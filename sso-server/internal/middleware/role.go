package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
)

// RoleMiddleware creates role-based access control middleware
func RoleMiddleware(userRepo *repository.UserRepository, allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context (set by AuthMiddleware)
		userID := c.Locals("user_id")
		if userID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized - no user in context",
			})
		}

		userIDStr := userID.(string)

		// Fetch user with roles
		user, err := userRepo.GetByID(c.Context(), userIDStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized - user not found",
			})
		}

		// Load user roles
		var userRoles []models.Role
		if err := userRepo.GetDB().Model(user).Association("Roles").Find(&userRoles); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to load user roles",
			})
		}

		// Check if user has any of the allowed roles
		for _, userRole := range userRoles {
			for _, allowedRole := range allowedRoles {
				if userRole.Name == allowedRole {
					// Store user roles in context for handlers
					c.Locals("user_roles", userRoles)
					c.Locals("user", user)
					return c.Next()
				}
			}
		}

		// No matching role found
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden - insufficient permissions",
		})
	}
}
