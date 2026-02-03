package repository

import (
	"context"
	"errors"

	"github.com/sso-project/sso-server/internal/models"
	"gorm.io/gorm"
)

var (
	ErrScopeNotFound = errors.New("oauth2 scope not found")
)

// OAuth2ScopeRepository handles OAuth2 scope data persistence
type OAuth2ScopeRepository struct {
	db *gorm.DB
}

// NewOAuth2ScopeRepository creates a new OAuth2ScopeRepository
func NewOAuth2ScopeRepository(db *gorm.DB) *OAuth2ScopeRepository {
	return &OAuth2ScopeRepository{db: db}
}

// Create creates a new scope
func (r *OAuth2ScopeRepository) Create(ctx context.Context, scope *models.OAuth2Scope) error {
	return r.db.WithContext(ctx).Create(scope).Error
}

// GetAll retrieves all scopes
func (r *OAuth2ScopeRepository) GetAll(ctx context.Context) ([]*models.OAuth2Scope, error) {
	var scopes []*models.OAuth2Scope
	err := r.db.WithContext(ctx).
		Order("name ASC").
		Find(&scopes).Error
	return scopes, err
}

// GetByName retrieves a scope by its name
func (r *OAuth2ScopeRepository) GetByName(ctx context.Context, name string) (*models.OAuth2Scope, error) {
	var scope models.OAuth2Scope
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&scope).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrScopeNotFound
		}
		return nil, err
	}

	return &scope, nil
}

// GetByNames retrieves multiple scopes by their names
func (r *OAuth2ScopeRepository) GetByNames(ctx context.Context, names []string) ([]*models.OAuth2Scope, error) {
	var scopes []*models.OAuth2Scope
	err := r.db.WithContext(ctx).
		Where("name IN ?", names).
		Find(&scopes).Error
	return scopes, err
}

// GetDefaultScopes retrieves all default scopes
func (r *OAuth2ScopeRepository) GetDefaultScopes(ctx context.Context) ([]*models.OAuth2Scope, error) {
	var scopes []*models.OAuth2Scope
	err := r.db.WithContext(ctx).
		Where("is_default = ?", true).
		Find(&scopes).Error
	return scopes, err
}

// ValidateScopes checks if all requested scopes exist and are valid
func (r *OAuth2ScopeRepository) ValidateScopes(ctx context.Context, requestedScopes []string) (bool, error) {
	if len(requestedScopes) == 0 {
		return true, nil
	}

	scopes, err := r.GetByNames(ctx, requestedScopes)
	if err != nil {
		return false, err
	}

	// Check if we found all requested scopes
	if len(scopes) != len(requestedScopes) {
		return false, nil
	}

	return true, nil
}
