package service

import (
	"context"
	"errors"

	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
)

// PermissionService handles business logic for permissions
type PermissionService struct {
	permissionRepo *repository.PermissionRepository
}

// NewPermissionService creates a new permission service
func NewPermissionService(permissionRepo *repository.PermissionRepository) *PermissionService {
	return &PermissionService{
		permissionRepo: permissionRepo,
	}
}

// CreatePermission creates a new permission
func (s *PermissionService) CreatePermission(ctx context.Context, name, description, resource, action string) (*models.Permission, error) {
	if name == "" || resource == "" || action == "" {
		return nil, errors.New("name, resource, and action are required")
	}

	permission := &models.Permission{
		Name:        name,
		Description: description,
		Resource:    resource,
		Action:      action,
	}

	if err := s.permissionRepo.Create(ctx, permission); err != nil {
		return nil, err
	}

	return permission, nil
}

// GetPermission retrieves a permission by ID
func (s *PermissionService) GetPermission(ctx context.Context, id string) (*models.Permission, error) {
	return s.permissionRepo.GetByID(ctx, id)
}

// UpdatePermission updates a permission
func (s *PermissionService) UpdatePermission(ctx context.Context, id, name, description, resource, action string) (*models.Permission, error) {
	permission, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		permission.Name = name
	}
	permission.Description = description
	if resource != "" {
		permission.Resource = resource
	}
	if action != "" {
		permission.Action = action
	}

	if err := s.permissionRepo.Update(ctx, permission); err != nil {
		return nil, err
	}

	return permission, nil
}

// DeletePermission deletes a permission
func (s *PermissionService) DeletePermission(ctx context.Context, id string) error {
	return s.permissionRepo.Delete(ctx, id)
}

// ListPermissions retrieves permissions with pagination
func (s *PermissionService) ListPermissions(ctx context.Context, page, limit int) ([]*models.Permission, int64, error) {
	offset := (page - 1) * limit
	return s.permissionRepo.List(ctx, offset, limit)
}
