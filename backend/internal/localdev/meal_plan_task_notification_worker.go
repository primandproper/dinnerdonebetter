package localdev

import (
	"context"
	"fmt"

	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	notificationsstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notificationsstore"
	mealplantasknotifications "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_notifications"
	"github.com/primandproper/platform-go/v14/notifications/push"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/platform-go/v14/workqueue"
	workqueuecfg "github.com/primandproper/platform-go/v14/workqueue/config"
	"github.com/primandproper/primitives-go/v2/database"
	platformnotifications "github.com/primandproper/primitives-go/v2/notifications/mobile"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// NewMealPlanTaskNotificationWorker builds the prep task reminder worker over the given database,
// pushing through the given sender, and returns it with a close function for its queue.
//
// It is the notification half of what StartSagaWorker does for finalization, and exists for the
// same reason: the API server neither discovers tasks owed a reminder nor sends one, because both
// are the scheduler's job, so an in-process harness that wants to assert on reminders has to
// stand the worker up itself.
//
// The sender is a parameter rather than built from configuration because that is the whole point
// of running this in a test: a harness wants to record what was pushed, and there is no APNs to
// push it to. Everything else is real — the queue is a real workqueue.Queue over the real table,
// so the claim, the lease and the completion are the ones production runs.
//
// Unlike StartSagaWorker this starts no loop. The worker is driven by a scheduled job rather than
// by a Run of its own, so a caller advances it by calling Work, which is also what lets a test
// assert on what one pass did rather than waiting to see whether a second one happened.
//
// queueName partitions the table. Production has one name and uses it; a caller that runs two of
// these at once wants two, because two claimants on one logical queue is exactly what the lease
// permits — either may claim any item, so a test counting what its own sender received would
// otherwise be racing a sibling for the right to send it.
func NewMealPlanTaskNotificationWorker(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	databaseClient database.Client,
	sender platformnotifications.PushNotificationSender,
	queueName string,
	queueOpts ...func(*workqueue.Config),
) (worker *mealplantasknotifications.Worker, closeQueue func(context.Context) error, err error) {
	metricsProvider := metricsnoop.NewMetricsProvider()

	auditRepo, err := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, metricsProvider, databaseClient)
	if err != nil {
		return nil, nil, fmt.Errorf("building audit log repository: %w", err)
	}

	uploads, err := UploadsRegistry(logger, tracerProvider, databaseClient)
	if err != nil {
		return nil, nil, fmt.Errorf("building upload registry store: %w", err)
	}

	identityStore, err := platformidentity.NewSQLStore(databaseClient,
		platformidentity.WithTablePrefix(ddbidentity.TablePrefix),
		platformidentity.WithStoreLogger(logger),
		platformidentity.WithStoreTracerProvider(tracerProvider),
	)
	if err != nil {
		return nil, nil, err
	}

	mealPlanningRepo := mealplanningrepo.ProvideMealPlanningRepository(logger, tracerProvider, auditRepo, identityStore, databaseClient, nil, uploads)
	notificationsRepo, err := notificationsstore.ProvideAdapter(ctx, logger, tracerProvider, metricsnoop.NewMetricsProvider(), auditRepo, nil, databaseClient)
	if err != nil {
		return nil, nil, fmt.Errorf("building notifications repository: %w", err)
	}

	fanout, err := push.NewFanout(notificationsRepo.Registry(), sender,
		push.WithLogger(logger),
		push.WithTracerProvider(tracerProvider),
		push.WithMetricsProvider(metricsProvider))
	if err != nil {
		return nil, nil, fmt.Errorf("building push fanout: %w", err)
	}

	// The queue's own defaults otherwise. MaxAttempts is left unlimited unless a caller says
	// otherwise, so a harness cannot silently stall an item somebody is asserting on.
	cfg := &workqueue.Config{Name: queueName}
	for _, opt := range queueOpts {
		opt(cfg)
	}

	queue, err := workqueuecfg.NewQueue[string](ctx, cfg, databaseClient,
		workqueuecfg.WithLogger(logger),
		workqueuecfg.WithTracerProvider(tracerProvider),
		workqueuecfg.WithMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("building meal plan task notification queue: %w", err)
	}

	return mealplantasknotifications.NewWorker(
			logger,
			tracerProvider,
			&mealplantasknotifications.TaskQueue{Queue: queue},
			mealPlanningRepo,
			identityStore,
			databaseClient,
			fanout,
		),
		queue.Close,
		nil
}
