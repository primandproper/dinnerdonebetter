package datachangemessagehandler

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	queuemessages "github.com/primandproper/dinnerdonebetter/backend/internal/queues/messages"
	coreemails "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/emails"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/notifications/mail"
	"github.com/primandproper/platform-go/v15/notifications/push"
	"github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/analytics"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/email"
	"github.com/primandproper/primitives-go/v2/jobs"
	"github.com/primandproper/primitives-go/v2/messagequeue"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
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

// OutboundNotificationHandler answers, for one event, whether it is a domain's to notify about
// and if so the mail to send: handled is false for an event the handler does not recognize, and
// the caller tries the next one. emailType labels the mail in the logs.
//
// Each domain contributes one through RegisterAsyncDataChangeMessageHandler; the handler over
// platform's own identity events is this package's.
type OutboundNotificationHandler func(ctx context.Context, event *webhooks.Envelope) (handled bool, emailType string, emails []*queuemessages.OutboundEmailMessage, err error)

// OutboundNotifier builds one domain's OutboundNotificationHandler from the injector, once the
// process has registered what it reads through.
type OutboundNotifier func(i do.Injector) (OutboundNotificationHandler, error)

var errRequiredDataIsNil = errors.New("required data is nil")

// AsyncDataChangeMessageHandler is a cross-cutting event router that dispatches domain events to
// email and mobile notifications, and runs the Syncer that applies each search index's events.
// It names no domain: which events imply a mail is each domain's OutboundNotificationHandler's
// to say, and which indexes it drains is the Registry's.
//
// It does not publish index events. It used to: a handler picked a row ID out of a data change
// message and published an event onto the index's topic, which made indexing a dual write one
// hop downstream of the write it described. Index events are now enqueued into the outbox by
// the transaction that changed the row — see internal/recordingspine — and reach
// this process the same way every other message does, on the topic its Syncer consumes.
type AsyncDataChangeMessageHandler struct {
	tracer                                    tracing.Tracer
	internalOpsRepo                           internalops.InternalOpsDataManager
	logger                                    logging.Logger
	outboundEmailsPublisher                   messagequeue.Publisher
	outboundEmailsExecutionTimeHistogram      metrics.Float64Histogram
	analyticsEventReporter                    analytics.EventReporter
	dataChangesExecutionTimeHistogram         metrics.Float64Histogram
	mobileNotificationsPublisher              messagequeue.Publisher
	emailer                                   email.Emailer
	mailDrainer                               *mail.Drainer
	directory                                 platformidentity.Store
	db                                        database.Client
	consumerProvider                          messagequeue.ConsumerProvider
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
	directory platformidentity.Store,
	db database.Client,
	internalOpsRepo internalops.InternalOpsDataManager,
	consumerProvider messagequeue.ConsumerProvider,
	publisherProvider messagequeue.PublisherProvider,
	analyticsEventReporter analytics.EventReporter,
	emailer email.Emailer,
	metricsProvider metrics.Provider,
	searchSyncers []SearchSyncer,
	outboundNotificationHandlers []OutboundNotificationHandler,
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

	// The mail platform's identity, sign-in, password reset and waitlist doors queue, taken off
	// its own topic, rendered in this application's words and sent. platform owns the transport;
	// the wording, and the one read it needs, is internal/services/identity/emails.
	mailDrainer, err := mail.NewDrainer(emailer, coreemails.NewRenderer(directory, db, cfg.BaseURL),
		mail.WithLogger(logger),
		mail.WithTracerProvider(tracerProvider),
		mail.WithMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("configuring the queued mail drainer: %w", err)
	}

	handler := &AsyncDataChangeMessageHandler{
		tracer:                               tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:                               logging.NewNamedLogger(logger, o11yName),
		tracerProvider:                       tracerProvider,
		metricsProvider:                      metricsProvider,
		poolsConfig:                          cfg.Pools,
		deadLetter:                           deadLetter,
		directory:                            directory,
		db:                                   db,
		internalOpsRepo:                      internalOpsRepo,
		consumerProvider:                     consumerProvider,
		analyticsEventReporter:               analyticsEventReporter,
		outboundEmailsPublisher:              outboundEmailsPublisher,
		queuesConfig:                         cfg.Queues,
		mobileNotificationsPublisher:         mobileNotificationsPublisher,
		emailer:                              emailer,
		mailDrainer:                          mailDrainer,
		dataChangesExecutionTimeHistogram:    dataChangesExecutionTimeHistogram,
		outboundEmailsExecutionTimeHistogram: outboundEmailsExecutionTimeHistogram,
		mobileNotificationsExecutionTimeHistogram: mobileNotificationsExecutionTimeHistogram,
		messagesProcessedCounter:                  messagesProcessedCounter,
		messageDecodeErrorsCounter:                messageDecodeErrorsCounter,
		handlerErrorsCounter:                      handlerErrorsCounter,
		emailsSentCounter:                         emailsSentCounter,
		emailsFailedCounter:                       emailsFailedCounter,
		searchSyncers:                             searchSyncers,
		pushFanout:                                pushFanout,
		baseURL:                                   cfg.BaseURL,
	}

	// The domains' handlers as handed in, then this package's own over platform's identity
	// events. The first to claim an event answers for it, so two handlers claiming one event
	// type is a mistake the order would hide; every handler here claims a disjoint set.
	handler.outboundNotificationHandlers = append(slices.Clone(outboundNotificationHandlers), handler.handleIdentityOutboundNotification)

	// Built last, because the specs read the handler's own event handler factories.
	if handler.poolGroup, err = newPoolGroup(ctx, handler); err != nil {
		return nil, fmt.Errorf("configuring job pool group: %w", err)
	}

	return handler, nil
}
