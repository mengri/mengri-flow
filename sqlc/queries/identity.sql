-- name: CreateIdentity :execresult
INSERT INTO account_identities (id, account_id, login_type, external_id, external_meta_json, created_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetIdentityByProviderID :one
SELECT id, account_id, login_type, external_id, external_meta_json, created_at, deleted_at
FROM account_identities WHERE login_type = ? AND external_id = ? AND deleted_at IS NULL;

-- name: ListIdentitiesByAccountID :many
SELECT id, account_id, login_type, external_id, external_meta_json, created_at, deleted_at
FROM account_identities WHERE account_id = ? AND deleted_at IS NULL;

-- name: CountActiveIdentities :one
SELECT COUNT(*) FROM account_identities WHERE account_id = ? AND deleted_at IS NULL;

-- name: SoftDeleteIdentity :exec
UPDATE account_identities SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL;
