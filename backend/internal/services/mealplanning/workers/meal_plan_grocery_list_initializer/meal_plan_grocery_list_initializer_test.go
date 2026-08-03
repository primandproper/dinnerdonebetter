package mealplangrocerylistinitializer

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	grocerylistpreparation2 "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v9/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildNewMealPlanGroceryListInitializerForTest(t *testing.T) *Worker {
	t.Helper()

	x, err := NewMealPlanGroceryListInitializer(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		&mealplanningmock.RepositoryMock{},
		grocerylistpreparation2.NewGroceryListCreator(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider()),
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
				// the account is the ordering key for the events the repository emits, and a
				// background job has no session to read it from — the worker has to pass it.
				BelongsToAccount: fakes.BuildFakeID(),
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
			InitializeMealPlanGroceryListFunc: func(_ context.Context, mealPlanID, accountID string, inputs []*mealplanning.MealPlanGroceryListItemDatabaseCreationInput) ([]*mealplanning.MealPlanGroceryListItem, error) {
				assert.Equal(t, expectedMealPlans[0].ID, mealPlanID)
				assert.Equal(t, expectedMealPlans[0].BelongsToAccount, accountID)
				// every generated input must be handed over exactly as the grocery list creator produced it.
				assert.Equal(t, firstMealPlanExpectedGroceryListItemInputs, inputs)

				created := make([]*mealplanning.MealPlanGroceryListItem, len(inputs))
				for i := range inputs {
					created[i] = fakes.BuildFakeMealPlanGroceryListItem()
				}

				return created, nil
			},
		}

		w.dataManager = mdm

		assert.NoError(t, w.Work(ctx))

		assert.Len(t, mdm.GetFinalizedMealPlansWithUninitializedGroceryListsCalls(), 1)
		assert.Len(t, mdm.InitializeMealPlanGroceryListCalls(), 1)
		assert.Len(t, mglm.GenerateGroceryListInputsCalls(), 1)
	})
}
