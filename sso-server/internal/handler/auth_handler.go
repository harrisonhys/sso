package handler

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sso-project/sso-server/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
	jwtService  *service.JWTService
	totpService *service.TOTPService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	authService *service.AuthService,
	jwtService *service.JWTService,
	totpService *service.TOTPService,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtService:  jwtService,
		totpService: totpService,
	}
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Success           bool        `json:"success"`
	RequiresTwoFactor bool        `json:"requires_two_factor,omitempty"`
	TempToken         string      `json:"temp_token,omitempty"`
	AccessToken       string      `json:"access_token,omitempty"`
	RefreshToken      string      `json:"refresh_token,omitempty"`
	SessionToken      string      `json:"session_token,omitempty"`
	User              interface{} `json:"user,omitempty"`
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[HANDLER_DEBUG] Failed to parse body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	log.Printf("[HANDLER_DEBUG] Login request for email: %s", req.Email)
	log.Printf("[HANDLER_DEBUG] Password length: %d", len(req.Password))

	// Get client info
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	log.Printf("[HANDLER_DEBUG] IP: %s, UserAgent: %s", ipAddress, userAgent)

	// Attempt login
	log.Printf("[HANDLER_DEBUG] Calling authService.Login...")
	result, err := h.authService.Login(c.Context(), req.Email, req.Password, ipAddress, userAgent)
	if err != nil {
		log.Printf("[HANDLER_DEBUG] Login failed with error: %v (type: %T)", err, err)
		switch err {
		case service.ErrInvalidCredentials:
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid email or password",
			})
		case service.ErrAccountLocked:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Account is locked. Please try again later.",
			})
		case service.ErrAccountInactive:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Account is inactive",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Login failed",
			})
		}
	}

	log.Printf("[HANDLER_DEBUG] Login successful, requires2FA: %v", result.RequiresTwoFactor)

	// If 2FA is required
	if result.RequiresTwoFactor {
		return c.JSON(LoginResponse{
			Success:           true,
			RequiresTwoFactor: true,
			TempToken:         result.TempToken,
		})
	}

	// Generate JWT tokens
	accessToken, err := h.jwtService.GenerateAccessToken(result.User)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(result.User.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate refresh token",
		})
	}

	// Set session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    result.Session.SessionToken,
		Expires:  result.Session.ExpiresAt,
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "Lax",
	})

	return c.JSON(LoginResponse{
		Success:      true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionToken: result.Session.SessionToken,
		User: fiber.Map{
			"id":    result.User.ID,
			"email": result.User.Email,
			"name":  result.User.Name,
		},
	})
}

// Verify2FARequest represents 2FA verification request
type Verify2FARequest struct {
	TempToken string `json:"temp_token" validate:"required"`
	Code      string `json:"code" validate:"required"`
}

// Verify2FA handles POST /auth/verify-2fa
func (h *AuthHandler) Verify2FA(c *fiber.Ctx) error {
	var req Verify2FARequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get user ID and temp session from temp token
	// The temp token is stored as a regular session but with short expiry
	session, err := h.authService.GetSessionByToken(c.Context(), req.TempToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired temporary token",
		})
	}

	// Verify the session is a 2FA pending session (within expiry time)
	if time.Now().After(session.ExpiresAt) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Temporary token has expired. Please login again",
		})
	}

	// Get user
	user := &session.User
	if user.TwoFactorAuth == nil || !user.TwoFactorAuth.Enabled {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "2FA is not enabled for this account",
		})
	}

	// Verify TOTP code
	isValid, err := h.totpService.VerifyCode(c.Context(), user.ID, req.Code)
	if err != nil || !isValid {
		// TODO: Increment failed 2FA attempts and lock after max attempts
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid verification code",
		})
	}

	// Delete the temporary session
	h.authService.DeleteSessionByToken(c.Context(), req.TempToken)

	// Get client info
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	// Create a real session now that 2FA is verified
	realSession, err := h.authService.CreateSessionAfter2FA(c.Context(), user.ID, ipAddress, userAgent)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create session",
		})
	}

	// Generate JWT tokens
	accessToken, err := h.jwtService.GenerateAccessToken(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate access token",
		})
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate refresh token",
		})
	}

	// Set session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    realSession.SessionToken,
		Expires:  realSession.ExpiresAt,
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "Lax",
	})

	// Log successful 2FA verification
	h.authService.LogAudit(c.Context(), &user.ID, "2fa_verified", "authentication", ipAddress, userAgent, "")

	return c.JSON(fiber.Map{
		"success":       true,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"session_token": realSession.SessionToken,
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Get session token from cookie or header
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		sessionToken = c.Get("Authorization")
		// Remove "Bearer " prefix if present
		if len(sessionToken) > 7 && sessionToken[:7] == "Bearer " {
			sessionToken = sessionToken[7:]
		}
	}

	if sessionToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No session token provided",
		})
	}

	// Get client info
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")

	// Logout
	if err := h.authService.Logout(c.Context(), sessionToken, ipAddress, userAgent); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Logout failed",
		})
	}

	// Clear cookie
	c.ClearCookie("session_token")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Logged out successfully",
	})
}

// RefreshSession handles POST /auth/refresh
func (h *AuthHandler) RefreshSession(c *fiber.Ctx) error {
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "No session token provided",
		})
	}

	session, err := h.authService.RefreshSession(c.Context(), sessionToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Failed to refresh session",
		})
	}

	// Update cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    session.SessionToken,
		Expires:  session.ExpiresAt,
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	return c.JSON(fiber.Map{
		"success":    true,
		"message":    "Session refreshed",
		"expires_at": session.ExpiresAt,
	})
}
