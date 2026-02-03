package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// StringSlice is a custom type for JSON array of strings
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// User represents a user in the system
type User struct {
	ID                  string     `gorm:"column:id;primaryKey;type:char(36)" json:"id"`
	Email               string     `gorm:"column:email;unique;not null;type:varchar(255)" json:"email"`
	PasswordHash        string     `gorm:"column:password_hash;not null" json:"-"`
	Name                string     `gorm:"column:name;not null" json:"name"`
	EmailVerified       bool       `gorm:"column:email_verified;default:false" json:"email_verified"`
	IsActive            bool       `gorm:"column:is_active;default:true" json:"is_active"`
	IsLocked            bool       `gorm:"column:is_locked;default:false" json:"is_locked"`
	FailedLoginAttempts int        `gorm:"column:failed_login_attempts;default:0" json:"-"`
	LockedUntil         *time.Time `gorm:"column:locked_until" json:"locked_until,omitempty"`
	PasswordChangedAt   time.Time  `gorm:"column:password_changed_at;not null" json:"password_changed_at"`
	LastLoginAt         *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
	CreatedAt           time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at" json:"updated_at"`

	// Relationships
	Roles         []Role         `gorm:"many2many:user_roles" json:"roles,omitempty"`
	TwoFactorAuth *TwoFactorAuth `gorm:"foreignKey:UserID" json:"two_factor_auth,omitempty"`
}

// TableName specifies the table name
func (User) TableName() string {
	return "users"
}

// Role represents a role in the system
type Role struct {
	ID          string    `gorm:"column:id;primaryKey;type:char(36)" json:"id"`
	Name        string    `gorm:"column:name;unique;not null;type:varchar(255)" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`

	// Relationships
	Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}

// TableName specifies the table name
func (Role) TableName() string {
	return "roles"
}

// Permission represents a permission in the system
type Permission struct {
	ID          string    `gorm:"column:id;primaryKey;type:char(36)" json:"id"`
	Name        string    `gorm:"column:name;unique;not null;type:varchar(255)" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	Resource    string    `gorm:"column:resource;not null;type:varchar(255)" json:"resource"`
	Action      string    `gorm:"column:action;not null;type:varchar(255)" json:"action"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName specifies the table name
func (Permission) TableName() string {
	return "permissions"
}

// Session represents a user session
type Session struct {
	ID             string    `gorm:"column:id;primaryKey;type:char(36)" json:"id"`
	UserID         string    `gorm:"column:user_id;not null;type:char(36)" json:"user_id"`
	SessionToken   string    `gorm:"column:session_token;unique;not null;type:varchar(255)" json:"session_token"`
	IPAddress      string    `gorm:"column:ip_address;not null" json:"ip_address"`
	UserAgent      string    `gorm:"column:user_agent" json:"user_agent"`
	ExpiresAt      time.Time `gorm:"column:expires_at;not null" json:"expires_at"`
	LastActivityAt time.Time `gorm:"column:last_activity_at;not null" json:"last_activity_at"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name
func (Session) TableName() string {
	return "sessions"
}

// TwoFactorAuth represents 2FA configuration for a user
type TwoFactorAuth struct {
	ID                   string     `gorm:"column:id;primaryKey;type:char(36)" json:"id"`
	UserID               string     `gorm:"column:user_id;unique;not null;type:char(36)" json:"user_id"`
	SecretEncrypted      string     `gorm:"column:secret_encrypted;not null" json:"-"`
	Enabled              bool       `gorm:"column:enabled;default:false" json:"enabled"`
	BackupCodesEncrypted string     `gorm:"column:backup_codes_encrypted" json:"-"`
	EnabledAt            *time.Time `gorm:"column:enabled_at" json:"enabled_at,omitempty"`
	CreatedAt            time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName specifies the table name
func (TwoFactorAuth) TableName() string {
	return "two_factor_auth"
}

// OAuthClient represents an OAuth2 client application
type OAuthClient struct {
	ID            string    `gorm:"column:id;primaryKey;type:char(36)" json:"id"`
	ClientID      string    `gorm:"column:client_id;unique;not null" json:"client_id"`
	ClientSecret  string    `gorm:"column:client_secret;not null" json:"-"`
	Name          string    `gorm:"column:name;not null" json:"name"`
	RedirectURIs  string    `gorm:"column:redirect_uris;not null" json:"redirect_uris"`
	AllowedScopes string    `gorm:"column:allowed_scopes" json:"allowed_scopes"`
	IsActive      bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName specifies the table name
func (OAuthClient) TableName() string {
	return "oauth_clients"
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        string     `gorm:"primaryKey" json:"id"`
	UserID    string     `gorm:"not null;index;type:char(36)" json:"user_id"`
	Email     string     `gorm:"not null;index;type:varchar(255)" json:"email"`
	Token     string     `gorm:"not null;unique;index;type:varchar(255)" json:"token"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	Used      bool       `gorm:"default:false" json:"used"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the table name
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// PasswordHistory represents password history for a user
type PasswordHistory struct {
	ID           string    `gorm:"column:id;primaryKey;type:char(36)" json:"id"`
	UserID       string    `gorm:"column:user_id;not null;type:char(36)" json:"user_id"`
	PasswordHash string    `gorm:"column:password_hash;not null" json:"-"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName specifies the table name
func (PasswordHistory) TableName() string {
	return "password_history"
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        string    `gorm:"column:id;primaryKey;type:char(36)" json:"id"`
	UserID    *string   `gorm:"column:user_id;type:char(36)" json:"user_id,omitempty"`
	Action    string    `gorm:"column:action;not null;type:varchar(255)" json:"action"`
	Resource  string    `gorm:"column:resource;not null;type:varchar(255)" json:"resource"`
	IPAddress string    `gorm:"column:ip_address;not null" json:"ip_address"`
	UserAgent string    `gorm:"column:user_agent" json:"user_agent"`
	Details   string    `gorm:"column:details" json:"details,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name
func (AuditLog) TableName() string {
	return "audit_logs"
}

// SystemConfig represents a system configuration entry
type SystemConfig struct {
	ID          string    `gorm:"column:id;primaryKey;type:char(36)" json:"id"`
	ConfigKey   string    `gorm:"column:config_key;unique;not null;type:varchar(255)" json:"config_key"`
	ConfigValue string    `gorm:"column:config_value;not null" json:"config_value"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName specifies the table name
func (SystemConfig) TableName() string {
	return "system_config"
}
