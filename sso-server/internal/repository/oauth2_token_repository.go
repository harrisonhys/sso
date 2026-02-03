package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/sso-project/sso-server/internal/models"
	"gorm.io/gorm"
)

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token has expired")
	ErrTokenRevoked  = errors.New("token has been revoked")
)

// OAuth2TokenRepository handles OAuth2 token persistence
type OAuth2TokenRepository struct {
	db *gorm.DB
}

// NewOAuth2TokenRepository creates a new OAuth2TokenRepository
func NewOAuth2TokenRepository(db *gorm.DB) *OAuth2TokenRepository {
	return &OAuth2TokenRepository{db: db}
}

// ========== Access Tokens ==========

// CreateAccessToken stores a new access token (stores hash, not actual token)
func (r *OAuth2TokenRepository) CreateAccessToken(ctx context.Context, token *models.OAuth2AccessToken) error {
	// Hash the token before storing
	if token.TokenHash == "" {
		return errors.New("token hash is required")
	}
	return r.db.WithContext(ctx).Create(token).Error
}

// GetAccessTokenByHash retrieves an access token by its hash
func (r *OAuth2TokenRepository) GetAccessTokenByHash(ctx context.Context, tokenHash string) (*models.OAuth2AccessToken, error) {
	var token models.OAuth2AccessToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&token).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	// Check expiry
	if time.Now().After(token.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	return &token, nil
}

// RevokeAccessToken invalidates an access token
func (r *OAuth2TokenRepository) RevokeAccessToken(ctx context.Context, tokenID string) error {
	return r.db.WithContext(ctx).
		Delete(&models.OAuth2AccessToken{}, "id = ?", tokenID).Error
}

// DeleteExpiredAccessTokens removes expired access tokens
func (r *OAuth2TokenRepository) DeleteExpiredAccessTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.OAuth2AccessToken{}).Error
}

// ========== Refresh Tokens ==========

// CreateRefreshToken stores a new refresh token
func (r *OAuth2TokenRepository) CreateRefreshToken(ctx context.Context, token *models.OAuth2RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetRefreshToken retrieves a refresh token by its value
func (r *OAuth2TokenRepository) GetRefreshToken(ctx context.Context, token string) (*models.OAuth2RefreshToken, error) {
	var refreshToken models.OAuth2RefreshToken
	err := r.db.WithContext(ctx).
		Where("token = ?", token).
		First(&refreshToken).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	// Check if revoked
	if refreshToken.Revoked {
		return nil, ErrTokenRevoked
	}

	// Check expiry
	if time.Now().After(refreshToken.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	return &refreshToken, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (r *OAuth2TokenRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Model(&models.OAuth2RefreshToken{}).
		Where("token = ?", token).
		Update("revoked", true).Error
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (r *OAuth2TokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&models.OAuth2RefreshToken{}).
		Where("user_id = ?", userID).
		Update("revoked", true).Error
}

// RevokeAllClientUserTokens revokes all refresh tokens for a client-user combination
func (r *OAuth2TokenRepository) RevokeAllClientUserTokens(ctx context.Context, clientID, userID string) error {
	return r.db.WithContext(ctx).
		Model(&models.OAuth2RefreshToken{}).
		Where("client_id = ? AND user_id = ?", clientID, userID).
		Update("revoked", true).Error
}

// DeleteExpiredRefreshTokens removes expired refresh tokens
func (r *OAuth2TokenRepository) DeleteExpiredRefreshTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.OAuth2RefreshToken{}).Error
}

// ========== Utility Functions ==========

// HashToken creates a SHA-256 hash of a token for storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
