-- Create oauth2_access_tokens table
CREATE TABLE IF NOT EXISTS oauth2_access_tokens (
    id CHAR(36) PRIMARY KEY,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    client_id VARCHAR(255) NOT NULL,
    user_id CHAR(36),
    scopes JSON,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_token_hash (token_hash),
    INDEX idx_client (client_id),
    INDEX idx_user (user_id),
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create oauth2_refresh_tokens table
CREATE TABLE IF NOT EXISTS oauth2_refresh_tokens (
    id CHAR(36) PRIMARY KEY,
    token VARCHAR(255) NOT NULL UNIQUE,
    access_token_id CHAR(36),
    client_id VARCHAR(255) NOT NULL,
    user_id CHAR(36) NOT NULL,
    scopes JSON,
    expires_at DATETIME NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_token (token),
    INDEX idx_access_token (access_token_id),
    INDEX idx_client_user (client_id, user_id),
    INDEX idx_expires (expires_at),
    INDEX idx_revoked (revoked)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
