package manager

import (
	"context"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v5/filtering"
)

type (
	IdentityDataManager interface {
		AcceptAccountInvitation(ctx context.Context, accountID, accountInvitationID string, input *identity.AccountInvitationUpdateRequestInput) error
		RejectAccountInvitation(ctx context.Context, accountID, accountInvitationID string, input *identity.AccountInvitationUpdateRequestInput) error
		CancelAccountInvitation(ctx context.Context, accountID, accountInvitationID, note string) error
		ArchiveAccount(ctx context.Context, accountID, ownerID string) error
		ArchiveUserMembership(ctx context.Context, userID, accountID string) error
		ArchiveUser(ctx context.Context, userID string) error
		CreateAccount(ctx context.Context, input *identity.AccountCreationRequestInput) (*identity.Account, error)
		CreateAccountInvitation(ctx context.Context, userID, accountID string, input *identity.AccountInvitationCreationRequestInput) (*identity.AccountInvitation, error)
		CreateUser(ctx context.Context, registrationInput *identity.UserRegistrationInput) (*identity.UserCreationResponse, error)
		GetAccount(ctx context.Context, accountID string) (*identity.Account, error)
		GetAccountInvitation(ctx context.Context, accountID, accountInvitationID string) (*identity.AccountInvitation, error)
		GetAccounts(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error)
		GetReceivedAccountInvitations(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error)
		GetSentAccountInvitations(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error)
		GetUser(ctx context.Context, userID string) (*identity.User, error)
		GetUserByEmail(ctx context.Context, email string) (*identity.User, error)
		GetUserByUsername(ctx context.Context, username string) (*identity.User, error)
		GetAdminUserByUsername(ctx context.Context, username string) (*identity.User, error)
		GetUsers(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error)
		GetUsersForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error)
		SearchForUsers(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error)
		SetDefaultAccount(ctx context.Context, userID, accountID string) error
		TransferAccountOwnership(ctx context.Context, accountID string, input *identity.AccountOwnershipTransferInput) error
		UpdateAccount(ctx context.Context, accountID string, input *identity.AccountUpdateRequestInput) error
		UpdateAccountBillingFields(ctx context.Context, accountID string, billingStatus, subscriptionPlanID, paymentProcessorCustomerID *string, lastPaymentProviderSyncOccurredAt *time.Time) error
		UpdateAccountMemberPermissions(ctx context.Context, userID, accountID string, input *identity.ModifyUserPermissionsInput) error
		UpdateUserDetails(ctx context.Context, userID string, input *identity.UserDetailsUpdateRequestInput) error
		UpdateUserEmailAddress(ctx context.Context, userID, newEmail string) error
		UpdateUserUsername(ctx context.Context, userID, newUsername string) error
		SetUserAvatar(ctx context.Context, userID, uploadedMediaID string) error
		AdminUpdateUserStatus(ctx context.Context, input *identity.UserAccountStatusUpdateInput) error
		AdminSetPasswordChangeRequired(ctx context.Context, userID string, requiresChange bool) error
		UserRequiresPasswordChange(ctx context.Context, userID string) (bool, error)

		// Session/context helpers used by auth service and interceptors.
		GetDefaultAccountIDForUser(ctx context.Context, userID string) (string, error)
		BuildSessionContextDataForUser(ctx context.Context, userID, activeAccountID string) (*sessions.ContextData, error)
	}
)
