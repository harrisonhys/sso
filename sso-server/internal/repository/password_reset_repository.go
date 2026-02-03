package repository

import (
	"context"
	"time"

	"github.com/sso-project/sso-server/internal/database"
	"github.com/sso-project/sso-server/internal/models"
)

// PasswordResetTokenRepository handles password reset token operations
type PasswordResetTokenRepository struct {
	db *database.DB
}

// NewPasswordResetTokenRepository creates a new password reset token repository
func NewPasswordResetTokenRepository(db *database.DB) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{db: db}
}

// Create creates a new password reset token
func (r *PasswordResetTokenRepository) Create(ctx context.Context, token *models.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// FindByToken finds a password reset token by token string
func (r *PasswordResetTokenRepository) FindByToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	var resetToken models.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND used = ? AND expires_at > ?", token, false, time.Now()).
		First(&resetToken).Error
	if err != nil {
		return nil, err
	}
	return &resetToken, nil
}

// FindByEmail finds the most recent unused token for an email
func (r *PasswordResetTokenRepository) FindByEmail(ctx context.Context, email string) (*models.PasswordResetToken, error) {
	var resetToken models.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("email = ? AND used = ? AND expires_at > ?", email, false, time.Now()).
		Order("created_at DESC").
		First(&resetToken).Error
	if err != nil {
		return nil, err
	}
	return &resetToken, nil
}

// MarkAsUsed marks a token as used
func (r *PasswordResetTokenRepository) MarkAsUsed(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Model(&models.PasswordResetToken{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"used":    true,
			"used_at": time.Now(),
		}).Error
}

// DeleteExpired deletes all expired tokens
func (r *PasswordResetTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.PasswordResetToken{}).Error
}

// DeleteByEmail deletes all tokens for a specific email
func (r *PasswordResetTokenRepository) DeleteByEmail(ctx context.Context, email string) error {
	return r.db.WithContext(ctx).
		Where("email = ?", email).
		Delete(&models.PasswordResetToken{}).Error
}
