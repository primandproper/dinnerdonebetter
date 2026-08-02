package mealplangrocerylistinitializer

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers"

	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/hashicorp/go-multierror"
)

const (
	serviceName = "meal_plan_grocery_list_initializer"
)

var _ workers.Worker = (*Worker)(nil)

type Worker struct {
	logger                  logging.Logger
	tracer                  tracing.Tracer
	dataManager             mealplanning.Repository // TODO: make this less potent
	recordsProcessedCounter metrics.Int64Counter
	groceryListCreator      grocerylistpreparation.GroceryListCreator
}

// NewMealPlanGroceryListInitializer builds the grocery list initializer.
//
// It constructs no publisher: the grocery list item created events are enqueued into the outbox
// inside the transaction that writes the items. This job used to publish them a second time after
// the repository had already emitted them transactionally, so every item announced itself twice.
func NewMealPlanGroceryListInitializer(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	dataManager mealplanning.Repository,
	groceryListCreator grocerylistpreparation.GroceryListCreator,
) (*Worker, error) {
	recordsProcessedCounter, err := metricsProvider.NewInt64Counter("meal_plan_grocery_list_initializer.records_processed")
	if err != nil {
		return nil, err
	}

	return &Worker{
		recordsProcessedCounter: recordsProcessedCounter,
		dataManager:             dataManager,
		groceryListCreator:      groceryListCreator,
		logger:                  logging.NewNamedLogger(logger, serviceName),
		tracer:                  tracing.NewNamedTracer(tracerProvider, serviceName),
	}, nil
}

func (w *Worker) Work(ctx context.Context) error {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	logger := w.logger.Clone()

	mealPlans, err := w.dataManager.GetFinalizedMealPlansWithUninitializedGroceryLists(ctx)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "getting finalized meal plan data")
	}

	logger = logger.WithValue("meal_plan_quantity", len(mealPlans))

	if len(mealPlans) > 0 {
		logger.Info("attempting to initialize grocery lists for meal plans")
	}

	errorResult := &multierror.Error{}

	for _, mealPlan := range mealPlans {
		l := logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlan.ID)

		var dbInputs []*mealplanning.MealPlanGroceryListItemDatabaseCreationInput
		dbInputs, err = w.groceryListCreator.GenerateGroceryListInputs(ctx, mealPlan)
		if err != nil {
			errorResult = multierror.Append(errorResult, err)
			l.Error("failed to generate grocery list inputs for meal plan", err)
			continue
		}

		l = l.WithValue("to_create", len(dbInputs))
		l.Info("creating grocery list items for meal plan")

		// The items, their data change events, and the flag saying the list was initialized all
		// commit together. Creating them one at a time meant a failure partway through left the
		// earlier items behind, and the retry — which regenerates the whole list with fresh IDs —
		// wrote them again.
		var createdItems []*mealplanning.MealPlanGroceryListItem
		createdItems, err = w.dataManager.InitializeMealPlanGroceryList(ctx, mealPlan.ID, mealPlan.BelongsToAccount, dbInputs)
		if err != nil {
			errorResult = multierror.Append(errorResult, err)
			l.Error("failed to initialize grocery list for meal plan", err)
			continue
		}

		w.recordsProcessedCounter.Add(ctx, int64(len(createdItems)))

		l.Info("marked meal plan as grocery list initialized")
	}

	return errorResult.ErrorOrNil()
}
