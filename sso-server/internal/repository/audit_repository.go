package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sso-project/sso-server/internal/database"
	"github.com/sso-project/sso-server/internal/models"
	"gorm.io/gorm"
)

// AuditLogRepository handles audit log operations
type AuditLogRepository struct {
	db *database.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *database.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log entry
func (r *AuditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(log).Error
}

// List retrieves audit logs with pagination and filters
func (r *AuditLogRepository) List(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]*models.AuditLog, int64, error) {
	var logs []*models.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AuditLog{})

	// Apply filters
	if userID, ok := filters["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}
	if action, ok := filters["action"]; ok {
		query = query.Where("action = ?", action)
	}
	if resource, ok := filters["resource"]; ok {
		query = query.Where("resource = ?", resource)
	}

	// Count total
	query.Count(&total)

	// Get paginated results
	result := query.
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return logs, total, nil
}

// DeleteOld deletes audit logs older than specified days
func (r *AuditLogRepository) DeleteOld(ctx context.Context, days int) error {
	return r.db.WithContext(ctx).
		Where("created_at < DATE_SUB(NOW(), INTERVAL ? DAY)", days).
		Delete(&models.AuditLog{}).Error
}

// GetDB returns the underlying GORM DB instance
func (r *AuditLogRepository) GetDB() *gorm.DB {
	return r.db.DB
}
