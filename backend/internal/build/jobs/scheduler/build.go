package scheduler

import (
	"context"

	dataprivacybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/dataprivacy"
	mobilenotificationscheduler "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/mobile_notification_scheduler"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/sagas"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	commentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	identityrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	internalopsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/internalops"
	issuereportsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	notificationsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notifications"
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

	"github.com/primandproper/platform-go/v12/database"
	databasecfg "github.com/primandproper/platform-go/v12/database/config"
	"github.com/primandproper/platform-go/v12/database/postgres"
	"github.com/primandproper/platform-go/v12/distributedlock"
	distributedlockcfg "github.com/primandproper/platform-go/v12/distributedlock/config"
	"github.com/primandproper/platform-go/v12/jobs"
	"github.com/primandproper/platform-go/v12/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v12/messagequeue/config"
	"github.com/primandproper/platform-go/v12/observability"
	"github.com/primandproper/platform-go/v12/observability/logging"
	loggingcfg "github.com/primandproper/platform-go/v12/observability/logging/config"
	"github.com/primandproper/platform-go/v12/observability/metrics"
	metricscfg "github.com/primandproper/platform-go/v12/observability/metrics/config"
	"github.com/primandproper/platform-go/v12/observability/tracing"
	tracingcfg "github.com/primandproper/platform-go/v12/observability/tracing/config"
	operationscfg "github.com/primandproper/platform-go/v12/operations/config"

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

	// repositories
	//
	// This process holds every domain repository, which it did not before, because the data
	// privacy fulfillment worker runs here and one export is a fan-out over all of them. That
	// is the cost of moving the gather off the message queue and onto a claimed row: the
	// process that claims has to be able to answer. It is paid once at startup — the pools and
	// the tracer are constructed here regardless — rather than per request.
	auditlogentries.RegisterAuditLogRepository(i)
	commentsrepo.RegisterCommentsRepository(i)
	identityrepo.RegisterIdentityRepository(i)
	internalopsrepo.RegisterInternalOpsRepository(i)
	issuereportsrepo.RegisterIssueReportsRepository(i)
	notificationsrepo.RegisterNotificationsRepository(i)
	paymentsrepo.RegisterPaymentsRepository(i)
	settingsrepo.RegisterSettingsRepository(i)
	uploadedmediarepo.RegisterUploadedMediaRepository(i)
	waitlistsrepo.RegisterWaitlistsRepository(i)
	// This also registers the webhook Store and Dispatcher, which this process needs in both
	// directions: dispatch happens inside the transaction that causes the event, and the meal
	// plan finalizer emits events like any request does.
	webhooksrepo.RegisterWebhooksRepository(i)

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

	// The mobile notification scheduler is constructed here rather than through its own
	// RegisterConfigs: that one also registers a bare messagequeue.Publisher bound to the
	// mobile notifications topic, which in a container shared with five other jobs would be
	// an unlabeled publisher that anything asking for "a publisher" would receive.
	do.Provide[*mobilenotificationscheduler.Scheduler](i, func(i do.Injector) (*mobilenotificationscheduler.Scheduler, error) {
		publisher, err := do.MustInvoke[messagequeue.PublisherProvider](i).NewPublisher(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[*queuescfg.Config](i).MobileNotificationsTopicName,
		)
		if err != nil {
			return nil, err
		}

		return mobilenotificationscheduler.NewScheduler(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[identity.Repository](i),
			publisher,
		), nil
	})

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
