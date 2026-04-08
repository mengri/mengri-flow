CREATE TABLE IF NOT EXISTS accounts (
    id          VARCHAR(36)  PRIMARY KEY,
    email       VARCHAR(255) NOT NULL,
    username    VARCHAR(50)  NOT NULL,
    display_name VARCHAR(100) NOT NULL DEFAULT '',
    status      VARCHAR(30)  NOT NULL DEFAULT 'PENDING_ACTIVATION',
    role        VARCHAR(20)  NOT NULL DEFAULT 'user',
    activated_at DATETIME(3) NULL,
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_email (email),
    UNIQUE KEY uk_username (username),
    KEY idx_status (status),
    KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
