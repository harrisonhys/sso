package models

import "time"

// ==================== OAuth2 Models ====================

// OAuth2Client represents a registered OAuth2 client application
type OAuth2Client struct {
	ID            string      `gorm:"column:id;primaryKey" json:"id"`
	ClientID      string      `gorm:"column:client_id;uniqueIndex;type:varchar(255)" json:"client_id"`
	ClientSecret  string      `gorm:"column:client_secret_hash" json:"-"` // Never expose in JSON
	Name          string      `gorm:"column:name" json:"name"`
	Description   string      `gorm:"column:description" json:"description"`
	RedirectURIs  StringSlice `gorm:"column:redirect_uris;type:json" json:"redirect_uris"`
	AllowedScopes StringSlice `gorm:"column:allowed_scopes;type:json" json:"allowed_scopes"`
	GrantTypes    StringSlice `gorm:"column:grant_types;type:json" json:"grant_types"`
	IsPublic      bool        `gorm:"column:is_public;default:false" json:"is_public"`
	IsActive      bool        `gorm:"column:is_active;default:true" json:"is_active"`
	OwnerUserID   *string     `gorm:"column:owner_user_id;type:char(36)" json:"owner_user_id,omitempty"`
	CreatedAt     time.Time   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time   `gorm:"column:updated_at" json:"updated_at"`
}

func (OAuth2Client) TableName() string {
	return "oauth2_clients"
}

// OAuth2Scope represents an OAuth2 permission scope
type OAuth2Scope struct {
	ID          string    `gorm:"column:id;primaryKey" json:"id"`
	Name        string    `gorm:"column:name;uniqueIndex;type:varchar(255)" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	IsDefault   bool      `gorm:"column:is_default;default:false" json:"is_default"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

func (OAuth2Scope) TableName() string {
	return "oauth2_scopes"
}

// OAuth2AuthorizationCode represents an authorization code in the OAuth2 flow
type OAuth2AuthorizationCode struct {
	ID                  string      `gorm:"column:id;primaryKey" json:"id"`
	Code                string      `gorm:"column:code;uniqueIndex;type:varchar(255)" json:"code"`
	ClientID            string      `gorm:"column:client_id;type:varchar(255)" json:"client_id"`
	UserID              string      `gorm:"column:user_id;type:char(36)" json:"user_id"`
	RedirectURI         string      `gorm:"column:redirect_uri" json:"redirect_uri"`
	Scopes              StringSlice `gorm:"column:scopes;type:json" json:"scopes"`
	CodeChallenge       *string     `gorm:"column:code_challenge" json:"code_challenge,omitempty"`
	CodeChallengeMethod *string     `gorm:"column:code_challenge_method" json:"code_challenge_method,omitempty"`
	ExpiresAt           time.Time   `gorm:"column:expires_at" json:"expires_at"`
	Used                bool        `gorm:"column:used;default:false" json:"used"`
	CreatedAt           time.Time   `gorm:"column:created_at" json:"created_at"`
}

func (OAuth2AuthorizationCode) TableName() string {
	return "oauth2_authorization_codes"
}

// OAuth2AccessToken represents an OAuth2 access token
type OAuth2AccessToken struct {
	ID        string      `gorm:"column:id;primaryKey" json:"id"`
	TokenHash string      `gorm:"column:token_hash;uniqueIndex;type:varchar(255)" json:"-"` // Store hash, not actual token
	ClientID  string      `gorm:"column:client_id;type:varchar(255)" json:"client_id"`
	UserID    *string     `gorm:"column:user_id;type:char(36)" json:"user_id,omitempty"` // NULL for client_credentials
	Scopes    StringSlice `gorm:"column:scopes;type:json" json:"scopes"`
	ExpiresAt time.Time   `gorm:"column:expires_at" json:"expires_at"`
	CreatedAt time.Time   `gorm:"column:created_at" json:"created_at"`
}

func (OAuth2AccessToken) TableName() string {
	return "oauth2_access_tokens"
}

// OAuth2RefreshToken represents an OAuth2 refresh token
type OAuth2RefreshToken struct {
	ID            string      `gorm:"column:id;primaryKey" json:"id"`
	Token         string      `gorm:"column:token;uniqueIndex;type:varchar(255)" json:"-"` // Never expose in JSON
	AccessTokenID *string     `gorm:"column:access_token_id;type:char(36)" json:"access_token_id,omitempty"`
	ClientID      string      `gorm:"column:client_id;type:varchar(255)" json:"client_id"`
	UserID        string      `gorm:"column:user_id;type:char(36)" json:"user_id"`
	Scopes        StringSlice `gorm:"column:scopes;type:json" json:"scopes"`
	ExpiresAt     time.Time   `gorm:"column:expires_at" json:"expires_at"`
	Revoked       bool        `gorm:"column:revoked;default:false" json:"revoked"`
	CreatedAt     time.Time   `gorm:"column:created_at" json:"created_at"`
}

func (OAuth2RefreshToken) TableName() string {
	return "oauth2_refresh_tokens"
}

// OAuth2Consent represents a user's consent to an OAuth2 client
type OAuth2Consent struct {
	ID        string      `gorm:"column:id;primaryKey" json:"id"`
	UserID    string      `gorm:"column:user_id;type:char(36)" json:"user_id"`
	ClientID  string      `gorm:"column:client_id;type:varchar(255)" json:"client_id"`
	Scopes    StringSlice `gorm:"column:scopes;type:json" json:"scopes"`
	GrantedAt time.Time   `gorm:"column:granted_at" json:"granted_at"`
	UpdatedAt time.Time   `gorm:"column:updated_at" json:"updated_at"`
}

func (OAuth2Consent) TableName() string {
	return "oauth2_consents"
}
