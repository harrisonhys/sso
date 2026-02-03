package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
	"github.com/sso-project/sso-server/internal/utils"
)

var (
	ErrTokenExpired   = errors.New("reset token has expired")
	ErrTokenUsed      = errors.New("reset token has already been used")
	ErrTokenNotFound  = errors.New("reset token not found")
	ErrPasswordReused = errors.New("password was recently used")
)

// PasswordService handles password operations
type PasswordService struct {
	userRepo       *repository.UserRepository
	resetTokenRepo *repository.PasswordResetTokenRepository
	sessionRepo    repository.SessionStore
	historyRepo    *repository.PasswordHistoryRepository
	policy         utils.PasswordPolicy
	historyCount   int
	tokenExpiry    time.Duration
}

// NewPasswordService creates a new password service
func NewPasswordService(
	userRepo *repository.UserRepository,
	resetTokenRepo *repository.PasswordResetTokenRepository,
	sessionRepo repository.SessionStore,
	historyRepo *repository.PasswordHistoryRepository,
	policy utils.PasswordPolicy,
	historyCount int,
	tokenExpiry time.Duration,
) *PasswordService {
	return &PasswordService{
		userRepo:       userRepo,
		resetTokenRepo: resetTokenRepo,
		sessionRepo:    sessionRepo,
		historyRepo:    historyRepo,
		policy:         policy,
		historyCount:   historyCount,
		tokenExpiry:    tokenExpiry,
	}
}

// ChangePassword changes a user's password (requires current password)
func (s *PasswordService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := utils.ComparePassword(user.PasswordHash, currentPassword); err != nil {
		return errors.New("current password is incorrect")
	}

	// Validate new password
	if err := utils.ValidatePassword(newPassword, s.policy); err != nil {
		return err
	}

	// Check password history to prevent reuse
	if s.historyCount > 0 {
		recentPasswords, err := s.historyRepo.GetRecentPasswords(ctx, userID, s.historyCount)
		if err != nil {
			// Log error but don't fail - password history is not critical
		}
		for _, oldPassword := range recentPasswords {
			if err := utils.ComparePassword(oldPassword.PasswordHash, newPassword); err == nil {
				return ErrPasswordReused
			}
		}
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Save old password hash BEFORE updating
	oldPasswordHash := user.PasswordHash

	// Update password
	user.PasswordHash = hashedPassword
	now := time.Now()
	user.PasswordChangedAt = now

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Add OLD password to history (not the new one!)
	if s.historyCount > 0 {
		if err := s.historyRepo.Create(ctx, userID, oldPasswordHash); err != nil {
			// Log error but don't fail
		}
		// Cleanup old passwords beyond history limit
		if err := s.historyRepo.CleanupOldPasswords(ctx, userID, s.historyCount); err != nil {
			// Log error but don't fail
		}
	}

	// Invalidate all user sessions for security
	if err := s.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		// Log error but don't fail the password change
		// Worst case: old sessions remain valid until expiry
	}

	return nil
}

// GenerateResetToken generates a password reset token and stores it
func (s *PasswordService) GenerateResetToken(ctx context.Context, email string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists - still return success
		return "", nil
	}

	// Generate secure token
	tokenString, err := utils.GenerateRandomToken(32)
	if err != nil {
		return "", err
	}

	token := &models.PasswordResetToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Email:     email,
		Token:     tokenString,
		ExpiresAt: time.Now().Add(s.tokenExpiry),
		Used:      false,
	}

	// Store token in database
	if err := s.resetTokenRepo.Create(ctx, token); err != nil {
		return "", err
	}

	return tokenString, nil
}

// ResetPassword resets a password using a token
func (s *PasswordService) ResetPassword(ctx context.Context, tokenString, newPassword string) error {
	// Find and validate token
	token, err := s.resetTokenRepo.FindByToken(ctx, tokenString)
	if err != nil {
		return ErrTokenNotFound
	}

	// Check if token is still valid
	if token.Used {
		return ErrTokenUsed
	}
	if time.Now().After(token.ExpiresAt) {
		return ErrTokenExpired
	}

	// Validate new password
	if err := utils.ValidatePassword(newPassword, s.policy); err != nil {
		return err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return err
	}

	// Check password history to prevent reuse
	if s.historyCount > 0 {
		recentPasswords, err := s.historyRepo.GetRecentPasswords(ctx, user.ID, s.historyCount)
		if err != nil {
			// Log error but don't fail - password history is not critical
		}
		for _, oldPassword := range recentPasswords {
			if err := utils.ComparePassword(oldPassword.PasswordHash, newPassword); err == nil {
				return ErrPasswordReused
			}
		}
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Save old password hash BEFORE updating
	oldPasswordHash := user.PasswordHash

	// Update password
	user.PasswordHash = hashedPassword
	now := time.Now()
	user.PasswordChangedAt = now

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Mark token as used
	if err := s.resetTokenRepo.MarkAsUsed(ctx, tokenString); err != nil {
		// Log error but don't fail - password was already changed
	}

	// Add OLD password to history (not the new one!)
	if s.historyCount > 0 {
		if err := s.historyRepo.Create(ctx, user.ID, oldPasswordHash); err != nil {
			// Log error but don't fail
		}
		// Cleanup old passwords beyond history limit
		if err := s.historyRepo.CleanupOldPasswords(ctx, user.ID, s.historyCount); err != nil {
			// Log error but don't fail
		}
	}

	// Invalidate all user sessions for security
	if err := s.sessionRepo.DeleteByUserID(ctx, token.UserID); err != nil {
		// Log error but don't fail the password reset
	}

	return nil
}

// CheckPasswordExpiry checks if a user's password has expired
func (s *PasswordService) CheckPasswordExpiry(user *models.User, expiryDays int) bool {
	if expiryDays <= 0 {
		return false // Password expiry disabled
	}

	expiryDate := user.PasswordChangedAt.AddDate(0, 0, expiryDays)
	return time.Now().After(expiryDate)
}
