-- name: CreateActivationToken :execresult
INSERT INTO activation_tokens (token_hash, account_id, expires_at, created_at) VALUES (?, ?, ?, ?);

-- name: GetActivationTokenByHash :one
SELECT token_hash, account_id, expires_at, used_at, created_at
FROM activation_tokens WHERE token_hash = ?;

-- name: MarkActivationTokenUsed :exec
UPDATE activation_tokens SET used_at = ? WHERE token_hash = ? AND used_at IS NULL;

-- name: InvalidateTokensByAccountID :exec
UPDATE activation_tokens SET used_at = NOW() WHERE account_id = ? AND used_at IS NULL;
