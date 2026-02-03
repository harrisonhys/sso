package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sso-project/sso-server/internal/models"
	"github.com/sso-project/sso-server/internal/repository"
	"github.com/sso-project/sso-server/internal/testutil"
)

func TestUserRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &models.User{
			Email:        "test@example.com",
			Name:         "Test User",
			PasswordHash: "hashed_password",
			IsActive:     true,
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotEmpty(t, user.ID)

		// Verify in DB
		savedUser, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Email, savedUser.Email)
	})

	t.Run("duplicate_email", func(t *testing.T) {
		testutil.CleanupDB(t, db)

		user1 := &models.User{Email: "dup@example.com", Name: "U1", PasswordHash: "pw"}
		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		user2 := &models.User{Email: "dup@example.com", Name: "U2", PasswordHash: "pw"}
		err = repo.Create(ctx, user2)
		require.Error(t, err)
		assert.Equal(t, repository.ErrAlreadyExists, err)
	})
}

func TestUserRepository_Get(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		Email:        "get@example.com",
		Name:         "Get User",
		PasswordHash: "pw",
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("get_by_id_success", func(t *testing.T) {
		found, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("get_by_id_not_found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "non-existent")
		require.Error(t, err)
		assert.Equal(t, repository.ErrNotFound, err)
	})

	t.Run("get_by_email_success", func(t *testing.T) {
		found, err := repo.GetByEmail(ctx, user.Email)
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
	})

	t.Run("get_by_email_not_found", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "missing@example.com")
		require.Error(t, err)
		assert.Equal(t, repository.ErrNotFound, err)
	})
}

func TestUserRepository_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "update@example.com", Name: "Old Name"}
	require.NoError(t, repo.Create(ctx, user))

	t.Run("update_success", func(t *testing.T) {
		user.Name = "New Name"
		err := repo.Update(ctx, user)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "New Name", updated.Name)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "delete@example.com"}
	require.NoError(t, repo.Create(ctx, user))

	t.Run("delete_success", func(t *testing.T) {
		err := repo.Delete(ctx, user.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, user.ID)
		assert.Equal(t, repository.ErrNotFound, err)
	})

	t.Run("delete_not_found", func(t *testing.T) {
		err := repo.Delete(ctx, "non-existent")
		require.Equal(t, repository.ErrNotFound, err)
	})
}

func TestUserRepository_ListWithFilters(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Seed users
	names := []string{"Alpha", "Beta", "Charlie", "Delta"}
	for _, name := range names {
		repo.Create(ctx, &models.User{
			Email: name + "@example.com",
			Name:  name,
		})
	}

	t.Run("list_all", func(t *testing.T) {
		users, total, err := repo.ListWithFilters(ctx, 0, 10, "")
		require.NoError(t, err)
		assert.Equal(t, int64(4), total)
		assert.Len(t, users, 4)
	})

	t.Run("search_filter", func(t *testing.T) {
		users, total, err := repo.ListWithFilters(ctx, 0, 10, "Alph")
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "Alpha", users[0].Name)
	})
}

func TestUserRepository_AccountLocking(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{Email: "lock@example.com"}
	require.NoError(t, repo.Create(ctx, user))

	t.Run("increment_failed_attempts", func(t *testing.T) {
		err := repo.IncrementFailedAttempts(ctx, user.ID)
		require.NoError(t, err)

		u, _ := repo.GetByID(ctx, user.ID)
		assert.Equal(t, 1, u.FailedLoginAttempts)
	})

	t.Run("unlock_account", func(t *testing.T) {
		// Manually lock
		lockTime := time.Now().Add(time.Hour)
		db.DB.Model(user).Updates(map[string]interface{}{
			"locked_until":          lockTime,
			"failed_login_attempts": 5,
		})

		err := repo.UnlockAccount(ctx, user.ID)
		require.NoError(t, err)

		u, _ := repo.GetByID(ctx, user.ID)
		assert.Equal(t, 0, u.FailedLoginAttempts)
		assert.Nil(t, u.LockedUntil)
	})
}
