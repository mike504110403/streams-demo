-- 004_create_chat_messages.up.sql
-- Sprint 3: 建立 chat_messages 表（彈幕/聊天訊息）

CREATE TABLE IF NOT EXISTS chat_messages (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id  UUID         NOT NULL REFERENCES streams(id) ON DELETE CASCADE,
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content    TEXT         NOT NULL,
    type       VARCHAR(20)  NOT NULL DEFAULT 'chat',  -- 'chat' 一般彈幕, 'system' 系統訊息（XX進入直播間）, 'gift' 禮物訊息
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 索引：依直播間查詢訊息
CREATE INDEX idx_chat_messages_stream_id ON chat_messages(stream_id);

-- 索引：依時間排序查詢
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at);

-- 複合索引：查詢某直播間的訊息並按時間排序（最常用查詢）
CREATE INDEX idx_chat_messages_stream_created ON chat_messages(stream_id, created_at);
