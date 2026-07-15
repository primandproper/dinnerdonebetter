package mock

import (
	"context"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager"

	"github.com/primandproper/platform-go/v4/filtering"

	"github.com/stretchr/testify/mock"
)

var _ manager.IdentityDataManager = (*IdentityDataManager)(nil)

// IdentityDataManager is a mock type for the IdentityDataManager interface.
type IdentityDataManager struct {
	mock.Mock
}

// AcceptAccountInvitation is a mock function.
func (m *IdentityDataManager) AcceptAccountInvitation(ctx context.Context, accountID, accountInvitationID string, input *identity.AccountInvitationUpdateRequestInput) error {
	return m.Called(ctx, accountID, accountInvitationID, input).Error(0)
}

// RejectAccountInvitation is a mock function.
func (m *IdentityDataManager) RejectAccountInvitation(ctx context.Context, accountID, accountInvitationID string, input *identity.AccountInvitationUpdateRequestInput) error {
	return m.Called(ctx, accountID, accountInvitationID, input).Error(0)
}

// CancelAccountInvitation is a mock function.
func (m *IdentityDataManager) CancelAccountInvitation(ctx context.Context, accountID, accountInvitationID, note string) error {
	return m.Called(ctx, accountID, accountInvitationID, note).Error(0)
}

// ArchiveAccount is a mock function.
func (m *IdentityDataManager) ArchiveAccount(ctx context.Context, accountID, ownerID string) error {
	return m.Called(ctx, accountID, ownerID).Error(0)
}

// ArchiveUserMembership is a mock function.
func (m *IdentityDataManager) ArchiveUserMembership(ctx context.Context, userID, accountID string) error {
	return m.Called(ctx, userID, accountID).Error(0)
}

// ArchiveUser is a mock function.
func (m *IdentityDataManager) ArchiveUser(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

// CreateAccount is a mock function.
func (m *IdentityDataManager) CreateAccount(ctx context.Context, input *identity.AccountCreationRequestInput) (*identity.Account, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*identity.Account), returnValues.Error(1)
}

// CreateAccountInvitation is a mock function.
func (m *IdentityDataManager) CreateAccountInvitation(ctx context.Context, userID, accountID string, input *identity.AccountInvitationCreationRequestInput) (*identity.AccountInvitation, error) {
	returnValues := m.Called(ctx, userID, accountID, input)
	return returnValues.Get(0).(*identity.AccountInvitation), returnValues.Error(1)
}

// CreateUser is a mock function.
func (m *IdentityDataManager) CreateUser(ctx context.Context, registrationInput *identity.UserRegistrationInput) (*identity.UserCreationResponse, error) {
	returnValues := m.Called(ctx, registrationInput)
	return returnValues.Get(0).(*identity.UserCreationResponse), returnValues.Error(1)
}

// GetAccount is a mock function.
func (m *IdentityDataManager) GetAccount(ctx context.Context, accountID string) (*identity.Account, error) {
	returnValues := m.Called(ctx, accountID)
	return returnValues.Get(0).(*identity.Account), returnValues.Error(1)
}

// GetAccountInvitation is a mock function.
func (m *IdentityDataManager) GetAccountInvitation(ctx context.Context, accountID, accountInvitationID string) (*identity.AccountInvitation, error) {
	returnValues := m.Called(ctx, accountID, accountInvitationID)
	return returnValues.Get(0).(*identity.AccountInvitation), returnValues.Error(1)
}

// GetAccounts is a mock function.
func (m *IdentityDataManager) GetAccounts(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
	returnValues := m.Called(ctx, userID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[identity.Account]), returnValues.Error(1)
}

// GetReceivedAccountInvitations is a mock function.
func (m *IdentityDataManager) GetReceivedAccountInvitations(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
	returnValues := m.Called(ctx, userID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[identity.AccountInvitation]), returnValues.Error(1)
}

// GetSentAccountInvitations is a mock function.
func (m *IdentityDataManager) GetSentAccountInvitations(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
	returnValues := m.Called(ctx, userID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[identity.AccountInvitation]), returnValues.Error(1)
}

// GetUser is a mock function.
func (m *IdentityDataManager) GetUser(ctx context.Context, userID string) (*identity.User, error) {
	returnValues := m.Called(ctx, userID)
	return returnValues.Get(0).(*identity.User), returnValues.Error(1)
}

// GetUserByEmail is a mock function.
func (m *IdentityDataManager) GetUserByEmail(ctx context.Context, email string) (*identity.User, error) {
	returnValues := m.Called(ctx, email)
	return returnValues.Get(0).(*identity.User), returnValues.Error(1)
}

// GetUserByUsername is a mock function.
func (m *IdentityDataManager) GetUserByUsername(ctx context.Context, username string) (*identity.User, error) {
	returnValues := m.Called(ctx, username)
	return returnValues.Get(0).(*identity.User), returnValues.Error(1)
}

// GetAdminUserByUsername is a mock function.
func (m *IdentityDataManager) GetAdminUserByUsername(ctx context.Context, username string) (*identity.User, error) {
	returnValues := m.Called(ctx, username)
	return returnValues.Get(0).(*identity.User), returnValues.Error(1)
}

// GetDefaultAccountIDForUser is a mock function.
func (m *IdentityDataManager) GetDefaultAccountIDForUser(ctx context.Context, userID string) (string, error) {
	returnValues := m.Called(ctx, userID)
	return returnValues.String(0), returnValues.Error(1)
}

// BuildSessionContextDataForUser is a mock function.
func (m *IdentityDataManager) BuildSessionContextDataForUser(ctx context.Context, userID, activeAccountID string) (*sessions.ContextData, error) {
	returnValues := m.Called(ctx, userID, activeAccountID)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*sessions.ContextData), returnValues.Error(1)
}

// GetUsers is a mock function.
func (m *IdentityDataManager) GetUsers(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[identity.User]), returnValues.Error(1)
}

// GetUsersForAccount is a mock function.
func (m *IdentityDataManager) GetUsersForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
	returnValues := m.Called(ctx, accountID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[identity.User]), returnValues.Error(1)
}

// SearchForUsers is a mock function.
func (m *IdentityDataManager) SearchForUsers(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
	returnValues := m.Called(ctx, query, useSearchService, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[identity.User]), returnValues.Error(1)
}

// SetDefaultAccount is a mock function.
func (m *IdentityDataManager) SetDefaultAccount(ctx context.Context, userID, accountID string) error {
	return m.Called(ctx, userID, accountID).Error(0)
}

// TransferAccountOwnership is a mock function.
func (m *IdentityDataManager) TransferAccountOwnership(ctx context.Context, accountID string, input *identity.AccountOwnershipTransferInput) error {
	return m.Called(ctx, accountID, input).Error(0)
}

// UpdateAccount is a mock function.
func (m *IdentityDataManager) UpdateAccount(ctx context.Context, accountID string, input *identity.AccountUpdateRequestInput) error {
	return m.Called(ctx, accountID, input).Error(0)
}

// UpdateAccountBillingFields is a mock function.
func (m *IdentityDataManager) UpdateAccountBillingFields(ctx context.Context, accountID string, billingStatus, subscriptionPlanID, paymentProcessorCustomerID *string, lastPaymentProviderSyncOccurredAt *time.Time) error {
	return m.Called(ctx, accountID, billingStatus, subscriptionPlanID, paymentProcessorCustomerID, lastPaymentProviderSyncOccurredAt).Error(0)
}

// UpdateAccountMemberPermissions is a mock function.
func (m *IdentityDataManager) UpdateAccountMemberPermissions(ctx context.Context, userID, accountID string, input *identity.ModifyUserPermissionsInput) error {
	return m.Called(ctx, userID, accountID, input).Error(0)
}

// UpdateUserDetails is a mock function.
func (m *IdentityDataManager) UpdateUserDetails(ctx context.Context, userID string, input *identity.UserDetailsUpdateRequestInput) error {
	return m.Called(ctx, userID, input).Error(0)
}

// UpdateUserEmailAddress is a mock function.
func (m *IdentityDataManager) UpdateUserEmailAddress(ctx context.Context, userID, newEmail string) error {
	return m.Called(ctx, userID, newEmail).Error(0)
}

// UpdateUserUsername is a mock function.
func (m *IdentityDataManager) UpdateUserUsername(ctx context.Context, userID, newUsername string) error {
	return m.Called(ctx, userID, newUsername).Error(0)
}

// SetUserAvatar is a mock function.
func (m *IdentityDataManager) SetUserAvatar(ctx context.Context, userID, uploadedMediaID string) error {
	return m.Called(ctx, userID, uploadedMediaID).Error(0)
}

// AdminUpdateUserStatus is a mock function.
func (m *IdentityDataManager) AdminUpdateUserStatus(ctx context.Context, input *identity.UserAccountStatusUpdateInput) error {
	return m.Called(ctx, input).Error(0)
}

// AdminSetPasswordChangeRequired is a mock function.
func (m *IdentityDataManager) AdminSetPasswordChangeRequired(ctx context.Context, userID string, requiresChange bool) error {
	return m.Called(ctx, userID, requiresChange).Error(0)
}

// UserRequiresPasswordChange is a mock function.
func (m *IdentityDataManager) UserRequiresPasswordChange(ctx context.Context, userID string) (bool, error) {
	returnValues := m.Called(ctx, userID)
	return returnValues.Bool(0), returnValues.Error(1)
}
