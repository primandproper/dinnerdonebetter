// Package registration provides consolidated DI registration functions for the mealplanning domain.
// Domain: mealplanning — remove this package when swapping the domain.
package registration

import (
	domaindataprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	grocerylistpreparation "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	mealplanningmgr "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers"
	mealplanningprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/privacy"
	recipeanalysis "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc"
	eatingindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

func registerRepository(i do.Injector) {
	// The repository writes its data change events into the outbox as part of the same
	// transaction as the row they describe; see internal/repositories/postgres/events.
	events.RegisterOutboxEmitter(i)
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
	// The API only ever starts a finalization saga — the admin RPC that used to run the three
	// pipeline jobs on demand now runs this. Advancing belongs to the scheduler's saga worker,
	// which is what keeps a durable process from being tied to the lifetime of a request.
	mealplanfinalization.RegisterStarter(i)
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
