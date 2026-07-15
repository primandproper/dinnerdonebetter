package identitymock

import (
	"context"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v4/filtering"

	"github.com/stretchr/testify/mock"
)

var _ identity.Repository = (*RepositoryMock)(nil)

type RepositoryMock struct {
	mock.Mock
}

// UpdateUserAccountStatus is a mock function.
func (m *RepositoryMock) UpdateUserAccountStatus(ctx context.Context, userID string, input *identity.UserAccountStatusUpdateInput) error {
	return m.Called(ctx, userID, input).Error(0)
}

// SetUserRequiresPasswordChange is a mock function.
func (m *RepositoryMock) SetUserRequiresPasswordChange(ctx context.Context, userID string, requiresChange bool) error {
	return m.Called(ctx, userID, requiresChange).Error(0)
}

// UserRequiresPasswordChange is a mock function.
func (m *RepositoryMock) UserRequiresPasswordChange(ctx context.Context, userID string) (bool, error) {
	returnValues := m.Called(ctx, userID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// AccountExists is a mock function.
func (m *RepositoryMock) AccountExists(ctx context.Context, accountID, userID string) (bool, error) {
	returnValues := m.Called(ctx, accountID, userID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetAccount is a mock function.
func (m *RepositoryMock) GetAccount(ctx context.Context, accountID string) (*identity.Account, error) {
	returnValues := m.Called(ctx, accountID)
	return returnValues.Get(0).(*identity.Account), returnValues.Error(1)
}

// GetAllAccounts is a mock function.
func (m *RepositoryMock) GetAllAccounts(ctx context.Context, results chan []*identity.Account, bucketSize uint16) error {
	return m.Called(ctx, results, bucketSize).Error(0)
}

// GetAccounts is a mock function.
func (m *RepositoryMock) GetAccounts(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.Account], error) {
	returnValues := m.Called(ctx, userID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[identity.Account]), returnValues.Error(1)
}

// CreateAccount is a mock function.
func (m *RepositoryMock) CreateAccount(ctx context.Context, input *identity.AccountDatabaseCreationInput) (*identity.Account, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*identity.Account), returnValues.Error(1)
}

// UpdateAccount is a mock function.
func (m *RepositoryMock) UpdateAccount(ctx context.Context, updated *identity.Account) error {
	return m.Called(ctx, updated).Error(0)
}

// UpdateAccountBillingFields is a mock function.
func (m *RepositoryMock) UpdateAccountBillingFields(ctx context.Context, accountID string, billingStatus, subscriptionPlanID, paymentProcessorCustomerID *string, lastPaymentProviderSyncOccurredAt *time.Time) error {
	return m.Called(ctx, accountID, billingStatus, subscriptionPlanID, paymentProcessorCustomerID, lastPaymentProviderSyncOccurredAt).Error(0)
}

// ArchiveAccount is a mock function.
func (m *RepositoryMock) ArchiveAccount(ctx context.Context, accountID, userID string) error {
	return m.Called(ctx, accountID, userID).Error(0)
}

// AccountInvitationExists is a mock function.
func (m *RepositoryMock) AccountInvitationExists(ctx context.Context, accountInvitationID string) (bool, error) {
	returnValues := m.Called(ctx, accountInvitationID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetAccountInvitationByAccountAndID is a mock function.
func (m *RepositoryMock) GetAccountInvitationByAccountAndID(ctx context.Context, accountID, accountInvitationID string) (*identity.AccountInvitation, error) {
	returnValues := m.Called(ctx, accountID, accountInvitationID)
	return returnValues.Get(0).(*identity.AccountInvitation), returnValues.Error(1)
}

// GetAccountInvitationByTokenAndID is a mock function.
func (m *RepositoryMock) GetAccountInvitationByTokenAndID(ctx context.Context, accountInvitationID, token string) (*identity.AccountInvitation, error) {
	returnValues := m.Called(ctx, accountInvitationID, token)
	return returnValues.Get(0).(*identity.AccountInvitation), returnValues.Error(1)
}

// GetAccountInvitationByEmailAndToken is a mock function.
func (m *RepositoryMock) GetAccountInvitationByEmailAndToken(ctx context.Context, emailAddress, token string) (*identity.AccountInvitation, error) {
	returnValues := m.Called(ctx, emailAddress, token)
	return returnValues.Get(0).(*identity.AccountInvitation), returnValues.Error(1)
}

// GetPendingAccountInvitationsFromUser is a mock function.
func (m *RepositoryMock) GetPendingAccountInvitationsFromUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
	returnValues := m.Called(ctx, userID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[identity.AccountInvitation]), returnValues.Error(1)
}

// GetPendingAccountInvitationsForUser is a mock function.
func (m *RepositoryMock) GetPendingAccountInvitationsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.AccountInvitation], error) {
	returnValues := m.Called(ctx, userID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[identity.AccountInvitation]), returnValues.Error(1)
}

// CreateAccountInvitation is a mock function.
func (m *RepositoryMock) CreateAccountInvitation(ctx context.Context, input *identity.AccountInvitationDatabaseCreationInput) (*identity.AccountInvitation, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*identity.AccountInvitation), returnValues.Error(1)
}

// CancelAccountInvitation is a mock function.
func (m *RepositoryMock) CancelAccountInvitation(ctx context.Context, accountID, accountInvitationID, note string) error {
	return m.Called(ctx, accountID, accountInvitationID, note).Error(0)
}

// AcceptAccountInvitation is a mock function.
func (m *RepositoryMock) AcceptAccountInvitation(ctx context.Context, accountID, accountInvitationID, token, note string) error {
	return m.Called(ctx, accountID, accountInvitationID, token, note).Error(0)
}

// RejectAccountInvitation is a mock function.
func (m *RepositoryMock) RejectAccountInvitation(ctx context.Context, accountID, accountInvitationID, note string) error {
	return m.Called(ctx, accountID, accountInvitationID, note).Error(0)
}

// ArchiveAccountInvitation is a mock function.
func (m *RepositoryMock) ArchiveAccountInvitation(ctx context.Context, accountID, userID string) error {
	return m.Called(ctx, accountID, userID).Error(0)
}

// GetPasswordResetTokenByToken implements our interface requirements.
func (m *RepositoryMock) GetPasswordResetTokenByToken(ctx context.Context, passwordResetTokenID string) (*identity.PasswordResetToken, error) {
	returnValues := m.Called(ctx, passwordResetTokenID)

	return returnValues.Get(0).(*identity.PasswordResetToken), returnValues.Error(1)
}

// CreatePasswordResetToken implements our interface requirements.
func (m *RepositoryMock) CreatePasswordResetToken(ctx context.Context, input *identity.PasswordResetTokenDatabaseCreationInput) (*identity.PasswordResetToken, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*identity.PasswordResetToken), returnValues.Error(1)
}

// RedeemPasswordResetToken implements our interface requirements.
func (m *RepositoryMock) RedeemPasswordResetToken(ctx context.Context, passwordResetTokenID string) error {
	return m.Called(ctx, passwordResetTokenID).Error(0)
}

// SetUserAvatar is a mock function.
func (m *RepositoryMock) SetUserAvatar(ctx context.Context, userID, uploadedMediaID string) error {
	return m.Called(ctx, userID, uploadedMediaID).Error(0)
}

// UpdateUserUsername is a mock function.
func (m *RepositoryMock) UpdateUserUsername(ctx context.Context, userID, newUsername string) error {
	return m.Called(ctx, userID, newUsername).Error(0)
}

// UpdateUserEmailAddress is a mock function.
func (m *RepositoryMock) UpdateUserEmailAddress(ctx context.Context, userID, newEmailAddress string) error {
	return m.Called(ctx, userID, newEmailAddress).Error(0)
}

// UpdateUserDetails is a mock function.
func (m *RepositoryMock) UpdateUserDetails(ctx context.Context, userID string, input *identity.UserDetailsDatabaseUpdateInput) error {
	return m.Called(ctx, userID, input).Error(0)
}

// GetUserByEmailAddressVerificationToken is a mock function.
func (m *RepositoryMock) GetUserByEmailAddressVerificationToken(ctx context.Context, token string) (*identity.User, error) {
	returnValues := m.Called(ctx, token)
	return returnValues.Get(0).(*identity.User), returnValues.Error(1)
}

// MarkUserEmailAddressAsVerified is a mock function.
func (m *RepositoryMock) MarkUserEmailAddressAsVerified(ctx context.Context, userID, token string) error {
	return m.Called(ctx, userID, token).Error(0)
}

// GetUser is a mock function.
func (m *RepositoryMock) GetUser(ctx context.Context, userID string) (*identity.User, error) {
	returnValues := m.Called(ctx, userID)
	return returnValues.Get(0).(*identity.User), returnValues.Error(1)
}

// GetUserWithUnverifiedTwoFactorSecret is a mock function.
func (m *RepositoryMock) GetUserWithUnverifiedTwoFactorSecret(ctx context.Context, userID string) (*identity.User, error) {
	returnValues := m.Called(ctx, userID)
	return returnValues.Get(0).(*identity.User), returnValues.Error(1)
}

// GetUserByEmail is a mock function.
func (m *RepositoryMock) GetUserByEmail(ctx context.Context, email string) (*identity.User, error) {
	returnValues := m.Called(ctx, email)
	return returnValues.Get(0).(*identity.User), returnValues.Error(1)
}

// MarkUserTwoFactorSecretAsVerified is a mock function.
func (m *RepositoryMock) MarkUserTwoFactorSecretAsVerified(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

// MarkUserEmailAddressAsUnverified is a mock function.
func (m *RepositoryMock) MarkUserEmailAddressAsUnverified(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

// MarkUserTwoFactorSecretAsUnverified is a mock function.
func (m *RepositoryMock) MarkUserTwoFactorSecretAsUnverified(ctx context.Context, userID, newSecret string) error {
	return m.Called(ctx, userID, newSecret).Error(0)
}

// GetUserByUsername is a mock function.
func (m *RepositoryMock) GetUserByUsername(ctx context.Context, username string) (*identity.User, error) {
	returnValues := m.Called(ctx, username)
	return returnValues.Get(0).(*identity.User), returnValues.Error(1)
}

// GetAdminUserByUsername is a mock function.
func (m *RepositoryMock) GetAdminUserByUsername(ctx context.Context, username string) (*identity.User, error) {
	returnValues := m.Called(ctx, username)
	return returnValues.Get(0).(*identity.User), returnValues.Error(1)
}

// SearchForUsersByUsername is a mock function.
func (m *RepositoryMock) SearchForUsersByUsername(ctx context.Context, usernameQuery string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
	returnValues := m.Called(ctx, usernameQuery, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[identity.User]), returnValues.Error(1)
}

// GetUsersWithIDs is a mock function.
func (m *RepositoryMock) GetUsersWithIDs(ctx context.Context, ids []string) ([]*identity.User, error) {
	returnValues := m.Called(ctx, ids)
	return returnValues.Get(0).([]*identity.User), returnValues.Error(1)
}

// GetUsers is a mock function.
func (m *RepositoryMock) GetUsers(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[identity.User]), returnValues.Error(1)
}

// GetUsersForAccount is a mock function.
func (m *RepositoryMock) GetUsersForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[identity.User], error) {
	returnValues := m.Called(ctx, accountID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[identity.User]), returnValues.Error(1)
}

// CreateUser is a mock function.
func (m *RepositoryMock) CreateUser(ctx context.Context, input *identity.UserDatabaseCreationInput) (*identity.User, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*identity.User), returnValues.Error(1)
}

// UpdateUserPassword is a mock function.
func (m *RepositoryMock) UpdateUserPassword(ctx context.Context, userID, newHash string) error {
	return m.Called(ctx, userID, newHash).Error(0)
}

// ArchiveUser is a mock function.
func (m *RepositoryMock) ArchiveUser(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

// DeleteUser is a mock function.
func (m *RepositoryMock) DeleteUser(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

// GetEmailAddressVerificationTokenForUser is a mock function.
func (m *RepositoryMock) GetEmailAddressVerificationTokenForUser(ctx context.Context, userID string) (string, error) {
	returnValues := m.Called(ctx, userID)
	return returnValues.String(0), returnValues.Error(1)
}

// GetUserIDsThatNeedSearchIndexing is a mock function.
func (m *RepositoryMock) GetUserIDsThatNeedSearchIndexing(ctx context.Context) ([]string, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]string), returnValues.Error(1)
}

func (m *RepositoryMock) MarkUserAsIndexed(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

// BuildSessionContextDataForUser satisfies our interface contract.
func (m *RepositoryMock) BuildSessionContextDataForUser(ctx context.Context, userID, activeAccountID string) (*sessions.ContextData, error) {
	returnValues := m.Called(ctx, userID, activeAccountID)

	return returnValues.Get(0).(*sessions.ContextData), returnValues.Error(1)
}

// GetDefaultAccountIDForUser satisfies our interface contract.
func (m *RepositoryMock) GetDefaultAccountIDForUser(ctx context.Context, userID string) (string, error) {
	returnValues := m.Called(ctx, userID)

	return returnValues.Get(0).(string), returnValues.Error(1)
}

// MarkAccountAsUserDefault implements the interface.
func (m *RepositoryMock) MarkAccountAsUserDefault(ctx context.Context, userID, accountID string) error {
	return m.Called(ctx, userID, accountID).Error(0)
}

// UserIsMemberOfAccount implements the interface.
func (m *RepositoryMock) UserIsMemberOfAccount(ctx context.Context, userID, accountID string) (bool, error) {
	returnValues := m.Called(ctx, userID, accountID)

	return returnValues.Bool(0), returnValues.Error(1)
}

// AddUserToAccount implements the interface.
func (m *RepositoryMock) AddUserToAccount(ctx context.Context, input *identity.AccountUserMembershipDatabaseCreationInput) error {
	return m.Called(ctx, input).Error(0)
}

// RemoveUserFromAccount implements the interface.
func (m *RepositoryMock) RemoveUserFromAccount(ctx context.Context, userID, accountID string) error {
	return m.Called(ctx, userID, accountID).Error(0)
}

// ModifyUserPermissions implements the interface.
func (m *RepositoryMock) ModifyUserPermissions(ctx context.Context, accountID, userID string, input *identity.ModifyUserPermissionsInput) error {
	return m.Called(ctx, userID, accountID, input).Error(0)
}

// TransferAccountOwnership implements the interface.
func (m *RepositoryMock) TransferAccountOwnership(ctx context.Context, accountID string, input *identity.AccountOwnershipTransferInput) error {
	return m.Called(ctx, accountID, input).Error(0)
}

// CreateWebAuthnCredential is a mock function.
func (m *RepositoryMock) CreateWebAuthnCredential(ctx context.Context, input *identity.WebAuthnCredentialCreationInput) (*identity.WebAuthnCredential, error) {
	returnValues := m.Called(ctx, input)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*identity.WebAuthnCredential), returnValues.Error(1)
}

// GetWebAuthnCredentialsForUser is a mock function.
func (m *RepositoryMock) GetWebAuthnCredentialsForUser(ctx context.Context, userID string) ([]*identity.WebAuthnCredential, error) {
	returnValues := m.Called(ctx, userID)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).([]*identity.WebAuthnCredential), returnValues.Error(1)
}

// GetWebAuthnCredentialByCredentialID is a mock function.
func (m *RepositoryMock) GetWebAuthnCredentialByCredentialID(ctx context.Context, credentialID []byte) (*identity.WebAuthnCredential, error) {
	returnValues := m.Called(ctx, credentialID)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*identity.WebAuthnCredential), returnValues.Error(1)
}

// UpdateWebAuthnCredentialSignCount is a mock function.
func (m *RepositoryMock) UpdateWebAuthnCredentialSignCount(ctx context.Context, id string, signCount uint32) error {
	return m.Called(ctx, id, signCount).Error(0)
}

// ArchiveWebAuthnCredential is a mock function.
func (m *RepositoryMock) ArchiveWebAuthnCredential(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

// ArchiveWebAuthnCredentialForUser is a mock function.
func (m *RepositoryMock) ArchiveWebAuthnCredentialForUser(ctx context.Context, id, userID string) error {
	return m.Called(ctx, id, userID).Error(0)
}
