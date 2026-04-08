-- name: CreateAuditEvent :execresult
INSERT INTO audit_events (id, actor_id, target_account_id, event_type, result, ip, ua, metadata_json, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListAuditEventsByAccount :many
SELECT id, actor_id, target_account_id, event_type, result, ip, ua, metadata_json, created_at
FROM audit_events WHERE target_account_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: CountAuditEventsByAccount :one
SELECT COUNT(*) FROM audit_events WHERE target_account_id = ?;
