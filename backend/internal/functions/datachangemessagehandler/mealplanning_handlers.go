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
	eatingindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

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

// handleMealPlanningSearchIndexUpdate handles search index updates for meal planning domain events.
// Returns true if the event was handled, false if it should fall through to other handlers.
func (a *AsyncDataChangeMessageHandler) handleMealPlanningSearchIndexUpdate(
	ctx context.Context,
	changeMessage *audit.DataChangeMessage,
) (bool, error) {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	logger := a.logger.WithValue("event_type", changeMessage.EventType)

	switch changeMessage.EventType {
	case mealplanning.RecipeCreatedServiceEventType,
		mealplanning.RecipeUpdatedServiceEventType,
		mealplanning.RecipeArchivedServiceEventType:
		rowID := rowIDFromEventContext(changeMessage, mealplanningkeys.RecipeIDKey)
		if rowID == "" {
			return true, observability.PrepareAndLogError(errRequiredDataIsNil, logger, span, "updating search index for Recipe")
		}
		if err := a.publishIndexEvent(ctx, eatingindexing.IndexTypeRecipes, rowID, changeMessage.EventType == mealplanning.RecipeArchivedServiceEventType); err != nil {
			return true, observability.PrepareAndLogError(err, logger, span, "publishing search index update")
		}

		return true, nil
	case mealplanning.MealCreatedServiceEventType,
		mealplanning.MealUpdatedServiceEventType,
		mealplanning.MealArchivedServiceEventType:
		rowID := rowIDFromEventContext(changeMessage, mealplanningkeys.MealIDKey)
		if rowID == "" {
			return true, observability.PrepareAndLogError(errRequiredDataIsNil, logger, span, "updating search index for Meal")
		}
		if err := a.publishIndexEvent(ctx, eatingindexing.IndexTypeMeals, rowID, changeMessage.EventType == mealplanning.MealArchivedServiceEventType); err != nil {
			return true, observability.PrepareAndLogError(err, logger, span, "publishing search index update")
		}

		return true, nil
	case mealplanning.ValidIngredientCreatedServiceEventType,
		mealplanning.ValidIngredientUpdatedServiceEventType,
		mealplanning.ValidIngredientArchivedServiceEventType:
		rowID := rowIDFromEventContext(changeMessage, mealplanningkeys.ValidIngredientIDKey)
		if rowID == "" {
			return true, observability.PrepareAndLogError(errRequiredDataIsNil, logger, span, "updating search index for ValidIngredient")
		}
		if err := a.publishIndexEvent(ctx, eatingindexing.IndexTypeValidIngredients, rowID, changeMessage.EventType == mealplanning.ValidIngredientArchivedServiceEventType); err != nil {
			return true, observability.PrepareAndLogError(err, logger, span, "publishing search index update")
		}

		return true, nil
	case mealplanning.ValidInstrumentCreatedServiceEventType,
		mealplanning.ValidInstrumentUpdatedServiceEventType,
		mealplanning.ValidInstrumentArchivedServiceEventType:
		rowID := rowIDFromEventContext(changeMessage, mealplanningkeys.ValidInstrumentIDKey)
		if rowID == "" {
			return true, observability.PrepareAndLogError(errRequiredDataIsNil, logger, span, "updating search index for ValidInstrument")
		}
		if err := a.publishIndexEvent(ctx, eatingindexing.IndexTypeValidInstruments, rowID, changeMessage.EventType == mealplanning.ValidInstrumentArchivedServiceEventType); err != nil {
			return true, observability.PrepareAndLogError(err, logger, span, "publishing search index update")
		}

		return true, nil
	case mealplanning.ValidMeasurementUnitCreatedServiceEventType,
		mealplanning.ValidMeasurementUnitUpdatedServiceEventType,
		mealplanning.ValidMeasurementUnitArchivedServiceEventType:
		rowID := rowIDFromEventContext(changeMessage, mealplanningkeys.ValidMeasurementUnitIDKey)
		if rowID == "" {
			return true, observability.PrepareAndLogError(errRequiredDataIsNil, logger, span, "updating search index for ValidMeasurementUnit")
		}
		if err := a.publishIndexEvent(ctx, eatingindexing.IndexTypeValidMeasurementUnits, rowID, changeMessage.EventType == mealplanning.ValidMeasurementUnitArchivedServiceEventType); err != nil {
			return true, observability.PrepareAndLogError(err, logger, span, "publishing search index update")
		}

		return true, nil
	case mealplanning.ValidPreparationCreatedServiceEventType,
		mealplanning.ValidPreparationUpdatedServiceEventType,
		mealplanning.ValidPreparationArchivedServiceEventType:
		rowID := rowIDFromEventContext(changeMessage, mealplanningkeys.ValidPreparationIDKey)
		if rowID == "" {
			return true, observability.PrepareAndLogError(errRequiredDataIsNil, logger, span, "updating search index for ValidPreparation")
		}
		if err := a.publishIndexEvent(ctx, eatingindexing.IndexTypeValidPreparations, rowID, changeMessage.EventType == mealplanning.ValidPreparationArchivedServiceEventType); err != nil {
			return true, observability.PrepareAndLogError(err, logger, span, "publishing search index update")
		}

		return true, nil
	case mealplanning.ValidIngredientStateCreatedServiceEventType,
		mealplanning.ValidIngredientStateUpdatedServiceEventType,
		mealplanning.ValidIngredientStateArchivedServiceEventType:
		rowID := rowIDFromEventContext(changeMessage, mealplanningkeys.ValidIngredientStateIDKey)
		if rowID == "" {
			return true, observability.PrepareAndLogError(errRequiredDataIsNil, logger, span, "updating search index for ValidIngredientState")
		}
		if err := a.publishIndexEvent(ctx, eatingindexing.IndexTypeValidIngredientStates, rowID, changeMessage.EventType == mealplanning.ValidIngredientStateArchivedServiceEventType); err != nil {
			return true, observability.PrepareAndLogError(err, logger, span, "publishing search index update")
		}

		return true, nil
	case mealplanning.ValidIngredientMeasurementUnitCreatedServiceEventType,
		mealplanning.ValidIngredientMeasurementUnitUpdatedServiceEventType,
		mealplanning.ValidIngredientMeasurementUnitArchivedServiceEventType:
		rowID := rowIDFromEventContext(changeMessage, mealplanningkeys.ValidIngredientMeasurementUnitIDKey)
		if rowID == "" {
			return true, observability.PrepareAndLogError(errRequiredDataIsNil, logger, span, "updating search index for ValidIngredientMeasurementUnit")
		}
		if err := a.publishIndexEvent(ctx, eatingindexing.IndexTypeRecipes, rowID, changeMessage.EventType == mealplanning.ValidIngredientMeasurementUnitArchivedServiceEventType); err != nil {
			return true, observability.PrepareAndLogError(err, logger, span, "publishing search index update")
		}

		return true, nil
	case mealplanning.ValidPreparationInstrumentCreatedServiceEventType,
		mealplanning.ValidPreparationInstrumentUpdatedServiceEventType,
		mealplanning.ValidPreparationInstrumentArchivedServiceEventType:
		rowID := rowIDFromEventContext(changeMessage, mealplanningkeys.ValidPreparationInstrumentIDKey)
		if rowID == "" {
			return true, observability.PrepareAndLogError(errRequiredDataIsNil, logger, span, "updating search index for ValidPreparationInstrument")
		}
		if err := a.publishIndexEvent(ctx, eatingindexing.IndexTypeRecipes, rowID, changeMessage.EventType == mealplanning.ValidPreparationInstrumentArchivedServiceEventType); err != nil {
			return true, observability.PrepareAndLogError(err, logger, span, "publishing search index update")
		}

		return true, nil
	case mealplanning.ValidIngredientPreparationCreatedServiceEventType,
		mealplanning.ValidIngredientPreparationUpdatedServiceEventType,
		mealplanning.ValidIngredientPreparationArchivedServiceEventType:
		rowID := rowIDFromEventContext(changeMessage, mealplanningkeys.ValidIngredientPreparationIDKey)
		if rowID == "" {
			return true, observability.PrepareAndLogError(errRequiredDataIsNil, logger, span, "updating search index for ValidIngredientPreparation")
		}
		if err := a.publishIndexEvent(ctx, eatingindexing.IndexTypeRecipes, rowID, changeMessage.EventType == mealplanning.ValidIngredientPreparationArchivedServiceEventType); err != nil {
			return true, observability.PrepareAndLogError(err, logger, span, "publishing search index update")
		}

		return true, nil
	default:
		return false, nil
	}
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
