package model

import (
	"time"

	"github.com/google/uuid"
)

// User 對應 DB users 表
type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	Phone        *string    `json:"phone"`
	PasswordHash string     `json:"-"`
	Nickname     string     `json:"nickname"`
	AvatarURL    *string    `json:"avatar_url"`
	Bio          *string    `json:"bio"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// RegisterRequest 註冊請求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Nickname string `json:"nickname" binding:"required,min=1,max=50"`
}

// LoginRequest 登入請求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest 更新個人資料請求
type UpdateProfileRequest struct {
	Nickname  *string `json:"nickname" binding:"omitempty,min=1,max=50"`
	AvatarURL *string `json:"avatar_url" binding:"omitempty,url"`
	Bio       *string `json:"bio" binding:"omitempty,max=500"`
}

// UserResponse 用戶回應（不含敏感資訊）
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone,omitempty"`
	Nickname  string    `json:"nickname"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Bio       *string   `json:"bio,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthResponse 認證回應
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

// ToResponse 將 User 轉換為 UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Phone:     u.Phone,
		Nickname:  u.Nickname,
		AvatarURL: u.AvatarURL,
		Bio:       u.Bio,
		CreatedAt: u.CreatedAt,
	}
}
