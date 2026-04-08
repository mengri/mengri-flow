-- name: CreateCredential :execresult
INSERT INTO account_credentials (account_id, password_hash, password_updated_at) VALUES (?, ?, ?);

-- name: GetCredentialByAccountID :one
SELECT account_id, password_hash, password_updated_at
FROM account_credentials WHERE account_id = ?;

-- name: UpdateCredentialPassword :exec
UPDATE account_credentials SET password_hash = ?, password_updated_at = ? WHERE account_id = ?;
