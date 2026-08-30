package keys

const (
	idSuffix = ".id"

	// PasswordResetTokenKey is the standard key for referring to a password reset token's ID.
	PasswordResetTokenKey = "password_reset_token"
	// PasswordResetTokenIDKey is the standard key for referring to a password reset token's ID.
	PasswordResetTokenIDKey = PasswordResetTokenKey + idSuffix
	// PasswordResetTokenSecretKey is the standard key for referring to the reset token
	// itself — the secret that goes in the link.
	//
	// It appears in exactly one place: the context of the data change message that asks
	// for the reset email, because the store keeps only a digest and the secret exists
	// once, in the issuance. It must never be attached to a span, a log line, or an audit
	// entry. The email verification token travels the same way, for the same reason.
	/* #nosec G101 */
	PasswordResetTokenSecretKey = PasswordResetTokenKey + ".secret"

	// UserSessionKey is the standard key for referring to a user session.
	UserSessionKey = "user_session"
	// UserSessionIDKey is the standard key for referring to a user session's ID.
	UserSessionIDKey = UserSessionKey + idSuffix
)
