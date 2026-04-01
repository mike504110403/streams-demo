package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SRSHandler SRS Callback HTTP Handler（內部使用，不走 JWT）
type SRSHandler struct{}

// NewSRSHandler 建立 SRSHandler
func NewSRSHandler() *SRSHandler {
	return &SRSHandler{}
}

// OnPublish POST /api/v1/internal/srs/on_publish — 推流開始回呼
// SRS 要求回傳 HTTP 200 + code 0 表示允許推流
func (h *SRSHandler) OnPublish(c *gin.Context) {
	log.Printf("[SRS] on_publish callback received")
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "not implemented"})
}

// OnUnpublish POST /api/v1/internal/srs/on_unpublish — 推流結束回呼
func (h *SRSHandler) OnUnpublish(c *gin.Context) {
	log.Printf("[SRS] on_unpublish callback received")
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "not implemented"})
}

// OnPlay POST /api/v1/internal/srs/on_play — 觀眾開始播放回呼
func (h *SRSHandler) OnPlay(c *gin.Context) {
	log.Printf("[SRS] on_play callback received")
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "not implemented"})
}

// OnStop POST /api/v1/internal/srs/on_stop — 觀眾停止播放回呼
func (h *SRSHandler) OnStop(c *gin.Context) {
	log.Printf("[SRS] on_stop callback received")
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "not implemented"})
}
