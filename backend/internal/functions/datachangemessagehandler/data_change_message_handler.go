package datachangemessagehandler

import (
	"context"
	"errors"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/push"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"

	"github.com/primandproper/platform-go/v13/analytics"
	"github.com/primandproper/platform-go/v13/email"
	"github.com/primandproper/platform-go/v13/encoding"
	"github.com/primandproper/platform-go/v13/jobs"
	"github.com/primandproper/platform-go/v13/messagequeue"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	o11yName = "async_data_change_message_handler"

	topicDataChanges         = "data_changes"
	topicOutboundEmails      = "outbound_emails"
	topicMobileNotifications = "mobile_notifications"

	statusSuccess = "success"
	statusFailure = "failure"
	unknownValue  = "unknown"
)

// OutboundNotificationHandler handles outbound notifications for a domain's events.
// Returns true if the event was handled. May return emails to be published by the caller.
type OutboundNotificationHandler func(ctx context.Context, msg *audit.DataChangeMessage, user *identity.User) (handled bool, emailType string, emails []*queuemessages.OutboundEmailMessage, err error)

var errRequiredDataIsNil = errors.New("required data is nil")

// AsyncDataChangeMessageHandler is a cross-cutting event router that dispatches domain events to
// email and mobile notifications, and runs the Syncer that applies each search index's events.
// It necessarily references all domain repositories and event types. Domain-specific handler
// logic lives in dedicated files (e.g., mealplanning_handlers.go) to keep concerns separable.
//
// It does not publish index events. It used to: a handler picked a row ID out of a data change
// message and published an event onto the index's topic, which made indexing a dual write one
// hop downstream of the write it described. Index events are now enqueued into the outbox by
// the transaction that changed the row — see internal/repositories/postgres/events — and reach
// this process the same way every other message does, on the topic its Syncer consumes.
type AsyncDataChangeMessageHandler struct {
	tracer                                    tracing.Tracer
	internalOpsRepo                           internalops.InternalOpsDataManager
	logger                                    logging.Logger
	decoder                                   encoding.ServerEncoderDecoder
	outboundEmailsPublisher                   messagequeue.Publisher
	outboundEmailsExecutionTimeHistogram      metrics.Float64Histogram
	analyticsEventReporter                    analytics.EventReporter
	dataChangesExecutionTimeHistogram         metrics.Float64Histogram
	mobileNotificationsPublisher              messagequeue.Publisher
	emailer                                   email.Emailer
	identityRepo                              identity.Repository
	consumerProvider                          messagequeue.ConsumerProvider
	mealPlanRepo                              mealplanning.Repository
	pushFanout                                *push.Fanout
	handlerErrorsCounter                      metrics.Int64Counter
	messageDecodeErrorsCounter                metrics.Int64Counter
	messagesProcessedCounter                  metrics.Int64Counter
	emailsSentCounter                         metrics.Int64Counter
	emailsFailedCounter                       metrics.Int64Counter
	mobileNotificationsExecutionTimeHistogram metrics.Float64Histogram
	tracerProvider                            tracing.Provider
	metricsProvider                           metrics.Provider
	searchSyncers                             []SearchSyncer
	deadLetter                                jobs.DeadLetterFunc
	poolGroup                                 *jobs.PoolGroup
	queuesConfig                              queuescfg.Config
	baseURL                                   string
	outboundNotificationHandlers              []OutboundNotificationHandler
	poolsConfig                               config.WorkerPoolsConfig
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
	tracerProvider tracing.Provider,
	cfg *config.AsyncMessageHandlerConfig,
	identityRepo identity.Repository,
	internalOpsRepo internalops.InternalOpsDataManager,
	consumerProvider messagequeue.ConsumerProvider,
	publisherProvider messagequeue.PublisherProvider,
	analyticsEventReporter analytics.EventReporter,
	emailer email.Emailer,
	metricsProvider metrics.Provider,
	decoder encoding.ServerEncoderDecoder,
	searchSyncers []SearchSyncer,
	mealPlanRepo mealplanning.Repository,
	pushFanout *push.Fanout,
) (*AsyncDataChangeMessageHandler, error) {
	dataChangesExecutionTimeHistogram, err := metricsProvider.NewFloat64Histogram("data_changes_execution_time")
	if err != nil {
		return nil, fmt.Errorf("setting up dataChanges execution time histogram: %w", err)
	}

	outboundEmailsExecutionTimeHistogram, err := metricsProvider.NewFloat64Histogram("outbound_emails_execution_time")
	if err != nil {
		return nil, fmt.Errorf("setting up outboundEmails execution time histogram: %w", err)
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

	outboundEmailsPublisher, err := publisherProvider.NewPublisher(ctx, cfg.Queues.OutboundEmailsTopicName)
	if err != nil {
		return nil, fmt.Errorf("configuring outbound emails publisher: %w", err)
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

	handler := &AsyncDataChangeMessageHandler{
		tracer:                               tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:                               logging.NewNamedLogger(logger, o11yName),
		tracerProvider:                       tracerProvider,
		metricsProvider:                      metricsProvider,
		poolsConfig:                          cfg.Pools,
		deadLetter:                           deadLetter,
		identityRepo:                         identityRepo,
		internalOpsRepo:                      internalOpsRepo,
		consumerProvider:                     consumerProvider,
		analyticsEventReporter:               analyticsEventReporter,
		outboundEmailsPublisher:              outboundEmailsPublisher,
		queuesConfig:                         cfg.Queues,
		mobileNotificationsPublisher:         mobileNotificationsPublisher,
		emailer:                              emailer,
		dataChangesExecutionTimeHistogram:    dataChangesExecutionTimeHistogram,
		outboundEmailsExecutionTimeHistogram: outboundEmailsExecutionTimeHistogram,
		mobileNotificationsExecutionTimeHistogram: mobileNotificationsExecutionTimeHistogram,
		messagesProcessedCounter:                  messagesProcessedCounter,
		messageDecodeErrorsCounter:                messageDecodeErrorsCounter,
		handlerErrorsCounter:                      handlerErrorsCounter,
		emailsSentCounter:                         emailsSentCounter,
		emailsFailedCounter:                       emailsFailedCounter,
		decoder:                                   decoder,
		searchSyncers:                             searchSyncers,
		mealPlanRepo:                              mealPlanRepo,
		pushFanout:                                pushFanout,
		baseURL:                                   cfg.BaseURL,
	}

	// Register domain-specific event handlers.
	// When adding or removing a domain from this template, update these registrations.
	handler.outboundNotificationHandlers = []OutboundNotificationHandler{
		handler.handleMealPlanningOutboundNotification,
		handler.handleIdentityOutboundNotification,
	}

	// Built last, because the specs read the handler's own event handler factories.
	if handler.poolGroup, err = newPoolGroup(ctx, handler); err != nil {
		return nil, fmt.Errorf("configuring job pool group: %w", err)
	}

	return handler, nil
}
