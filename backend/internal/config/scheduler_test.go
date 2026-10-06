package config

import (
	"context"
	"math"
	"testing"
	"time"
	// Embedded so these tests resolve America/Chicago on a host without the zoneinfo
	// database, the same way cmd/workers/scheduler does.
	_ "time/tzdata"

	"github.com/primandproper/primitives-go/v2/distributedlock/noop"
	"github.com/primandproper/primitives-go/v2/jobs"
	jobscfg "github.com/primandproper/primitives-go/v2/jobs/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultScheduledJobsConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := defaultScheduledJobsConfig()

		require.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("every enabled job leases for longer than it may run", func(t *testing.T) {
		t.Parallel()

		// The lease is not renewed while a job runs. A Timeout at or above LeaseTTL means a
		// slow job can still be running when its lease lapses, at which point a second
		// replica may start the same job — the failure jobs_scheduler_leases_expired counts.
		for name, job := range enabledDefaultJobs() {
			assert.Positive(t, job.Timeout, "%s has no timeout, so it can run until its lease lapses", name)
			assert.Greater(t, job.LeaseTTL, job.Timeout, "%s can outlive its lease", name)
		}
	})

	T.Run("every enabled cron job finishes before its next fire", func(t *testing.T) {
		t.Parallel()

		// A calendar's headroom varies — a job at "0 9 * * 1-5" has three days of it on
		// Friday night and one on Monday — so the bound that matters is the tightest gap
		// the expression ever produces, not the average one. A Timeout above that gap means
		// the job is still running when the scheduler wants to start it again, which the
		// scheduler reports as an overrun and skips.
		for name, cfg := range enabledDefaultJobs() {
			if cfg.Schedule == "" {
				continue
			}

			job, err := cfg.Job(name, func(context.Context) error { return nil })
			require.NoError(t, err, "%s has an unparseable schedule", name)

			assert.Greater(t, smallestGap(job.Schedule), cfg.Timeout, "%s can still be running when it is next due", name)
		}
	})
}

func TestDefaultScheduledJobsConfig_registersWithTheScheduler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		// The invariants above are ours; these are the scheduler's, and it enforces them at
		// Register — a job that sets both an interval and a schedule, one that sets neither,
		// one whose expression will never come true ("0 0 30 2 *" parses cleanly), and two
		// jobs sharing a name. Without this, all four are a crash loop at rollout rather
		// than a failure here.
		schedulerCfg := defaultJobsSchedulerConfig()
		schedulerCfg.EnsureDefaults()

		scheduler, err := jobs.NewScheduler(t.Context(), &schedulerCfg.Scheduler, noop.NewLocker())
		require.NoError(t, err)

		noopRun := func(context.Context) error { return nil }

		for name, jobCfg := range enabledDefaultJobs() {
			job, jobErr := jobCfg.Job(name, noopRun)
			require.NoError(t, jobErr, "building %s", name)

			assert.NoError(t, scheduler.Register(job), "registering %s", name)
		}
	})
}

// enabledDefaultJobs returns every job defaultScheduledJobsConfig enables, by name. The domain
// jobs are listed alongside the rest because the invariants their callers assert hold for any
// job the scheduler runs, whatever domain it came from.
func enabledDefaultJobs() map[string]jobscfg.JobConfig {
	cfg := defaultScheduledJobsConfig()

	all := map[string]jobscfg.JobConfig{
		"search_data_index_scheduler":    cfg.SearchDataIndexScheduler,
		"queue_test":                     cfg.QueueTest,
		"meal_plan_finalization_starter": cfg.MealPlanning.MealPlanFinalizationStarter,
		"meal_plan_task_notifications":   cfg.MealPlanning.MealPlanTaskNotifications,
		"data_privacy_sweep":             cfg.DataPrivacySweep,
		"audit_retention_sweeper":        cfg.AuditRetentionSweeper,
		"metering_flusher":               cfg.MeteringFlusher,
	}

	for name := range all {
		if all[name].Disabled {
			delete(all, name)
		}
	}

	return all
}

// smallestGap reports the shortest interval between consecutive fires of a schedule over a full
// year, which covers the tightest gap of any expression that repeats on a daily, weekly, monthly,
// or annual cycle. A year also spans both daylight saving transitions, where a wall-clock schedule
// in a zone that observes them stretches one gap in spring — the fires an hour apart on either
// side of the missing hour land two hours apart in real time.
func smallestGap(schedule jobs.Schedule) time.Duration {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	// Measured from the first fire rather than from start, so that the distance between an
	// arbitrary instant and the fire after it is not mistaken for a gap between two fires.
	previous := schedule.Next(start)

	smallest := time.Duration(math.MaxInt64)
	for next := schedule.Next(previous); !next.IsZero() && next.Before(end); next = schedule.Next(previous) {
		if gap := next.Sub(previous); gap < smallest {
			smallest = gap
		}
		previous = next
	}

	return smallest
}
