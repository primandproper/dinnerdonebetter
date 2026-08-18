package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"

	"github.com/primandproper/platform-go/v11/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServiceImpl_AdminSetPasswordChangeRequired(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUserID := fake.BuildFakeID()

		identityDataManager.AdminSetPasswordChangeRequiredFunc = func(_ context.Context, userID string, requiresChange bool) error {
			assert.Equal(t, exampleUserID, userID)
			assert.True(t, requiresChange)

			return nil
		}

		request := &identitysvc.AdminSetPasswordChangeRequiredRequest{
			TargetUserId:           exampleUserID,
			RequiresPasswordChange: true,
		}

		result, err := service.AdminSetPasswordChangeRequired(buildAdminSessionContextForTest(t), request)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)

		request := &identitysvc.AdminSetPasswordChangeRequiredRequest{
			TargetUserId:           fake.BuildFakeID(),
			RequiresPasswordChange: true,
		}

		result, err := service.AdminSetPasswordChangeRequired(t.Context(), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.AdminSetPasswordChangeRequiredFunc = func(_ context.Context, _ string, requiresChange bool) error {
			assert.True(t, requiresChange)

			return errors.New("update error")
		}

		request := &identitysvc.AdminSetPasswordChangeRequiredRequest{
			TargetUserId:           fake.BuildFakeID(),
			RequiresPasswordChange: true,
		}

		result, err := service.AdminSetPasswordChangeRequired(buildAdminSessionContextForTest(t), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})

	t.Run("with insufficient permissions", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)

		request := &identitysvc.AdminSetPasswordChangeRequiredRequest{
			TargetUserId:           fake.BuildFakeID(),
			RequiresPasswordChange: true,
		}

		result, err := service.AdminSetPasswordChangeRequired(buildInsufficientPermissionsSessionContextForTest(t), request)

		require.Error(t, err)
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

		service, identityDataManager := buildTestService(t)

		exampleUserID := fake.BuildFakeID()

		identityDataManager.AdminUpdateUserStatusFunc = func(_ context.Context, input *identity.UserAccountStatusUpdateInput) error {
			assert.True(t, input.TargetUserID == exampleUserID && input.NewStatus == identity.GoodStandingUserAccountStatus.String())

			return nil
		}

		request := &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: exampleUserID,
			NewStatus:    identity.GoodStandingUserAccountStatus.String(),
			Reason:       "Admin update for testing",
		}

		result, err := service.AdminUpdateUserStatus(buildAdminSessionContextForTest(t), request)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)

		request := &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: fake.BuildFakeID(),
			NewStatus:    identity.GoodStandingUserAccountStatus.String(),
		}

		result, err := service.AdminUpdateUserStatus(t.Context(), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.AdminUpdateUserStatusFunc = func(_ context.Context, _ *identity.UserAccountStatusUpdateInput) error {
			return errors.New("update error")
		}

		request := &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: fake.BuildFakeID(),
			NewStatus:    identity.GoodStandingUserAccountStatus.String(),
		}

		result, err := service.AdminUpdateUserStatus(buildAdminSessionContextForTest(t), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})

	t.Run("with insufficient permissions", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)

		request := &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: fake.BuildFakeID(),
			NewStatus:    identity.BannedUserAccountStatus.String(),
		}

		result, err := service.AdminUpdateUserStatus(buildInsufficientPermissionsSessionContextForTest(t), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, grpcErr.Code())
	})

	t.Run("with banned status", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUserID := fake.BuildFakeID()

		identityDataManager.AdminUpdateUserStatusFunc = func(_ context.Context, input *identity.UserAccountStatusUpdateInput) error {
			assert.True(t, input.TargetUserID == exampleUserID && input.NewStatus == identity.BannedUserAccountStatus.String())

			return nil
		}

		request := &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: exampleUserID,
			NewStatus:    identity.BannedUserAccountStatus.String(),
			Reason:       "User violated terms of service",
		}

		result, err := service.AdminUpdateUserStatus(buildAdminSessionContextForTest(t), request)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with unverified status", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUserID := fake.BuildFakeID()

		identityDataManager.AdminUpdateUserStatusFunc = func(_ context.Context, input *identity.UserAccountStatusUpdateInput) error {
			assert.True(t, input.TargetUserID == exampleUserID && input.NewStatus == identity.UnverifiedAccountStatus.String())

			return nil
		}

		request := &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: exampleUserID,
			NewStatus:    identity.UnverifiedAccountStatus.String(),
			Reason:       "Reset verification status",
		}

		result, err := service.AdminUpdateUserStatus(buildAdminSessionContextForTest(t), request)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})
}
