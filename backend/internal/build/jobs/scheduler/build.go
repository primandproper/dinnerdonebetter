package scheduler

import (
	"context"

	mobilenotificationscheduler "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/mobile_notification_scheduler"
	searchdataindexscheduler "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/search_data_index_scheduler"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	identityrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	internalopsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/internalops"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	queuetest "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/workers/queue_test"
	mealplanfinalizer "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalizer"
	mealplangrocerylistinitializer "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_grocery_list_initializer"
	mealplantaskcreator "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_creator"

	"github.com/primandproper/platform-go/v9/database"
	databasecfg "github.com/primandproper/platform-go/v9/database/config"
	"github.com/primandproper/platform-go/v9/database/postgres"
	"github.com/primandproper/platform-go/v9/distributedlock"
	distributedlockcfg "github.com/primandproper/platform-go/v9/distributedlock/config"
	"github.com/primandproper/platform-go/v9/jobs"
	"github.com/primandproper/platform-go/v9/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v9/messagequeue/config"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	loggingcfg "github.com/primandproper/platform-go/v9/observability/logging/config"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	metricscfg "github.com/primandproper/platform-go/v9/observability/metrics/config"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	tracingcfg "github.com/primandproper/platform-go/v9/observability/tracing/config"
	"github.com/primandproper/platform-go/v9/search/text/indexing"

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
	auditlogentries.RegisterAuditLogRepository(i)
	identityrepo.RegisterIdentityRepository(i)
	internalopsrepo.RegisterInternalOpsRepository(i)
	// Domain: mealplanning
	events.RegisterOutboxEmitter(i)
	mealplanningrepo.RegisterMealPlanningRepository(i)
	grocerylistpreparation.RegisterGroceryListCreator(i)
	recipeanalysis.RegisterRecipeAnalyzer(i)

	// the periodic jobs themselves
	mealplanfinalizer.RegisterMealPlanFinalizer(i)
	mealplangrocerylistinitializer.RegisterMealPlanGroceryListInitializer(i)
	mealplantaskcreator.RegisterMealPlanTaskCreator(i)
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
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[identity.Repository](i),
			publisher,
		), nil
	})

	do.Provide[map[string]indexing.Function](i, func(i do.Injector) (map[string]indexing.Function, error) {
		return searchdataindexscheduler.ProvideIndexFunctions(
			do.MustInvoke[identity.Repository](i),
			do.MustInvoke[mealplanning.Repository](i),
		), nil
	})
	indexing.RegisterIndexScheduler(i, do.MustInvoke[*queuescfg.Config](i).SearchIndexRequestsTopicName)

	// the lock that decides which replica runs a given tick
	do.Provide[distributedlock.Locker](i, func(i do.Injector) (distributedlock.Locker, error) {
		return distributedlockcfg.NewLocker(
			do.MustInvoke[context.Context](i),
			&do.MustInvoke[*config.ScheduledJobsConfig](i).Lock,
			do.MustInvoke[database.Client](i),
			distributedlockcfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			distributedlockcfg.WithTracerProvider(do.MustInvoke[tracing.TracerProvider](i)),
			distributedlockcfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	RegisterScheduler(i)
	RegisterOutboxRelay(i)

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
