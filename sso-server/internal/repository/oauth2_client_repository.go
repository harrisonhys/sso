package repository

import (
	"context"
	"errors"

	"github.com/sso-project/sso-server/internal/models"
	"gorm.io/gorm"
)

var (
	ErrClientNotFound      = errors.New("oauth2 client not found")
	ErrInvalidClientSecret = errors.New("invalid client secret")
)

// OAuth2ClientRepository handles OAuth2 client data persistence
type OAuth2ClientRepository struct {
	db *gorm.DB
}

// NewOAuth2ClientRepository creates a new OAuth2ClientRepository
func NewOAuth2ClientRepository(db *gorm.DB) *OAuth2ClientRepository {
	return &OAuth2ClientRepository{db: db}
}

// Create creates a new OAuth2 client
func (r *OAuth2ClientRepository) Create(ctx context.Context, client *models.OAuth2Client) error {
	return r.db.WithContext(ctx).Create(client).Error
}

// GetByClientID retrieves a client by its client_id
func (r *OAuth2ClientRepository) GetByClientID(ctx context.Context, clientID string) (*models.OAuth2Client, error) {
	var client models.OAuth2Client
	err := r.db.WithContext(ctx).
		Where("client_id = ?", clientID).
		First(&client).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClientNotFound
		}
		return nil, err
	}

	return &client, nil
}

// GetByID retrieves a client by its internal ID
func (r *OAuth2ClientRepository) GetByID(ctx context.Context, id string) (*models.OAuth2Client, error) {
	var client models.OAuth2Client
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&client).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClientNotFound
		}
		return nil, err
	}

	return &client, nil
}

// Update updates an existing client
func (r *OAuth2ClientRepository) Update(ctx context.Context, client *models.OAuth2Client) error {
	return r.db.WithContext(ctx).Save(client).Error
}

// Delete soft-deletes a client by setting is_active to false
func (r *OAuth2ClientRepository) Delete(ctx context.Context, clientID string) error {
	return r.db.WithContext(ctx).
		Model(&models.OAuth2Client{}).
		Where("client_id = ?", clientID).
		Update("is_active", false).Error
}

// GetByOwnerID retrieves all clients owned by a specific user
func (r *OAuth2ClientRepository) GetByOwnerID(ctx context.Context, ownerID string) ([]*models.OAuth2Client, error) {
	var clients []*models.OAuth2Client
	err := r.db.WithContext(ctx).
		Where("owner_user_id = ? AND is_active = ?", ownerID, true).
		Order("created_at DESC").
		Find(&clients).Error

	return clients, err
}

// GetAll retrieves all active clients (for admin purposes)
func (r *OAuth2ClientRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.OAuth2Client, int64, error) {
	var clients []*models.OAuth2Client
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).
		Model(&models.OAuth2Client{}).
		Where("is_active = ?", true).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch page
	query := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Find(&clients).Error
	return clients, total, err
}

// ValidateRedirectURI checks if a redirect URI is valid for the client
func (r *OAuth2ClientRepository) ValidateRedirectURI(ctx context.Context, clientID, redirectURI string) (bool, error) {
	client, err := r.GetByClientID(ctx, clientID)
	if err != nil {
		return false, err
	}

	// Check if redirect URI is in the list of allowed URIs
	for _, uri := range client.RedirectURIs {
		if uri == redirectURI {
			return true, nil
		}
	}

	return false, nil
}

// ValidateScopes checks if requested scopes are allowed for the client
func (r *OAuth2ClientRepository) ValidateScopes(ctx context.Context, clientID string, requestedScopes []string) (bool, error) {
	client, err := r.GetByClientID(ctx, clientID)
	if err != nil {
		return false, err
	}

	// If no specific scopes are allowed, deny all
	if len(client.AllowedScopes) == 0 {
		return false, nil
	}

	// Check if all requested scopes are in the allowed list
	allowedMap := make(map[string]bool)
	for _, scope := range client.AllowedScopes {
		allowedMap[scope] = true
	}

	for _, requested := range requestedScopes {
		if !allowedMap[requested] {
			return false, nil
		}
	}

	return true, nil
}

// GetDB returns the underlying GORM DB instance
func (r *OAuth2ClientRepository) GetDB() *gorm.DB {
	return r.db
}
