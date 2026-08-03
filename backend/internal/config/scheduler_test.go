package config

import (
	"context"
	"math"
	"testing"
	"time"
	// Embedded so these tests resolve America/Chicago on a host without the zoneinfo
	// database, the same way cmd/workers/scheduler does.
	_ "time/tzdata"

	"github.com/primandproper/platform-go/v9/distributedlock/noop"
	"github.com/primandproper/platform-go/v9/jobs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduledJobConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &ScheduledJobConfig{
			Enabled:  true,
			Interval: time.Minute,
			Timeout:  time.Minute,
			LeaseTTL: 5 * time.Minute,
		}

		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("with cron schedule", func(t *testing.T) {
		t.Parallel()

		cfg := &ScheduledJobConfig{
			Enabled:  true,
			Schedule: "0 3 * * *",
			Timeout:  time.Minute,
			LeaseTTL: 5 * time.Minute,
		}

		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("with cron schedule naming its own timezone", func(t *testing.T) {
		t.Parallel()

		cfg := &ScheduledJobConfig{
			Enabled:  true,
			Schedule: "CRON_TZ=America/Chicago 0 8-21 * * *",
			Timeout:  time.Minute,
			LeaseTTL: 5 * time.Minute,
		}

		assert.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("with disabled job", func(t *testing.T) {
		t.Parallel()

		// A disabled job is never registered, so an empty schedule is not an error.
		assert.NoError(t, (&ScheduledJobConfig{}).ValidateWithContext(t.Context()))
	})

	T.Run("with enabled job setting neither an interval nor a schedule", func(t *testing.T) {
		t.Parallel()

		cfg := &ScheduledJobConfig{
			Enabled:  true,
			LeaseTTL: time.Minute,
		}

		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("with enabled job setting both an interval and a schedule", func(t *testing.T) {
		t.Parallel()

		// The scheduler rejects this at Register rather than picking one, so config must
		// not be able to express it either.
		cfg := &ScheduledJobConfig{
			Enabled:  true,
			Interval: time.Minute,
			Schedule: "0 3 * * *",
			LeaseTTL: 5 * time.Minute,
		}

		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("with an unparseable cron schedule", func(t *testing.T) {
		t.Parallel()

		cfg := &ScheduledJobConfig{
			Enabled:  true,
			Schedule: "not a cron expression",
			LeaseTTL: 5 * time.Minute,
		}

		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("with enabled job missing a lease TTL", func(t *testing.T) {
		t.Parallel()

		cfg := &ScheduledJobConfig{
			Enabled:  true,
			Interval: time.Minute,
		}

		assert.Error(t, cfg.ValidateWithContext(t.Context()))
	})
}

func TestScheduledJobConfig_Job(T *testing.T) {
	T.Parallel()

	noopRun := func(context.Context) error { return nil }

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &ScheduledJobConfig{
			Interval:   time.Minute,
			Timeout:    30 * time.Second,
			LeaseTTL:   5 * time.Minute,
			RunOnStart: true,
		}

		job, err := cfg.Job("example", noopRun)

		require.NoError(t, err)
		assert.Equal(t, "example", job.Name)
		assert.Equal(t, time.Minute, job.Interval)
		assert.Equal(t, 30*time.Second, job.Timeout)
		assert.Equal(t, 5*time.Minute, job.LeaseTTL)
		assert.True(t, job.RunOnStart)
		// jobs.Job rejects a job that sets both, so an interval-shaped job must leave the
		// Schedule field nil rather than carrying an interface wrapping a zero value.
		assert.Nil(t, job.Schedule)
	})

	T.Run("with a cron schedule", func(t *testing.T) {
		t.Parallel()

		job, err := (&ScheduledJobConfig{Schedule: "0 3 * * *"}).Job("example", noopRun)

		require.NoError(t, err)
		require.NotNil(t, job.Schedule)
		assert.Zero(t, job.Interval)

		// Midnight UTC on a Tuesday, so the next fire is 03:00 the same day.
		assert.Equal(t,
			time.Date(2026, 1, 6, 3, 0, 0, 0, time.UTC),
			job.Schedule.Next(time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)).UTC(),
		)
	})

	T.Run("with a schedule naming its own timezone", func(t *testing.T) {
		t.Parallel()

		// A CRON_TZ prefix beats the scheduler's own Timezone, which is how one job opts
		// into a calendar the rest do not share.
		job, err := (&ScheduledJobConfig{Schedule: "CRON_TZ=America/Chicago 0 8 * * *"}).Job("example", noopRun)

		require.NoError(t, err)
		require.NotNil(t, job.Schedule)

		// 08:00 Chicago on a January Tuesday is 14:00 UTC, Central being six hours behind
		// outside daylight saving.
		assert.Equal(t,
			time.Date(2026, 1, 6, 14, 0, 0, 0, time.UTC),
			job.Schedule.Next(time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)).UTC(),
		)
	})

	T.Run("with an unparseable schedule", func(t *testing.T) {
		t.Parallel()

		job, err := (&ScheduledJobConfig{Schedule: "0 3 * *"}).Job("example", noopRun)

		assert.Error(t, err)
		assert.Zero(t, job)
	})
}

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

			job, err := cfg.Job(name, nil)
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
		cfg := defaultScheduledJobsConfig()
		cfg.Scheduler.EnsureDefaults()

		scheduler, err := jobs.NewScheduler(t.Context(), &cfg.Scheduler, noop.NewLocker())
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
func enabledDefaultJobs() map[string]ScheduledJobConfig {
	cfg := defaultScheduledJobsConfig()

	all := map[string]ScheduledJobConfig{
		"search_data_index_scheduler":    cfg.SearchDataIndexScheduler,
		"mobile_notification_scheduler":  cfg.MobileNotificationScheduler,
		"queue_test":                     cfg.QueueTest,
		"meal_plan_finalization_starter": cfg.MealPlanning.MealPlanFinalizationStarter,
	}

	for name := range all {
		if !all[name].Enabled {
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
