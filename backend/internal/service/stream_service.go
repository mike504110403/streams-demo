package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/streams-demo/backend/config"
	"github.com/streams-demo/backend/internal/model"
	"github.com/streams-demo/backend/internal/repository"
)

var (
	ErrStreamNotFound    = errors.New("stream not found")
	ErrStreamNotOwner    = errors.New("you are not the owner of this stream")
	ErrStreamAlreadyEnded = errors.New("stream has already ended")
)

const (
	listCacheTTL   = 30 * time.Second
	listCachePrefix = "streams:list:"
)

// ViewerCounter 觀看人數查詢介面（避免直接依賴 ws 套件）
type ViewerCounter interface {
	ViewerCount(streamID uuid.UUID) int
}

// StreamService 直播間業務邏輯
type StreamService struct {
	repo          *repository.StreamRepository
	rdb           *redis.Client
	cfg           *config.Config
	viewerCounter ViewerCounter
}

// NewStreamService 建立 StreamService
func NewStreamService(repo *repository.StreamRepository, rdb *redis.Client, cfg *config.Config) *StreamService {
	return &StreamService{
		repo: repo,
		rdb:  rdb,
		cfg:  cfg,
	}
}

// SetViewerCounter 設定觀看人數查詢器（在 Hub 初始化後注入，避免循環依賴）
func (s *StreamService) SetViewerCounter(vc ViewerCounter) {
	s.viewerCounter = vc
}

// CreateStream 建立直播間，產生 stream_key 並回傳含推流地址的 response
func (s *StreamService) CreateStream(ctx context.Context, userID uuid.UUID, req model.CreateStreamRequest) (*model.StreamResponse, error) {
	streamKey := uuid.New().String()

	stream, err := s.repo.CreateStream(ctx, userID, req.Title, req.CoverURL, streamKey)
	if err != nil {
		return nil, fmt.Errorf("建立直播間失敗: %w", err)
	}

	// 清除列表快取
	_ = s.InvalidateListCache(ctx)

	resp := s.toOwnerResponse(stream)
	return &resp, nil
}

// ListStreams 分頁查詢直播列表，優先從 Redis 快取取得
func (s *StreamService) ListStreams(ctx context.Context, page, limit int) (*model.ListStreamsResponse, error) {
	offset := (page - 1) * limit
	cacheKey := fmt.Sprintf("%spage:%d:limit:%d", listCachePrefix, page, limit)

	// 嘗試從 Redis 快取取得
	cached, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		var resp model.ListStreamsResponse
		if jsonErr := json.Unmarshal([]byte(cached), &resp); jsonErr == nil {
			log.Printf("直播列表快取命中: %s", cacheKey)
			return &resp, nil
		}
	}

	// 快取 miss，查 DB
	streams, total, err := s.repo.ListStreams(ctx, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("查詢直播列表失敗: %w", err)
	}

	// 組裝 response（觀眾視角，含拉流地址），並從 Hub 即時取得 viewer_count
	streamResponses := make([]model.StreamResponse, 0, len(streams))
	for _, stream := range streams {
		sr := s.toPlayerResponse(stream)
		// 如果有 ViewerCounter（Hub），用即時人數覆蓋 DB 的冗餘欄位
		if s.viewerCounter != nil {
			sr.ViewerCount = s.viewerCounter.ViewerCount(stream.ID)
		}
		streamResponses = append(streamResponses, sr)
	}

	// 依 viewer_count 降序重排（因為 DB 的 viewer_count 可能不即時）
	sort.Slice(streamResponses, func(i, j int) bool {
		return streamResponses[i].ViewerCount > streamResponses[j].ViewerCount
	})

	resp := &model.ListStreamsResponse{
		Streams: streamResponses,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}

	// 存入 Redis 快取
	if data, jsonErr := json.Marshal(resp); jsonErr == nil {
		if setErr := s.rdb.Set(ctx, cacheKey, data, listCacheTTL).Err(); setErr != nil {
			log.Printf("寫入直播列表快取失敗: %v", setErr)
		}
	}

	return resp, nil
}

// GetStream 取得直播間詳情，回傳含拉流地址的 response
func (s *StreamService) GetStream(ctx context.Context, id uuid.UUID) (*model.StreamResponse, error) {
	stream, err := s.repo.GetStreamByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}
		return nil, fmt.Errorf("查詢直播間失敗: %w", err)
	}

	resp := s.toPlayerResponse(stream)
	// 如果有 ViewerCounter（Hub），用即時人數覆蓋 DB 的冗餘欄位
	if s.viewerCounter != nil {
		resp.ViewerCount = s.viewerCounter.ViewerCount(stream.ID)
	}
	return &resp, nil
}

// UpdateStream 更新直播間標題/封面，需驗證擁有者
func (s *StreamService) UpdateStream(ctx context.Context, id uuid.UUID, userID uuid.UUID, req model.UpdateStreamRequest) (*model.StreamResponse, error) {
	// 驗證擁有者
	stream, err := s.repo.GetStreamByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}
		return nil, fmt.Errorf("查詢直播間失敗: %w", err)
	}

	if stream.UserID != userID {
		return nil, ErrStreamNotOwner
	}

	updated, err := s.repo.UpdateStream(ctx, id, req.Title, req.CoverURL)
	if err != nil {
		return nil, fmt.Errorf("更新直播間失敗: %w", err)
	}

	// 清除列表快取
	_ = s.InvalidateListCache(ctx)

	resp := s.toOwnerResponse(updated)
	return &resp, nil
}

// EndStream 結束直播，需驗證擁有者
func (s *StreamService) EndStream(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.StreamResponse, error) {
	// 驗證擁有者
	stream, err := s.repo.GetStreamByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return nil, ErrStreamNotFound
		}
		return nil, fmt.Errorf("查詢直播間失敗: %w", err)
	}

	if stream.UserID != userID {
		return nil, ErrStreamNotOwner
	}

	if stream.Status == model.StreamStatusEnded {
		return nil, ErrStreamAlreadyEnded
	}

	ended, err := s.repo.EndStream(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("結束直播失敗: %w", err)
	}

	// 清除列表快取
	_ = s.InvalidateListCache(ctx)

	resp := s.toOwnerResponse(ended)
	return &resp, nil
}

// HandlePublish SRS on_publish 回調：更新直播狀態為 live
func (s *StreamService) HandlePublish(ctx context.Context, streamKey string) error {
	stream, err := s.repo.GetStreamByStreamKey(ctx, streamKey)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return ErrStreamNotFound
		}
		return fmt.Errorf("查詢直播間失敗: %w", err)
	}

	_, err = s.repo.UpdateStreamStatus(ctx, stream.ID, model.StreamStatusLive)
	if err != nil {
		return fmt.Errorf("更新直播狀態失敗: %w", err)
	}

	log.Printf("直播開始推流: stream_key=%s, stream_id=%s", streamKey, stream.ID)

	// 清除列表快取
	_ = s.InvalidateListCache(ctx)

	return nil
}

// HandleUnpublish SRS on_unpublish 回調：更新直播狀態為 ended
func (s *StreamService) HandleUnpublish(ctx context.Context, streamKey string) error {
	stream, err := s.repo.GetStreamByStreamKey(ctx, streamKey)
	if err != nil {
		if errors.Is(err, repository.ErrStreamNotFound) {
			return ErrStreamNotFound
		}
		return fmt.Errorf("查詢直播間失敗: %w", err)
	}

	_, err = s.repo.UpdateStreamStatus(ctx, stream.ID, model.StreamStatusEnded)
	if err != nil {
		return fmt.Errorf("更新直播狀態失敗: %w", err)
	}

	log.Printf("直播停止推流: stream_key=%s, stream_id=%s", streamKey, stream.ID)

	// 清除列表快取
	_ = s.InvalidateListCache(ctx)

	return nil
}

// InvalidateListCache 清除 Redis 中所有 streams:list:* 快取
func (s *StreamService) InvalidateListCache(ctx context.Context) error {
	iter := s.rdb.Scan(ctx, 0, listCachePrefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		if err := s.rdb.Del(ctx, iter.Val()).Err(); err != nil {
			log.Printf("清除快取 key 失敗 %s: %v", iter.Val(), err)
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("掃描快取 key 失敗: %w", err)
	}
	return nil
}

// === 內部方法：組裝 response，使用 config 的 URL ===

// toOwnerResponse 給擁有者看的 response（含推流地址）
func (s *StreamService) toOwnerResponse(stream *model.Stream) model.StreamResponse {
	resp := stream.ToResponse()
	resp.StreamKey = stream.StreamKey
	resp.RTMPURL = s.cfg.SRSRtmpURL + "/" + stream.StreamKey
	return resp
}

// toPlayerResponse 給觀眾看的 response（含拉流地址）
func (s *StreamService) toPlayerResponse(stream *model.Stream) model.StreamResponse {
	resp := stream.ToResponse()
	resp.FLVURL = s.cfg.SRSHttpURL + "/" + stream.StreamKey + ".flv"
	resp.HLSURL = s.cfg.SRSHttpURL + "/" + stream.StreamKey + ".m3u8"
	return resp
}
