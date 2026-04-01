-- 003_auth_revamp.up.sql
-- 認證系統改版：phone 改為必填、移除 password_hash、新增 verification_codes + user_oauth_providers

-- ============================================================
-- 1. 修改 users 表
-- ============================================================

-- MVP 階段無正式用戶，先清除可能有 NULL phone 的測試資料
UPDATE users SET phone = CONCAT('+886900', LPAD(FLOOR(RANDOM() * 1000000)::TEXT, 6, '0'))
WHERE phone IS NULL;

-- phone 改為 NOT NULL（主要認證欄位）
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;

-- email 改為 NULLABLE，保留 UNIQUE（PostgreSQL UNIQUE 允許多個 NULL）
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

-- 移除 password_hash 欄位（不再需要密碼認證）
ALTER TABLE users DROP COLUMN password_hash;

-- ============================================================
-- 2. 新增 verification_codes 表（簡訊驗證碼）
-- ============================================================

CREATE TABLE verification_codes (
    id         BIGSERIAL    PRIMARY KEY,
    phone      VARCHAR(20)  NOT NULL,
    code       VARCHAR(6)   NOT NULL,
    expires_at TIMESTAMPTZ  NOT NULL,
    used       BOOLEAN      DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_verification_codes_phone ON verification_codes(phone);
CREATE INDEX idx_verification_codes_expires_at ON verification_codes(expires_at);

-- ============================================================
-- 3. 新增 user_oauth_providers 表（第三方 OAuth 登入）
-- ============================================================

CREATE TABLE user_oauth_providers (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider         VARCHAR(20)  NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    email            VARCHAR(255),
    name             VARCHAR(100),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_user_oauth_provider ON user_oauth_providers(provider, provider_user_id);
CREATE INDEX idx_user_oauth_user_id ON user_oauth_providers(user_id);

-- 複用 users 表的 update_updated_at_column() function
CREATE TRIGGER trigger_user_oauth_providers_updated_at
    BEFORE UPDATE ON user_oauth_providers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
