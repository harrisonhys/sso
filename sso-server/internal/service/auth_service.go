package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
	"github.com/sso-project/sso-server/internal/utils"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account is locked due to too many failed attempts")
	ErrAccountInactive    = errors.New("account is not active")
	ErrPasswordExpired    = errors.New("password has expired")
	ErrTwoFactorRequired  = errors.New("two-factor authentication required")
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo        *repository.UserRepository
	sessionService  *SessionService
	auditRepo       *repository.AuditLogRepository
	maxAttempts     int
	lockoutDuration time.Duration
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo *repository.UserRepository,
	sessionService *SessionService,
	auditRepo *repository.AuditLogRepository,
	maxAttempts int,
	lockoutDuration time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		sessionService:  sessionService,
		auditRepo:       auditRepo,
		maxAttempts:     maxAttempts,
		lockoutDuration: lockoutDuration,
	}
}

// LoginResult contains the result of a login attempt
type LoginResult struct {
	User              *models.User
	Session           *models.Session
	RequiresTwoFactor bool
	TempToken         string // Temporary token for 2FA verification
}

// Login authenticates a user with email and password
func (s *AuthService) Login(ctx context.Context, email, password, ipAddress, userAgent string) (*LoginResult, error) {
	log.Printf("[AUTH_SERVICE_DEBUG] Login called for email: %s", email)

	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		log.Printf("[AUTH_SERVICE_DEBUG] GetByEmail error: %v", err)
		// Log failed attempt
		s.logAudit(ctx, nil, "login_failed", "authentication", ipAddress, userAgent, "user not found")
		return nil, ErrInvalidCredentials
	}

	log.Printf("[AUTH_SERVICE_DEBUG] User found: ID=%s, Email=%s, IsActive=%v, IsLocked=%v", user.ID, user.Email, user.IsActive, user.IsLocked)

	// Check if account is locked
	if user.IsLocked {
		if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
			s.logAudit(ctx, &user.ID, "login_failed", "authentication", ipAddress, userAgent, "account locked")
			return nil, ErrAccountLocked
		}
		// Unlock account if lockout period has passed
		user.IsLocked = false
		user.LockedUntil = nil
		user.FailedLoginAttempts = 0
		s.userRepo.Update(ctx, user)
	}

	// Check if account is active
	if !user.IsActive {
		s.logAudit(ctx, &user.ID, "login_failed", "authentication", ipAddress, userAgent, "account inactive")
		return nil, ErrAccountInactive
	}

	// Verify password
	log.Printf("[AUTH_DEBUG] Attempting to verify password for user: %s", email)
	log.Printf("[AUTH_DEBUG] Password hash from DB (first 30 chars): %s...", user.PasswordHash[:min(30, len(user.PasswordHash))])
	log.Printf("[AUTH_DEBUG] Password length: %d", len(password))

	if err := utils.ComparePassword(user.PasswordHash, password); err != nil {
		log.Printf("[AUTH_DEBUG] Password comparison failed: %v", err)
		// Increment failed attempts in database
		s.userRepo.IncrementFailedAttempts(ctx, user.ID)

		// Re-fetch user to get updated failed attempts count
		user, err = s.userRepo.GetByEmail(ctx, email)
		if err != nil {
			return nil, err
		}

		// Lock account if max attempts reached
		if user.FailedLoginAttempts >= s.maxAttempts {
			user.IsLocked = true
			lockoutUntil := time.Now().Add(s.lockoutDuration)
			user.LockedUntil = &lockoutUntil
			s.userRepo.Update(ctx, user)

			s.logAudit(ctx, &user.ID, "account_locked", "authentication", ipAddress, userAgent, "max login attempts exceeded")
			return nil, ErrAccountLocked
		}

		s.logAudit(ctx, &user.ID, "login_failed", "authentication", ipAddress, userAgent, "invalid password")
		return nil, ErrInvalidCredentials
	}

	log.Printf("[AUTH_DEBUG] Password verified successfully!")

	// Reset failed attempts on successful password verification
	if user.FailedLoginAttempts > 0 {
		log.Printf("[AUTH_DEBUG] Resetting failed attempts for user %s (current: %d)", user.ID, user.FailedLoginAttempts)
		err := s.userRepo.ResetFailedAttempts(ctx, user.ID)
		if err != nil {
			log.Printf("[AUTH_DEBUG] ERROR resetting failed attempts: %v", err)
		} else {
			log.Printf("[AUTH_DEBUG] Successfully reset failed attempts")
		}
	}

	// Check if 2FA is enabled
	if user.TwoFactorAuth != nil && user.TwoFactorAuth.Enabled {
		// Create a temporary session with short expiry for 2FA verification (5 minutes)
		tempSession, err := s.sessionService.CreateSession(ctx, user.ID, ipAddress, userAgent)
		if err != nil {
			return nil, err
		}

		// Override expiry to make it short-lived
		tempSession.ExpiresAt = time.Now().Add(5 * time.Minute)

		s.logAudit(ctx, &user.ID, "2fa_required", "authentication", ipAddress, userAgent, "")

		return &LoginResult{
			User:              user,
			Session:           tempSession,
			RequiresTwoFactor: true,
			TempToken:         tempSession.SessionToken,
		}, nil
	}

	// Create session for non-2FA users
	session, err := s.sessionService.CreateSession(ctx, user.ID, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	// Update last login time only (don't overwrite other fields like failed_login_attempts)
	now := time.Now()
	err = s.userRepo.UpdateLastLogin(ctx, user.ID, &now)
	if err != nil {
		log.Printf("[AUTH_DEBUG] Failed to update last login: %v", err)
	}

	s.logAudit(ctx, &user.ID, "login_success", "authentication", ipAddress, userAgent, "")

	return &LoginResult{
		User:              user,
		Session:           session,
		RequiresTwoFactor: false,
	}, nil
}

// Logout terminates a user session
func (s *AuthService) Logout(ctx context.Context, sessionToken, ipAddress, userAgent string) error {
	session, err := s.sessionService.ValidateSession(ctx, sessionToken)
	if err != nil {
		return err
	}

	s.logAudit(ctx, &session.UserID, "logout", "authentication", ipAddress, userAgent, "")

	return s.sessionService.TerminateSession(ctx, sessionToken)
}

// RefreshSession renews a session
func (s *AuthService) RefreshSession(ctx context.Context, sessionToken string) (*models.Session, error) {
	return s.sessionService.RenewSession(ctx, sessionToken)
}

// GetSessionByToken retrieves a session by token (for 2FA temp token validation)
func (s *AuthService) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	// Use ValidateSession which checks expiry
	return s.sessionService.ValidateSession(ctx, token)
}

// DeleteSessionByToken deletes a session by token (cleanup temp 2FA session)
func (s *AuthService) DeleteSessionByToken(ctx context.Context, token string) error {
	return s.sessionService.TerminateSession(ctx, token)
}

// CreateSessionAfter2FA creates a new session after 2FA verification
func (s *AuthService) CreateSessionAfter2FA(ctx context.Context, userID, ipAddress, userAgent string) (*models.Session, error) {
	// Create a real session (not a temp 2FA session)
	session, err := s.sessionService.CreateSession(ctx, userID, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	// Update last login
	user, err := s.userRepo.GetByID(ctx, userID)
	if err == nil {
		now := time.Now()
		user.LastLoginAt = &now
		s.userRepo.Update(ctx, user)
	}

	// Log successful 2FA login
	s.logAudit(ctx, &userID, "login_success_2fa", "authentication", ipAddress, userAgent, "2FA verified")

	return session, nil
}

// LogAudit is a public wrapper for logging audit events
func (s *AuthService) LogAudit(ctx context.Context, userID *string, action, resource, ipAddress, userAgent, details string) {
	s.logAudit(ctx, userID, action, resource, ipAddress, userAgent, details)
}

// logAudit creates an audit log entry
func (s *AuthService) logAudit(ctx context.Context, userID *string, action, resource, ipAddress, userAgent, details string) {
	auditLog := &models.AuditLog{
		ID:        uuid.New().String(),
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   details,
		CreatedAt: time.Now(),
	}

	// Log errors but don't fail the main operation
	if err := s.auditRepo.Create(ctx, auditLog); err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
