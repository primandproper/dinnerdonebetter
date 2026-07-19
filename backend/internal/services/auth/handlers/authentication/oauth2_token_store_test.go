package authentication

import (
	"errors"
	"testing"
	"time"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/fakes"
	oauthmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	"github.com/primandproper/platform-go/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"CreateOAuth2ClientToken",
			testutils.ContextMatcher,
			testutils.MatchType[*types.OAuth2ClientTokenDatabaseCreationInput](),
		).Return(token, nil)

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.Create(ctx, tokenInfo)

		assert.NoError(t, err)
		// Verify that the expiration times were set
		assert.Equal(t, 24*time.Hour, tokenInfo.GetAccessExpiresIn())
		assert.Equal(t, 72*time.Hour, tokenInfo.GetRefreshExpiresIn())

		mock.AssertExpectationsForObjects(t, dataManager)
	})

	T.Run("with database error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		token := fakes.BuildFakeOAuth2ClientToken()
		tokenInfo := convertTokenToImpl(token)

		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"CreateOAuth2ClientToken",
			testutils.ContextMatcher,
			testutils.MatchType[*types.OAuth2ClientTokenDatabaseCreationInput](),
		).Return((*types.OAuth2ClientToken)(nil), errors.New("database error"))

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.Create(ctx, tokenInfo)

		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, dataManager)
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

		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"DeleteOAuth2ClientTokenByCode",
			testutils.ContextMatcher,
			code,
		).Return(nil)

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.RemoveByCode(ctx, code)

		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, dataManager)
	})

	T.Run("with database error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		code := "test-code"

		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"DeleteOAuth2ClientTokenByCode",
			testutils.ContextMatcher,
			code,
		).Return(errors.New("database error"))

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.RemoveByCode(ctx, code)

		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, dataManager)
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

		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"DeleteOAuth2ClientTokenByAccess",
			testutils.ContextMatcher,
			access,
		).Return(nil)

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.RemoveByAccess(ctx, access)

		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, dataManager)
	})

	T.Run("with database error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		access := "test-access"

		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"DeleteOAuth2ClientTokenByAccess",
			testutils.ContextMatcher,
			access,
		).Return(errors.New("database error"))

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.RemoveByAccess(ctx, access)

		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, dataManager)
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

		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"DeleteOAuth2ClientTokenByRefresh",
			testutils.ContextMatcher,
			refresh,
		).Return(nil)

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.RemoveByRefresh(ctx, refresh)

		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, dataManager)
	})

	T.Run("with database error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		refresh := "test-refresh"

		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"DeleteOAuth2ClientTokenByRefresh",
			testutils.ContextMatcher,
			refresh,
		).Return(errors.New("database error"))

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		err := store.RemoveByRefresh(ctx, refresh)

		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, dataManager)
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

		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"GetOAuth2ClientTokenByCode",
			testutils.ContextMatcher,
			token.Code,
		).Return(token, nil)

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		result, err := store.GetByCode(ctx, token.Code)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, token.ClientID, result.GetClientID())
		assert.Equal(t, token.Code, result.GetCode())

		mock.AssertExpectationsForObjects(t, dataManager)
	})

	T.Run("with database error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		code := "test-code"

		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"GetOAuth2ClientTokenByCode",
			testutils.ContextMatcher,
			code,
		).Return((*types.OAuth2ClientToken)(nil), errors.New("database error"))

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		result, err := store.GetByCode(ctx, code)

		assert.Error(t, err)
		assert.Nil(t, result)

		mock.AssertExpectationsForObjects(t, dataManager)
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

		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"GetOAuth2ClientTokenByAccess",
			testutils.ContextMatcher,
			token.Access,
		).Return(token, nil)

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		result, err := store.GetByAccess(ctx, token.Access)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, token.ClientID, result.GetClientID())
		assert.Equal(t, token.Access, result.GetAccess())

		mock.AssertExpectationsForObjects(t, dataManager)
	})

	T.Run("with database error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		access := "test-access"

		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"GetOAuth2ClientTokenByAccess",
			testutils.ContextMatcher,
			access,
		).Return((*types.OAuth2ClientToken)(nil), errors.New("database error"))

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		result, err := store.GetByAccess(ctx, access)

		assert.Error(t, err)
		assert.Nil(t, result)

		mock.AssertExpectationsForObjects(t, dataManager)
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

		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"GetOAuth2ClientTokenByRefresh",
			testutils.ContextMatcher,
			token.Refresh,
		).Return(token, nil)

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		result, err := store.GetByRefresh(ctx, token.Refresh)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, token.ClientID, result.GetClientID())
		assert.Equal(t, token.Refresh, result.GetRefresh())

		mock.AssertExpectationsForObjects(t, dataManager)
	})

	T.Run("with database error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		refresh := "test-refresh"

		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"GetOAuth2ClientTokenByRefresh",
			testutils.ContextMatcher,
			refresh,
		).Return((*types.OAuth2ClientToken)(nil), errors.New("database error"))

		store := &oauth2TokenStoreImpl{
			tracer:      tracer,
			logger:      logger,
			dataManager: dataManager,
		}

		result, err := store.GetByRefresh(ctx, refresh)

		assert.Error(t, err)
		assert.Nil(t, result)

		mock.AssertExpectationsForObjects(t, dataManager)
	})
}
