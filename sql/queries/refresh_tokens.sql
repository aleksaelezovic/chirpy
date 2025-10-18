-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, user_id, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT users.* FROM users JOIN refresh_tokens ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.revoked_at IS NOT NULL
  AND refresh_tokens.token = $1
  AND refresh_tokens.expires_at > NOW()
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW() WHERE token = $1;
