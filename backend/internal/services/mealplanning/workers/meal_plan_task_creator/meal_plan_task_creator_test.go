package mealplantaskcreator

import (
	"context"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"

	"github.com/primandproper/platform-go/v7/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v7/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v7/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v7/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v7/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v7/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildNewMealPlanTaskCreatorForTest(t *testing.T) *Worker {
	t.Helper()

	ctx := t.Context()
	cfg := &msgconfig.QueuesConfig{DataChangesTopicName: "data_changes"}

	pp := &mockpublishers.PublisherProviderMock{
		NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{
				PublishFunc:      func(_ context.Context, _ any) error { return nil },
				PublishAsyncFunc: func(_ context.Context, _ any) {},
				StopFunc:         func() {},
			}, nil
		},
	}

	x, err := NewMealPlanTaskCreator(
		ctx,
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		&recipeanalysis.RecipeAnalyzerMock{},
		&mealplanningmock.RepositoryMock{},
		pp,
		metricsnoop.NewMetricsProvider(),
		cfg,
	)
	require.NoError(t, err)

	return x
}

func TestWorker_Work(T *testing.T) {
	T.Parallel()

	T.Run("with nothing to do", func(t *testing.T) {
		t.Parallel()

		w := buildNewMealPlanTaskCreatorForTest(t)
		assert.NotNil(t, w)

		ctx := t.Context()

		mdm := &mealplanningmock.RepositoryMock{
			GetFinalizedMealPlanIDsForTheNextWeekFunc: func(context.Context) ([]*mealplanning.FinalizedMealPlanDatabaseResult, error) {
				return []*mealplanning.FinalizedMealPlanDatabaseResult{}, nil
			},
		}
		w.dataManager = mdm

		assert.NoError(t, w.Work(ctx))

		assert.Len(t, mdm.GetFinalizedMealPlanIDsForTheNextWeekCalls(), 1)
		assert.Empty(t, mdm.CreateMealPlanTasksForMealPlanOptionCalls())
	})

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		w := buildNewMealPlanTaskCreatorForTest(t)
		assert.NotNil(t, w)

		ctx := t.Context()

		exampleMeal := fakes.BuildFakeMeal()
		exampleMealPlan := fakes.BuildFakeMealPlan()

		exampleMealPlanEvent := fakes.BuildFakeMealPlanEvent()
		exampleMealPlanEvent.BelongsToMealPlan = exampleMealPlan.ID
		now := time.Now().Add(0).Truncate(time.Second).UTC()
		inThreeDays := now.Add((time.Hour * 24) * 3).Add(0).Truncate(time.Second).UTC()
		inOneWeek := now.Add((time.Hour * 24) * 7).Add(0).Truncate(time.Second).UTC()
		exampleMealPlanEvent.StartsAt = inThreeDays
		exampleMealPlanEvent.EndsAt = inOneWeek

		exampleMealPlanOption := fakes.BuildFakeMealPlanOption()
		exampleMealPlanOption.BelongsToMealPlanEvent = exampleMealPlanEvent.ID
		exampleMealPlanOption.Meal = *exampleMeal

		recipeStepID := fakes.BuildFakeID()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipe := &mealplanning.Recipe{
			Name: "Recipe 1",
			ID:   exampleRecipeID,
			Steps: []*mealplanning.RecipeStep{
				{
					BelongsToRecipe: exampleRecipeID,
					ID:              recipeStepID,
					Preparation:     mealplanning.ValidPreparation{Name: "dice"},
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Ingredient: &mealplanning.ValidIngredient{
								MinStorageTemperatureInCelsius: new(float32(2.5)),
								PluralName:                     "chicken breasts",
								StorageInstructions:            "keep frozen",
								Name:                           "chicken breast",
								ID:                             fakes.BuildFakeID(),
							},
							Name:                "chicken breast",
							ID:                  fakes.BuildFakeID(),
							BelongsToRecipeStep: recipeStepID,
							MeasurementUnit:     mealplanning.ValidMeasurementUnit{Name: "gram", PluralName: "grams"},
							MinQuantity:         900,

							MaxQuantity: new(float32(900)),
						},
					},
					Products: []*mealplanning.RecipeStepProduct{
						{
							Name:                "diced chicken breast",
							Type:                mealplanning.RecipeStepProductIngredientType,
							BelongsToRecipeStep: recipeStepID,
							ID:                  fakes.BuildFakeID(),
							MeasurementUnit:     &mealplanning.ValidMeasurementUnit{},
						},
					},
				},
			},
		}

		recipeMap := map[string]*mealplanning.Recipe{
			exampleRecipe.ID: exampleRecipe,
		}

		exampleFinalizedMealPlanResult := &mealplanning.FinalizedMealPlanDatabaseResult{
			MealPlanID:       exampleMealPlan.ID,
			MealPlanEventID:  exampleMealPlanEvent.ID,
			MealPlanOptionID: exampleMealPlanOption.ID,
			MealID:           exampleMeal.ID,
			RecipeIDs: []string{
				exampleRecipe.ID,
			},
		}

		exampleFinalizedMealPlanResults := []*mealplanning.FinalizedMealPlanDatabaseResult{exampleFinalizedMealPlanResult}

		createdMealPlanTasks := fakes.BuildFakeMealPlanTasksList().Data

		expectedReturnResults := []*mealplanning.MealPlanTaskDatabaseCreationInput{
			{
				CreationExplanation: t.Name(),
				MealPlanOptionID:    exampleFinalizedMealPlanResult.MealPlanOptionID,
			},
		}

		mdm := &mealplanningmock.RepositoryMock{
			GetFinalizedMealPlanIDsForTheNextWeekFunc: func(context.Context) ([]*mealplanning.FinalizedMealPlanDatabaseResult, error) {
				return exampleFinalizedMealPlanResults, nil
			},
			GetRecipeFunc: func(_ context.Context, recipeID string) (*mealplanning.Recipe, error) {
				recipe, ok := recipeMap[recipeID]
				assert.True(t, ok, "unexpected recipe fetched: %s", recipeID)

				return recipe, nil
			},
			CreateMealPlanTasksForMealPlanOptionFunc: func(_ context.Context, inputs []*mealplanning.MealPlanTaskDatabaseCreationInput) ([]*mealplanning.MealPlanTask, error) {
				assert.NotNil(t, inputs)

				return createdMealPlanTasks, nil
			},
			MarkMealPlanAsHavingTasksCreatedFunc: func(_ context.Context, mealPlanID string) error {
				assert.Equal(t, exampleMealPlan.ID, mealPlanID)

				return nil
			},
		}

		mockAnalyzer := &recipeanalysis.RecipeAnalyzerMock{
			GenerateMealPlanTasksForRecipeFunc: func(_ context.Context, mealPlanOptionID string, recipe *mealplanning.Recipe) ([]*mealplanning.MealPlanTaskDatabaseCreationInput, error) {
				assert.Equal(t, exampleFinalizedMealPlanResult.MealPlanOptionID, mealPlanOptionID)
				assert.Equal(t, exampleRecipe, recipe)

				return expectedReturnResults, nil
			},
		}

		w.analyzer = mockAnalyzer
		w.dataManager = mdm

		mmp := &mockpublishers.PublisherMock{
			PublishFunc: func(_ context.Context, _ any) error { return nil },
		}
		w.postUpdatesPublisher = mmp

		assert.NoError(t, w.Work(ctx))

		assert.Len(t, mdm.GetFinalizedMealPlanIDsForTheNextWeekCalls(), 1)
		assert.Len(t, mdm.GetRecipeCalls(), 1)
		assert.Len(t, mdm.CreateMealPlanTasksForMealPlanOptionCalls(), 1)
		assert.Len(t, mdm.MarkMealPlanAsHavingTasksCreatedCalls(), 1)
		assert.Len(t, mockAnalyzer.GenerateMealPlanTasksForRecipeCalls(), 1)
	})
}
