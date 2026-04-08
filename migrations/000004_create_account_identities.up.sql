CREATE TABLE IF NOT EXISTS account_identities (
    id          VARCHAR(36)  PRIMARY KEY,
    account_id  VARCHAR(36)  NOT NULL,
    login_type  VARCHAR(30)  NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    external_meta_json TEXT,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    deleted_at  DATETIME(3)  NULL,
    UNIQUE KEY uk_login_type_external_id (login_type, external_id),
    UNIQUE KEY uk_account_login_type (account_id, login_type),
    KEY idx_account_id (account_id),
    CONSTRAINT fk_identity_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
