package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sso-project/sso-server/internal/database"
	"github.com/sso-project/sso-server/internal/models"
	"gorm.io/gorm"
)

// RoleRepository handles role data operations
type RoleRepository struct {
	db *database.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *database.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create creates a new role
func (r *RoleRepository) Create(ctx context.Context, role *models.Role) error {
	if role.ID == "" {
		role.ID = uuid.New().String()
	}

	result := r.db.WithContext(ctx).Create(role)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrAlreadyExists
		}
		return result.Error
	}
	return nil
}

// GetByID retrieves a role by ID
func (r *RoleRepository) GetByID(ctx context.Context, id string) (*models.Role, error) {
	var role models.Role
	result := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &role, nil
}

// GetByName retrieves a role by name
func (r *RoleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	result := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "name = ?", name)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &role, nil
}

// Update updates a role
func (r *RoleRepository) Update(ctx context.Context, role *models.Role) error {
	result := r.db.WithContext(ctx).Save(role)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a role
func (r *RoleRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.Role{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves roles with pagination
func (r *RoleRepository) List(ctx context.Context, offset, limit int) ([]*models.Role, int64, error) {
	var roles []*models.Role
	var total int64

	// Count total
	r.db.WithContext(ctx).Model(&models.Role{}).Count(&total)

	// Get paginated results
	result := r.db.WithContext(ctx).
		Preload("Permissions").
		Offset(offset).
		Limit(limit).
		Order("name ASC").
		Find(&roles)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return roles, total, nil
}

// AddPermission adds a permission to a role
func (r *RoleRepository) AddPermission(ctx context.Context, roleID, permissionID string) error {
	return r.db.WithContext(ctx).Exec(
		"INSERT IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)",
		roleID, permissionID,
	).Error
}

// RemovePermission removes a permission from a role
func (r *RoleRepository) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	return r.db.WithContext(ctx).Exec(
		"DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?",
		roleID, permissionID,
	).Error
}

// GetDB returns the underlying GORM DB instance
func (r *RoleRepository) GetDB() *gorm.DB {
	return r.db.DB
}
