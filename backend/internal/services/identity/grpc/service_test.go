package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	managermock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager/mock"
	uploadedmediamock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/mock"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"

	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"
	mockuploads "github.com/primandproper/platform-go/v9/uploads/mock"

	"github.com/stretchr/testify/assert"
)

func buildTestService(t *testing.T) (*serviceImpl, *managermock.IdentityDataManagerMock) {
	t.Helper()
	service, identityDataManager, _ := buildTestServiceWithUploadMocks(t)
	return service, identityDataManager
}

func buildTestServiceWithUploadMocks(t *testing.T) (*serviceImpl, *managermock.IdentityDataManagerMock, *uploadedmediamock.RepositoryMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	identityDataManager := &managermock.IdentityDataManagerMock{}
	uploadedMediaRepo := &uploadedmediamock.RepositoryMock{}
	uploadManager := &mockuploads.UploadManagerMock{}

	service := &serviceImpl{
		tracer:               tracer,
		logger:               logger,
		identityDataManager:  identityDataManager,
		uploadedMediaManager: uploadedMediaRepo,
		uploadManager:        uploadManager,
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID:             identityfakes.BuildFakeID(),
					AccountStatus:      identity.GoodStandingUserAccountStatus.String(),
					ServicePermissions: authorization.NewServiceRolePermissionChecker([]string{authorization.ServiceUserRole.String()}, nil),
				},
				ActiveAccountID: identityfakes.BuildFakeID(),
				AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
					identityfakes.BuildFakeID(): authorization.NewAccountRolePermissionChecker(nil),
				},
			}, nil
		},
	}

	return service, identityDataManager, uploadedMediaRepo
}

// buildTestServiceWithAccountMembership returns a service whose authenticated session is a member of
// (and has the given account as its active account) accountID, satisfying handler ownership checks.
func buildTestServiceWithAccountMembership(t *testing.T, accountID string) (*serviceImpl, *managermock.IdentityDataManagerMock) {
	t.Helper()

	service, identityDataManager := buildTestService(t)
	service.sessionContextDataFetcher = func(context.Context) (*sessions.ContextData, error) {
		return &sessions.ContextData{
			Requester: sessions.RequesterInfo{
				UserID:             identityfakes.BuildFakeID(),
				AccountStatus:      identity.GoodStandingUserAccountStatus.String(),
				ServicePermissions: authorization.NewServiceRolePermissionChecker([]string{authorization.ServiceUserRole.String()}, nil),
			},
			ActiveAccountID: accountID,
			AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
				accountID: authorization.NewAccountRolePermissionChecker(nil),
			},
		}, nil
	}

	return service, identityDataManager
}

func buildTestServiceWithSessionError(t *testing.T) *serviceImpl {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	identityDataManager := &managermock.IdentityDataManagerMock{}

	service := &serviceImpl{
		tracer:               tracer,
		logger:               logger,
		identityDataManager:  identityDataManager,
		uploadedMediaManager: &uploadedmediamock.RepositoryMock{},
		uploadManager:        &mockuploads.UploadManagerMock{},
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return nil, errors.New("session error")
		},
	}

	return service
}

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		sessionContextDataFetcher := func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{}, nil
		}
		identityDataManager := &managermock.IdentityDataManagerMock{}

		uploadedMediaManager := &uploadedmediamock.RepositoryMock{}
		uploadManager := &mockuploads.UploadManagerMock{}
		service := NewService(logger, tracerProvider, sessionContextDataFetcher, identityDataManager, uploadedMediaManager, uploadManager)

		assert.NotNil(t, service)
		assert.Implements(t, (*identitysvc.IdentityServiceServer)(nil), service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
		assert.NotNil(t, impl.sessionContextDataFetcher)
		assert.Equal(t, identityDataManager, impl.identityDataManager)
		assert.Equal(t, uploadedMediaManager, impl.uploadedMediaManager)
		assert.Equal(t, uploadManager, impl.uploadManager)
	})
}

func TestServiceImpl_buildResponseDetails(t *testing.T) {
	t.Parallel()

	t.Run("with valid session context", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)
		ctx := t.Context()

		result := service.buildResponseDetails(ctx, nil)

		assert.NotNil(t, result)
		assert.IsType(t, &types.ResponseDetails{}, result)
		assert.NotEmpty(t, result.CurrentAccountId)
	})

	t.Run("with span", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)
		ctx, span := service.tracer.StartSpan(t.Context())
		defer span.End()

		result := service.buildResponseDetails(ctx, span)

		assert.NotNil(t, result)
		assert.IsType(t, &types.ResponseDetails{}, result)
		assert.NotEmpty(t, result.TraceId)
		assert.NotEmpty(t, result.CurrentAccountId)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)
		ctx := t.Context()

		result := service.buildResponseDetails(ctx, nil)

		assert.NotNil(t, result)
		assert.IsType(t, &types.ResponseDetails{}, result)
		assert.Empty(t, result.CurrentAccountId)
	})

	t.Run("with nil span", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)
		ctx := t.Context()

		result := service.buildResponseDetails(ctx, nil)

		assert.NotNil(t, result)
		assert.IsType(t, &types.ResponseDetails{}, result)
		assert.Empty(t, result.TraceId)
		assert.NotEmpty(t, result.CurrentAccountId)
	})
}
