package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/streams-demo/backend/internal/model"
)

var (
	ErrStreamNotFound = errors.New("stream not found")
)

// StreamRepository 直播間資料存取層
type StreamRepository struct {
	db *pgxpool.Pool
}

// NewStreamRepository 建立 StreamRepository
func NewStreamRepository(db *pgxpool.Pool) *StreamRepository {
	return &StreamRepository{db: db}
}

// streamColumns 統一的 stream 查詢欄位
const streamColumns = `id, user_id, title, cover_url, stream_key, status, started_at, ended_at, viewer_count, created_at, updated_at`

// scanStream 統一掃描 stream row
func scanStream(row pgx.Row) (*model.Stream, error) {
	s := &model.Stream{}
	err := row.Scan(
		&s.ID, &s.UserID, &s.Title, &s.CoverURL,
		&s.StreamKey, &s.Status, &s.StartedAt, &s.EndedAt,
		&s.ViewerCount, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStreamNotFound
		}
		return nil, err
	}
	return s, nil
}

// CreateStream 建立直播間
func (r *StreamRepository) CreateStream(ctx context.Context, userID uuid.UUID, title string, coverURL *string, streamKey string) (*model.Stream, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO streams (user_id, title, cover_url, stream_key)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+streamColumns,
		userID, title, coverURL, streamKey,
	)
	return scanStream(row)
}

// GetStreamByID 用 ID 查詢直播間
func (r *StreamRepository) GetStreamByID(ctx context.Context, id uuid.UUID) (*model.Stream, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+streamColumns+` FROM streams WHERE id = $1`, id,
	)
	return scanStream(row)
}

// ListStreams 分頁查詢直播列表，按 viewer_count DESC, created_at DESC 排序
func (r *StreamRepository) ListStreams(ctx context.Context, offset, limit int) ([]*model.Stream, int, error) {
	// 查詢總數
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM streams WHERE status != 'ended'`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查詢列表
	rows, err := r.db.Query(ctx,
		`SELECT `+streamColumns+` FROM streams
		 WHERE status != 'ended'
		 ORDER BY viewer_count DESC, created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var streams []*model.Stream
	for rows.Next() {
		s := &model.Stream{}
		err := rows.Scan(
			&s.ID, &s.UserID, &s.Title, &s.CoverURL,
			&s.StreamKey, &s.Status, &s.StartedAt, &s.EndedAt,
			&s.ViewerCount, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		streams = append(streams, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return streams, total, nil
}

// UpdateStream 更新直播間標題和封面
func (r *StreamRepository) UpdateStream(ctx context.Context, id uuid.UUID, title *string, coverURL *string) (*model.Stream, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE streams
		 SET title     = COALESCE($2, title),
		     cover_url = COALESCE($3, cover_url)
		 WHERE id = $1
		 RETURNING `+streamColumns,
		id, title, coverURL,
	)
	return scanStream(row)
}

// EndStream 結束直播（更新狀態為 ended 並記錄結束時間）
func (r *StreamRepository) EndStream(ctx context.Context, id uuid.UUID) (*model.Stream, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE streams
		 SET status = 'ended', ended_at = NOW()
		 WHERE id = $1
		 RETURNING `+streamColumns,
		id,
	)
	return scanStream(row)
}

// GetStreamByStreamKey 用 stream_key 查詢直播間（給 SRS callback 用）
func (r *StreamRepository) GetStreamByStreamKey(ctx context.Context, streamKey string) (*model.Stream, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+streamColumns+` FROM streams WHERE stream_key = $1`, streamKey,
	)
	return scanStream(row)
}

// UpdateStreamStatus 更新直播狀態
func (r *StreamRepository) UpdateStreamStatus(ctx context.Context, id uuid.UUID, status model.StreamStatus) (*model.Stream, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE streams
		 SET status = $2, started_at = CASE WHEN $2 = 'live' THEN NOW() ELSE started_at END,
		     ended_at = CASE WHEN $2 = 'ended' THEN NOW() ELSE ended_at END
		 WHERE id = $1
		 RETURNING `+streamColumns,
		id, string(status),
	)
	return scanStream(row)
}

// UpdateViewerCount 更新 DB 中的觀看人數
func (r *StreamRepository) UpdateViewerCount(ctx context.Context, id uuid.UUID, count int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE streams SET viewer_count = $2 WHERE id = $1`,
		id, count,
	)
	return err
}
