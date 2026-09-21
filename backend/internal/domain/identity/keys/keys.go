package keys

const (
	idSuffix = ".id"

	// AccountIDKey is the standard key for referring to an account ID.
	AccountIDKey = "account" + idSuffix
	// AccountInvitationKey is the standard key for referring to an account ID.
	AccountInvitationKey = "account_invitation"
	// AccountInvitationIDKey is the standard key for referring to an account ID.
	AccountInvitationIDKey = AccountInvitationKey + idSuffix
	// DestinationAccountIDKey is the context key for the destination account ID (e.g. in invitation events).
	DestinationAccountIDKey = "destination_account"
	// AccountInvitationTokenKey carries the secret half of an invitation link on the event
	// that announces the invitation.
	//
	// It is on the event because it is the one fact about an invitation that no read can
	// hand back: the column holds a digest, and the only moment the secret exists is the
	// write that minted it. The mail this event exists to trigger cannot be composed
	// without it, which is why platform hands the unredacted invitation to the hook.
	/* #nosec G101 */
	AccountInvitationTokenKey = "account_invitation.token"
	// UserIDKey is the standard key for referring to a user ID (re-exported for domain use).
	UserIDKey = "user" + idSuffix
	// UserEmailAddressKey is the standard key for referring to a user's email address.
	UserEmailAddressKey = "user.email_address"
	// UsernameKey is the standard key for referring to a username (re-exported for domain use).
	UsernameKey = "user.username"
	// #nosec G101 UserEmailVerificationTokenKey is the standard key for referring to a username.
	UserEmailVerificationTokenKey = "user.email_verification_token"
)
