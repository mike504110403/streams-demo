package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streams-demo/backend/internal/model"
	"github.com/streams-demo/backend/internal/service"
)

// AuthHandler 認證相關 HTTP Handler
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 建立 AuthHandler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// SendCode POST /api/v1/auth/send-code
func (h *AuthHandler) SendCode(c *gin.Context) {
	var req model.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	retryAfter, err := h.authService.SendCode(c.Request.Context(), req.Phone)
	if err != nil {
		if errors.Is(err, service.ErrTooManyRequests) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "請等待 60 秒後再試",
				"retry_after": retryAfter,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"message":    "驗證碼已發送",
		"expires_in": 300,
	}})
}

// Register POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPhoneExists):
			c.JSON(http.StatusConflict, gin.H{"message": "此手機號碼已被註冊"})
		case errors.Is(err, service.ErrInvalidCode):
			c.JSON(http.StatusBadRequest, gin.H{"message": "驗證碼錯誤，請重新輸入"})
		case errors.Is(err, service.ErrCodeExpired):
			c.JSON(http.StatusBadRequest, gin.H{"message": "驗證碼已過期，請重新發送"})
		case errors.Is(err, service.ErrCodeAlreadyUsed):
			c.JSON(http.StatusBadRequest, gin.H{"message": "驗證碼已使用，請重新發送"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

// Login POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPhoneNotFound):
			c.JSON(http.StatusNotFound, gin.H{"message": "此手機號碼尚未註冊"})
		case errors.Is(err, service.ErrInvalidCode):
			c.JSON(http.StatusBadRequest, gin.H{"message": "驗證碼錯誤，請重新輸入"})
		case errors.Is(err, service.ErrCodeExpired):
			c.JSON(http.StatusBadRequest, gin.H{"message": "驗證碼已過期，請重新發送"})
		case errors.Is(err, service.ErrCodeAlreadyUsed):
			c.JSON(http.StatusBadRequest, gin.H{"message": "驗證碼已使用，請重新發送"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// OAuthApple POST /api/v1/auth/oauth/apple
func (h *AuthHandler) OAuthApple(c *gin.Context) {
	var req model.OAuthAppleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, isNew, err := h.authService.OAuthApple(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrOAuthFailed) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Apple 登入失敗，請重試"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	status := http.StatusOK
	if isNew {
		status = http.StatusCreated
	}
	c.JSON(status, gin.H{"data": resp})
}

// OAuthGoogle POST /api/v1/auth/oauth/google
func (h *AuthHandler) OAuthGoogle(c *gin.Context) {
	var req model.OAuthGoogleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, isNew, err := h.authService.OAuthGoogle(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrOAuthFailed) {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Google 登入失敗，請重試"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	status := http.StatusOK
	if isNew {
		status = http.StatusCreated
	}
	c.JSON(status, gin.H{"data": resp})
}

// RefreshToken POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// Logout POST /api/v1/auth/logout (需認證)
func (h *AuthHandler) Logout(c *gin.Context) {
	token, exists := c.Get("access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), token.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "logged out successfully"}})
}
