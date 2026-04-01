package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
)

// Room 代表一個直播間的聊天室
type Room struct {
	streamID  uuid.UUID
	clients   map[*Client]bool
	mu        sync.RWMutex
	join      chan *Client
	leave     chan *Client
	broadcast chan []byte
	done      chan struct{}
}

// NewRoom 建立 Room
func NewRoom(streamID uuid.UUID) *Room {
	return &Room{
		streamID:  streamID,
		clients:   make(map[*Client]bool),
		join:      make(chan *Client, 16),
		leave:     make(chan *Client, 16),
		broadcast: make(chan []byte, 256),
		done:      make(chan struct{}),
	}
}

// Run 啟動 Room 事件迴圈
func (r *Room) Run() {
	for {
		select {
		case client := <-r.join:
			r.mu.Lock()
			r.clients[client] = true
			count := len(r.clients)
			r.mu.Unlock()

			// 廣播「XXX 進入直播間」
			sysMsg := NewSystemMessage(client.nickname+" 進入直播間", count)
			if data, err := json.Marshal(sysMsg); err == nil {
				r.broadcastToAll(data)
			}
			log.Printf("[WS] %s 加入 room %s (人數: %d)", client.nickname, r.streamID, count)

		case client := <-r.leave:
			r.mu.Lock()
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
			}
			count := len(r.clients)
			r.mu.Unlock()

			// 廣播「XXX 離開直播間」
			sysMsg := NewSystemMessage(client.nickname+" 離開直播間", count)
			if data, err := json.Marshal(sysMsg); err == nil {
				r.broadcastToAll(data)
			}
			log.Printf("[WS] %s 離開 room %s (人數: %d)", client.nickname, r.streamID, count)

		case msg := <-r.broadcast:
			r.broadcastToAll(msg)

		case <-r.done:
			// Room 關閉，清理所有連線
			r.mu.Lock()
			for client := range r.clients {
				close(client.send)
				delete(r.clients, client)
			}
			r.mu.Unlock()
			return
		}
	}
}

// Join 將 client 加入 Room
func (r *Room) Join(client *Client) {
	r.join <- client
}

// Close 關閉 Room
func (r *Room) Close() {
	select {
	case <-r.done:
		// 已關閉
	default:
		close(r.done)
	}
}

// ClientCount 取得連線人數
func (r *Room) ClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

// broadcastToAll 發送訊息給所有客戶端
func (r *Room) broadcastToAll(data []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for client := range r.clients {
		select {
		case client.send <- data:
		default:
			// send buffer 滿了，跳過（避免阻塞）
			log.Printf("[WS] 跳過慢客戶端 user=%s", client.userID)
		}
	}
}
