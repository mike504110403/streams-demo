package ws

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

const (
	writeTimeout = 10 * time.Second
	readLimit    = 4096
)

// MessagePersister 訊息持久化回調
type MessagePersister func(userID uuid.UUID, content string)

// Client 代表一個 WebSocket 連線
type Client struct {
	conn      *websocket.Conn
	room      *Room
	userID    uuid.UUID
	nickname  string
	send      chan []byte
	onMessage MessagePersister
	readOnly  bool // 未登入用戶為唯讀模式，只能接收彈幕不能發送
}

// NewClient 建立 Client
func NewClient(conn *websocket.Conn, room *Room, userID uuid.UUID, nickname string, onMessage MessagePersister) *Client {
	return &Client{
		conn:      conn,
		room:      room,
		userID:    userID,
		nickname:  nickname,
		send:      make(chan []byte, 64),
		onMessage: onMessage,
		readOnly:  false,
	}
}

// NewReadOnlyClient 建立唯讀 Client（未登入用戶，只能接收彈幕不能發送）
func NewReadOnlyClient(conn *websocket.Conn, room *Room) *Client {
	return &Client{
		conn:     conn,
		room:     room,
		userID:   uuid.Nil,
		nickname: "訪客",
		send:     make(chan []byte, 64),
		readOnly: true,
	}
}

// ReadPump 從 WebSocket 讀取訊息
func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.room.leave <- c
	}()

	c.conn.SetReadLimit(readLimit)

	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			// 正常關閉或錯誤都走這裡
			log.Printf("[WS] 讀取結束 user=%s: %v", c.userID, err)
			return
		}

		// 唯讀用戶不能發送彈幕，忽略訊息
		if c.readOnly {
			log.Printf("[WS] 唯讀用戶嘗試發送訊息，已忽略")
			continue
		}

		var msg IncomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("[WS] 訊息格式錯誤: %v", err)
			continue
		}

		if msg.Content == "" || len(msg.Content) > 500 {
			continue
		}

		// 建立廣播訊息
		outMsg := NewChatMessage(c.userID.String(), c.nickname, msg.Content)
		outData, err := json.Marshal(outMsg)
		if err != nil {
			continue
		}

		c.room.broadcast <- outData

		// 非同步持久化
		if c.onMessage != nil {
			go c.onMessage(c.userID, msg.Content)
		}
	}
}

// WritePump 將訊息寫入 WebSocket
func (c *Client) WritePump(ctx context.Context) {
	defer c.conn.CloseNow()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				// send channel 已關閉
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				log.Printf("[WS] 寫入失敗 user=%s: %v", c.userID, err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
