package grpc

import (
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	identitysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v3/database/filtering"
	"github.com/primandproper/platform-go/v3/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServiceImpl_AcceptAccountInvitation(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInvitationID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.AcceptAccountInvitation), testutils.ContextMatcher, mock.AnythingOfType("string"), exampleInvitationID, mock.AnythingOfType("*identity.AccountInvitationUpdateRequestInput")).Return(nil)

		request := &identitysvc.AcceptAccountInvitationRequest{
			AccountInvitationId: exampleInvitationID,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: "invitation-token",
				Note:  "Accepting invitation",
			},
		}

		result, err := service.AcceptAccountInvitation(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.AcceptAccountInvitationRequest{
			AccountInvitationId: identityfakes.BuildFakeID(),
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: "invitation-token",
			},
		}

		result, err := service.AcceptAccountInvitation(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInvitationID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.AcceptAccountInvitation), testutils.ContextMatcher, mock.AnythingOfType("string"), exampleInvitationID, mock.AnythingOfType("*identity.AccountInvitationUpdateRequestInput")).Return(errors.New("accept error"))

		request := &identitysvc.AcceptAccountInvitationRequest{
			AccountInvitationId: exampleInvitationID,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: "invitation-token",
			},
		}

		result, err := service.AcceptAccountInvitation(t.Context(), request)

		assert.Error(t, err)
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

		exampleInvitationID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.RejectAccountInvitation), testutils.ContextMatcher, mock.AnythingOfType("string"), exampleInvitationID, mock.AnythingOfType("*identity.AccountInvitationUpdateRequestInput")).Return(nil)

		request := &identitysvc.RejectAccountInvitationRequest{
			AccountInvitationId: exampleInvitationID,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: "invitation-token",
				Note:  "Rejecting invitation",
			},
		}

		result, err := service.RejectAccountInvitation(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.RejectAccountInvitationRequest{
			AccountInvitationId: identityfakes.BuildFakeID(),
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: "invitation-token",
			},
		}

		result, err := service.RejectAccountInvitation(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInvitationID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.RejectAccountInvitation), testutils.ContextMatcher, mock.AnythingOfType("string"), exampleInvitationID, mock.AnythingOfType("*identity.AccountInvitationUpdateRequestInput")).Return(errors.New("reject error"))

		request := &identitysvc.RejectAccountInvitationRequest{
			AccountInvitationId: exampleInvitationID,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Token: "invitation-token",
			},
		}

		result, err := service.RejectAccountInvitation(t.Context(), request)

		assert.Error(t, err)
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

		exampleInvitationID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.CancelAccountInvitation), testutils.ContextMatcher, mock.AnythingOfType("string"), exampleInvitationID, "Cancelling invitation").Return(nil)

		request := &identitysvc.CancelAccountInvitationRequest{
			AccountInvitationId: exampleInvitationID,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Note: "Cancelling invitation",
			},
		}

		result, err := service.CancelAccountInvitation(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.CancelAccountInvitationRequest{
			AccountInvitationId: identityfakes.BuildFakeID(),
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Note: "Cancelling invitation",
			},
		}

		result, err := service.CancelAccountInvitation(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInvitationID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.CancelAccountInvitation), testutils.ContextMatcher, mock.AnythingOfType("string"), exampleInvitationID, "Cancelling invitation").Return(errors.New("cancel error"))

		request := &identitysvc.CancelAccountInvitationRequest{
			AccountInvitationId: exampleInvitationID,
			Input: &identitysvc.AccountInvitationUpdateRequestInput{
				Note: "Cancelling invitation",
			},
		}

		result, err := service.CancelAccountInvitation(t.Context(), request)

		assert.Error(t, err)
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

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetAccountInvitation), testutils.ContextMatcher, mock.AnythingOfType("string"), exampleInvitation.ID).Return(exampleInvitation, nil)

		request := &identitysvc.GetAccountInvitationRequest{
			AccountInvitationId: exampleInvitation.ID,
		}

		result, err := service.GetAccountInvitation(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.NotNil(t, result.Result)
		assert.Equal(t, exampleInvitation.ID, result.Result.Id)
		assert.Equal(t, exampleInvitation.ToEmail, result.Result.ToEmail)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.GetAccountInvitationRequest{
			AccountInvitationId: identityfakes.BuildFakeID(),
		}

		result, err := service.GetAccountInvitation(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInvitationID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetAccountInvitation), testutils.ContextMatcher, mock.AnythingOfType("string"), exampleInvitationID).Return((*identity.AccountInvitation)(nil), errors.New("get error"))

		request := &identitysvc.GetAccountInvitationRequest{
			AccountInvitationId: exampleInvitationID,
		}

		result, err := service.GetAccountInvitation(t.Context(), request)

		assert.Error(t, err)
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

		ctx := t.Context()
		service, identityDataManager := buildTestService(t)

		exampleInvitations := &filtering.QueryFilteredResult[identity.AccountInvitation]{
			Data: []*identity.AccountInvitation{
				identityfakes.BuildFakeAccountInvitation(),
				identityfakes.BuildFakeAccountInvitation(),
			},
		}

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetReceivedAccountInvitations), testutils.ContextMatcher, mock.AnythingOfType("string"), testutils.QueryFilterMatcher).Return(exampleInvitations, nil)

		pageSize := uint32(25)
		request := &identitysvc.GetReceivedAccountInvitationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetReceivedAccountInvitations(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.Equal(t, len(exampleInvitations.Data), len(result.Results))
		for i := range result.Results {
			assert.Equal(t, result.Results[i].Id, exampleInvitations.Data[i].ID)
		}
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		pageSize := uint32(25)
		request := &identitysvc.GetReceivedAccountInvitationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetReceivedAccountInvitations(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetReceivedAccountInvitations), testutils.ContextMatcher, mock.AnythingOfType("string"), testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[identity.AccountInvitation])(nil), errors.New("get error"))

		pageSize := uint32(25)
		request := &identitysvc.GetReceivedAccountInvitationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetReceivedAccountInvitations(ctx, request)

		assert.Error(t, err)
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

		ctx := t.Context()
		service, identityDataManager := buildTestService(t)

		exampleInvitations := &filtering.QueryFilteredResult[identity.AccountInvitation]{
			Data: []*identity.AccountInvitation{
				identityfakes.BuildFakeAccountInvitation(),
				identityfakes.BuildFakeAccountInvitation(),
			},
		}

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetSentAccountInvitations), testutils.ContextMatcher, mock.AnythingOfType("string"), testutils.QueryFilterMatcher).Return(exampleInvitations,
			nil)

		pageSize := uint32(25)
		request := &identitysvc.GetSentAccountInvitationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetSentAccountInvitations(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.Equal(t, len(exampleInvitations.Data), len(result.Results))
		for i := range result.Results {
			assert.Equal(t, result.Results[i].Id, exampleInvitations.Data[i].ID)
		}
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		pageSize := uint32(25)
		request := &identitysvc.GetSentAccountInvitationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetSentAccountInvitations(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetSentAccountInvitations), testutils.ContextMatcher, mock.AnythingOfType("string"), testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[identity.AccountInvitation])(nil), errors.New("get error"))

		pageSize := uint32(25)
		request := &identitysvc.GetSentAccountInvitationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetSentAccountInvitations(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}
