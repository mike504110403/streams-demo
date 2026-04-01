package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/streams-demo/backend/internal/model"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrPhoneAlreadyExists = errors.New("phone already exists")
	ErrOAuthNotFound      = errors.New("oauth provider not found")
)

// UserRepository 用戶資料存取層
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository 建立 UserRepository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// userColumns 統一的 user 查詢欄位（auth revamp 後無 password_hash）
const userColumns = `id, email, phone, nickname, avatar_url, bio, created_at, updated_at`

// scanUser 統一掃描 user row
func scanUser(row pgx.Row) (*model.User, error) {
	user := &model.User{}
	err := row.Scan(
		&user.ID, &user.Email, &user.Phone,
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

// CreateUserByPhone 用手機號碼建立新用戶
func (r *UserRepository) CreateUserByPhone(ctx context.Context, phone, nickname string) (*model.User, error) {
	row := r.db.QueryRow(ctx,
		`INSERT INTO users (phone, nickname)
		 VALUES ($1, $2)
		 RETURNING `+userColumns,
		phone, nickname,
	)
	user, err := scanUser(row)
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, ErrPhoneAlreadyExists
		}
		return nil, err
	}
	return user, nil
}

// CreateUserForOAuth 為 OAuth 用戶建立帳號（phone 為空字串，後續可綁定）
func (r *UserRepository) CreateUserForOAuth(ctx context.Context, nickname string, avatarURL *string) (*model.User, error) {
	// OAuth 用戶初始沒有手機號碼，用 UUID 佔位以滿足 NOT NULL + UNIQUE
	placeholder := "oauth_" + uuid.New().String()
	row := r.db.QueryRow(ctx,
		`INSERT INTO users (phone, nickname, avatar_url)
		 VALUES ($1, $2, $3)
		 RETURNING `+userColumns,
		placeholder, nickname, avatarURL,
	)
	return scanUser(row)
}

// GetUserByPhone 用手機號碼查詢用戶
func (r *UserRepository) GetUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE phone = $1`, phone,
	)
	return scanUser(row)
}

// GetUserByID 用 ID 查詢用戶
func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = $1`, id,
	)
	return scanUser(row)
}

// UpdateProfile 更新用戶個人資料
func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, nickname *string, avatarURL *string, bio *string) (*model.User, error) {
	row := r.db.QueryRow(ctx,
		`UPDATE users
		 SET nickname   = COALESCE($2, nickname),
		     avatar_url = COALESCE($3, avatar_url),
		     bio        = COALESCE($4, bio)
		 WHERE id = $1
		 RETURNING `+userColumns,
		id, nickname, avatarURL, bio,
	)
	return scanUser(row)
}

// === 驗證碼相關 ===

// SaveVerificationCode 儲存驗證碼
func (r *UserRepository) SaveVerificationCode(ctx context.Context, phone, code string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO verification_codes (phone, code, expires_at) VALUES ($1, $2, $3)`,
		phone, code, expiresAt,
	)
	return err
}

// GetLatestVerificationCode 取得最新未使用的驗證碼
func (r *UserRepository) GetLatestVerificationCode(ctx context.Context, phone string) (code string, expiresAt time.Time, used bool, err error) {
	err = r.db.QueryRow(ctx,
		`SELECT code, expires_at, used
		 FROM verification_codes
		 WHERE phone = $1
		 ORDER BY created_at DESC
		 LIMIT 1`,
		phone,
	).Scan(&code, &expiresAt, &used)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", time.Time{}, false, errors.New("no verification code found")
	}
	return
}

// GetLatestCodeCreatedAt 取得最新驗證碼的建立時間（用於 60 秒限制）
func (r *UserRepository) GetLatestCodeCreatedAt(ctx context.Context, phone string) (time.Time, error) {
	var createdAt time.Time
	err := r.db.QueryRow(ctx,
		`SELECT created_at FROM verification_codes
		 WHERE phone = $1
		 ORDER BY created_at DESC LIMIT 1`,
		phone,
	).Scan(&createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, nil
	}
	return createdAt, err
}

// MarkVerificationCodeUsed 標記驗證碼為已使用
func (r *UserRepository) MarkVerificationCodeUsed(ctx context.Context, phone, code string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE verification_codes SET used = TRUE
		 WHERE phone = $1 AND code = $2 AND used = FALSE`,
		phone, code,
	)
	return err
}

// === OAuth 相關 ===

// FindUserByOAuth 用 OAuth provider + provider_user_id 查詢用戶
func (r *UserRepository) FindUserByOAuth(ctx context.Context, provider, providerUserID string) (*model.User, error) {
	row := r.db.QueryRow(ctx,
		`SELECT u.`+userColumns+`
		 FROM users u
		 INNER JOIN user_oauth_providers p ON u.id = p.user_id
		 WHERE p.provider = $1 AND p.provider_user_id = $2`,
		provider, providerUserID,
	)
	// 手動 scan，因為有表別名但欄位一樣
	user := &model.User{}
	err := row.Scan(
		&user.ID, &user.Email, &user.Phone,
		&user.Nickname, &user.AvatarURL, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOAuthNotFound
		}
		return nil, err
	}
	return user, nil
}

// CreateOAuthProvider 建立 OAuth provider 記錄
func (r *UserRepository) CreateOAuthProvider(ctx context.Context, userID uuid.UUID, provider, providerUserID, email, name string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_oauth_providers (user_id, provider, provider_user_id, email, name)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID, provider, providerUserID, email, name,
	)
	return err
}

// isDuplicateKeyError 檢查是否為 PostgreSQL unique violation (23505)
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "23505") || contains(errStr, "duplicate key")
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
