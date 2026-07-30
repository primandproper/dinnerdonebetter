package config

// This file contains environment config rendering for meal planning domain workers.
// Domain: mealplanning — remove this file when swapping the domain.

import "time"

// defaultMealPlanningScheduledJobsConfig returns the schedules for the meal planning domain's
// periodic jobs, which run inside the shared scheduler.
//
// The finalizer's lease is the longest of the three because its cost scales with how many meal
// plans expired since the last run: a voting deadline that lands during an outage produces a
// backlog, and the first run afterwards works through all of it.
func defaultMealPlanningScheduledJobsConfig() MealPlanningScheduledJobsConfig {
	return MealPlanningScheduledJobsConfig{
		MealPlanFinalizer: ScheduledJobConfig{
			Enabled:  true,
			Interval: 5 * time.Minute,
			Timeout:  10 * time.Minute,
			LeaseTTL: 15 * time.Minute,
		},
		MealPlanGroceryListInitializer: ScheduledJobConfig{
			Enabled:  true,
			Interval: time.Minute,
			Timeout:  2 * time.Minute,
			LeaseTTL: 5 * time.Minute,
		},
		MealPlanTaskCreator: ScheduledJobConfig{
			Enabled:  true,
			Interval: time.Minute,
			Timeout:  2 * time.Minute,
			LeaseTTL: 5 * time.Minute,
		},
	}
}
