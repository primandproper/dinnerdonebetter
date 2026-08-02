package datachangemessagehandler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	notificationsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/manager"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v9/analytics"
	"github.com/primandproper/platform-go/v9/email"
	"github.com/primandproper/platform-go/v9/encoding"
	"github.com/primandproper/platform-go/v9/httpclient"
	"github.com/primandproper/platform-go/v9/jobs"
	"github.com/primandproper/platform-go/v9/messagequeue"
	platformnotifications "github.com/primandproper/platform-go/v9/notifications/mobile"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	"github.com/primandproper/platform-go/v9/uploads"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	o11yName = "async_data_change_message_handler"

	topicDataChanges              = "data_changes"
	topicOutboundEmails           = "outbound_emails"
	topicSearchIndexRequests      = "search_index_requests"
	topicWebhookExecutionRequests = "webhook_execution_requests"
	topicUserDataAggregation      = "user_data_aggregation"
	topicMobileNotifications      = "mobile_notifications"

	statusSuccess = "success"
	statusFailure = "failure"
	unknownValue  = "unknown"
)

// SearchIndexEventHandler handles search index updates for a domain's events.
// Returns true if the event was handled, false to fall through to other handlers.
type SearchIndexEventHandler func(ctx context.Context, msg *audit.DataChangeMessage) (handled bool, err error)

// OutboundNotificationHandler handles outbound notifications for a domain's events.
// Returns true if the event was handled. May return emails to be published by the caller.
type OutboundNotificationHandler func(ctx context.Context, msg *audit.DataChangeMessage, user *identity.User) (handled bool, emailType string, emails []*queuemessages.OutboundEmailMessage, err error)

var (
	errRequiredDataIsNil = errors.New("required data is nil")

	errPoolsAlreadyStarted = errors.New("job pools already started")
)

// AsyncDataChangeMessageHandler is a cross-cutting event router that dispatches domain events to
// search indexing, email, webhooks, and mobile notifications. It necessarily references all domain
// repositories and event types. Domain-specific handler logic lives in dedicated files
// (e.g., mealplanning_handlers.go) to keep concerns separable.
type AsyncDataChangeMessageHandler struct {
	uploadManager                             uploads.UploadManager
	tracer                                    tracing.Tracer
	dataPrivacyRepo                           dataprivacy.Repository
	internalOpsRepo                           internalops.InternalOpsDataManager
	logger                                    logging.Logger
	decoder                                   encoding.ServerEncoderDecoder
	webhookExecutionTimestampHistogram        metrics.Float64Histogram
	userDataAggregationExecutionTimeHistogram metrics.Float64Histogram
	outboundEmailsPublisher                   messagequeue.Publisher
	webhookRepo                               webhooks.Repository
	outboundEmailsExecutionTimeHistogram      metrics.Float64Histogram
	analyticsEventReporter                    analytics.EventReporter
	dataChangesExecutionTimeHistogram         metrics.Float64Histogram
	webhookExecutionRequestPublisher          messagequeue.Publisher
	mobileNotificationsPublisher              messagequeue.Publisher
	emailer                                   email.Emailer
	identityRepo                              identity.Repository
	searchDataIndexPublisher                  messagequeue.Publisher
	consumerProvider                          messagequeue.ConsumerProvider
	searchIndexRequestsExecutionTimeHistogram metrics.Float64Histogram
	badDeviceTokensArchivedCounter            metrics.Int64Counter
	pushNotificationsSentCounter              metrics.Int64Counter
	mealPlanRepo                              mealplanning.Repository
	passwordResetTokenDataManager             auth.PasswordResetTokenDataManager
	notificationsRepo                         notificationsmanager.NotificationsDataManager
	pushNotificationSender                    platformnotifications.PushNotificationSender
	handlerErrorsCounter                      metrics.Int64Counter
	messageDecodeErrorsCounter                metrics.Int64Counter
	messagesProcessedCounter                  metrics.Int64Counter
	emailsSentCounter                         metrics.Int64Counter
	emailsFailedCounter                       metrics.Int64Counter
	mobileNotificationsExecutionTimeHistogram metrics.Float64Histogram
	tracerProvider                            tracing.TracerProvider
	metricsProvider                           metrics.Provider
	mealPlanningDataIndexer                   *mealplanningindexing.MealPlanningDataIndexer
	userDataIndexer                           *identityindexing.UserDataIndexer
	webhookHTTPClient                         *http.Client
	deadLetter                                jobs.DeadLetterFunc
	queuesConfig                              queuescfg.Config
	baseURL                                   string
	searchIndexHandlers                       []SearchIndexEventHandler
	outboundNotificationHandlers              []OutboundNotificationHandler
	pools                                     []*jobs.Pool
	nonWebhookEventTypes                      []string
	poolsConfig                               config.WorkerPoolsConfig
	poolsWG                                   sync.WaitGroup
	nonWebhookEventTypesHat                   sync.RWMutex
}

func (a *AsyncDataChangeMessageHandler) SetNonWebhookEventTypes(nonWebhookEventTypes []string) {
	a.nonWebhookEventTypesHat.Lock()
	defer a.nonWebhookEventTypesHat.Unlock()
	a.nonWebhookEventTypes = nonWebhookEventTypes
}

func (a *AsyncDataChangeMessageHandler) recordMessagesProcessed(ctx context.Context, topic, status string) {
	a.messagesProcessedCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("topic", topic),
		attribute.String("status", status),
	))
}

func NewAsyncDataChangeMessageHandler(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	cfg *config.AsyncMessageHandlerConfig,
	identityRepo identity.Repository,
	dataPrivacyRepo dataprivacy.Repository,
	webhookRepo webhooks.Repository,
	internalOpsRepo internalops.InternalOpsDataManager,
	consumerProvider messagequeue.ConsumerProvider,
	publisherProvider messagequeue.PublisherProvider,
	analyticsEventReporter analytics.EventReporter,
	emailer email.Emailer,
	uploadManager uploads.UploadManager,
	metricsProvider metrics.Provider,
	decoder encoding.ServerEncoderDecoder,
	coreDataIndexer *identityindexing.UserDataIndexer,
	eatingDataIndexer *mealplanningindexing.MealPlanningDataIndexer,
	mealPlanRepo mealplanning.Repository,
	passwordResetTokenDataManager auth.PasswordResetTokenDataManager,
	notificationsRepo notificationsmanager.NotificationsDataManager,
	pushNotificationSender platformnotifications.PushNotificationSender,
) (*AsyncDataChangeMessageHandler, error) {
	dataChangesExecutionTimeHistogram, err := metricsProvider.NewFloat64Histogram("data_changes_execution_time")
	if err != nil {
		return nil, fmt.Errorf("setting up dataChanges execution time histogram: %w", err)
	}

	outboundEmailsExecutionTimeHistogram, err := metricsProvider.NewFloat64Histogram("outbound_emails_execution_time")
	if err != nil {
		return nil, fmt.Errorf("setting up outboundEmails execution time histogram: %w", err)
	}

	searchIndexRequestsExecutionTimeHistogram, err := metricsProvider.NewFloat64Histogram("search_index_requests_execution_time")
	if err != nil {
		return nil, fmt.Errorf("setting up searchIndexRequests execution time histogram: %w", err)
	}

	userDataAggregationExecutionTimeHistogram, err := metricsProvider.NewFloat64Histogram("user_data_aggregation_execution_time")
	if err != nil {
		return nil, fmt.Errorf("setting up userDataAggregation execution time histogram: %w", err)
	}

	webhookExecutionTimestampHistogram, err := metricsProvider.NewFloat64Histogram("webhook_requests_execution_time")
	if err != nil {
		return nil, fmt.Errorf("setting up webhookExecutionRequests execution time histogram: %w", err)
	}

	mobileNotificationsExecutionTimeHistogram, err := metricsProvider.NewFloat64Histogram("mobile_notifications_execution_time")
	if err != nil {
		return nil, fmt.Errorf("setting up mobileNotifications execution time histogram: %w", err)
	}

	messagesProcessedCounter, err := metricsProvider.NewInt64Counter("messages_processed_total")
	if err != nil {
		return nil, fmt.Errorf("setting up messages processed counter: %w", err)
	}

	messageDecodeErrorsCounter, err := metricsProvider.NewInt64Counter("message_decode_errors_total")
	if err != nil {
		return nil, fmt.Errorf("setting up message decode errors counter: %w", err)
	}

	handlerErrorsCounter, err := metricsProvider.NewInt64Counter("handler_errors_total")
	if err != nil {
		return nil, fmt.Errorf("setting up handler errors counter: %w", err)
	}

	emailsSentCounter, err := metricsProvider.NewInt64Counter("emails_sent_total")
	if err != nil {
		return nil, fmt.Errorf("setting up emails sent counter: %w", err)
	}

	emailsFailedCounter, err := metricsProvider.NewInt64Counter("emails_failed_total")
	if err != nil {
		return nil, fmt.Errorf("setting up emails failed counter: %w", err)
	}

	pushNotificationsSentCounter, err := metricsProvider.NewInt64Counter("push_notifications_sent_total")
	if err != nil {
		return nil, fmt.Errorf("setting up push notifications sent counter: %w", err)
	}

	badDeviceTokensArchivedCounter, err := metricsProvider.NewInt64Counter("bad_device_tokens_archived_total")
	if err != nil {
		return nil, fmt.Errorf("setting up bad device tokens archived counter: %w", err)
	}

	outboundEmailsPublisher, err := publisherProvider.NewPublisher(ctx, cfg.Queues.OutboundEmailsTopicName)
	if err != nil {
		return nil, fmt.Errorf("configuring outbound emails publisher: %w", err)
	}

	searchDataIndexPublisher, err := publisherProvider.NewPublisher(ctx, cfg.Queues.SearchIndexRequestsTopicName)
	if err != nil {
		return nil, fmt.Errorf("configuring search indexing publisher: %w", err)
	}

	webhookExecutionRequestPublisher, err := publisherProvider.NewPublisher(ctx, cfg.Queues.WebhookExecutionRequestsTopicName)
	if err != nil {
		return nil, fmt.Errorf("configuring webhook execution requests publisher: %w", err)
	}

	mobileNotificationsPublisher, err := publisherProvider.NewPublisher(ctx, cfg.Queues.MobileNotificationsTopicName)
	if err != nil {
		return nil, fmt.Errorf("configuring mobile notifications publisher: %w", err)
	}

	// A pool with no dead-letter destination drops exhausted messages, so this is built here
	// rather than lazily: a broken dead-letter topic should fail startup, not the first
	// message that needs it.
	deadLetter, err := jobs.NewTopicDeadLetter(ctx, publisherProvider, cfg.Pools.DeadLetterTopicName)
	if err != nil {
		return nil, fmt.Errorf("configuring dead letter publisher: %w", err)
	}

	// One client for every delivery: a client built per delivery gets its own connection pool,
	// so every webhook pays for a TLS handshake that no subsequent delivery can reuse.
	webhookHTTPClient := httpclient.NewHTTPClient(httpclient.WithTracing(true))

	handler := &AsyncDataChangeMessageHandler{
		tracer:                               tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:                               logging.NewNamedLogger(logger, o11yName),
		tracerProvider:                       tracerProvider,
		metricsProvider:                      metricsProvider,
		poolsConfig:                          cfg.Pools,
		webhookHTTPClient:                    webhookHTTPClient,
		deadLetter:                           deadLetter,
		nonWebhookEventTypes:                 []string{},
		identityRepo:                         identityRepo,
		dataPrivacyRepo:                      dataPrivacyRepo,
		webhookRepo:                          webhookRepo,
		internalOpsRepo:                      internalOpsRepo,
		consumerProvider:                     consumerProvider,
		analyticsEventReporter:               analyticsEventReporter,
		outboundEmailsPublisher:              outboundEmailsPublisher,
		searchDataIndexPublisher:             searchDataIndexPublisher,
		queuesConfig:                         cfg.Queues,
		webhookExecutionRequestPublisher:     webhookExecutionRequestPublisher,
		mobileNotificationsPublisher:         mobileNotificationsPublisher,
		emailer:                              emailer,
		uploadManager:                        uploadManager,
		dataChangesExecutionTimeHistogram:    dataChangesExecutionTimeHistogram,
		outboundEmailsExecutionTimeHistogram: outboundEmailsExecutionTimeHistogram,
		searchIndexRequestsExecutionTimeHistogram: searchIndexRequestsExecutionTimeHistogram,
		userDataAggregationExecutionTimeHistogram: userDataAggregationExecutionTimeHistogram,
		webhookExecutionTimestampHistogram:        webhookExecutionTimestampHistogram,
		mobileNotificationsExecutionTimeHistogram: mobileNotificationsExecutionTimeHistogram,
		messagesProcessedCounter:                  messagesProcessedCounter,
		messageDecodeErrorsCounter:                messageDecodeErrorsCounter,
		handlerErrorsCounter:                      handlerErrorsCounter,
		emailsSentCounter:                         emailsSentCounter,
		emailsFailedCounter:                       emailsFailedCounter,
		pushNotificationsSentCounter:              pushNotificationsSentCounter,
		badDeviceTokensArchivedCounter:            badDeviceTokensArchivedCounter,
		decoder:                                   decoder,
		userDataIndexer:                           coreDataIndexer,
		mealPlanningDataIndexer:                   eatingDataIndexer,
		mealPlanRepo:                              mealPlanRepo,
		passwordResetTokenDataManager:             passwordResetTokenDataManager,
		notificationsRepo:                         notificationsRepo,
		pushNotificationSender:                    pushNotificationSender,
		baseURL:                                   cfg.BaseURL,
	}

	// Register domain-specific event handlers.
	// When adding or removing a domain from this template, update these registrations.
	handler.searchIndexHandlers = []SearchIndexEventHandler{
		handler.handleMealPlanningSearchIndexUpdate,
		handler.handleIdentitySearchIndexUpdate,
	}
	handler.outboundNotificationHandlers = []OutboundNotificationHandler{
		handler.handleMealPlanningOutboundNotification,
		handler.handleIdentityOutboundNotification,
	}

	return handler, nil
}
