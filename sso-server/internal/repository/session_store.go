package repository

import (
	"context"

	"github.com/sso-project/sso-server/internal/models"
)

// SessionStore defines the interface for session storage operations
// allowing switching between different storage backends (Database, Redis)
type SessionStore interface {
	// Create creates a new session
	Create(ctx context.Context, session *models.Session) error

	// GetByToken retrieves a session by token
	GetByToken(ctx context.Context, token string) (*models.Session, error)

	// GetByUserID retrieves all sessions for a user
	GetByUserID(ctx context.Context, userID string) ([]*models.Session, error)

	// Update updates a session
	Update(ctx context.Context, session *models.Session) error

	// Delete deletes a session
	Delete(ctx context.Context, id string) error

	// DeleteByToken deletes a session by token
	DeleteByToken(ctx context.Context, token string) error

	// DeleteByUserID deletes all sessions for a user
	DeleteByUserID(ctx context.Context, userID string) error

	// DeleteExpired deletes all expired sessions
	DeleteExpired(ctx context.Context) error

	// CleanupExpired is a background task to clean up expired sessions (if needed)
	CleanupExpired(ctx context.Context) error
}
