package datachangemessagehandler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v11/jobs"
	"github.com/primandproper/platform-go/v11/messagequeue"
	msgqueuemock "github.com/primandproper/platform-go/v11/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v11/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v11/observability/metrics/noop"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v11/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// blockingConsumer stands in for a real messagequeue.Consumer: it blocks in Consume until the
// pool closes its stop channel, which is what lets a test assert that a drain actually happened
// rather than that a slice was emptied.
type blockingConsumer struct {
	started chan struct{}
	stopped chan struct{}
	once    sync.Once
}

func newBlockingConsumer() *blockingConsumer {
	return &blockingConsumer{
		started: make(chan struct{}),
		stopped: make(chan struct{}),
	}
}

func (c *blockingConsumer) Consume(ctx context.Context, _ chan<- error) {
	c.once.Do(func() { close(c.started) })
	<-ctx.Done()
	close(c.stopped)
}

func buildTestPoolsHandler(t *testing.T, consumerProvider messagequeue.ConsumerProvider) *AsyncDataChangeMessageHandler {
	t.Helper()

	queues := queuescfg.Config{
		DataChangesTopicName:         "data-changes",
		OutboundEmailsTopicName:      "outbound-emails",
		SearchIndexRequestsTopicName: "search-index-requests",
		MobileNotificationsTopicName: "mobile-notifications",
	}

	handler := &AsyncDataChangeMessageHandler{
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
		// One search index, not nine: the point of the assertions below is that a pool is
		// built per registered Syncer, and one is enough to show the fan-out happens.
		searchSyncers: []SearchSyncer{
			{
				Topic:  "search-users",
				Handle: func(context.Context, []byte) error { return nil },
			},
		},
	}

	group, err := newPoolGroup(t.Context(), handler)
	require.NoError(t, err)
	handler.poolGroup = group

	return handler
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

				return newBlockingConsumer(), nil
			},
		}

		handler := buildTestPoolsHandler(t, consumerProvider)

		require.NoError(t, handler.Start(ctx))
		t.Cleanup(func() {
			closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer cancel()
			assert.NoError(t, handler.Close(closeCtx))
		})

		expectedTopics := []string{
			"data-changes",
			"outbound-emails",
			"search-users",
			"mobile-notifications",
		}

		// What the group says it drains, and what it actually subscribed to, are separate
		// claims — the first is resolved at construction and the second is the broker's record.
		assert.ElementsMatch(t, expectedTopics, handler.poolGroup.Topics())

		hat.Lock()
		defer hat.Unlock()
		assert.ElementsMatch(t, expectedTopics, topics)
	})

	T.Run("with already started pools", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		consumerProvider := &msgqueuemock.ConsumerProviderMock{
			NewConsumerFunc: func(context.Context, string, messagequeue.ConsumerFunc) (messagequeue.Consumer, error) {
				return newBlockingConsumer(), nil
			},
		}

		handler := buildTestPoolsHandler(t, consumerProvider)

		require.NoError(t, handler.Start(ctx))
		t.Cleanup(func() {
			closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer cancel()
			assert.NoError(t, handler.Close(closeCtx))
		})

		assert.ErrorIs(t, handler.Start(ctx), jobs.ErrPoolGroupStarted)
	})

	T.Run("with error building a consumer", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := errors.New("blah")

		var (
			hat       sync.Mutex
			consumers []*blockingConsumer
			calls     int
		)

		consumerProvider := &msgqueuemock.ConsumerProviderMock{
			NewConsumerFunc: func(context.Context, string, messagequeue.ConsumerFunc) (messagequeue.Consumer, error) {
				hat.Lock()
				defer hat.Unlock()

				// Fail partway through so the partial-start cleanup path is what runs.
				calls++
				if calls > 2 {
					return nil, expected
				}

				c := newBlockingConsumer()
				consumers = append(consumers, c)

				return c, nil
			},
		}

		handler := buildTestPoolsHandler(t, consumerProvider)

		require.ErrorIs(t, handler.Start(ctx), expected)

		// The pools that did start are drained rather than left consuming.
		hat.Lock()
		started := make([]*blockingConsumer, len(consumers))
		copy(started, consumers)
		hat.Unlock()

		require.Len(t, started, 2)
		for _, c := range started {
			select {
			case <-c.stopped:
			case <-time.After(5 * time.Second):
				t.Fatal("timed out waiting for a partially started consumer to stop")
			}
		}

		// A group is single-use, so a failed start is final rather than something to retry.
		assert.ErrorIs(t, handler.Start(ctx), jobs.ErrPoolGroupStarted)
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
				c := newBlockingConsumer()

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

		require.NoError(t, handler.Close(closeCtx))

		for _, c := range started {
			select {
			case <-c.stopped:
			case <-time.After(5 * time.Second):
				t.Fatal("timed out waiting for consumer to stop")
			}
		}
	})

	T.Run("with a second close", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		consumerProvider := &msgqueuemock.ConsumerProviderMock{
			NewConsumerFunc: func(context.Context, string, messagequeue.ConsumerFunc) (messagequeue.Consumer, error) {
				return newBlockingConsumer(), nil
			},
		}

		handler := buildTestPoolsHandler(t, consumerProvider)

		require.NoError(t, handler.Start(ctx))

		closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		require.NoError(t, handler.Close(closeCtx))
		assert.NoError(t, handler.Close(closeCtx))
	})

	T.Run("without started pools", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, buildTestPoolsHandler(t, &msgqueuemock.ConsumerProviderMock{}).Close(t.Context()))
	})

	T.Run("closes each syncer's stamper", func(t *testing.T) {
		t.Parallel()

		handler := buildTestPoolsHandler(t, &msgqueuemock.ConsumerProviderMock{})

		var closed int
		handler.searchSyncers[0].Close = func(context.Context) error {
			closed++

			return nil
		}

		require.NoError(t, handler.Close(t.Context()))

		// The stamper holds the last interval's last_indexed_at writes in memory, so a
		// shutdown that does not close it drops them.
		assert.Equal(t, 1, closed)
	})

	T.Run("with a stamper that will not close", func(t *testing.T) {
		t.Parallel()

		handler := buildTestPoolsHandler(t, &msgqueuemock.ConsumerProviderMock{})

		expected := errors.New("blah")
		handler.searchSyncers[0].Close = func(context.Context) error { return expected }

		assert.ErrorIs(t, handler.Close(t.Context()), expected)
	})
}

func TestNewPoolGroup(T *testing.T) {
	T.Parallel()

	T.Run("with two syncers on one topic", func(t *testing.T) {
		t.Parallel()

		handler := buildTestPoolsHandler(t, &msgqueuemock.ConsumerProviderMock{})
		handler.searchSyncers = append(handler.searchSyncers, SearchSyncer{
			Topic:  handler.searchSyncers[0].Topic,
			Handle: func(context.Context, []byte) error { return nil },
		})

		// Caught at construction rather than from the middle of a start that has already
		// brought other pools up.
		_, err := newPoolGroup(t.Context(), handler)
		assert.ErrorIs(t, err, jobs.ErrDuplicateTopic)
	})
}
