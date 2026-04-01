-- name: InsertChatMessage :one
INSERT INTO chat_messages (stream_id, user_id, content, type)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetChatMessagesByStream :many
SELECT cm.*, u.nickname, u.avatar_url
FROM chat_messages cm
JOIN users u ON u.id = cm.user_id
WHERE cm.stream_id = $1
ORDER BY cm.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetRecentChatMessages :many
SELECT cm.*, u.nickname, u.avatar_url
FROM chat_messages cm
JOIN users u ON u.id = cm.user_id
WHERE cm.stream_id = $1
ORDER BY cm.created_at DESC
LIMIT $2;
