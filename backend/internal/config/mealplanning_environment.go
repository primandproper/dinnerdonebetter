package config

// This file contains environment config rendering for meal planning domain workers.
// Domain: mealplanning — remove this file when swapping the domain.

import "time"

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
	}
}
