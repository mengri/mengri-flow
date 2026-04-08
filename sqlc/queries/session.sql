-- name: CreateSession :execresult
INSERT INTO sessions (id, account_id, refresh_token_hash, device_info_json, ip, expires_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetSessionByRefreshTokenHash :one
SELECT id, account_id, refresh_token_hash, device_info_json, ip, expires_at, created_at
FROM sessions WHERE refresh_token_hash = ? AND expires_at > NOW();

-- name: DeleteSession :exec
DELETE FROM sessions WHERE refresh_token_hash = ?;

-- name: DeleteSessionsByAccountID :exec
DELETE FROM sessions WHERE account_id = ? AND refresh_token_hash != ?;

-- name: CountSessionsByAccountID :one
SELECT COUNT(*) FROM sessions WHERE account_id = ? AND expires_at > NOW();
