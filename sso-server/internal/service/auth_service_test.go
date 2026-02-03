package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
	"github.com/sso-project/sso-server/internal/testutil"
	"github.com/sso-project/sso-server/internal/utils"
)

func TestAuthService_Login_Success(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	// Create repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	auditRepo := repository.NewAuditLogRepository(db)

	// Create services
	sessionService := NewSessionService(sessionRepo, 24*time.Hour)
	authService := NewAuthService(userRepo, sessionService, auditRepo, 5, 30*time.Minute)

	// Create test user
	user := testutil.CreateTestUserWithPassword(t, db, "test@example.com", "TestPassword123!")

	// Test
	ctx := context.Background()
	result, err := authService.Login(ctx, "test@example.com", "TestPassword123!", "127.0.0.1", "Test Agent")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, user.ID, result.User.ID)
	assert.Equal(t, user.Email, result.User.Email)
	assert.NotNil(t, result.Session)
	assert.False(t, result.RequiresTwoFactor)
	assert.NotEmpty(t, result.Session.SessionToken)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	auditRepo := repository.NewAuditLogRepository(db)

	sessionService := NewSessionService(sessionRepo, 24*time.Hour)
	authService := NewAuthService(userRepo, sessionService, auditRepo, 5, 30*time.Minute)

	// Create test user
	testutil.CreateTestUserWithPassword(t, db, "test@example.com", "CorrectPassword123!")

	// Test with wrong password
	ctx := context.Background()
	result, err := authService.Login(ctx, "test@example.com", "WrongPassword", "127.0.0.1", "Test Agent")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, result)

	// Verify failed attempts incremented
	user, _ := userRepo.GetByEmail(ctx, "test@example.com")
	assert.Equal(t, 1, user.FailedLoginAttempts)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	auditRepo := repository.NewAuditLogRepository(db)

	sessionService := NewSessionService(sessionRepo, 24*time.Hour)
	authService := NewAuthService(userRepo, sessionService, auditRepo, 5, 30*time.Minute)

	// Test with non-existent user
	ctx := context.Background()
	result, err := authService.Login(ctx, "nonexistent@example.com", "password", "127.0.0.1", "Test Agent")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, result)
}

func TestAuthService_Login_AccountLocked(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	auditRepo := repository.NewAuditLogRepository(db)

	sessionService := NewSessionService(sessionRepo, 24*time.Hour)
	authService := NewAuthService(userRepo, sessionService, auditRepo, 5, 30*time.Minute)

	// Create locked user
	testutil.CreateLockedUser(t, db, "locked@example.com")

	// Test
	ctx := context.Background()
	result, err := authService.Login(ctx, "locked@example.com", "TestPassword123!", "127.0.0.1", "Test Agent")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrAccountLocked, err)
	assert.Nil(t, result)
}

func TestAuthService_Login_MaxAttemptsLockout(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	auditRepo := repository.NewAuditLogRepository(db)

	sessionService := NewSessionService(sessionRepo, 24*time.Hour)
	maxAttempts := 3
	authService := NewAuthService(userRepo, sessionService, auditRepo, maxAttempts, 30*time.Minute)

	// Create test user
	testutil.CreateTestUserWithPassword(t, db, "test@example.com", "CorrectPassword123!")

	ctx := context.Background()

	// Attempt login with wrong password multiple times
	for i := 0; i < maxAttempts; i++ {
		result, err := authService.Login(ctx, "test@example.com", "WrongPassword", "127.0.0.1", "Test Agent")
		assert.Error(t, err)
		assert.Nil(t, result)
	}

	// Verify account is now locked
	user, _ := userRepo.GetByEmail(ctx, "test@example.com")
	assert.True(t, user.IsLocked)
	assert.NotNil(t, user.LockedUntil)
	assert.True(t, user.LockedUntil.After(time.Now()))

	// Next attempt should return account locked error
	result, err := authService.Login(ctx, "test@example.com", "CorrectPassword123!", "127.0.0.1", "Test Agent")
	assert.Error(t, err)
	assert.Equal(t, ErrAccountLocked, err)
	assert.Nil(t, result)
}

func TestAuthService_Login_InactiveAccount(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	auditRepo := repository.NewAuditLogRepository(db)

	sessionService := NewSessionService(sessionRepo, 24*time.Hour)
	authService := NewAuthService(userRepo, sessionService, auditRepo, 5, 30*time.Minute)

	// Create inactive user
	hashedPassword, _ := utils.HashPassword("TestPassword123!")
	userID := uuid.New().String()
	user := &models.User{
		ID:                userID,
		Email:             "inactive@example.com",
		PasswordHash:      hashedPassword,
		Name:              "Inactive User",
		EmailVerified:     true,
		PasswordChangedAt: time.Now(),
	}
	err := db.DB.Create(user).Error
	require.NoError(t, err)

	// Explicitly set IsActive to false (override GORM default)
	err = db.DB.Model(&models.User{}).Where("id = ?", userID).Update("is_active", false).Error
	require.NoError(t, err)

	// Test
	ctx := context.Background()
	result, err := authService.Login(ctx, "inactive@example.com", "TestPassword123!", "127.0.0.1", "Test Agent")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrAccountInactive, err)
	assert.Nil(t, result)
}

func TestAuthService_Login_Requires2FA(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	auditRepo := repository.NewAuditLogRepository(db)

	sessionService := NewSessionService(sessionRepo, 24*time.Hour)
	authService := NewAuthService(userRepo, sessionService, auditRepo, 5, 30*time.Minute)

	// Create user with 2FA enabled
	hashedPassword, _ := utils.HashPassword("TestPassword123!")
	userID := uuid.New().String()
	user := &models.User{
		ID:                userID,
		Email:             "2fa@example.com",
		PasswordHash:      hashedPassword,
		Name:              "2FA User",
		IsActive:          true,
		EmailVerified:     true,
		PasswordChangedAt: time.Now(),
	}
	err := db.DB.Create(user).Error
	require.NoError(t, err)

	// Create 2FA record separately
	twoFA := &models.TwoFactorAuth{
		ID:              uuid.New().String(),
		UserID:          userID,
		SecretEncrypted: "test-secret",
		Enabled:         true,
	}
	err = db.DB.Create(twoFA).Error
	require.NoError(t, err)

	// Test
	ctx := context.Background()
	result, err := authService.Login(ctx, "2fa@example.com", "TestPassword123!", "127.0.0.1", "Test Agent")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.RequiresTwoFactor)
	assert.NotEmpty(t, result.TempToken)
	assert.NotNil(t, result.Session)

	// Verify temp session has short expiry (5 minutes)
	assert.True(t, result.Session.ExpiresAt.Before(time.Now().Add(6*time.Minute)))
	assert.True(t, result.Session.ExpiresAt.After(time.Now().Add(4*time.Minute)))
}

func TestAuthService_Login_ResetFailedAttemptsOnSuccess(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	auditRepo := repository.NewAuditLogRepository(db)

	sessionService := NewSessionService(sessionRepo, 24*time.Hour)
	authService := NewAuthService(userRepo, sessionService, auditRepo, 5, 30*time.Minute)

	// Create test user
	testutil.CreateTestUserWithPassword(t, db, "test@example.com", "CorrectPassword123!")

	ctx := context.Background()

	// Fail login twice
	authService.Login(ctx, "test@example.com", "WrongPassword", "127.0.0.1", "Test Agent")
	authService.Login(ctx, "test@example.com", "WrongPassword", "127.0.0.1", "Test Agent")

	// Verify failed attempts incremented
	user, _ := userRepo.GetByEmail(ctx, "test@example.com")
	assert.Equal(t, 2, user.FailedLoginAttempts, "Should have 2 failed attempts")

	// Successful login (should reset failed attempts)
	result, err := authService.Login(ctx, "test@example.com", "CorrectPassword123!", "127.0.0.1", "Test Agent")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Now fail login ONE more time
	authService.Login(ctx, "test@example.com", "WrongPassword", "127.0.0.1", "Test Agent")

	// Verify failed attempts is 1 (not 3), proving reset worked
	user, _ = userRepo.GetByEmail(ctx, "test@example.com")
	assert.Equal(t, 1, user.FailedLoginAttempts,
		"Failed attempts should be 1 (reset after success), not 3 (continuing from before)")
}

func TestAuthService_Logout_Success(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	auditRepo := repository.NewAuditLogRepository(db)

	sessionService := NewSessionService(sessionRepo, 24*time.Hour)
	authService := NewAuthService(userRepo, sessionService, auditRepo, 5, 30*time.Minute)

	// Create user and session
	user := testutil.CreateTestUser(t, db, "test@example.com")
	session := testutil.CreateTestSession(t, db, user.ID)

	// Test
	ctx := context.Background()
	err := authService.Logout(ctx, session.SessionToken, "127.0.0.1", "Test Agent")

	// Assert
	assert.NoError(t, err)

	// Verify session is terminated
	_, err = sessionService.ValidateSession(ctx, session.SessionToken)
	assert.Error(t, err) // Should error because session is deleted
}

func TestAuthService_RefreshSession_Success(t *testing.T) {
	// Setup
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewDatabaseSessionStore(db)
	auditRepo := repository.NewAuditLogRepository(db)

	sessionService := NewSessionService(sessionRepo, 24*time.Hour)
	authService := NewAuthService(userRepo, sessionService, auditRepo, 5, 30*time.Minute)

	// Create user and session
	user := testutil.CreateTestUser(t, db, "test@example.com")
	session := testutil.CreateTestSession(t, db, user.ID)
	originalExpiry := session.ExpiresAt

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Test
	ctx := context.Background()
	refreshedSession, err := authService.RefreshSession(ctx, session.SessionToken)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, refreshedSession)
	assert.Equal(t, session.SessionToken, refreshedSession.SessionToken)
	assert.True(t, refreshedSession.ExpiresAt.After(originalExpiry))
}
