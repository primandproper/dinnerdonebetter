package scheduler

import (
	"context"

	commentstargets "github.com/primandproper/dinnerdonebetter/backend/internal/build/comments"
	dataprivacybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/sagas"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	notificationsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/manager"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/push"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	commentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	identityrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	internalopsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/internalops"
	issuereportsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	paymentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/payments"
	settingsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/settings"
	uploadedmediarepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	waitlistsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"
	webhooksrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooks"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"
	queuetest "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/workers/queue_test"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"
	mealplantasknotifications "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_notifications"

	"github.com/primandproper/platform-go/v13/database"
	databasecfg "github.com/primandproper/platform-go/v13/database/config"
	"github.com/primandproper/platform-go/v13/database/postgres"
	"github.com/primandproper/platform-go/v13/distributedlock"
	distributedlockcfg "github.com/primandproper/platform-go/v13/distributedlock/config"
	"github.com/primandproper/platform-go/v13/jobs"
	msgconfig "github.com/primandproper/platform-go/v13/messagequeue/config"
	notificationscfg "github.com/primandproper/platform-go/v13/notifications/mobile/config"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	loggingcfg "github.com/primandproper/platform-go/v13/observability/logging/config"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	metricscfg "github.com/primandproper/platform-go/v13/observability/metrics/config"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	tracingcfg "github.com/primandproper/platform-go/v13/observability/tracing/config"
	operationscfg "github.com/primandproper/platform-go/v13/operations/config"

	"github.com/samber/do/v2"
)

// BuildInjector creates and configures the dependency injection container.
//
// The container is the union of what the six periodic jobs used to build separately, one
// short-lived process each. Consolidating them means the connection pools, the tracer, and the
// repositories are constructed once at startup rather than once per tick — which is most of the
// cost of a job that runs every minute.
func BuildInjector(
	ctx context.Context,
	cfg *config.SchedulerConfig,
) *do.RootScope {
	i := do.New()

	do.ProvideValue(i, ctx)
	do.ProvideValue(i, cfg)

	RegisterConfigs(i)

	// platform providers
	observability.RegisterO11yConfigs(i)
	tracingcfg.RegisterTracerProvider(i)
	loggingcfg.RegisterLogger(i)
	metricscfg.RegisterMetricsProvider(i)
	databasecfg.RegisterClientConfig(i)
	postgres.RegisterDatabaseClient(i)
	msgconfig.RegisterMessageQueue(i)
	// The push sender. This process delivers prep task reminders itself rather than handing
	// them to the async message handler over a topic — see the meal plan task notification
	// worker for why the send has to happen under the queue lease that claimed the task.
	notificationscfg.RegisterPushSender(i)

	// repositories
	//
	// This process holds every domain repository, which it did not before, because the data
	// privacy fulfillment worker runs here and one export is a fan-out over all of them. That
	// is the cost of moving the gather off the message queue and onto a claimed row: the
	// process that claims has to be able to answer. It is paid once at startup — the pools and
	// the tracer are constructed here regardless — rather than per request.
	auditlogentries.RegisterAuditLogRepository(i)
	// No existence checks on the catalog: this process reads and erases comments
	// but never writes one, and the catalog gates writes rather than reads.
	commentstargets.RegisterReadOnlyTargets(i)
	commentsrepo.RegisterCommentsRepository(i)
	identityrepo.RegisterIdentityRepository(i)
	internalopsrepo.RegisterInternalOpsRepository(i)
	issuereportsrepo.RegisterIssueReportsRepository(i)
	paymentsrepo.RegisterPaymentsRepository(i)
	uploadedmediarepo.RegisterUploadedMediaRepository(i)
	settingsrepo.RegisterSettingsRepository(i)
	waitlistsrepo.RegisterWaitlistsRepository(i)
	// This also registers the webhook Store and Dispatcher, which this process needs in both
	// directions: dispatch happens inside the transaction that causes the event, and the meal
	// plan finalizer emits events like any request does.
	webhooksrepo.RegisterWebhooksRepository(i)

	// The notifications manager and the push fan-out over it. The manager registers the
	// notifications repository on its own, which is why that one is absent from the list
	// above. The fan-out is the part of mobile notifications that has nothing to do with why
	// one is owed: device tokens in, pushes out, dead tokens retired — and the async message
	// handler builds the same one.
	notificationsmanager.RegisterNotificationsDataManager(i)
	push.RegisterFanout(i)

	// The data privacy machinery: the bucket and cipher artifacts are written with, the
	// registry of who holds data about a person, the worker that fulfills, and the sweeper
	// that expires. A deployment that ran the worker and not the sweeper would accumulate
	// artifacts forever, which is why the sweeper is a registered job rather than a flag.
	dataprivacycfg.RegisterArtifactStorage(i)
	dataprivacybuild.RegisterRegistry(i)
	dataprivacybuild.RegisterOperationsRegistry(i)
	dataprivacybuild.RegisterSweeper(i)

	// The operations tier privacy requests are now fulfilled through. This process runs the
	// whole of it: the store and queue it shares with the API server, and the worker that
	// claims operations and runs the kinds the registry above holds.
	operationscfg.RegisterStore(i)
	operationscfg.RegisterQueue(i)
	operationscfg.RegisterService(i)
	operationscfg.RegisterWorker(i)

	// The delivery side: the worker that claims the dispatch rows the write side above
	// produces, signs them, and sends them.
	RegisterWebhookWorker(i)
	// Domain: mealplanning
	events.RegisterOutboxEmitter(i)
	mealplanningrepo.RegisterMealPlanningRepository(i)
	grocerylistpreparation.RegisterGroceryListCreator(i)
	recipeanalysis.RegisterRecipeAnalyzer(i)

	// the periodic jobs themselves
	mealplanfinalization.RegisterStarter(i)
	queuetest.RegisterQueueTest(i)

	do.Provide[*queuetest.JobParams](i, func(i do.Injector) (*queuetest.JobParams, error) {
		return &queuetest.JobParams{Queues: *do.MustInvoke[*queuescfg.Config](i)}, nil
	})

	// The prep task reminder queue and the worker that fills and drains it. The queue owns a
	// goroutine and is closed by cmd/ddb's shutdown rather than by the container, which is
	// what every other background component in this process does with its Close.
	mealplantasknotifications.RegisterQueue(i)
	mealplantasknotifications.RegisterWorker(i)

	// The Reindexers this process drives on a schedule. They come from the same registration
	// as the Syncers the consumer runs, because the two are halves of keeping one index right.
	identityindexing.RegisterUserReindexer(i)
	mealplanningindexing.RegisterIndexSyncers(i)

	// the lock that decides which replica runs a given tick
	do.Provide[distributedlock.Locker](i, func(i do.Injector) (distributedlock.Locker, error) {
		return distributedlockcfg.NewLocker(
			do.MustInvoke[context.Context](i),
			&do.MustInvoke[*config.ScheduledJobsConfig](i).Lock,
			do.MustInvoke[database.Client](i),
			distributedlockcfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			distributedlockcfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			distributedlockcfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	// The saga machinery, and the one worker that advances every definition in the process.
	// Registered after the lock, which it takes a per-instance scope of.
	sagas.RegisterSagas(i)
	sagas.RegisterSagaWorker(i)

	RegisterMeteringFlusher(i)

	RegisterScheduler(i)
	RegisterOutboxRelay(i)
	RegisterRetentionSweeper(i)

	return i
}

// Build builds the scheduler.
func Build(
	ctx context.Context,
	cfg *config.SchedulerConfig,
) (*jobs.Scheduler, error) {
	i := BuildInjector(ctx, cfg)
	return do.MustInvoke[*jobs.Scheduler](i), nil
}
