package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streams-demo/backend/internal/model"
	"github.com/streams-demo/backend/internal/service"
)

// SRSHandler SRS Callback HTTP Handler（內部使用，不走 JWT）
type SRSHandler struct {
	streamService *service.StreamService
}

// NewSRSHandler 建立 SRSHandler
func NewSRSHandler(streamService *service.StreamService) *SRSHandler {
	return &SRSHandler{streamService: streamService}
}

// OnPublish POST /api/v1/internal/srs/on_publish — 推流開始回呼
// SRS 要求回傳 HTTP 200 + code 0 表示允許推流
func (h *SRSHandler) OnPublish(c *gin.Context) {
	var req model.SRSCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[SRS] on_publish 解析請求失敗: %v", err)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "invalid request"})
		return
	}

	log.Printf("[SRS] on_publish: stream_key=%s, ip=%s, app=%s", req.Stream, req.IP, req.App)

	if err := h.streamService.HandlePublish(c.Request.Context(), req.Stream); err != nil {
		if errors.Is(err, service.ErrStreamNotFound) {
			log.Printf("[SRS] on_publish 拒絕：找不到對應的直播間 stream_key=%s", req.Stream)
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "stream not found"})
			return
		}
		log.Printf("[SRS] on_publish 處理失敗: %v", err)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0})
}

// OnUnpublish POST /api/v1/internal/srs/on_unpublish — 推流結束回呼
func (h *SRSHandler) OnUnpublish(c *gin.Context) {
	var req model.SRSCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[SRS] on_unpublish 解析請求失敗: %v", err)
		c.JSON(http.StatusOK, gin.H{"code": 0})
		return
	}

	log.Printf("[SRS] on_unpublish: stream_key=%s, ip=%s", req.Stream, req.IP)

	if err := h.streamService.HandleUnpublish(c.Request.Context(), req.Stream); err != nil {
		log.Printf("[SRS] on_unpublish 處理失敗: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"code": 0})
}

// OnPlay POST /api/v1/internal/srs/on_play — 觀眾開始播放回呼
func (h *SRSHandler) OnPlay(c *gin.Context) {
	var req model.SRSCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[SRS] on_play 解析請求失敗: %v", err)
		c.JSON(http.StatusOK, gin.H{"code": 0})
		return
	}

	log.Printf("[SRS] on_play: stream_key=%s, ip=%s, client_id=%s", req.Stream, req.IP, req.ClientID)
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

// OnStop POST /api/v1/internal/srs/on_stop — 觀眾停止播放回呼
func (h *SRSHandler) OnStop(c *gin.Context) {
	var req model.SRSCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[SRS] on_stop 解析請求失敗: %v", err)
		c.JSON(http.StatusOK, gin.H{"code": 0})
		return
	}

	log.Printf("[SRS] on_stop: stream_key=%s, ip=%s, client_id=%s", req.Stream, req.IP, req.ClientID)
	c.JSON(http.StatusOK, gin.H{"code": 0})
}
