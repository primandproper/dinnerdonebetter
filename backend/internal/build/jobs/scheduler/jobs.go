package scheduler

import (
	"context"
	"errors"

	mobilenotificationscheduler "github.com/primandproper/dinnerdonebetter/backend/internal/build/jobs/mobile_notification_scheduler"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"
	queuetest "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/workers/queue_test"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"

	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	"github.com/primandproper/platform-go/v13/distributedlock"
	"github.com/primandproper/platform-go/v13/jobs"
	"github.com/primandproper/platform-go/v13/metering"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/retention"
	searchsync "github.com/primandproper/platform-go/v13/search/sync"

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
			jobs.WithSchedulerTracerProvider(do.MustInvoke[tracing.Provider](i)),
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
				// The reindex backstop. It used to be the only thing keeping the indexes
				// current — a sampler that published an index request for every row that
				// looked stale — and it is now the slow half of a pair, behind a change
				// feed that keeps up in the ordinary case.
				//
				// Every index is walked on one tick, sequentially. They are walked rather
				// than sampled, so this is proportional to the tables rather than to the
				// change rate; running them concurrently would multiply that load against
				// the same database for no gain in a job with a whole tick to finish in.
				run: runReindexers(i),
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
				// The sweep reports what each policy pruned and records its own failures;
				// a policy that fails must not stop the others, and the counts are
				// already metrics. The error is returned so a sweep that could not run
				// at all shows up as a failed job rather than a silent one.
				run: func(ctx context.Context) error {
					_, sweepErr := do.MustInvoke[*retention.Sweeper](i).Sweep(ctx)

					return sweepErr
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

// runReindexers walks every search index against its source, one after another.
//
// A failure does not stop the others: the indexes are independent, and an Algolia outage on one
// of them is no reason to leave the other eight un-rebuilt. The errors are joined so the job
// still reports as failed, with all of what went wrong rather than the first of it.
func runReindexers(i do.Injector) func(context.Context) error {
	return func(ctx context.Context) error {
		reindexers := []interface {
			Reindex(context.Context) (*searchsync.ReindexResult, error)
		}{
			do.MustInvoke[*searchsync.Reindexer[identityindexing.UserSearchSubset]](i),
			do.MustInvoke[*searchsync.Reindexer[mealplanningindexing.MealSearchSubset]](i),
			do.MustInvoke[*searchsync.Reindexer[mealplanningindexing.RecipeSearchSubset]](i),
			do.MustInvoke[*searchsync.Reindexer[mealplanningindexing.ValidIngredientSearchSubset]](i),
			do.MustInvoke[*searchsync.Reindexer[mealplanningindexing.ValidInstrumentSearchSubset]](i),
			do.MustInvoke[*searchsync.Reindexer[mealplanningindexing.ValidMeasurementUnitSearchSubset]](i),
			do.MustInvoke[*searchsync.Reindexer[mealplanningindexing.ValidPreparationSearchSubset]](i),
			do.MustInvoke[*searchsync.Reindexer[mealplanningindexing.ValidIngredientStateSearchSubset]](i),
			do.MustInvoke[*searchsync.Reindexer[mealplanningindexing.ValidVesselSearchSubset]](i),
		}

		var errs []error
		for _, reindexer := range reindexers {
			if _, err := reindexer.Reindex(ctx); err != nil {
				errs = append(errs, err)
			}
		}

		return errors.Join(errs...)
	}
}
