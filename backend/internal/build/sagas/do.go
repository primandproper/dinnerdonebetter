// Package sagas wires the durable saga machinery: the registry of definitions this build can
// run, the store they live in, the lifecycle event publisher, and the typed runners callers
// start them through.
//
// Every process that starts or advances a saga registers all of it. A Runner refuses a
// definition name it has not seen and a Worker marks an instance stuck rather than guessing at
// one, so a process holding a partially-populated registry is one that fails on the instances it
// cannot see rather than one that quietly does less.
//
// The Worker is not registered here at all. Advancing is background work, and the scheduler
// process that does it is composed from a service.Config whose Saga block registers platform's
// worker, its store, and the retention job that prunes finished instances. That process registers
// RegisterSagas beside it; the API server, which starts sagas and never advances them, registers
// RegisterSagaStore as well, since it has no Saga block to build one from.
package sagas

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"

	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/platform-go/v15/saga"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterSagas registers the saga registry, event publisher, and runners.
//
// Prerequisites: a saga.Store — platform's, from a service.Config Saga block, or RegisterSagaStore.
func RegisterSagas(i do.Injector) {
	do.Provide[*saga.Registry](i, func(i do.Injector) (*saga.Registry, error) {
		registry := saga.NewRegistry()

		// Domain: mealplanning — swapping the domain replaces this block and nothing else in
		// this file.
		if err := mealplanfinalization.Register(
			registry,
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[recipeanalysis.RecipeAnalyzer](i),
			do.MustInvoke[grocerylistpreparation.GroceryListCreator](i),
			do.MustInvoke[logging.Logger](i),
		); err != nil {
			return nil, err
		}

		return registry, nil
	})

	// Lifecycle events go into the outbox this process already runs, in the transaction that
	// records the transition they describe. That is the difference between a subscriber that
	// can always read the instance an event names and one that sometimes cannot.
	do.Provide[saga.EventPublisher](i, func(i do.Injector) (saga.EventPublisher, error) {
		return saga.NewOutboxPublisher(do.MustInvoke[*outbox.Writer](i))
	})

	// Domain: mealplanning — one runner per state type, over the one shared store.
	do.Provide[saga.Runner[mealplanning.MealPlanFinalizationState]](i, func(i do.Injector) (saga.Runner[mealplanning.MealPlanFinalizationState], error) {
		return saga.NewRunner[mealplanning.MealPlanFinalizationState](
			do.MustInvoke[database.Client](i),
			do.MustInvoke[saga.Store](i),
			do.MustInvoke[*saga.Registry](i),
			saga.WithRunnerEventPublisher(do.MustInvoke[saga.EventPublisher](i)),
			saga.WithRunnerLogger(do.MustInvoke[logging.Logger](i)),
			saga.WithRunnerTracerProvider(do.MustInvoke[tracing.Provider](i)),
			saga.WithRunnerMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

// RegisterSagaStore registers the store saga instances live in, for a process with no service.Config
// Saga block to build platform's from. It is the same SQL store over the same table, with the
// package's default prefix, which is what the migrations render it under.
func RegisterSagaStore(i do.Injector) {
	do.Provide[saga.Store](i, func(i do.Injector) (saga.Store, error) {
		return saga.NewSQLStore(
			do.MustInvoke[database.Client](i),
			saga.WithStoreLogger(do.MustInvoke[logging.Logger](i)),
			saga.WithStoreTracerProvider(do.MustInvoke[tracing.Provider](i)),
			saga.WithStoreMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
