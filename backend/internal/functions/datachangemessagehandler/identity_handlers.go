package datachangemessagehandler

import (
	"context"
	"fmt"

	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
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
)

// handleIdentityOutboundNotification turns platform's identity, sign-in and password reset events
// into the mail and the mobile notifications they imply.
//
// platform's events carry no secret, so they drive only the mail and pushes that need none. Every
// mail that carries a link — a verification link, at registration or asked for again, a reset
// link, a username reminder, an invitation — is handed by platform to notifications/mail's
// QueuedMailer once its write commits, and is rendered and sent by the mail Drainer this process
// runs on its own topic. None of it passes through here.
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
		return true, "", nil, a.userRegistered(ctx, logger, span, event)

	case passwordreset.EventTokenRedeemed:
		emailType = "password reset token redeemed"
		msg, err = a.passwordResetRedeemed(ctx, logger, span, event)

	// Through either door: the signed-in one, which is signin's write, and the identity
	// service's own, which an operator's reset of somebody's password goes through.
	case signin.EventPasswordUpdated, platformidentity.EventUserPasswordChanged:
		emailType = "password changed"
		msg, err = a.passwordChanged(ctx, logger, span, event)

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

// userRegistered welcomes somebody: the analytics platform learns of them. A registration
// through an invitation is also that invitation answered yes, so the household is told as it is
// for an acceptance. The verification mail is not this event's: signin mails the first link
// through the QueuedMailer after the registration commits, and the mail Drainer sends it.
func (a *AsyncDataChangeMessageHandler) userRegistered(ctx context.Context, logger logging.Logger, span tracing.Span, event *webhooks.Envelope) error {
	payload, err := payloadAs[platformidentity.UserEvent](event)
	if err != nil {
		return err
	}

	user, err := a.user(ctx, logger, payload.UserID)
	if err != nil {
		return err
	}

	if err = a.analyticsEventReporter.AddUser(ctx, user.ID, map[string]any{"accountID": payload.AccountID}); err != nil {
		observability.AcknowledgeError(err, logger, span, "notifying customer data platform")
	}

	if payload.InvitationID != "" {
		if err = a.notifyHousehold(ctx, logger, span, payload.AccountID, user); err != nil {
			return err
		}
	}

	return nil
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
