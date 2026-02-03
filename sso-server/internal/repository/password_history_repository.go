package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sso-project/sso-server/internal/database"
	"github.com/sso-project/sso-server/internal/models"
)

// PasswordHistoryRepository handles password history operations
type PasswordHistoryRepository struct {
	db *database.DB
}

// NewPasswordHistoryRepository creates a new password history repository
func NewPasswordHistoryRepository(db *database.DB) *PasswordHistoryRepository {
	return &PasswordHistoryRepository{db: db}
}

// Create adds a password to history
func (r *PasswordHistoryRepository) Create(ctx context.Context, userID, passwordHash string) error {
	history := &models.PasswordHistory{
		ID:           uuid.New().String(),
		UserID:       userID,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	return r.db.WithContext(ctx).Create(history).Error
}

// GetRecentPasswords fetches the last N passwords for a user
func (r *PasswordHistoryRepository) GetRecentPasswords(ctx context.Context, userID string, count int) ([]models.PasswordHistory, error) {
	var history []models.PasswordHistory
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(count).
		Find(&history).Error
	if err != nil {
		return nil, err
	}
	return history, nil
}

// CleanupOldPasswords deletes password history beyond the keep count
func (r *PasswordHistoryRepository) CleanupOldPasswords(ctx context.Context, userID string, keepCount int) error {
	// Get IDs of passwords to keep
	var idsToKeep []string
	err := r.db.WithContext(ctx).
		Model(&models.PasswordHistory{}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(keepCount).
		Pluck("id", &idsToKeep).Error
	if err != nil {
		return err
	}

	// Delete old passwords
	if len(idsToKeep) > 0 {
		return r.db.WithContext(ctx).
			Where("user_id = ? AND id NOT IN ?", userID, idsToKeep).
			Delete(&models.PasswordHistory{}).Error
	}

	// If no passwords to keep, delete all
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.PasswordHistory{}).Error
}

// DeleteByUserID deletes all password history for a user
func (r *PasswordHistoryRepository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.PasswordHistory{}).Error
}

// Count returns the total number of password history entries for a user
func (r *PasswordHistoryRepository) Count(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.PasswordHistory{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
