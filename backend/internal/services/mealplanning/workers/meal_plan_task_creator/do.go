package mealplantaskcreator

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterMealPlanTaskCreator registers the meal plan task creator with the injector.
func RegisterMealPlanTaskCreator(i do.Injector) {
	do.Provide[*Worker](i, func(i do.Injector) (*Worker, error) {
		return NewMealPlanTaskCreator(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[recipeanalysis.RecipeAnalyzer](i),
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})
}
