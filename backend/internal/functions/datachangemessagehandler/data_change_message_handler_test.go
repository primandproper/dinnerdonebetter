package datachangemessagehandler

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	internalopsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops/mock"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	identitymock "github.com/primandproper/platform-go/v15/identity/mock"
	platformnotificationsmock "github.com/primandproper/platform-go/v15/notifications/mock"
	"github.com/primandproper/platform-go/v15/notifications/push"
	"github.com/primandproper/platform-go/v15/webhooks"
	analyticsmock "github.com/primandproper/primitives-go/v2/analytics/mock"
	"github.com/primandproper/primitives-go/v2/database"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	emailmock "github.com/primandproper/primitives-go/v2/email/mock"
	"github.com/primandproper/primitives-go/v2/messagequeue"
	msgqueuemock "github.com/primandproper/primitives-go/v2/messagequeue/mock"
	noopnotifications "github.com/primandproper/primitives-go/v2/notifications/mobile/noop"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	mockmetrics "github.com/primandproper/primitives-go/v2/observability/metrics/mock"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

//nolint:gocritic // I know this returns too many things
func buildTestAsyncDataChangeMessageHandler(t *testing.T) (*AsyncDataChangeMessageHandler, *identitymock.StoreMock, *msgqueuemock.ConsumerProviderMock, *msgqueuemock.PublisherProviderMock, *analyticsmock.EventReporterMock, *emailmock.EmailerMock, *mockmetrics.ProviderMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())

	identityRepo := &identitymock.StoreMock{}
	consumerProvider := &msgqueuemock.ConsumerProviderMock{}
	publisherProvider := &msgqueuemock.PublisherProviderMock{}
	analyticsEventReporter := &analyticsmock.EventReporterMock{}
	emailer := &emailmock.EmailerMock{}
	metricsProvider := &mockmetrics.ProviderMock{}
	// Create mock indexers with noop implementations for testing
	searchSyncers := []SearchSyncer{}

	// Set up mock publishers for the indexers to prevent nil pointer dereferences
	mockPublisher := &msgqueuemock.PublisherMock{
		PublishFunc:      func(_ context.Context, _ any, _ ...messagequeue.PublishOption) error { return nil },
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
		StopFunc:         func() {},
	}
	publisherProvider.NewPublisherFunc = func(_ context.Context, _ string) (messagequeue.Publisher, error) {
		return mockPublisher, nil
	}

	// Set up mock histograms and counters
	noopProvider := metricsnoop.NewMetricsProvider()
	noopHistogram, _ := noopProvider.NewFloat64Histogram("test")
	noopCounter, _ := noopProvider.NewInt64Counter("test")
	metricsProvider.NewFloat64HistogramFunc = func(_ string, _ ...metric.Float64HistogramOption) (metrics.Float64Histogram, error) {
		return noopHistogram, nil
	}
	metricsProvider.NewInt64CounterFunc = func(_ string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
		return noopCounter, nil
	}

	internalOpsRepo := &internalopsmock.InternalOpsDataManagerMock{}

	pushFanout, err := push.NewFanout(&platformnotificationsmock.RegistryMock{},
		noopnotifications.NewPushNotificationSender(),
		push.WithLogger(logger), push.WithMetricsProvider(noopProvider))
	require.NoError(t, err)

	handler := &AsyncDataChangeMessageHandler{
		directory: identityRepo,
		db: &mockdatabase.ClientMock{
			ReaderFunc: func() database.SQLQueryExecutor { return nil },
			WriterFunc: func() database.SQLQueryExecutor { return nil },
		},
		internalOpsRepo:                      internalOpsRepo,
		consumerProvider:                     consumerProvider,
		analyticsEventReporter:               analyticsEventReporter,
		emailer:                              emailer,
		searchSyncers:                        searchSyncers,
		logger:                               logger,
		tracer:                               tracer,
		dataChangesExecutionTimeHistogram:    noopHistogram,
		outboundEmailsExecutionTimeHistogram: noopHistogram,
		mobileNotificationsExecutionTimeHistogram: noopHistogram,
		messagesProcessedCounter:                  noopCounter,
		messageDecodeErrorsCounter:                noopCounter,
		handlerErrorsCounter:                      noopCounter,
		emailsSentCounter:                         noopCounter,
		emailsFailedCounter:                       noopCounter,
		queuesConfig:                              queuescfg.Config{},
		outboundEmailsPublisher:                   mockPublisher,
		mobileNotificationsPublisher:              mockPublisher,
		pushFanout:                                pushFanout,
	}

	// The handler over platform's identity events alone: the domains' handlers are theirs to
	// test, and the list the constructor builds is asserted in TestNewAsyncDataChangeMessageHandler.
	handler.outboundNotificationHandlers = []OutboundNotificationHandler{
		handler.handleIdentityOutboundNotification,
	}

	return handler, identityRepo, consumerProvider, publisherProvider, analyticsEventReporter, emailer, metricsProvider
}

func TestNewAsyncDataChangeMessageHandler(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		cfg := &config.AsyncMessageHandlerConfig{
			Queues: queuescfg.Config{
				DataChangesTopicName:         "data-changes",
				OutboundEmailsTopicName:      "outbound-emails",
				MobileNotificationsTopicName: "mobile-notifications",
				QueuedMailTopicName:          "queued-mail",
			},
			Pools: config.WorkerPoolsConfig{
				DeadLetterTopicName: "dead-letter",
			},
		}
		identityRepo := &identitymock.StoreMock{}
		consumerProvider := &msgqueuemock.ConsumerProviderMock{}
		publisherProvider := &msgqueuemock.PublisherProviderMock{}
		analyticsEventReporter := &analyticsmock.EventReporterMock{}
		emailer := &emailmock.EmailerMock{}
		metricsProvider := &mockmetrics.ProviderMock{}
		// Empty rather than populated: this asserts the handler carries what it was given,
		// and a Syncer needs a live index to construct. What each Syncer does with an event
		// is covered where the Source is, in platform-go's search/sync/source.
		searchSyncers := []SearchSyncer{}

		// Set up metrics expectations
		noopProvider := metricsnoop.NewMetricsProvider()
		noopHistogram, _ := noopProvider.NewFloat64Histogram("test")
		noopCounter, _ := noopProvider.NewInt64Counter("test")
		metricsProvider.NewFloat64HistogramFunc = func(_ string, _ ...metric.Float64HistogramOption) (metrics.Float64Histogram, error) {
			return noopHistogram, nil
		}
		metricsProvider.NewInt64CounterFunc = func(_ string, _ ...metric.Int64CounterOption) (metrics.Int64Counter, error) {
			return noopCounter, nil
		}

		// Set up publisher expectations
		mockPublisher := &msgqueuemock.PublisherMock{
			PublishFunc:      func(_ context.Context, _ any, _ ...messagequeue.PublishOption) error { return nil },
			PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
			StopFunc:         func() {},
		}
		publisherProvider.NewPublisherFunc = func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return mockPublisher, nil
		}

		internalOpsRepo := &internalopsmock.InternalOpsDataManagerMock{}

		pushFanout, err := push.NewFanout(&platformnotificationsmock.RegistryMock{},
			noopnotifications.NewPushNotificationSender(),
			push.WithLogger(logger), push.WithMetricsProvider(noopProvider))
		require.NoError(t, err)

		// One domain handler, claiming nothing, to assert the list's shape: the domains'
		// first, in the order given, and the identity handler after them.
		domainHandler := func(context.Context, *webhooks.Envelope) (bool, string, []*queuemessages.OutboundEmailMessage, error) {
			return false, "", nil, nil
		}

		handler, err := NewAsyncDataChangeMessageHandler(
			ctx,
			logger,
			tracerProvider,
			cfg,
			identityRepo,
			&mockdatabase.ClientMock{
				ReaderFunc: func() database.SQLQueryExecutor { return nil },
				WriterFunc: func() database.SQLQueryExecutor { return nil },
			},
			internalOpsRepo,
			consumerProvider,
			publisherProvider,
			analyticsEventReporter,
			emailer,
			metricsProvider,
			searchSyncers,
			[]OutboundNotificationHandler{domainHandler},
			pushFanout,
		)

		require.NoError(t, err)
		assert.NotNil(t, handler)
		assert.Equal(t, identityRepo, handler.directory)
		assert.Equal(t, consumerProvider, handler.consumerProvider)
		assert.Equal(t, analyticsEventReporter, handler.analyticsEventReporter)
		assert.Equal(t, emailer, handler.emailer)
		assert.Equal(t, searchSyncers, handler.searchSyncers)

		// The domain's handler, then the identity one: two entries, and the second is the
		// handler's own. Functions compare by nothing, so the shape is what is asserted.
		require.Len(t, handler.outboundNotificationHandlers, 2)
		handled, _, _, err := handler.outboundNotificationHandlers[0](ctx, &webhooks.Envelope{EventType: "anything"})
		require.NoError(t, err)
		assert.False(t, handled)

		// metricsProvider and publisherProvider are moq mocks - no testify assertion needed
	})
}
