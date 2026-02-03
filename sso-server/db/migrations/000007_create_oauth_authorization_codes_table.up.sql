-- Create oauth_authorization_codes table
CREATE TABLE IF NOT EXISTS oauth_authorization_codes (
    id CHAR(36) PRIMARY KEY,
    code VARCHAR(255) NOT NULL UNIQUE,
    user_id CHAR(36) NOT NULL,
    client_id CHAR(36) NOT NULL,
    redirect_uri VARCHAR(500) NOT NULL,
    scope VARCHAR(500) NOT NULL,
    expires_at DATETIME NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
   FOREIGN KEY (client_id) REFERENCES oauth_clients(id) ON DELETE CASCADE,
    
    INDEX idx_code (code),
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
