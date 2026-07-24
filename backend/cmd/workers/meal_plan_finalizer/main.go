package main

import (
	"context"
	"fmt"
	"log"

	mealplanfinalizerbuild "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/jobs/meal_plan_finalizer"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/telemetry"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	mealplanfinalizer "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalizer"

	"github.com/samber/do/v2"
	_ "go.uber.org/automaxprocs"
)

func doTheThing(ctx context.Context) error {
	config.ConditionallyCease()

	cfg, err := config.LoadConfigFromEnvironment[config.MealPlanFinalizerConfig]()
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}
	cfg.Database.RunMigrations = false

	i := mealplanfinalizerbuild.BuildInjector(ctx, cfg)

	// Flush telemetry on exit so this short-lived CronJob pod exports its spans/metrics before it exits.
	defer telemetry.Flush(ctx, i)

	worker := do.MustInvoke[*mealplanfinalizer.Worker](i)

	if _, err = worker.Work(ctx); err != nil {
		return fmt.Errorf("error finalizing meal plans: %w", err)
	}

	return nil
}

func main() {
	if err := doTheThing(context.Background()); err != nil {
		log.Fatal(err)
	}
}
