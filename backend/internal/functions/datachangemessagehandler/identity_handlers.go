package datachangemessagehandler

import (
	"context"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	authkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/keys"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"
	coreemails "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/emails"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	notifications "github.com/primandproper/primitives-go/v2/notifications/mobile"
	"github.com/primandproper/primitives-go/v2/observability"
)

// handleIdentityOutboundNotification handles outbound notifications for identity domain events.
func (a *AsyncDataChangeMessageHandler) handleIdentityOutboundNotification(
	ctx context.Context,
	changeMessage *audit.DataChangeMessage,
	user *platformidentity.User,
) (
	handled bool,
	emailType string,
	outboundEmailMessages []*queuemessages.OutboundEmailMessage,
	err error,
) {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	logger := a.logger.WithValue("event_type", changeMessage.EventType)

	var (
		msg *queuemessages.OutboundEmailMessage
	)

	switch changeMessage.EventType {
	case ddbidentity.UserSignedUpServiceEventType:
		emailType = "user signup"
		if err = a.analyticsEventReporter.AddUser(ctx, changeMessage.UserID, changeMessage.Context); err != nil {
			observability.AcknowledgeError(err, logger, span, "notifying customer data platform")
		}

		emailVerificationToken := stringFromEventContext(changeMessage, identitykeys.UserEmailVerificationTokenKey)
		if emailVerificationToken == "" {
			return true, emailType, nil, observability.PrepareError(fmt.Errorf("email verification token required"), span, "building address verification email")
		}

		msg, err = coreemails.BuildVerifyEmailAddressEmail(user, emailVerificationToken, a.baseURL)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "building address verification email")
		}
		outboundEmailMessages = append(outboundEmailMessages, msg)

	case ddbidentity.UserEmailAddressVerificationEmailRequestedEventType:
		emailType = "email address verification"
		emailVerificationToken := stringFromEventContext(changeMessage, identitykeys.UserEmailVerificationTokenKey)
		if emailVerificationToken == "" {
			return true, emailType, nil, observability.PrepareError(fmt.Errorf("email verification token required"), span, "building address verification email")
		}

		msg, err = coreemails.BuildVerifyEmailAddressEmail(user, emailVerificationToken, a.baseURL)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "building address verification email")
		}
		outboundEmailMessages = append(outboundEmailMessages, msg)

	case ddbidentity.PasswordResetTokenCreatedEventType:
		emailType = "password reset request"
		// The secret arrives on the message, and there is nowhere to read it back from: the
		// store holds a digest of it. Same shape as the email verification token above.
		resetToken := stringFromEventContext(changeMessage, authkeys.PasswordResetTokenSecretKey)
		if resetToken == "" {
			return true, emailType, nil, observability.PrepareError(fmt.Errorf("password reset token created event requires %s in context", authkeys.PasswordResetTokenSecretKey), span, "building password reset email")
		}

		msg, err = coreemails.BuildGeneratedPasswordResetTokenEmail(user, resetToken, a.baseURL)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "building password reset token created email")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)

	case ddbidentity.UsernameReminderRequestedEventType:
		emailType = "username reminder"
		msg, err = coreemails.BuildUsernameReminderEmail(user, a.baseURL)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "building username reminder email")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)

	case ddbidentity.PasswordResetTokenRedeemedEventType:
		emailType = "password reset token redeemed"
		msg, err = coreemails.BuildPasswordResetTokenRedeemedEmail(user, a.baseURL)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "building password reset token redemption email")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)

	case ddbidentity.PasswordChangedEventType:
		emailType = "password reset token redeemed"
		msg, err = coreemails.BuildPasswordChangedEmail(user, a.baseURL)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "building password reset token email")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)

	case ddbidentity.AccountInvitationCreatedServiceEventType:
		emailType = "account invitation created"
		invitationID := stringFromEventContext(changeMessage, identitykeys.AccountInvitationIDKey)
		destinationAccountID, ok := changeMessage.Context[identitykeys.DestinationAccountIDKey].(string)
		if !ok {
			destinationAccountID = ""
		}
		if invitationID == "" || destinationAccountID == "" {
			return true, emailType, nil, observability.PrepareError(fmt.Errorf("account invitation created event requires %s and %s in context", identitykeys.AccountInvitationIDKey, identitykeys.DestinationAccountIDKey), span, "building invite member email")
		}

		var accountInvite *platformidentity.Invitation
		accountInvite, err = a.directory.GetInvitation(ctx, a.db.Reader(), ddbidentity.Scope(), invitationID)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "getting account invitation")
		}
		if accountInvite == nil {
			return true, emailType, nil, observability.PrepareError(fmt.Errorf("account invitation not found"), span, "building invite member email")
		}

		// The token comes off the event rather than off the row. The column holds a digest
		// and no read fills the secret in, so the read above answers with an empty token —
		// and a link composed from it would be a link that cannot be followed, with nothing
		// reporting the difference. The hook that wrote this event held the invitation
		// unredacted, which is the one moment the secret exists.
		accountInvite.Token = stringFromEventContext(changeMessage, identitykeys.AccountInvitationTokenKey)
		if accountInvite.Token == "" {
			return true, emailType, nil, observability.PrepareError(
				fmt.Errorf("account invitation created event carries no %s", identitykeys.AccountInvitationTokenKey),
				span, "building invite member email")
		}

		msg, err = coreemails.BuildInviteMemberEmail(user, accountInvite, a.baseURL)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "building email message")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)

	case ddbidentity.AccountInvitationAcceptedServiceEventType:
		destinationAccountID, ok := changeMessage.Context[identitykeys.DestinationAccountIDKey].(string)
		if !ok || destinationAccountID == "" {
			logger.Debug(fmt.Sprintf("account invitation accepted: missing %s in context, skipping mobile notification", identitykeys.DestinationAccountIDKey))
			return true, "", nil, nil
		}
		acceptedUserID := changeMessage.UserID

		// Paged to the end rather than one page: a household past the first page would
		// otherwise have its later members told nothing.
		var members []string
		members, err = ddbidentity.MembersOfAccount(ctx, a.directory, a.db.Reader(), destinationAccountID)
		if err != nil {
			return true, "", nil, observability.PrepareAndLogError(err, logger, span, "getting users for account")
		}

		var recipientUserIDs []string
		for _, memberID := range members {
			if memberID != "" && memberID != acceptedUserID {
				recipientUserIDs = append(recipientUserIDs, memberID)
			}
		}
		if len(recipientUserIDs) == 0 {
			return true, "", nil, nil
		}

		// DisplayName is never empty on a read — a row whose column is blank reads its
		// handle back — so the three-way fallback this replaced had two branches that
		// could not be reached.
		displayName := "Someone"
		if user != nil && user.DisplayName != "" {
			displayName = user.DisplayName
		}

		mobileReq := &notifications.MobileNotificationRequest{
			RequestType:      ddbidentity.MobileNotificationRequestTypeHouseholdInvitationAccepted,
			RecipientUserIDs: recipientUserIDs,
			Title:            "Someone joined your household",
			Body:             fmt.Sprintf("%s joined your household", displayName),
			Context: map[string]string{
				ddbidentity.ExcludedUserIDContextKey: acceptedUserID,
			},
		}
		if err = a.mobileNotificationsPublisher.Publish(ctx, mobileReq); err != nil {
			return true, "", nil, observability.PrepareAndLogError(err, logger, span, "publishing household invitation accepted mobile notification")
		}

		return true, "", nil, nil

	default:
		return false, "", nil, nil
	}

	return true, emailType, outboundEmailMessages, nil
}
