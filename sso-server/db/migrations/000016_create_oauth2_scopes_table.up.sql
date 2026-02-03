-- Create oauth2_scopes table
CREATE TABLE IF NOT EXISTS oauth2_scopes (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_name (name),
    INDEX idx_default (is_default)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default scopes
INSERT INTO oauth2_scopes (id, name, description, is_default) VALUES
    (UUID(), 'openid', 'OpenID Connect authentication', true),
    (UUID(), 'profile', 'Access user profile information (name, picture, etc.)', false),
    (UUID(), 'email', 'Access user email address', false),
    (UUID(), 'offline_access', 'Access resources offline (grants refresh token)', false);
