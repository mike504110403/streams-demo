package ws

import (
	"log"
	"sync"

	"github.com/google/uuid"
)

// Hub 管理所有直播間的 WebSocket Room
type Hub struct {
	mu    sync.RWMutex
	rooms map[uuid.UUID]*Room
}

// NewHub 建立 Hub
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[uuid.UUID]*Room),
	}
}

// GetOrCreateRoom 取得或建立直播間 Room
func (h *Hub) GetOrCreateRoom(streamID uuid.UUID) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[streamID]; ok {
		return room
	}

	room := NewRoom(streamID)
	h.rooms[streamID] = room
	go room.Run()
	log.Printf("[WS] Room 已建立: stream_id=%s", streamID)
	return room
}

// RemoveRoom 移除空的 Room
func (h *Hub) RemoveRoom(streamID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[streamID]; ok {
		room.Close()
		delete(h.rooms, streamID)
		log.Printf("[WS] Room 已移除: stream_id=%s", streamID)
	}
}

// GetRoom 取得 Room（可能為 nil）
func (h *Hub) GetRoom(streamID uuid.UUID) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[streamID]
}

// ViewerCount 取得某直播間的觀看人數
func (h *Hub) ViewerCount(streamID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if room, ok := h.rooms[streamID]; ok {
		return room.ClientCount()
	}
	return 0
}
