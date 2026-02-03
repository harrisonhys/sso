package testutil

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/sso-project/sso-server/internal/database"
	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/utils"
)

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, db *database.DB, email string) *models.User {
	t.Helper()

	hashedPassword, err := utils.HashPassword("TestPassword123!")
	require.NoError(t, err)

	user := &models.User{
		ID:                  uuid.New().String(),
		Email:               email,
		PasswordHash:        hashedPassword,
		Name:                "Test User",
		IsActive:            true,
		EmailVerified:       true,
		FailedLoginAttempts: 0,
		PasswordChangedAt:   time.Now(),
	}

	err = db.DB.Create(user).Error
	require.NoError(t, err, "Failed to create test user")

	return user
}

// CreateTestUserWithPassword creates a test user with a specific password
func CreateTestUserWithPassword(t *testing.T, db *database.DB, email, password string) *models.User {
	t.Helper()

	hashedPassword, err := utils.HashPassword(password)
	require.NoError(t, err)

	user := &models.User{
		ID:                  uuid.New().String(),
		Email:               email,
		PasswordHash:        hashedPassword,
		Name:                "Test User",
		IsActive:            true,
		EmailVerified:       true,
		FailedLoginAttempts: 0,
		PasswordChangedAt:   time.Now(),
	}

	err = db.DB.Create(user).Error
	require.NoError(t, err, "Failed to create test user")

	return user
}

// CreateTestSession creates a test session for a user
func CreateTestSession(t *testing.T, db *database.DB, userID string) *models.Session {
	t.Helper()

	token, err := utils.GenerateRandomToken(32)
	require.NoError(t, err)

	session := &models.Session{
		ID:             uuid.New().String(),
		UserID:         userID,
		SessionToken:   token,
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		LastActivityAt: time.Now(),
		IPAddress:      "127.0.0.1",
		UserAgent:      "Test Agent",
	}

	err = db.DB.Create(session).Error
	require.NoError(t, err, "Failed to create test session")

	return session
}

// CreateTestOAuth2Client creates a test OAuth2 client
func CreateTestOAuth2Client(t *testing.T, db *database.DB) *models.OAuthClient {
	t.Helper()

	secret, err := utils.GenerateRandomToken(32)
	require.NoError(t, err)

	client := &models.OAuthClient{
		ID:            uuid.New().String(),
		ClientID:      "test-client-id",
		ClientSecret:  secret,
		Name:          "Test Client",
		RedirectURIs:  "http://localhost:3000/callback",
		AllowedScopes: "openid profile email",
		IsActive:      true,
	}

	err = db.DB.Create(client).Error
	require.NoError(t, err, "Failed to create test OAuth2 client")

	return client
}

// CreateTestPasswordReset creates a test password reset token
func CreateTestPasswordReset(t *testing.T, db *database.DB, userID, email string) *models.PasswordResetToken {
	t.Helper()

	token, err := utils.GenerateRandomToken(32)
	require.NoError(t, err)

	reset := &models.PasswordResetToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}

	err = db.DB.Create(reset).Error
	require.NoError(t, err, "Failed to create test password reset")

	return reset
}

// CreateTestRole creates a test role
func CreateTestRole(t *testing.T, db *database.DB, name string) *models.Role {
	t.Helper()

	role := &models.Role{
		ID:          uuid.New().String(),
		Name:        name,
		Description: "Test role: " + name,
	}

	err := db.DB.Create(role).Error
	require.NoError(t, err, "Failed to create test role")

	return role
}

// CreateTestPermission creates a test permission
func CreateTestPermission(t *testing.T, db *database.DB, name string) *models.Permission {
	t.Helper()

	permission := &models.Permission{
		ID:          uuid.New().String(),
		Name:        name,
		Description: "Test permission: " + name,
		Resource:    "test",
		Action:      "read",
	}

	err := db.DB.Create(permission).Error
	require.NoError(t, err, "Failed to create test permission")

	return permission
}

// CreateExpiredSession creates an expired session for testing cleanup
func CreateExpiredSession(t *testing.T, db *database.DB, userID string) *models.Session {
	t.Helper()

	token, err := utils.GenerateRandomToken(32)
	require.NoError(t, err)

	session := &models.Session{
		ID:             uuid.New().String(),
		UserID:         userID,
		SessionToken:   token,
		ExpiresAt:      time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		LastActivityAt: time.Now().Add(-1 * time.Hour),
		IPAddress:      "127.0.0.1",
		UserAgent:      "Test Agent",
	}

	err = db.DB.Create(session).Error
	require.NoError(t, err, "Failed to create expired session")

	return session
}

// CreateLockedUser creates a locked user account for testing
func CreateLockedUser(t *testing.T, db *database.DB, email string) *models.User {
	t.Helper()

	hashedPassword, err := utils.HashPassword("TestPassword123!")
	require.NoError(t, err)

	lockedUntil := time.Now().Add(1 * time.Hour)
	user := &models.User{
		ID:                  uuid.New().String(),
		Email:               email,
		PasswordHash:        hashedPassword,
		Name:                "Locked User",
		IsActive:            true, // Active but locked
		IsLocked:            true, // This is the lock flag
		EmailVerified:       true,
		FailedLoginAttempts: 5,
		LockedUntil:         &lockedUntil,
		PasswordChangedAt:   time.Now(),
	}

	err = db.DB.Create(user).Error
	require.NoError(t, err, "Failed to create locked user")

	return user
}
