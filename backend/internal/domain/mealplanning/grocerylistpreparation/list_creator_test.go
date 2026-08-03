package grocerylistpreparation

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"

	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// quantityEpsilon is the relative tolerance for asserting on generated grocery list
// quantities. Those quantities are aggregated across recipes and scaled by portion
// count before being rounded to a tenth, so the arithmetic can land a few ULPs off an
// exactly-representable expectation. Because the final value is quantized to a tenth,
// any genuinely wrong quantity differs by at least 0.1 — several orders of magnitude
// more than this epsilon — so a tolerance this tight still catches real regressions
// while ignoring float32 drift (~1.2e-7 relative).
const quantityEpsilon = 1e-6

func Test_groceryListCreator_GenerateGroceryListInputs(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		listGenerator := &groceryListCreator{
			logger: loggingnoop.NewLogger(),
			tracer: tracing.NewTracerForTest(t.Name()),
		}

		onion := fakes.BuildFakeValidIngredient()
		carrot := fakes.BuildFakeValidIngredient()
		celery := fakes.BuildFakeValidIngredient()
		salt := fakes.BuildFakeValidIngredient()
		grams := fakes.BuildFakeValidMeasurementUnit()

		// Set up IDs for options, recipes, and steps
		option1ID := fakes.BuildFakeID()
		option2ID := fakes.BuildFakeID()
		option3ID := fakes.BuildFakeID()
		option4ID := fakes.BuildFakeID()
		option5ID := fakes.BuildFakeID()
		recipe1ID := fakes.BuildFakeID()
		recipe2ID := fakes.BuildFakeID()
		recipe3ID := fakes.BuildFakeID()
		recipe4ID := fakes.BuildFakeID()
		recipe5ID := fakes.BuildFakeID()
		step1ID := fakes.BuildFakeID()
		step2ID := fakes.BuildFakeID()
		step3ID := fakes.BuildFakeID()
		step4ID := fakes.BuildFakeID()
		step5ID := fakes.BuildFakeID()

		expectedMealPlan := &mealplanning.MealPlan{
			ID: fakes.BuildFakeID(),
			Events: []*mealplanning.MealPlanEvent{
				{
					Options: []*mealplanning.MealPlanOption{
						{
							ID:        option1ID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 1.0,
										Recipe: mealplanning.Recipe{
											ID: recipe1ID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: step1ID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														{
															Ingredient:  onion,
															MinQuantity: 100,

															MaxQuantity:     new(float32(100)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
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
							ID:        option2ID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 1.0,
										Recipe: mealplanning.Recipe{
											ID: recipe2ID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: step2ID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														{
															Ingredient:  carrot,
															MinQuantity: 100,

															MaxQuantity:     new(float32(100)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
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
							ID:        option3ID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 1.0,
										Recipe: mealplanning.Recipe{
											ID: recipe3ID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: step3ID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														{
															Ingredient:  celery,
															MinQuantity: 100,

															MaxQuantity:     new(float32(100)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
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
							ID:        option4ID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 1.0,
										Recipe: mealplanning.Recipe{
											ID: recipe4ID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: step4ID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														{
															Ingredient:  salt,
															MinQuantity: 100,

															MaxQuantity:     new(float32(100)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
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
							ID:        option5ID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 1.0,
										Recipe: mealplanning.Recipe{
											ID: recipe5ID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: step5ID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														{
															Ingredient:  onion,
															MinQuantity: 100,

															MaxQuantity:     new(float32(100)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
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

		expectedMap := map[string]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput{
			onion.ID: {
				Status:                 mealplanning.MealPlanGroceryListItemStatusNeeds,
				ValidMeasurementUnitID: grams.ID,
				ValidIngredientID:      onion.ID,
				BelongsToMealPlan:      expectedMealPlan.ID,
				// consolidated across option1/recipe1/step1 and option5/recipe5/step5, so no single source attribution
				BelongsToMealPlanOption: nil,
				RecipeID:                nil,
				RecipeStepID:            nil,
				MinQuantityNeeded:       200,

				MaxQuantityNeeded: new(float32(200)),
			},
			carrot.ID: {
				Status:                  mealplanning.MealPlanGroceryListItemStatusNeeds,
				ValidMeasurementUnitID:  grams.ID,
				ValidIngredientID:       carrot.ID,
				BelongsToMealPlan:       expectedMealPlan.ID,
				BelongsToMealPlanOption: &option2ID,
				RecipeID:                &recipe2ID,
				RecipeStepID:            &step2ID,
				MinQuantityNeeded:       100,

				MaxQuantityNeeded: new(float32(100)),
			},
			celery.ID: {
				Status:                  mealplanning.MealPlanGroceryListItemStatusNeeds,
				ValidMeasurementUnitID:  grams.ID,
				ValidIngredientID:       celery.ID,
				BelongsToMealPlan:       expectedMealPlan.ID,
				BelongsToMealPlanOption: &option3ID,
				RecipeID:                &recipe3ID,
				RecipeStepID:            &step3ID,
				MinQuantityNeeded:       100,

				MaxQuantityNeeded: new(float32(100)),
			},
			salt.ID: {
				Status:                  mealplanning.MealPlanGroceryListItemStatusNeeds,
				ValidMeasurementUnitID:  grams.ID,
				ValidIngredientID:       salt.ID,
				BelongsToMealPlan:       expectedMealPlan.ID,
				BelongsToMealPlanOption: &option4ID,
				RecipeID:                &recipe4ID,
				RecipeStepID:            &step4ID,
				MinQuantityNeeded:       100,

				MaxQuantityNeeded: new(float32(100)),
			},
		}

		actual, err := listGenerator.GenerateGroceryListInputs(ctx, expectedMealPlan)
		require.NoError(t, err)

		actualMap := map[string]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput{}
		for i := range actual {
			actualMap[actual[i].ValidIngredientID] = actual[i]
			expectedMap[actual[i].ValidIngredientID].ID = actual[i].ID
		}

		assert.Equal(t, expectedMap, actualMap)
	})

	T.Run("with scaling", func(t *testing.T) {
		t.Parallel()

		listGenerator := &groceryListCreator{
			logger: loggingnoop.NewLogger(),
			tracer: tracing.NewTracerForTest(t.Name()),
		}

		onion := fakes.BuildFakeValidIngredient()
		carrot := fakes.BuildFakeValidIngredient()
		celery := fakes.BuildFakeValidIngredient()
		grams := fakes.BuildFakeValidMeasurementUnit()

		// Set up IDs for options, recipes, and steps
		option1ID := fakes.BuildFakeID()
		option2ID := fakes.BuildFakeID()
		option3ID := fakes.BuildFakeID()
		recipe1ID := fakes.BuildFakeID()
		recipe2ID := fakes.BuildFakeID()
		recipe3ID := fakes.BuildFakeID()
		step1ID := fakes.BuildFakeID()
		step2ID := fakes.BuildFakeID()
		step3ID := fakes.BuildFakeID()

		expectedMealPlan := &mealplanning.MealPlan{
			ID: fakes.BuildFakeID(),
			Events: []*mealplanning.MealPlanEvent{
				{
					Options: []*mealplanning.MealPlanOption{
						{
							ID:        option1ID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 1.0,
										Recipe: mealplanning.Recipe{
											ID: recipe1ID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: step1ID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														{
															Ingredient:  onion,
															MinQuantity: 100,

															MaxQuantity:     new(float32(100)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
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
							ID:        option2ID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 2.0,
										Recipe: mealplanning.Recipe{
											ID: recipe2ID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: step2ID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														{
															Ingredient:  carrot,
															MinQuantity: 100,

															MaxQuantity:     new(float32(100)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
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
							ID:        option3ID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 3.0,
										Recipe: mealplanning.Recipe{
											ID: recipe3ID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: step3ID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														{
															Ingredient:  celery,
															MinQuantity: 100,

															MaxQuantity:     new(float32(100)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
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

		expectedMap := map[string]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput{
			onion.ID: {
				Status:                  mealplanning.MealPlanGroceryListItemStatusNeeds,
				ValidMeasurementUnitID:  grams.ID,
				ValidIngredientID:       onion.ID,
				BelongsToMealPlan:       expectedMealPlan.ID,
				BelongsToMealPlanOption: &option1ID,
				RecipeID:                &recipe1ID,
				RecipeStepID:            &step1ID,
				MinQuantityNeeded:       100,

				MaxQuantityNeeded: new(float32(100)),
			},
			carrot.ID: {
				Status:                  mealplanning.MealPlanGroceryListItemStatusNeeds,
				ValidMeasurementUnitID:  grams.ID,
				ValidIngredientID:       carrot.ID,
				BelongsToMealPlan:       expectedMealPlan.ID,
				BelongsToMealPlanOption: &option2ID,
				RecipeID:                &recipe2ID,
				RecipeStepID:            &step2ID,
				MinQuantityNeeded:       200,

				MaxQuantityNeeded: new(float32(200)),
			},
			celery.ID: {
				Status:                  mealplanning.MealPlanGroceryListItemStatusNeeds,
				ValidMeasurementUnitID:  grams.ID,
				ValidIngredientID:       celery.ID,
				BelongsToMealPlan:       expectedMealPlan.ID,
				BelongsToMealPlanOption: &option3ID,
				RecipeID:                &recipe3ID,
				RecipeStepID:            &step3ID,
				MinQuantityNeeded:       300,

				MaxQuantityNeeded: new(float32(300)),
			},
		}

		actual, err := listGenerator.GenerateGroceryListInputs(ctx, expectedMealPlan)
		require.NoError(t, err)

		actualMap := map[string]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput{}
		for i := range actual {
			actualMap[actual[i].ValidIngredientID] = actual[i]
			expectedMap[actual[i].ValidIngredientID].ID = actual[i].ID
		}

		assert.Equal(t, expectedMap, actualMap)
	})

	T.Run("with ingredient scale_factor", func(t *testing.T) {
		t.Parallel()

		listGenerator := &groceryListCreator{
			logger: loggingnoop.NewLogger(),
			tracer: tracing.NewTracerForTest(t.Name()),
		}

		onion := fakes.BuildFakeValidIngredient()
		grams := fakes.BuildFakeValidMeasurementUnit()
		optionID := fakes.BuildFakeID()
		recipeID := fakes.BuildFakeID()
		stepID := fakes.BuildFakeID()

		// Recipe scaled 2x; ingredient has scale_factor 0.5 so effective scale = 2 * 0.5 = 1.0
		// Quantity min 100 -> 100 (unchanged). Without scale_factor it would be 200.
		expectedMealPlan := &mealplanning.MealPlan{
			ID: fakes.BuildFakeID(),
			Events: []*mealplanning.MealPlanEvent{
				{
					Options: []*mealplanning.MealPlanOption{
						{
							ID:        optionID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 2.0,
										Recipe: mealplanning.Recipe{
											ID: recipeID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: stepID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														{
															Ingredient:  onion,
															MinQuantity: 100,

															MaxQuantity:     new(float32(100)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
															ScaleFactor:     0.5,
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
		actual, err := listGenerator.GenerateGroceryListInputs(ctx, expectedMealPlan)
		require.NoError(t, err)
		require.Len(t, actual, 1)
		// effectiveScale = 2.0 * 0.5 = 1.0, so 100 * 1.0 = 100
		assert.InEpsilon(t, float32(100), actual[0].MinQuantityNeeded, quantityEpsilon)
		assert.InEpsilon(t, float32(100), *actual[0].MaxQuantityNeeded, quantityEpsilon)
	})

	T.Run("with option groups", func(t *testing.T) {
		t.Parallel()

		listGenerator := &groceryListCreator{
			logger: loggingnoop.NewLogger(),
			tracer: tracing.NewTracerForTest(t.Name()),
		}

		spaghetti := fakes.BuildFakeValidIngredient()
		angelHair := fakes.BuildFakeValidIngredient()
		onion := fakes.BuildFakeValidIngredient()
		grams := fakes.BuildFakeValidMeasurementUnit()

		// Set up IDs
		optionID := fakes.BuildFakeID()
		recipeID := fakes.BuildFakeID()
		stepID := fakes.BuildFakeID()

		expectedMealPlan := &mealplanning.MealPlan{
			ID: fakes.BuildFakeID(),
			Events: []*mealplanning.MealPlanEvent{
				{
					Options: []*mealplanning.MealPlanOption{
						{
							ID:        optionID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 1.0,
										Recipe: mealplanning.Recipe{
											ID: recipeID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: stepID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														// Option group: same index (0), different option_index
														{
															Ingredient:  spaghetti,
															MinQuantity: 100,

															MaxQuantity:     new(float32(100)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
														},
														{
															Ingredient:  angelHair,
															MinQuantity: 100,

															MaxQuantity:     new(float32(100)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     1,
														},
														// Non-option ingredient at different index
														{
															Ingredient:  onion,
															MinQuantity: 50,

															MaxQuantity:     new(float32(50)),
															MeasurementUnit: *grams,
															Index:           1,
															OptionIndex:     0,
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

		actual, err := listGenerator.GenerateGroceryListInputs(ctx, expectedMealPlan)
		require.NoError(t, err)

		// Should have 2 items: spaghetti (default optionIndex=0) and onion (non-option)
		// angelHair (optionIndex=1) is NOT included because no selection was made, so we default to optionIndex=0
		assert.Len(t, actual, 2)

		// Find items by ingredient ID
		actualMap := make(map[string]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput)
		for i := range actual {
			actualMap[actual[i].ValidIngredientID] = actual[i]
		}

		// Verify spaghetti (option group item - default selection)
		spaghettiItem, ok := actualMap[spaghetti.ID]
		assert.True(t, ok, "spaghetti item should exist (default optionIndex=0)")
		assert.NotNil(t, spaghettiItem.BelongsToMealPlanOption)
		assert.Equal(t, optionID, *spaghettiItem.BelongsToMealPlanOption)
		assert.NotNil(t, spaghettiItem.RecipeID)
		assert.Equal(t, recipeID, *spaghettiItem.RecipeID)
		assert.NotNil(t, spaghettiItem.RecipeStepID)
		assert.Equal(t, stepID, *spaghettiItem.RecipeStepID)
		assert.NotNil(t, spaghettiItem.IngredientIndex)
		assert.Equal(t, uint16(0), *spaghettiItem.IngredientIndex)
		assert.NotNil(t, spaghettiItem.OptionIndex)
		assert.Equal(t, uint16(0), *spaghettiItem.OptionIndex)
		assert.InEpsilon(t, float32(100), spaghettiItem.MinQuantityNeeded, quantityEpsilon)

		// Verify angelHair is NOT present (was not selected, and optionIndex=1 is not the default)
		_, ok = actualMap[angelHair.ID]
		assert.False(t, ok, "angelHair item should NOT exist (optionIndex=1 is not selected)")

		// Verify onion (non-option item, should still have recipe context)
		onionItem, ok := actualMap[onion.ID]
		assert.True(t, ok, "onion item should exist")
		assert.NotNil(t, onionItem.BelongsToMealPlanOption)
		assert.Equal(t, optionID, *onionItem.BelongsToMealPlanOption)
		assert.NotNil(t, onionItem.RecipeID)
		assert.Equal(t, recipeID, *onionItem.RecipeID)
		assert.NotNil(t, onionItem.RecipeStepID)
		assert.Equal(t, stepID, *onionItem.RecipeStepID)
		assert.InEpsilon(t, float32(50), onionItem.MinQuantityNeeded, quantityEpsilon)
	})

	T.Run("with option groups and aggregation", func(t *testing.T) {
		t.Parallel()

		listGenerator := &groceryListCreator{
			logger: loggingnoop.NewLogger(),
			tracer: tracing.NewTracerForTest(t.Name()),
		}

		spaghetti := fakes.BuildFakeValidIngredient()
		angelHair := fakes.BuildFakeValidIngredient()
		onion := fakes.BuildFakeValidIngredient()
		grams := fakes.BuildFakeValidMeasurementUnit()

		// Set up IDs
		option1ID := fakes.BuildFakeID()
		option2ID := fakes.BuildFakeID()
		recipe1ID := fakes.BuildFakeID()
		recipe2ID := fakes.BuildFakeID()
		step1ID := fakes.BuildFakeID()
		step2ID := fakes.BuildFakeID()

		expectedMealPlan := &mealplanning.MealPlan{
			ID: fakes.BuildFakeID(),
			Events: []*mealplanning.MealPlanEvent{
				{
					Options: []*mealplanning.MealPlanOption{
						{
							ID:        option1ID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 1.0,
										Recipe: mealplanning.Recipe{
											ID: recipe1ID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: step1ID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														// Option group
														{
															Ingredient:  spaghetti,
															MinQuantity: 100,

															MaxQuantity:     new(float32(100)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
														},
														{
															Ingredient:  angelHair,
															MinQuantity: 100,

															MaxQuantity:     new(float32(100)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     1,
														},
														// Non-option ingredient
														{
															Ingredient:  onion,
															MinQuantity: 50,

															MaxQuantity:     new(float32(50)),
															MeasurementUnit: *grams,
															Index:           1,
															OptionIndex:     0,
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
							ID:        option2ID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 1.0,
										Recipe: mealplanning.Recipe{
											ID: recipe2ID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: step2ID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														// Same onion ingredient, should aggregate
														{
															Ingredient:  onion,
															MinQuantity: 50,

															MaxQuantity:     new(float32(50)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
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

		actual, err := listGenerator.GenerateGroceryListInputs(ctx, expectedMealPlan)
		require.NoError(t, err)

		// Should have 2 items: spaghetti (default optionIndex=0) and onion (aggregated)
		// angelHair (optionIndex=1) is NOT included because no selection was made
		assert.Len(t, actual, 2)

		actualMap := make(map[string]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput)
		for i := range actual {
			actualMap[actual[i].ValidIngredientID] = actual[i]
		}

		// Verify spaghetti (option group item - default selection)
		spaghettiItem, ok := actualMap[spaghetti.ID]
		assert.True(t, ok, "spaghetti item should exist (default optionIndex=0)")
		assert.InEpsilon(t, float32(100), spaghettiItem.MinQuantityNeeded, quantityEpsilon)
		assert.NotNil(t, spaghettiItem.BelongsToMealPlanOption)
		assert.Equal(t, option1ID, *spaghettiItem.BelongsToMealPlanOption)

		// Verify angelHair is NOT present (was not selected)
		_, ok = actualMap[angelHair.ID]
		assert.False(t, ok, "angelHair item should NOT exist (optionIndex=1 is not selected)")

		// Verify onion (non-option item, should be aggregated)
		onionItem, ok := actualMap[onion.ID]
		assert.True(t, ok)
		assert.InEpsilon(t, float32(100), onionItem.MinQuantityNeeded, quantityEpsilon, "onion should be aggregated (50 + 50)")
		// Consolidated across two options/recipes/steps, so attribution is cleared rather than
		// misattributed to the first contributor.
		assert.Nil(t, onionItem.BelongsToMealPlanOption)
		assert.Nil(t, onionItem.RecipeID)
		assert.Nil(t, onionItem.RecipeStepID)
	})

	T.Run("with option groups and user selection", func(t *testing.T) {
		t.Parallel()

		listGenerator := &groceryListCreator{
			logger: loggingnoop.NewLogger(),
			tracer: tracing.NewTracerForTest(t.Name()),
		}

		spaghetti := fakes.BuildFakeValidIngredient()
		angelHair := fakes.BuildFakeValidIngredient()
		onion := fakes.BuildFakeValidIngredient()
		grams := fakes.BuildFakeValidMeasurementUnit()

		// Set up IDs
		optionID := fakes.BuildFakeID()
		recipeID := fakes.BuildFakeID()
		stepID := fakes.BuildFakeID()

		expectedMealPlan := &mealplanning.MealPlan{
			ID: fakes.BuildFakeID(),
			// User has selected optionIndex=1 (angelHair) instead of the default (spaghetti)
			Selections: []*mealplanning.MealPlanRecipeOptionSelection{
				{
					RecipeStepID:        stepID,
					IngredientIndex:     0,
					SelectedOptionIndex: 1, // User selected angelHair (optionIndex=1)
					SelectionType:       mealplanning.MealPlanRecipeOptionSelectionTypeIngredient,
				},
			},
			Events: []*mealplanning.MealPlanEvent{
				{
					Options: []*mealplanning.MealPlanOption{
						{
							ID:        optionID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 1.0,
										Recipe: mealplanning.Recipe{
											ID: recipeID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: stepID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														// Alternative A: spaghetti (index=0, optionIndex=0)
														{
															ID:              fakes.BuildFakeID(),
															Ingredient:      spaghetti,
															MeasurementUnit: *grams,
															MinQuantity:     100,
															Index:           0,
															OptionIndex:     0,
														},
														// Alternative B: angelHair (index=0, optionIndex=1)
														{
															ID:              fakes.BuildFakeID(),
															Ingredient:      angelHair,
															MeasurementUnit: *grams,
															MinQuantity:     100,
															Index:           0,
															OptionIndex:     1,
														},
														// Non-option ingredient at different index
														{
															ID:              fakes.BuildFakeID(),
															Ingredient:      onion,
															MeasurementUnit: *grams,
															MinQuantity:     50,
															Index:           1,
															OptionIndex:     0,
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

		actual, err := listGenerator.GenerateGroceryListInputs(ctx, expectedMealPlan)
		require.NoError(t, err)

		// Should have 2 items: angelHair (selected optionIndex=1) and onion (non-option)
		// spaghetti (optionIndex=0) is NOT included because user selected optionIndex=1
		assert.Len(t, actual, 2)

		// Find items by ingredient ID
		actualMap := make(map[string]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput)
		for i := range actual {
			actualMap[actual[i].ValidIngredientID] = actual[i]
		}

		// Verify spaghetti is NOT present (user selected optionIndex=1, not 0)
		_, ok := actualMap[spaghetti.ID]
		assert.False(t, ok, "spaghetti item should NOT exist (user selected optionIndex=1)")

		// Verify angelHair IS present (user's selection)
		angelHairItem, ok := actualMap[angelHair.ID]
		assert.True(t, ok, "angelHair item should exist (user selected optionIndex=1)")
		assert.NotNil(t, angelHairItem.BelongsToMealPlanOption)
		assert.Equal(t, optionID, *angelHairItem.BelongsToMealPlanOption)
		assert.NotNil(t, angelHairItem.RecipeID)
		assert.Equal(t, recipeID, *angelHairItem.RecipeID)
		assert.NotNil(t, angelHairItem.RecipeStepID)
		assert.Equal(t, stepID, *angelHairItem.RecipeStepID)
		assert.NotNil(t, angelHairItem.IngredientIndex)
		assert.Equal(t, uint16(0), *angelHairItem.IngredientIndex)
		assert.NotNil(t, angelHairItem.OptionIndex)
		assert.Equal(t, uint16(1), *angelHairItem.OptionIndex)
		assert.InEpsilon(t, float32(100), angelHairItem.MinQuantityNeeded, quantityEpsilon)

		// Verify onion (non-option item)
		onionItem, ok := actualMap[onion.ID]
		assert.True(t, ok, "onion item should exist")
		assert.NotNil(t, onionItem.BelongsToMealPlanOption)
		assert.Equal(t, optionID, *onionItem.BelongsToMealPlanOption)
		assert.NotNil(t, onionItem.RecipeID)
		assert.Equal(t, recipeID, *onionItem.RecipeID)
		assert.NotNil(t, onionItem.RecipeStepID)
		assert.Equal(t, stepID, *onionItem.RecipeStepID)
		assert.InEpsilon(t, float32(50), onionItem.MinQuantityNeeded, quantityEpsilon)
	})

	T.Run("with associated recipes", func(t *testing.T) {
		t.Parallel()

		listGenerator := &groceryListCreator{
			logger: loggingnoop.NewLogger(),
			tracer: tracing.NewTracerForTest(t.Name()),
		}

		// Ingredients for main recipe
		chicken := fakes.BuildFakeValidIngredient()
		// Ingredients for associated recipe (dressing)
		oliveOil := fakes.BuildFakeValidIngredient()
		lemon := fakes.BuildFakeValidIngredient()
		grams := fakes.BuildFakeValidMeasurementUnit()

		// Set up IDs
		optionID := fakes.BuildFakeID()
		mainRecipeID := fakes.BuildFakeID()
		associatedRecipeID := fakes.BuildFakeID()
		mainStepID := fakes.BuildFakeID()
		associatedStepID := fakes.BuildFakeID()

		expectedMealPlan := &mealplanning.MealPlan{
			ID: fakes.BuildFakeID(),
			Events: []*mealplanning.MealPlanEvent{
				{
					Options: []*mealplanning.MealPlanOption{
						{
							ID:        optionID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 1.0,
										Recipe: mealplanning.Recipe{
											ID: mainRecipeID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: mainStepID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														{
															Ingredient:  chicken,
															MinQuantity: 500,

															MaxQuantity:     new(float32(500)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
														},
													},
												},
											},
											// Associated recipe (e.g., Caesar Dressing)
											AssociatedRecipes: []*mealplanning.Recipe{
												{
													ID: associatedRecipeID,
													Steps: []*mealplanning.RecipeStep{
														{
															ID: associatedStepID,
															Ingredients: []*mealplanning.RecipeStepIngredient{
																{
																	Ingredient:  oliveOil,
																	MinQuantity: 100,

																	MaxQuantity:     new(float32(100)),
																	MeasurementUnit: *grams,
																	Index:           0,
																	OptionIndex:     0,
																},
																{
																	Ingredient:  lemon,
																	MinQuantity: 50,

																	MaxQuantity:     new(float32(50)),
																	MeasurementUnit: *grams,
																	Index:           1,
																	OptionIndex:     0,
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
			},
		}

		ctx := t.Context()

		actual, err := listGenerator.GenerateGroceryListInputs(ctx, expectedMealPlan)
		require.NoError(t, err)

		// Should have 3 items: chicken (from main recipe), oliveOil and lemon (from associated recipe)
		assert.Len(t, actual, 3)

		// Build a map for easier lookup
		actualMap := make(map[string]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput)
		for i := range actual {
			actualMap[actual[i].ValidIngredientID] = actual[i]
		}

		// Verify chicken (from main recipe)
		chickenItem, ok := actualMap[chicken.ID]
		assert.True(t, ok, "chicken item should exist")
		assert.NotNil(t, chickenItem.RecipeID)
		assert.Equal(t, mainRecipeID, *chickenItem.RecipeID)
		assert.NotNil(t, chickenItem.RecipeStepID)
		assert.Equal(t, mainStepID, *chickenItem.RecipeStepID)
		assert.InEpsilon(t, float32(500), chickenItem.MinQuantityNeeded, quantityEpsilon)

		// Verify oliveOil (from associated recipe)
		oliveOilItem, ok := actualMap[oliveOil.ID]
		assert.True(t, ok, "oliveOil item should exist from associated recipe")
		assert.NotNil(t, oliveOilItem.RecipeID)
		assert.Equal(t, associatedRecipeID, *oliveOilItem.RecipeID)
		assert.NotNil(t, oliveOilItem.RecipeStepID)
		assert.Equal(t, associatedStepID, *oliveOilItem.RecipeStepID)
		assert.InEpsilon(t, float32(100), oliveOilItem.MinQuantityNeeded, quantityEpsilon)

		// Verify lemon (from associated recipe)
		lemonItem, ok := actualMap[lemon.ID]
		assert.True(t, ok, "lemon item should exist from associated recipe")
		assert.NotNil(t, lemonItem.RecipeID)
		assert.Equal(t, associatedRecipeID, *lemonItem.RecipeID)
		assert.NotNil(t, lemonItem.RecipeStepID)
		assert.Equal(t, associatedStepID, *lemonItem.RecipeStepID)
		assert.InEpsilon(t, float32(50), lemonItem.MinQuantityNeeded, quantityEpsilon)
	})

	T.Run("with associated recipes and scaling", func(t *testing.T) {
		t.Parallel()

		listGenerator := &groceryListCreator{
			logger: loggingnoop.NewLogger(),
			tracer: tracing.NewTracerForTest(t.Name()),
		}

		// Ingredients for main recipe
		chicken := fakes.BuildFakeValidIngredient()
		// Ingredients for associated recipe
		oliveOil := fakes.BuildFakeValidIngredient()
		grams := fakes.BuildFakeValidMeasurementUnit()

		// Set up IDs
		optionID := fakes.BuildFakeID()
		mainRecipeID := fakes.BuildFakeID()
		associatedRecipeID := fakes.BuildFakeID()
		mainStepID := fakes.BuildFakeID()
		associatedStepID := fakes.BuildFakeID()

		expectedMealPlan := &mealplanning.MealPlan{
			ID: fakes.BuildFakeID(),
			Events: []*mealplanning.MealPlanEvent{
				{
					Options: []*mealplanning.MealPlanOption{
						{
							ID:        optionID,
							Chosen:    true,
							MealScale: 2.0, // Meal scaled 2x
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 1.5, // Recipe scaled 1.5x
										Recipe: mealplanning.Recipe{
											ID: mainRecipeID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: mainStepID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														{
															Ingredient:  chicken,
															MinQuantity: 500,

															MaxQuantity:     new(float32(500)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
														},
													},
												},
											},
											AssociatedRecipes: []*mealplanning.Recipe{
												{
													ID: associatedRecipeID,
													Steps: []*mealplanning.RecipeStep{
														{
															ID: associatedStepID,
															Ingredients: []*mealplanning.RecipeStepIngredient{
																{
																	Ingredient:  oliveOil,
																	MinQuantity: 100,

																	MaxQuantity:     new(float32(100)),
																	MeasurementUnit: *grams,
																	Index:           0,
																	OptionIndex:     0,
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
			},
		}

		ctx := t.Context()

		actual, err := listGenerator.GenerateGroceryListInputs(ctx, expectedMealPlan)
		require.NoError(t, err)

		// Should have 2 items: chicken and oliveOil
		assert.Len(t, actual, 2)

		// Build a map for easier lookup
		actualMap := make(map[string]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput)
		for i := range actual {
			actualMap[actual[i].ValidIngredientID] = actual[i]
		}

		// Verify chicken (from main recipe) - should be scaled: 500 * 1.5 * 2.0 = 1500
		chickenItem, ok := actualMap[chicken.ID]
		assert.True(t, ok, "chicken item should exist")
		assert.InEpsilon(t, float32(1500), chickenItem.MinQuantityNeeded, quantityEpsilon)

		// Verify oliveOil (from associated recipe) - should also be scaled: 100 * 1.5 * 2.0 = 300
		oliveOilItem, ok := actualMap[oliveOil.ID]
		assert.True(t, ok, "oliveOil item should exist from associated recipe")
		assert.InEpsilon(t, float32(300), oliveOilItem.MinQuantityNeeded, quantityEpsilon)
	})

	T.Run("with associated recipes and aggregation", func(t *testing.T) {
		t.Parallel()

		listGenerator := &groceryListCreator{
			logger: loggingnoop.NewLogger(),
			tracer: tracing.NewTracerForTest(t.Name()),
		}

		// Ingredient that appears in both main recipe and associated recipe
		salt := fakes.BuildFakeValidIngredient()
		chicken := fakes.BuildFakeValidIngredient()
		grams := fakes.BuildFakeValidMeasurementUnit()

		// Set up IDs
		optionID := fakes.BuildFakeID()
		mainRecipeID := fakes.BuildFakeID()
		associatedRecipeID := fakes.BuildFakeID()
		mainStepID := fakes.BuildFakeID()
		associatedStepID := fakes.BuildFakeID()

		expectedMealPlan := &mealplanning.MealPlan{
			ID: fakes.BuildFakeID(),
			Events: []*mealplanning.MealPlanEvent{
				{
					Options: []*mealplanning.MealPlanOption{
						{
							ID:        optionID,
							Chosen:    true,
							MealScale: 1.0,
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 1.0,
										Recipe: mealplanning.Recipe{
											ID: mainRecipeID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: mainStepID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														{
															Ingredient:  chicken,
															MinQuantity: 500,

															MaxQuantity:     new(float32(500)),
															MeasurementUnit: *grams,
															Index:           0,
															OptionIndex:     0,
														},
														{
															Ingredient:  salt,
															MinQuantity: 10,

															MaxQuantity:     new(float32(10)),
															MeasurementUnit: *grams,
															Index:           1,
															OptionIndex:     0,
														},
													},
												},
											},
											AssociatedRecipes: []*mealplanning.Recipe{
												{
													ID: associatedRecipeID,
													Steps: []*mealplanning.RecipeStep{
														{
															ID: associatedStepID,
															Ingredients: []*mealplanning.RecipeStepIngredient{
																{
																	Ingredient:  salt, // Same salt ingredient
																	MinQuantity: 5,

																	MaxQuantity:     new(float32(5)),
																	MeasurementUnit: *grams,
																	Index:           0,
																	OptionIndex:     0,
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
			},
		}

		ctx := t.Context()

		actual, err := listGenerator.GenerateGroceryListInputs(ctx, expectedMealPlan)
		require.NoError(t, err)

		// Should have 2 items: chicken and salt (aggregated from main + associated)
		assert.Len(t, actual, 2)

		// Build a map for easier lookup
		actualMap := make(map[string]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput)
		for i := range actual {
			actualMap[actual[i].ValidIngredientID] = actual[i]
		}

		// Verify chicken
		chickenItem, ok := actualMap[chicken.ID]
		assert.True(t, ok, "chicken item should exist")
		assert.InEpsilon(t, float32(500), chickenItem.MinQuantityNeeded, quantityEpsilon)

		// Verify salt (should be aggregated: 10 + 5 = 15)
		saltItem, ok := actualMap[salt.ID]
		assert.True(t, ok, "salt item should exist and be aggregated")
		assert.InEpsilon(t, float32(15), saltItem.MinQuantityNeeded, quantityEpsilon, "salt should be aggregated from main recipe (10) and associated recipe (5)")
		// Same option contributed both amounts, so option attribution survives, but the recipe/step
		// attribution is cleared because the contributions came from different recipes/steps.
		assert.NotNil(t, saltItem.BelongsToMealPlanOption)
		assert.Equal(t, optionID, *saltItem.BelongsToMealPlanOption)
		assert.Nil(t, saltItem.RecipeID)
		assert.Nil(t, saltItem.RecipeStepID)
	})

	T.Run("rounds quantities to nearest tenth", func(t *testing.T) {
		t.Parallel()

		listGenerator := &groceryListCreator{
			logger: loggingnoop.NewLogger(),
			tracer: tracing.NewTracerForTest(t.Name()),
		}

		carrot := fakes.BuildFakeValidIngredient()
		thyme := fakes.BuildFakeValidIngredient()
		pounds := fakes.BuildFakeValidMeasurementUnit()
		sprigs := fakes.BuildFakeValidMeasurementUnit()

		optionID := fakes.BuildFakeID()
		recipeID := fakes.BuildFakeID()
		stepID := fakes.BuildFakeID()

		expectedMealPlan := &mealplanning.MealPlan{
			ID: fakes.BuildFakeID(),
			Events: []*mealplanning.MealPlanEvent{
				{
					Options: []*mealplanning.MealPlanOption{
						{
							ID:        optionID,
							Chosen:    true,
							MealScale: 1.34, // Scale that produces fractional totals
							Meal: mealplanning.Meal{
								Components: []*mealplanning.MealComponent{
									{
										RecipeScale: 1.0,
										Recipe: mealplanning.Recipe{
											ID: recipeID,
											Steps: []*mealplanning.RecipeStep{
												{
													ID: stepID,
													Ingredients: []*mealplanning.RecipeStepIngredient{
														{
															Ingredient:  carrot,
															MinQuantity: 3.01,

															MaxQuantity:     new(float32(3.01)),
															MeasurementUnit: *pounds,
															Index:           0,
															OptionIndex:     0,
														},
														{
															Ingredient:  thyme,
															MinQuantity: 1.5,

															MaxQuantity:     new(float32(1.5)),
															MeasurementUnit: *sprigs,
															Index:           1,
															OptionIndex:     0,
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

		actual, err := listGenerator.GenerateGroceryListInputs(ctx, expectedMealPlan)
		require.NoError(t, err)

		actualMap := make(map[string]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput)
		for i := range actual {
			actualMap[actual[i].ValidIngredientID] = actual[i]
		}

		// 3.01 * 1.34 = 4.0334 -> should round to 4.0
		carrotItem, ok := actualMap[carrot.ID]
		assert.True(t, ok, "carrot item should exist")
		assert.InEpsilon(t, float32(4.0), carrotItem.MinQuantityNeeded, quantityEpsilon, "carrot quantity should be rounded to nearest tenth")
		assert.NotNil(t, carrotItem.MaxQuantityNeeded)
		assert.InEpsilon(t, float32(4.0), *carrotItem.MaxQuantityNeeded, quantityEpsilon)

		// 1.5 * 1.34 = 2.01 -> should round to 2.0
		thymeItem, ok := actualMap[thyme.ID]
		assert.True(t, ok, "thyme item should exist")
		assert.InEpsilon(t, float32(2.0), thymeItem.MinQuantityNeeded, quantityEpsilon, "thyme quantity should be rounded to nearest tenth")
		assert.NotNil(t, thymeItem.MaxQuantityNeeded)
		assert.InEpsilon(t, float32(2.0), *thymeItem.MaxQuantityNeeded, quantityEpsilon)
	})
}
