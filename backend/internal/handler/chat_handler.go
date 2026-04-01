package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streams-demo/backend/internal/repository"
	"github.com/streams-demo/backend/internal/ws"
)

// ChatHandler 聊天相關 HTTP Handler
type ChatHandler struct {
	chatRepo *repository.ChatRepository
	hub      *ws.Hub
}

// NewChatHandler 建立 ChatHandler
func NewChatHandler(chatRepo *repository.ChatRepository, hub *ws.Hub) *ChatHandler {
	return &ChatHandler{
		chatRepo: chatRepo,
		hub:      hub,
	}
}

// GetMessages GET /api/v1/streams/:id/messages — 聊天歷史（cursor 分頁）
func (h *ChatHandler) GetMessages(c *gin.Context) {
	streamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stream id"})
		return
	}

	// 解析 limit（預設 50，最大 200）
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	// 解析 before cursor（RFC3339 timestamp）
	var before time.Time
	if beforeStr := c.Query("before"); beforeStr != "" {
		before, err = time.Parse(time.RFC3339Nano, beforeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid before timestamp, use RFC3339 format"})
			return
		}
	}

	// 查詢 limit+1 筆來判斷是否還有更多
	messages, err := h.chatRepo.GetMessagesByStreamID(c.Request.Context(), streamID, limit, before)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	// 反轉為時間正序（DB 查出來是 DESC）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	// 確保回傳空陣列而非 null
	if messages == nil {
		messages = []repository.ChatMessage{}
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"has_more": hasMore,
	})
}

// GetViewers GET /api/v1/streams/:id/viewers — 即時觀看人數
func (h *ChatHandler) GetViewers(c *gin.Context) {
	streamID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stream id"})
		return
	}

	count := h.hub.ViewerCount(streamID)

	c.JSON(http.StatusOK, gin.H{
		"stream_id":    streamID,
		"viewer_count": count,
	})
}
