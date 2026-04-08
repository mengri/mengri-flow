CREATE TABLE IF NOT EXISTS audit_events (
    id                VARCHAR(36)  PRIMARY KEY,
    actor_id          VARCHAR(36)  NOT NULL DEFAULT '',
    target_account_id VARCHAR(36)  NOT NULL DEFAULT '',
    event_type        VARCHAR(50)  NOT NULL,
    result            VARCHAR(10)  NOT NULL DEFAULT 'success',
    ip                VARCHAR(45)  NOT NULL DEFAULT '',
    ua                VARCHAR(500) NOT NULL DEFAULT '',
    metadata_json     TEXT,
    created_at        DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_target_account (target_account_id),
    KEY idx_event_type (event_type),
    KEY idx_created_at (created_at),
    KEY idx_actor (actor_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
