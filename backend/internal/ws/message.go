package ws

import "time"

// MessageType WebSocket 訊息類型
type MessageType string

const (
	MessageTypeChat   MessageType = "chat"
	MessageTypeSystem MessageType = "system"
)

// IncomingMessage 客戶端發送的訊息
type IncomingMessage struct {
	Type    MessageType `json:"type"`
	Content string      `json:"content"`
}

// OutgoingMessage 伺服器廣播的訊息
type OutgoingMessage struct {
	Type        MessageType `json:"type"`
	UserID      string      `json:"user_id,omitempty"`
	Nickname    string      `json:"nickname"`
	Content     string      `json:"content"`
	Timestamp   int64       `json:"timestamp"`
	ViewerCount int         `json:"viewer_count,omitempty"`
}

// NewChatMessage 建立聊天訊息
func NewChatMessage(userID, nickname, content string) OutgoingMessage {
	return OutgoingMessage{
		Type:      MessageTypeChat,
		UserID:    userID,
		Nickname:  nickname,
		Content:   content,
		Timestamp: time.Now().Unix(),
	}
}

// NewSystemMessage 建立系統訊息（進入/離開直播間）
func NewSystemMessage(content string, viewerCount int) OutgoingMessage {
	return OutgoingMessage{
		Type:        MessageTypeSystem,
		Nickname:    "系統",
		Content:     content,
		Timestamp:   time.Now().Unix(),
		ViewerCount: viewerCount,
	}
}
