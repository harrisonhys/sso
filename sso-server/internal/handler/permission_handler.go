package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sso-project/sso-server/internal/repository"
	"github.com/sso-project/sso-server/internal/service"
)

// PermissionHandler handles permission management endpoints
type PermissionHandler struct {
	permissionService *service.PermissionService
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(permissionService *service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permissionService: permissionService}
}

// GetPermissions returns paginated list of permissions
func (h *PermissionHandler) GetPermissions(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	permissions, total, err := h.permissionService.ListPermissions(c.Context(), page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch permissions",
		})
	}

	return c.JSON(fiber.Map{
		"permissions": permissions,
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

// GetPermission returns a single permission details
func (h *PermissionHandler) GetPermission(c *fiber.Ctx) error {
	id := c.Params("id")

	permission, err := h.permissionService.GetPermission(c.Context(), id)
	if err != nil {
		if err == repository.ErrNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "permission not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch permission",
		})
	}

	return c.JSON(permission)
}

// CreatePermission creates a new permission
func (h *PermissionHandler) CreatePermission(c *fiber.Ctx) error {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Resource    string `json:"resource"`
		Action      string `json:"action"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	permission, err := h.permissionService.CreatePermission(c.Context(), req.Name, req.Description, req.Resource, req.Action)
	if err != nil {
		if err == repository.ErrAlreadyExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "permission already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(permission)
}

// UpdatePermission updates a permission
func (h *PermissionHandler) UpdatePermission(c *fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Resource    string `json:"resource"`
		Action      string `json:"action"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	permission, err := h.permissionService.UpdatePermission(c.Context(), id, req.Name, req.Description, req.Resource, req.Action)
	if err != nil {
		if err == repository.ErrNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "permission not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update permission",
		})
	}

	return c.JSON(permission)
}

// DeletePermission deletes a permission
func (h *PermissionHandler) DeletePermission(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.permissionService.DeletePermission(c.Context(), id); err != nil {
		if err == repository.ErrNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "permission not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete permission",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Permission deleted successfully",
	})
}
