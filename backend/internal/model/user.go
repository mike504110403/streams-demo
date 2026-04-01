package model

import (
	"time"

	"github.com/google/uuid"
)

// User 對應 DB users 表（auth revamp 後無 password_hash）
type User struct {
	ID        uuid.UUID `json:"id"`
	Email     *string   `json:"email"`
	Phone     string    `json:"phone"`
	Nickname  string    `json:"nickname"`
	AvatarURL *string   `json:"avatar_url"`
	Bio       *string   `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SendCodeRequest 發送驗證碼請求
type SendCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// RegisterRequest 註冊請求（手機號碼 + 驗證碼）
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Code     string `json:"code" binding:"required,len=6"`
	Nickname string `json:"nickname" binding:"required,min=1,max=50"`
}

// LoginRequest 登入請求（手機號碼 + 驗證碼）
type LoginRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required,len=6"`
}

// OAuthAppleRequest Apple OAuth 登入請求
type OAuthAppleRequest struct {
	Code    string           `json:"code" binding:"required"`
	IDToken string           `json:"id_token" binding:"required"`
	User    *OAuthAppleUser  `json:"user,omitempty"`
}

// OAuthAppleUser Apple 首次授權時提供的用戶資訊
type OAuthAppleUser struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// OAuthGoogleRequest Google OAuth 登入請求
type OAuthGoogleRequest struct {
	Credential string `json:"credential" binding:"required"`
}

// UpdateProfileRequest 更新個人資料請求
type UpdateProfileRequest struct {
	Nickname  *string `json:"nickname" binding:"omitempty,min=1,max=50"`
	AvatarURL *string `json:"avatar_url" binding:"omitempty"`
	Bio       *string `json:"bio" binding:"omitempty,max=500"`
}

// UserResponse 用戶回應（不含敏感資訊）
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Phone     string    `json:"phone"`
	Email     *string   `json:"email,omitempty"`
	Nickname  string    `json:"nickname"`
	AvatarURL *string   `json:"avatar_url"`
	Bio       *string   `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthResponse 認證回應
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

// OAuthResponse OAuth 認證回應（含 is_new_user）
type OAuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	IsNewUser    bool         `json:"is_new_user"`
	User         UserResponse `json:"user"`
}

// ToResponse 將 User 轉換為 UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Phone:     u.Phone,
		Email:     u.Email,
		Nickname:  u.Nickname,
		AvatarURL: u.AvatarURL,
		Bio:       u.Bio,
		CreatedAt: u.CreatedAt,
	}
}
