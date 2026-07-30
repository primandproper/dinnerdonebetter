package datachangemessagehandler

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	dataprivacymock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy/mock"
	identitymock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/mock"
	internalopsmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/internalops/mock"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	notificationsmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/mock"
	webhooksmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/mock"
	identityindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/indexing"
	mealplanningindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	analyticsmock "github.com/primandproper/platform-go/v8/analytics/mock"
	emailmock "github.com/primandproper/platform-go/v8/email/mock"
	encodingmock "github.com/primandproper/platform-go/v8/encoding/mock"
	"github.com/primandproper/platform-go/v8/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v8/messagequeue/config"
	msgqueuemock "github.com/primandproper/platform-go/v8/messagequeue/mock"
	noopnotifications "github.com/primandproper/platform-go/v8/notifications/mobile/noop"
	loggingnoop "github.com/primandproper/platform-go/v8/observability/logging/noop"
	"github.com/primandproper/platform-go/v8/observability/metrics"
	mockmetrics "github.com/primandproper/platform-go/v8/observability/metrics/mock"
	metricsnoop "github.com/primandproper/platform-go/v8/observability/metrics/noop"
	"github.com/primandproper/platform-go/v8/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v8/observability/tracing/noop"
	uploadsmock "github.com/primandproper/platform-go/v8/uploads/mock"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/metric"
)

// noopPasswordResetTokenDataManager implements auth.PasswordResetTokenDataManager for tests that do not exercise the password reset flow.
type noopPasswordResetTokenDataManager struct{}

func (noopPasswordResetTokenDataManager) GetPasswordResetTokenByID(context.Context, string) (*auth.PasswordResetToken, error) {
	return nil, nil
}
func (noopPasswordResetTokenDataManager) GetPasswordResetTokenByToken(context.Context, string) (*auth.PasswordResetToken, error) {
	return nil, nil
}
func (noopPasswordResetTokenDataManager) CreatePasswordResetToken(context.Context, *auth.PasswordResetTokenDatabaseCreationInput) (*auth.PasswordResetToken, error) {
	return nil, nil
}
func (noopPasswordResetTokenDataManager) RedeemPasswordResetToken(context.Context, string) error {
	return nil
}

//nolint:gocritic // I know this returns too many things
func buildTestAsyncDataChangeMessageHandler(t *testing.T) (*AsyncDataChangeMessageHandler, *identitymock.RepositoryMock, *webhooksmock.RepositoryMock, *msgqueuemock.ConsumerProviderMock, *msgqueuemock.PublisherProviderMock, *analyticsmock.EventReporterMock, *emailmock.EmailerMock, *uploadsmock.UploadManagerMock, *mockmetrics.ProviderMock, *encodingmock.ServerEncoderDecoderMock, *dataprivacymock.RepositoryMock) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())

	identityRepo := &identitymock.RepositoryMock{}
	webhookRepo := &webhooksmock.RepositoryMock{}
	consumerProvider := &msgqueuemock.ConsumerProviderMock{}
	publisherProvider := &msgqueuemock.PublisherProviderMock{}
	analyticsEventReporter := &analyticsmock.EventReporterMock{}
	emailer := &emailmock.EmailerMock{}
	uploadManager := &uploadsmock.UploadManagerMock{}
	metricsProvider := &mockmetrics.ProviderMock{}
	decoder := &encodingmock.ServerEncoderDecoderMock{}
	dataPrivacyRepo := &dataprivacymock.RepositoryMock{}

	// Create mock indexers with noop implementations for testing
	userDataIndexer := &identityindexing.UserDataIndexer{}
	mealPlanningDataIndexer := &mealplanningindexing.MealPlanningDataIndexer{}

	// Set up mock publishers for the indexers to prevent nil pointer dereferences
	mockPublisher := &msgqueuemock.PublisherMock{
		PublishFunc:      func(_ context.Context, _ any) error { return nil },
		PublishAsyncFunc: func(_ context.Context, _ any) {},
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
	notificationsRepo := &notificationsmock.RepositoryMock{}
	pushNotificationSender := noopnotifications.NewPushNotificationSender()

	handler := &AsyncDataChangeMessageHandler{
		identityRepo:                         identityRepo,
		webhookRepo:                          webhookRepo,
		internalOpsRepo:                      internalOpsRepo,
		consumerProvider:                     consumerProvider,
		analyticsEventReporter:               analyticsEventReporter,
		emailer:                              emailer,
		uploadManager:                        uploadManager,
		decoder:                              decoder,
		userDataIndexer:                      userDataIndexer,
		mealPlanningDataIndexer:              mealPlanningDataIndexer,
		logger:                               logger,
		tracer:                               tracer,
		nonWebhookEventTypes:                 []string{},
		dataChangesExecutionTimeHistogram:    noopHistogram,
		outboundEmailsExecutionTimeHistogram: noopHistogram,
		webhookExecutionTimestampHistogram:   noopHistogram,
		userDataAggregationExecutionTimeHistogram: noopHistogram,
		searchIndexRequestsExecutionTimeHistogram: noopHistogram,
		mobileNotificationsExecutionTimeHistogram: noopHistogram,
		messagesProcessedCounter:                  noopCounter,
		messageDecodeErrorsCounter:                noopCounter,
		handlerErrorsCounter:                      noopCounter,
		emailsSentCounter:                         noopCounter,
		emailsFailedCounter:                       noopCounter,
		pushNotificationsSentCounter:              noopCounter,
		badDeviceTokensArchivedCounter:            noopCounter,
		queuesConfig: msgconfig.QueuesConfig{
			SearchIndexRequestsTopicName: "search-index-requests",
		},
		searchDataIndexPublisher:         mockPublisher,
		outboundEmailsPublisher:          mockPublisher,
		webhookExecutionRequestPublisher: mockPublisher,
		mobileNotificationsPublisher:     mockPublisher,
		dataPrivacyRepo:                  dataPrivacyRepo,
		mealPlanRepo:                     mealPlanRepo,
		passwordResetTokenDataManager:    noopPasswordResetTokenDataManager{},
		notificationsRepo:                notificationsRepo,
		pushNotificationSender:           pushNotificationSender,
	}

	handler.searchIndexHandlers = []SearchIndexEventHandler{
		handler.handleMealPlanningSearchIndexUpdate,
		handler.handleIdentitySearchIndexUpdate,
	}
	handler.outboundNotificationHandlers = []OutboundNotificationHandler{
		handler.handleMealPlanningOutboundNotification,
		handler.handleIdentityOutboundNotification,
	}

	return handler, identityRepo, webhookRepo, consumerProvider, publisherProvider, analyticsEventReporter, emailer, uploadManager, metricsProvider, decoder, dataPrivacyRepo
}

func TestNewAsyncDataChangeMessageHandler(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		cfg := &config.AsyncMessageHandlerConfig{
			Queues: msgconfig.QueuesConfig{
				OutboundEmailsTopicName:           "outbound-emails",
				SearchIndexRequestsTopicName:      "search-index-requests",
				WebhookExecutionRequestsTopicName: "webhook-execution-requests",
				MobileNotificationsTopicName:      "mobile-notifications",
			},
			Pools: config.WorkerPoolsConfig{
				DeadLetterTopicName: "dead-letter",
			},
		}
		identityRepo := &identitymock.RepositoryMock{}
		dataPrivacyRepo := &dataprivacymock.RepositoryMock{}
		webhookRepo := &webhooksmock.RepositoryMock{}
		consumerProvider := &msgqueuemock.ConsumerProviderMock{}
		publisherProvider := &msgqueuemock.PublisherProviderMock{}
		analyticsEventReporter := &analyticsmock.EventReporterMock{}
		emailer := &emailmock.EmailerMock{}
		uploadManager := &uploadsmock.UploadManagerMock{}
		metricsProvider := &mockmetrics.ProviderMock{}
		decoder := &encodingmock.ServerEncoderDecoderMock{}
		coreDataIndexer := &identityindexing.UserDataIndexer{}
		eatingDataIndexer := &mealplanningindexing.MealPlanningDataIndexer{}

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
			PublishFunc:      func(_ context.Context, _ any) error { return nil },
			PublishAsyncFunc: func(_ context.Context, _ any) {},
			StopFunc:         func() {},
		}
		publisherProvider.NewPublisherFunc = func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return mockPublisher, nil
		}

		internalOpsRepo := &internalopsmock.InternalOpsDataManagerMock{}
		mealPlanRepo := &mealplanningmock.RepositoryMock{}
		prtManager := noopPasswordResetTokenDataManager{}
		notificationsRepo := &notificationsmock.RepositoryMock{}
		pushNotificationSender := noopnotifications.NewPushNotificationSender()

		handler, err := NewAsyncDataChangeMessageHandler(
			ctx,
			logger,
			tracerProvider,
			cfg,
			identityRepo,
			dataPrivacyRepo,
			webhookRepo,
			internalOpsRepo,
			consumerProvider,
			publisherProvider,
			analyticsEventReporter,
			emailer,
			uploadManager,
			metricsProvider,
			decoder,
			coreDataIndexer,
			eatingDataIndexer,
			mealPlanRepo,
			prtManager,
			notificationsRepo,
			pushNotificationSender,
		)

		assert.NoError(t, err)
		assert.NotNil(t, handler)
		assert.Equal(t, identityRepo, handler.identityRepo)
		assert.Equal(t, webhookRepo, handler.webhookRepo)
		assert.Equal(t, consumerProvider, handler.consumerProvider)
		assert.Equal(t, analyticsEventReporter, handler.analyticsEventReporter)
		assert.Equal(t, emailer, handler.emailer)
		assert.Equal(t, uploadManager, handler.uploadManager)
		assert.Equal(t, decoder, handler.decoder)
		assert.Equal(t, coreDataIndexer, handler.userDataIndexer)
		assert.Equal(t, eatingDataIndexer, handler.mealPlanningDataIndexer)

		// metricsProvider and publisherProvider are moq mocks - no testify assertion needed
	})
}

func TestAsyncDataChangeMessageHandler_SetNonWebhookEventTypes(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		eventTypes := []string{"event1", "event2", "event3"}
		handler.SetNonWebhookEventTypes(eventTypes)

		handler.nonWebhookEventTypesHat.RLock()
		assert.Equal(t, eventTypes, handler.nonWebhookEventTypes)
		handler.nonWebhookEventTypesHat.RUnlock()
	})
}
