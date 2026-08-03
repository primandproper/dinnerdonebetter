package mealplanfinalization

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	"github.com/primandproper/platform-go/v9/saga"

	"github.com/samber/do/v2"
)

// Register adds the meal plan finalization definition to a saga registry.
//
// It takes the registry rather than the injector because a registry is only usable once every
// definition the process can run is on it: a Runner refuses a name it has not seen, and a Worker
// marks an instance stuck rather than guessing at one. Registering during the registry's own
// construction is what makes that an invariant instead of a wiring order somebody has to
// remember.
func Register(
	registry *saga.Registry,
	dataManager mealplanning.Repository,
	analyzer recipeanalysis.RecipeAnalyzer,
	groceryListCreator grocerylistpreparation.GroceryListCreator,
	logger logging.Logger,
) error {
	return saga.Register(registry, saga.Definition[mealplanning.MealPlanFinalizationState]{
		Name:  mealplanning.MealPlanFinalizationSagaName,
		Steps: steps(dataManager, analyzer, groceryListCreator, logger),
	})
}

// RegisterStarter registers the meal plan finalization saga starter with the injector.
func RegisterStarter(i do.Injector) {
	do.Provide[*Starter](i, func(i do.Injector) (*Starter, error) {
		return NewStarter(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[saga.Runner[mealplanning.MealPlanFinalizationState]](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})
}
