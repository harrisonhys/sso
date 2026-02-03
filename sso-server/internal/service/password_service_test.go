package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sso-project/sso-server/internal/repository"
	"github.com/sso-project/sso-server/internal/testutil"
	"github.com/sso-project/sso-server/internal/utils"
)

// Test ChangePassword - Success
func TestPasswordService_ChangePassword_Success(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	resetTokenRepo := repository.NewPasswordResetTokenRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	historyRepo := repository.NewPasswordHistoryRepository(db)

	policy := utils.PasswordPolicy{
		MinLength:      8,
		RequireUpper:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}

	passwordService := NewPasswordService(
		userRepo,
		resetTokenRepo,
		sessionRepo,
		historyRepo,
		policy,
		3, // history count
		24*time.Hour,
	)

	// Create test user
	user := testutil.CreateTestUserWithPassword(t, db, "test@example.com", "OldPassword123!")

	// Create a session that should be invalidated
	testutil.CreateTestSession(t, db, user.ID)

	ctx := context.Background()

	// Test: Change password
	err := passwordService.ChangePassword(ctx, user.ID, "OldPassword123!", "NewPassword456!")
	require.NoError(t, err)

	// Verify: Password updated
	updatedUser, err := userRepo.GetByID(ctx, user.ID)
	require.NoError(t, err)

	// Verify new password works
	err = utils.ComparePassword(updatedUser.PasswordHash, "NewPassword456!")
	assert.NoError(t, err, "New password should work")

	// Verify old password doesn't work
	err = utils.ComparePassword(updatedUser.PasswordHash, "OldPassword123!")
	assert.Error(t, err, "Old password should not work")

	// Verify PasswordChangedAt updated
	assert.True(t, updatedUser.PasswordChangedAt.After(user.PasswordChangedAt))

	// Verify old sessions invalidated
	sessions, err := sessionRepo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, len(sessions), "All sessions should be invalidated")
}

// Test ChangePassword - Wrong Current Password
func TestPasswordService_ChangePassword_WrongCurrentPassword(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	resetTokenRepo := repository.NewPasswordResetTokenRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	historyRepo := repository.NewPasswordHistoryRepository(db)

	policy := utils.PasswordPolicy{
		MinLength:      8,
		RequireUpper:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}

	passwordService := NewPasswordService(
		userRepo,
		resetTokenRepo,
		sessionRepo,
		historyRepo,
		policy,
		3,
		24*time.Hour,
	)

	// Create test user
	user := testutil.CreateTestUserWithPassword(t, db, "test@example.com", "CorrectPassword123!")

	ctx := context.Background()

	// Test: Attempt change with wrong current password
	err := passwordService.ChangePassword(ctx, user.ID, "WrongPassword!", "NewPassword456!")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "current password is incorrect")

	// Verify: Password NOT changed
	updatedUser, err := userRepo.GetByID(ctx, user.ID)
	require.NoError(t, err)

	err = utils.ComparePassword(updatedUser.PasswordHash, "CorrectPassword123!")
	assert.NoError(t, err, "Original password should still work")
}

// Test ChangePassword - Weak New Password
func TestPasswordService_ChangePassword_WeakNewPassword(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	resetTokenRepo := repository.NewPasswordResetTokenRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	historyRepo := repository.NewPasswordHistoryRepository(db)

	policy := utils.PasswordPolicy{
		MinLength:      8,
		RequireUpper:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}

	passwordService := NewPasswordService(
		userRepo,
		resetTokenRepo,
		sessionRepo,
		historyRepo,
		policy,
		3,
		24*time.Hour,
	)

	// Create test user
	user := testutil.CreateTestUserWithPassword(t, db, "test@example.com", "OldPassword123!")

	ctx := context.Background()

	// Test: Attempt change to weak password (no special char)
	err := passwordService.ChangePassword(ctx, user.ID, "OldPassword123!", "weakpassword")
	assert.Error(t, err)

	// Verify: Password NOT changed
	updatedUser, err := userRepo.GetByID(ctx, user.ID)
	require.NoError(t, err)

	err = utils.ComparePassword(updatedUser.PasswordHash, "OldPassword123!")
	assert.NoError(t, err, "Original password should still work")
}

// Test ChangePassword - Password Reused
func TestPasswordService_ChangePassword_PasswordReused(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	resetTokenRepo := repository.NewPasswordResetTokenRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	historyRepo := repository.NewPasswordHistoryRepository(db)

	policy := utils.PasswordPolicy{
		MinLength:      8,
		RequireUpper:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}

	passwordService := NewPasswordService(
		userRepo,
		resetTokenRepo,
		sessionRepo,
		historyRepo,
		policy,
		3, // Remember last 3 passwords
		24*time.Hour,
	)

	// Create test user
	user := testutil.CreateTestUserWithPassword(t, db, "test@example.com", "Password1!")

	ctx := context.Background()

	// Change password once
	err := passwordService.ChangePassword(ctx, user.ID, "Password1!", "Password2!")
	require.NoError(t, err)

	// Test: Attempt to change back to old password
	err = passwordService.ChangePassword(ctx, user.ID, "Password2!", "Password1!")
	assert.Error(t, err)
	assert.Equal(t, ErrPasswordReused, err)

	// Verify: Password still Password2!
	updatedUser, err := userRepo.GetByID(ctx, user.ID)
	require.NoError(t, err)

	err = utils.ComparePassword(updatedUser.PasswordHash, "Password2!")
	assert.NoError(t, err, "Current password should still work")
}

// Test ChangePassword - Password History
func TestPasswordService_ChangePassword_PasswordHistory(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	resetTokenRepo := repository.NewPasswordResetTokenRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	historyRepo := repository.NewPasswordHistoryRepository(db)

	policy := utils.PasswordPolicy{
		MinLength:      8,
		RequireUpper:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}

	passwordService := NewPasswordService(
		userRepo,
		resetTokenRepo,
		sessionRepo,
		historyRepo,
		policy,
		3, // Remember last 3 passwords
		24*time.Hour,
	)

	// Create test user
	user := testutil.CreateTestUserWithPassword(t, db, "test@example.com", "Password1!")

	ctx := context.Background()

	// Change password 3 times
	err := passwordService.ChangePassword(ctx, user.ID, "Password1!", "Password2!")
	require.NoError(t, err)

	err = passwordService.ChangePassword(ctx, user.ID, "Password2!", "Password3!")
	require.NoError(t, err)

	err = passwordService.ChangePassword(ctx, user.ID, "Password3!", "Password4!")
	require.NoError(t, err)

	// Verify: All 3 old passwords in history
	history, err := historyRepo.GetRecentPasswords(ctx, user.ID, 3)
	require.NoError(t, err)
	assert.Equal(t, 3, len(history), "Should have 3 passwords in history")

	// Test: Attempt to use password from history (Password2!)
	err = passwordService.ChangePassword(ctx, user.ID, "Password4!", "Password2!")
	assert.Error(t, err)
	assert.Equal(t, ErrPasswordReused, err)
}

// Test ChangePassword - User Not Found
func TestPasswordService_ChangePassword_UserNotFound(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	resetTokenRepo := repository.NewPasswordResetTokenRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	historyRepo := repository.NewPasswordHistoryRepository(db)

	policy := utils.PasswordPolicy{
		MinLength:      8,
		RequireUpper:   true,
		RequireNumber:  true,
		RequireSpecial: true,
	}

	passwordService := NewPasswordService(
		userRepo,
		resetTokenRepo,
		sessionRepo,
		historyRepo,
		policy,
		3,
		24*time.Hour,
	)

	ctx := context.Background()

	// Test: Attempt change for non-existent user
	err := passwordService.ChangePassword(ctx, "non-existent-id", "OldPassword!", "NewPassword!")
	assert.Error(t, err)
}
