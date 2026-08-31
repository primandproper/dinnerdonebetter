package datachangemessagehandler

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	identitymock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/mock"
	internalopsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops/mock"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	notificationsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/push"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	analyticsmock "github.com/primandproper/platform-go/v13/analytics/mock"
	emailmock "github.com/primandproper/platform-go/v13/email/mock"
	encodingmock "github.com/primandproper/platform-go/v13/encoding/mock"
	"github.com/primandproper/platform-go/v13/messagequeue"
	msgqueuemock "github.com/primandproper/platform-go/v13/messagequeue/mock"
	noopnotifications "github.com/primandproper/platform-go/v13/notifications/mobile/noop"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	mockmetrics "github.com/primandproper/platform-go/v13/observability/metrics/mock"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
)

//nolint:gocritic // I know this returns too many things
func buildTestAsyncDataChangeMessageHandler(t *testing.T) (*AsyncDataChangeMessageHandler, *identitymock.RepositoryMock, *msgqueuemock.ConsumerProviderMock, *msgqueuemock.PublisherProviderMock, *analyticsmock.EventReporterMock, *emailmock.EmailerMock, *mockmetrics.ProviderMock, *encodingmock.ServerEncoderDecoderMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())

	identityRepo := &identitymock.RepositoryMock{}
	consumerProvider := &msgqueuemock.ConsumerProviderMock{}
	publisherProvider := &msgqueuemock.PublisherProviderMock{}
	analyticsEventReporter := &analyticsmock.EventReporterMock{}
	emailer := &emailmock.EmailerMock{}
	metricsProvider := &mockmetrics.ProviderMock{}
	decoder := &encodingmock.ServerEncoderDecoderMock{}
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
	mealPlanRepo := &mealplanningmock.RepositoryMock{}

	pushFanout, err := push.NewFanout(logger, &notificationsmock.RepositoryMock{}, noopnotifications.NewPushNotificationSender(), noopProvider)
	require.NoError(t, err)

	handler := &AsyncDataChangeMessageHandler{
		identityRepo:                         identityRepo,
		internalOpsRepo:                      internalOpsRepo,
		consumerProvider:                     consumerProvider,
		analyticsEventReporter:               analyticsEventReporter,
		emailer:                              emailer,
		decoder:                              decoder,
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
		queuesConfig: queuescfg.Config{
			SearchIndexRequestsTopicName: "search-index-requests",
		},
		outboundEmailsPublisher:      mockPublisher,
		mobileNotificationsPublisher: mockPublisher,
		mealPlanRepo:                 mealPlanRepo,
		pushFanout:                   pushFanout,
	}

	handler.outboundNotificationHandlers = []OutboundNotificationHandler{
		handler.handleMealPlanningOutboundNotification,
		handler.handleIdentityOutboundNotification,
	}

	return handler, identityRepo, consumerProvider, publisherProvider, analyticsEventReporter, emailer, metricsProvider, decoder
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
				SearchIndexRequestsTopicName: "search-index-requests",
				MobileNotificationsTopicName: "mobile-notifications",
			},
			Pools: config.WorkerPoolsConfig{
				DeadLetterTopicName: "dead-letter",
			},
		}
		identityRepo := &identitymock.RepositoryMock{}
		consumerProvider := &msgqueuemock.ConsumerProviderMock{}
		publisherProvider := &msgqueuemock.PublisherProviderMock{}
		analyticsEventReporter := &analyticsmock.EventReporterMock{}
		emailer := &emailmock.EmailerMock{}
		metricsProvider := &mockmetrics.ProviderMock{}
		decoder := &encodingmock.ServerEncoderDecoderMock{}
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
		mealPlanRepo := &mealplanningmock.RepositoryMock{}

		pushFanout, err := push.NewFanout(logger, &notificationsmock.RepositoryMock{}, noopnotifications.NewPushNotificationSender(), noopProvider)
		require.NoError(t, err)

		handler, err := NewAsyncDataChangeMessageHandler(
			ctx,
			logger,
			tracerProvider,
			cfg,
			identityRepo,
			internalOpsRepo,
			consumerProvider,
			publisherProvider,
			analyticsEventReporter,
			emailer,
			metricsProvider,
			decoder,
			searchSyncers,
			mealPlanRepo,
			pushFanout,
		)

		require.NoError(t, err)
		assert.NotNil(t, handler)
		assert.Equal(t, identityRepo, handler.identityRepo)
		assert.Equal(t, consumerProvider, handler.consumerProvider)
		assert.Equal(t, analyticsEventReporter, handler.analyticsEventReporter)
		assert.Equal(t, emailer, handler.emailer)
		assert.Equal(t, decoder, handler.decoder)
		assert.Equal(t, searchSyncers, handler.searchSyncers)

		// metricsProvider and publisherProvider are moq mocks - no testify assertion needed
	})
}
