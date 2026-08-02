package datachangemessagehandler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	dataprivacykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy/keys"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"

	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	"github.com/primandproper/platform-go/v9/retry"
	"github.com/primandproper/platform-go/v9/uploads"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// UserDataAggregationEventHandler handles user data aggregation requests for GDPR/CCPA compliance.
func (a *AsyncDataChangeMessageHandler) UserDataAggregationEventHandler(topicName string) func(ctx context.Context, rawMsg []byte) error {
	return func(ctx context.Context, rawMsg []byte) error {
		ctx, span := a.tracer.StartSpan(ctx)
		defer span.End()

		start := time.Now()
		status := statusSuccess

		defer func() {
			a.userDataAggregationExecutionTimeHistogram.Record(ctx, float64(time.Since(start).Milliseconds()),
				metric.WithAttributes(attribute.String("status", status)))
			a.recordMessagesProcessed(ctx, topicUserDataAggregation, status)
		}()

		var userDataCollectionRequest dataprivacy.UserDataAggregationRequest
		if err := a.decoder.DecodeBytes(ctx, rawMsg, &userDataCollectionRequest); err != nil {
			a.messageDecodeErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicUserDataAggregation)))
			status = statusFailure
			// Unretryable: a payload that fails to decode will fail to decode on every
			// remaining attempt, and each of those is latency the healthy messages behind
			// it spend waiting. Straight to the dead-letter topic.
			return retry.Unretryable(fmt.Errorf("decoding JSON body: %w", err))
		}

		if userDataCollectionRequest.TestID != "" {
			return a.handleQueueTestMessage(ctx, a.logger.WithSpan(span), span, userDataCollectionRequest.TestID, topicName)
		}

		// RequestID carries the disclosure record's ID so its status can be updated as the work progresses.
		disclosureID := userDataCollectionRequest.RequestID

		logger := a.logger.WithValue(dataprivacykeys.UserDataAggregationReportIDKey, userDataCollectionRequest.ReportID).
			WithValue(identitykeys.UserIDKey, userDataCollectionRequest.UserID)
		tracing.AttachToSpan(span, dataprivacykeys.UserDataAggregationReportIDKey, userDataCollectionRequest.ReportID)
		tracing.AttachToSpan(span, identitykeys.UserIDKey, userDataCollectionRequest.UserID)
		logger.Info("loaded payload, aggregating user data")

		// Fetch the user's complete data collection
		collection, err := a.dataPrivacyRepo.FetchUserDataCollection(ctx, userDataCollectionRequest.UserID)
		if err != nil {
			a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicUserDataAggregation)))
			status = statusFailure
			a.markDisclosureFailed(ctx, logger, span, disclosureID)
			return observability.PrepareAndLogError(err, logger, span, "fetching user data collection")
		}

		logger.Info("compiled user data payload, marshaling")

		// Marshal the collection to JSON
		collectionBytes, err := json.Marshal(collection)
		if err != nil {
			a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicUserDataAggregation)))
			status = statusFailure
			a.markDisclosureFailed(ctx, logger, span, disclosureID)
			return observability.PrepareAndLogError(err, logger, span, "marshaling collection")
		}

		logger.Info("saving file to object storage")

		// Save to object storage with report ID as filename
		if err = uploads.SaveFile(ctx, a.uploadManager, fmt.Sprintf("%s.json", userDataCollectionRequest.ReportID), collectionBytes); err != nil {
			a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicUserDataAggregation)))
			status = statusFailure
			a.markDisclosureFailed(ctx, logger, span, disclosureID)
			return observability.PrepareAndLogError(err, logger, span, "saving collection")
		}

		// Mark the disclosure completed so the requesting user can fetch the report.
		if disclosureID != "" {
			if err = a.dataPrivacyRepo.MarkUserDataDisclosureCompleted(ctx, disclosureID, userDataCollectionRequest.ReportID); err != nil {
				a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicUserDataAggregation)))
				status = statusFailure
				return observability.PrepareAndLogError(err, logger, span, "marking disclosure completed")
			}
		}

		logger.Info("user data aggregation complete")

		return nil
	}
}

// markDisclosureFailed best-effort flips a disclosure to the failed status. A missing disclosure ID (e.g. a legacy
// message) is a no-op, and a failure to record the status is logged rather than returned so the original error is preserved.
func (a *AsyncDataChangeMessageHandler) markDisclosureFailed(ctx context.Context, logger logging.Logger, span tracing.Span, disclosureID string) {
	if disclosureID == "" {
		return
	}

	if err := a.dataPrivacyRepo.MarkUserDataDisclosureFailed(ctx, disclosureID); err != nil {
		observability.AcknowledgeError(err, logger, span, "marking disclosure failed")
	}
}
