package authentication

import (
	"context"

	authkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/keys"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v14/authentication/passwordreset"
	"github.com/primandproper/platform-go/v14/authentication/signin"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
)

// SignInMailers hands the mail platform's sign-in and password reset doors send to this
// application's own mail pipeline: an event on the outbox, which the data change message
// handler renders into the email.
//
// A mail is not sent inside the write it follows — signin and passwordreset call a mailer once
// the write has committed — so each event is written on a transaction of its own.
type SignInMailers struct {
	db      database.Client
	emitter *events.Emitter
	logger  logging.Logger
}

var (
	_ signin.VerificationMailer   = (*SignInMailers)(nil)
	_ signin.HandleReminderMailer = (*SignInMailers)(nil)
	_ passwordreset.Mailer        = (*SignInMailers)(nil)
)

// NewSignInMailers builds the mailers platform's sign-in and password reset services send
// their links through.
func NewSignInMailers(logger logging.Logger, db database.Client, emitter *events.Emitter) *SignInMailers {
	return &SignInMailers{
		db:      db,
		emitter: emitter,
		logger:  logging.NewNamedLogger(logger, "signin_mailers"),
	}
}

// SendVerification queues the verification mail, carrying the link's secret on the event: the
// store holds only a digest, so the event is where the mail is rendered from.
func (m *SignInMailers) SendVerification(ctx context.Context, mail *signin.VerificationMail) error {
	if mail == nil || mail.User == nil {
		return platformerrors.New("verification mail names nobody")
	}

	return m.emit(ctx, mail.User.ID, ddbidentity.UserEmailAddressVerificationEmailRequestedEventType,
		map[string]any{identitykeys.UserEmailVerificationTokenKey: mail.Token})
}

// SendHandleReminder queues the username reminder mail.
func (m *SignInMailers) SendHandleReminder(ctx context.Context, mail *signin.HandleReminderMail) error {
	if mail == nil || mail.User == nil {
		return platformerrors.New("handle reminder names nobody")
	}

	return m.emit(ctx, mail.User.ID, ddbidentity.UsernameReminderRequestedEventType, nil)
}

// SendPasswordReset queues the reset mail, carrying the link's secret on the event for
// SendVerification's reason. The issuance is the only place the secret will ever exist.
func (m *SignInMailers) SendPasswordReset(ctx context.Context, mail *passwordreset.Mail) error {
	if mail == nil || mail.User == nil || mail.Issuance == nil || mail.Issuance.Token == nil {
		return platformerrors.New("password reset mail names nobody")
	}

	return m.emit(ctx, mail.User.ID, ddbidentity.PasswordResetTokenCreatedEventType,
		map[string]any{
			authkeys.PasswordResetTokenIDKey:     mail.Issuance.Token.ID,
			authkeys.PasswordResetTokenSecretKey: mail.Issuance.Secret,
		})
}

func (m *SignInMailers) emit(ctx context.Context, userID, eventType string, metadata map[string]any) error {
	logger := m.logger.WithValue(identitykeys.UserIDKey, userID)

	if err := m.db.WithTransaction(ctx, func(tx database.Tx) error {
		return m.emitter.Emit(ctx, tx, logger, eventType, "", metadata, events.WithUserID(userID))
	}); err != nil {
		return platformerrors.Wrapf(err, "queueing %s", eventType)
	}

	return nil
}
