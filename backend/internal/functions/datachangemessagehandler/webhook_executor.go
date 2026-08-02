package datachangemessagehandler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v9/encoding"
	"github.com/primandproper/platform-go/v9/observability"
	platformkeys "github.com/primandproper/platform-go/v9/observability/keys"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	"github.com/primandproper/platform-go/v9/retry"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func (a *AsyncDataChangeMessageHandler) WebhookExecutionRequestsEventHandler(topicName string) func(ctx context.Context, rawMsg []byte) error {
	return func(ctx context.Context, rawMsg []byte) error {
		ctx, span := a.tracer.StartSpan(ctx)
		defer span.End()

		start := time.Now()
		status := statusSuccess

		defer func() {
			a.webhookExecutionTimestampHistogram.Record(ctx, float64(time.Since(start).Milliseconds()),
				metric.WithAttributes(attribute.String("status", status)))
			a.recordMessagesProcessed(ctx, topicWebhookExecutionRequests, status)
		}()

		var webhookExecutionRequest webhooks.WebhookExecutionRequest
		if err := a.decoder.DecodeBytes(ctx, rawMsg, &webhookExecutionRequest); err != nil {
			a.messageDecodeErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicWebhookExecutionRequests)))
			status = statusFailure
			// Unretryable: a payload that fails to decode will fail to decode on every
			// remaining attempt, and each of those is latency the healthy messages behind
			// it spend waiting. Straight to the dead-letter topic.
			return retry.Unretryable(fmt.Errorf("decoding JSON body: %w", err))
		}

		if webhookExecutionRequest.TestID != "" {
			return a.handleQueueTestMessage(ctx, a.logger.WithSpan(span), span, webhookExecutionRequest.TestID, topicName)
		}

		if err := a.handleWebhookExecutionRequest(ctx, &webhookExecutionRequest); err != nil {
			a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicWebhookExecutionRequests)))
			status = statusFailure
			return fmt.Errorf("handling webhook execution request: %w", err)
		}

		return nil
	}
}

func (a *AsyncDataChangeMessageHandler) handleWebhookExecutionRequest(
	ctx context.Context,
	webhookExecutionRequest *webhooks.WebhookExecutionRequest,
) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	if webhookExecutionRequest == nil {
		return errRequiredDataIsNil
	}

	logger := a.logger.WithValue(platformkeys.RequestIDKey, webhookExecutionRequest.RequestID)

	account, err := a.identityRepo.GetAccount(ctx, webhookExecutionRequest.AccountID)
	if err != nil {
		// A missing account is terminal: there is nothing to sign the delivery with, and no
		// amount of redelivery will bring it back. Anything else is presumed transient and
		// returned so the queue retries.
		if errors.Is(err, sql.ErrNoRows) {
			observability.AcknowledgeError(err, logger, span, "getting account")
			return nil
		}

		return observability.PrepareAndLogError(err, logger, span, "getting account")
	}

	webhook, err := a.webhookRepo.GetWebhook(ctx, webhookExecutionRequest.WebhookID, webhookExecutionRequest.AccountID)
	if err != nil {
		// Same split: an archived webhook has nothing left to deliver to, but a connection
		// blip or a timeout must not read as "this webhook is gone" and drop the delivery.
		if errors.Is(err, sql.ErrNoRows) {
			observability.AcknowledgeError(err, logger, span, "getting webhook")
			return nil
		}

		return observability.PrepareAndLogError(err, logger, span, "getting webhook")
	}

	var payloadBody []byte
	switch webhook.ContentType {
	case encoding.ContentTypeJSON.String():
		payloadBody, err = json.Marshal(webhookExecutionRequest.Payload)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "marshaling webhook payload")
		}
	case encoding.ContentTypeXML.String():
		payloadBody, err = xml.Marshal(webhookExecutionRequest.Payload)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "marshaling webhook payload")
		}
	}

	req, err := http.NewRequestWithContext(ctx, webhook.Method, webhook.URL, bytes.NewReader(payloadBody))
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "creating webhook request")
	}

	req.Header.Set("Content-Type", webhook.ContentType)

	decryptedKey, err := hex.DecodeString(account.WebhookEncryptionKey)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "decoding webhook encryption key")
	}

	digest := hmac.New(sha256.New, decryptedKey)
	digest.Write(payloadBody)
	req.Header.Set("X-Dinner-Done-Better-Signature", hex.EncodeToString(digest.Sum(nil)))

	res, err := a.webhookHTTPClient.Do(req) //nolint:gosec // G704: webhook URL is admin-configured; webhooks intentionally deliver to external URLs
	if err != nil {
		// Return the error so the message is retried at the queue level rather than silently dropped.
		return observability.PrepareAndLogError(err, logger, span, "executing webhook request")
	}
	defer func() {
		if err = res.Body.Close(); err != nil {
			logger.Error("closing response body", err)
		}
	}()

	logger = logger.WithResponse(res)
	tracing.AttachResponseToSpan(span, res)

	if res.StatusCode < 200 || res.StatusCode > 299 {
		// Log the real status code (err is nil here) and return an error so the delivery is retried.
		return observability.PrepareAndLogError(
			fmt.Errorf("webhook responded with unexpected status code %d", res.StatusCode),
			logger, span, "executing webhook request",
		)
	}

	return nil
}
