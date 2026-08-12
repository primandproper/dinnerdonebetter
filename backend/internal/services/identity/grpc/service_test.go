package grpc

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	managermock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager/mock"
	uploadedmediamock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/mock"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"

	loggingnoop "github.com/primandproper/platform-go/v10/observability/logging/noop"
	"github.com/primandproper/platform-go/v10/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v10/observability/tracing/noop"
	mockuploads "github.com/primandproper/platform-go/v10/uploads/mock"

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
	}

	return service, identityDataManager, uploadedMediaRepo
}

// buildSessionContextForTest returns a context whose session is an ordinary, good-standing user.
func buildSessionContextForTest(t *testing.T) context.Context {
	t.Helper()

	return buildSessionContextForAccount(t, identityfakes.BuildFakeID())
}

// buildSessionContextForAccount returns a context whose session is a member of (and has as its
// active account) accountID, satisfying handler ownership checks.
func buildSessionContextForAccount(t *testing.T, accountID string) context.Context {
	t.Helper()

	return sessions.AttachToContext(t.Context(), &sessions.ContextData{
		Requester: sessions.RequesterInfo{
			UserID:             identityfakes.BuildFakeID(),
			AccountStatus:      identity.GoodStandingUserAccountStatus.String(),
			ServicePermissions: authorization.NewServiceRolePermissionChecker([]string{authorization.ServiceUserRole.String()}, nil),
		},
		ActiveAccountID: accountID,
		AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
			accountID: authorization.NewAccountRolePermissionChecker(nil),
		},
	})
}

// buildAdminSessionContextForTest returns a context whose session holds service-admin authority.
func buildAdminSessionContextForTest(t *testing.T) context.Context {
	t.Helper()

	accountID := identityfakes.BuildFakeID()

	return sessions.AttachToContext(t.Context(), &sessions.ContextData{
		Requester: sessions.RequesterInfo{
			UserID:             identityfakes.BuildFakeID(),
			AccountStatus:      identity.GoodStandingUserAccountStatus.String(),
			ServicePermissions: authorization.NewServiceRolePermissionChecker([]string{authorization.ServiceAdminRole.String()}, authorization.ServiceAdminPermissions),
		},
		ActiveAccountID: accountID,
		AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
			accountID: authorization.NewAccountRolePermissionChecker(authorization.AccountMemberPermissions),
		},
	})
}

// buildInsufficientPermissionsSessionContextForTest returns a context whose session holds no
// service-admin authority, for the handlers that must reject it.
func buildInsufficientPermissionsSessionContextForTest(t *testing.T) context.Context {
	t.Helper()

	return buildSessionContextForAccount(t, identityfakes.BuildFakeID())
}

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		identityDataManager := &managermock.IdentityDataManagerMock{}

		uploadedMediaManager := &uploadedmediamock.RepositoryMock{}
		uploadManager := &mockuploads.UploadManagerMock{}
		service := NewService(logger, tracerProvider, identityDataManager, uploadedMediaManager, uploadManager)

		assert.NotNil(t, service)
		assert.Implements(t, (*identitysvc.IdentityServiceServer)(nil), service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
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
		ctx := buildSessionContextForTest(t)

		result := service.buildResponseDetails(ctx, nil)

		assert.NotNil(t, result)
		assert.IsType(t, &types.ResponseDetails{}, result)
		assert.NotEmpty(t, result.CurrentAccountId)
	})

	t.Run("with span", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)
		ctx, span := service.tracer.StartSpan(buildSessionContextForTest(t))
		defer span.End()

		result := service.buildResponseDetails(ctx, span)

		assert.NotNil(t, result)
		assert.IsType(t, &types.ResponseDetails{}, result)
		assert.NotEmpty(t, result.TraceId)
		assert.NotEmpty(t, result.CurrentAccountId)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)
		ctx := t.Context()

		result := service.buildResponseDetails(ctx, nil)

		assert.NotNil(t, result)
		assert.IsType(t, &types.ResponseDetails{}, result)
		assert.Empty(t, result.CurrentAccountId)
	})

	t.Run("with nil span", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)
		ctx := buildSessionContextForTest(t)

		result := service.buildResponseDetails(ctx, nil)

		assert.NotNil(t, result)
		assert.IsType(t, &types.ResponseDetails{}, result)
		assert.Empty(t, result.TraceId)
		assert.NotEmpty(t, result.CurrentAccountId)
	})
}
