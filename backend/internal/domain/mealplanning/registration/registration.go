// Package registration provides consolidated DI registration functions for the mealplanning domain.
// Domain: mealplanning — remove this package when swapping the domain.
package registration

import (
	domaindataprivacy "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	grocerylistpreparation "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	mealplanningmgr "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/managers"
	mealplanningprivacy "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/privacy"
	recipeanalysis "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	mealplanningrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	mealplanningsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/grpc"
	eatingindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/indexing"
	mealplanfinalizer "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalizer"
	mealplangrocerylistinitializer "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_grocery_list_initializer"
	mealplantaskcreator "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_creator"

	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/tracing"

	"github.com/samber/do/v2"
)

func registerRepository(i do.Injector) {
	mealplanningrepo.RegisterMealPlanningRepository(i)
}

func registerDataPrivacyCollector(i do.Injector) {
	do.Provide[[]domaindataprivacy.UserDataCollector](i, func(i do.Injector) ([]domaindataprivacy.UserDataCollector, error) {
		return []domaindataprivacy.UserDataCollector{
			mealplanningprivacy.NewCollector(
				do.MustInvoke[mealplanning.Repository](i),
				do.MustInvoke[logging.Logger](i),
				do.MustInvoke[tracing.TracerProvider](i),
			),
		}, nil
	})
}

// RegisterForGRPCAPI registers all mealplanning components needed by the gRPC API server.
func RegisterForGRPCAPI(i do.Injector) {
	registerRepository(i)
	registerDataPrivacyCollector(i)
	mealplanningmgr.RegisterManagers(i)
	mealplanningsvc.RegisterMealPlanningService(i)
	mealplanfinalizer.RegisterMealPlanFinalizer(i)
	mealplangrocerylistinitializer.RegisterMealPlanGroceryListInitializer(i)
	mealplantaskcreator.RegisterMealPlanTaskCreator(i)
	recipeanalysis.RegisterRecipeAnalyzer(i)
	grocerylistpreparation.RegisterGroceryListCreator(i)
}

// RegisterForDataChangeHandler registers mealplanning components needed by the async message handler.
func RegisterForDataChangeHandler(i do.Injector) {
	registerRepository(i)
	registerDataPrivacyCollector(i)
	eatingindexing.RegisterMealPlanningDataIndexer(i)
}

// RegisterForSearchIndexScheduler registers mealplanning components needed by the search index scheduler.
func RegisterForSearchIndexScheduler(i do.Injector) {
	registerRepository(i)
	eatingindexing.RegisterMealPlanningDataIndexer(i)
}
