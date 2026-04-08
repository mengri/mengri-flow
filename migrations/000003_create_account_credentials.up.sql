CREATE TABLE IF NOT EXISTS account_credentials (
    account_id       VARCHAR(36) PRIMARY KEY,
    password_hash    VARCHAR(255) NOT NULL,
    password_updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    CONSTRAINT fk_credentials_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
