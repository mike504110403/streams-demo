package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streams-demo/backend/internal/model"
	"github.com/streams-demo/backend/internal/repository"
	"github.com/streams-demo/backend/internal/service"
)

// UserHandler 用戶相關 HTTP Handler
type UserHandler struct {
	authService *service.AuthService
}

// NewUserHandler 建立 UserHandler
func NewUserHandler(authService *service.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

// GetMe GET /api/v1/users/me (需認證)
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.authService.GetUserByID(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user.ToResponse()})
}

// UpdateMe PUT /api/v1/users/me (需認證)
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 至少要有一個欄位
	if req.Nickname == nil && req.AvatarURL == nil && req.Bio == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field must be provided"})
		return
	}

	user, err := h.authService.UpdateProfile(c.Request.Context(), userID.(uuid.UUID), req)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user.ToResponse()})
}
