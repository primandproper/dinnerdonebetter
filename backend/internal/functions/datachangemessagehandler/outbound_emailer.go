package datachangemessagehandler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/retry"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func (a *AsyncDataChangeMessageHandler) OutboundEmailsEventHandler(topicName string) func(ctx context.Context, rawMsg []byte) error {
	return func(ctx context.Context, rawMsg []byte) error {
		ctx, span := a.tracer.StartSpan(ctx)
		defer span.End()

		start := time.Now()
		status := statusSuccess

		defer func() {
			a.outboundEmailsExecutionTimeHistogram.Record(ctx, float64(time.Since(start).Milliseconds()),
				metric.WithAttributes(attribute.String("status", status)))
			a.recordMessagesProcessed(ctx, topicOutboundEmails, status)
		}()

		var emailMessage queuemessages.OutboundEmailMessage
		if err := json.NewDecoder(bytes.NewReader(rawMsg)).Decode(&emailMessage); err != nil {
			a.messageDecodeErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicOutboundEmails)))
			status = statusFailure
			// Unretryable: a payload that fails to decode will fail to decode on every
			// remaining attempt, and each of those is latency the healthy messages behind
			// it spend waiting. Straight to the dead-letter topic.
			return retry.Unretryable(fmt.Errorf("decoding JSON body: %w", err))
		}

		if emailMessage.TestID != "" {
			return a.handleQueueTestMessage(ctx, a.logger.WithSpan(span), span, emailMessage.TestID, topicName)
		}

		if err := a.handleEmailRequest(ctx, &emailMessage); err != nil {
			a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicOutboundEmails)))
			status = statusFailure
			return fmt.Errorf("handling outbound email request: %w", err)
		}

		return nil
	}
}

func (a *AsyncDataChangeMessageHandler) handleEmailRequest(
	ctx context.Context,
	mail *queuemessages.OutboundEmailMessage,
) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	if mail == nil {
		return errRequiredDataIsNil
	}

	if err := a.emailer.SendEmail(ctx, &mail.OutboundEmailMessage); err != nil {
		observability.AcknowledgeError(err, a.logger, span, "sending email")
		a.emailsFailedCounter.Add(ctx, 1)
		return fmt.Errorf("sending email: %w", err)
	}

	a.emailsSentCounter.Add(ctx, 1)

	if err := a.analyticsEventReporter.EventOccurred(ctx, "email_sent", mail.UserID, map[string]any{
		"toAddress":   mail.ToAddress,
		"toName":      mail.ToName,
		"fromAddress": mail.FromAddress,
		"fromName":    mail.FromName,
		"subject":     mail.Subject,
	}); err != nil {
		observability.AcknowledgeError(err, a.logger, span, "notifying customer data platform")
	}

	return nil
}
