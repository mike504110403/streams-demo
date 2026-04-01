package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streams-demo/backend/internal/model"
	"github.com/streams-demo/backend/internal/service"
)

// StreamHandler 直播間相關 HTTP Handler
type StreamHandler struct {
	streamService *service.StreamService
}

// NewStreamHandler 建立 StreamHandler
func NewStreamHandler(streamService *service.StreamService) *StreamHandler {
	return &StreamHandler{streamService: streamService}
}

// CreateStream POST /api/v1/streams — 建立直播間
func (h *StreamHandler) CreateStream(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req model.CreateStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.streamService.CreateStream(c.Request.Context(), userID.(uuid.UUID), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

// ListStreams GET /api/v1/streams — 直播列表
func (h *StreamHandler) ListStreams(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	resp, err := h.streamService.ListStreams(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// GetStream GET /api/v1/streams/:id — 直播間詳情
func (h *StreamHandler) GetStream(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stream id"})
		return
	}

	resp, err := h.streamService.GetStream(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrStreamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "直播間不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// UpdateStream PUT /api/v1/streams/:id — 更新直播間
func (h *StreamHandler) UpdateStream(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stream id"})
		return
	}

	var req model.UpdateStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.streamService.UpdateStream(c.Request.Context(), id, userID.(uuid.UUID), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStreamNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "直播間不存在"})
		case errors.Is(err, service.ErrStreamNotOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "無權操作此直播間"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// DeleteStream DELETE /api/v1/streams/:id — 結束直播
func (h *StreamHandler) DeleteStream(c *gin.Context) {
	userID, _ := c.Get("user_id")

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stream id"})
		return
	}

	resp, err := h.streamService.EndStream(c.Request.Context(), id, userID.(uuid.UUID))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrStreamNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "直播間不存在"})
		case errors.Is(err, service.ErrStreamNotOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "無權操作此直播間"})
		case errors.Is(err, service.ErrStreamAlreadyEnded):
			c.JSON(http.StatusConflict, gin.H{"error": "直播已結束"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}
