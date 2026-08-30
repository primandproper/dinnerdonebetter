package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	grpcfiltering "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServiceImpl_AcceptAccountInvitation(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInvitationID := fake.BuildFakeID()

		identityDataManager.AcceptAccountInvitationFunc = func(_ context.Context, _ string, accountInvitationID string, _ *identity.AccountInvitationUpdateRequestInput) error {
			assert.Equal(t, exampleInvitationID, accountInvitationID)

			return nil
		}

		request := &identitysvc.AcceptAccountInvitationRequest{
			AccountInvitationId: exampleInvitationID,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: "invitation-token",
				Note:  "Accepting invitation",
			},
		}

		result, err := service.AcceptAccountInvitation(buildSessionContextForTest(t), request)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)

		request := &identitysvc.AcceptAccountInvitationRequest{
			AccountInvitationId: fake.BuildFakeID(),
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: "invitation-token",
			},
		}

		result, err := service.AcceptAccountInvitation(t.Context(), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInvitationID := fake.BuildFakeID()

		identityDataManager.AcceptAccountInvitationFunc = func(_ context.Context, _ string, accountInvitationID string, _ *identity.AccountInvitationUpdateRequestInput) error {
			assert.Equal(t, exampleInvitationID, accountInvitationID)

			return errors.New("accept error")
		}

		request := &identitysvc.AcceptAccountInvitationRequest{
			AccountInvitationId: exampleInvitationID,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: "invitation-token",
			},
		}

		result, err := service.AcceptAccountInvitation(buildSessionContextForTest(t), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_RejectAccountInvitation(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInvitationID := fake.BuildFakeID()

		identityDataManager.RejectAccountInvitationFunc = func(_ context.Context, _ string, accountInvitationID string, _ *identity.AccountInvitationUpdateRequestInput) error {
			assert.Equal(t, exampleInvitationID, accountInvitationID)

			return nil
		}

		request := &identitysvc.RejectAccountInvitationRequest{
			AccountInvitationId: exampleInvitationID,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: "invitation-token",
				Note:  "Rejecting invitation",
			},
		}

		result, err := service.RejectAccountInvitation(buildSessionContextForTest(t), request)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)

		request := &identitysvc.RejectAccountInvitationRequest{
			AccountInvitationId: fake.BuildFakeID(),
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: "invitation-token",
			},
		}

		result, err := service.RejectAccountInvitation(t.Context(), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInvitationID := fake.BuildFakeID()

		identityDataManager.RejectAccountInvitationFunc = func(_ context.Context, _ string, accountInvitationID string, _ *identity.AccountInvitationUpdateRequestInput) error {
			assert.Equal(t, exampleInvitationID, accountInvitationID)

			return errors.New("reject error")
		}

		request := &identitysvc.RejectAccountInvitationRequest{
			AccountInvitationId: exampleInvitationID,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: "invitation-token",
			},
		}

		result, err := service.RejectAccountInvitation(buildSessionContextForTest(t), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_CancelAccountInvitation(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInvitationID := fake.BuildFakeID()

		identityDataManager.CancelAccountInvitationFunc = func(_ context.Context, _ string, accountInvitationID string, note string) error {
			assert.Equal(t, exampleInvitationID, accountInvitationID)
			assert.Equal(t, "Cancelling invitation", note)

			return nil
		}

		request := &identitysvc.CancelAccountInvitationRequest{
			AccountInvitationId: exampleInvitationID,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Note: "Cancelling invitation",
			},
		}

		result, err := service.CancelAccountInvitation(buildSessionContextForTest(t), request)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)

		request := &identitysvc.CancelAccountInvitationRequest{
			AccountInvitationId: fake.BuildFakeID(),
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Note: "Cancelling invitation",
			},
		}

		result, err := service.CancelAccountInvitation(t.Context(), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInvitationID := fake.BuildFakeID()

		identityDataManager.CancelAccountInvitationFunc = func(_ context.Context, _ string, accountInvitationID string, note string) error {
			assert.Equal(t, exampleInvitationID, accountInvitationID)
			assert.Equal(t, "Cancelling invitation", note)

			return errors.New("cancel error")
		}

		request := &identitysvc.CancelAccountInvitationRequest{
			AccountInvitationId: exampleInvitationID,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Note: "Cancelling invitation",
			},
		}

		result, err := service.CancelAccountInvitation(buildSessionContextForTest(t), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_GetAccountInvitation(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInvitation := identityfakes.BuildFakeAccountInvitation()

		identityDataManager.GetAccountInvitationFunc = func(_ context.Context, _ string, accountInvitationID string) (*identity.AccountInvitation, error) {
			assert.Equal(t, exampleInvitation.ID, accountInvitationID)

			return exampleInvitation, nil
		}

		request := &identitysvc.GetAccountInvitationRequest{
			AccountInvitationId: exampleInvitation.ID,
		}

		result, err := service.GetAccountInvitation(buildSessionContextForTest(t), request)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.NotNil(t, result.Result)
		assert.Equal(t, exampleInvitation.ID, result.Result.Id)
		assert.Equal(t, exampleInvitation.ToEmail, result.Result.ToEmail)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)

		request := &identitysvc.GetAccountInvitationRequest{
			AccountInvitationId: fake.BuildFakeID(),
		}

		result, err := service.GetAccountInvitation(t.Context(), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInvitationID := fake.BuildFakeID()

		identityDataManager.GetAccountInvitationFunc = func(_ context.Context, _ string, accountInvitationID string) (*identity.AccountInvitation, error) {
			assert.Equal(t, exampleInvitationID, accountInvitationID)

			return nil, errors.New("get error")
		}

		request := &identitysvc.GetAccountInvitationRequest{
			AccountInvitationId: exampleInvitationID,
		}

		result, err := service.GetAccountInvitation(buildSessionContextForTest(t), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_GetReceivedAccountInvitations(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, identityDataManager := buildTestService(t)

		exampleInvitations := &filtering.QueryFilteredResult[identity.AccountInvitation]{
			Data: []*identity.AccountInvitation{
				identityfakes.BuildFakeAccountInvitation(),
				identityfakes.BuildFakeAccountInvitation(),
			},
		}

		identityDataManager.GetReceivedAccountInvitationsFunc = func(_ context.Context, _ string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
			return exampleInvitations, nil
		}

		pageSize := uint32(25)
		request := &identitysvc.GetReceivedAccountInvitationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetReceivedAccountInvitations(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.Len(t, result.Results, len(exampleInvitations.Data))
		for i := range result.Results {
			assert.Equal(t, result.Results[i].Id, exampleInvitations.Data[i].ID)
		}
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)

		pageSize := uint32(25)
		request := &identitysvc.GetReceivedAccountInvitationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetReceivedAccountInvitations(t.Context(), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, identityDataManager := buildTestService(t)

		identityDataManager.GetReceivedAccountInvitationsFunc = func(_ context.Context, _ string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
			return nil, errors.New("get error")
		}

		pageSize := uint32(25)
		request := &identitysvc.GetReceivedAccountInvitationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetReceivedAccountInvitations(ctx, request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_GetSentAccountInvitations(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		service, identityDataManager := buildTestService(t)

		exampleInvitations := &filtering.QueryFilteredResult[identity.AccountInvitation]{
			Data: []*identity.AccountInvitation{
				identityfakes.BuildFakeAccountInvitation(),
				identityfakes.BuildFakeAccountInvitation(),
			},
		}

		identityDataManager.GetSentAccountInvitationsFunc = func(_ context.Context, _ string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
			return exampleInvitations, nil
		}

		pageSize := uint32(25)
		request := &identitysvc.GetSentAccountInvitationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetSentAccountInvitations(ctx, request)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.Len(t, result.Results, len(exampleInvitations.Data))
		for i := range result.Results {
			assert.Equal(t, result.Results[i].Id, exampleInvitations.Data[i].ID)
		}
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service, _ := buildTestService(t)

		pageSize := uint32(25)
		request := &identitysvc.GetSentAccountInvitationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetSentAccountInvitations(t.Context(), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.GetSentAccountInvitationsFunc = func(_ context.Context, _ string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
			return nil, errors.New("get error")
		}

		pageSize := uint32(25)
		request := &identitysvc.GetSentAccountInvitationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetSentAccountInvitations(buildSessionContextForTest(t), request)

		require.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}
