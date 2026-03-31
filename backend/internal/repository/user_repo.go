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
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// UserRepository 用戶資料存取層
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository 建立 UserRepository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser 建立新用戶
func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash, nickname string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, nickname)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, phone, password_hash, nickname, avatar_url, bio, created_at, updated_at`,
		email, passwordHash, nickname,
	).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash,
		&user.Nickname, &user.AvatarURL, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		// 檢查是否為 unique violation (email 重複)
		if isDuplicateKeyError(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, err
	}
	return user, nil
}

// GetUserByEmail 用 email 查詢用戶
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, phone, password_hash, nickname, avatar_url, bio, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash,
		&user.Nickname, &user.AvatarURL, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetUserByID 用 ID 查詢用戶
func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, phone, password_hash, nickname, avatar_url, bio, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash,
		&user.Nickname, &user.AvatarURL, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// UpdateProfile 更新用戶個人資料
func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, nickname *string, avatarURL *string, bio *string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx,
		`UPDATE users
		 SET nickname   = COALESCE($2, nickname),
		     avatar_url = COALESCE($3, avatar_url),
		     bio        = COALESCE($4, bio)
		 WHERE id = $1
		 RETURNING id, email, phone, password_hash, nickname, avatar_url, bio, created_at, updated_at`,
		id, nickname, avatarURL, bio,
	).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash,
		&user.Nickname, &user.AvatarURL, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// isDuplicateKeyError 檢查是否為 PostgreSQL unique violation (23505)
func isDuplicateKeyError(err error) bool {
	// pgx 會回傳 *pgconn.PgError，檢查 SQLSTATE 23505
	if err == nil {
		return false
	}
	return contains(err.Error(), "23505") || contains(err.Error(), "duplicate key")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
