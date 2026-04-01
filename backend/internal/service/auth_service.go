package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/streams-demo/backend/internal/model"
	"github.com/streams-demo/backend/internal/repository"
)

var (
	ErrInvalidCode      = errors.New("invalid or expired verification code")
	ErrCodeExpired      = errors.New("verification code expired")
	ErrCodeAlreadyUsed  = errors.New("verification code already used")
	ErrPhoneNotFound    = errors.New("phone number not registered")
	ErrPhoneExists      = errors.New("phone number already registered")
	ErrInvalidToken     = errors.New("invalid or expired token")
	ErrTooManyRequests  = errors.New("too many requests")
	ErrOAuthFailed      = errors.New("oauth verification failed")
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
	codeExpiry           = 5 * time.Minute
	codeCooldown         = 60 * time.Second
	mockVerificationCode = "123456"
)

// AuthService 認證業務邏輯
type AuthService struct {
	userRepo  *repository.UserRepository
	rdb       *redis.Client
	jwtSecret []byte
	debugMode bool
}

// NewAuthService 建立 AuthService
func NewAuthService(userRepo *repository.UserRepository, rdb *redis.Client, jwtSecret string, debugMode bool) *AuthService {
	if debugMode {
		log.Println("[DEBUG] AuthService: DEBUG_MODE 啟用，驗證碼將跳過比對")
	}
	return &AuthService{
		userRepo:  userRepo,
		rdb:       rdb,
		jwtSecret: []byte(jwtSecret),
		debugMode: debugMode,
	}
}

// SendCode 發送 SMS 驗證碼（MVP mock：固定 123456）
func (s *AuthService) SendCode(ctx context.Context, phone string) (int, error) {
	// 檢查 60 秒冷卻期
	lastCreated, err := s.userRepo.GetLatestCodeCreatedAt(ctx, phone)
	if err != nil {
		return 0, fmt.Errorf("failed to check code cooldown: %w", err)
	}
	if !lastCreated.IsZero() {
		elapsed := time.Since(lastCreated)
		if elapsed < codeCooldown {
			remaining := int(codeCooldown.Seconds() - elapsed.Seconds())
			return remaining, ErrTooManyRequests
		}
	}

	// 產生驗證碼（MVP 固定 123456）
	code := mockVerificationCode
	expiresAt := time.Now().Add(codeExpiry)

	// 存入 DB
	if err := s.userRepo.SaveVerificationCode(ctx, phone, code, expiresAt); err != nil {
		return 0, fmt.Errorf("failed to save verification code: %w", err)
	}

	// Mock SMS：輸出到 console
	log.Printf("[MOCK SMS] Phone: %s, Code: %s", phone, code)

	return 0, nil
}

// Register 用手機號碼 + 驗證碼註冊
func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, error) {
	// 驗證驗證碼
	if err := s.verifyCode(ctx, req.Phone, req.Code); err != nil {
		return nil, err
	}

	// 檢查手機號碼是否已存在
	_, err := s.userRepo.GetUserByPhone(ctx, req.Phone)
	if err == nil {
		return nil, ErrPhoneExists
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check phone: %w", err)
	}

	// 建立用戶
	user, err := s.userRepo.CreateUserByPhone(ctx, req.Phone, req.Nickname)
	if err != nil {
		if errors.Is(err, repository.ErrPhoneAlreadyExists) {
			return nil, ErrPhoneExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 標記驗證碼為已使用
	_ = s.userRepo.MarkVerificationCodeUsed(ctx, req.Phone, req.Code)

	return s.generateAuthResponse(ctx, user)
}

// Login 用手機號碼 + 驗證碼登入
func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	// 驗證驗證碼
	if err := s.verifyCode(ctx, req.Phone, req.Code); err != nil {
		return nil, err
	}

	// 查詢用戶
	user, err := s.userRepo.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrPhoneNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 標記驗證碼為已使用
	_ = s.userRepo.MarkVerificationCodeUsed(ctx, req.Phone, req.Code)

	return s.generateAuthResponse(ctx, user)
}

// OAuthApple Apple OAuth 登入/自動註冊（MVP mock：不真正驗證 Apple token）
func (s *AuthService) OAuthApple(ctx context.Context, req model.OAuthAppleRequest) (*model.OAuthResponse, bool, error) {
	// MVP：從 id_token 產生一個穩定的 mock sub（用 hash 確保同一 token 產生同一 sub）
	sub := "apple_" + hashString(req.IDToken)

	// 查詢是否已綁定
	user, err := s.userRepo.FindUserByOAuth(ctx, "apple", sub)
	if err == nil {
		// 已有帳號，直接登入
		resp, err := s.generateAuthResponse(ctx, user)
		if err != nil {
			return nil, false, err
		}
		return &model.OAuthResponse{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
			IsNewUser:    false,
			User:         resp.User,
		}, false, nil
	}
	if !errors.Is(err, repository.ErrOAuthNotFound) {
		return nil, false, fmt.Errorf("failed to find oauth user: %w", err)
	}

	// 新用戶：自動建立帳號
	nickname := "Apple 用戶"
	email := ""
	if req.User != nil {
		if req.User.Name != "" {
			nickname = req.User.Name
		}
		email = req.User.Email
	}

	newUser, err := s.userRepo.CreateUserForOAuth(ctx, nickname, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create oauth user: %w", err)
	}

	// 建立 OAuth provider 記錄
	if err := s.userRepo.CreateOAuthProvider(ctx, newUser.ID, "apple", sub, email, nickname); err != nil {
		return nil, false, fmt.Errorf("failed to create oauth provider: %w", err)
	}

	resp, err := s.generateAuthResponse(ctx, newUser)
	if err != nil {
		return nil, false, err
	}
	return &model.OAuthResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		IsNewUser:    true,
		User:         resp.User,
	}, true, nil
}

// OAuthGoogle Google OAuth 登入/自動註冊（MVP mock：不真正驗證 Google token）
func (s *AuthService) OAuthGoogle(ctx context.Context, req model.OAuthGoogleRequest) (*model.OAuthResponse, bool, error) {
	// MVP：從 credential 產生一個穩定的 mock sub
	sub := "google_" + hashString(req.Credential)

	// 查詢是否已綁定
	user, err := s.userRepo.FindUserByOAuth(ctx, "google", sub)
	if err == nil {
		resp, err := s.generateAuthResponse(ctx, user)
		if err != nil {
			return nil, false, err
		}
		return &model.OAuthResponse{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
			IsNewUser:    false,
			User:         resp.User,
		}, false, nil
	}
	if !errors.Is(err, repository.ErrOAuthNotFound) {
		return nil, false, fmt.Errorf("failed to find oauth user: %w", err)
	}

	// 新用戶
	nickname := "Google 用戶"
	avatarURL := "https://fastly.jsdelivr.net/npm/@vant/assets/cat.jpeg"

	newUser, err := s.userRepo.CreateUserForOAuth(ctx, nickname, &avatarURL)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create oauth user: %w", err)
	}

	if err := s.userRepo.CreateOAuthProvider(ctx, newUser.ID, "google", sub, "", nickname); err != nil {
		return nil, false, fmt.Errorf("failed to create oauth provider: %w", err)
	}

	resp, err := s.generateAuthResponse(ctx, newUser)
	if err != nil {
		return nil, false, err
	}
	return &model.OAuthResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		IsNewUser:    true,
		User:         resp.User,
	}, true, nil
}

// RefreshToken 刷新 token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*model.AuthResponse, error) {
	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	userIDStr, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to check refresh token: %w", err)
	}

	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id in token: %w", err)
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	s.rdb.Del(ctx, key)

	return s.generateAuthResponse(ctx, user)
}

// Logout 登出（將 access token 加入黑名單）
func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	claims, err := s.parseToken(accessToken)
	if err != nil {
		return nil
	}

	exp, err := claims.GetExpirationTime()
	if err != nil {
		return nil
	}
	ttl := time.Until(exp.Time)
	if ttl <= 0 {
		return nil
	}

	key := fmt.Sprintf("blacklist:%s", accessToken)
	if err := s.rdb.Set(ctx, key, "1", ttl).Err(); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// IsTokenBlacklisted 檢查 token 是否在黑名單中
func (s *AuthService) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", token)
	_, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ValidateAccessToken 驗證 access token 並回傳 user ID
func (s *AuthService) ValidateAccessToken(token string) (uuid.UUID, error) {
	claims, err := s.parseToken(token)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "access" {
		return uuid.Nil, ErrInvalidToken
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}

// GetUserByID 取得用戶資料
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.userRepo.GetUserByID(ctx, id)
}

// UpdateProfile 更新用戶資料
func (s *AuthService) UpdateProfile(ctx context.Context, id uuid.UUID, req model.UpdateProfileRequest) (*model.User, error) {
	return s.userRepo.UpdateProfile(ctx, id, req.Nickname, req.AvatarURL, req.Bio)
}

// === 內部方法 ===

// verifyCode 驗證 SMS 驗證碼
func (s *AuthService) verifyCode(ctx context.Context, phone, code string) error {
	// DEBUG_MODE：跳過驗證碼比對，任意碼都能通過
	if s.debugMode {
		log.Printf("[DEBUG] verifyCode: DEBUG_MODE 啟用，跳過驗證碼比對 (phone=%s)", phone)
		return nil
	}

	storedCode, expiresAt, used, err := s.userRepo.GetLatestVerificationCode(ctx, phone)
	if err != nil {
		return ErrInvalidCode
	}
	if used {
		return ErrCodeAlreadyUsed
	}
	if time.Now().After(expiresAt) {
		return ErrCodeExpired
	}
	if storedCode != code {
		return ErrInvalidCode
	}
	return nil
}

// generateAuthResponse 產生包含 access/refresh token 的回應
func (s *AuthService) generateAuthResponse(ctx context.Context, user *model.User) (*model.AuthResponse, error) {
	now := time.Now()

	accessToken, err := s.createToken(user.ID.String(), "access", now, accessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken, err := s.createToken(user.ID.String(), "refresh", now, refreshTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	if err := s.rdb.Set(ctx, key, user.ID.String(), refreshTokenDuration).Err(); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
	}, nil
}

// createToken 建立 JWT token
func (s *AuthService) createToken(subject, tokenType string, now time.Time, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":  subject,
		"type": tokenType,
		"iat":  now.Unix(),
		"exp":  now.Add(duration).Unix(),
		"jti":  uuid.New().String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// parseToken 解析並驗證 JWT token
func (s *AuthService) parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// hashString 用 SHA256 產生穩定的 hash（用於 OAuth mock sub）
func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8])
}
