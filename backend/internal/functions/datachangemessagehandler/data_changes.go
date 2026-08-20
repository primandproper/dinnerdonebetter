package datachangemessagehandler

import (
	"context"
	"fmt"
	"time"

	analyticsevents "github.com/primandproper/dinnerdonebetter/backend/internal/domain/analytics"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	"github.com/primandproper/platform-go/v12/observability"
	"github.com/primandproper/platform-go/v12/retry"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func (a *AsyncDataChangeMessageHandler) DataChangesEventHandler(topicName string) func(ctx context.Context, rawMsg []byte) error {
	return func(ctx context.Context, rawMsg []byte) error {
		ctx, span := a.tracer.StartSpan(ctx)
		defer span.End()

		start := time.Now()
		status := statusSuccess
		eventType := unknownValue

		defer func() {
			a.dataChangesExecutionTimeHistogram.Record(ctx, float64(time.Since(start).Milliseconds()),
				metric.WithAttributes(
					attribute.String("status", status),
					attribute.String("event_type", eventType),
				))
			a.recordMessagesProcessed(ctx, topicDataChanges, status)
		}()

		var dataChangeMessage audit.DataChangeMessage
		if err := a.decoder.DecodeBytes(ctx, rawMsg, &dataChangeMessage); err != nil {
			a.messageDecodeErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicDataChanges)))
			status = statusFailure
			// Unretryable: a payload that fails to decode will fail to decode on every
			// remaining attempt, and each of those is latency the healthy messages behind
			// it spend waiting. Straight to the dead-letter topic.
			return retry.Unretryable(fmt.Errorf("decoding message body: %w", err))
		}

		eventType = dataChangeMessage.EventType

		if err := a.handleDataChangeMessage(ctx, &dataChangeMessage, topicName); err != nil {
			a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicDataChanges)))
			status = statusFailure
			return observability.PrepareAndLogError(err, a.logger, span, "handling data change message")
		}

		return nil
	}
}

func (a *AsyncDataChangeMessageHandler) handleDataChangeMessage(
	ctx context.Context,
	changeMessage *audit.DataChangeMessage,
	topicName string,
) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	if changeMessage == nil {
		return errRequiredDataIsNil
	}

	logger := a.logger.WithValue("event_type", changeMessage.EventType)

	// Non-empty TestID triggers queue test behavior (acknowledge and skip business logic)
	testID := changeMessage.TestID
	if testID == "" && changeMessage.Context != nil {
		if v, ok := changeMessage.Context["test_id"].(string); ok {
			testID = v
		}
	}
	if testID != "" {
		return a.handleQueueTestMessage(ctx, logger, span, testID, topicName)
	}

	// Only the events someone has a question about, rather than every event carrying a user
	// ID. See internal/domain/analytics: this consumer used to forward the whole topic, which
	// meant a third-party vendor received every create, update and archive the application
	// performed — including the authentication activity the webhook layer refuses to deliver.
	//
	// The empty-event-type check is gone because Reportable("") is false, which is the same
	// answer arrived at from the allowlist rather than from a guard beside it.
	if changeMessage.UserID != "" && analyticsevents.Reportable(changeMessage.EventType) {
		if err := a.analyticsEventReporter.EventOccurred(ctx, changeMessage.EventType, changeMessage.UserID, changeMessage.Context); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "notifying customer data platform")
		}
	}

	// Outbound notifications are the only fan-out left here, so it runs inline.
	//
	// Webhook deliveries used to be one of these, one queue message per subscriber; they are
	// now dispatch rows written inside the transaction that caused the event, by
	// internal/repositories/postgres/events, and claimed by the delivery worker. Search index
	// events used to be another; they are now outbox rows written by that same transaction.
	// Both moved for the same reason: a fan-out performed downstream of a commit can fail on
	// its own, leaving durable state and everything derived from it disagreeing.
	if err := a.handleOutboundNotifications(ctx, changeMessage); err != nil {
		observability.AcknowledgeError(err, logger, span, "notifying customer(s)")
	}

	return nil
}

func (a *AsyncDataChangeMessageHandler) handleOutboundNotifications(
	ctx context.Context,
	changeMessage *audit.DataChangeMessage,
) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	if changeMessage == nil {
		return fmt.Errorf("nil data change message")
	}

	// Events from background jobs may have no UserID; skip notifications.
	if changeMessage.UserID == "" {
		return nil
	}

	user, err := a.identityRepo.GetUser(ctx, changeMessage.UserID)
	if err != nil {
		return observability.PrepareAndLogError(err, a.logger, span, "getting user")
	}

	for _, handler := range a.outboundNotificationHandlers {
		handled, emailType, emails, handlerErr := handler(ctx, changeMessage, user)
		if handlerErr != nil {
			return handlerErr
		}
		if handled {
			if len(emails) > 0 {
				a.logger.WithValue("email_type", emailType).WithValue("outbound_emails_to_send", len(emails)).Info("publishing email requests")
			}
			for _, oem := range emails {
				if pubErr := a.outboundEmailsPublisher.Publish(ctx, oem); pubErr != nil {
					observability.AcknowledgeError(pubErr, a.logger, span, "publishing %s request email", emailType)
				}
			}
			return nil
		}
	}

	return nil
}

// stringFromEventContext returns a string value from the data change message context.
// The value may be a string or []byte depending on message serialization.
func stringFromEventContext(changeMessage *audit.DataChangeMessage, key string) string {
	if changeMessage == nil || changeMessage.Context == nil {
		return ""
	}

	v, ok := changeMessage.Context[key]
	if !ok {
		return ""
	}

	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return ""
	}
}
