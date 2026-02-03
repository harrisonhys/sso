package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sso-project/sso-server/internal/database"
	"github.com/sso-project/sso-server/internal/models"
	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyExists = errors.New("record already exists")
	ErrInvalidInput  = errors.New("invalid input")
)

// UserRepository handles user data operations
type UserRepository struct {
	db *database.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	result := r.db.DB.WithContext(ctx).Create(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) || strings.Contains(result.Error.Error(), "UNIQUE constraint failed") {
			return ErrAlreadyExists
		}
		return result.Error
	}
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	result := r.db.DB.WithContext(ctx).
		Preload("Roles.Permissions").
		Preload("TwoFactorAuth").
		First(&user, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := r.db.DB.WithContext(ctx).
		Preload("Roles.Permissions").
		Preload("TwoFactorAuth").
		First(&user, "email = ?", email)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	result := r.db.DB.WithContext(ctx).Save(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes a user (soft delete)
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	result := r.db.DB.WithContext(ctx).Delete(&models.User{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves users with pagination
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// Count total
	r.db.DB.WithContext(ctx).Model(&models.User{}).Count(&total)

	// Get paginated results
	result := r.db.DB.WithContext(ctx).
		Preload("Roles").
		Offset(offset).
		Limit(limit).
		Find(&users)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return users, total, nil
}

// AssignRole assigns a role to a user
func (r *UserRepository) AssignRole(ctx context.Context, userID, roleID string) error {
	return r.db.DB.WithContext(ctx).Exec(
		"INSERT IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)",
		userID, roleID,
	).Error
}

// RemoveRole removes a role from a user
func (r *UserRepository) RemoveRole(ctx context.Context, userID, roleID string) error {
	return r.db.DB.WithContext(ctx).Exec(
		"DELETE FROM user_roles WHERE user_id = ? AND role_id = ?",
		userID, roleID,
	).Error
}

// IncrementFailedAttempts increments failed login attempts
func (r *UserRepository) IncrementFailedAttempts(ctx context.Context, id string) error {
	return r.db.DB.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		UpdateColumn("failed_login_attempts", gorm.Expr("failed_login_attempts + ?", 1)).
		Error
}

// ResetFailedAttempts resets failed login attempts to 0
func (r *UserRepository) ResetFailedAttempts(ctx context.Context, id string) error {
	result := r.db.DB.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		UpdateColumn("failed_login_attempts", 0)

	if result.Error != nil {
		return result.Error
	}

	// Log for debugging
	if result.RowsAffected == 0 {
		return errors.New("no rows affected - user not found")
	}

	return nil
}

// GetDB returns the underlying GORM DB instance
func (r *UserRepository) GetDB() *gorm.DB {
	return r.db.DB
}

// ListWithFilters retrieves users with pagination and filters
func (r *UserRepository) ListWithFilters(ctx context.Context, offset, limit int, search string) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	query := r.db.DB.WithContext(ctx).Model(&models.User{})

	// Apply search filter
	if search != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Count total
	query.Count(&total)

	// Get paginated results
	result := query.
		Preload("Roles").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return users, total, nil
}

// CountUsers counts total users with optional filter
func (r *UserRepository) CountUsers(ctx context.Context, search string) (int64, error) {
	var count int64
	query := r.db.DB.WithContext(ctx).Model(&models.User{})

	if search != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// UpdateUserStatus updates user active status
func (r *UserRepository) UpdateUserStatus(ctx context.Context, userID string, isActive bool) error {
	return r.db.DB.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("is_active", isActive).
		Error
}

// UnlockAccount unlocks a locked user account
func (r *UserRepository) UnlockAccount(ctx context.Context, userID string) error {
	return r.db.DB.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"failed_login_attempts": 0,
			"locked_until":          nil,
		}).
		Error
}

// UpdateLastLogin updates only the last_login_at field
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string, lastLoginAt *time.Time) error {
	return r.db.DB.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		UpdateColumn("last_login_at", lastLoginAt).
		Error
}
