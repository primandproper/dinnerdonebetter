package datachangemessagehandler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	mealplanningnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/notifications"
	domainnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"

	"github.com/primandproper/platform-go/v12/filtering"
	notifications "github.com/primandproper/platform-go/v12/notifications/mobile"
	"github.com/primandproper/platform-go/v12/observability"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/retry"

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

		switch req.RequestType {
		case mealplanningnotifications.MobileNotificationRequestTypeMealPlanTask:
			if err := a.handleMealPlanTaskNotification(ctx, &req); err != nil {
				a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicMobileNotifications)))
				status = statusFailure
				return err
			}
			return nil
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
func (a *AsyncDataChangeMessageHandler) handleHouseholdInvitationAcceptedNotification(ctx context.Context, req *notifications.MobileNotificationRequest) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	if len(req.RecipientUserIDs) == 0 {
		return nil
	}

	deviceTokens, err := a.collectDeviceTokensForUsers(ctx, req.RecipientUserIDs)
	if err != nil {
		return observability.PrepareAndLogError(err, a.logger, span, "collecting device tokens")
	}
	if len(deviceTokens) == 0 {
		return nil
	}

	logger := a.logger.WithValue("recipient_count", len(req.RecipientUserIDs)).WithValue("device_count", len(deviceTokens))
	for _, t := range deviceTokens {
		a.sendPushToDevice(ctx, logger, t, req)
	}
	return nil
}

// sendPushToDevice sends a push notification to a device and handles BadDeviceToken by archiving.
// Returns true if the send succeeded.
func (a *AsyncDataChangeMessageHandler) sendPushToDevice(ctx context.Context, logger logging.Logger, t *domainnotifications.UserDeviceToken, req *notifications.MobileNotificationRequest) bool {
	msg := notifications.PushMessage{Title: req.Title, Body: req.Body, BadgeCount: req.BadgeCount}
	if sendErr := a.pushNotificationSender.SendPush(ctx, t.Platform, t.DeviceToken, msg); sendErr != nil {
		logger.WithValue("user_device_token_id", t.ID).WithValue("error", sendErr).Error("sending push notification to device", sendErr)
		a.pushNotificationsSentCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("request_type", req.RequestType),
			attribute.String("status", statusFailure),
		))
		if strings.Contains(sendErr.Error(), "BadDeviceToken") {
			if archiveErr := a.notificationsRepo.ArchiveUserDeviceToken(ctx, t.BelongsToUser, t.ID); archiveErr != nil {
				logger.WithValue("user_device_token_id", t.ID).Error("archiving invalid device token", archiveErr)
			} else {
				a.badDeviceTokensArchivedCounter.Add(ctx, 1)
				logger.WithValue("user_device_token_id", t.ID).Info("archived invalid device token (BadDeviceToken from APNs)")
			}
		}
		return false
	}
	a.pushNotificationsSentCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("request_type", req.RequestType),
		attribute.String("status", statusSuccess),
	))
	return true
}

func (a *AsyncDataChangeMessageHandler) collectDeviceTokensForUsers(ctx context.Context, userIDs []string) ([]*domainnotifications.UserDeviceToken, error) {
	var tokens []*domainnotifications.UserDeviceToken
	filter := filtering.DefaultQueryFilter()
	for _, userID := range userIDs {
		result, err := a.notificationsRepo.GetUserDeviceTokens(ctx, userID, filter, nil)
		if err != nil {
			return nil, fmt.Errorf("getting device tokens for user %s: %w", userID, err)
		}
		for _, t := range result.Data {
			if t != nil && t.DeviceToken != "" {
				tokens = append(tokens, t)
			}
		}
	}
	return tokens, nil
}
