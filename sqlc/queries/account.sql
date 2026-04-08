-- name: GetAccountByID :one
SELECT id, email, username, display_name, status, role, activated_at, created_at, updated_at
FROM accounts WHERE id = ?;

-- name: GetAccountByEmail :one
SELECT id, email, username, display_name, status, role, activated_at, created_at, updated_at
FROM accounts WHERE email = ?;

-- name: GetAccountByUsername :one
SELECT id, email, username, display_name, status, role, activated_at, created_at, updated_at
FROM accounts WHERE username = ?;

-- name: CreateAccount :execresult
INSERT INTO accounts (id, email, username, display_name, status, role, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateAccountStatus :exec
UPDATE accounts SET status = ?, updated_at = ? WHERE id = ?;

-- name: UpdateAccountActivation :exec
UPDATE accounts SET status = 'ACTIVE', activated_at = ?, updated_at = ? WHERE id = ?;

-- name: ListAccounts :many
SELECT id, email, username, display_name, status, role, activated_at, created_at, updated_at
FROM accounts ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: CountAccounts :one
SELECT COUNT(*) FROM accounts;
