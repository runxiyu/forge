-- name: GetUserCreds :one
SELECT id, COALESCE(password_hash, '') FROM users WHERE username = $1;

-- name: InsertSession :exec
INSERT INTO sessions (user_id, token_hash, expires_at) VALUES ($1, $2, $3);

-- name: GetUserFromSession :one
SELECT user_id, COALESCE(username, '') FROM users u JOIN sessions s ON u.id = s.user_id WHERE s.token_hash = $1;
