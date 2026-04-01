-- name: CreateStream :one
INSERT INTO streams (user_id, title, cover_url, stream_key, status)
VALUES ($1, $2, $3, $4, 'pending')
RETURNING *;

-- name: GetStreamByID :one
SELECT * FROM streams
WHERE id = $1;

-- name: GetStreamByStreamKey :one
SELECT * FROM streams
WHERE stream_key = $1;

-- name: ListLiveStreams :many
SELECT * FROM streams
WHERE status = 'live'
ORDER BY viewer_count DESC
LIMIT $1 OFFSET $2;

-- name: UpdateStreamStatus :one
UPDATE streams
SET status = $2,
    started_at = CASE WHEN $2 = 'live' THEN NOW() ELSE started_at END
WHERE id = $1
RETURNING *;

-- name: UpdateStreamInfo :one
UPDATE streams
SET title = $2, cover_url = $3
WHERE id = $1
RETURNING *;

-- name: UpdateViewerCount :exec
UPDATE streams
SET viewer_count = $2
WHERE id = $1;

-- name: EndStream :one
UPDATE streams
SET status = 'ended', ended_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListStreamsByUser :many
SELECT * FROM streams
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
