package service

import (
	"context"
	"errors"

	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
)

// RoleService handles business logic for roles
type RoleService struct {
	roleRepo       *repository.RoleRepository
	permissionRepo *repository.PermissionRepository
}

// NewRoleService creates a new role service
func NewRoleService(roleRepo *repository.RoleRepository, permissionRepo *repository.PermissionRepository) *RoleService {
	return &RoleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(ctx context.Context, name, description string) (*models.Role, error) {
	if name == "" {
		return nil, errors.New("role name is required")
	}

	role := &models.Role{
		Name:        name,
		Description: description,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

// GetRole retrieves a role by ID
func (s *RoleService) GetRole(ctx context.Context, id string) (*models.Role, error) {
	return s.roleRepo.GetByID(ctx, id)
}

// UpdateRole updates a role
func (s *RoleService) UpdateRole(ctx context.Context, id, name, description string) (*models.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		role.Name = name
	}
	role.Description = description

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

// DeleteRole deletes a role
func (s *RoleService) DeleteRole(ctx context.Context, id string) error {
	return s.roleRepo.Delete(ctx, id)
}

// ListRoles retrieves roles with pagination
func (s *RoleService) ListRoles(ctx context.Context, page, limit int) ([]*models.Role, int64, error) {
	offset := (page - 1) * limit
	return s.roleRepo.List(ctx, offset, limit)
}

// AssignPermissions assigns permissions to a role
func (s *RoleService) AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	// Verify role exists
	if _, err := s.roleRepo.GetByID(ctx, roleID); err != nil {
		return err
	}

	// Verify permissions exist and assign
	for _, permID := range permissionIDs {
		if _, err := s.permissionRepo.GetByID(ctx, permID); err != nil {
			return err
		}
		if err := s.roleRepo.AddPermission(ctx, roleID, permID); err != nil {
			return err
		}
	}

	return nil
}

// RemovePermission removes a permission from a role
func (s *RoleService) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	return s.roleRepo.RemovePermission(ctx, roleID, permissionID)
}
