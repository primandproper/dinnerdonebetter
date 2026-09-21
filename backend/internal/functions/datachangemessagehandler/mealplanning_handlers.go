package datachangemessagehandler

import (
	"context"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityroster "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	eatingemails "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/emails"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/observability"
)

// handleMealPlanningOutboundNotification handles outbound notifications for meal planning domain events.
func (a *AsyncDataChangeMessageHandler) handleMealPlanningOutboundNotification(
	ctx context.Context,
	changeMessage *audit.DataChangeMessage,
	_ *platformidentity.User,
) (
	handled bool,
	emailType string,
	outgoingMessages []*queuemessages.OutboundEmailMessage,
	err error,
) {
	if changeMessage.EventType != mealplanning.MealPlanCreatedServiceEventType {
		return false, "", nil, nil
	}

	msgs, err := a.handleMealPlanCreatedNotification(ctx, changeMessage)
	if err != nil {
		return true, "meal plan created", nil, err
	}

	return true, "meal plan created", msgs, nil
}

// handleMealPlanCreatedNotification builds email notifications for a newly created meal plan.
func (a *AsyncDataChangeMessageHandler) handleMealPlanCreatedNotification(
	ctx context.Context,
	changeMessage *audit.DataChangeMessage,
) ([]*queuemessages.OutboundEmailMessage, error) {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	logger := a.logger.WithValue("event_type", changeMessage.EventType)

	mealPlanID, ok := changeMessage.Context[mealplanningkeys.MealPlanIDKey].(string)
	if !ok {
		mealPlanID = ""
	}
	if mealPlanID == "" || changeMessage.AccountID == "" {
		return nil, observability.PrepareError(fmt.Errorf("meal plan created event requires meal_plan.id and accountID in context"), span, "publishing meal plan created email")
	}

	mealPlan, err := a.mealPlanRepo.GetMealPlan(ctx, mealPlanID, changeMessage.AccountID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting meal plan for created email")
	}
	if mealPlan == nil {
		return nil, observability.PrepareError(fmt.Errorf("meal plan is nil"), span, "publishing meal plan created email")
	}

	// The roster, paged to the end. The read this replaced answered with an Account
	// carrying its whole membership list, so a household past the first page would have
	// had its later members silently left off the mailing.
	members, err := identityroster.MembersOfAccount(ctx, a.directory, a.db.Reader(), mealPlan.BelongsToAccount)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting account members")
	}

	users, err := a.directory.ListUsersByIDs(ctx, a.db.Reader(), ddbidentity.Scope(), members)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting account members")
	}

	var outboundEmailMessages []*queuemessages.OutboundEmailMessage
	for _, member := range users {
		if member.EmailAddressVerifiedAt != nil {
			msg, emailErr := eatingemails.BuildMealPlanCreatedEmail(member, mealPlan, a.baseURL)
			if emailErr != nil {
				return nil, observability.PrepareAndLogError(emailErr, logger, span, "building meal plan created email")
			}

			outboundEmailMessages = append(outboundEmailMessages, msg)
		}
	}

	return outboundEmailMessages, nil
}
