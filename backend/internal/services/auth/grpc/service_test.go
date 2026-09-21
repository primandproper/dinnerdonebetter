package grpc

import (
	"context"
	"testing"

	authenticationmock "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/mock"
	authmanagermock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/managers/mock"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	identitymock "github.com/primandproper/platform-go/v14/identity/mock"

	"github.com/primandproper/primitives-go/v2/database"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/encoding"
	"github.com/primandproper/primitives-go/v2/featureflags"
	featureflagsmock "github.com/primandproper/primitives-go/v2/featureflags/mock"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
)

var (
	testAccountID = identifiers.New()
	testUserID    = identifiers.New()
)

func buildTestService(t *testing.T) (*serviceImpl, *identitymock.StoreMock, *authmanagermock.AuthManagerInterfaceMock, *authenticationmock.ManagerMock, *featureflagsmock.FeatureFlagManagerMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracerProvider := tracingnoop.NewTracerProvider()
	tracer := tracing.NewTracerForTest(t.Name())
	directory := &identitymock.StoreMock{}
	authManager := &authmanagermock.AuthManagerInterfaceMock{}
	authenticationManager := &authenticationmock.ManagerMock{}
	featureFlagManager := &featureflagsmock.FeatureFlagManagerMock{
		CanUseFeatureFunc: func(_ context.Context, _ string, _ featureflags.EvaluationContext) (bool, error) {
			return false, nil
		},
		GetStringValueFunc: func(_ context.Context, _ string, defaultValue string, _ featureflags.EvaluationContext) (string, error) {
			return defaultValue, nil
		},
		GetInt64ValueFunc: func(_ context.Context, _ string, defaultValue int64, _ featureflags.EvaluationContext) (int64, error) {
			return defaultValue, nil
		},
		GetFloat64ValueFunc: func(_ context.Context, _ string, defaultValue float64, _ featureflags.EvaluationContext) (float64, error) {
			return defaultValue, nil
		},
		GetObjectValueFunc: func(_ context.Context, _ string, defaultValue any, _ featureflags.EvaluationContext) (any, error) {
			return defaultValue, nil
		},
		CloseFunc: func() error { return nil },
	}

	jsonEncoder := encoding.NewServerEncoderDecoder(encoding.ContentTypeJSON, encoding.WithLogger(logger), encoding.WithTracerProvider(tracerProvider))

	service := &serviceImpl{
		tracer:    tracer,
		logger:    logger,
		directory: directory,
		// A client whose executors are nil, because the store is a mock and ignores them:
		// what the handlers actually need a client for is somewhere to ask for one.
		db: &mockdatabase.ClientMock{
			ReaderFunc: func() database.SQLQueryExecutor { return nil },
			WriterFunc: func() database.SQLQueryExecutor { return nil },
		},
		authManager:           authManager,
		authenticationManager: authenticationManager,
		featureFlagManager:    featureFlagManager,
		jsonEncoder:           jsonEncoder,
	}

	return service, directory, authManager, authenticationManager, featureFlagManager
}

func TestNewAuthService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		directory := &identitymock.StoreMock{}
		authManager := &authmanagermock.AuthManagerInterfaceMock{}
		authenticationManager := &authenticationmock.ManagerMock{}

		featureFlagManager := &featureflagsmock.FeatureFlagManagerMock{
			CanUseFeatureFunc: func(_ context.Context, _ string, _ featureflags.EvaluationContext) (bool, error) {
				return false, nil
			},
			CloseFunc: func() error { return nil },
		}
		service := NewAuthService(logger, tracerProvider, directory, nil, authManager, authenticationManager, featureFlagManager, nil)

		assert.NotNil(t, service)
		assert.Implements(t, (*authsvc.AuthServiceServer)(nil), service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
		assert.Equal(t, directory, impl.directory)
		assert.Equal(t, authManager, impl.authManager)
		assert.Equal(t, authenticationManager, impl.authenticationManager)
		assert.Equal(t, featureFlagManager, impl.featureFlagManager)
	})
}
