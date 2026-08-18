package managers

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"

	"github.com/primandproper/platform-go/v11/messagequeue"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/metrics"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	textsearchcfg "github.com/primandproper/platform-go/v11/search/text/config"

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
			do.MustInvoke[*queuescfg.Config](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
			do.MustInvoke[recipeanalysis.RecipeAnalyzer](i),
			do.MustInvoke[*textsearchcfg.Config](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[mealPlanFinalizationStarter](i),
		)
	})
}
