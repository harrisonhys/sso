package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sso-project/sso-server/internal/database"
	"github.com/sso-project/sso-server/internal/models"
	"gorm.io/gorm"
)

// PermissionRepository handles permission data operations
type PermissionRepository struct {
	db *database.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *database.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// Create creates a new permission
func (r *PermissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	if permission.ID == "" {
		permission.ID = uuid.New().String()
	}

	result := r.db.WithContext(ctx).Create(permission)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrAlreadyExists
		}
		return result.Error
	}
	return nil
}

// GetByID retrieves a permission by ID
func (r *PermissionRepository) GetByID(ctx context.Context, id string) (*models.Permission, error) {
	var permission models.Permission
	result := r.db.WithContext(ctx).
		First(&permission, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &permission, nil
}

// GetByName retrieves a permission by name
func (r *PermissionRepository) GetByName(ctx context.Context, name string) (*models.Permission, error) {
	var permission models.Permission
	result := r.db.WithContext(ctx).
		First(&permission, "name = ?", name)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &permission, nil
}

// Update updates a permission
func (r *PermissionRepository) Update(ctx context.Context, permission *models.Permission) error {
	result := r.db.WithContext(ctx).Save(permission)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a permission
func (r *PermissionRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Permission{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves permissions with pagination
func (r *PermissionRepository) List(ctx context.Context, offset, limit int) ([]*models.Permission, int64, error) {
	var permissions []*models.Permission
	var total int64

	// Count total
	r.db.WithContext(ctx).Model(&models.Permission{}).Count(&total)

	// Get paginated results
	result := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("resource ASC, action ASC").
		Find(&permissions)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return permissions, total, nil
}

// GetDB returns the underlying GORM DB instance
func (r *PermissionRepository) GetDB() *gorm.DB {
	return r.db.DB
}
