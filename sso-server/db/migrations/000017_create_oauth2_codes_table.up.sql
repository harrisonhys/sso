-- Create oauth2_authorization_codes table
CREATE TABLE IF NOT EXISTS oauth2_authorization_codes (
    id CHAR(36) PRIMARY KEY,
    code VARCHAR(255) NOT NULL UNIQUE,
    client_id VARCHAR(255) NOT NULL,
    user_id CHAR(36) NOT NULL,
    redirect_uri VARCHAR(500) NOT NULL,
    scopes JSON,
    code_challenge VARCHAR(255),
    code_challenge_method VARCHAR(10),
    expires_at DATETIME NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_code (code),
    INDEX idx_client_user (client_id, user_id),
    INDEX idx_expires (expires_at),
    INDEX idx_used (used)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
