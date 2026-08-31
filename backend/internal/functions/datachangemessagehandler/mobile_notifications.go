package datachangemessagehandler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	notifications "github.com/primandproper/platform-go/v13/notifications/mobile"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/retry"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// MobileNotificationsEventHandler handles mobile notification requests from the mobile_notifications queue.
// It decodes the request, validates it, and routes to the appropriate type-specific handler.
func (a *AsyncDataChangeMessageHandler) MobileNotificationsEventHandler(topicName string) func(ctx context.Context, rawMsg []byte) error {
	return func(ctx context.Context, rawMsg []byte) error {
		ctx, span := a.tracer.StartSpan(ctx)
		defer span.End()

		start := time.Now()
		status := statusSuccess
		requestType := unknownValue

		defer func() {
			a.mobileNotificationsExecutionTimeHistogram.Record(ctx, float64(time.Since(start).Milliseconds()),
				metric.WithAttributes(
					attribute.String("status", status),
					attribute.String("request_type", requestType),
				))
			a.recordMessagesProcessed(ctx, topicMobileNotifications, status)
		}()

		var req notifications.MobileNotificationRequest
		if err := json.NewDecoder(bytes.NewReader(rawMsg)).Decode(&req); err != nil {
			a.messageDecodeErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicMobileNotifications)))
			status = statusFailure
			// Unretryable: a payload that fails to decode will fail to decode on every
			// remaining attempt, and each of those is latency the healthy messages behind
			// it spend waiting. Straight to the dead-letter topic.
			return retry.Unretryable(fmt.Errorf("decoding mobile notification request: %w", err))
		}

		if req.TestID != "" {
			return a.handleQueueTestMessage(ctx, a.logger.WithSpan(span), span, req.TestID, topicName)
		}

		if req.Title == "" || req.Body == "" {
			a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicMobileNotifications)))
			status = statusFailure
			return fmt.Errorf("title and body are required")
		}
		if req.RequestType == "" {
			a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicMobileNotifications)))
			status = statusFailure
			return fmt.Errorf("request type is required")
		}

		requestType = req.RequestType

		// Meal plan task reminders are deliberately not here. They are scheduled work rather
		// than a reaction to an event, and they are now claimed from a work queue by the job
		// that owns them — see internal/services/mealplanning/workers/
		// meal_plan_task_notifications. What is left on this topic is what genuinely arrives
		// as a message.
		switch req.RequestType {
		case identity.MobileNotificationRequestTypeHouseholdInvitationAccepted:
			if err := a.handleHouseholdInvitationAcceptedNotification(ctx, &req); err != nil {
				a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicMobileNotifications)))
				status = statusFailure
				return err
			}
			return nil
		default:
			a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicMobileNotifications)))
			status = statusFailure
			return fmt.Errorf("unknown request type: %q", req.RequestType)
		}
	}
}

// handleHouseholdInvitationAcceptedNotification sends push notifications to household members when someone joins.
// RecipientUserIDs excludes the newly accepted user; ExcludedUserIDContextKey in context is for validation.
//
// Whether anybody was actually reached is not acted on. This notification is a courtesy about
// something that has already happened, and there is nothing to leave undone: the invitation was
// accepted regardless of whose phone heard about it.
func (a *AsyncDataChangeMessageHandler) handleHouseholdInvitationAcceptedNotification(ctx context.Context, req *notifications.MobileNotificationRequest) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	msg := notifications.PushMessage{Title: req.Title, Body: req.Body, BadgeCount: req.BadgeCount}

	if _, err := a.pushFanout.Send(ctx, req.RequestType, req.RecipientUserIDs, msg); err != nil {
		return observability.PrepareAndLogError(err, a.logger, span, "sending household invitation notification")
	}

	return nil
}
