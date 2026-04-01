-- 003_auth_revamp.down.sql
-- 回滾認證系統改版

-- ============================================================
-- 1. 移除 user_oauth_providers 表
-- ============================================================

DROP TRIGGER IF EXISTS trigger_user_oauth_providers_updated_at ON user_oauth_providers;
DROP TABLE IF EXISTS user_oauth_providers;

-- ============================================================
-- 2. 移除 verification_codes 表
-- ============================================================

DROP TABLE IF EXISTS verification_codes;

-- ============================================================
-- 3. 還原 users 表
-- ============================================================

-- 加回 password_hash 欄位（預設空字串，因為原始資料已遺失）
ALTER TABLE users ADD COLUMN password_hash VARCHAR(255) NOT NULL DEFAULT '';

-- email 還原為 NOT NULL（先填入假資料避免違反約束）
UPDATE users SET email = CONCAT('user_', id::TEXT, '@placeholder.local')
WHERE email IS NULL;
ALTER TABLE users ALTER COLUMN email SET NOT NULL;

-- phone 還原為 NULLABLE
ALTER TABLE users ALTER COLUMN phone DROP NOT NULL;

-- 移除 password_hash 的預設值（原始 schema 沒有 DEFAULT）
ALTER TABLE users ALTER COLUMN password_hash DROP DEFAULT;
