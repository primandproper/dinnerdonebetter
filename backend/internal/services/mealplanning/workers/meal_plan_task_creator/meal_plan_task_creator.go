package mealplantaskcreator

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers"

	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/hashicorp/go-multierror"
)

const (
	serviceName = "meal_plan_task_creator"
)

var _ workers.Worker = (*Worker)(nil)

type Worker struct {
	logger                  logging.Logger
	tracer                  tracing.Tracer
	analyzer                recipeanalysis.RecipeAnalyzer
	dataManager             mealplanning.Repository
	processedRecordsCounter metrics.Int64Counter
}

// NewMealPlanTaskCreator builds the task creator.
//
// It constructs no publisher: the meal plan task created events are enqueued into the outbox inside
// the transaction that writes the tasks, so there is nothing left for this job to publish.
func NewMealPlanTaskCreator(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	analyzer recipeanalysis.RecipeAnalyzer,
	dataManager mealplanning.Repository,
	metricsProvider metrics.Provider,
) (*Worker, error) {
	processedRecordsCounter, err := metricsProvider.NewInt64Counter("meal_plan_task_creator.records_processed")
	if err != nil {
		return nil, err
	}

	return &Worker{
		analyzer:                analyzer,
		dataManager:             dataManager,
		processedRecordsCounter: processedRecordsCounter,
		logger:                  logging.NewNamedLogger(logger, serviceName),
		tracer:                  tracing.NewNamedTracer(tracerProvider, serviceName),
	}, nil
}

func (w *Worker) Work(ctx context.Context) error {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	logger := w.logger.Clone()

	mealPlansAndSteps, err := w.determineCreatableMealPlanTasks(ctx)
	if err != nil {
		return observability.PrepareError(err, nil, "determining creatable steps")
	}

	logger = logger.WithValue("creatable_steps_qty", len(mealPlansAndSteps))

	result := &multierror.Error{}
	for mealPlanID, steps := range mealPlansAndSteps {
		l := logger.Clone().WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID).WithValue("creatable_prep_step_qty", len(steps))

		// The tasks, their data change events, and the flag saying the plan has had its tasks
		// created all commit together. A plan that fails here is left untouched and retried whole
		// on the next run, rather than retried against tasks that already exist.
		createdMealPlanTasks, creationErr := w.dataManager.CreateMealPlanTasksForMealPlan(ctx, mealPlanID, steps)
		if creationErr != nil {
			result = multierror.Append(result, creationErr)
			observability.AcknowledgeError(creationErr, l, span, "creating meal plan tasks for meal plan")
			continue
		}

		w.processedRecordsCounter.Add(ctx, int64(len(createdMealPlanTasks)))
	}

	return result.ErrorOrNil()
}

// determineCreatableMealPlanTasks determines which meal plan tasks are creatable for a recipe.
func (w *Worker) determineCreatableMealPlanTasks(ctx context.Context) (map[string][]*mealplanning.MealPlanTaskDatabaseCreationInput, error) {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	logger := w.logger.Clone()
	logger.Info("fetching finalized meal plan IDs to determine creatable steps")

	results, err := w.dataManager.GetFinalizedMealPlanIDsForTheNextWeek(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting finalized meal plan data for the next week")
	}

	if len(results) > 0 {
		logger = logger.WithValue("steps_to_create", len(results))
		logger.Info("determining creatable steps")
	}

	inputs := map[string][]*mealplanning.MealPlanTaskDatabaseCreationInput{}
	for _, result := range results {
		l := logger.Clone().WithValues(map[string]any{
			mealplanningkeys.MealPlanIDKey:       result.MealPlanID,
			mealplanningkeys.MealPlanEventIDKey:  result.MealPlanEventID,
			mealplanningkeys.MealPlanOptionIDKey: result.MealPlanOptionID,
			mealplanningkeys.MealIDKey:           result.MealID,
			"recipe_ids":                         result.RecipeIDs,
		})
		l.Info("fetching meal plan event")

		if _, ok := inputs[result.MealPlanID]; !ok {
			inputs[result.MealPlanID] = []*mealplanning.MealPlanTaskDatabaseCreationInput{}
		}

		for _, recipeID := range result.RecipeIDs {
			recipe, getRecipeErr := w.dataManager.GetRecipe(ctx, recipeID)
			if getRecipeErr != nil {
				return nil, observability.PrepareAndLogError(getRecipeErr, l, span, "fetching recipe")
			}

			creatableSteps, determineStepsErr := w.analyzer.GenerateMealPlanTasksForRecipe(ctx, result.MealPlanOptionID, recipe)
			if determineStepsErr != nil {
				return nil, observability.PrepareAndLogError(determineStepsErr, l, span, "fetching recipe")
			}

			inputs[result.MealPlanID] = append(inputs[result.MealPlanID], creatableSteps...)
		}
	}

	return inputs, nil
}
