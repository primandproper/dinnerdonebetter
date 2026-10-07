package managers

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/clock"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	textsearchcfg "github.com/primandproper/primitives-go/v2/search/text/config"

	"github.com/samber/do/v2"
)

// RegisterManagers registers the meal planning manager with the injector.
func RegisterManagers(i do.Injector) {
	do.Provide[mealPlanFinalizationStarter](i, func(i do.Injector) (mealPlanFinalizationStarter, error) {
		return do.MustInvoke[*mealplanfinalization.Starter](i), nil
	})

	do.Provide[MealPlanningManager](i, func(i do.Injector) (MealPlanningManager, error) {
		return NewMealPlanningManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[mealplanning.Repository](i),
			identity.NewAccountRoster(do.MustInvoke[platformidentity.Store](i), do.MustInvoke[database.Client](i).Reader()),
			do.MustInvoke[clock.Clock](i),
			do.MustInvoke[recipeanalysis.RecipeAnalyzer](i),
			do.MustInvoke[*textsearchcfg.Config](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[mealPlanFinalizationStarter](i),
		)
	})
}
