package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	identitysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/identity"

	"github.com/primandproper/platform-go/v7/filtering"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServiceImpl_ArchiveAccount(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, identityDataManager := buildTestService(t)

		exampleAccountID := identityfakes.BuildFakeID()

		identityDataManager.ArchiveAccountFunc = func(_ context.Context, accountID string, _ string) error {
			assert.Equal(t, exampleAccountID, accountID)

			return nil
		}

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

		identityDataManager.ArchiveAccountFunc = func(_ context.Context, accountID string, _ string) error {
			assert.Equal(t, exampleAccountID, accountID)

			return errors.New("database error")
		}

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

		identityDataManager.CreateAccountFunc = func(_ context.Context, input *identity.AccountCreationRequestInput) (*identity.Account, error) {
			assert.True(t, input.Name == exampleAccount.Name && input.BelongsToUser != "")

			return exampleAccount, nil
		}

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

		identityDataManager.CreateAccountFunc = func(_ context.Context, _ *identity.AccountCreationRequestInput) (*identity.Account, error) {
			return nil, errors.New("creation error")
		}

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

		identityDataManager.CreateAccountInvitationFunc = func(_ context.Context, _ string, _ string, input *identity.AccountInvitationCreationRequestInput) (*identity.AccountInvitation, error) {
			assert.True(t, input.ToEmail == exampleInvitation.ToEmail && input.ToName == exampleInvitation.ToName)

			return exampleInvitation, nil
		}

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

		identityDataManager.CreateAccountInvitationFunc = func(_ context.Context, _ string, _ string, _ *identity.AccountInvitationCreationRequestInput) (*identity.AccountInvitation, error) {
			return nil, errors.New("creation error")
		}

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

		exampleAccount := identityfakes.BuildFakeAccount()

		service, identityDataManager := buildTestServiceWithAccountMembership(t, exampleAccount.ID)

		identityDataManager.GetAccountFunc = func(_ context.Context, accountID string) (*identity.Account, error) {
			assert.Equal(t, exampleAccount.ID, accountID)

			return exampleAccount, nil
		}

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

		exampleAccountID := identityfakes.BuildFakeID()

		service, identityDataManager := buildTestServiceWithAccountMembership(t, exampleAccountID)

		identityDataManager.GetAccountFunc = func(_ context.Context, accountID string) (*identity.Account, error) {
			assert.Equal(t, exampleAccountID, accountID)

			return nil, errors.New("database error")
		}

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

		identityDataManager.GetAccountsFunc = func(_ context.Context, _ string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
			return exampleAccounts, nil
		}

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

		identityDataManager.GetAccountsFunc = func(_ context.Context, _ string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
			return nil, errors.New("database error")
		}

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

		identityDataManager.SetDefaultAccountFunc = func(_ context.Context, _ string, accountID string) error {
			assert.Equal(t, exampleAccountID, accountID)

			return nil
		}

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

		identityDataManager.SetDefaultAccountFunc = func(_ context.Context, _ string, accountID string) error {
			assert.Equal(t, exampleAccountID, accountID)

			return errors.New("update error")
		}

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

		identityDataManager.TransferAccountOwnershipFunc = func(_ context.Context, _ string, _ *identity.AccountOwnershipTransferInput) error {
			return nil
		}

		// AccountId is left unset so it derives from the authenticated session's active account.
		request := &identitysvc.TransferAccountOwnershipRequest{
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

		identityDataManager.TransferAccountOwnershipFunc = func(_ context.Context, _ string, _ *identity.AccountOwnershipTransferInput) error {
			return errors.New("transfer error")
		}

		// AccountId is left unset so it derives from the authenticated session's active account.
		request := &identitysvc.TransferAccountOwnershipRequest{
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

		identityDataManager.UpdateAccountFunc = func(_ context.Context, _ string, _ *identity.AccountUpdateRequestInput) error {
			return nil
		}

		// AccountId is left unset so it derives from the authenticated session's active account.
		request := &identitysvc.UpdateAccountRequest{
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

		identityDataManager.UpdateAccountFunc = func(_ context.Context, _ string, _ *identity.AccountUpdateRequestInput) error {
			return errors.New("update error")
		}

		// AccountId is left unset so it derives from the authenticated session's active account.
		request := &identitysvc.UpdateAccountRequest{
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

		identityDataManager.UpdateAccountMemberPermissionsFunc = func(_ context.Context, _ string, accountID string, _ *identity.ModifyUserPermissionsInput) error {
			assert.Equal(t, exampleUserID, accountID)

			return nil
		}

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

		identityDataManager.UpdateAccountMemberPermissionsFunc = func(_ context.Context, _ string, _ string, _ *identity.ModifyUserPermissionsInput) error {
			return errors.New("update error")
		}

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

		// the account arg is derived from the authenticated session, not the request.
		identityDataManager.ArchiveUserMembershipFunc = func(_ context.Context, userID string, _ string) error {
			assert.Equal(t, exampleUserID, userID)

			return nil
		}

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

		// the account arg is derived from the authenticated session, not the request.
		identityDataManager.ArchiveUserMembershipFunc = func(_ context.Context, userID string, _ string) error {
			assert.Equal(t, exampleUserID, userID)

			return errors.New("archive error")
		}

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
