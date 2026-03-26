-- name: GetUserByID :one
SELECT id, username, email, password, status, created_at, updated_at
FROM users
WHERE id = ?;

-- name: GetUserByEmail :one
SELECT id, username, email, password, status, created_at, updated_at
FROM users
WHERE email = ?;

-- name: ListUsers :many
SELECT id, username, email, password, status, created_at, updated_at
FROM users
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: CountUsers :one
SELECT COUNT(*) as total FROM users;

-- name: CreateUser :execresult
INSERT INTO users (username, email, password, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateUser :exec
UPDATE users
SET username = ?, email = ?, status = ?, updated_at = ?
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;
