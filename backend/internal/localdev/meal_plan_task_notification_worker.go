package localdev

import (
	"context"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/push"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	notificationsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notifications"
	mealplantasknotifications "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_notifications"

	"github.com/primandproper/platform-go/v13/database"
	platformnotifications "github.com/primandproper/platform-go/v13/notifications/mobile"
	"github.com/primandproper/platform-go/v13/observability/logging"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/workqueue"
	workqueuecfg "github.com/primandproper/platform-go/v13/workqueue/config"
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

	identityRepo := identityrepo.ProvideIdentityRepository(logger, tracerProvider, auditRepo, databaseClient, nil, uploads)
	mealPlanningRepo := mealplanningrepo.ProvideMealPlanningRepository(logger, tracerProvider, auditRepo, identityRepo, databaseClient, nil, uploads)
	notificationsRepo := notificationsrepo.ProvideNotificationsRepository(logger, tracerProvider, auditRepo, nil, databaseClient, nil)

	fanout, err := push.NewFanout(logger, notificationsRepo, sender, metricsProvider)
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
			identityRepo,
			fanout,
		),
		queue.Close,
		nil
}
