package identity

// The mail requests this application queues for platform's identity, sign-in and password reset
// doors, through authentication.SignInMailers and identitystore.InvitationMailer.
//
// They are not store events. The write each one follows is recorded by platform's own hooks —
// signin.EventVerificationEmailRequested, passwordreset.EventTokenIssued,
// identity.EventInvitationCreated — and those events
// deliberately carry no secret. The mail cannot be rendered without one: the store holds a
// digest of the link's token, and the issuance is the only moment the secret exists. So
// platform hands the secret to a mailer once the write has committed, and the mailer puts it on
// one of these, for the data change message handler to render. Every one of them is Internal in
// the webhook catalog for that reason; see internal/domain/webhooks/catalog.
const (
	// PasswordResetTokenCreatedEventType indicates a password reset mail was requested; it carries the reset link.
	PasswordResetTokenCreatedEventType = "password_reset_token_created"
	// UsernameReminderRequestedEventType indicates a username reminder mail was requested.
	UsernameReminderRequestedEventType = "username_reminder_requested"
	// UserEmailAddressVerificationEmailRequestedEventType indicates another verification mail was requested; it carries the verification link.
	UserEmailAddressVerificationEmailRequestedEventType = "user_email_address_verification_email_requested"
	// AccountInvitationMailRequestedEventType indicates an invitation mail was requested; it carries the invitation link.
	AccountInvitationMailRequestedEventType = "account_invitation_mail_requested"
)
