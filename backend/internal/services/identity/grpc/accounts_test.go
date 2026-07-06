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

func TestServiceImpl_ArchiveAccount(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleAccountID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.ArchiveAccount), testutils.ContextMatcher, exampleAccountID, mock.AnythingOfType("string")).Return(nil)

		request := &identitysvc.ArchiveAccountRequest{
			AccountId: exampleAccountID,
		}

		result, err := service.ArchiveAccount(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.ArchiveAccountRequest{
			AccountId: identityfakes.BuildFakeID(),
		}

		result, err := service.ArchiveAccount(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleAccountID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.ArchiveAccount), testutils.ContextMatcher, exampleAccountID, mock.AnythingOfType("string")).Return(errors.New("database error"))

		request := &identitysvc.ArchiveAccountRequest{
			AccountId: exampleAccountID,
		}

		result, err := service.ArchiveAccount(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_CreateAccount(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleAccount := identityfakes.BuildFakeAccount()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.CreateAccount), testutils.ContextMatcher, mock.MatchedBy(func(input *identity.AccountCreationRequestInput) bool {
			return input.Name == exampleAccount.Name &&
				input.BelongsToUser != ""
		})).Return(exampleAccount, nil)

		request := &identitysvc.CreateAccountRequest{
			Input: &identitysvc.AccountCreationRequestInput{
				Name:         exampleAccount.Name,
				ContactPhone: exampleAccount.ContactPhone,
				AddressLine1: exampleAccount.AddressLine1,
				City:         exampleAccount.City,
				State:        exampleAccount.State,
				ZipCode:      exampleAccount.ZipCode,
				Country:      exampleAccount.Country,
			},
		}

		result, err := service.CreateAccount(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.NotNil(t, result.Created)
		assert.Equal(t, exampleAccount.ID, result.Created.Id)
		assert.Equal(t, exampleAccount.Name, result.Created.Name)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.CreateAccountRequest{
			Input: &identitysvc.AccountCreationRequestInput{
				Name: "Test Account",
			},
		}

		result, err := service.CreateAccount(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.CreateAccount), testutils.ContextMatcher, mock.AnythingOfType("*identity.AccountCreationRequestInput")).Return((*identity.Account)(nil), errors.New("creation error"))

		request := &identitysvc.CreateAccountRequest{
			Input: &identitysvc.AccountCreationRequestInput{
				Name: "Test Account",
			},
		}

		result, err := service.CreateAccount(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_CreateAccountInvitation(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleInvitation := identityfakes.BuildFakeAccountInvitation()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.CreateAccountInvitation), testutils.ContextMatcher, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.MatchedBy(func(input *identity.AccountInvitationCreationRequestInput) bool {
			return input.ToEmail == exampleInvitation.ToEmail &&
				input.ToName == exampleInvitation.ToName
		})).Return(exampleInvitation, nil)

		request := &identitysvc.CreateAccountInvitationRequest{
			Input: &identitysvc.AccountInvitationCreationRequestInput{
				ToEmail: exampleInvitation.ToEmail,
				ToName:  exampleInvitation.ToName,
				Note:    exampleInvitation.Note,
			},
		}

		result, err := service.CreateAccountInvitation(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.NotNil(t, result.Created)
		assert.Equal(t, exampleInvitation.ID, result.Created.Id)
		assert.Equal(t, exampleInvitation.ToEmail, result.Created.ToEmail)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.CreateAccountInvitationRequest{
			Input: &identitysvc.AccountInvitationCreationRequestInput{
				ToEmail: "test@example.com",
			},
		}

		result, err := service.CreateAccountInvitation(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.CreateAccountInvitation), testutils.ContextMatcher, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*identity.AccountInvitationCreationRequestInput")).Return((*identity.AccountInvitation)(nil), errors.New("creation error"))

		request := &identitysvc.CreateAccountInvitationRequest{
			Input: &identitysvc.AccountInvitationCreationRequestInput{
				ToEmail: "test@example.com",
			},
		}

		result, err := service.CreateAccountInvitation(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_GetAccount(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleAccount := identityfakes.BuildFakeAccount()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetAccount), testutils.ContextMatcher, exampleAccount.ID).Return(exampleAccount, nil)

		request := &identitysvc.GetAccountRequest{
			AccountId: exampleAccount.ID,
		}

		result, err := service.GetAccount(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.NotNil(t, result.Result)
		assert.Equal(t, exampleAccount.ID, result.Result.Id)
		assert.Equal(t, exampleAccount.Name, result.Result.Name)
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleAccountID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetAccount), testutils.ContextMatcher, exampleAccountID).Return((*identity.Account)(nil), errors.New("database error"))

		request := &identitysvc.GetAccountRequest{
			AccountId: exampleAccountID,
		}

		result, err := service.GetAccount(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_GetAccounts(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, identityDataManager := buildTestService(t)

		exampleAccounts := &filtering.QueryFilteredResult[identity.Account]{
			Data: []*identity.Account{
				identityfakes.BuildFakeAccount(),
				identityfakes.BuildFakeAccount(),
			},
		}

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetAccounts), testutils.ContextMatcher, mock.AnythingOfType("string"), testutils.QueryFilterMatcher).Return(exampleAccounts, nil)

		pageSize := uint32(25)
		request := &identitysvc.GetAccountsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetAccounts(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.Equal(t, len(exampleAccounts.Data), len(result.Results))
		for i := range result.Results {
			assert.Equal(t, result.Results[i].Id, exampleAccounts.Data[i].ID)
		}
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		pageSize := uint32(25)
		request := &identitysvc.GetAccountsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetAccounts(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.GetAccounts), testutils.ContextMatcher, mock.AnythingOfType("string"), testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[identity.Account])(nil), errors.New("database error"))

		pageSize := uint32(25)
		request := &identitysvc.GetAccountsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &pageSize,
			},
		}

		result, err := service.GetAccounts(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_SetDefaultAccount(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleAccountID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.SetDefaultAccount), testutils.ContextMatcher, mock.AnythingOfType("string"), exampleAccountID).Return(nil)

		request := &identitysvc.SetDefaultAccountRequest{
			AccountId: exampleAccountID,
		}

		result, err := service.SetDefaultAccount(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.True(t, result.Success)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.SetDefaultAccountRequest{
			AccountId: identityfakes.BuildFakeID(),
		}

		result, err := service.SetDefaultAccount(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleAccountID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.SetDefaultAccount), testutils.ContextMatcher, mock.AnythingOfType("string"), exampleAccountID).Return(errors.New("update error"))

		request := &identitysvc.SetDefaultAccountRequest{
			AccountId: exampleAccountID,
		}

		result, err := service.SetDefaultAccount(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_TransferAccountOwnership(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleAccountID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.TransferAccountOwnership), testutils.ContextMatcher, mock.AnythingOfType("string"), mock.AnythingOfType("*identity.AccountOwnershipTransferInput")).Return(nil)

		request := &identitysvc.TransferAccountOwnershipRequest{
			AccountId: exampleAccountID,
			Input: &identitysvc.AccountOwnershipTransferInput{
				CurrentOwner: identityfakes.BuildFakeID(),
				NewOwner:     identityfakes.BuildFakeID(),
				Reason:       "Transfer for testing",
			},
		}

		result, err := service.TransferAccountOwnership(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
		assert.True(t, result.Success)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.TransferAccountOwnershipRequest{
			AccountId: identityfakes.BuildFakeID(),
			Input: &identitysvc.AccountOwnershipTransferInput{
				CurrentOwner: identityfakes.BuildFakeID(),
				NewOwner:     identityfakes.BuildFakeID(),
			},
		}

		result, err := service.TransferAccountOwnership(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.TransferAccountOwnership), testutils.ContextMatcher, mock.AnythingOfType("string"), mock.AnythingOfType("*identity.AccountOwnershipTransferInput")).Return(errors.New("transfer error"))

		request := &identitysvc.TransferAccountOwnershipRequest{
			AccountId: identityfakes.BuildFakeID(),
			Input: &identitysvc.AccountOwnershipTransferInput{
				CurrentOwner: identityfakes.BuildFakeID(),
				NewOwner:     identityfakes.BuildFakeID(),
			},
		}

		result, err := service.TransferAccountOwnership(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_UpdateAccount(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleAccountID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateAccount), testutils.ContextMatcher, mock.AnythingOfType("string"), mock.AnythingOfType("*identity.AccountUpdateRequestInput")).Return(nil)

		request := &identitysvc.UpdateAccountRequest{
			AccountId: exampleAccountID,
			Input: &identitysvc.AccountUpdateRequestInput{
				Name:         new("Updated Account Name"),
				ContactPhone: new("555-0123"),
			},
		}

		result, err := service.UpdateAccount(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.UpdateAccountRequest{
			AccountId: identityfakes.BuildFakeID(),
			Input: &identitysvc.AccountUpdateRequestInput{
				Name: new("Updated Account Name"),
			},
		}

		result, err := service.UpdateAccount(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateAccount), testutils.ContextMatcher, mock.AnythingOfType("string"), mock.AnythingOfType("*identity.AccountUpdateRequestInput")).Return(errors.New("update error"))

		request := &identitysvc.UpdateAccountRequest{
			AccountId: identityfakes.BuildFakeID(),
			Input: &identitysvc.AccountUpdateRequestInput{
				Name: new("Updated Account Name"),
			},
		}

		result, err := service.UpdateAccount(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_UpdateAccountMemberPermissions(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUserID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateAccountMemberPermissions), testutils.ContextMatcher, mock.AnythingOfType("string"), exampleUserID, mock.AnythingOfType("*identity.ModifyUserPermissionsInput")).Return(nil)

		request := &identitysvc.UpdateAccountMemberPermissionsRequest{
			UserId: exampleUserID,
			Input: &identitysvc.ModifyUserPermissionsInput{
				NewRole: "account_admin",
				Reason:  "Promotion for good work",
			},
		}

		result, err := service.UpdateAccountMemberPermissions(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with session error", func(t *testing.T) {
		t.Parallel()

		service := buildTestServiceWithSessionError(t)

		request := &identitysvc.UpdateAccountMemberPermissionsRequest{
			UserId: identityfakes.BuildFakeID(),
			Input: &identitysvc.ModifyUserPermissionsInput{
				NewRole: "account_admin",
			},
		}

		result, err := service.UpdateAccountMemberPermissions(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, grpcErr.Code())
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		identityDataManager.On(reflection.GetMethodName(identityDataManager.UpdateAccountMemberPermissions), testutils.ContextMatcher, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*identity.ModifyUserPermissionsInput")).Return(errors.New("update error"))

		request := &identitysvc.UpdateAccountMemberPermissionsRequest{
			UserId: identityfakes.BuildFakeID(),
			Input: &identitysvc.ModifyUserPermissionsInput{
				NewRole: "account_admin",
			},
		}

		result, err := service.UpdateAccountMemberPermissions(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}

func TestServiceImpl_ArchiveUserMembership(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUserID := identityfakes.BuildFakeID()
		exampleAccountID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.ArchiveUserMembership), testutils.ContextMatcher, exampleUserID, exampleAccountID).Return(nil)

		request := &identitysvc.ArchiveUserMembershipRequest{
			UserId:    exampleUserID,
			AccountId: exampleAccountID,
		}

		result, err := service.ArchiveUserMembership(t.Context(), request)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ResponseDetails)
	})

	t.Run("with error from data manager", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleUserID := identityfakes.BuildFakeID()
		exampleAccountID := identityfakes.BuildFakeID()

		identityDataManager.On(reflection.GetMethodName(identityDataManager.ArchiveUserMembership), testutils.ContextMatcher, exampleUserID, exampleAccountID).Return(errors.New("archive error"))

		request := &identitysvc.ArchiveUserMembershipRequest{
			UserId:    exampleUserID,
			AccountId: exampleAccountID,
		}

		result, err := service.ArchiveUserMembership(t.Context(), request)

		assert.Error(t, err)
		assert.Nil(t, result)

		grpcErr, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, grpcErr.Code())
	})
}
