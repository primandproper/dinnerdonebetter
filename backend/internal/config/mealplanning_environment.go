package config

// This file contains environment config rendering for meal planning domain workers.
// Domain: mealplanning — remove this file when swapping the domain.

import (
	"time"

	"github.com/primandproper/platform-go/v13/workqueue"
)

// defaultMealPlanningScheduledJobsConfig returns the schedules for the meal planning domain's
// periodic jobs, which run inside the shared scheduler.
//
// A minute, where the finalizer this replaces ran every five. It can afford to: the job writes
// one row per plan and does none of the pipeline work, so its cost scales with how many plans
// came due rather than with what has to happen to them. The saga worker polls every second and
// picks each plan up as soon as its instance lands, which makes this interval the whole of the
// pipeline's scheduling latency instead of the first of three.
func defaultMealPlanningScheduledJobsConfig() MealPlanningScheduledJobsConfig {
	return MealPlanningScheduledJobsConfig{
		MealPlanFinalizationStarter: ScheduledJobConfig{
			Enabled:  true,
			Interval: time.Minute,
			Timeout:  2 * time.Minute,
			LeaseTTL: 5 * time.Minute,
		},
		// Push notifications, so the hours are the point: this fires on the hour from 08:00
		// to 21:00 US Central and never overnight. It carries its own zone because it is the
		// one job here whose correctness is a fact about people rather than about load, and
		// because the scheduler's own default is deliberately UTC.
		//
		// Hourly rather than once in the morning: a task created after the day's last fire
		// waits for the next one, and is dropped entirely if its event starts first — so the
		// gap between fires bounds how much short notice the app can give.
		//
		// Per-user timezones are the real answer here and a much larger conversation; one
		// zone's waking hours are strictly better than every two minutes in the meantime.
		MealPlanTaskNotifications: ScheduledJobConfig{
			Enabled:  true,
			Schedule: "CRON_TZ=America/Chicago 0 8-21 * * *",
			// One pass now sends the pushes as well as finding the work, and it drains
			// the queue in batches until it is empty, so the bound is an hour's backlog
			// of tasks times however many devices their recipients have registered.
			Timeout: 10 * time.Minute,
			// Comfortably past the timeout: the lock is not renewed while the job runs,
			// and a second replica starting this job while the first is still pushing is
			// the one way to send a reminder twice.
			LeaseTTL: 20 * time.Minute,
		},
		MealPlanTaskNotificationQueue: defaultMealPlanTaskNotificationQueueConfig(),
	}
}

// defaultMealPlanTaskNotificationQueueConfig returns the queue the notification worker fills and
// drains.
//
// MaxAttempts is the setting that matters and the reason the queue is here at all. Five claims
// is generous for a push that is going to work and short enough that a task nobody can be
// notified for — an account that no longer exists, a context row that went missing — stops being
// retried and starts being counted, instead of failing on every tick until its event starts.
// Stalled items are not deleted, so the keys stay readable and an operator can re-enqueue them.
//
// Retention is a day, which is long enough that a completed task's row is still there to explain
// why nothing was sent again, and short enough that the table stays small.
//
// No NotifyChannel: this queue is drained on a schedule with hours between passes, so waking a
// claimer the instant a row lands would buy nothing. The listener would be pure machinery.
func defaultMealPlanTaskNotificationQueueConfig() workqueue.Config {
	return workqueue.Config{
		MaxAttempts:   5,
		MaxClaimBatch: 100,
		Retention:     24 * time.Hour,
	}
}
