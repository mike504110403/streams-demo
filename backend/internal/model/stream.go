package model

import (
	"time"

	"github.com/google/uuid"
)

// StreamStatus 直播狀態
type StreamStatus string

const (
	StreamStatusPending StreamStatus = "pending"
	StreamStatusLive    StreamStatus = "live"
	StreamStatusEnded   StreamStatus = "ended"
)

// Stream 對應 DB streams 表
type Stream struct {
	ID          uuid.UUID    `json:"id"`
	UserID      uuid.UUID    `json:"user_id"`
	Title       string       `json:"title"`
	CoverURL    *string      `json:"cover_url"`
	StreamKey   string       `json:"stream_key"`
	Status      StreamStatus `json:"status"`
	StartedAt   *time.Time   `json:"started_at"`
	EndedAt     *time.Time   `json:"ended_at"`
	ViewerCount int          `json:"viewer_count"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// CreateStreamRequest 建立直播間請求
type CreateStreamRequest struct {
	Title    string  `json:"title" binding:"required,min=1,max=100"`
	CoverURL *string `json:"cover_url" binding:"omitempty"`
}

// UpdateStreamRequest 更新直播間請求
type UpdateStreamRequest struct {
	Title    *string `json:"title" binding:"omitempty,min=1,max=100"`
	CoverURL *string `json:"cover_url" binding:"omitempty"`
}

// StreamResponse 直播間回應
type StreamResponse struct {
	ID          uuid.UUID    `json:"id"`
	UserID      uuid.UUID    `json:"user_id"`
	Title       string       `json:"title"`
	CoverURL    *string      `json:"cover_url"`
	StreamKey   string       `json:"stream_key,omitempty"`
	Status      StreamStatus `json:"status"`
	StartedAt   *time.Time   `json:"started_at,omitempty"`
	EndedAt     *time.Time   `json:"ended_at,omitempty"`
	ViewerCount int          `json:"viewer_count"`
	RTMPURL     string       `json:"rtmp_url,omitempty"`
	FLVURL      string       `json:"flv_url,omitempty"`
	HLSURL      string       `json:"hls_url,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
}

// ToResponse 將 Stream 轉換為 StreamResponse（不含推流地址）
func (s *Stream) ToResponse() StreamResponse {
	return StreamResponse{
		ID:          s.ID,
		UserID:      s.UserID,
		Title:       s.Title,
		CoverURL:    s.CoverURL,
		Status:      s.Status,
		StartedAt:   s.StartedAt,
		EndedAt:     s.EndedAt,
		ViewerCount: s.ViewerCount,
		CreatedAt:   s.CreatedAt,
	}
}

// ToOwnerResponse 將 Stream 轉換為 StreamResponse（含推流地址，給擁有者看）
func (s *Stream) ToOwnerResponse() StreamResponse {
	resp := s.ToResponse()
	resp.StreamKey = s.StreamKey
	resp.RTMPURL = "rtmp://localhost:1935/live/" + s.StreamKey
	return resp
}

// ToPlayerResponse 將 Stream 轉換為 StreamResponse（含拉流地址，給觀眾看）
func (s *Stream) ToPlayerResponse() StreamResponse {
	resp := s.ToResponse()
	resp.FLVURL = "http://localhost:8080/live/" + s.StreamKey + ".flv"
	resp.HLSURL = "http://localhost:8080/live/" + s.StreamKey + ".m3u8"
	return resp
}

// SRSCallbackRequest SRS 回呼請求
type SRSCallbackRequest struct {
	Action   string `json:"action"`
	ClientID string `json:"client_id"`
	IP       string `json:"ip"`
	Vhost    string `json:"vhost"`
	App      string `json:"app"`
	Stream   string `json:"stream"`
	Param    string `json:"param"`
}

// ListStreamsResponse 直播列表回應
type ListStreamsResponse struct {
	Streams []StreamResponse `json:"streams"`
	Total   int              `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
}
