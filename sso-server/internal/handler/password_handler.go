package handler

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/sso-project/sso-server/internal/service"
)

// PasswordHandler handles password management endpoints
type PasswordHandler struct {
	passwordService *service.PasswordService
	emailService    *service.EmailService
	userRepo        interface {
		GetByEmail(ctx interface{}, email string) (interface{}, error)
	}
}

// NewPasswordHandler creates a new password handler
func NewPasswordHandler(
	passwordService *service.PasswordService,
	emailService *service.EmailService,
) *PasswordHandler {
	return &PasswordHandler{
		passwordService: passwordService,
		emailService:    emailService,
	}
}

// ForgotPasswordRequest represents forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents reset password request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ChangePasswordRequest represents change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// ForgotPassword handles POST /password/forgot
func (h *PasswordHandler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Generate reset token
	token, err := h.passwordService.GenerateResetToken(c.Context(), req.Email)
	if err != nil {
		log.Printf("Error generating reset token: %v", err)
		// Still return success to prevent email enumeration
	}

	// Send reset email (only if token was generated)
	if token != "" {
		// Note: In production, get user details to personalize email
		err = h.emailService.SendPasswordResetEmail(req.Email, "User", token)
		if err != nil {
			log.Printf("Error sending reset email: %v", err)
			// Don't fail request - token is still valid
		}
	}

	// Always return success to prevent email enumeration
	return c.JSON(fiber.Map{
		"success": true,
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles POST /password/reset
func (h *PasswordHandler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Reset password using token
	err := h.passwordService.ResetPassword(c.Context(), req.Token, req.NewPassword)
	if err != nil {
		switch err {
		case service.ErrTokenNotFound:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid or expired reset token",
			})
		case service.ErrTokenExpired:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Reset token has expired. Please request a new one",
			})
		case service.ErrTokenUsed:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Reset token has already been used",
			})
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Password has been reset successfully. Please login with your new password",
	})
}

// ChangePassword handles POST /password/change (requires authentication)
func (h *PasswordHandler) ChangePassword(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Change password
	err := h.passwordService.ChangePassword(c.Context(), userID.(string), req.CurrentPassword, req.NewPassword)
	if err != nil {
		if err.Error() == "current password is incorrect" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Current password is incorrect",
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Password changed successfully. All sessions have been terminated for security",
	})
}
