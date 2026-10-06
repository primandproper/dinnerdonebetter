package dataprivacy

import (
	"context"
	"database/sql"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	platformdataprivacy "github.com/primandproper/platform-go/v15/dataprivacy"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/email"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// newCompletionNotifier builds who the Fulfiller tells when a subject access request finishes:
// platform's EmailNotifier, addressed to the subject's own address, sent through the outbox.
//
// The Fulfiller runs in the scheduler, which has no mail provider of its own, and a mail sent from
// the operation would be a mail provider's outage reported as a fulfillment's failure. So the
// notifier's emailer queues the message on the outbound emails topic instead, in a transaction of
// its own, and the async message handler sends it — the transport every other mail here takes
// to the provider. The message carries no secret: artifacts are encrypted, so platform mints no
// download link for one, and the mail tells the subject to sign in.
func newCompletionNotifier(db database.Client, directory platformidentity.Store, topic string) (*platformdataprivacy.EmailNotifier, error) {
	emailer, err := newOutboxEmailer(db, topic)
	if err != nil {
		return nil, err
	}

	return platformdataprivacy.NewEmailNotifier(
		emailer,
		platformdataprivacy.Recipient{Address: branding.FromEmail, Name: branding.CompanyName},
		subjectRecipient(db, directory),
	)
}

// subjectRecipient resolves a subject to the address their directory row names.
//
// A subject the directory no longer has is nobody to mail rather than a failure: an erasure that
// removed the row is the system working, and platform takes a nil recipient as that decision.
func subjectRecipient(db database.Client, directory platformidentity.Store) platformdataprivacy.RecipientResolver {
	return func(ctx context.Context, subject platformdataprivacy.Subject) (*platformdataprivacy.Recipient, error) {
		user, err := directory.GetUser(ctx, db.Reader(), tenancy.Global(), subject.ID)
		switch {
		case errors.Is(err, platformidentity.ErrUserNotFound), errors.Is(err, sql.ErrNoRows):
			return nil, nil
		case err != nil:
			return nil, err
		}

		return &platformdataprivacy.Recipient{Address: user.EmailAddress, Name: user.DisplayName}, nil
	}
}

// outboxEmailer is an email.Emailer that sends nothing itself: it queues the message on the
// outbound emails topic, for the async message handler to send.
type outboxEmailer struct {
	db     database.Client
	writer *outbox.Writer
	topic  string
}

var _ email.Emailer = (*outboxEmailer)(nil)

// newOutboxEmailer builds an outboxEmailer over a writer of its own, so no side effect registered
// on the process's shared writer is handed the message.
func newOutboxEmailer(db database.Client, topic string) (*outboxEmailer, error) {
	if db == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil database client")
	}

	if topic == "" {
		return nil, platformerrors.Wrap(outbox.ErrEmptyTopic, "building the completion notifier's emailer")
	}

	writer, err := outbox.NewWriter(db.Dialect())
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the completion notifier's outbox writer")
	}

	return &outboxEmailer{db: db, writer: writer, topic: topic}, nil
}

// SendEmail implements email.Emailer.
func (e *outboxEmailer) SendEmail(ctx context.Context, msg *email.OutboundEmailMessage) error {
	if msg == nil {
		return platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil outbound email")
	}

	return e.db.WithTransaction(ctx, func(tx database.Tx) error {
		return e.writer.Enqueue(ctx, tx, outbox.Message{
			Topic:   e.topic,
			Payload: &queuemessages.OutboundEmailMessage{OutboundEmailMessage: *msg},
		})
	})
}

// outboundEmailsTopic is the topic the async message handler sends mail from.
func outboundEmailsTopic(queues *queuescfg.Config) string {
	if queues == nil {
		return ""
	}

	return queues.OutboundEmailsTopicName
}
