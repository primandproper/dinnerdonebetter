package managers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"

	"github.com/primandproper/platform-go/v7/filtering"
)

var (
	_ AuthManagerInterface = (*AuthManager)(nil)
)

// AuthManagerInterface defines the methods that the auth manager must implement.
type AuthManagerInterface interface {
	CheckUserPermissions(ctx context.Context, input *auth.UserPermissionsRequestInput) (*auth.UserPermissionsResponse, error)
	PasswordResetTokenRedemption(ctx context.Context, input *auth.PasswordResetTokenRedemptionRequestInput) error
	NewTOTPSecret(ctx context.Context, input *auth.TOTPSecretRefreshInput) (*auth.TOTPSecretRefreshResponse, error)
	RequestEmailVerificationEmail(ctx context.Context) error
	CreatePasswordResetToken(ctx context.Context, input *auth.PasswordResetTokenCreationRequestInput) error
	RequestUsernameReminder(ctx context.Context, input *auth.UsernameReminderRequestInput) error
	VerifyUserEmailAddress(ctx context.Context, input *auth.EmailAddressVerificationRequestInput) error
	VerifyUserEmailAddressByToken(ctx context.Context, token string) error
	TOTPSecretVerification(ctx context.Context, input *auth.TOTPSecretVerificationInput) error
	UpdatePassword(ctx context.Context, input *auth.PasswordUpdateInput) error
	GetActiveSessionsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[auth.UserSession], error)
	RevokeSession(ctx context.Context, sessionID, userID string) error
	RevokeAllSessionsForUserExcept(ctx context.Context, userID, currentSessionID string) error
	RevokeAllSessionsForUser(ctx context.Context, userID string) error
}
