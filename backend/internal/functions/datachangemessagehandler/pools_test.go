package datachangemessagehandler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	queuescfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v9/jobs"
	"github.com/primandproper/platform-go/v9/messagequeue"
	msgqueuemock "github.com/primandproper/platform-go/v9/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v9/observability/metrics/noop"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// blockingConsumer stands in for a real messagequeue.Consumer: it blocks in Consume until the
// pool closes its stop channel, which is what lets a test assert that Close actually drains
// rather than returning while consumers are still running.
type blockingConsumer struct {
	started chan struct{}
	once    sync.Once
}

func (c *blockingConsumer) Consume(ctx context.Context, _ chan<- error) {
	c.once.Do(func() { close(c.started) })
	<-ctx.Done()
}

func buildTestPoolsHandler(t *testing.T, consumerProvider messagequeue.ConsumerProvider) *AsyncDataChangeMessageHandler {
	t.Helper()

	queues := queuescfg.Config{
		DataChangesTopicName:              "data-changes",
		OutboundEmailsTopicName:           "outbound-emails",
		SearchIndexRequestsTopicName:      "search-index-requests",
		WebhookExecutionRequestsTopicName: "webhook-execution-requests",
		UserDataAggregationTopicName:      "user-data-aggregation",
		MobileNotificationsTopicName:      "mobile-notifications",
	}

	return &AsyncDataChangeMessageHandler{
		logger:           loggingnoop.NewLogger(),
		tracer:           tracing.NewTracerForTest(t.Name()),
		tracerProvider:   tracingnoop.NewTracerProvider(),
		metricsProvider:  metricsnoop.NewMetricsProvider(),
		consumerProvider: consumerProvider,
		queuesConfig:     queues,
		poolsConfig: config.WorkerPoolsConfig{
			DeadLetterTopicName: "dead-letter",
		},
		deadLetter: func(context.Context, jobs.DeadLetter) error { return nil },
	}
}

func TestAsyncDataChangeMessageHandler_Start(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		var (
			hat    sync.Mutex
			topics []string
		)

		consumerProvider := &msgqueuemock.ConsumerProviderMock{
			NewConsumerFunc: func(_ context.Context, topic string, _ messagequeue.ConsumerFunc) (messagequeue.Consumer, error) {
				hat.Lock()
				topics = append(topics, topic)
				hat.Unlock()

				return &blockingConsumer{started: make(chan struct{})}, nil
			},
		}

		handler := buildTestPoolsHandler(t, consumerProvider)

		require.NoError(t, handler.Start(ctx))
		t.Cleanup(func() {
			closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer cancel()
			assert.NoError(t, handler.Close(closeCtx))
		})

		assert.Len(t, handler.pools, 6)

		hat.Lock()
		defer hat.Unlock()
		assert.ElementsMatch(t, []string{
			"data-changes",
			"outbound-emails",
			"search-index-requests",
			"webhook-execution-requests",
			"user-data-aggregation",
			"mobile-notifications",
		}, topics)
	})

	T.Run("with already started pools", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		consumerProvider := &msgqueuemock.ConsumerProviderMock{
			NewConsumerFunc: func(context.Context, string, messagequeue.ConsumerFunc) (messagequeue.Consumer, error) {
				return &blockingConsumer{started: make(chan struct{})}, nil
			},
		}

		handler := buildTestPoolsHandler(t, consumerProvider)

		require.NoError(t, handler.Start(ctx))
		t.Cleanup(func() {
			closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer cancel()
			assert.NoError(t, handler.Close(closeCtx))
		})

		assert.ErrorIs(t, handler.Start(ctx), errPoolsAlreadyStarted)
	})

	T.Run("with error building a consumer", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := errors.New("blah")

		var calls int

		consumerProvider := &msgqueuemock.ConsumerProviderMock{
			NewConsumerFunc: func(context.Context, string, messagequeue.ConsumerFunc) (messagequeue.Consumer, error) {
				// Fail partway through so the partial-start cleanup path is what runs.
				calls++
				if calls > 2 {
					return nil, expected
				}

				return &blockingConsumer{started: make(chan struct{})}, nil
			},
		}

		handler := buildTestPoolsHandler(t, consumerProvider)

		assert.ErrorIs(t, handler.Start(ctx), expected)
		// The pools that did start are closed rather than left consuming.
		assert.Empty(t, handler.pools)
	})
}

func TestAsyncDataChangeMessageHandler_Close(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		var (
			hat       sync.Mutex
			consumers []*blockingConsumer
		)

		consumerProvider := &msgqueuemock.ConsumerProviderMock{
			NewConsumerFunc: func(context.Context, string, messagequeue.ConsumerFunc) (messagequeue.Consumer, error) {
				c := &blockingConsumer{started: make(chan struct{})}

				hat.Lock()
				consumers = append(consumers, c)
				hat.Unlock()

				return c, nil
			},
		}

		handler := buildTestPoolsHandler(t, consumerProvider)

		require.NoError(t, handler.Start(ctx))

		// Wait for every consumer to actually be consuming, so Close is exercised against
		// running pools rather than racing their startup.
		hat.Lock()
		started := make([]*blockingConsumer, len(consumers))
		copy(started, consumers)
		hat.Unlock()

		for _, c := range started {
			select {
			case <-c.started:
			case <-time.After(5 * time.Second):
				t.Fatal("timed out waiting for consumer to start")
			}
		}

		closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		assert.NoError(t, handler.Close(closeCtx))
		assert.Empty(t, handler.pools)
	})

	T.Run("without started pools", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, buildTestPoolsHandler(t, &msgqueuemock.ConsumerProviderMock{}).Close(t.Context()))
	})
}
