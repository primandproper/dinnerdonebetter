package mealplangrocerylistinitializer

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	grocerylistpreparation2 "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v3/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v3/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v3/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v3/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v3/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v3/observability/tracing/noop"
	"github.com/primandproper/platform-go/v3/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildNewMealPlanGroceryListInitializerForTest(t *testing.T) *Worker {
	t.Helper()

	ctx := t.Context()
	cfg := &msgconfig.QueuesConfig{DataChangesTopicName: "data_changes"}

	pp := &mockpublishers.PublisherProviderMock{
		ProvidePublisherFunc: func(_ context.Context, topic string) (messagequeue.Publisher, error) {
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
		&mealplanningmock.Repository{},
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
		mdm := &mealplanningmock.Repository{}

		mdm.On(reflection.GetMethodName(mdm.GetFinalizedMealPlansWithUninitializedGroceryLists), testutils.ContextMatcher).Return(expectedMealPlans, nil)

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

		expectedInputSets := map[string][]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput{
			expectedMealPlans[0].ID: firstMealPlanExpectedGroceryListItemInputs,
		}

		mglm := &grocerylistpreparation2.MockGroceryListCreator{}
		mglm.On(
			"GenerateGroceryListInputs",
			testutils.ContextMatcher,
			expectedMealPlans[0],
		).Return(firstMealPlanExpectedGroceryListItemInputs, nil)
		w.groceryListCreator = mglm

		pup := &mockpublishers.PublisherMock{
			PublishFunc: func(_ context.Context, _ any) error { return nil },
		}
		for _, inputs := range expectedInputSets {
			for _, input := range inputs {
				mdm.On(reflection.GetMethodName(mdm.CreateMealPlanGroceryListItem), testutils.ContextMatcher, input).Return(fakes.BuildFakeMealPlanGroceryListItem(), nil)
			}
		}

		mdm.On(reflection.GetMethodName(mdm.MarkMealPlanAsGroceryListInitialized), testutils.ContextMatcher, expectedMealPlans[0].ID).Return(nil)

		w.postUpdatesPublisher = pup
		w.dataManager = mdm

		assert.NoError(t, w.Work(ctx))
		mock.AssertExpectationsForObjects(t, mdm)
	})
}
