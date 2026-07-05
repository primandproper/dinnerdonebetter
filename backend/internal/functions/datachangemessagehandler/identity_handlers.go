package datachangemessagehandler

import (
	"context"
	"fmt"
	"strings"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	authkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	coreemails "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/emails"
	coreindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/indexing"

	"github.com/primandproper/platform-go/v3/database/filtering"
	"github.com/primandproper/platform-go/v3/email"
	notifications "github.com/primandproper/platform-go/v3/notifications/mobile"
	"github.com/primandproper/platform-go/v3/observability"
	textsearch "github.com/primandproper/platform-go/v3/search/text"
)

// handleIdentitySearchIndexUpdate handles search index updates for identity domain events.
func (a *AsyncDataChangeMessageHandler) handleIdentitySearchIndexUpdate(
	ctx context.Context,
	changeMessage *audit.DataChangeMessage,
) (bool, error) {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	logger := a.logger.WithValue("event_type", changeMessage.EventType)

	switch changeMessage.EventType {
	case identity.UserSignedUpServiceEventType,
		identity.UserArchivedServiceEventType,
		identity.EmailAddressChangedEventType,
		identity.UsernameChangedEventType,
		identity.UserDetailsChangedEventType,
		identity.UserEmailAddressVerifiedEventType:
		if changeMessage.UserID == "" {
			observability.AcknowledgeError(errRequiredDataIsNil, logger, span, "updating search index for User")
		}

		if err := a.searchDataIndexPublisher.Publish(ctx, &textsearch.IndexRequest{
			RowID:     changeMessage.UserID,
			IndexType: coreindexing.IndexTypeUsers,
			Delete:    changeMessage.EventType == identity.UserArchivedServiceEventType,
		}); err != nil {
			return true, observability.PrepareAndLogError(err, logger, span, "publishing search index update")
		}

		return true, nil
	default:
		return false, nil
	}
}

// handleIdentityOutboundNotification handles outbound notifications for identity domain events.
func (a *AsyncDataChangeMessageHandler) handleIdentityOutboundNotification(
	ctx context.Context,
	changeMessage *audit.DataChangeMessage,
	user *identity.User,
) (
	handled bool,
	emailType string,
	outboundEmailMessages []*email.OutboundEmailMessage,
	err error,
) {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	logger := a.logger.WithValue("event_type", changeMessage.EventType)

	var (
		msg *email.OutboundEmailMessage
	)

	switch changeMessage.EventType {
	case identity.UserSignedUpServiceEventType:
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

	case identity.UserEmailAddressVerificationEmailRequestedEventType:
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

	case identity.PasswordResetTokenCreatedEventType:
		emailType = "password reset request"
		tokenID := stringFromEventContext(changeMessage, authkeys.PasswordResetTokenIDKey)
		if tokenID == "" {
			return true, emailType, nil, observability.PrepareError(fmt.Errorf("password reset token created event requires password_reset_token.id in context"), span, "building password reset email")
		}

		var prt *auth.PasswordResetToken
		prt, err = a.passwordResetTokenDataManager.GetPasswordResetTokenByID(ctx, tokenID)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "getting password reset token")
		}
		if prt == nil {
			return true, emailType, nil, observability.PrepareError(fmt.Errorf("password reset token not found"), span, "building password reset email")
		}

		msg, err = coreemails.BuildGeneratedPasswordResetTokenEmail(user, prt, a.baseURL)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "building password reset token created email")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)

	case identity.UsernameReminderRequestedEventType:
		emailType = "username reminder"
		msg, err = coreemails.BuildUsernameReminderEmail(user, a.baseURL)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "building username reminder email")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)

	case identity.PasswordResetTokenRedeemedEventType:
		emailType = "password reset token redeemed"
		msg, err = coreemails.BuildPasswordResetTokenRedeemedEmail(user, a.baseURL)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "building password reset token redemption email")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)

	case identity.PasswordChangedEventType:
		emailType = "password reset token redeemed"
		msg, err = coreemails.BuildPasswordChangedEmail(user, a.baseURL)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "building password reset token email")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)

	case identity.AccountInvitationCreatedServiceEventType:
		emailType = "account invitation created"
		invitationID := stringFromEventContext(changeMessage, identitykeys.AccountInvitationIDKey)
		destinationAccountID, ok := changeMessage.Context[identitykeys.DestinationAccountIDKey].(string)
		if !ok {
			destinationAccountID = ""
		}
		if invitationID == "" || destinationAccountID == "" {
			return true, emailType, nil, observability.PrepareError(fmt.Errorf("account invitation created event requires %s and %s in context", identitykeys.AccountInvitationIDKey, identitykeys.DestinationAccountIDKey), span, "building invite member email")
		}

		var accountInvite *identity.AccountInvitation
		accountInvite, err = a.identityRepo.GetAccountInvitationByAccountAndID(ctx, destinationAccountID, invitationID)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "getting account invitation")
		}
		if accountInvite == nil {
			return true, emailType, nil, observability.PrepareError(fmt.Errorf("account invitation not found"), span, "building invite member email")
		}

		msg, err = coreemails.BuildInviteMemberEmail(user, accountInvite, a.baseURL)
		if err != nil {
			return true, emailType, nil, observability.PrepareAndLogError(err, logger, span, "building email message")
		}

		outboundEmailMessages = append(outboundEmailMessages, msg)

	case identity.AccountInvitationAcceptedServiceEventType:
		destinationAccountID, ok := changeMessage.Context[identitykeys.DestinationAccountIDKey].(string)
		if !ok || destinationAccountID == "" {
			logger.Debug(fmt.Sprintf("account invitation accepted: missing %s in context, skipping mobile notification", identitykeys.DestinationAccountIDKey))
			return true, "", nil, nil
		}
		acceptedUserID := changeMessage.UserID

		var usersResult *filtering.QueryFilteredResult[identity.User]
		usersResult, err = a.identityRepo.GetUsersForAccount(ctx, destinationAccountID, filtering.DefaultQueryFilter())
		if err != nil {
			return true, "", nil, observability.PrepareAndLogError(err, logger, span, "getting users for account")
		}

		var recipientUserIDs []string
		for _, u := range usersResult.Data {
			if u != nil && u.ID != "" && u.ID != acceptedUserID {
				recipientUserIDs = append(recipientUserIDs, u.ID)
			}
		}
		if len(recipientUserIDs) == 0 {
			return true, "", nil, nil
		}

		displayName := "Someone"
		if user != nil {
			if user.FirstName != "" || user.LastName != "" {
				displayName = strings.TrimSpace(user.FirstName + " " + user.LastName)
			} else if user.Username != "" {
				displayName = user.Username
			}
		}

		mobileReq := &notifications.MobileNotificationRequest{
			RequestType:      identity.MobileNotificationRequestTypeHouseholdInvitationAccepted,
			RecipientUserIDs: recipientUserIDs,
			Title:            "Someone joined your household",
			Body:             fmt.Sprintf("%s joined your household", displayName),
			Context: map[string]string{
				identity.ExcludedUserIDContextKey: acceptedUserID,
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
