# SSO System - Database Design

## 1. Entity Relationship Diagram

```mermaid
erDiagram
    USERS ||--o{ USER_ROLES : has
    USERS ||--o| TWO_FACTOR_AUTH : has
    USERS ||--o{ SESSIONS : has
    USERS ||--o{ PASSWORD_RESET_TOKENS : has
    USERS ||--o{ PASSWORD_HISTORY : has
    USERS ||--o{ AUDIT_LOGS : creates
    
    ROLES ||--o{ USER_ROLES : has
    ROLES ||--o{ ROLE_PERMISSIONS : has
    
    PERMISSIONS ||--o{ ROLE_PERMISSIONS : has
    
    OAUTH_CLIENTS ||--o{ OAUTH_AUTHORIZATION_CODES : has
    OAUTH_CLIENTS ||--o{ OAUTH_REFRESH_TOKENS : has
    
    USERS ||--o{ OAUTH_AUTHORIZATION_CODES : has
    USERS ||--o{ OAUTH_REFRESH_TOKENS : has
    
    USERS {
        uuid id PK
        string email UK
        string password_hash
        string name
        boolean email_verified
        boolean is_active
        boolean is_locked
        int failed_login_attempts
        datetime locked_until
        datetime password_changed_at
        datetime last_login_at
        datetime created_at
        datetime updated_at
    }
    
    ROLES {
        uuid id PK
        string name UK
        string description
        datetime created_at
        datetime updated_at
    }
    
    PERMISSIONS {
        uuid id PK
        string name UK
        string description
        string resource
        string action
        datetime created_at
        datetime updated_at
    }
    
    USER_ROLES {
        uuid user_id FK
        uuid role_id FK
        datetime created_at
    }
    
    ROLE_PERMISSIONS {
        uuid role_id FK
        uuid permission_id FK
        datetime created_at
    }
    
    OAUTH_CLIENTS {
        uuid id PK
        string client_id UK
        string client_secret
        string name
        string redirect_uris
        string allowed_scopes
        boolean is_active
        datetime created_at
        datetime updated_at
    }
    
    OAUTH_AUTHORIZATION_CODES {
        uuid id PK
        string code UK
        uuid user_id FK
        uuid client_id FK
        string redirect_uri
        string scope
        datetime expires_at
        boolean used
        datetime created_at
    }
    
    OAUTH_REFRESH_TOKENS {
        uuid id PK
        string token UK
        uuid user_id FK
        uuid client_id FK
        string scope
        datetime expires_at
        boolean revoked
        datetime created_at
    }
    
    SESSIONS {
        uuid id PK
        uuid user_id FK
        string session_token UK
        string ip_address
        string user_agent
        datetime expires_at
        datetime last_activity_at
        datetime created_at
    }
    
    TWO_FACTOR_AUTH {
        uuid id PK
        uuid user_id FK UK
        string secret_encrypted
        boolean enabled
        string backup_codes_encrypted
        datetime enabled_at
        datetime created_at
        datetime updated_at
    }
    
    PASSWORD_RESET_TOKENS {
        uuid id PK
        uuid user_id FK
        string token UK
        datetime expires_at
        boolean used
        datetime created_at
    }
    
    PASSWORD_HISTORY {
        uuid id PK
        uuid user_id FK
        string password_hash
        datetime created_at
    }
    
    AUDIT_LOGS {
        uuid id PK
        uuid user_id FK
        string action
        string resource
        string ip_address
        string user_agent
        text details
        datetime created_at
    }
    
    SYSTEM_CONFIG {
        uuid id PK
        string config_key UK
        text config_value
        datetime updated_at
        datetime created_at
    }
```

## 2. Table Schemas

### 2.1 Users Table

```sql
CREATE TABLE users (
    id CHAR(36) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    is_locked BOOLEAN DEFAULT FALSE,
    failed_login_attempts INT DEFAULT 0,
    locked_until DATETIME NULL,
    password_changed_at DATETIME NOT NULL,
    last_login_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_email (email),
    INDEX idx_is_active (is_active),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Column Descriptions:**
- `id`: UUID primary key
- `email`: Unique email address for login
- `password_hash`: bcrypt hashed password
- `name`: Display name
- `email_verified`: Email verification status
- `is_active`: Account active status
- `is_locked`: Account locked status (after failed attempts)
- `failed_login_attempts`: Counter for failed login attempts
- `locked_until`: Automatic unlock timestamp
- `password_changed_at`: Last password change (for expiry check)
- `last_login_at`: Last successful login timestamp

### 2.2 Roles Table

```sql
CREATE TABLE roles (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Default Roles:**
- `super_admin`: Full system access
- `admin`: Management dashboard access
- `user`: Basic user access

### 2.3 Permissions Table

```sql
CREATE TABLE permissions (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_resource_action (resource, action),
    INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Permission Format:** `resource:action`
- Examples: `users:read`, `users:write`, `roles:manage`, `config:write`

### 2.4 User Roles Table (Junction)

```sql
CREATE TABLE user_roles (
    user_id CHAR(36) NOT NULL,
    role_id CHAR(36) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    
    INDEX idx_user_id (user_id),
    INDEX idx_role_id (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 2.5 Role Permissions Table (Junction)

```sql
CREATE TABLE role_permissions (
    role_id CHAR(36) NOT NULL,
    permission_id CHAR(36) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    
    INDEX idx_role_id (role_id),
    INDEX idx_permission_id (permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 2.6 OAuth Clients Table

```sql
CREATE TABLE oauth_clients (
    id CHAR(36) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL UNIQUE,
    client_secret VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    redirect_uris TEXT NOT NULL,
    allowed_scopes VARCHAR(500) DEFAULT 'openid,profile,email',
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_client_id (client_id),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Column Descriptions:**
- `client_id`: Public client identifier
- `client_secret`: Hashed secret for confidential clients
- `redirect_uris`: JSON array of allowed redirect URIs
- `allowed_scopes`: Comma-separated list of allowed OAuth scopes

### 2.7 OAuth Authorization Codes Table

```sql
CREATE TABLE oauth_authorization_codes (
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
```

**Notes:**
- Authorization codes expire in 10 minutes
- Single-use only (marked as used after exchange)
- Automatically cleaned up after expiry

### 2.8 OAuth Refresh Tokens Table

```sql
CREATE TABLE oauth_refresh_tokens (
    id CHAR(36) PRIMARY KEY,
    token VARCHAR(255) NOT NULL UNIQUE,
    user_id CHAR(36) NOT NULL,
    client_id CHAR(36) NOT NULL,
    scope VARCHAR(500) NOT NULL,
    expires_at DATETIME NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES oauth_clients(id) ON DELETE CASCADE,
    
    INDEX idx_token (token),
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Notes:**
- Refresh tokens expire in 7 days
- Can be revoked manually
- Token rotation on refresh

### 2.9 Sessions Table

```sql
CREATE TABLE sessions (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    session_token VARCHAR(255) NOT NULL UNIQUE,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    expires_at DATETIME NOT NULL,
    last_activity_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_session_token (session_token),
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Notes:**
- Session timeout configurable (default: 30 minutes)
- Sliding expiry on activity
- Supports multiple concurrent sessions per user

### 2.10 Two Factor Auth Table

```sql
CREATE TABLE two_factor_auth (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL UNIQUE,
    secret_encrypted VARCHAR(500) NOT NULL,
    enabled BOOLEAN DEFAULT FALSE,
    backup_codes_encrypted TEXT,
    enabled_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Column Descriptions:**
- `secret_encrypted`: AES-256 encrypted TOTP secret
- `enabled`: 2FA activation status
- `backup_codes_encrypted`: Encrypted JSON array of backup codes
- `enabled_at`: Timestamp when 2FA was first enabled

### 2.11 Password Reset Tokens Table

```sql
CREATE TABLE password_reset_tokens (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_token (token),
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Notes:**
- Tokens expire in 1 hour
- Single-use only
- Invalidated on password change

### 2.12 Password History Table

```sql
CREATE TABLE password_history (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_user_id_created (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Notes:**
- Keep last 5 passwords per user
- Prevent password reuse
- Automatically prune old entries

### 2.13 Audit Logs Table

```sql
CREATE TABLE audit_logs (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NULL,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    details TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at),
    INDEX idx_resource (resource)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Logged Actions:**
- `login_success`, `login_failed`
- `logout`
- `password_changed`, `password_reset`
- `2fa_enabled`, `2fa_disabled`
- `user_created`, `user_updated`, `user_deleted`
- `role_assigned`, `role_revoked`
- `config_changed`

### 2.14 System Config Table

```sql
CREATE TABLE system_config (
    id CHAR(36) PRIMARY KEY,
    config_key VARCHAR(100) NOT NULL UNIQUE,
    config_value TEXT NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Configuration Keys:**
```json
{
  "password_policy": {
    "min_length": 8,
    "require_uppercase": true,
    "require_lowercase": true,
    "require_number": true,
    "require_special": true,
    "prevent_reuse_count": 5,
    "expiry_days": 90,
    "warning_days": 7
  },
  "session": {
    "timeout_minutes": 30,
    "max_concurrent_sessions": 5
  },
  "security": {
    "max_login_attempts": 5,
    "lockout_duration_minutes": 30,
    "enforce_2fa": false,
    "password_reset_token_expiry_hours": 1
  },
  "rate_limit": {
    "login_per_minute": 5,
    "api_per_minute": 100
  }
}
```

## 3. Indexes Strategy

### 3.1 Primary Keys
- All tables use UUID as primary key for better distribution

### 3.2 Foreign Keys
- All foreign keys have corresponding indexes
- CASCADE delete for dependent data
- SET NULL for audit logs

### 3.3 Query Optimization Indexes
- **Login Query**: `idx_email` on users table
- **Permission Check**: `idx_user_id` on user_roles, `idx_role_id` on role_permissions
- **Token Validation**: `idx_token` on refresh_tokens and reset_tokens
- **Session Lookup**: `idx_session_token` on sessions
- **Audit Trail**: `idx_user_id`, `idx_created_at` on audit_logs

### 3.4 Composite Indexes
- `idx_user_id_created` on password_history for efficient history lookup
- `idx_resource_action` on permissions for fast permission matching

## 4. Sample Data

### 4.1 Default Roles

```sql
INSERT INTO roles (id, name, description) VALUES
(UUID(), 'super_admin', 'Full system access including system configuration'),
(UUID(), 'admin', 'Management dashboard access for user and role management'),
(UUID(), 'user', 'Basic user access to client applications');
```

### 4.2 Default Permissions

```sql
INSERT INTO permissions (id, name, description, resource, action) VALUES
(UUID(), 'users:read', 'Read user data', 'users', 'read'),
(UUID(), 'users:write', 'Create and update users', 'users', 'write'),
(UUID(), 'users:delete', 'Delete users', 'users', 'delete'),
(UUID(), 'roles:read', 'Read roles', 'roles', 'read'),
(UUID(), 'roles:write', 'Create and update roles', 'roles', 'write'),
(UUID(), 'roles:delete', 'Delete roles', 'roles', 'delete'),
(UUID(), 'permissions:read', 'Read permissions', 'permissions', 'read'),
(UUID(), 'permissions:write', 'Manage permissions', 'permissions', 'write'),
(UUID(), 'config:read', 'Read system configuration', 'config', 'read'),
(UUID(), 'config:write', 'Update system configuration', 'config', 'write'),
(UUID(), 'audit:read', 'Read audit logs', 'audit', 'read'),
(UUID(), 'clients:read', 'Read OAuth clients', 'clients', 'read'),
(UUID(), 'clients:write', 'Manage OAuth clients', 'clients', 'write');
```

### 4.3 Default System Configuration

```sql
INSERT INTO system_config (id, config_key, config_value) VALUES
(UUID(), 'password_policy', '{"min_length":8,"require_uppercase":true,"require_lowercase":true,"require_number":true,"require_special":true,"prevent_reuse_count":5,"expiry_days":90,"warning_days":7}'),
(UUID(), 'session', '{"timeout_minutes":30,"max_concurrent_sessions":5}'),
(UUID(), 'security', '{"max_login_attempts":5,"lockout_duration_minutes":30,"enforce_2fa":false,"password_reset_token_expiry_hours":1}'),
(UUID(), 'rate_limit', '{"login_per_minute":5,"api_per_minute":100}');
```

## 5. Database Maintenance

### 5.1 Cleanup Jobs (Scheduled)

**Daily Cleanup:**
```sql
-- Clean expired authorization codes (older than 1 day)
DELETE FROM oauth_authorization_codes 
WHERE expires_at < DATE_SUB(NOW(), INTERVAL 1 DAY);

-- Clean expired password reset tokens
DELETE FROM password_reset_tokens 
WHERE expires_at < NOW();

-- Clean expired sessions
DELETE FROM sessions 
WHERE expires_at < NOW();
```

**Weekly Cleanup:**
```sql
-- Clean old audit logs (keep 90 days)
DELETE FROM audit_logs 
WHERE created_at < DATE_SUB(NOW(), INTERVAL 90 DAY);

-- Clean old password history (keep last 5 per user)
DELETE ph1 FROM password_history ph1
LEFT JOIN (
    SELECT id FROM password_history ph2
    WHERE ph2.user_id = ph1.user_id
    ORDER BY created_at DESC
    LIMIT 5
) ph3 ON ph1.id = ph3.id
WHERE ph3.id IS NULL;
```

### 5.2 Backup Strategy

**Full Backup:**
- Daily at 2 AM
- Retention: 30 days
- Automated to cloud storage

**Incremental Backup:**
- Every 6 hours
- Retention: 7 days

**Point-in-Time Recovery:**
- Binary log enabled
- 7 days retention

### 5.3 Performance Monitoring

**Query Analysis:**
- Slow query log enabled (>1s)
- Regular EXPLAIN analysis
- Index optimization

**Table Statistics:**
```sql
-- Check table sizes
SELECT 
    table_name,
    ROUND(((data_length + index_length) / 1024 / 1024), 2) AS "Size (MB)",
    table_rows
FROM information_schema.TABLES
WHERE table_schema = 'sso_db'
ORDER BY (data_length + index_length) DESC;
```

## 6. Migration Strategy

### 6.1 Version Control
- Use migration tool (e.g., golang-migrate, goose)
- Sequential numbered migrations
- Up and down migrations
- Atomic transactions

### 6.2 Deployment Process
1. Backup database
2. Run migrations
3. Verify schema
4. Run tests
5. Rollback if errors

### 6.3 Zero-Downtime Migrations
- Additive changes first
- Backward compatible
- Feature flags for new columns
- Deprecate before delete

---

**Document Version**: 1.0  
**Last Updated**: 2026-01-26  
**Database Version**: MySQL 8.0+
