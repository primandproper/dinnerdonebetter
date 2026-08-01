package config

import (
	"testing"
	"time"

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

	T.Run("with disabled job", func(t *testing.T) {
		t.Parallel()

		// A disabled job is never registered, so an empty schedule is not an error.
		assert.NoError(t, (&ScheduledJobConfig{}).ValidateWithContext(t.Context()))
	})

	T.Run("with enabled job missing an interval", func(t *testing.T) {
		t.Parallel()

		cfg := &ScheduledJobConfig{
			Enabled:  true,
			LeaseTTL: time.Minute,
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

func TestDefaultScheduledJobsConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := defaultScheduledJobsConfig()

		require.NoError(t, cfg.ValidateWithContext(t.Context()))
	})

	T.Run("every enabled job leases for longer than it may run", func(t *testing.T) {
		t.Parallel()

		cfg := defaultScheduledJobsConfig()

		// The lease is not renewed while a job runs. A Timeout at or above LeaseTTL means a
		// slow job can still be running when its lease lapses, at which point a second
		// replica may start the same job — the failure jobs_scheduler_leases_expired counts.
		for name, job := range map[string]ScheduledJobConfig{
			"search_data_index_scheduler":        cfg.SearchDataIndexScheduler,
			"mobile_notification_scheduler":      cfg.MobileNotificationScheduler,
			"queue_test":                         cfg.QueueTest,
			"meal_plan_finalizer":                cfg.MealPlanning.MealPlanFinalizer,
			"meal_plan_grocery_list_initializer": cfg.MealPlanning.MealPlanGroceryListInitializer,
			"meal_plan_task_creator":             cfg.MealPlanning.MealPlanTaskCreator,
		} {
			if !job.Enabled {
				continue
			}

			assert.Positive(t, job.Timeout, "%s has no timeout, so it can run until its lease lapses", name)
			assert.Greater(t, job.LeaseTTL, job.Timeout, "%s can outlive its lease", name)
		}
	})
}
