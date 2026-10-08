package scheduler

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	mealplanningregistration "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/registration"
	"github.com/primandproper/dinnerdonebetter/backend/internal/searchindexes"
	queuetest "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/workers/queue_test"

	platformdataprivacy "github.com/primandproper/platform-go/v15/dataprivacy"
	"github.com/primandproper/platform-go/v15/metering"
	"github.com/primandproper/platform-go/v15/retention"
	"github.com/primandproper/primitives-go/v2/jobs"
	jobscfg "github.com/primandproper/primitives-go/v2/jobs/config"

	"github.com/samber/do/v2"
)

// Job names. These are also the distributed lock keys, so renaming one lets an old replica and
// a new replica both run that job during a rollout.
const (
	jobSearchDataIndexScheduler = "search_data_index_scheduler"
	jobQueueTest                = "queue_test"
	jobDataPrivacySweep         = "data_privacy_sweep"
	jobAuditRetentionSweeper    = "audit_retention_sweeper"
	jobMeteringFlusher          = "metering_flusher"
)

// RegisterJobs registers this application's scheduled jobs, every enabled one already rendered, as
// the []jobs.Job service.New hands to the scheduler.
//
// The scheduler is not built here. platform builds it from the service.Config JobsScheduler block,
// with the lock that keeps each run to one replica, and New registers these on it together with
// the jobs platform schedules for itself — operations' recovery and reap, and saga retention — in
// one call, so a duplicate name or an invalid job anywhere fails the boot rather than leaving a
// schedule that is partly what was asked for.
//
// The jobs are this application's own, which are rendered here, followed by each domain's,
// which the domain renders from its own block of the config and contributes already built.
func RegisterJobs(i do.Injector) {
	do.Provide[[]jobs.Job](i, func(i do.Injector) ([]jobs.Job, error) {
		jobsCfg := do.MustInvoke[*config.ScheduledJobsConfig](i)

		registrations := []struct {
			run  func(ctx context.Context) error
			cfg  *jobscfg.JobConfig
			name string
		}{
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

		for _, contribute := range []func(do.Injector) ([]jobs.Job, error){
			// Domain: mealplanning
			mealplanningregistration.ScheduledJobs,
		} {
			contributed, err := contribute(i)
			if err != nil {
				return nil, err
			}

			scheduled = append(scheduled, contributed...)
		}

		return scheduled, nil
	})
}

// runReindexers walks every search index against its source, one after another.
//
// A failure does not stop the others: the indexes are independent, and an Algolia outage on one
// of them is no reason to leave the other eight un-rebuilt. ReindexAll joins the errors, so the
// job still reports as failed, with all of what went wrong rather than the first of it — and it
// walks the Registry rather than a list kept here, so an index added to it is rebuilt without
// this changing.
func runReindexers(i do.Injector) func(context.Context) error {
	return func(ctx context.Context) error {
		_, err := do.MustInvoke[*searchindexes.Registry](i).ReindexAll(ctx)

		return err
	}
}
