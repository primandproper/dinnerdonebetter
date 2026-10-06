package datachangemessagehandler

import (
	"context"
	"fmt"

	authkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"
	coreemails "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/emails"

	"github.com/primandproper/platform-go/v15/authentication/passwordreset"
	"github.com/primandproper/platform-go/v15/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/webhooks"
	notifications "github.com/primandproper/primitives-go/v2/notifications/mobile"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// handleIdentityOutboundNotification turns the events about people — platform's identity, sign-in
// and password reset events, and this application's own three mail requests — into the mail and
// the mobile notifications they imply.
//
// platform's events carry what the mail needs and nothing it does not: a registration carries its
// verification link and an invitation its token, because the store holds a digest of each and
// the event is the one place the secret survives; a reset or a verification link asked for
// later carries nothing, which is why those two mails, and the handle reminder, arrive as this
// application's own mail requests with the secret on them (authentication.SignInMailers).
func (a *AsyncDataChangeMessageHandler) handleIdentityOutboundNotification(
	ctx context.Context,
	event *webhooks.Envelope,
) (
	handled bool,
	emailType string,
	outboundEmailMessages []*queuemessages.OutboundEmailMessage,
	err error,
) {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	logger := a.logger.WithValue("event_type", event.EventType.String())

	var msg *queuemessages.OutboundEmailMessage

	switch event.EventType {
	case platformidentity.EventUserRegistered:
		emailType = "user signup"
		msg, err = a.userRegistered(ctx, logger, span, event)

	case webhooks.EventType(ddbidentity.UserEmailAddressVerificationEmailRequestedEventType):
		emailType = "email address verification"
		msg, err = a.verificationEmailRequested(ctx, logger, span, event)

	case webhooks.EventType(ddbidentity.PasswordResetTokenCreatedEventType):
		emailType = "password reset request"
		msg, err = a.passwordResetRequested(ctx, logger, span, event)

	case webhooks.EventType(ddbidentity.UsernameReminderRequestedEventType):
		emailType = "username reminder"
		msg, err = a.usernameReminderRequested(ctx, logger, span, event)

	case passwordreset.EventTokenRedeemed:
		emailType = "password reset token redeemed"
		msg, err = a.passwordResetRedeemed(ctx, logger, span, event)

	// Through either door: the signed-in one, which is signin's write, and the identity
	// service's own, which an operator's reset of somebody's password goes through.
	case signin.EventPasswordUpdated, platformidentity.EventUserPasswordChanged:
		emailType = "password changed"
		msg, err = a.passwordChanged(ctx, logger, span, event)

	case platformidentity.EventInvitationCreated:
		emailType = "account invitation created"
		msg, err = a.invitationSent(ctx, logger, span, event)

	case platformidentity.EventInvitationAccepted:
		return true, "", nil, a.invitationAccepted(ctx, logger, span, event)

	default:
		return false, "", nil, nil
	}

	if err != nil {
		return true, emailType, nil, err
	}

	if msg != nil {
		outboundEmailMessages = append(outboundEmailMessages, msg)
	}

	return true, emailType, outboundEmailMessages, nil
}

// userRegistered welcomes somebody: the analytics platform learns of them, and they are mailed
// the link that proves their address. A registration through an invitation is also that
// invitation answered yes, so the household is told as it is for an acceptance.
func (a *AsyncDataChangeMessageHandler) userRegistered(ctx context.Context, logger logging.Logger, span tracing.Span, event *webhooks.Envelope) (*queuemessages.OutboundEmailMessage, error) {
	payload, err := payloadAs[platformidentity.UserEvent](event)
	if err != nil {
		return nil, err
	}

	user, err := a.user(ctx, logger, payload.UserID)
	if err != nil {
		return nil, err
	}

	// The account and nothing else: the verification link on this event is a bearer secret
	// the vendor has no business holding.
	if err = a.analyticsEventReporter.AddUser(ctx, user.ID, map[string]any{"accountID": payload.AccountID}); err != nil {
		observability.AcknowledgeError(err, logger, span, "notifying customer data platform")
	}

	if payload.EmailAddressVerificationToken == "" {
		return nil, observability.PrepareError(fmt.Errorf("email verification token required"), span, "building address verification email")
	}

	msg, err := coreemails.BuildVerifyEmailAddressEmail(user, payload.EmailAddressVerificationToken, a.baseURL)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "building address verification email")
	}

	if payload.InvitationID != "" {
		if err = a.notifyHousehold(ctx, logger, span, payload.AccountID, user); err != nil {
			return nil, err
		}
	}

	return msg, nil
}

// verificationEmailRequested mails another verification link. The link's secret is on this
// application's own mail request, because platform's signin.EventVerificationEmailRequested
// deliberately carries none.
func (a *AsyncDataChangeMessageHandler) verificationEmailRequested(ctx context.Context, logger logging.Logger, span tracing.Span, event *webhooks.Envelope) (*queuemessages.OutboundEmailMessage, error) {
	request, user, err := a.mailRequest(ctx, logger, event)
	if err != nil {
		return nil, err
	}

	token := stringFromEventContext(request, identitykeys.UserEmailVerificationTokenKey)
	if token == "" {
		return nil, observability.PrepareError(fmt.Errorf("email verification token required"), span, "building address verification email")
	}

	msg, err := coreemails.BuildVerifyEmailAddressEmail(user, token, a.baseURL)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "building address verification email")
	}

	return msg, nil
}

// passwordResetRequested mails the reset link. The secret arrives on the mail request and there
// is nowhere to read it back from: the store holds a digest of it.
func (a *AsyncDataChangeMessageHandler) passwordResetRequested(ctx context.Context, logger logging.Logger, span tracing.Span, event *webhooks.Envelope) (*queuemessages.OutboundEmailMessage, error) {
	request, user, err := a.mailRequest(ctx, logger, event)
	if err != nil {
		return nil, err
	}

	resetToken := stringFromEventContext(request, authkeys.PasswordResetTokenSecretKey)
	if resetToken == "" {
		return nil, observability.PrepareError(fmt.Errorf("password reset mail request carries no %s", authkeys.PasswordResetTokenSecretKey), span, "building password reset email")
	}

	msg, err := coreemails.BuildGeneratedPasswordResetTokenEmail(user, resetToken, a.baseURL)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "building password reset token created email")
	}

	return msg, nil
}

func (a *AsyncDataChangeMessageHandler) usernameReminderRequested(ctx context.Context, logger logging.Logger, span tracing.Span, event *webhooks.Envelope) (*queuemessages.OutboundEmailMessage, error) {
	_, user, err := a.mailRequest(ctx, logger, event)
	if err != nil {
		return nil, err
	}

	msg, err := coreemails.BuildUsernameReminderEmail(user, a.baseURL)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "building username reminder email")
	}

	return msg, nil
}

// mailRequest decodes one of this application's own mail requests and reads the person it is for.
func (a *AsyncDataChangeMessageHandler) mailRequest(ctx context.Context, logger logging.Logger, event *webhooks.Envelope) (*datachanges.Message, *platformidentity.User, error) {
	request, err := payloadAs[datachanges.Message](event)
	if err != nil {
		return nil, nil, err
	}

	user, err := a.user(ctx, logger, request.UserID)
	if err != nil {
		return nil, nil, err
	}

	return request, user, nil
}

// passwordResetRedeemed tells somebody their password was reset through a mailed link. platform
// records the redemption on the transaction that spends the link and writes the password, so the
// mail and the write commit together or not at all.
func (a *AsyncDataChangeMessageHandler) passwordResetRedeemed(ctx context.Context, logger logging.Logger, span tracing.Span, event *webhooks.Envelope) (*queuemessages.OutboundEmailMessage, error) {
	payload, err := payloadAs[passwordreset.TokenEvent](event)
	if err != nil {
		return nil, err
	}

	user, err := a.user(ctx, logger, payload.UserID)
	if err != nil {
		return nil, err
	}

	msg, err := coreemails.BuildPasswordResetTokenRedeemedEmail(user, a.baseURL)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "building password reset token redemption email")
	}

	return msg, nil
}

// passwordChanged tells somebody their password changed. The two events that say so have
// different payload types, so each is read as its own.
func (a *AsyncDataChangeMessageHandler) passwordChanged(ctx context.Context, logger logging.Logger, span tracing.Span, event *webhooks.Envelope) (*queuemessages.OutboundEmailMessage, error) {
	var userID string

	switch event.EventType {
	case signin.EventPasswordUpdated:
		payload, err := payloadAs[signin.UserEvent](event)
		if err != nil {
			return nil, err
		}

		userID = payload.UserID
	default:
		payload, err := payloadAs[platformidentity.UserEvent](event)
		if err != nil {
			return nil, err
		}

		userID = payload.UserID
	}

	user, err := a.user(ctx, logger, userID)
	if err != nil {
		return nil, err
	}

	msg, err := coreemails.BuildPasswordChangedEmail(user, a.baseURL)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "building password changed email")
	}

	return msg, nil
}

// invitationSent mails the invitation, in the sender's name.
//
// The token comes off the event rather than off the row. The column holds a digest and no read
// fills the secret in, so the read below answers with an empty token — and a link composed from it
// would be a link that cannot be followed, with nothing reporting the difference. platform's hook
// held the invitation unredacted, which is the one moment the secret exists, and put it on the
// event for exactly this.
func (a *AsyncDataChangeMessageHandler) invitationSent(ctx context.Context, logger logging.Logger, span tracing.Span, event *webhooks.Envelope) (*queuemessages.OutboundEmailMessage, error) {
	payload, err := payloadAs[platformidentity.InvitationEvent](event)
	if err != nil {
		return nil, err
	}

	if payload.InvitationID == "" || payload.AccountID == "" {
		return nil, observability.PrepareError(fmt.Errorf("invitation created event names no invitation or no account"), span, "building invite member email")
	}

	if payload.Token == "" {
		return nil, observability.PrepareError(fmt.Errorf("invitation created event carries no token"), span, "building invite member email")
	}

	sender, err := a.user(ctx, logger, payload.FromUser)
	if err != nil {
		return nil, err
	}

	invitation, err := a.directory.GetInvitation(ctx, a.db.Reader(), tenancy.Global(), payload.InvitationID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting account invitation")
	}

	if invitation == nil {
		return nil, observability.PrepareError(fmt.Errorf("account invitation not found"), span, "building invite member email")
	}

	invitation.Token = payload.Token

	msg, err := coreemails.BuildInviteMemberEmail(sender, invitation, a.baseURL)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "building email message")
	}

	return msg, nil
}

// invitationAccepted tells the household somebody joined. The acceptor is who the invitation
// now names; platform fills ToUser in when it is answered.
func (a *AsyncDataChangeMessageHandler) invitationAccepted(ctx context.Context, logger logging.Logger, span tracing.Span, event *webhooks.Envelope) error {
	payload, err := payloadAs[platformidentity.InvitationEvent](event)
	if err != nil {
		return err
	}

	if payload.ToUser == nil || *payload.ToUser == "" {
		return observability.PrepareError(fmt.Errorf("invitation accepted event names no acceptor"), span, "publishing household invitation accepted mobile notification")
	}

	joined, err := a.user(ctx, logger, *payload.ToUser)
	if err != nil {
		return err
	}

	return a.notifyHousehold(ctx, logger, span, payload.AccountID, joined)
}

// notifyHousehold pushes "somebody joined" to every other member of the account.
func (a *AsyncDataChangeMessageHandler) notifyHousehold(ctx context.Context, logger logging.Logger, span tracing.Span, accountID string, joined *platformidentity.User) error {
	if accountID == "" {
		return observability.PrepareError(fmt.Errorf("event names no account"), span, "publishing household invitation accepted mobile notification")
	}

	// Paged to the end rather than one page: a household past the first page would
	// otherwise have its later members told nothing.
	members, err := ddbidentity.MembersOfAccount(ctx, a.directory, a.db.Reader(), accountID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "getting users for account")
	}

	var recipientUserIDs []string
	for _, memberID := range members {
		if memberID != "" && memberID != joined.ID {
			recipientUserIDs = append(recipientUserIDs, memberID)
		}
	}

	if len(recipientUserIDs) == 0 {
		return nil
	}

	// DisplayName is never empty on a read — a row whose column is blank reads its
	// handle back — so the three-way fallback this replaced had two branches that
	// could not be reached.
	displayName := "Someone"
	if joined.DisplayName != "" {
		displayName = joined.DisplayName
	}

	request := &notifications.MobileNotificationRequest{
		RequestType:      ddbidentity.MobileNotificationRequestTypeHouseholdInvitationAccepted,
		RecipientUserIDs: recipientUserIDs,
		Title:            "Someone joined your household",
		Body:             fmt.Sprintf("%s joined your household", displayName),
		Context: map[string]string{
			ddbidentity.ExcludedUserIDContextKey: joined.ID,
		},
	}

	if err = a.mobileNotificationsPublisher.Publish(ctx, request); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "publishing household invitation accepted mobile notification")
	}

	return nil
}
