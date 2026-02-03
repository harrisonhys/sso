package repository

import (
	"context"
	"errors"
	"time"

	"github.com/sso-project/sso-server/internal/models"
	"gorm.io/gorm"
)

var (
	ErrCodeNotFound = errors.New("authorization code not found")
	ErrCodeExpired  = errors.New("authorization code has expired")
	ErrCodeUsed     = errors.New("authorization code has already been used")
)

// OAuth2CodeRepository handles OAuth2 authorization code persistence
type OAuth2CodeRepository struct {
	db *gorm.DB
}

// NewOAuth2CodeRepository creates a new OAuth2CodeRepository
func NewOAuth2CodeRepository(db *gorm.DB) *OAuth2CodeRepository {
	return &OAuth2CodeRepository{db: db}
}

// Create stores a new authorization code
func (r *OAuth2CodeRepository) Create(ctx context.Context, code *models.OAuth2AuthorizationCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

// GetByCode retrieves an authorization code by its value
func (r *OAuth2CodeRepository) GetByCode(ctx context.Context, code string) (*models.OAuth2AuthorizationCode, error) {
	var authCode models.OAuth2AuthorizationCode
	err := r.db.WithContext(ctx).
		Where("code = ?", code).
		First(&authCode).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCodeNotFound
		}
		return nil, err
	}

	return &authCode, nil
}

// MarkAsUsed marks an authorization code as used
func (r *OAuth2CodeRepository) MarkAsUsed(ctx context.Context, code string) error {
	return r.db.WithContext(ctx).
		Model(&models.OAuth2AuthorizationCode{}).
		Where("code = ?", code).
		Update("used", true).Error
}

// DeleteExpired removes expired authorization codes
func (r *OAuth2CodeRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.OAuth2AuthorizationCode{}).Error
}

// ValidateCode validates an authorization code
func (r *OAuth2CodeRepository) ValidateCode(ctx context.Context, code, clientID, redirectURI string) (*models.OAuth2AuthorizationCode, error) {
	authCode, err := r.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Check if code has been used
	if authCode.Used {
		return nil, ErrCodeUsed
	}

	// Check if code has expired
	if time.Now().After(authCode.ExpiresAt) {
		return nil, ErrCodeExpired
	}

	// Validate client ID
	if authCode.ClientID != clientID {
		return nil, errors.New("client ID mismatch")
	}

	// Validate redirect URI
	if authCode.RedirectURI != redirectURI {
		return nil, errors.New("redirect URI mismatch")
	}

	return authCode, nil
}

// GetByClientAndUser retrieves authorization codes for a specific client and user
func (r *OAuth2CodeRepository) GetByClientAndUser(ctx context.Context, clientID, userID string) ([]*models.OAuth2AuthorizationCode, error) {
	var codes []*models.OAuth2AuthorizationCode
	err := r.db.WithContext(ctx).
		Where("client_id = ? AND user_id = ? AND used = ? AND expires_at > ?",
			clientID, userID, false, time.Now()).
		Order("created_at DESC").
		Find(&codes).Error

	return codes, err
}
