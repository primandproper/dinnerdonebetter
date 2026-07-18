package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	managermock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager/mock"
	identitysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	"github.com/primandproper/platform-go/v5/observability/tracing"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func buildTestServiceWithAdminPermissions(t *testing.T) (*serviceImpl, *managermock.IdentityDataManager) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	identityDataManager := &managermock.IdentityDataManager{}

	service := &serviceImpl{
		tracer:              tracer,
		logger:              logger,
		identityDataManager: identityDataManager,
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID:             identityfakes.BuildFakeID(),
					AccountStatus:      identity.GoodStandingUserAccountStatus.String(),
					ServicePermissions: authorization.NewServiceRolePermissionChecker([]string{authorization.ServiceAdminRole.String()}, authorization.ServiceAdminPermissions),
				},
				ActiveAccountID: identityfakes.BuildFakeID(),
				AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
					identityfakes.BuildFakeID(): authorization.NewAccountRolePermissionChecker(authorization.AccountMemberPermissions),
				},
			}, nil
		},
	}

	return service, identityDataManager
}

func buildTestServiceWithInsufficientPermissions(t *testing.T) *serviceImpl {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	identityDataManager := &managermock.IdentityDataManager{}

	service := &serviceImpl{
		tracer:              tracer,
		logger:              logger,
		identityDataManager: identityDataManager,
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

	return service
}

func TestServiceImpl_AdminSetPasswordChangeRequired(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestServiceWithAdminPermissions(t)

		exampleUserID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.AdminSetPasswordChangeRequired), testutils.ContextMatcher, exampleUserID, true).Return(nil)

		request := &identitysvc.AdminSetPasswordChangeRequiredRequest{
			TargetUserId:           exampleUserID,
			RequiresPasswordChange: true,
		}

		result, err := service.AdminSetPasswordChangeRequired(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.AdminSetPasswordChangeRequiredRequest{
			TargetUserId:           identityfakes.BuildFakeID(),
			RequiresPasswordChange: true,
		}

		result, err := service.AdminSetPasswordChangeRequired(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestServiceWithAdminPermissions(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.AdminSetPasswordChangeRequired), testutils.ContextMatcher, mock.AnythingOfType("string"), true).Return(errors.New("update error"))

		request := &identitysvc.AdminSetPasswordChangeRequiredRequest{
			TargetUserId:           identityfakes.BuildFakeID(),
			RequiresPasswordChange: true,
		}

		result, err := service.AdminSetPasswordChangeRequired(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_AdminUpdateUserStatus(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestServiceWithAdminPermissions(t)

		exampleUserID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.AdminUpdateUserStatus), testutils.ContextMatcher, mock.MatchedBy(func(input *identity.UserAccountStatusUpdateInput) bool {
			return input.TargetUserID == exampleUserID &&
				input.NewStatus == identity.GoodStandingUserAccountStatus.String()
		})).Return(nil)

		request := &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: exampleUserID,
			NewStatus:    identity.GoodStandingUserAccountStatus.String(),
			Reason:       "Admin update for testing",
		}

		result, err := service.AdminUpdateUserStatus(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: identityfakes.BuildFakeID(),
			NewStatus:    identity.GoodStandingUserAccountStatus.String(),
		}

		result, err := service.AdminUpdateUserStatus(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestServiceWithAdminPermissions(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.AdminUpdateUserStatus), testutils.ContextMatcher, mock.AnythingOfType("*identity.UserAccountStatusUpdateInput")).Return(errors.New("update error"))

		request := &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: identityfakes.BuildFakeID(),
			NewStatus:    identity.GoodStandingUserAccountStatus.String(),
		}

		result, err := service.AdminUpdateUserStatus(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})

	t.Run("with banned status", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestServiceWithAdminPermissions(t)

		exampleUserID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.AdminUpdateUserStatus), testutils.ContextMatcher, mock.MatchedBy(func(input *identity.UserAccountStatusUpdateInput) bool {
			return input.TargetUserID == exampleUserID &&
				input.NewStatus == identity.BannedUserAccountStatus.String()
		})).Return(nil)

		request := &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: exampleUserID,
			NewStatus:    identity.BannedUserAccountStatus.String(),
			Reason:       "User violated terms of service",
		}

		result, err := service.AdminUpdateUserStatus(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with unverified status", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestServiceWithAdminPermissions(t)

		exampleUserID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.AdminUpdateUserStatus), testutils.ContextMatcher, mock.MatchedBy(func(input *identity.UserAccountStatusUpdateInput) bool {
			return input.TargetUserID == exampleUserID &&
				input.NewStatus == identity.UnverifiedAccountStatus.String()
		})).Return(nil)

		request := &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: exampleUserID,
			NewStatus:    identity.UnverifiedAccountStatus.String(),
			Reason:       "Reset verification status",
		}

		result, err := service.AdminUpdateUserStatus(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})
}
