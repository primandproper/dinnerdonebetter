package authentication

import (
	"context"
	"errors"
	"testing"
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/fakes"
	oauthmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/mock"

	loggingnoop "github.com/primandproper/platform-go/v11/observability/logging/noop"
	"github.com/primandproper/platform-go/v11/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuth2TokenStoreImpl_Create(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		token := fakes.BuildFakeOAuth2ClientToken()
		tokenInfo := convertTokenToImpl(token)

		dataManager := &oauthmock.RepositoryMock{
			CreateOAuth2ClientTokenFunc: func(_ context.Context, input *types.OAuth2ClientTokenDatabaseCreationInput) (*types.OAuth2ClientToken, error) {
				assert.NotNil(t, input)
				return token, nil
			},
		}

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.Create(ctx, tokenInfo)

		require.NoError(t, err)
		// Verify that the expiration times were set
		assert.Equal(t, 24*time.Hour, tokenInfo.GetAccessExpiresIn())
		assert.Equal(t, 72*time.Hour, tokenInfo.GetRefreshExpiresIn())

		assert.Len(t, dataManager.CreateOAuth2ClientTokenCalls(), 1)
	})

	T.Run("with database error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		token := fakes.BuildFakeOAuth2ClientToken()
		tokenInfo := convertTokenToImpl(token)

		dataManager := &oauthmock.RepositoryMock{
			CreateOAuth2ClientTokenFunc: func(_ context.Context, input *types.OAuth2ClientTokenDatabaseCreationInput) (*types.OAuth2ClientToken, error) {
				assert.NotNil(t, input)
				return nil, errors.New("database error")
			},
		}

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.Create(ctx, tokenInfo)

		require.Error(t, err)

		assert.Len(t, dataManager.CreateOAuth2ClientTokenCalls(), 1)
	})
}

func TestOAuth2TokenStoreImpl_RemoveByCode(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		code := "test-code"

		dataManager := &oauthmock.RepositoryMock{
			DeleteOAuth2ClientTokenByCodeFunc: func(_ context.Context, actual string) error {
				assert.Equal(t, code, actual)
				return nil
			},
		}

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.RemoveByCode(ctx, code)

		require.NoError(t, err)

		assert.Len(t, dataManager.DeleteOAuth2ClientTokenByCodeCalls(), 1)
	})

	T.Run("with database error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		code := "test-code"

		dataManager := &oauthmock.RepositoryMock{
			DeleteOAuth2ClientTokenByCodeFunc: func(_ context.Context, actual string) error {
				assert.Equal(t, code, actual)
				return errors.New("database error")
			},
		}

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.RemoveByCode(ctx, code)

		require.Error(t, err)

		assert.Len(t, dataManager.DeleteOAuth2ClientTokenByCodeCalls(), 1)
	})
}

func TestOAuth2TokenStoreImpl_RemoveByAccess(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		access := "test-access"

		dataManager := &oauthmock.RepositoryMock{
			DeleteOAuth2ClientTokenByAccessFunc: func(_ context.Context, actual string) error {
				assert.Equal(t, access, actual)
				return nil
			},
		}

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.RemoveByAccess(ctx, access)

		require.NoError(t, err)

		assert.Len(t, dataManager.DeleteOAuth2ClientTokenByAccessCalls(), 1)
	})

	T.Run("with database error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		access := "test-access"

		dataManager := &oauthmock.RepositoryMock{
			DeleteOAuth2ClientTokenByAccessFunc: func(_ context.Context, actual string) error {
				assert.Equal(t, access, actual)
				return errors.New("database error")
			},
		}

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.RemoveByAccess(ctx, access)

		require.Error(t, err)

		assert.Len(t, dataManager.DeleteOAuth2ClientTokenByAccessCalls(), 1)
	})
}

func TestOAuth2TokenStoreImpl_RemoveByRefresh(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		refresh := "test-refresh"

		dataManager := &oauthmock.RepositoryMock{
			DeleteOAuth2ClientTokenByRefreshFunc: func(_ context.Context, actual string) error {
				assert.Equal(t, refresh, actual)
				return nil
			},
		}

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.RemoveByRefresh(ctx, refresh)

		require.NoError(t, err)

		assert.Len(t, dataManager.DeleteOAuth2ClientTokenByRefreshCalls(), 1)
	})

	T.Run("with database error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		refresh := "test-refresh"

		dataManager := &oauthmock.RepositoryMock{
			DeleteOAuth2ClientTokenByRefreshFunc: func(_ context.Context, actual string) error {
				assert.Equal(t, refresh, actual)
				return errors.New("database error")
			},
		}

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.RemoveByRefresh(ctx, refresh)

		require.Error(t, err)

		assert.Len(t, dataManager.DeleteOAuth2ClientTokenByRefreshCalls(), 1)
	})
}

func TestOAuth2TokenStoreImpl_GetByCode(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		token := fakes.BuildFakeOAuth2ClientToken()

		dataManager := &oauthmock.RepositoryMock{
			GetOAuth2ClientTokenByCodeFunc: func(_ context.Context, actual string) (*types.OAuth2ClientToken, error) {
				assert.Equal(t, token.Code, actual)
				return token, nil
			},
		}

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		result, err := store.GetByCode(ctx, token.Code)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, token.ClientID, result.GetClientID())
		assert.Equal(t, token.Code, result.GetCode())

		assert.Len(t, dataManager.GetOAuth2ClientTokenByCodeCalls(), 1)
	})

	T.Run("with database error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		code := "test-code"

		dataManager := &oauthmock.RepositoryMock{
			GetOAuth2ClientTokenByCodeFunc: func(_ context.Context, actual string) (*types.OAuth2ClientToken, error) {
				assert.Equal(t, code, actual)
				return nil, errors.New("database error")
			},
		}

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		result, err := store.GetByCode(ctx, code)

		require.Error(t, err)
		assert.Nil(t, result)

		assert.Len(t, dataManager.GetOAuth2ClientTokenByCodeCalls(), 1)
	})
}

func TestOAuth2TokenStoreImpl_GetByAccess(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		token := fakes.BuildFakeOAuth2ClientToken()

		dataManager := &oauthmock.RepositoryMock{
			GetOAuth2ClientTokenByAccessFunc: func(_ context.Context, actual string) (*types.OAuth2ClientToken, error) {
				assert.Equal(t, token.Access, actual)
				return token, nil
			},
		}

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		result, err := store.GetByAccess(ctx, token.Access)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, token.ClientID, result.GetClientID())
		assert.Equal(t, token.Access, result.GetAccess())

		assert.Len(t, dataManager.GetOAuth2ClientTokenByAccessCalls(), 1)
	})

	T.Run("with database error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		access := "test-access"

		dataManager := &oauthmock.RepositoryMock{
			GetOAuth2ClientTokenByAccessFunc: func(_ context.Context, actual string) (*types.OAuth2ClientToken, error) {
				assert.Equal(t, access, actual)
				return nil, errors.New("database error")
			},
		}

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		result, err := store.GetByAccess(ctx, access)

		require.Error(t, err)
		assert.Nil(t, result)

		assert.Len(t, dataManager.GetOAuth2ClientTokenByAccessCalls(), 1)
	})
}

func TestOAuth2TokenStoreImpl_GetByRefresh(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		token := fakes.BuildFakeOAuth2ClientToken()

		dataManager := &oauthmock.RepositoryMock{
			GetOAuth2ClientTokenByRefreshFunc: func(_ context.Context, actual string) (*types.OAuth2ClientToken, error) {
				assert.Equal(t, token.Refresh, actual)
				return token, nil
			},
		}

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		result, err := store.GetByRefresh(ctx, token.Refresh)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, token.ClientID, result.GetClientID())
		assert.Equal(t, token.Refresh, result.GetRefresh())

		assert.Len(t, dataManager.GetOAuth2ClientTokenByRefreshCalls(), 1)
	})

	T.Run("with database error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		refresh := "test-refresh"

		dataManager := &oauthmock.RepositoryMock{
			GetOAuth2ClientTokenByRefreshFunc: func(_ context.Context, actual string) (*types.OAuth2ClientToken, error) {
				assert.Equal(t, refresh, actual)
				return nil, errors.New("database error")
			},
		}

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		result, err := store.GetByRefresh(ctx, refresh)

		require.Error(t, err)
		assert.Nil(t, result)

		assert.Len(t, dataManager.GetOAuth2ClientTokenByRefreshCalls(), 1)
	})
}
