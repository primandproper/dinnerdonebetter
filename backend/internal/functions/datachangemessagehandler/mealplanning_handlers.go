package datachangemessagehandler

import (
	"context"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	eatingemails "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/emails"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/notifications"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	notifications "github.com/primandproper/platform-go/v10/notifications/mobile"
	"github.com/primandproper/platform-go/v10/observability"
)

// handleMealPlanTaskNotification processes a meal plan task reminder notification.
// It performs idempotency checks, sends push notifications to recipients, and marks the task as notified.
func (a *AsyncDataChangeMessageHandler) handleMealPlanTaskNotification(ctx context.Context, req *notifications.MobileNotificationRequest) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	mealPlanTaskID := ""
	if req.Context != nil {
		mealPlanTaskID = req.Context[mealplanningnotifications.MealPlanTaskIDContextKey]
	}
	if mealPlanTaskID == "" {
		return fmt.Errorf("meal plan task notification requires mealPlanTaskID in context")
	}

	logger := a.logger.WithValue(mealplanningkeys.MealPlanTaskIDKey, mealPlanTaskID)

	sent, err := a.mealPlanRepo.MealPlanTaskNotificationHasBeenSent(ctx, mealPlanTaskID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "checking if notification already sent")
	}
	if sent {
		return nil // idempotent skip
	}

	if len(req.RecipientUserIDs) == 0 {
		return a.mealPlanRepo.MarkMealPlanTaskNotificationSent(ctx, mealPlanTaskID)
	}

	deviceTokens, err := a.collectDeviceTokensForUsers(ctx, req.RecipientUserIDs)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "collecting device tokens")
	}
	if len(deviceTokens) == 0 {
		return a.mealPlanRepo.MarkMealPlanTaskNotificationSent(ctx, mealPlanTaskID)
	}

	atLeastOneSent := false
	for _, t := range deviceTokens {
		if a.sendPushToDevice(ctx, logger, t, req) {
			atLeastOneSent = true
		}
	}

	if !atLeastOneSent {
		return nil // Don't mark as notified if every attempt failed; scheduler will retry
	}
	return a.mealPlanRepo.MarkMealPlanTaskNotificationSent(ctx, mealPlanTaskID)
}

// handleMealPlanningOutboundNotification handles outbound notifications for meal planning domain events.
func (a *AsyncDataChangeMessageHandler) handleMealPlanningOutboundNotification(
	ctx context.Context,
	changeMessage *audit.DataChangeMessage,
	_ *identity.User,
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

	account, err := a.identityRepo.GetAccount(ctx, mealPlan.BelongsToAccount)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting account")
	}

	var outboundEmailMessages []*queuemessages.OutboundEmailMessage
	for _, member := range account.Members {
		if member.BelongsToUser.EmailAddressVerifiedAt != nil {
			msg, emailErr := eatingemails.BuildMealPlanCreatedEmail(member.BelongsToUser, mealPlan, a.baseURL)
			if emailErr != nil {
				return nil, observability.PrepareAndLogError(emailErr, logger, span, "building meal plan created email")
			}

			outboundEmailMessages = append(outboundEmailMessages, msg)
		}
	}

	return outboundEmailMessages, nil
}
