CREATE TABLE IF NOT EXISTS activation_tokens (
    token_hash  VARCHAR(64)  PRIMARY KEY,
    account_id  VARCHAR(36)  NOT NULL,
    expires_at  DATETIME(3)  NOT NULL,
    used_at     DATETIME(3)  NULL,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_account_id (account_id),
    KEY idx_expires_at (expires_at),
    CONSTRAINT fk_activation_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
