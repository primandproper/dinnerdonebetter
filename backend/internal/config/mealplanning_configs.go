package config

// This file contains config types for meal planning domain workers.
// Domain: mealplanning — remove this file when swapping the domain.

import (
	"context"
	"fmt"

	"github.com/primandproper/platform-go/v13/workqueue"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

type (
	// MealPlanningScheduledJobsConfig carries the schedules for the meal planning domain's
	// periodic jobs. They run inside the shared jobs.Scheduler; see SchedulerConfig.
	MealPlanningScheduledJobsConfig struct {
		_ struct{} `json:"-"`

		// MealPlanFinalizationStarter replaces the three jobs that used to poll for expired
		// plans, for finalized plans without tasks, and for finalized plans without a grocery
		// list. It only claims plans and writes a saga instance for each; the saga worker does
		// the pipeline, so this interval is the delay before a plan enters the pipeline rather
		// than the delay before it comes out the other end.
		MealPlanFinalizationStarter ScheduledJobConfig `envPrefix:"MEAL_PLAN_FINALIZATION_STARTER_" json:"mealPlanFinalizationStarter,omitzero"`

		// MealPlanTaskNotifications sends each prep task's reminder push. It replaces the
		// job that published one message per unnotified task onto the mobile notifications
		// topic and let the consumer decide when the task had been dealt with; this one
		// claims tasks from MealPlanTaskNotificationQueue and sends under the lease, so
		// the queue's idea of done and the database's are the same fact.
		//
		// It sits with the meal planning jobs rather than among the generic ones because
		// prep task reminders are the only thing it has ever sent. The topic it used to
		// publish to still exists and still carries notifications from other domains.
		MealPlanTaskNotifications ScheduledJobConfig `envPrefix:"MEAL_PLAN_TASK_NOTIFICATIONS_" json:"mealPlanTaskNotifications,omitzero"`

		// MealPlanTaskNotificationQueue is the leased queue that job fills and drains.
		//
		// Its Name is left unset and filled in at construction: one table holds every
		// logical queue, partitioned by name, and workqueue.Config deliberately has no
		// default for it, so an unnamed queue would share rows with every other unnamed
		// one. What is worth setting here is MaxAttempts — the ceiling that stops a task
		// nobody can be notified for from being retried forever.
		MealPlanTaskNotificationQueue workqueue.Config `envPrefix:"MEAL_PLAN_TASK_NOTIFICATION_QUEUE_" json:"mealPlanTaskNotificationQueue,omitzero"`
	}
)

var _ validation.ValidatableWithContext = (*MealPlanningScheduledJobsConfig)(nil)

// ValidateWithContext validates a MealPlanningScheduledJobsConfig struct.
func (cfg *MealPlanningScheduledJobsConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	// MealPlanTaskNotificationQueue is deliberately not among these. Its Name is filled in at
	// construction and its remaining knobs are defaulted there, so validating it here would
	// reject every configuration that had sensibly left both alone. workqueue.New defaults and
	// validates it once the name is set.
	validators := map[string]func(context.Context) error{
		"MealPlanFinalizationStarter": cfg.MealPlanFinalizationStarter.ValidateWithContext,
		"MealPlanTaskNotifications":   cfg.MealPlanTaskNotifications.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}
