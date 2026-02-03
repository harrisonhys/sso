package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sso-project/sso-server/internal/database"
	"github.com/sso-project/sso-server/internal/models"
	"gorm.io/gorm"
)

// ConfigRepository handles system configuration operations
type ConfigRepository struct {
	db *database.DB
}

// NewConfigRepository creates a new config repository
func NewConfigRepository(db *database.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// GetByKey retrieves a config by key
func (r *ConfigRepository) GetByKey(ctx context.Context, key string) (*models.SystemConfig, error) {
	var config models.SystemConfig
	result := r.db.WithContext(ctx).
		First(&config, "config_key = ?", key)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &config, nil
}

// CreateOrUpdate creates or updates a config
func (r *ConfigRepository) CreateOrUpdate(ctx context.Context, key, value string) (*models.SystemConfig, error) {
	var config models.SystemConfig

	// Check if exists
	err := r.db.WithContext(ctx).First(&config, "config_key = ?", key).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new
			config = models.SystemConfig{
				ID:          uuid.New().String(),
				ConfigKey:   key,
				ConfigValue: value,
			}
			if err := r.db.WithContext(ctx).Create(&config).Error; err != nil {
				return nil, err
			}
			return &config, nil
		}
		return nil, err
	}

	// Update existing
	config.ConfigValue = value
	if err := r.db.WithContext(ctx).Save(&config).Error; err != nil {
		return nil, err
	}

	return &config, nil
}

// List retrieves all configs
func (r *ConfigRepository) List(ctx context.Context) ([]*models.SystemConfig, error) {
	var configs []*models.SystemConfig

	result := r.db.WithContext(ctx).
		Order("config_key ASC").
		Find(&configs)

	if result.Error != nil {
		return nil, result.Error
	}

	return configs, nil
}

// GetDB returns the underlying GORM DB instance
func (r *ConfigRepository) GetDB() *gorm.DB {
	return r.db.DB
}
