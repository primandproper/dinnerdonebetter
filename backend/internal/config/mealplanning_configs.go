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

		MealPlanFinalizer              ScheduledJobConfig `envPrefix:"MEAL_PLAN_FINALIZER_"                json:"mealPlanFinalizer"`
		MealPlanGroceryListInitializer ScheduledJobConfig `envPrefix:"MEAL_PLAN_GROCERY_LIST_INITIALIZER_" json:"mealPlanGroceryListInitializer"`
		MealPlanTaskCreator            ScheduledJobConfig `envPrefix:"MEAL_PLAN_TASK_CREATOR_"             json:"mealPlanTaskCreator"`
	}
)

var _ validation.ValidatableWithContext = (*MealPlanningScheduledJobsConfig)(nil)

// ValidateWithContext validates a MealPlanningScheduledJobsConfig struct.
func (cfg *MealPlanningScheduledJobsConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"MealPlanFinalizer":              cfg.MealPlanFinalizer.ValidateWithContext,
		"MealPlanGroceryListInitializer": cfg.MealPlanGroceryListInitializer.ValidateWithContext,
		"MealPlanTaskCreator":            cfg.MealPlanTaskCreator.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}
