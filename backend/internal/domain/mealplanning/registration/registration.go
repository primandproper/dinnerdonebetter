// Package registration provides consolidated DI registration functions for the mealplanning domain.
// Domain: mealplanning — remove this package when swapping the domain.
package registration

import (
	grocerylistpreparation "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	mealplanningmgr "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers"
	recipeanalysis "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc"
	eatingindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"

	"github.com/samber/do/v2"
)

func registerRepository(i do.Injector) {
	// The repository writes its data change events into the outbox as part of the same
	// transaction as the row they describe; see internal/repositories/postgres/events.
	events.RegisterOutboxEmitter(i)
	mealplanningrepo.RegisterMealPlanningRepository(i)
}

// RegisterForGRPCAPI registers all mealplanning components needed by the gRPC API server.
func RegisterForGRPCAPI(i do.Injector) {
	registerRepository(i)
	mealplanningmgr.RegisterManagers(i)
	mealplanningsvc.RegisterMealPlanningService(i)
	// The API only ever starts a finalization saga — the admin RPC that used to run the three
	// pipeline jobs on demand now runs this. Advancing belongs to the scheduler's saga worker,
	// which is what keeps a durable process from being tied to the lifetime of a request.
	mealplanfinalization.RegisterStarter(i)
	recipeanalysis.RegisterRecipeAnalyzer(i)
	grocerylistpreparation.RegisterGroceryListCreator(i)
}

// RegisterForDataChangeHandler registers mealplanning components needed by the async message
// handler, which consumes the index topics and applies their events through the Syncers.
func RegisterForDataChangeHandler(i do.Injector) {
	registerRepository(i)
	eatingindexing.RegisterIndexSyncers(i)
}

// RegisterForSearchIndexScheduler registers mealplanning components needed by the scheduler,
// which drives the Reindexers. They come from the same registration as the Syncers: the two are
// halves of keeping one index right, and the process resolves whichever half it runs.
func RegisterForSearchIndexScheduler(i do.Injector) {
	registerRepository(i)
	eatingindexing.RegisterIndexSyncers(i)
}
