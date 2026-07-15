package mealplantaskcreator

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers"

	"github.com/primandproper/platform-go/v4/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v4/messagequeue/config"
	"github.com/primandproper/platform-go/v4/observability"
	"github.com/primandproper/platform-go/v4/observability/logging"
	"github.com/primandproper/platform-go/v4/observability/metrics"
	"github.com/primandproper/platform-go/v4/observability/tracing"

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
	postUpdatesPublisher    messagequeue.Publisher
	processedRecordsCounter metrics.Int64Counter
}

func NewMealPlanTaskCreator(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	analyzer recipeanalysis.RecipeAnalyzer,
	dataManager mealplanning.Repository,
	publisherProvider messagequeue.PublisherProvider,
	metricsProvider metrics.Provider,
	cfg *msgconfig.QueuesConfig,
) (*Worker, error) {
	postUpdatesPublisher, err := publisherProvider.NewPublisher(ctx, cfg.DataChangesTopicName)
	if err != nil {
		return nil, err
	}

	processedRecordsCounter, err := metricsProvider.NewInt64Counter("meal_plan_task_creator.records_processed")
	if err != nil {
		return nil, err
	}

	return &Worker{
		analyzer:                analyzer,
		dataManager:             dataManager,
		postUpdatesPublisher:    postUpdatesPublisher,
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

		createdMealPlanTasks, creationErr := w.dataManager.CreateMealPlanTasksForMealPlanOption(ctx, steps)
		if creationErr != nil {
			result = multierror.Append(result, creationErr)
			observability.AcknowledgeError(creationErr, l, span, "creating meal plan tasks for meal plan option")
			continue
		}

		w.processedRecordsCounter.Add(ctx, int64(len(createdMealPlanTasks)))

		for _, createdTask := range createdMealPlanTasks {
			if publishErr := w.postUpdatesPublisher.Publish(ctx, &audit.DataChangeMessage{
				EventType: mealplanning.MealPlanTaskCreatedServiceEventType,
				Context: map[string]any{
					mealplanningkeys.MealPlanIDKey:     mealPlanID,
					mealplanningkeys.MealPlanTaskIDKey: createdTask.ID,
					"meal_plan_task":                   createdTask,
				},
			}); publishErr != nil {
				observability.AcknowledgeError(publishErr, l, span, "publishing data change event")
			}
		}

		if err = w.dataManager.MarkMealPlanAsHavingTasksCreated(ctx, mealPlanID); err != nil {
			result = multierror.Append(result, err)
		}
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
