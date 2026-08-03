package queuetest

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v9/identifiers"
	"github.com/primandproper/platform-go/v9/messagequeue"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	serviceName              = "queue_test"
	pollInterval             = 500 * time.Millisecond
	timeout                  = 30 * time.Second
	queueTestRoundTripMetric = "queue_test_round_trip_ms"
	metricStatusSuccess      = "success"
	metricStatusTimeout      = "timeout"
)

// JobParams holds configuration for the queue test job.
type JobParams struct {
	Queues queuescfg.Config
}

// Job sends a test message to a random queue and reports round-trip latency.
type Job struct {
	internalOpsRepo    internalops.InternalOpsDataManager
	publisherProvider  messagequeue.PublisherProvider
	logger             logging.Logger
	tracer             tracing.Tracer
	roundTripHistogram metrics.Float64Histogram
	queues             queuescfg.Config
}

// NewJob creates a new queue test job.
func NewJob(
	internalOpsRepo internalops.InternalOpsDataManager,
	publisherProvider messagequeue.PublisherProvider,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	params *JobParams,
) (*Job, error) {
	histogram, err := metricsProvider.NewFloat64Histogram(queueTestRoundTripMetric)
	if err != nil {
		return nil, fmt.Errorf("creating queue test round trip histogram: %w", err)
	}
	return &Job{
		internalOpsRepo:    internalOpsRepo,
		publisherProvider:  publisherProvider,
		logger:             logging.NewNamedLogger(logger, serviceName),
		tracer:             tracing.NewNamedTracer(tracerProvider, serviceName),
		roundTripHistogram: histogram,
		queues:             params.Queues,
	}, nil
}

// topicNames returns the topic names from config for random selection.
//
// Webhook execution and user data aggregation are not among them any more: a delivery and an
// export are both rows a worker claims rather than messages on a topic, so there is no queue
// here to probe.
func (j *Job) topicNames() []string {
	return []string{
		j.queues.DataChangesTopicName,
		j.queues.OutboundEmailsTopicName,
		j.queues.SearchIndexRequestsTopicName,
		j.queues.MobileNotificationsTopicName,
	}
}

// Do sends a test message to a random topic and reports the round-trip time as a metric.
func (j *Job) Do(ctx context.Context) error {
	ctx, span := j.tracer.StartSpan(ctx)
	defer span.End()

	topics := j.topicNames()
	topicName := topics[rand.IntN(len(topics))] //nolint:gosec // G404: test data selection does not require cryptographic randomness

	testID := identifiers.New()
	start := time.Now()

	if err := j.internalOpsRepo.CreateQueueTestMessage(ctx, testID, topicName); err != nil {
		return fmt.Errorf("creating queue test message record: %w", err)
	}

	// BuildQueueTestMessage expects logical names (e.g. "data_changes"); extract from full path if needed.
	logicalName := topicName
	if idx := strings.LastIndex(topicName, "/"); idx >= 0 {
		logicalName = topicName[idx+1:]
	}

	msg, err := internalops.BuildQueueTestMessage(logicalName, testID, "")
	if err != nil {
		return fmt.Errorf("building queue test message: %w", err)
	}

	publisher, err := j.publisherProvider.NewPublisher(ctx, topicName)
	if err != nil {
		return fmt.Errorf("initializing publisher for %s: %w", topicName, err)
	}

	if err = publisher.Publish(ctx, msg); err != nil {
		return fmt.Errorf("publishing test message: %w", err)
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	deadline := time.After(timeout)

	for {
		select {
		case <-deadline:
			j.roundTripHistogram.Record(ctx, float64(time.Since(start).Milliseconds()),
				metric.WithAttributes(
					attribute.String("topic", topicName),
					attribute.String("status", metricStatusTimeout),
				))
			return fmt.Errorf("timed out waiting for queue test message acknowledgment on %s", topicName)
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			record, pollErr := j.internalOpsRepo.GetQueueTestMessage(ctx, testID)
			if pollErr != nil {
				j.logger.Error("polling for queue test message acknowledgment", pollErr)
				continue
			}
			if record.AcknowledgedAt != nil {
				roundTripMs := time.Since(start).Milliseconds()
				j.roundTripHistogram.Record(ctx, float64(roundTripMs),
					metric.WithAttributes(
						attribute.String("topic", topicName),
						attribute.String("status", metricStatusSuccess),
					))
				j.logger.WithValue("topic", topicName).WithValue("roundTripMs", roundTripMs).Info("queue test completed")
				return nil
			}
		}
	}
}
