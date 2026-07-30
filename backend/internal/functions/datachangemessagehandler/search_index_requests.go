package datachangemessagehandler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	coreindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/indexing"
	eatingindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	textsearch "github.com/primandproper/platform-go/v8/search/text"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func (a *AsyncDataChangeMessageHandler) SearchIndexRequestsEventHandler(topicName string) func(ctx context.Context, rawMsg []byte) error {
	return func(ctx context.Context, rawMsg []byte) error {
		ctx, span := a.tracer.StartSpan(ctx)
		defer span.End()

		start := time.Now()
		status := statusSuccess
		indexType := unknownValue

		defer func() {
			a.searchIndexRequestsExecutionTimeHistogram.Record(ctx, float64(time.Since(start).Milliseconds()),
				metric.WithAttributes(
					attribute.String("status", status),
					attribute.String("index_type", indexType),
				))
			a.recordMessagesProcessed(ctx, topicSearchIndexRequests, status)
		}()

		var searchIndexRequest textsearch.IndexRequest
		if err := json.NewDecoder(bytes.NewReader(rawMsg)).Decode(&searchIndexRequest); err != nil {
			a.messageDecodeErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicSearchIndexRequests)))
			status = statusFailure
			return fmt.Errorf("decoding JSON body: %w", err)
		}

		if searchIndexRequest.TestID != "" {
			return a.handleQueueTestMessage(ctx, a.logger.WithSpan(span), span, searchIndexRequest.TestID, topicName)
		}

		indexType = searchIndexRequest.IndexType

		switch searchIndexRequest.IndexType {
		case eatingindexing.IndexTypeRecipes,
			eatingindexing.IndexTypeMeals,
			eatingindexing.IndexTypeValidIngredients,
			eatingindexing.IndexTypeValidInstruments,
			eatingindexing.IndexTypeValidMeasurementUnits,
			eatingindexing.IndexTypeValidPreparations,
			eatingindexing.IndexTypeValidIngredientStates,
			eatingindexing.IndexTypeValidVessels:
			// we don't want to retry indexing perpetually in the event of a fundamental error, so we just log it and move on
			if err := a.mealPlanningDataIndexer.HandleIndexRequest(ctx, &searchIndexRequest); err != nil {
				a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicSearchIndexRequests)))
				status = statusFailure
				return fmt.Errorf("handling search indexing request: %w", err)
			}

		case coreindexing.IndexTypeUsers:
			// we don't want to retry indexing perpetually in the event of a fundamental error, so we just log it and move on
			if err := a.userDataIndexer.HandleIndexRequest(ctx, &searchIndexRequest); err != nil {
				a.handlerErrorsCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("topic", topicSearchIndexRequests)))
				status = statusFailure
				return fmt.Errorf("handling search indexing request: %w", err)
			}
		}

		return nil
	}
}
