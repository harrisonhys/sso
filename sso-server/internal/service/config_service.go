package service

import (
	"context"

	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
)

// ConfigService handles business logic for system configuration
type ConfigService struct {
	configRepo *repository.ConfigRepository
}

// NewConfigService creates a new config service
func NewConfigService(configRepo *repository.ConfigRepository) *ConfigService {
	return &ConfigService{
		configRepo: configRepo,
	}
}

// GetConfig retrieves a configuration by key
func (s *ConfigService) GetConfig(ctx context.Context, key string) (*models.SystemConfig, error) {
	return s.configRepo.GetByKey(ctx, key)
}

// UpdateConfig updates a configuration
func (s *ConfigService) UpdateConfig(ctx context.Context, key, value string) (*models.SystemConfig, error) {
	return s.configRepo.CreateOrUpdate(ctx, key, value)
}

// GetAllConfigs retrieves all configurations
func (s *ConfigService) GetAllConfigs(ctx context.Context) ([]*models.SystemConfig, error) {
	return s.configRepo.List(ctx)
}
