package handler

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/streams-demo/backend/internal/repository"
	"github.com/streams-demo/backend/internal/service"
	"github.com/streams-demo/backend/internal/ws"
	"nhooyr.io/websocket"
)

// WSHandler WebSocket 處理器
type WSHandler struct {
	hub         *ws.Hub
	authService *service.AuthService
	chatRepo    *repository.ChatRepository
}

// NewWSHandler 建立 WSHandler
func NewWSHandler(hub *ws.Hub, authService *service.AuthService, chatRepo *repository.ChatRepository) *WSHandler {
	return &WSHandler{
		hub:         hub,
		authService: authService,
		chatRepo:    chatRepo,
	}
}

// HandleChat 處理 WebSocket 連線：ws/chat/:stream_id?token=xxx
func (h *WSHandler) HandleChat(c *gin.Context) {
	// 解析 stream_id
	streamIDStr := c.Param("stream_id")
	streamID, err := uuid.Parse(streamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stream_id"})
		return
	}

	// 從 query param 取得 token（WebSocket 無法用 header）
	token := c.Query("token")
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = authHeader[7:]
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	// 驗證 token
	userID, err := h.authService.ValidateAccessToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// 取得用戶暱稱
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	// 升級為 WebSocket
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // MVP: 允許所有來源
	})
	if err != nil {
		log.Printf("[WS] 升級失敗: %v", err)
		return
	}

	// 取得或建立 Room
	room := h.hub.GetOrCreateRoom(streamID)

	// 建立 Client，帶持久化回調
	persister := func(uid uuid.UUID, content string) {
		if err := h.chatRepo.InsertMessage(context.Background(), streamID, uid, content, "chat"); err != nil {
			log.Printf("[WS] 持久化失敗: %v", err)
		}
	}
	client := ws.NewClient(conn, room, userID, user.Nickname, persister)

	// 註冊到 Room
	room.Join(client)

	// 啟動讀寫 pump
	ctx := c.Request.Context()
	go client.WritePump(ctx)
	client.ReadPump(ctx) // blocking，結束時 client 會自動 leave room
}
