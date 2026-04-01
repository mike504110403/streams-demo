-- 002_create_streams.up.sql
-- Sprint 2: 建立 streams 表

CREATE TABLE IF NOT EXISTS streams (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID         NOT NULL REFERENCES users(id),
    title         VARCHAR(100) NOT NULL,
    cover_url     VARCHAR(500),
    stream_key    VARCHAR(100) UNIQUE NOT NULL,
    status        VARCHAR(20)  NOT NULL DEFAULT 'pending',
    started_at    TIMESTAMPTZ,
    ended_at      TIMESTAMPTZ,
    viewer_count  INT          NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 索引：大廳列表查詢（按狀態篩選，按建立時間排序）
CREATE INDEX idx_streams_status_created_at ON streams (status, created_at DESC);

-- 索引：查詢某用戶的直播歷史
CREATE INDEX idx_streams_user_id_created_at ON streams (user_id, created_at DESC);

-- 複用 users 表的 update_updated_at_column() function，自動更新 updated_at
CREATE TRIGGER trigger_streams_updated_at
    BEFORE UPDATE ON streams
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
