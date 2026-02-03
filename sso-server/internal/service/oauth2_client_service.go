package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/google/uuid"
	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
	"github.com/sso-project/sso-server/internal/utils"
)

var (
	ErrClientNotFound = errors.New("oauth2 client not found")
	ErrInvalidClient  = errors.New("invalid client credentials")
)

// OAuth2ClientService handles OAuth2 client business logic
type OAuth2ClientService struct {
	clientRepo *repository.OAuth2ClientRepository
	scopeRepo  *repository.OAuth2ScopeRepository
}

// NewOAuth2ClientService creates a new OAuth2ClientService
func NewOAuth2ClientService(
	clientRepo *repository.OAuth2ClientRepository,
	scopeRepo *repository.OAuth2ScopeRepository,
) *OAuth2ClientService {
	return &OAuth2ClientService{
		clientRepo: clientRepo,
		scopeRepo:  scopeRepo,
	}
}

// RegisterClientRequest represents a client registration request
type RegisterClientRequest struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	RedirectURIs  []string `json:"redirect_uris"`
	AllowedScopes []string `json:"allowed_scopes"`
	GrantTypes    []string `json:"grant_types"`
	IsPublic      bool     `json:"is_public"`
	OwnerUserID   *string  `json:"owner_user_id,omitempty"`
}

// RegisterClient creates a new OAuth2 client
func (s *OAuth2ClientService) RegisterClient(ctx context.Context, req RegisterClientRequest) (*models.OAuth2Client, string, error) {
	// Validate grant types
	validGrantTypes := map[string]bool{
		"authorization_code": true,
		"refresh_token":      true,
		"client_credentials": true,
	}

	for _, gt := range req.GrantTypes {
		if !validGrantTypes[gt] {
			return nil, "", errors.New("invalid grant type: " + gt)
		}
	}

	// Validate scopes exist
	if len(req.AllowedScopes) > 0 {
		valid, err := s.scopeRepo.ValidateScopes(ctx, req.AllowedScopes)
		if err != nil {
			return nil, "", err
		}
		if !valid {
			return nil, "", errors.New("one or more invalid scopes")
		}
	}

	// Generate client ID and secret
	clientID := s.generateClientID()
	clientSecret := s.generateClientSecret()

	// Hash the client secret
	hashedSecret, err := utils.HashPassword(clientSecret)
	if err != nil {
		return nil, "", err
	}

	client := &models.OAuth2Client{
		ID:            uuid.New().String(),
		ClientID:      clientID,
		ClientSecret:  hashedSecret,
		Name:          req.Name,
		Description:   req.Description,
		RedirectURIs:  req.RedirectURIs,
		AllowedScopes: req.AllowedScopes,
		GrantTypes:    req.GrantTypes,
		IsPublic:      req.IsPublic,
		IsActive:      true,
		OwnerUserID:   req.OwnerUserID,
	}

	if err := s.clientRepo.Create(ctx, client); err != nil {
		return nil, "", err
	}

	// Return client and the PLAIN secret (only time it's accessible)
	return client, clientSecret, nil
}

// ValidateClient validates client credentials
func (s *OAuth2ClientService) ValidateClient(ctx context.Context, clientID, clientSecret string) (*models.OAuth2Client, error) {
	client, err := s.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, ErrInvalidClient
	}

	if !client.IsActive {
		return nil, errors.New("client is not active")
	}

	// Public clients don't require secret validation
	if client.IsPublic {
		return client, nil
	}

	// Validate client secret
	if err := utils.ComparePassword(client.ClientSecret, clientSecret); err != nil {
		return nil, ErrInvalidClient
	}

	return client, nil
}

// GetClient retrieves a client by ID
func (s *OAuth2ClientService) GetClient(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	return s.clientRepo.GetByClientID(ctx, clientID)
}

// ListClients retrieves paginated list of clients
func (s *OAuth2ClientService) ListClients(ctx context.Context, page, limit int) ([]*models.OAuth2Client, int64, error) {
	offset := (page - 1) * limit
	return s.clientRepo.GetAll(ctx, limit, offset)
}

// RegenerateClientSecret generates a new client secret
func (s *OAuth2ClientService) RegenerateClientSecret(ctx context.Context, clientID string) (string, error) {
	client, err := s.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return "", err
	}

	// Generate new secret
	newSecret := s.generateClientSecret()
	hashedSecret, err := utils.HashPassword(newSecret)
	if err != nil {
		return "", err
	}

	client.ClientSecret = hashedSecret
	if err := s.clientRepo.Update(ctx, client); err != nil {
		return "", err
	}

	return newSecret, nil
}

// RevokeClient deactivates a client
func (s *OAuth2ClientService) RevokeClient(ctx context.Context, clientID string) error {
	return s.clientRepo.Delete(ctx, clientID)
}

// ValidateRedirectURI validates if a redirect URI is allowed for the client
func (s *OAuth2ClientService) ValidateRedirectURI(ctx context.Context, clientID, redirectURI string) (bool, error) {
	return s.clientRepo.ValidateRedirectURI(ctx, clientID, redirectURI)
}

// ValidateScopes validates if requested scopes are allowed for the client
func (s *OAuth2ClientService) ValidateScopes(ctx context.Context, clientID string, requestedScopes []string) (bool, error) {
	return s.clientRepo.ValidateScopes(ctx, clientID, requestedScopes)
}

// generateClientID generates a random client ID
func (s *OAuth2ClientService) generateClientID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// generateClientSecret generates a random client secret
func (s *OAuth2ClientService) generateClientSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
