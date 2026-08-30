package auth

import (
	"context"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// TablePrefix namespaces platform-go's password reset token table, rendering
// ddb_password_reset_tokens.
//
// A prefix rather than the platform's empty default, and here the collision is not
// hypothetical: the platform's table is named password_reset_tokens, which is exactly the
// name 00003_auth.sql gave the hand-written one this replaces. Its DDL says CREATE TABLE
// IF NOT EXISTS, so against a database still holding the old table the platform's schema
// would be a silent no-op followed by a store reading columns that are not there. The
// migration drops the old table as well, and the prefix means a database where that drop
// somehow did not run fails loudly rather than quietly reading the wrong rows.
//
// It is referenced by the migration that creates the table and by the Store that reads and
// writes it; a prefix that differs between the two is the misconfiguration nobody notices
// until a reset link says it was never issued.
const TablePrefix = "ddb"

const (
	// TwoFactorSecretVerifiedServiceEventType indicates a user's two factor secret was verified.
	/* #nosec G101 */
	TwoFactorSecretVerifiedServiceEventType = "two_factor_secret_verified"
	// TwoFactorDeactivatedServiceEventType indicates a user's two factor secret was deactivated and verified_at timestamp was reset.
	/* #nosec G101 */
	TwoFactorDeactivatedServiceEventType = "two_factor_deactivated"
	// TwoFactorSecretChangedServiceEventType indicates a user's two factor secret was changed and verified_at timestamp was reset.
	/* #nosec G101 */
	TwoFactorSecretChangedServiceEventType = "two_factor_secret_changed"
	// PasswordResetTokenCreatedEventType indicates a user created a password reset token.
	PasswordResetTokenCreatedEventType = "password_reset_token_created"
	// PasswordResetTokenRedeemedEventType indicates a user created a password reset token.
	PasswordResetTokenRedeemedEventType = "password_reset_token_redeemed"
	// PasswordChangedEventType indicates a user changed their password.
	PasswordChangedEventType = "password_changed"
	// EmailAddressChangedEventType indicates a user changed their email address.
	EmailAddressChangedEventType = "email_address_changed"
	// UsernameChangedEventType indicates a user changed their username.
	UsernameChangedEventType = "username_changed"
	// UserAvatarChangedEventType indicates a user changed their avatar.
	UserAvatarChangedEventType = "user_avatar_changed"
	// UserDetailsChangedEventType indicates a user changed their information.
	UserDetailsChangedEventType = "user_details_changed"
	// UsernameReminderRequestedEventType indicates a user requested a username reminder.
	UsernameReminderRequestedEventType = "username_reminder_requested"
	// UserLoggedInServiceEventType indicates a user has logged in.
	UserLoggedInServiceEventType = "user_logged_in"
	// UserLoggedOutServiceEventType indicates a user has logged out.
	UserLoggedOutServiceEventType = "user_logged_out"
	// UserChangedActiveAccountServiceEventType indicates a user switched their active account.
	UserChangedActiveAccountServiceEventType = "changed_active_account"
	// UserEmailAddressVerifiedEventType indicates a user created a password reset token.
	UserEmailAddressVerifiedEventType = "user_email_address_verified"
	// UserEmailAddressVerificationEmailRequestedEventType indicates a user created a password reset token.
	UserEmailAddressVerificationEmailRequestedEventType = "user_email_address_verification_email_requested"
)

type (
	// UserStatusResponse is what we encode when a user wants to check auth status.
	UserStatusResponse struct {
		_ struct{} `json:"-"`

		UserID                   string `json:"userID"`
		AccountStatus            string `json:"accountStatus,omitempty"`
		AccountStatusExplanation string `json:"accountStatusExplanation"`
		ActiveAccount            string `json:"activeAccount,omitempty"`
		UserIsAuthenticated      bool   `json:"isAuthenticated"`
	}

	// TokenResponse is used to return a JWT to a user.
	TokenResponse struct {
		_            struct{}  `json:"-"`
		ExpiresUTC   time.Time `json:"expires"`
		UserID       string    `json:"userID"`
		AccountID    string    `json:"accountID"`
		AccessToken  string    `json:"accessToken"`
		RefreshToken string    `json:"refreshToken"`
	}

	// UserPermissionsRequestInput is what we decode when a user wants to check permission status.
	UserPermissionsRequestInput struct {
		_ struct{} `json:"-"`

		Permissions []string `json:"permissions"`
	}

	// UserPermissionsResponse is what we encode when a user wants to check permission status.
	UserPermissionsResponse struct {
		_ struct{} `json:"-"`

		Permissions map[string]bool `json:"permissions"`
	}

	// ChangeActiveAccountInput represents what a User could set as input for switching accounts.
	ChangeActiveAccountInput struct {
		_ struct{} `json:"-"`

		AccountID string `json:"accountID"`
	}

	// AuthDataService describes a structure capable of handling passwords and authorization requests.
	//
	// The four handlers are the OAuth 2.1 authorization server's HTTP surface. Registration
	// is absent on purpose: an OAuth2 client here is created through the permission-gated
	// gRPC surface rather than by an anonymous RFC 7591 POST.
	AuthDataService interface {
		AuthorizeHandler(res http.ResponseWriter, req *http.Request)
		TokenHandler(res http.ResponseWriter, req *http.Request)
		RevokeHandler(res http.ResponseWriter, req *http.Request)
		AuthorizationServerMetadataHandler(res http.ResponseWriter, req *http.Request)
	}
)

var _ validation.ValidatableWithContext = (*ChangeActiveAccountInput)(nil)

// ValidateWithContext validates a ChangeActiveAccountInput.
func (x *ChangeActiveAccountInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.AccountID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*UserPermissionsRequestInput)(nil)

// ValidateWithContext validates a UserPermissionsRequestInput.
func (x *UserPermissionsRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Permissions, validation.Required),
	)
}
