package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sso-project/sso-server/internal/service"
)

// AuthMiddleware creates auth middleware
func AuthMiddleware(sessionService *service.SessionService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get session token from cookie or Authorization header
		sessionToken := c.Cookies("session_token")
		if sessionToken == "" {
			authHeader := c.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				sessionToken = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if sessionToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized - No session token",
			})
		}

		// Validate session
		session, err := sessionService.ValidateSession(c.Context(), sessionToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized - Invalid or expired session",
			})
		}

		// Store session and user in context
		c.Locals("session", session)
		c.Locals("user_id", session.UserID)

		return c.Next()
	}
}
