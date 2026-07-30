package indexing

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterMealPlanningDataIndexer registers the meal planning data indexer with the injector.
func RegisterMealPlanningDataIndexer(i do.Injector) {
	do.Provide[*MealPlanningDataIndexer](i, func(i do.Injector) (*MealPlanningDataIndexer, error) {
		return NewMealPlanningDataIndexer(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[RecipeTextSearcher](i),
			do.MustInvoke[MealTextSearcher](i),
			do.MustInvoke[ValidIngredientTextSearcher](i),
			do.MustInvoke[ValidInstrumentTextSearcher](i),
			do.MustInvoke[ValidMeasurementUnitTextSearcher](i),
			do.MustInvoke[ValidPreparationTextSearcher](i),
			do.MustInvoke[ValidIngredientStateTextSearcher](i),
			do.MustInvoke[ValidVesselTextSearcher](i),
		), nil
	})
}
