package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/streams-demo/backend/internal/model"
	"github.com/streams-demo/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrEmailExists        = errors.New("email already registered")
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
)

// AuthService 認證業務邏輯
type AuthService struct {
	userRepo  *repository.UserRepository
	rdb       *redis.Client
	jwtSecret []byte
}

// NewAuthService 建立 AuthService
func NewAuthService(userRepo *repository.UserRepository, rdb *redis.Client, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		rdb:       rdb,
		jwtSecret: []byte(jwtSecret),
	}
}

// Register 註冊新用戶
func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, error) {
	// bcrypt 加密密碼
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 建立用戶
	user, err := s.userRepo.CreateUser(ctx, req.Email, string(hash), req.Nickname)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return nil, ErrEmailExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 產生 JWT tokens
	return s.generateAuthResponse(ctx, user)
}

// Login 用戶登入
func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	// 查詢用戶
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 驗證密碼
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 產生 JWT tokens
	return s.generateAuthResponse(ctx, user)
}

// RefreshToken 刷新 token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*model.AuthResponse, error) {
	// 從 Redis 驗證 refresh token
	key := fmt.Sprintf("refresh_token:%s", refreshToken)
	userIDStr, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to check refresh token: %w", err)
	}

	// 同時驗證 JWT 簽名
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// 確認 token type 是 refresh
	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return nil, ErrInvalidToken
	}

	// 解析 user ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id in token: %w", err)
	}

	// 查詢用戶
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 刪除舊的 refresh token
	s.rdb.Del(ctx, key)

	// 產生新的 token pair
	return s.generateAuthResponse(ctx, user)
}

// Logout 登出（將 access token 加入黑名單）
func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	// 解析 token 取得剩餘有效期
	claims, err := s.parseToken(accessToken)
	if err != nil {
		// token 已經無效，不需要加入黑名單
		return nil
	}

	// 計算剩餘 TTL
	exp, err := claims.GetExpirationTime()
	if err != nil {
		return nil
	}
	ttl := time.Until(exp.Time)
	if ttl <= 0 {
		return nil
	}

	// 加入黑名單
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

// generateAuthResponse 產生包含 access/refresh token 的回應
func (s *AuthService) generateAuthResponse(ctx context.Context, user *model.User) (*model.AuthResponse, error) {
	now := time.Now()

	// 產生 access token
	accessToken, err := s.createToken(user.ID.String(), "access", now, accessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// 產生 refresh token
	refreshToken, err := s.createToken(user.ID.String(), "refresh", now, refreshTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	// 將 refresh token 存入 Redis
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
