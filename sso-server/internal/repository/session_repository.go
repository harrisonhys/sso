package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sso-project/sso-server/internal/database"
	"github.com/sso-project/sso-server/internal/models"
	"gorm.io/gorm"
)

// DatabaseSessionStore handles session data operations using Database
type DatabaseSessionStore struct {
	db *database.DB
}

// NewDatabaseSessionStore creates a new database session repository
func NewDatabaseSessionStore(db *database.DB) *DatabaseSessionStore {
	return &DatabaseSessionStore{db: db}
}

// Create creates a new session
func (r *DatabaseSessionStore) Create(ctx context.Context, session *models.Session) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(session).Error
}

// GetByToken retrieves a session by token
func (r *DatabaseSessionStore) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).
		Preload("User").
		First(&session, "session_token = ?", token).Error

	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetByUserID retrieves all sessions for a user
func (r *DatabaseSessionStore) GetByUserID(ctx context.Context, userID string) ([]*models.Session, error) {
	var sessions []*models.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&sessions).Error

	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// Update updates a session
func (r *DatabaseSessionStore) Update(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// Delete deletes a session
func (r *DatabaseSessionStore) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Session{}, "id = ?", id).Error
}

// DeleteByToken deletes a session by token
func (r *DatabaseSessionStore) DeleteByToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Delete(&models.Session{}, "session_token = ?", token).Error
}

// DeleteByUserID deletes all sessions for a user
func (r *DatabaseSessionStore) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Delete(&models.Session{}, "user_id = ?", userID).Error
}

// DeleteExpired deletes all expired sessions
func (r *DatabaseSessionStore) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < NOW()").
		Delete(&models.Session{}).Error
}

// CleanupExpired is a background task to clean up expired sessions
func (r *DatabaseSessionStore) CleanupExpired(ctx context.Context) error {
	return r.DeleteExpired(ctx)
}

// GetDB returns the underlying GORM DB instance
func (r *DatabaseSessionStore) GetDB() *gorm.DB {
	return r.db.DB
}
