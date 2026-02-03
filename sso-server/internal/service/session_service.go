package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
	"github.com/sso-project/sso-server/internal/utils"
)

// SessionService handles session operations
type SessionService struct {
	sessionRepo repository.SessionStore
	timeout     time.Duration
}

// NewSessionService creates a new session service
func NewSessionService(sessionRepo repository.SessionStore, timeout time.Duration) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		timeout:     timeout,
	}
}

// CreateSession creates a new session for a user
func (s *SessionService) CreateSession(ctx context.Context, userID, ipAddress, userAgent string) (*models.Session, error) {
	// Generate secure session token
	sessionToken, err := utils.GenerateRandomToken(64)
	if err != nil {
		return nil, err
	}

	session := &models.Session{
		ID:             uuid.New().String(),
		UserID:         userID,
		SessionToken:   sessionToken,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		ExpiresAt:      time.Now().Add(s.timeout),
		LastActivityAt: time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// ValidateSession validates a session token and returns the session
func (s *SessionService) ValidateSession(ctx context.Context, token string) (*models.Session, error) {
	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		s.sessionRepo.Delete(ctx, session.ID)
		return nil, ErrExpiredToken
	}

	return session, nil
}

// RenewSession renews a session (sliding window)
func (s *SessionService) RenewSession(ctx context.Context, token string) (*models.Session, error) {
	session, err := s.ValidateSession(ctx, token)
	if err != nil {
		return nil, err
	}

	// Update expiry and last activity
	session.ExpiresAt = time.Now().Add(s.timeout)
	session.LastActivityAt = time.Now()

	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// TerminateSession terminates a specific session
func (s *SessionService) TerminateSession(ctx context.Context, token string) error {
	return s.sessionRepo.DeleteByToken(ctx, token)
}

// TerminateUserSessions terminates all sessions for a user
func (s *SessionService) TerminateUserSessions(ctx context.Context, userID string) error {
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

// GetUserSessions retrieves all active sessions for a user
func (s *SessionService) GetUserSessions(ctx context.Context, userID string) ([]*models.Session, error) {
	sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Filter out expired sessions
	activeSessions := make([]*models.Session, 0)
	now := time.Now()

	for _, session := range sessions {
		if session.ExpiresAt.After(now) {
			activeSessions = append(activeSessions, session)
		} else {
			// Clean up expired session
			s.sessionRepo.Delete(ctx, session.ID)
		}
	}

	return activeSessions, nil
}

// CleanupExpiredSessions removes all expired sessions (background job)
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	return s.sessionRepo.DeleteExpired(ctx)
}
