package config

// This file contains config types for meal planning domain workers.
// Domain: mealplanning — remove this file when swapping the domain.

import (
	"context"
	"fmt"

	analyticscfg "github.com/primandproper/platform-go/v4/analytics/config"
	databasecfg "github.com/primandproper/platform-go/v4/database/config"
	msgconfig "github.com/primandproper/platform-go/v4/messagequeue/config"
	"github.com/primandproper/platform-go/v4/observability"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

type (
	// MealPlanFinalizerConfig configures an instance of the meal plan finalizer job.
	MealPlanFinalizerConfig struct {
		_ struct{} `json:"-"`

		Queues        msgconfig.QueuesConfig `envPrefix:"QUEUES_"        json:"queues"`
		Events        msgconfig.Config       `envPrefix:"EVENTS_"        json:"events"`
		Observability observability.Config   `envPrefix:"OBSERVABILITY_" json:"observability"`
		Database      databasecfg.Config     `envPrefix:"DATABASE_"      json:"database"`
	}

	// MealPlanGroceryListInitializerConfig configures an instance of the meal plan grocery list initializer job.
	MealPlanGroceryListInitializerConfig struct {
		_ struct{} `json:"-"`

		Queues        msgconfig.QueuesConfig `envPrefix:"QUEUES_"        json:"queues"`
		Analytics     analyticscfg.Config    `envPrefix:"ANALYTICS_"     json:"analytics"`
		Events        msgconfig.Config       `envPrefix:"EVENTS_"        json:"events"`
		Observability observability.Config   `envPrefix:"OBSERVABILITY_" json:"observability"`
		Database      databasecfg.Config     `envPrefix:"DATABASE_"      json:"database"`
	}

	// MealPlanTaskCreatorConfig configures an instance of the meal plan task creator job.
	MealPlanTaskCreatorConfig struct {
		_ struct{} `json:"-"`

		Queues        msgconfig.QueuesConfig `envPrefix:"QUEUES_"        json:"queues"`
		Analytics     analyticscfg.Config    `envPrefix:"ANALYTICS_"     json:"analytics"`
		Events        msgconfig.Config       `envPrefix:"EVENTS_"        json:"events"`
		Observability observability.Config   `envPrefix:"OBSERVABILITY_" json:"observability"`
		Database      databasecfg.Config     `envPrefix:"DATABASE_"      json:"database"`
	}
)

var _ validation.ValidatableWithContext = (*MealPlanFinalizerConfig)(nil)

// ValidateWithContext validates a MealPlanFinalizerConfig struct.
func (cfg *MealPlanFinalizerConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Observability": cfg.Observability.ValidateWithContext,
		"Database":      cfg.Database.ValidateWithContext,
		"Queues":        cfg.Queues.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*MealPlanGroceryListInitializerConfig)(nil)

// ValidateWithContext validates a MealPlanGroceryListInitializerConfig struct.
func (cfg *MealPlanGroceryListInitializerConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Analytics":     cfg.Analytics.ValidateWithContext,
		"Observability": cfg.Observability.ValidateWithContext,
		"Database":      cfg.Database.ValidateWithContext,
		"Queues":        cfg.Queues.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*MealPlanTaskCreatorConfig)(nil)

// ValidateWithContext validates a MealPlanTaskCreatorConfig struct.
func (cfg *MealPlanTaskCreatorConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Analytics":     cfg.Analytics.ValidateWithContext,
		"Observability": cfg.Observability.ValidateWithContext,
		"Database":      cfg.Database.ValidateWithContext,
		"Queues":        cfg.Queues.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}
