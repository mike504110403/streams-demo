package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ChatMessage 聊天訊息 DB model
type ChatMessage struct {
	ID        uuid.UUID `json:"id"`
	StreamID  uuid.UUID `json:"stream_id"`
	UserID    uuid.UUID `json:"user_id"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Nickname  string    `json:"nickname"`
	AvatarURL *string   `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatRepository 聊天訊息資料存取層
type ChatRepository struct {
	db *pgxpool.Pool
}

// NewChatRepository 建立 ChatRepository
func NewChatRepository(db *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{db: db}
}

// InsertMessage 儲存聊天訊息
func (r *ChatRepository) InsertMessage(ctx context.Context, streamID, userID uuid.UUID, content, msgType string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO chat_messages (stream_id, user_id, content, type) VALUES ($1, $2, $3, $4)`,
		streamID, userID, content, msgType,
	)
	return err
}

// GetRecentMessages 取得最近 N 筆聊天訊息
func (r *ChatRepository) GetRecentMessages(ctx context.Context, streamID uuid.UUID, limit int) ([]ChatMessage, error) {
	rows, err := r.db.Query(ctx,
		`SELECT cm.id, cm.stream_id, cm.user_id, cm.content, cm.type, u.nickname, u.avatar_url, cm.created_at
		 FROM chat_messages cm
		 JOIN users u ON u.id = cm.user_id
		 WHERE cm.stream_id = $1
		 ORDER BY cm.created_at DESC
		 LIMIT $2`,
		streamID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.StreamID, &m.UserID, &m.Content, &m.Type, &m.Nickname, &m.AvatarURL, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	// 反轉順序（DB 查出來是 DESC，前端要 ASC）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, rows.Err()
}

// GetMessagesByStreamID 取得指定直播間的歷史訊息（cursor 分頁）
// before 為 zero time 時不加 cursor 條件，limit+1 用來判斷 has_more
func (r *ChatRepository) GetMessagesByStreamID(ctx context.Context, streamID uuid.UUID, limit int, before time.Time) ([]ChatMessage, error) {
	var rows pgx.Rows
	var err error

	if before.IsZero() {
		rows, err = r.db.Query(ctx,
			`SELECT cm.id, cm.stream_id, cm.user_id, cm.content, cm.type, u.nickname, u.avatar_url, cm.created_at
			 FROM chat_messages cm
			 JOIN users u ON u.id = cm.user_id
			 WHERE cm.stream_id = $1
			 ORDER BY cm.created_at DESC
			 LIMIT $2`,
			streamID, limit+1,
		)
	} else {
		rows, err = r.db.Query(ctx,
			`SELECT cm.id, cm.stream_id, cm.user_id, cm.content, cm.type, u.nickname, u.avatar_url, cm.created_at
			 FROM chat_messages cm
			 JOIN users u ON u.id = cm.user_id
			 WHERE cm.stream_id = $1 AND cm.created_at < $2
			 ORDER BY cm.created_at DESC
			 LIMIT $3`,
			streamID, before, limit+1,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.StreamID, &m.UserID, &m.Content, &m.Type, &m.Nickname, &m.AvatarURL, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, rows.Err()
}
