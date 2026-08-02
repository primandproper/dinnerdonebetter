package authentication

import (
	"context"
	"errors"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/fakes"
	oauthmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/mock"

	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func TestNewOAuth2ClientStore(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		domain := "example.com"
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")
		dataManager := &oauthmock.RepositoryMock{}

		store := newOAuth2ClientStore(domain, logger, tracer, dataManager)

		assert.NotNil(t, store)
		impl, ok := store.(*oauth2ClientStoreImpl)
		assert.True(t, ok)
		assert.Equal(t, domain, impl.domain)
		assert.NotNil(t, impl.tracer)
		assert.NotNil(t, impl.logger)
		assert.Equal(t, dataManager, impl.dataManager)
	})
}

func TestOAuth2ClientStoreImpl_GetByID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		domain := "example.com"
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		client := fakes.BuildFakeOAuth2Client()
		dataManager := &oauthmock.RepositoryMock{
			GetOAuth2ClientByClientIDFunc: func(_ context.Context, clientID string) (*types.OAuth2Client, error) {
				assert.Equal(t, client.ID, clientID)
				return client, nil
			},
		}

		store := &oauth2ClientStoreImpl{
			domain:      domain,
			logger:      logger,
			tracer:      tracer,
			dataManager: dataManager,
		}

		result, err := store.GetByID(ctx, client.ID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, client.ID, result.GetID())
		assert.Equal(t, client.ClientSecret, result.GetSecret())
		assert.Equal(t, domain, result.GetDomain())
		assert.False(t, result.IsPublic())

		assert.Len(t, dataManager.GetOAuth2ClientByClientIDCalls(), 1)
	})

	T.Run("with error getting client", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		domain := "example.com"
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		clientID := "test-client-id"
		dataManager := &oauthmock.RepositoryMock{
			GetOAuth2ClientByClientIDFunc: func(_ context.Context, actualClientID string) (*types.OAuth2Client, error) {
				assert.Equal(t, clientID, actualClientID)
				return nil, errors.New("database error")
			},
		}

		store := &oauth2ClientStoreImpl{
			domain:      domain,
			logger:      logger,
			tracer:      tracer,
			dataManager: dataManager,
		}

		result, err := store.GetByID(ctx, clientID)

		assert.Error(t, err)
		assert.Nil(t, result)

		assert.Len(t, dataManager.GetOAuth2ClientByClientIDCalls(), 1)
	})
}
