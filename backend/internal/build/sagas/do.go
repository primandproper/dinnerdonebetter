// Package sagas wires the durable saga machinery: the registry of definitions this build can
// run, the store they live in, the lifecycle event publisher, and the typed runners callers
// start them through.
//
// Every process that starts or advances a saga registers all of it. A Runner refuses a
// definition name it has not seen and a Worker marks an instance stuck rather than guessing at
// one, so a process holding a partially-populated registry is one that fails on the instances it
// cannot see rather than one that quietly does less.
//
// The Worker itself is registered separately — see RegisterSagaWorker — because advancing is
// background work and belongs in the process that does background work, not in the one serving
// requests.
package sagas

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"

	"github.com/primandproper/platform-go/v12/database"
	"github.com/primandproper/platform-go/v12/distributedlock"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/metrics"
	"github.com/primandproper/platform-go/v12/observability/tracing"
	"github.com/primandproper/platform-go/v12/outbox"
	"github.com/primandproper/platform-go/v12/saga"

	"github.com/samber/do/v2"
)

// RegisterSagas registers the saga registry, store, event publisher, and runners.
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

	do.Provide[saga.Store](i, func(i do.Injector) (saga.Store, error) {
		return saga.NewSQLStore(
			do.MustInvoke[database.Client](i),
			saga.WithStoreLogger(do.MustInvoke[logging.Logger](i)),
			saga.WithStoreTracerProvider(do.MustInvoke[tracing.Provider](i)),
			saga.WithStoreMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
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
			do.MustInvoke[saga.Store](i),
			do.MustInvoke[*saga.Registry](i),
			saga.WithRunnerEventPublisher(do.MustInvoke[saga.EventPublisher](i)),
			saga.WithRunnerLogger(do.MustInvoke[logging.Logger](i)),
			saga.WithRunnerTracerProvider(do.MustInvoke[tracing.Provider](i)),
			saga.WithRunnerMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

// RegisterSagaWorker registers the worker that advances every saga in the process.
//
// It is built without an idempotency manager, and that is a decision rather than an omission.
// The manager suppresses a step whose result was recorded but whose instance row did not catch
// up, and it does so from a store that commits separately from the step — so for a step that
// writes to this database it is a weaker guarantee than the step already has. Meal plan
// finalization's steps each write their work and the flag saying they did it in one transaction,
// and re-read that flag before doing anything; a step that reached out to something that cannot
// join a transaction would need the manager, and there is not one yet.
func RegisterSagaWorker(i do.Injector) {
	do.Provide[*saga.Worker](i, func(i do.Injector) (*saga.Worker, error) {
		// The per-instance lock, distinct from the per-job lock the scheduler holds: that one
		// decides which replica runs a tick, this one stops two workers stepping through the
		// same instance while a lease lapses.
		locker, err := distributedlock.NewScopedLocker(do.MustInvoke[distributedlock.Locker](i))
		if err != nil {
			return nil, err
		}

		return saga.NewWorker(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[*saga.WorkerConfig](i),
			do.MustInvoke[saga.Store](i),
			do.MustInvoke[*saga.Registry](i),
			locker,
			saga.WithWorkerEventPublisher(do.MustInvoke[saga.EventPublisher](i)),
			saga.WithWorkerLogger(do.MustInvoke[logging.Logger](i)),
			saga.WithWorkerTracerProvider(do.MustInvoke[tracing.Provider](i)),
			saga.WithWorkerMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
