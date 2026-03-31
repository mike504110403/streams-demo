-- name: CreateUser :one
INSERT INTO users (email, phone, password_hash, nickname, avatar_url, bio)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByPhone :one
SELECT * FROM users
WHERE phone = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET nickname = $2, avatar_url = $3, bio = $4
WHERE id = $1
RETURNING *;
