package scheduler

import (
	"context"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"
	queuetest "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/workers/queue_test"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"
	mealplantasknotifications "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_notifications"

	platformdataprivacy "github.com/primandproper/platform-go/v15/dataprivacy"
	"github.com/primandproper/platform-go/v15/metering"
	"github.com/primandproper/platform-go/v15/retention"
	searchsync "github.com/primandproper/platform-go/v15/searchsync"
	"github.com/primandproper/primitives-go/v2/jobs"
	jobscfg "github.com/primandproper/primitives-go/v2/jobs/config"

	"github.com/samber/do/v2"
)

// Job names. These are also the distributed lock keys, so renaming one lets an old replica and
// a new replica both run that job during a rollout.
const (
	jobMealPlanFinalizationStarter = "meal_plan_finalization_starter"
	jobMealPlanTaskNotifications   = "meal_plan_task_notifications"
	jobSearchDataIndexScheduler    = "search_data_index_scheduler"
	jobQueueTest                   = "queue_test"
	jobDataPrivacySweep            = "data_privacy_sweep"
	jobAuditRetentionSweeper       = "audit_retention_sweeper"
	jobMeteringFlusher             = "metering_flusher"
)

// RegisterJobs registers this application's scheduled jobs, every enabled one already rendered, as
// the []jobs.Job service.New hands to the scheduler.
//
// The scheduler is not built here. platform builds it from the service.Config JobsScheduler block,
// with the lock that keeps each run to one replica, and New registers these on it together with
// the jobs platform schedules for itself — operations' recovery and reap, and saga retention — in
// one call, so a duplicate name or an invalid job anywhere fails the boot rather than leaving a
// schedule that is partly what was asked for.
func RegisterJobs(i do.Injector) {
	do.Provide[[]jobs.Job](i, func(i do.Injector) ([]jobs.Job, error) {
		jobsCfg := do.MustInvoke[*config.ScheduledJobsConfig](i)

		registrations := []struct {
			run  func(ctx context.Context) error
			cfg  *jobscfg.JobConfig
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
				name: jobMealPlanTaskNotifications,
				cfg:  &jobsCfg.MealPlanning.MealPlanTaskNotifications,
				// One pass enqueues every task still owed a reminder, drains what the
				// queue hands over, and sends under the lease. The count of pushes it
				// sent has nowhere to go here — the queue and the fan-out both record
				// their own counters — so it is dropped the way the finalization
				// starter's is.
				run: func(ctx context.Context) error {
					_, workErr := do.MustInvoke[*mealplantasknotifications.Worker](i).Work(ctx)

					return workErr
				},
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

		var scheduled []jobs.Job

		for idx := range registrations {
			r := &registrations[idx]

			if r.cfg.Disabled {
				continue
			}

			job, err := r.cfg.Job(r.name, r.run)
			if err != nil {
				return nil, err
			}

			scheduled = append(scheduled, job)
		}

		return scheduled, nil
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
