package waitlists

import (
	"context"
	"errors"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	"github.com/primandproper/platform-go/v14/outbox"
	waitlistsgrpc "github.com/primandproper/platform-go/v14/waitlists/grpc"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/matcornic/hermes/v2"
)

var (
	errNilConfirmationMail  = errors.New("nil waitlist confirmation mail")
	errNoOutboundEmailTopic = errors.New("no outbound email topic is configured")
)

// confirmationMailer delivers the one message a pending signup is sent: "was this you?",
// with a link that says yes and a link that says no.
//
// It queues the mail rather than sending it. platform calls this after the signup has
// committed and reports an error here as the Join's, so a send made inline would turn a mail
// provider's bad minute into a signup form that fails. The message goes on the outbox, in a
// transaction of its own, on the topic every other outbound mail travels — the async message
// handler delivers it, and retries it, exactly as it does a password reset.
//
// The message is rendered here rather than downstream, which is the difference from the mail
// that follows a data change event. Those are built by the handler from the event and a read
// of the user it names; a signup page's visitor is frequently nobody's user, so there is no
// read to make, and the two links are live credentials that exist at this moment and never
// again — the table holds digests.
//
// Neither link is logged. They travel as far as the outbox row and the mail rendered from it,
// which is the same distance an invitation's token travels.
type confirmationMailer struct {
	db      database.Client
	writer  *outbox.Writer
	topic   string
	baseURL string
}

var _ waitlistsgrpc.ConfirmationMailer = (*confirmationMailer)(nil)

func newConfirmationMailer(db database.Client, writer *outbox.Writer, topic, baseURL string) (*confirmationMailer, error) {
	if topic == "" {
		return nil, errNoOutboundEmailTopic
	}

	return &confirmationMailer{db: db, writer: writer, topic: topic, baseURL: baseURL}, nil
}

// SendConfirmation queues the confirmation mail for a pending signup.
func (m *confirmationMailer) SendConfirmation(ctx context.Context, _ tenancy.Scope, mail *waitlistsgrpc.ConfirmationMail) error {
	message, err := buildConfirmationEmail(mail, m.baseURL)
	if err != nil {
		return err
	}

	return m.db.WithTransaction(ctx, func(tx database.Tx) error {
		return m.writer.Enqueue(ctx, tx, outbox.Message{
			Topic:   m.topic,
			Payload: message,
			Key:     mail.Signup.ID,
		})
	})
}

// buildConfirmationEmail renders the confirmation mail for a pending signup.
func buildConfirmationEmail(mail *waitlistsgrpc.ConfirmationMail, baseURL string) (*queuemessages.OutboundEmailMessage, error) {
	if mail == nil || mail.Signup == nil || mail.Confirm == nil || mail.Unsubscribe == nil {
		return nil, errNilConfirmationMail
	}

	e := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				fmt.Sprintf("Somebody asked for this address to join a %s waitlist.", branding.CompanyName),
			},
			Actions: []hermes.Action{
				{
					Instructions: "If it was you, confirm it and we'll keep your place:",
					Button: hermes.Button{
						Text: "Confirm my signup",
						Link: mail.Confirm.URL,
					},
				},
				{
					Instructions: "If it wasn't, or you've changed your mind, we won't write to this address about it again:",
					Button: hermes.Button{
						Text: "Take me off the list",
						Link: mail.Unsubscribe.URL,
					},
				},
			},
			Outros: []string{
				"Until the signup is confirmed, nobody is waiting on it.",
			},
		},
	}

	htmlContent, err := branding.BuildHermes(baseURL).GenerateHTML(e)
	if err != nil {
		return nil, fmt.Errorf("error rendering email template: %w", err)
	}

	return &queuemessages.OutboundEmailMessage{
		ToAddress:   mail.Signup.Contact,
		FromAddress: branding.FromEmail,
		FromName:    branding.CompanyName,
		Subject:     fmt.Sprintf("Confirm your %s waitlist signup", branding.CompanyName),
		HTMLContent: htmlContent,
		UserID:      mail.Signup.Subject.ID,
	}, nil
}
