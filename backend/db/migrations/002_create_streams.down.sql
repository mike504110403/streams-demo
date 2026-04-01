-- 002_create_streams.down.sql
-- Sprint 2: 回滾 streams 表

DROP TRIGGER IF EXISTS trigger_streams_updated_at ON streams;
DROP INDEX IF EXISTS idx_streams_user_id_created_at;
DROP INDEX IF EXISTS idx_streams_status_created_at;
DROP TABLE IF EXISTS streams;
