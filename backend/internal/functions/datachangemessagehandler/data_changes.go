package datachangemessagehandler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	analyticsevents "github.com/primandproper/dinnerdonebetter/backend/internal/domain/analytics"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/retry"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// DataChangesEventHandler consumes the data changes topic.
//
// Everything on the topic is a webhooks.Envelope: platform's Emitter wraps every event it
// publishes — this application's own and every platform store's — in one, naming the event, so
// that a payload type serving several events (identity.UserEvent is the body of eleven) can be
// told apart. The envelope is read here, routed on its event type, and each route decodes the
// payload into the type that event names.
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

		var event webhooks.Envelope
		err := json.Unmarshal(rawMsg, &event)
		if err == nil && event.EventType == "" {
			// A body that is not an envelope is something Emit did not write: a wiring
			// mistake a consumer should hear about rather than skip.
			err = webhooks.ErrEmptyEventType
		}

		if err != nil {
			a.messageDecodeErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicDataChanges)))
			status = statusFailure
			// Unretryable: a payload that fails to decode will fail to decode on every
			// remaining attempt, and each of those is latency the healthy messages behind
			// it spend waiting. Straight to the dead-letter topic.
			return retry.Unretryable(fmt.Errorf("decoding event envelope: %w", err))
		}

		eventType = event.EventType.String()

		// The queue check's probe travels as an envelope like everything else on the topic,
		// and is acknowledged by its ID.
		if event.EventType == internalops.QueueTestProbe {
			return a.handleQueueTestMessage(ctx, a.logger.WithSpan(span), span, event.ID, topicName)
		}

		if err = a.handleDataChangeMessage(ctx, &event); err != nil {
			a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicDataChanges)))
			status = statusFailure
			return observability.PrepareAndLogError(err, a.logger, span, "handling data change message")
		}

		return nil
	}
}

func (a *AsyncDataChangeMessageHandler) handleDataChangeMessage(ctx context.Context, event *webhooks.Envelope) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	if event == nil {
		return errRequiredDataIsNil
	}

	logger := a.logger.WithValue("event_type", event.EventType.String())

	if err := a.reportToAnalytics(ctx, event); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "notifying customer data platform")
	}

	// Outbound notifications are the only fan-out left here, so it runs inline.
	//
	// Webhook deliveries used to be one of these, one queue message per subscriber; they are
	// now dispatch rows written inside the transaction that caused the event, by
	// internal/repositories/postgres/events, and claimed by the delivery worker. Search index
	// events used to be another; they are now outbox rows written by that same transaction.
	// Both moved for the same reason: a fan-out performed downstream of a commit can fail on
	// its own, leaving durable state and everything derived from it disagreeing.
	if err := a.handleOutboundNotifications(ctx, event); err != nil {
		observability.AcknowledgeError(err, logger, span, "notifying customer(s)")
	}

	return nil
}

// reportToAnalytics tells the customer data platform about the events someone has a question
// about, and no others. Which those are, and how each payload is read, is
// internal/domain/analytics's decision; this only resolves who to attribute an event to when the
// payload names an account and nobody — a subscription is the account's, and the account's owner
// is who the product counts it against.
func (a *AsyncDataChangeMessageHandler) reportToAnalytics(ctx context.Context, event *webhooks.Envelope) error {
	report, ok, err := analyticsevents.Read(event)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	userID := report.UserID
	if userID == "" && report.AccountID != "" {
		account, accountErr := a.directory.GetAccount(ctx, a.db.Reader(), tenancy.Global(), report.AccountID)
		if accountErr != nil {
			return fmt.Errorf("resolving the owner of account %s: %w", report.AccountID, accountErr)
		}

		userID = account.OwnerUserID
	}

	// An event from a background job names nobody, and the analytics platform counts people.
	if userID == "" {
		return nil
	}

	return a.analyticsEventReporter.EventOccurred(ctx, event.EventType.String(), userID, report.Properties)
}

func (a *AsyncDataChangeMessageHandler) handleOutboundNotifications(ctx context.Context, event *webhooks.Envelope) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	if event == nil {
		return fmt.Errorf("nil data change event")
	}

	for _, handler := range a.outboundNotificationHandlers {
		handled, emailType, emails, handlerErr := handler(ctx, event)
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

// payloadAs decodes the event's payload as the type its event names.
func payloadAs[T any](event *webhooks.Envelope) (*T, error) {
	var payload T
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return nil, fmt.Errorf("decoding the payload of a %q event: %w", event.EventType, err)
	}

	return &payload, nil
}

// user reads the person an event is about, for the mail it is rendered into.
func (a *AsyncDataChangeMessageHandler) user(ctx context.Context, logger logging.Logger, userID string) (*platformidentity.User, error) {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, observability.PrepareError(fmt.Errorf("event names no user"), span, "getting user")
	}

	user, err := a.directory.GetUser(ctx, a.db.Reader(), tenancy.Global(), userID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting user")
	}

	return user, nil
}

// stringFromEventContext returns a string value from one of this application's own messages'
// context. The value may be a string or []byte depending on message serialization.
func stringFromEventContext(changeMessage *datachanges.Message, key string) string {
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
