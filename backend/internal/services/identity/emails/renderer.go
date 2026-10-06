package emails

import (
	"context"
	"errors"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	identity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/notifications/mail"
	waitlistsgrpc "github.com/primandproper/platform-go/v15/waitlists/grpc"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/email"
	"github.com/primandproper/primitives-go/v2/retry"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/matcornic/hermes/v2"
)

var (
	// ErrUnrenderedKind indicates a queued mail of a kind this application sends none of. It is
	// unretryable: it will be exactly as unknown on every later attempt.
	ErrUnrenderedKind = errors.New("this application renders no mail of that kind")

	errNilConfirmationMail = errors.New("nil waitlist confirmation mail")
)

// Renderer is this application's wording for every mail platform's identity, sign-in, password
// reset and waitlist doors owe, for notifications/mail's Drainer to send.
//
// platform owns the transport — the QueuedMailer that puts the mail on the outbox once the write
// it follows has committed, and the Drainer that takes it off — and hands this the same value the
// synchronous seam would have been given, secret included. What a mail says is this
// application's, and so is the one read it needs: the invitation names its sender by ID.
//
// A failure that is about the mail rather than the moment — a kind this application never
// mails, a recipient whose address the mail cannot go to — is unretryable, so the Drainer's pool
// dead-letters it at once rather than spending its attempts on a mail that will never render.
type Renderer struct {
	directory identity.Store
	db        database.Client
	baseURL   string
}

var _ mail.Renderer = (*Renderer)(nil)

// NewRenderer builds the renderer. directory and db are where an invitation's sender is read
// from; baseURL is where every link points.
func NewRenderer(directory identity.Store, db database.Client, baseURL string) *Renderer {
	return &Renderer{directory: directory, db: db, baseURL: baseURL}
}

// Render implements mail.Renderer.
func (r *Renderer) Render(ctx context.Context, m *mail.Mail) (*email.OutboundEmailMessage, error) {
	if m == nil {
		return nil, retry.Unretryable(mail.ErrNilMail)
	}

	var (
		msg *queuemessages.OutboundEmailMessage
		err error
	)

	switch m.Kind {
	case mail.KindInvitation:
		msg, err = r.invitation(ctx, m.Invitation)
	case mail.KindVerification:
		msg, err = permanent(BuildVerifyEmailAddressEmail(m.Verification.User, m.Verification.Token, r.baseURL))
	case mail.KindHandleReminder:
		msg, err = permanent(BuildUsernameReminderEmail(m.HandleReminder.User, r.baseURL))
	case mail.KindPasswordReset:
		msg, err = permanent(BuildGeneratedPasswordResetTokenEmail(m.PasswordReset.User, m.PasswordReset.Issuance.Secret, r.baseURL))
	case mail.KindWaitlistConfirmation:
		msg, err = permanent(BuildWaitlistConfirmationEmail(m.WaitlistConfirmation.Mail, r.baseURL))
	default:
		// The magic link door is mounted but refused — this deployment names no magic link
		// store — so a magic link mail is one nothing here should ever have queued.
		return nil, retry.Unretryable(fmt.Errorf("%w: %s", ErrUnrenderedKind, m.Kind))
	}

	if err != nil {
		return nil, err
	}

	return &msg.OutboundEmailMessage, nil
}

// invitation renders the invitation in the sender's name. The sender is read rather than carried,
// because the invitation names them by ID; a read that fails is retried, since the directory
// being unreachable is the moment's problem rather than the mail's.
func (r *Renderer) invitation(ctx context.Context, invitationMail *identity.InvitationMail) (*queuemessages.OutboundEmailMessage, error) {
	invitation := *invitationMail.Invitation
	// The row the mailer was handed is redacted; the token is beside it, and goes in the link.
	invitation.Token = invitationMail.Token

	sender, err := r.directory.GetUser(ctx, r.db.Reader(), tenancy.Global(), invitation.FromUser)
	if err != nil {
		return nil, fmt.Errorf("reading the sender of invitation %s: %w", invitation.ID, err)
	}

	return permanent(BuildInviteMemberEmail(sender, &invitation, r.baseURL))
}

// permanent marks a builder's failure unretryable. Every builder fails only on what it was
// handed — an address not yet verified, a token missing — and a mail that will not render now
// will not render on the next attempt either.
func permanent(msg *queuemessages.OutboundEmailMessage, err error) (*queuemessages.OutboundEmailMessage, error) {
	if err != nil {
		return nil, retry.Unretryable(err)
	}

	return msg, nil
}

// BuildWaitlistConfirmationEmail renders the one message a pending signup is sent: "was this
// you?", with a link that says yes and a link that says no.
//
// It is rendered from the signup rather than from a read of a user, because a signup page's
// visitor is frequently nobody's user. Neither link is logged: they are live credentials, and
// the table holds their digests.
func BuildWaitlistConfirmationEmail(confirmation *waitlistsgrpc.ConfirmationMail, baseURL string) (*queuemessages.OutboundEmailMessage, error) {
	if confirmation == nil || confirmation.Signup == nil || confirmation.Confirm == nil || confirmation.Unsubscribe == nil {
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
						Link: confirmation.Confirm.URL,
					},
				},
				{
					Instructions: "If it wasn't, or you've changed your mind, we won't write to this address about it again:",
					Button: hermes.Button{
						Text: "Take me off the list",
						Link: confirmation.Unsubscribe.URL,
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
		ToAddress:   confirmation.Signup.Contact,
		FromAddress: branding.FromEmail,
		FromName:    branding.CompanyName,
		Subject:     fmt.Sprintf("Confirm your %s waitlist signup", branding.CompanyName),
		HTMLContent: htmlContent,
		UserID:      confirmation.Signup.Subject.ID,
	}, nil
}
