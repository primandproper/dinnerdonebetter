package keys

const (
	idSuffix = ".id"

	// AccountIDKey is the standard key for referring to an account ID.
	AccountIDKey = "account" + idSuffix
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
)
