package identity

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
	// PasswordResetTokenRedeemedEventType indicates a user redeemed a password reset token.
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
	// UserImpersonatedServiceEventType indicates an operator was issued a token to act as a user.
	UserImpersonatedServiceEventType = "user_impersonated"
	// UserLoggedOutServiceEventType indicates a user has logged out.
	UserLoggedOutServiceEventType = "user_logged_out"
	// UserChangedActiveAccountServiceEventType indicates a user switched their active account.
	UserChangedActiveAccountServiceEventType = "changed_active_account"
	// UserEmailAddressVerifiedEventType indicates a user verified their email address.
	UserEmailAddressVerifiedEventType = "user_email_address_verified"
	// UserEmailAddressVerificationEmailRequestedEventType indicates a user requested an email address verification email.
	UserEmailAddressVerificationEmailRequestedEventType = "user_email_address_verification_email_requested"
)
