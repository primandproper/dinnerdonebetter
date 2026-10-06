package identitystore

import (
	"context"

	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
)

// InvitationMailer hands the invitation mail to this application's own mail pipeline: a mail
// request on the outbox, carrying the link's token, which the data change message handler
// renders into the email.
//
// platform calls it once the invitation has committed, and is the only party that hands over the
// token: the column holds a digest, identity.EventInvitationCreated carries none, and with a
// mailer configured the hooks see the invitation redacted. So the request is written on a
// transaction of its own, and is the one place the secret goes besides the invitee's inbox.
type InvitationMailer struct {
	db      database.Client
	emitter *events.Emitter
	logger  logging.Logger
}

var _ platformidentity.InvitationMailer = (*InvitationMailer)(nil)

// NewInvitationMailer builds the mailer platform's identity service sends invitations through.
func NewInvitationMailer(logger logging.Logger, db database.Client, emitter *events.Emitter) (*InvitationMailer, error) {
	switch {
	case db == nil:
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil database client")
	case emitter == nil:
		return nil, ErrNilEmitter
	}

	return &InvitationMailer{db: db, emitter: emitter, logger: logging.NewNamedLogger(logger, "invitation_mailer")}, nil
}

// SendInvitation queues the invitation mail in the sender's name, on the account it invites to.
func (m *InvitationMailer) SendInvitation(ctx context.Context, mail *platformidentity.InvitationMail) error {
	if mail == nil || mail.Invitation == nil || mail.Token == "" {
		return platformerrors.New("invitation mail names no invitation or carries no token")
	}

	invitation := mail.Invitation
	logger := m.logger.WithValue(identitykeys.AccountInvitationIDKey, invitation.ID)

	if err := m.db.WithTransaction(ctx, func(tx database.Tx) error {
		return m.emitter.Emit(ctx, tx, logger, ddbidentity.AccountInvitationMailRequestedEventType, invitation.BelongsToAccount,
			map[string]any{
				identitykeys.AccountInvitationIDKey:    invitation.ID,
				identitykeys.AccountInvitationTokenKey: mail.Token,
			},
			events.WithUserID(invitation.FromUser),
		)
	}); err != nil {
		return platformerrors.Wrap(err, "queueing the invitation mail")
	}

	return nil
}
