-- 002_create_ppt_records.sql
-- Creates ppt_records table for managing per-user PPT metadata.

CREATE TABLE IF NOT EXISTS ppt_records (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(120) NOT NULL,
    description VARCHAR(500) NULL,
    group_name VARCHAR(120) NOT NULL,
    relative_path VARCHAR(255) NOT NULL,
    canonical_path VARCHAR(512) NOT NULL,
    tags JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_ppt_records_user FOREIGN KEY (user_id) REFERENCES user_accounts(id) ON DELETE CASCADE,
    CONSTRAINT uq_ppt_records_user_group UNIQUE (user_id, group_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_ppt_records_user_created_at ON ppt_records(user_id, created_at DESC);
