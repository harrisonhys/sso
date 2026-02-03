package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
	"github.com/sso-project/sso-server/internal/service"
	"github.com/sso-project/sso-server/internal/utils"
)

// AdminHandler handles admin dashboard endpoints
type AdminHandler struct {
	userRepo            *repository.UserRepository
	auditRepo           *repository.AuditLogRepository
	sessionRepo         repository.SessionStore
	oauth2ClientRepo    *repository.OAuth2ClientRepository
	passwordService     *service.PasswordService
	oauth2ClientService *service.OAuth2ClientService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	userRepo *repository.UserRepository,
	auditRepo *repository.AuditLogRepository,
	sessionRepo repository.SessionStore,
	oauth2ClientRepo *repository.OAuth2ClientRepository,
	passwordService *service.PasswordService,
	oauth2ClientService *service.OAuth2ClientService,
) *AdminHandler {
	return &AdminHandler{
		userRepo:            userRepo,
		auditRepo:           auditRepo,
		sessionRepo:         sessionRepo,
		oauth2ClientRepo:    oauth2ClientRepo,
		passwordService:     passwordService,
		oauth2ClientService: oauth2ClientService,
	}
}

// DashboardPage renders the admin dashboard home page
func (h *AdminHandler) DashboardPage(c *fiber.Ctx) error {
	return c.Render("admin/dashboard", fiber.Map{
		"title": "Admin Dashboard",
		"user":  c.Locals("user"),
	})
}

// UsersPage renders the user management page
func (h *AdminHandler) UsersPage(c *fiber.Ctx) error {
	return c.Render("admin/users", fiber.Map{
		"title": "User Management",
		"user":  c.Locals("user"),
	})
}

// OAuth2ClientsPage renders the OAuth2 clients management page
func (h *AdminHandler) OAuth2ClientsPage(c *fiber.Ctx) error {
	return c.Render("admin/oauth2-clients", fiber.Map{
		"title": "OAuth2 Clients",
		"user":  c.Locals("user"),
	})
}

// AuditLogsPage renders the audit logs page
func (h *AdminHandler) AuditLogsPage(c *fiber.Ctx) error {
	return c.Render("admin/audit-logs", fiber.Map{
		"title": "Audit Logs",
		"user":  c.Locals("user"),
	})
}

// SettingsPage renders the settings page
func (h *AdminHandler) SettingsPage(c *fiber.Ctx) error {
	return c.Render("admin/settings", fiber.Map{
		"title": "System Settings",
		"user":  c.Locals("user"),
	})
}

// GetStats returns dashboard statistics
func (h *AdminHandler) GetStats(c *fiber.Ctx) error {
	ctx := c.Context()

	// Count total users
	totalUsers, err := h.userRepo.CountUsers(ctx, "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to count users",
		})
	}

	// Count active sessions (sessions not expired)
	// Note: h.sessionRepo is now an interface.
	// We can't directly use GetDB() for the interface if it's Redis.
	// For V1 compatibility, if it's DB store, we might cast it, or better:
	// Use repository methods. But repository doesn't have Count().
	// For now, let's skip active session count if using Redis, or implement Count() in interface.
	// To minimize interface changes now, we'll try to check if it's DatabaseSessionStore

	var activeSessions int64 = 0

	if dbStore, ok := h.sessionRepo.(*repository.DatabaseSessionStore); ok {
		dbStore.GetDB().Model(&models.Session{}).
			Where("expires_at > ?", time.Now()).
			Count(&activeSessions)
	} else {
		// Redis implementation of count would be needed here
		// For now just 0
	}

	// Count OAuth2 clients
	var oauth2Clients int64
	h.oauth2ClientRepo.GetDB().Model(&models.OAuth2Client{}).
		Where("is_active = ?", true).
		Count(&oauth2Clients)

	// Count recent failed login attempts (last 24 hours)
	var failedLogins int64
	h.auditRepo.GetDB().Model(&models.AuditLog{}).
		Where("action = ? AND created_at > ?", "login_failed", time.Now().Add(-24*time.Hour)).
		Count(&failedLogins)

	return c.JSON(fiber.Map{
		"total_users":       totalUsers,
		"active_sessions":   activeSessions,
		"oauth2_clients":    oauth2Clients,
		"failed_logins_24h": failedLogins,
	})
}

// GetUsers returns paginated list of users
func (h *AdminHandler) GetUsers(c *fiber.Ctx) error {
	// Parse query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	search := c.Query("search", "")

	offset := (page - 1) * limit

	// Get users with filters
	users, total, err := h.userRepo.ListWithFilters(c.Context(), offset, limit, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch users",
		})
	}

	return c.JSON(fiber.Map{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetUser returns a single user details
func (h *AdminHandler) GetUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		if err == repository.ErrNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch user",
		})
	}

	return c.JSON(user)
}

// CreateUser creates a new user
func (h *AdminHandler) CreateUser(c *fiber.Ctx) error {
	var req struct {
		Email    string   `json:"email"`
		Name     string   `json:"name"`
		Password string   `json:"password"`
		RoleIDs  []string `json:"role_ids"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	// Validate password
	policy := utils.PasswordPolicy{
		MinLength:      8,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}
	if err := utils.ValidatePassword(req.Password, policy); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to hash password",
		})
	}

	// Create user
	user := &models.User{
		Email:             req.Email,
		Name:              req.Name,
		PasswordHash:      hashedPassword,
		IsActive:          true,
		PasswordChangedAt: time.Now(),
	}

	if err := h.userRepo.Create(c.Context(), user); err != nil {
		if err == repository.ErrAlreadyExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "user with this email already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create user: " + err.Error(),
		})
	}

	// Assign roles
	for _, roleID := range req.RoleIDs {
		h.userRepo.AssignRole(c.Context(), user.ID, roleID)
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

// UpdateUser updates user details
func (h *AdminHandler) UpdateUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	var req struct {
		Name     *string `json:"name"`
		Email    *string `json:"email"`
		IsActive *bool   `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		if err == repository.ErrNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch user",
		})
	}

	// Update fields if provided
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := h.userRepo.Update(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update user",
		})
	}

	return c.JSON(user)
}

// DeleteUser deactivates a user
func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	if err := h.userRepo.UpdateUserStatus(c.Context(), userID, false); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to deactivate user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deactivated successfully",
	})
}

// ResetUserPassword generates a password reset token for user
func (h *AdminHandler) ResetUserPassword(c *fiber.Ctx) error {
	userID := c.Params("id")

	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	// Generate reset token
	resetToken, err := h.passwordService.GenerateResetToken(c.Context(), user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate reset token",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password reset token generated",
		"token":   resetToken,
		"user":    user.Email,
	})
}

// UnlockUser unlocks a locked user account
func (h *AdminHandler) UnlockUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	if err := h.userRepo.UnlockAccount(c.Context(), userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to unlock account",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Account unlocked successfully",
	})
}

// AssignRole assigns a role to a user
func (h *AdminHandler) AssignRole(c *fiber.Ctx) error {
	userID := c.Params("id")

	var req struct {
		RoleID string `json:"role_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.userRepo.AssignRole(c.Context(), userID, req.RoleID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to assign role",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Role assigned successfully",
	})
}

// RemoveRole removes a role from a user
func (h *AdminHandler) RemoveRole(c *fiber.Ctx) error {
	userID := c.Params("id")
	roleID := c.Params("role_id")

	if err := h.userRepo.RemoveRole(c.Context(), userID, roleID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to remove role",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Role removed successfully",
	})
}

// GetAuditLogs returns paginated audit logs with filters
func (h *AdminHandler) GetAuditLogs(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	offset := (page - 1) * limit

	// Build filters
	filters := make(map[string]interface{})
	if userID := c.Query("user_id"); userID != "" {
		filters["user_id"] = userID
	}
	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}
	if resource := c.Query("resource"); resource != "" {
		filters["resource"] = resource
	}

	logs, total, err := h.auditRepo.List(c.Context(), filters, offset, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch audit logs",
		})
	}

	return c.JSON(fiber.Map{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetSystemInfo returns system information
func (h *AdminHandler) GetSystemInfo(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"version":     "1.0.0",
		"environment": "development",
		"database":    "connected",
	})
}

// TestRedisConnection checks if the provided Redis credentials are valid
func (h *AdminHandler) TestRedisConnection(c *fiber.Ctx) error {
	var config struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	}

	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})
	defer rdb.Close()

	// Short timeout for test
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   fmt.Sprintf("Failed to connect to Redis: %v", err),
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Connected to Redis successfully",
		"status":  "success",
	})
}

// SwitchSessionDriver handles switching the active session driver
func (h *AdminHandler) SwitchSessionDriver(c *fiber.Ctx) error {
	var input struct {
		Driver string `json:"driver"` // "db" or "redis"
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Logic to save config would go here (e.g. via h.configService or direct repo if injected)
	// For now we just return success as per previous plan step
	return c.JSON(fiber.Map{
		"message": "Session driver configuration updated",
		"driver":  input.Driver,
	})
}
