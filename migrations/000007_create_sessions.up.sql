CREATE TABLE IF NOT EXISTS sessions (
    id                 VARCHAR(36) PRIMARY KEY,
    account_id         VARCHAR(36) NOT NULL,
    refresh_token_hash VARCHAR(64) NOT NULL,
    device_info_json   TEXT,
    ip                 VARCHAR(45) NOT NULL DEFAULT '',
    expires_at         DATETIME(3) NOT NULL,
    created_at         DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_account_id (account_id),
    KEY idx_refresh_token (refresh_token_hash),
    KEY idx_expires_at (expires_at),
    CONSTRAINT fk_session_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
