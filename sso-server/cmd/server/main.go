package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/sso-project/sso-server/internal/config"
	"github.com/sso-project/sso-server/internal/database"
	"github.com/sso-project/sso-server/internal/handler"
	appLogger "github.com/sso-project/sso-server/internal/logger"
	"github.com/sso-project/sso-server/internal/middleware"
	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
	"github.com/sso-project/sso-server/internal/service"
	"github.com/sso-project/sso-server/internal/utils"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLog := appLogger.New(cfg.Log.Level, cfg.Log.Format)
	defer appLog.Sync()

	appLog.Info("Starting SSO Server...")

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		appLog.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	appLog.Info("Database connected successfully")

	// Auto-migrate schema (for development)
	if err := db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Session{},
		&models.TwoFactorAuth{},
		&models.OAuthClient{},
		&models.PasswordResetToken{},
		&models.PasswordHistory{},
		&models.AuditLog{},
		&models.SystemConfig{},
		&models.OAuth2Client{},
		&models.OAuth2AuthorizationCode{},
		&models.OAuth2AccessToken{},
		&models.OAuth2RefreshToken{},
		&models.OAuth2Consent{},
		&models.OAuth2Scope{},
		// &models.OAuth2Scope{}, // Ensure this model exists if used
	); err != nil {
		appLog.Fatal("Failed to auto-migrate schema", "error", err)
	}
	appLog.Info("Database schema migrated successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Session Stores
	dbSessionStore := repository.NewDatabaseSessionStore(db)

	// Redis Client
	var redisClient *redis.Client
	if cfg.Redis.Host != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

		// Test Redis connection
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := redisClient.Ping(ctx).Err(); err != nil {
			appLog.Warn("Failed to connect to Redis", "error", err)
			redisClient = nil
		} else {
			appLog.Info("Redis connected successfully")
		}
		cancel()
	}

	var sessionStore repository.SessionStore = dbSessionStore

	// Logic to switch session store based on config (will be enhanced dynamically later)
	// For now we default to DB, but if Redis is configured we could use it?
	// User requirement is dynamic switch. For start, let's look at system_config.
	// ... we need to query system_config manually here or use configService.
	// For simplicity in main, we strictly start with DB or respecting ENV if provided.

	if redisClient != nil {
		appLog.Info("Redis available, but defaulting to Database session store until switched in Admin")
		// sessionStore = repository.NewRedisSessionStore(redisClient) // Uncomment to default to Redis if present
	}

	auditRepo := repository.NewAuditLogRepository(db)
	passwordResetRepo := repository.NewPasswordResetTokenRepository(db)
	passwordHistoryRepo := repository.NewPasswordHistoryRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	configRepo := repository.NewConfigRepository(db)

	appLog.Info("Repositories initialized")

	// Initialize email service
	emailService, err := service.NewEmailService(&service.EmailConfig{
		SMTPHost:     cfg.Email.SMTPHost,
		SMTPPort:     cfg.Email.SMTPPort,
		SMTPUser:     cfg.Email.SMTPUser,
		SMTPPassword: cfg.Email.SMTPPassword,
		FromEmail:    cfg.Email.FromEmail,
		FromName:     cfg.Email.FromName,
	})
	if err != nil {
		appLog.Warn("Failed to initialize email service - password reset emails will not be sent", "error", err)
		emailService = nil // Continue without email service in dev mode
	}

	// Initialize services
	jwtService, err := service.NewJWTService(
		cfg.Server.BaseURL,
		15*time.Minute, // Access token TTL
		7*24*time.Hour, // Refresh token TTL
	)
	if err != nil {
		appLog.Fatal("Failed to initialize JWT service", "error", err)
	}

	sessionService := service.NewSessionService(
		sessionStore, // Use Interface
		cfg.Session.Timeout,
	)

	authService := service.NewAuthService(
		userRepo,
		sessionService,
		auditRepo,
		cfg.Security.MaxLoginAttempts,
		cfg.Security.AccountLockoutDuration,
	)

	totpService := service.NewTOTPService(userRepo, "SSO Server")

	passwordPolicy := utils.PasswordPolicy{
		MinLength:      cfg.Security.PasswordMinLength,
		RequireUpper:   cfg.Security.PasswordRequireUpper,
		RequireLower:   cfg.Security.PasswordRequireLower,
		RequireNumber:  cfg.Security.PasswordRequireNumber,
		RequireSpecial: cfg.Security.PasswordRequireSpecial,
	}

	passwordService := service.NewPasswordService(
		userRepo,
		passwordResetRepo,
		sessionStore, // Use Interface
		passwordHistoryRepo,
		passwordPolicy,
		cfg.Security.PasswordHistoryCount,
		time.Hour, // Reset token expiry
	)

	// Initialize OAuth2 repositories
	oauth2ClientRepo := repository.NewOAuth2ClientRepository(db.DB)
	oauth2ScopeRepo := repository.NewOAuth2ScopeRepository(db.DB)
	oauth2CodeRepo := repository.NewOAuth2CodeRepository(db.DB)
	oauth2TokenRepo := repository.NewOAuth2TokenRepository(db.DB)
	oauth2ConsentRepo := repository.NewOAuth2ConsentRepository(db.DB)

	// Initialize Role, Permission & Config services
	permissionService := service.NewPermissionService(permissionRepo)
	roleService := service.NewRoleService(roleRepo, permissionRepo)
	configService := service.NewConfigService(configRepo)

	// Initialize OAuth2 services
	oauth2ClientService := service.NewOAuth2ClientService(oauth2ClientRepo, oauth2ScopeRepo)
	oauth2TokenService := service.NewOAuth2TokenService(
		oauth2TokenRepo,
		jwtService,
		cfg.OAuth2.AccessTokenExpiry,
		cfg.OAuth2.RefreshTokenExpiry,
	)
	oauth2AuthzService := service.NewOAuth2AuthorizationService(
		oauth2CodeRepo,
		oauth2ClientRepo,
		oauth2ConsentRepo,
		oauth2TokenService,
		cfg.OAuth2.AuthCodeExpiry,
		cfg.OAuth2.EnforcePKCE,
	)
	oauth2ConsentService := service.NewOAuth2ConsentService(oauth2ConsentRepo, oauth2ClientRepo, oauth2ScopeRepo)

	appLog.Info("Services initialized")

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, jwtService, totpService)
	passwordHandler := handler.NewPasswordHandler(passwordService, emailService)
	oauth2Handler := handler.NewOAuth2Handler(oauth2AuthzService, oauth2TokenService, oauth2ClientService, oauth2ConsentService, userRepo)
	oauth2AdminHandler := handler.NewOAuth2AdminHandler(oauth2ClientService, oauth2ConsentService)
	roleHandler := handler.NewRoleHandler(roleService)
	permissionHandler := handler.NewPermissionHandler(permissionService)
	configHandler := handler.NewConfigHandler(configService)
	adminHandler := handler.NewAdminHandler(
		userRepo,
		auditRepo,
		sessionStore, // Use Interface
		oauth2ClientRepo,
		passwordService,
		oauth2ClientService,
	)

	appLog.Info("Handlers initialized")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "SSO Server",
		ServerHeader: "SSO-Server",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			appLog.Error("Request error", "error", err, "path", c.Path())
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	// CORS configuration - must use specific origins when credentials are needed
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			// Allow all localhost origins for development
			return strings.HasPrefix(origin, "http://localhost:") ||
				strings.HasPrefix(origin, "http://127.0.0.1:")
		},
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
		AllowCredentials: true,
	}))

	// Public routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "SSO Server API",
			"version": "1.0.0",
			"status":  "operational",
		})
	})

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		// Check database health
		if err := db.Health(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":   "unhealthy",
				"service":  "sso-server",
				"database": "disconnected",
			})
		}

		return c.JSON(fiber.Map{
			"status":   "ok",
			"service":  "sso-server",
			"database": "connected",
		})
	})

	// Static UI registration moved to end to prevent shadowing API routes

	// Authentication routes (public)
	auth := app.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/logout", authHandler.Logout)
	auth.Post("/refresh", authHandler.RefreshSession)
	auth.Post("/verify-2fa", authHandler.Verify2FA)

	// Password management routes
	password := app.Group("/password")
	password.Post("/forgot", passwordHandler.ForgotPassword) // Public - request reset
	password.Post("/reset", passwordHandler.ResetPassword)   // Public - complete reset

	// Protected password routes (require authentication)
	password.Post("/change", middleware.AuthMiddleware(sessionService), passwordHandler.ChangePassword)

	// OAuth2 routes
	oauth2 := app.Group("/oauth2")
	oauth2.Get("/authorize", middleware.AuthMiddleware(sessionService), oauth2Handler.Authorize)                 // Protected - requires login
	oauth2.Post("/authorize/consent", middleware.AuthMiddleware(sessionService), oauth2Handler.AuthorizeConsent) // Protected - consent submission
	oauth2.Post("/token", oauth2Handler.Token)                                                                   // Public - token exchange
	oauth2.Post("/revoke", oauth2Handler.Revoke)                                                                 // Public - token revocation
	oauth2.Get("/userinfo", oauth2Handler.UserInfo)                                                              // Public (Bearer Auth) - user info

	// OAuth2 admin routes (require authentication)
	admin := app.Group("/admin")
	admin.Use(middleware.AuthMiddleware(sessionService))
	admin.Post("/oauth2/clients", oauth2AdminHandler.RegisterClient)
	admin.Get("/oauth2/clients/:client_id", oauth2AdminHandler.GetClient)
	admin.Post("/oauth2/clients/:client_id/regenerate-secret", oauth2AdminHandler.RegenerateSecret)
	admin.Delete("/oauth2/clients/:client_id", oauth2AdminHandler.RevokeClient)

	// Admin API routes (require authentication + admin role)
	adminAPI := app.Group("/admin/api")
	adminAPI.Use(middleware.AuthMiddleware(sessionService))
	adminAPI.Use(middleware.RoleMiddleware(userRepo, "admin", "super_admin"))

	// Dashboard statistics
	adminAPI.Get("/stats", adminHandler.GetStats)

	// User management
	adminAPI.Get("/users", adminHandler.GetUsers)
	adminAPI.Get("/users/:id", adminHandler.GetUser)
	adminAPI.Post("/users", adminHandler.CreateUser)
	adminAPI.Put("/users/:id", adminHandler.UpdateUser)
	adminAPI.Delete("/users/:id", adminHandler.DeleteUser)
	adminAPI.Post("/users/:id/reset-password", adminHandler.ResetUserPassword)
	adminAPI.Post("/users/:id/unlock", adminHandler.UnlockUser)
	adminAPI.Post("/users/:id/roles/:role_id", adminHandler.AssignRole)
	adminAPI.Delete("/users/:id/roles/:role_id", adminHandler.RemoveRole)

	// Audit logs
	adminAPI.Get("/audit-logs", adminHandler.GetAuditLogs)

	// Role management
	adminAPI.Get("/roles", roleHandler.GetRoles)
	adminAPI.Get("/roles/:id", roleHandler.GetRole)
	adminAPI.Post("/roles", roleHandler.CreateRole)
	adminAPI.Put("/roles/:id", roleHandler.UpdateRole)
	adminAPI.Delete("/roles/:id", roleHandler.DeleteRole)
	adminAPI.Post("/roles/:id/permissions", roleHandler.AssignPermissions)
	adminAPI.Delete("/roles/:id/permissions/:permission_id", roleHandler.RemovePermission)

	// Permission management
	adminAPI.Get("/permissions", permissionHandler.GetPermissions)
	adminAPI.Get("/permissions/:id", permissionHandler.GetPermission)
	adminAPI.Post("/permissions", permissionHandler.CreatePermission)
	adminAPI.Put("/permissions/:id", permissionHandler.UpdatePermission)
	adminAPI.Delete("/permissions/:id", permissionHandler.DeletePermission)

	// System info
	adminAPI.Get("/system", adminHandler.GetSystemInfo)

	// Configuration management
	adminAPI.Get("/config", configHandler.GetAllConfigs)
	adminAPI.Get("/config/:key", configHandler.GetConfig)
	adminAPI.Put("/config/:key", configHandler.UpdateConfig)

	// Session Config
	adminAPI.Post("/config/session/test-redis", adminHandler.TestRedisConnection)
	adminAPI.Post("/config/session/switch", adminHandler.SwitchSessionDriver)

	// OAuth2 clients (alternative endpoints)
	// OAuth2 clients (alternative endpoints)
	adminAPI.Get("/oauth2-clients", oauth2AdminHandler.GetClients)

	// User OAuth2 consent management
	user := app.Group("/user")
	user.Use(middleware.AuthMiddleware(sessionService))
	user.Get("/oauth2/consents", oauth2AdminHandler.GetUserConsents)
	user.Delete("/oauth2/consents/:client_id", oauth2AdminHandler.RevokeConsent)

	// Protected API routes (require authentication)

	api := app.Group("/api")
	api.Use(middleware.AuthMiddleware(sessionService))
	// Add protected routes here
	api.Get("/profile", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		return c.JSON(fiber.Map{
			"user_id": userID,
			"message": "This is a protected route",
		})
	})

	appLog.Info("Routes registered")

	// Serve Static UI (Login Page) if exists - Registered LAST to act as fallback
	if _, err := os.Stat("./static"); err == nil {
		appLog.Info("Serving static UI from ./static")
		app.Static("/", "./static")

		// SPA Fallback for non-API routes that weren't matched above
		app.Get("*", func(c *fiber.Ctx) error {
			// Double check we aren't intercepting APIs just in case
			path := c.Path()
			if strings.HasPrefix(path, "/api") ||
				strings.HasPrefix(path, "/auth") ||
				strings.HasPrefix(path, "/oauth2") ||
				strings.HasPrefix(path, "/admin") ||
				strings.HasPrefix(path, "/user") {
				return c.Next()
			}
			return c.SendFile("./static/index.html")
		})
	}

	// Start server
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	appLog.Info("Server starting on " + serverAddr)

	// Graceful shutdown
	go func() {
		if err := app.Listen(serverAddr); err != nil {
			appLog.Fatal("Server failed to start", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLog.Info("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		appLog.Error("Server shutdown error", "error", err)
	}
	appLog.Info("Server stopped")
}
