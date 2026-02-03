-- Create oauth2_consents table
CREATE TABLE IF NOT EXISTS oauth2_consents (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    scopes JSON NOT NULL,
    granted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_user_client (user_id, client_id),
    INDEX idx_user_client (user_id, client_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
