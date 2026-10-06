package keys

const (
	idSuffix = ".id"

	// AccountIDKey is the standard key for referring to an account ID.
	AccountIDKey = "account" + idSuffix
	// AccountInvitationIDKey names the invitation an invitation mail request is for.
	AccountInvitationIDKey = "account_invitation" + idSuffix
	// AccountInvitationTokenKey carries the secret half of an invitation link on the mail
	// request that asks for the invitation mail, and nowhere else: the column holds a digest,
	// and platform hands the secret to the invitation mailer once, after the invitation commits.
	/* #nosec G101 */
	AccountInvitationTokenKey = "account_invitation.token"
	// PasskeyIDKey is the standard key for referring to a passkey's ID.
	PasskeyIDKey = "passkey" + idSuffix
	// UserIDKey is the standard key for referring to a user ID (re-exported for domain use).
	UserIDKey = "user" + idSuffix
	// ImpersonatorIDKey is the operator acting through somebody else's identity.
	ImpersonatorIDKey = "impersonator" + idSuffix
	// UserEmailAddressKey is the standard key for referring to a user's email address.
	UserEmailAddressKey = "user.email_address"
	// UsernameKey is the standard key for referring to a username (re-exported for domain use).
	UsernameKey = "user.username"
	// #nosec G101 UserEmailVerificationTokenKey is the standard key for referring to a username.
	UserEmailVerificationTokenKey = "user.email_verification_token"
)
