package repository

import (
	"context"
	"errors"

	"github.com/sso-project/sso-server/internal/models"
	"gorm.io/gorm"
)

var (
	ErrConsentNotFound = errors.New("user consent not found")
)

// OAuth2ConsentRepository handles OAuth2 user consent persistence
type OAuth2ConsentRepository struct {
	db *gorm.DB
}

// NewOAuth2ConsentRepository creates a new OAuth2ConsentRepository
func NewOAuth2ConsentRepository(db *gorm.DB) *OAuth2ConsentRepository {
	return &OAuth2ConsentRepository{db: db}
}

// Create stores a new user consent
func (r *OAuth2ConsentRepository) Create(ctx context.Context, consent *models.OAuth2Consent) error {
	return r.db.WithContext(ctx).Create(consent).Error
}

// GetByUserAndClient retrieves consent for a specific user-client combination
func (r *OAuth2ConsentRepository) GetByUserAndClient(ctx context.Context, userID, clientID string) (*models.OAuth2Consent, error) {
	var consent models.OAuth2Consent
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND client_id = ?", userID, clientID).
		First(&consent).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConsentNotFound
		}
		return nil, err
	}

	return &consent, nil
}

// Update updates an existing consent
func (r *OAuth2ConsentRepository) Update(ctx context.Context, consent *models.OAuth2Consent) error {
	return r.db.WithContext(ctx).Save(consent).Error
}

// Upsert creates or updates consent
func (r *OAuth2ConsentRepository) Upsert(ctx context.Context, consent *models.OAuth2Consent) error {
	// Try to get existing consent
	existing, err := r.GetByUserAndClient(ctx, consent.UserID, consent.ClientID)
	if err != nil && !errors.Is(err, ErrConsentNotFound) {
		return err
	}

	if existing != nil {
		// Update existing
		existing.Scopes = consent.Scopes
		return r.Update(ctx, existing)
	}

	// Create new
	return r.Create(ctx, consent)
}

// Revoke removes a user's consent for a client
func (r *OAuth2ConsentRepository) Revoke(ctx context.Context, userID, clientID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND client_id = ?", userID, clientID).
		Delete(&models.OAuth2Consent{}).Error
}

// RevokeAll removes all consents for a user
func (r *OAuth2ConsentRepository) RevokeAll(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.OAuth2Consent{}).Error
}

// GetByUserID retrieves all consents for a user
func (r *OAuth2ConsentRepository) GetByUserID(ctx context.Context, userID string) ([]*models.OAuth2Consent, error) {
	var consents []*models.OAuth2Consent
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("granted_at DESC").
		Find(&consents).Error

	return consents, err
}

// CheckConsent verifies if a user has granted consent for specific scopes
func (r *OAuth2ConsentRepository) CheckConsent(ctx context.Context, userID, clientID string, requestedScopes []string) (bool, []string, error) {
	consent, err := r.GetByUserAndClient(ctx, userID, clientID)
	if err != nil {
		if errors.Is(err, ErrConsentNotFound) {
			return false, nil, nil
		}
		return false, nil, err
	}

	// Create map of granted scopes
	grantedMap := make(map[string]bool)
	for _, scope := range consent.Scopes {
		grantedMap[scope] = true
	}

	// Check if all requested scopes are granted
	var missingScopes []string
	for _, requested := range requestedScopes {
		if !grantedMap[requested] {
			missingScopes = append(missingScopes, requested)
		}
	}

	if len(missingScopes) > 0 {
		return false, missingScopes, nil
	}

	return true, nil, nil
}
