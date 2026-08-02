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
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"

	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func buildTestServiceWithAdminPermissions(t *testing.T) (*serviceImpl, *managermock.IdentityDataManagerMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	identityDataManager := &managermock.IdentityDataManagerMock{}

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
	identityDataManager := &managermock.IdentityDataManagerMock{}

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

		identityDataManager.AdminSetPasswordChangeRequiredFunc = func(_ context.Context, userID string, requiresChange bool) error {
			assert.Equal(t, exampleUserID, userID)
			assert.Equal(t, true, requiresChange)

			return nil
		}

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

		identityDataManager.AdminSetPasswordChangeRequiredFunc = func(_ context.Context, _ string, requiresChange bool) error {
			assert.Equal(t, true, requiresChange)

			return errors.New("update error")
		}

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

	t.Run("with insufficient permissions", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithInsufficientPermissions(t)

		request := &identitysvc.AdminSetPasswordChangeRequiredRequest{
			TargetUserId:           identityfakes.BuildFakeID(),
			RequiresPasswordChange: true,
		}

		result, err := service.AdminSetPasswordChangeRequired(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, grpcErr.Code())
	})
}

func TestServiceImpl_AdminUpdateUserStatus(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestServiceWithAdminPermissions(t)

		exampleUserID := identityfakes.BuildFakeID()

		identityDataManager.AdminUpdateUserStatusFunc = func(_ context.Context, input *identity.UserAccountStatusUpdateInput) error {
			assert.True(t, input.TargetUserID == exampleUserID && input.NewStatus == identity.GoodStandingUserAccountStatus.String())

			return nil
		}

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

		identityDataManager.AdminUpdateUserStatusFunc = func(_ context.Context, _ *identity.UserAccountStatusUpdateInput) error {
			return errors.New("update error")
		}

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

	t.Run("with insufficient permissions", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithInsufficientPermissions(t)

		request := &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: identityfakes.BuildFakeID(),
			NewStatus:    identity.BannedUserAccountStatus.String(),
		}

		result, err := service.AdminUpdateUserStatus(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, grpcErr.Code())
	})

	t.Run("with banned status", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestServiceWithAdminPermissions(t)

		exampleUserID := identityfakes.BuildFakeID()

		identityDataManager.AdminUpdateUserStatusFunc = func(_ context.Context, input *identity.UserAccountStatusUpdateInput) error {
			assert.True(t, input.TargetUserID == exampleUserID && input.NewStatus == identity.BannedUserAccountStatus.String())

			return nil
		}

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

		identityDataManager.AdminUpdateUserStatusFunc = func(_ context.Context, input *identity.UserAccountStatusUpdateInput) error {
			assert.True(t, input.TargetUserID == exampleUserID && input.NewStatus == identity.UnverifiedAccountStatus.String())

			return nil
		}

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
