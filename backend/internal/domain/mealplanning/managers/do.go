package managers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	mealplangrocerylistinitializer "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_grocery_list_initializer"
	mealplantaskcreator "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_creator"

	"github.com/primandproper/platform-go/v7/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v7/messagequeue/config"
	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/metrics"
	"github.com/primandproper/platform-go/v7/observability/tracing"
	textsearchcfg "github.com/primandproper/platform-go/v7/search/text/config"

	"github.com/samber/do/v2"
)

// RegisterManagers registers the meal planning manager with the injector.
func RegisterManagers(i do.Injector) {
	do.Provide[mealPlanTaskCreatorWorker](i, func(i do.Injector) (mealPlanTaskCreatorWorker, error) {
		return BuildMealPlanTaskCreatorWorker(do.MustInvoke[*mealplantaskcreator.Worker](i)), nil
	})

	do.Provide[mealPlanGroceryListInitializerWorker](i, func(i do.Injector) (mealPlanGroceryListInitializerWorker, error) {
		return BuildMealPlanGroceryListInitializerWorker(do.MustInvoke[*mealplangrocerylistinitializer.Worker](i)), nil
	})

	do.Provide[MealPlanningManager](i, func(i do.Injector) (MealPlanningManager, error) {
		return NewMealPlanningManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[*msgconfig.QueuesConfig](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
			do.MustInvoke[recipeanalysis.RecipeAnalyzer](i),
			do.MustInvoke[*textsearchcfg.Config](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[mealPlanGroceryListInitializerWorker](i),
			do.MustInvoke[mealPlanTaskCreatorWorker](i),
		)
	})
}

func BuildMealPlanTaskCreatorWorker(x *mealplantaskcreator.Worker) mealPlanTaskCreatorWorker {
	return x
}

func BuildMealPlanGroceryListInitializerWorker(x *mealplangrocerylistinitializer.Worker) mealPlanGroceryListInitializerWorker {
	return x
}
