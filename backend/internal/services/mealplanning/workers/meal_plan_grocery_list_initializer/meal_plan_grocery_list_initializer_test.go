package mealplangrocerylistinitializer

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	grocerylistpreparation2 "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v9/messagequeue"
	mockpublishers "github.com/primandproper/platform-go/v9/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v9/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildNewMealPlanGroceryListInitializerForTest(t *testing.T) *Worker {
	t.Helper()

	ctx := t.Context()
	cfg := &queuescfg.Config{DataChangesTopicName: "data_changes"}

	pp := &mockpublishers.PublisherProviderMock{
		NewPublisherFunc: func(_ context.Context, topic string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{
				PublishFunc:      func(_ context.Context, _ any) error { return nil },
				PublishAsyncFunc: func(_ context.Context, _ any) {},
				StopFunc:         func() {},
			}, nil
		},
	}

	x, err := NewMealPlanGroceryListInitializer(
		ctx,
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		pp,
		&mealplanningmock.RepositoryMock{},
		grocerylistpreparation2.NewGroceryListCreator(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider()),
		cfg,
	)
	require.NoError(t, err)

	return x
}

func TestMealPlanGroceryListInitializer_HandleMessage(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		w := buildNewMealPlanGroceryListInitializerForTest(t)
		assert.NotNil(t, w)

		onion := fakes.BuildFakeValidIngredient()
		carrot := fakes.BuildFakeValidIngredient()
		celery := fakes.BuildFakeValidIngredient()
		salt := fakes.BuildFakeValidIngredient()

		grams := fakes.BuildFakeValidMeasurementUnit()

		expectedMealPlans := []*mealplanning.MealPlan{
			{
				ID: fakes.BuildFakeID(),
				Events: []*mealplanning.MealPlanEvent{
					{
						Options: []*mealplanning.MealPlanOption{
							{
								Chosen: true,
								Meal: mealplanning.Meal{
									Components: []*mealplanning.MealComponent{
										{
											Recipe: mealplanning.Recipe{
												Steps: []*mealplanning.RecipeStep{
													{
														Ingredients: []*mealplanning.RecipeStepIngredient{
															{
																Ingredient:  onion,
																MinQuantity: 100,

																MaxQuantity:     new(float32(100)),
																MeasurementUnit: *grams,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						Options: []*mealplanning.MealPlanOption{
							{
								Chosen: true,
								Meal: mealplanning.Meal{
									Components: []*mealplanning.MealComponent{
										{
											Recipe: mealplanning.Recipe{
												Steps: []*mealplanning.RecipeStep{
													{
														Ingredients: []*mealplanning.RecipeStepIngredient{
															{
																Ingredient:  carrot,
																MinQuantity: 100,

																MaxQuantity:     new(float32(100)),
																MeasurementUnit: *grams,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						Options: []*mealplanning.MealPlanOption{
							{
								Chosen: true,
								Meal: mealplanning.Meal{
									Components: []*mealplanning.MealComponent{
										{
											Recipe: mealplanning.Recipe{
												Steps: []*mealplanning.RecipeStep{
													{
														Ingredients: []*mealplanning.RecipeStepIngredient{
															{
																Ingredient:  celery,
																MinQuantity: 100,

																MaxQuantity:     new(float32(100)),
																MeasurementUnit: *grams,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						Options: []*mealplanning.MealPlanOption{
							{
								Chosen: true,
								Meal: mealplanning.Meal{
									Components: []*mealplanning.MealComponent{
										{
											Recipe: mealplanning.Recipe{
												Steps: []*mealplanning.RecipeStep{
													{
														Ingredients: []*mealplanning.RecipeStepIngredient{
															{
																Ingredient:  salt,
																MinQuantity: 100,

																MaxQuantity:     new(float32(100)),
																MeasurementUnit: *grams,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						Options: []*mealplanning.MealPlanOption{
							{
								Chosen: true,
								Meal: mealplanning.Meal{
									Components: []*mealplanning.MealComponent{
										{
											Recipe: mealplanning.Recipe{
												Steps: []*mealplanning.RecipeStep{
													{
														Ingredients: []*mealplanning.RecipeStepIngredient{
															{
																Ingredient:  onion,
																MinQuantity: 100,

																MaxQuantity:     new(float32(100)),
																MeasurementUnit: *grams,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		ctx := t.Context()

		firstMealPlanExpectedGroceryListItemInputs := []*mealplanning.MealPlanGroceryListItemDatabaseCreationInput{
			{
				Status:                 mealplanning.MealPlanGroceryListItemStatusUnknown,
				ValidMeasurementUnitID: grams.ID,
				ValidIngredientID:      onion.ID,
				BelongsToMealPlan:      expectedMealPlans[0].ID,
				MinQuantityNeeded:      200,

				MaxQuantityNeeded: new(float32(200)),
			},
			{
				Status:                 mealplanning.MealPlanGroceryListItemStatusUnknown,
				ValidMeasurementUnitID: grams.ID,
				ValidIngredientID:      carrot.ID,
				BelongsToMealPlan:      expectedMealPlans[0].ID,
				MinQuantityNeeded:      100,

				MaxQuantityNeeded: new(float32(100)),
			},
			{
				Status:                 mealplanning.MealPlanGroceryListItemStatusUnknown,
				ValidMeasurementUnitID: grams.ID,
				ValidIngredientID:      celery.ID,
				BelongsToMealPlan:      expectedMealPlans[0].ID,
				MinQuantityNeeded:      100,

				MaxQuantityNeeded: new(float32(100)),
			},
			{
				Status:                 mealplanning.MealPlanGroceryListItemStatusUnknown,
				ValidMeasurementUnitID: grams.ID,
				ValidIngredientID:      salt.ID,
				BelongsToMealPlan:      expectedMealPlans[0].ID,
				MinQuantityNeeded:      100,

				MaxQuantityNeeded: new(float32(100)),
			},
		}

		// every generated input must be persisted exactly as the grocery list creator produced it.
		expectedInputs := map[*mealplanning.MealPlanGroceryListItemDatabaseCreationInput]bool{}
		for _, input := range firstMealPlanExpectedGroceryListItemInputs {
			expectedInputs[input] = true
		}

		mglm := &grocerylistpreparation2.GroceryListCreatorMock{
			GenerateGroceryListInputsFunc: func(_ context.Context, mealPlan *mealplanning.MealPlan) ([]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput, error) {
				assert.Equal(t, expectedMealPlans[0], mealPlan)

				return firstMealPlanExpectedGroceryListItemInputs, nil
			},
		}
		w.groceryListCreator = mglm

		mdm := &mealplanningmock.RepositoryMock{
			GetFinalizedMealPlansWithUninitializedGroceryListsFunc: func(context.Context) ([]*mealplanning.MealPlan, error) {
				return expectedMealPlans, nil
			},
			CreateMealPlanGroceryListItemFunc: func(_ context.Context, input *mealplanning.MealPlanGroceryListItemDatabaseCreationInput) (*mealplanning.MealPlanGroceryListItem, error) {
				assert.True(t, expectedInputs[input], "unexpected grocery list item input persisted")

				return fakes.BuildFakeMealPlanGroceryListItem(), nil
			},
			MarkMealPlanAsGroceryListInitializedFunc: func(_ context.Context, mealPlanID string) error {
				assert.Equal(t, expectedMealPlans[0].ID, mealPlanID)

				return nil
			},
		}

		pup := &mockpublishers.PublisherMock{
			PublishFunc: func(_ context.Context, _ any) error { return nil },
		}

		w.postUpdatesPublisher = pup
		w.dataManager = mdm

		assert.NoError(t, w.Work(ctx))

		assert.Len(t, mdm.GetFinalizedMealPlansWithUninitializedGroceryListsCalls(), 1)
		assert.Len(t, mdm.CreateMealPlanGroceryListItemCalls(), len(firstMealPlanExpectedGroceryListItemInputs))
		assert.Len(t, mdm.MarkMealPlanAsGroceryListInitializedCalls(), 1)
		assert.Len(t, mglm.GenerateGroceryListInputsCalls(), 1)
	})
}
