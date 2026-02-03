package service

import (
	"context"
	"errors"

	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
)

// OAuth2ConsentService handles user consent management
type OAuth2ConsentService struct {
	consentRepo *repository.OAuth2ConsentRepository
	clientRepo  *repository.OAuth2ClientRepository
	scopeRepo   *repository.OAuth2ScopeRepository
}

// NewOAuth2ConsentService creates a new OAuth2ConsentService
func NewOAuth2ConsentService(
	consentRepo *repository.OAuth2ConsentRepository,
	clientRepo *repository.OAuth2ClientRepository,
	scopeRepo *repository.OAuth2ScopeRepository,
) *OAuth2ConsentService {
	return &OAuth2ConsentService{
		consentRepo: consentRepo,
		clientRepo:  clientRepo,
		scopeRepo:   scopeRepo,
	}
}

// GrantConsent stores or updates user consent for a client
func (s *OAuth2ConsentService) GrantConsent(ctx context.Context, userID, clientID string, scopes []string) error {
	// Validate client exists
	_, err := s.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return err
	}

	// Validate scopes
	if len(scopes) > 0 {
		valid, err := s.scopeRepo.ValidateScopes(ctx, scopes)
		if err != nil {
			return err
		}
		if !valid {
			return errors.New("invalid scopes")
		}
	}

	// Upsert consent
	consent := &models.OAuth2Consent{
		UserID:   userID,
		ClientID: clientID,
		Scopes:   scopes,
	}

	return s.consentRepo.Upsert(ctx, consent)
}

// CheckConsent validates if user has granted consent for requested scopes
func (s *OAuth2ConsentService) CheckConsent(ctx context.Context, userID, clientID string, requestedScopes []string) (bool, []string, error) {
	return s.consentRepo.CheckConsent(ctx, userID, clientID, requestedScopes)
}

// GetUserConsents retrieves all consents for a user
func (s *OAuth2ConsentService) GetUserConsents(ctx context.Context, userID string) ([]*models.OAuth2Consent, error) {
	return s.consentRepo.GetByUserID(ctx, userID)
}

// RevokeConsent removes a user's consent for a client
func (s *OAuth2ConsentService) RevokeConsent(ctx context.Context, userID, clientID string) error {
	return s.consentRepo.Revoke(ctx, userID, clientID)
}

// RevokeAllConsents removes all consents for a user
func (s *OAuth2ConsentService) RevokeAllConsents(ctx context.Context, userID string) error {
	return s.consentRepo.RevokeAll(ctx, userID)
}

// GetConsentWithClientDetails retrieves consent with enriched client information
func (s *OAuth2ConsentService) GetConsentWithClientDetails(ctx context.Context, userID, clientID string) (map[string]interface{}, error) {
	// Get consent
	consent, err := s.consentRepo.GetByUserAndClient(ctx, userID, clientID)
	if err != nil {
		if errors.Is(err, repository.ErrConsentNotFound) {
			return nil, nil // No consent exists
		}
		return nil, err
	}

	// Get client details
	client, err := s.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}

	// Get scope details
	scopeDetails, err := s.scopeRepo.GetByNames(ctx, consent.Scopes)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"consent": consent,
		"client":  client,
		"scopes":  scopeDetails,
	}, nil
}
