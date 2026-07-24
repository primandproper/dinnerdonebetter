package main

import (
	"context"
	"fmt"
	"log"

	mealplangrocerylistinitializerbuild "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/jobs/meal_plan_grocery_list_initializer"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/build/telemetry"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	mealplangrocerylistinitializer "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_grocery_list_initializer"

	"github.com/samber/do/v2"
	_ "go.uber.org/automaxprocs"
)

func doTheThing(ctx context.Context) error {
	config.ConditionallyCease()

	cfg, err := config.LoadConfigFromEnvironment[config.MealPlanGroceryListInitializerConfig]()
	if err != nil {
		return fmt.Errorf("error getting config: %w", err)
	}
	cfg.Database.RunMigrations = false

	i := mealplangrocerylistinitializerbuild.BuildInjector(ctx, cfg)

	// Flush telemetry on exit so this short-lived CronJob pod exports its spans/metrics before it exits.
	defer telemetry.Flush(ctx, i)

	worker := do.MustInvoke[*mealplangrocerylistinitializer.Worker](i)

	if err = worker.Work(ctx); err != nil {
		return fmt.Errorf("error initializing meal plan grocery lists: %w", err)
	}

	return nil
}

func main() {
	if err := doTheThing(context.Background()); err != nil {
		log.Fatal(err)
	}
}
