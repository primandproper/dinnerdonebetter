package scheduler

import (
	"context"

	mobilenotificationscheduler "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/mobile_notification_scheduler"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	queuetest "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/workers/queue_test"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"

	"github.com/primandproper/platform-go/v9/audit"
	platformdataprivacy "github.com/primandproper/platform-go/v9/dataprivacy"
	"github.com/primandproper/platform-go/v9/distributedlock"
	"github.com/primandproper/platform-go/v9/jobs"
	"github.com/primandproper/platform-go/v9/metering"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	"github.com/primandproper/platform-go/v9/search/text/indexing"

	"github.com/samber/do/v2"
)

// Job names. These are also the distributed lock keys, so renaming one lets an old replica and
// a new replica both run that job during a rollout.
const (
	jobMealPlanFinalizationStarter = "meal_plan_finalization_starter"
	jobSearchDataIndexScheduler    = "search_data_index_scheduler"
	jobMobileNotificationScheduler = "mobile_notification_scheduler"
	jobQueueTest                   = "queue_test"
	jobDataPrivacySweep            = "data_privacy_sweep"
	jobAuditRetentionSweeper       = "audit_retention_sweeper"
	jobMeteringFlusher             = "metering_flusher"
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
				name: jobMealPlanFinalizationStarter,
				cfg:  &jobsCfg.MealPlanning.MealPlanFinalizationStarter,
				// The starter reports how many sagas it began; the scheduler has nowhere to
				// put a count, and the worker already records it as a metric.
				run: func(ctx context.Context) error {
					_, workErr := do.MustInvoke[*mealplanfinalization.Starter](i).Work(ctx)
					return workErr
				},
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
			{
				name: jobDataPrivacySweep,
				cfg:  &jobsCfg.DataPrivacySweep,
				// One pass does three things: deletes the artifacts of completed exports
				// past their expiry, cancels erasures whose confirmation window lapsed,
				// and samples the overdue gauge. The sweep result reports counts the
				// Sweeper has already recorded as metrics, so there is nothing here to
				// return them to.
				//
				// Without this job every export artifact ever written stays in the bucket
				// forever, and nothing about the request rows suggests otherwise. It is
				// the failure this whole adoption is most anxious about, which is why it
				// is a named registration rather than a flag.
				run: func(ctx context.Context) error {
					_, sweepErr := do.MustInvoke[*platformdataprivacy.Sweeper](i).Sweep(ctx)

					return sweepErr
				},
			},
			{
				name: jobAuditRetentionSweeper,
				cfg:  &jobsCfg.AuditRetentionSweeper,
				// Sweep reports how many entries it pruned and logs its own failures —
				// there is no caller to return them to, and a scope that fails must not
				// stop the others. The count is already a metric.
				run: func(ctx context.Context) error {
					do.MustInvoke[*audit.Sweeper](i).Sweep(ctx)

					return nil
				},
			},
			{
				name: jobMeteringFlusher,
				cfg:  &jobsCfg.MeteringFlusher,
				// Flush reports what it posted, settled, and reaped; the scheduler has
				// nowhere to put a result, and the flusher already records all three as
				// metrics — including the backlog gauge, which is the one instrument in
				// that package worth alerting on.
				run: func(ctx context.Context) error {
					_, workErr := do.MustInvoke[*metering.Flusher](i).Flush(ctx)

					return workErr
				},
			},
		}

		for idx := range registrations {
			r := &registrations[idx]

			if !r.cfg.Enabled {
				continue
			}

			job, jobErr := r.cfg.Job(r.name, r.run)
			if jobErr != nil {
				return nil, jobErr
			}

			if err = scheduler.Register(job); err != nil {
				return nil, err
			}
		}

		return scheduler, nil
	})
}
