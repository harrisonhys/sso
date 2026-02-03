package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/sso-project/sso-server/internal/database"
	"github.com/sso-project/sso-server/internal/models"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *database.DB {
	t.Helper()

	// Use in-memory SQLite for fast, isolated tests
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Quiet during tests
	})
	require.NoError(t, err, "Failed to connect to test database")

	// Auto-migrate all models
	err = gormDB.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.PasswordResetToken{},
		&models.PasswordHistory{},
		&models.Role{},
		&models.Permission{},
		&models.OAuthClient{},
		&models.TwoFactorAuth{},
		&models.AuditLog{},
		&models.SystemConfig{},
	)
	require.NoError(t, err, "Failed to migrate test database")

	// Get underlying SQL DB
	sqlDB, err := gormDB.DB()
	require.NoError(t, err, "Failed to get underlying SQL DB")

	// Wrap in database.DB
	return &database.DB{
		DB:  gormDB,
		SQL: sqlDB,
	}
}

// TeardownTestDB closes the database connection
func TeardownTestDB(t *testing.T, db *database.DB) {
	t.Helper()

	err := db.Close()
	if err != nil {
		t.Logf("Warning: Failed to close test database: %v", err)
	}
}

// CleanupDB truncates all tables for a fresh state between tests
func CleanupDB(t *testing.T, db *database.DB) {
	t.Helper()

	// Delete all records from tables (order matters for foreign keys)
	tables := []interface{}{
		&models.AuditLog{},
		&models.Session{},
		&models.PasswordResetToken{},
		&models.PasswordHistory{},
		&models.TwoFactorAuth{},
		&models.User{},
		&models.OAuthClient{},
		&models.Role{},
		&models.Permission{},
		&models.SystemConfig{},
	}

	for _, table := range tables {
		err := db.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error
		require.NoError(t, err, "Failed to cleanup table")
	}
}
