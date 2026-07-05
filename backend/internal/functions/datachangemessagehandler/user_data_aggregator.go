package datachangemessagehandler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	dataprivacykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy/keys"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"

	"github.com/primandproper/platform-go/v3/observability"
	"github.com/primandproper/platform-go/v3/observability/tracing"
	"github.com/primandproper/platform-go/v3/uploads"

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
			return fmt.Errorf("decoding JSON body: %w", err)
		}

		if userDataCollectionRequest.TestID != "" {
			return a.handleQueueTestMessage(ctx, a.logger.WithSpan(span), span, userDataCollectionRequest.TestID, topicName)
		}

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
			return observability.PrepareAndLogError(err, logger, span, "fetching user data collection")
		}

		logger.Info("compiled user data payload, marshaling")

		// Marshal the collection to JSON
		collectionBytes, err := json.Marshal(collection)
		if err != nil {
			a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicUserDataAggregation)))
			status = statusFailure
			return observability.PrepareAndLogError(err, logger, span, "marshaling collection")
		}

		logger.Info("saving file to object storage")

		// Save to object storage with report ID as filename
		if err = uploads.SaveFile(ctx, a.uploadManager, fmt.Sprintf("%s.json", userDataCollectionRequest.ReportID), collectionBytes); err != nil {
			a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicUserDataAggregation)))
			status = statusFailure
			return observability.PrepareAndLogError(err, logger, span, "saving collection")
		}

		logger.Info("user data aggregation complete")

		return nil
	}
}
