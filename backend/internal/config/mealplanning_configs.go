package config

// This file contains config types for meal planning domain workers.
// Domain: mealplanning — remove this file when swapping the domain.

import (
	"context"
	"fmt"

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
		MealPlanFinalizationStarter ScheduledJobConfig `envPrefix:"MEAL_PLAN_FINALIZATION_STARTER_" json:"mealPlanFinalizationStarter"`
	}
)

var _ validation.ValidatableWithContext = (*MealPlanningScheduledJobsConfig)(nil)

// ValidateWithContext validates a MealPlanningScheduledJobsConfig struct.
func (cfg *MealPlanningScheduledJobsConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"MealPlanFinalizationStarter": cfg.MealPlanFinalizationStarter.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}
