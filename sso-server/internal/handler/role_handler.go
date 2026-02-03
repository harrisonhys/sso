package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sso-project/sso-server/internal/repository"
	"github.com/sso-project/sso-server/internal/service"
)

// RoleHandler handles role management endpoints
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(roleService *service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// GetRoles returns paginated list of roles
func (h *RoleHandler) GetRoles(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	roles, total, err := h.roleService.ListRoles(c.Context(), page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch roles",
		})
	}

	return c.JSON(fiber.Map{
		"roles": roles,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetRole returns a single role details
func (h *RoleHandler) GetRole(c *fiber.Ctx) error {
	roleID := c.Params("id")

	role, err := h.roleService.GetRole(c.Context(), roleID)
	if err != nil {
		if err == repository.ErrNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "role not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch role",
		})
	}

	return c.JSON(role)
}

// CreateRole creates a new role
func (h *RoleHandler) CreateRole(c *fiber.Ctx) error {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	role, err := h.roleService.CreateRole(c.Context(), req.Name, req.Description)
	if err != nil {
		if err == repository.ErrAlreadyExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "role already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(role)
}

// UpdateRole updates a role
func (h *RoleHandler) UpdateRole(c *fiber.Ctx) error {
	roleID := c.Params("id")
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	role, err := h.roleService.UpdateRole(c.Context(), roleID, req.Name, req.Description)
	if err != nil {
		if err == repository.ErrNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "role not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update role",
		})
	}

	return c.JSON(role)
}

// DeleteRole deletes a role
func (h *RoleHandler) DeleteRole(c *fiber.Ctx) error {
	roleID := c.Params("id")

	if err := h.roleService.DeleteRole(c.Context(), roleID); err != nil {
		if err == repository.ErrNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "role not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete role",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Role deleted successfully",
	})
}

// AssignPermissions assigns permissions to a role
func (h *RoleHandler) AssignPermissions(c *fiber.Ctx) error {
	roleID := c.Params("id")
	var req struct {
		PermissionIDs []string `json:"permission_ids"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.roleService.AssignPermissions(c.Context(), roleID, req.PermissionIDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to assign permissions",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Permissions assigned successfully",
	})
}

// RemovePermission removes a permission from a role
func (h *RoleHandler) RemovePermission(c *fiber.Ctx) error {
	roleID := c.Params("id")
	permissionID := c.Params("permission_id")

	if err := h.roleService.RemovePermission(c.Context(), roleID, permissionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to remove permission",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Permission removed successfully",
	})
}
