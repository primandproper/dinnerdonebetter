package datachangemessagehandler

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v4/observability"

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
			return fmt.Errorf("decoding message body: %w", err)
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

	if changeMessage.UserID != "" && changeMessage.EventType != "" {
		if err := a.analyticsEventReporter.EventOccurred(ctx, changeMessage.EventType, changeMessage.UserID, changeMessage.Context); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "notifying customer data platform")
		}
	}

	var wg sync.WaitGroup

	wg.Go(func() {
		a.nonWebhookEventTypesHat.RLock()
		eventTypeIsValid := slices.Contains(a.nonWebhookEventTypes, changeMessage.EventType)
		a.nonWebhookEventTypesHat.RUnlock()

		if changeMessage.AccountID != "" && !eventTypeIsValid {
			relevantWebhooks, err := a.webhookRepo.GetWebhooksForAccountAndEvent(ctx, changeMessage.AccountID, changeMessage.EventType)
			if err != nil {
				observability.AcknowledgeError(err, logger, span, "getting webhooks")
				return
			}

			for _, webhook := range relevantWebhooks {
				if err = a.webhookExecutionRequestPublisher.Publish(ctx, &webhooks.WebhookExecutionRequest{
					WebhookID: webhook.ID,
					AccountID: changeMessage.AccountID,
					Payload:   changeMessage,
				}); err != nil {
					observability.AcknowledgeError(err, logger, span, "publishing webhook execution request")
				}
			}
		}
	})

	wg.Go(func() {
		if err := a.handleOutboundNotifications(ctx, changeMessage); err != nil {
			observability.AcknowledgeError(err, logger, span, "notifying customer(s)")
		}
	})

	wg.Go(func() {
		if err := a.handleSearchIndexUpdates(ctx, changeMessage); err != nil {
			observability.AcknowledgeError(err, logger, span, "updating search index)")
		}
	})

	wg.Wait()

	return nil
}

func (a *AsyncDataChangeMessageHandler) handleSearchIndexUpdates(
	ctx context.Context,
	changeMessage *audit.DataChangeMessage,
) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	for _, handler := range a.searchIndexHandlers {
		if handled, err := handler(ctx, changeMessage); handled || err != nil {
			return err
		}
	}

	a.logger.WithValue("event_type", changeMessage.EventType).Debug("event type not handled for search indexing")
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

// rowIDFromEventContext returns the row ID string from the data change message context.
// Producers publish only the ID (e.g. RecipeIDKey -> recipe.ID), not the full entity.
func rowIDFromEventContext(changeMessage *audit.DataChangeMessage, idKey string) string {
	return stringFromEventContext(changeMessage, idKey)
}
