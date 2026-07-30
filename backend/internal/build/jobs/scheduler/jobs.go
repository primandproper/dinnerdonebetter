package scheduler

import (
	"context"

	mobilenotificationscheduler "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/jobs/mobile_notification_scheduler"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	queuetest "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/internalops/workers/queue_test"
	mealplanfinalizer "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalizer"
	mealplangrocerylistinitializer "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_grocery_list_initializer"
	mealplantaskcreator "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_creator"

	"github.com/primandproper/platform-go/v8/distributedlock"
	"github.com/primandproper/platform-go/v8/jobs"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/metrics"
	"github.com/primandproper/platform-go/v8/observability/tracing"
	"github.com/primandproper/platform-go/v8/search/text/indexing"

	"github.com/samber/do/v2"
)

// Job names. These are also the distributed lock keys, so renaming one lets an old replica and
// a new replica both run that job during a rollout.
const (
	jobMealPlanFinalizer              = "meal_plan_finalizer"
	jobMealPlanGroceryListInitializer = "meal_plan_grocery_list_initializer"
	jobMealPlanTaskCreator            = "meal_plan_task_creator"
	jobSearchDataIndexScheduler       = "search_data_index_scheduler"
	jobMobileNotificationScheduler    = "mobile_notification_scheduler"
	jobQueueTest                      = "queue_test"
)

// RegisterScheduler registers the jobs.Scheduler, with every enabled job already registered on
// it, with the injector.
func RegisterScheduler(i do.Injector) {
	do.Provide[*jobs.Scheduler](i, func(i do.Injector) (*jobs.Scheduler, error) {
		jobsCfg := do.MustInvoke[*config.ScheduledJobsConfig](i)

		scheduler, err := jobs.NewScheduler(
			do.MustInvoke[context.Context](i),
			&jobsCfg.Scheduler,
			do.MustInvoke[distributedlock.Locker](i),
			jobs.WithSchedulerLogger(do.MustInvoke[logging.Logger](i)),
			jobs.WithSchedulerTracerProvider(do.MustInvoke[tracing.TracerProvider](i)),
			jobs.WithSchedulerMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
		if err != nil {
			return nil, err
		}

		registrations := []struct {
			run  func(ctx context.Context) error
			cfg  *config.ScheduledJobConfig
			name string
		}{
			{
				name: jobMealPlanFinalizer,
				cfg:  &jobsCfg.MealPlanning.MealPlanFinalizer,
				// The finalizer reports how many meal plans it changed; the scheduler has
				// nowhere to put a count, and the worker already records it as a metric.
				run: func(ctx context.Context) error {
					_, workErr := do.MustInvoke[*mealplanfinalizer.Worker](i).Work(ctx)
					return workErr
				},
			},
			{
				name: jobMealPlanGroceryListInitializer,
				cfg:  &jobsCfg.MealPlanning.MealPlanGroceryListInitializer,
				run:  do.MustInvoke[*mealplangrocerylistinitializer.Worker](i).Work,
			},
			{
				name: jobMealPlanTaskCreator,
				cfg:  &jobsCfg.MealPlanning.MealPlanTaskCreator,
				run:  do.MustInvoke[*mealplantaskcreator.Worker](i).Work,
			},
			{
				name: jobSearchDataIndexScheduler,
				cfg:  &jobsCfg.SearchDataIndexScheduler,
				run:  do.MustInvoke[*indexing.IndexScheduler](i).IndexTypes,
			},
			{
				name: jobMobileNotificationScheduler,
				cfg:  &jobsCfg.MobileNotificationScheduler,
				run:  do.MustInvoke[*mobilenotificationscheduler.Scheduler](i).ScheduleNotifications,
			},
			{
				name: jobQueueTest,
				cfg:  &jobsCfg.QueueTest,
				run:  do.MustInvoke[*queuetest.Job](i).Do,
			},
		}

		for idx := range registrations {
			r := &registrations[idx]

			if !r.cfg.Enabled {
				continue
			}

			if err = scheduler.Register(jobs.Job{
				Name:       r.name,
				Interval:   r.cfg.Interval,
				Timeout:    r.cfg.Timeout,
				LeaseTTL:   r.cfg.LeaseTTL,
				RunOnStart: r.cfg.RunOnStart,
				Run:        r.run,
			}); err != nil {
				return nil, err
			}
		}

		return scheduler, nil
	})
}
