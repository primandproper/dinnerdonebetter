package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth/managers"

	"github.com/primandproper/platform-go/v6/filtering"

	"github.com/stretchr/testify/mock"
)

var (
	_ managers.AuthManagerInterface = (*AuthManager)(nil)
)

// AuthManager is a mock implementation of the auth manager.
type AuthManager struct {
	mock.Mock
}

// CheckUserPermissions is a mock method.
func (m *AuthManager) CheckUserPermissions(ctx context.Context, input *auth.UserPermissionsRequestInput) (*auth.UserPermissionsResponse, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*auth.UserPermissionsResponse), args.Error(1)
}

// TOTPSecretVerification is a mock method.
func (m *AuthManager) TOTPSecretVerification(ctx context.Context, input *auth.TOTPSecretVerificationInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// NewTOTPSecret is a mock method.
func (m *AuthManager) NewTOTPSecret(ctx context.Context, input *auth.TOTPSecretRefreshInput) (*auth.TOTPSecretRefreshResponse, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*auth.TOTPSecretRefreshResponse), args.Error(1)
}

// UpdatePassword is a mock method.
func (m *AuthManager) UpdatePassword(ctx context.Context, input *auth.PasswordUpdateInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// RequestUsernameReminder is a mock method.
func (m *AuthManager) RequestUsernameReminder(ctx context.Context, input *auth.UsernameReminderRequestInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// CreatePasswordResetToken is a mock method.
func (m *AuthManager) CreatePasswordResetToken(ctx context.Context, input *auth.PasswordResetTokenCreationRequestInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// PasswordResetTokenRedemption is a mock method.
func (m *AuthManager) PasswordResetTokenRedemption(ctx context.Context, input *auth.PasswordResetTokenRedemptionRequestInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// RequestEmailVerificationEmail is a mock method.
func (m *AuthManager) RequestEmailVerificationEmail(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// VerifyUserEmailAddress is a mock method.
func (m *AuthManager) VerifyUserEmailAddress(ctx context.Context, input *auth.EmailAddressVerificationRequestInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// VerifyUserEmailAddressByToken is a mock method.
func (m *AuthManager) VerifyUserEmailAddressByToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// GetActiveSessionsForUser is a mock method.
func (m *AuthManager) GetActiveSessionsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[auth.UserSession], error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[auth.UserSession]), args.Error(1)
}

// RevokeSession is a mock method.
func (m *AuthManager) RevokeSession(ctx context.Context, sessionID, userID string) error {
	args := m.Called(ctx, sessionID, userID)
	return args.Error(0)
}

// RevokeAllSessionsForUserExcept is a mock method.
func (m *AuthManager) RevokeAllSessionsForUserExcept(ctx context.Context, userID, currentSessionID string) error {
	args := m.Called(ctx, userID, currentSessionID)
	return args.Error(0)
}

// RevokeAllSessionsForUser is a mock method.
func (m *AuthManager) RevokeAllSessionsForUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
