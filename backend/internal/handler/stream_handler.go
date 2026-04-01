package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// StreamHandler 直播間相關 HTTP Handler
type StreamHandler struct{}

// NewStreamHandler 建立 StreamHandler
func NewStreamHandler() *StreamHandler {
	return &StreamHandler{}
}

// CreateStream POST /api/v1/streams — 建立直播間
func (h *StreamHandler) CreateStream(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

// ListStreams GET /api/v1/streams — 直播列表
func (h *StreamHandler) ListStreams(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

// GetStream GET /api/v1/streams/:id — 直播間詳情
func (h *StreamHandler) GetStream(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

// UpdateStream PUT /api/v1/streams/:id — 更新直播間
func (h *StreamHandler) UpdateStream(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}

// DeleteStream DELETE /api/v1/streams/:id — 結束直播
func (h *StreamHandler) DeleteStream(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented"})
}
