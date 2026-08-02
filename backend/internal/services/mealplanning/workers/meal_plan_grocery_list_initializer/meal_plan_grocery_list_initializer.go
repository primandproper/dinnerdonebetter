package mealplangrocerylistinitializer

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	queuescfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/queues/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers"

	"github.com/primandproper/platform-go/v9/messagequeue"
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
	postUpdatesPublisher    messagequeue.Publisher
	recordsProcessedCounter metrics.Int64Counter
	groceryListCreator      grocerylistpreparation.GroceryListCreator
}

func NewMealPlanGroceryListInitializer(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	publisherProvider messagequeue.PublisherProvider,
	dataManager mealplanning.Repository,
	groceryListCreator grocerylistpreparation.GroceryListCreator,
	cfg *queuescfg.Config,
) (*Worker, error) {
	postUpdatesPublisher, err := publisherProvider.NewPublisher(ctx, cfg.DataChangesTopicName)
	if err != nil {
		return nil, err
	}

	recordsProcessedCounter, err := metricsProvider.NewInt64Counter("meal_plan_grocery_list_initializer.records_processed")
	if err != nil {
		return nil, err
	}

	return &Worker{
		recordsProcessedCounter: recordsProcessedCounter,
		postUpdatesPublisher:    postUpdatesPublisher,
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

		var createdCount int64
		for _, dbInput := range dbInputs {
			var createdItem *mealplanning.MealPlanGroceryListItem
			createdItem, err = w.dataManager.CreateMealPlanGroceryListItem(ctx, dbInput)
			if err != nil {
				errorResult = multierror.Append(errorResult, err)
				l.Error("failed to create grocery list for meal plan", err)
				continue
			}
			createdCount++

			if err = w.postUpdatesPublisher.Publish(ctx, &audit.DataChangeMessage{
				EventType: mealplanning.MealPlanGroceryListItemCreatedServiceEventType,
				Context: map[string]any{
					"groceryListItem": createdItem,
					mealplanningkeys.MealPlanGroceryListItemIDKey: createdItem.ID,
					mealplanningkeys.MealPlanIDKey:                dbInput.BelongsToMealPlan,
				},
			}); err != nil {
				l.Error("failed to write update message for meal plan grocery list item", err)
			}
		}
		w.recordsProcessedCounter.Add(ctx, createdCount)

		// Only mark the plan initialized once every item was created. If some items failed, leave the plan
		// unmarked so it is retried on a later run instead of being permanently missing items.
		if createdCount != int64(len(dbInputs)) {
			l.WithValue("created", createdCount).Info("not marking meal plan as grocery list initialized because some items failed to create")
			continue
		}

		if err = w.dataManager.MarkMealPlanAsGroceryListInitialized(ctx, mealPlan.ID); err != nil {
			errorResult = multierror.Append(errorResult, err)
			l.Error("failed to mark meal plan as grocery list initialized", err)
			continue
		}

		l.Info("marked meal plan as grocery list initialized")
	}

	return errorResult.ErrorOrNil()
}
