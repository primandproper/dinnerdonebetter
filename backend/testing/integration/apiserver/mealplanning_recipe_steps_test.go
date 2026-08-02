package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mpconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkRecipeStepEquality(t *testing.T, index int, expected, actual *mealplanning.RecipeStep) {
	t.Helper()

	assert.NotZero(t, actual.CreatedAt, "expected recipe step %d", index)
	assert.Equal(t, expected.MinEstimatedTimeInSeconds, actual.MinEstimatedTimeInSeconds, "expected recipe step %d", index)
	assert.Equal(t, expected.MaxEstimatedTimeInSeconds, actual.MaxEstimatedTimeInSeconds, "expected recipe step %d", index)
	assert.Equal(t, expected.MinTemperatureInCelsius, actual.MinTemperatureInCelsius, "expected recipe step %d", index)
	assert.Equal(t, expected.MaxTemperatureInCelsius, actual.MaxTemperatureInCelsius, "expected recipe step %d", index)
	assert.NotEmpty(t, actual.BelongsToRecipe, "expected recipe step %d", index)
	assert.Equal(t, expected.ConditionExpression, actual.ConditionExpression, "expected recipe step %d", index)
	assert.NotEmpty(t, actual.ID, "expected recipe step %d", index)
	assert.Equal(t, expected.Notes, actual.Notes, "expected recipe step %d", index)
	assert.Equal(t, expected.ExplicitInstructions, actual.ExplicitInstructions, "expected recipe step %d", index)
	checkRecipeMediaSliceEquality(t, index, expected.Media, actual.Media)
	checkRecipeStepProductSliceEquality(t, index, expected.Products, actual.Products)
	checkRecipeStepInstrumentSliceEquality(t, index, expected.Instruments, actual.Instruments)
	checkRecipeStepVesselSliceEquality(t, index, expected.Vessels, actual.Vessels)
	checkRecipeStepCompletionConditionSliceEquality(t, index, expected.CompletionConditions, actual.CompletionConditions)
	checkRecipeStepIngredientSliceEquality(t, index, expected.Ingredients, actual.Ingredients)
	checkValidPreparationEquality(t, index, &expected.Preparation, &actual.Preparation)
	assert.Equal(t, expected.Index, actual.Index, "expected recipe step %d", index)
	assert.Equal(t, expected.Optional, actual.Optional, "expected recipe step %d", index)
	assert.Equal(t, expected.StartTimerAutomatically, actual.StartTimerAutomatically, "expected recipe step %d", index)
}

//nolint:paralleltest // test is inexplicably flaky when run in parallel
func TestRecipeSteps_CompleteLifecycle(T *testing.T) {
	T.Run("should update", func(t *testing.T) {
		ctx := t.Context()

		createdValidIngredients, createdValidPreparation, createdRecipe := createRecipeForTest(t, nil)

		var (
			createdRecipeStep *mealplanning.RecipeStep
			stepIndex         int
		)
		for i, step := range createdRecipe.Steps {
			createdRecipeStep = step
			stepIndex = i
			break
		}

		require.NotNil(t, createdRecipeStep)

		newRecipeStep := fakes.BuildFakeRecipeStep()
		newRecipeStep.BelongsToRecipe = createdRecipe.ID
		newRecipeStep.Media = createdRecipeStep.Media
		newRecipeStep.Products = createdRecipeStep.Products
		newRecipeStep.Instruments = createdRecipeStep.Instruments
		newRecipeStep.Vessels = createdRecipeStep.Vessels
		newRecipeStep.CompletionConditions = createdRecipeStep.CompletionConditions
		newRecipeStep.Ingredients = createdRecipeStep.Ingredients
		for j := range newRecipeStep.Ingredients {
			newRecipeStep.Ingredients[j].Ingredient = createdValidIngredients[j]
		}

		updateInput := mpconverters.ConvertRecipeStepToRecipeStepUpdateRequestInput(newRecipeStep)
		updateInput.Preparation = createdValidPreparation

		createdRecipeStep.Update(updateInput)

		updateResponse, err := adminClient.UpdateRecipeStep(ctx, &mealplanninggrpc.UpdateRecipeStepRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStep.ID,
			Input:        converters.ConvertRecipeStepUpdateRequestInputToGRPCRecipeStepUpdateRequestInput(updateInput),
		})
		require.NoError(t, err)

		// Test the response from UpdateRecipeStep first
		checkRecipeStepEquality(t, stepIndex, createdRecipeStep, converters.ConvertGRPCRecipeStepToRecipeStep(updateResponse.Updated))
		assert.NotNil(t, updateResponse.Updated.LastUpdatedAt)

		// Also test the separate GetRecipeStep call
		actual, err := adminClient.GetRecipeStep(ctx, &mealplanninggrpc.GetRecipeStepRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStep.ID,
		})
		require.NoError(t, err)

		// assert recipe step equality for the separate get call
		checkRecipeStepEquality(t, stepIndex, createdRecipeStep, converters.ConvertGRPCRecipeStepToRecipeStep(actual.Result))
		assert.NotNil(t, actual.Result.LastUpdatedAt)

		_, err = adminClient.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStep.ID,
		})
		assert.NoError(t, err)

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipeSteps_Listing(T *testing.T) {
	T.Parallel()

	T.Run("should be readable in paginated form", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		createdValidIngredients, _, createdRecipe := createRecipeForTest(t, nil)
		createdValidPreparation := createValidPreparationForTest(t) // Use a new preparation to avoid duplicate VIP/VPI/VPV (createRecipeForTest already creates bridge entries)
		createdValidMeasurementUnit := createValidMeasurementUnitForTest(t)
		createdValidInstrument := createValidInstrumentForTest(t)
		createdValidIngredientState := createValidIngredientStateForTest(t)
		createdValidVessel := createValidVesselForTest(t)

		// Create bridge table entries for the new preparation (unique from createRecipeForTest's)
		var validIngredientPreparations []*mealplanning.ValidIngredientPreparation
		var validIngredientMeasurementUnits []*mealplanning.ValidIngredientMeasurementUnit
		for _, ingredient := range createdValidIngredients {
			vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, createdValidPreparation)
			validIngredientPreparations = append(validIngredientPreparations, vip)

			vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, createdValidMeasurementUnit)
			validIngredientMeasurementUnits = append(validIngredientMeasurementUnits, vimu)
		}

		// Create bridge table entry for preparation+instrument
		createdValidPreparationInstrument := createValidPreparationInstrumentWithEntitiesForTest(t, createdValidPreparation, createdValidInstrument)

		// Create bridge table entry for preparation+vessel
		createdValidPreparationVessel := createValidPreparationVesselWithEntitiesForTest(t, createdValidPreparation, createdValidVessel)

		var expected []*mealplanning.RecipeStep
		for i := range 5 {
			t.Logf("creating recipe step #%d", i+1)

			exampleRecipeStep := fakes.BuildFakeRecipeStep()
			exampleRecipeStep.BelongsToRecipe = createdRecipe.ID
			for j := range exampleRecipeStep.Ingredients {
				exampleRecipeStep.Ingredients[j].Ingredient = createdValidIngredients[j]
				exampleRecipeStep.Ingredients[j].MeasurementUnit = *createdValidMeasurementUnit
				// Set Index to array position (will be used when Index is nil in creation input)
				exampleRecipeStep.Ingredients[j].Index = uint16(j)
				exampleRecipeStep.Ingredients[j].OptionIndex = 0
			}

			for j := range exampleRecipeStep.Products {
				exampleRecipeStep.Products[j].MeasurementUnit = createdValidMeasurementUnit
			}

			for j := range exampleRecipeStep.Instruments {
				exampleRecipeStep.Instruments[j].Instrument = createdValidInstrument
				// Set Index to array position (will be used when Index is nil in creation input)
				exampleRecipeStep.Instruments[j].Index = uint16(j)
				exampleRecipeStep.Instruments[j].OptionIndex = 0
			}

			for j := range exampleRecipeStep.Vessels {
				exampleRecipeStep.Vessels[j].Vessel = createdValidVessel
				// Set Index to array position (will be used when Index is nil in creation input)
				exampleRecipeStep.Vessels[j].Index = uint16(j)
				exampleRecipeStep.Vessels[j].OptionIndex = 0
			}

			for j := range exampleRecipeStep.CompletionConditions {
				exampleRecipeStep.CompletionConditions[j].IngredientState = *createdValidIngredientState
				for k := range exampleRecipeStep.CompletionConditions[j].Ingredients {
					// Reference the first recipe step ingredient that will be created
					exampleRecipeStep.CompletionConditions[j].Ingredients[k].RecipeStepIngredient = exampleRecipeStep.Ingredients[0].ID
				}
			}

			exampleRecipeStepInput := mpconverters.ConvertRecipeStepToRecipeStepCreationRequestInput(exampleRecipeStep)
			exampleRecipeStepInput.PreparationID = createdValidPreparation.ID

			// Set bridge table IDs for ingredients
			for j := range exampleRecipeStepInput.Ingredients {
				exampleRecipeStepInput.Ingredients[j].ValidIngredientPreparationID = &validIngredientPreparations[j].ID
				exampleRecipeStepInput.Ingredients[j].ValidIngredientMeasurementUnitID = &validIngredientMeasurementUnits[j].ID
			}

			// Set bridge table IDs for instruments
			for j := range exampleRecipeStepInput.Instruments {
				exampleRecipeStepInput.Instruments[j].ValidPreparationInstrumentID = &createdValidPreparationInstrument.ID
			}

			// Set bridge table IDs for vessels
			for j := range exampleRecipeStepInput.Vessels {
				exampleRecipeStepInput.Vessels[j].ValidPreparationVesselID = &createdValidPreparationVessel.ID
			}

			// Set the preparation on the example recipe step for comparison
			exampleRecipeStep.Preparation = *createdValidPreparation

			createdRecipeStep, createdRecipeStepErr := adminClient.CreateRecipeStep(ctx, &mealplanninggrpc.CreateRecipeStepRequest{
				RecipeId: createdRecipe.ID,
				Input:    converters.ConvertRecipeStepCreationRequestInputToGRPCRecipeStepCreationRequestInput(exampleRecipeStepInput),
			})
			require.NoError(t, createdRecipeStepErr)

			// Update the completion condition ingredient references to match the actual created recipe step
			createdRecipeStepConverted := converters.ConvertGRPCRecipeStepToRecipeStep(createdRecipeStep.Created)
			for j := range exampleRecipeStep.CompletionConditions {
				for k := range exampleRecipeStep.CompletionConditions[j].Ingredients {
					// Use the actual created recipe step ingredient MealPlanTaskID
					exampleRecipeStep.CompletionConditions[j].Ingredients[k].RecipeStepIngredient = createdRecipeStepConverted.Ingredients[k].ID
				}
			}

			checkRecipeStepEquality(t, -1, exampleRecipeStep, createdRecipeStepConverted)

			recipeStep, err := adminClient.GetRecipeStep(ctx, &mealplanninggrpc.GetRecipeStepRequest{
				RecipeId:     createdRecipe.ID,
				RecipeStepId: createdRecipeStep.Created.Id,
			})
			require.NoError(t, err)
			require.Equal(t, createdRecipe.ID, recipeStep.Result.BelongsToRecipe)

			expected = append(expected, converters.ConvertGRPCRecipeStepToRecipeStep(recipeStep.Result))
			t.Logf("created step #%d", i)
		}

		// assert recipe step list equality
		actual, err := adminClient.GetRecipeSteps(ctx, &mealplanninggrpc.GetRecipeStepsRequest{RecipeId: createdRecipe.ID})
		require.NoError(t, err)
		assert.True(
			t,
			len(expected) <= len(actual.Results),
			"expected %d to be <= %d",
			len(expected),
			len(actual.Results),
		)

		for _, createdRecipeStep := range expected {
			_, err = adminClient.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{RecipeId: createdRecipe.ID, RecipeStepId: createdRecipeStep.ID})
			assert.NoError(t, err)
		}

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipeSteps_OwnershipDenial(T *testing.T) {
	T.Parallel()

	T.Run("non-owner cannot create recipe step", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, otherUserClient := createUserAndClientForTest(t)
		_, _, createdRecipe := createRecipeForTest(t, nil)

		exampleRecipeStep := fakes.BuildFakeRecipeStep()
		exampleRecipeStepInput := mpconverters.ConvertRecipeStepToRecipeStepCreationRequestInput(exampleRecipeStep)
		_, err := otherUserClient.CreateRecipeStep(ctx, &mealplanninggrpc.CreateRecipeStepRequest{
			RecipeId: createdRecipe.ID,
			Input:    converters.ConvertRecipeStepCreationRequestInputToGRPCRecipeStepCreationRequestInput(exampleRecipeStepInput),
		})
		require.Error(t, err)
	})

	T.Run("non-owner cannot update recipe step", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, otherUserClient := createUserAndClientForTest(t)
		_, _, createdRecipe := createRecipeForTest(t, nil)

		createdRecipeStep := createdRecipe.Steps[0]

		newRecipeStep := fakes.BuildFakeRecipeStep()
		updateInput := mpconverters.ConvertRecipeStepToRecipeStepUpdateRequestInput(newRecipeStep)
		_, err := otherUserClient.UpdateRecipeStep(ctx, &mealplanninggrpc.UpdateRecipeStepRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStep.ID,
			Input:        converters.ConvertRecipeStepUpdateRequestInputToGRPCRecipeStepUpdateRequestInput(updateInput),
		})
		require.Error(t, err)
	})

	T.Run("non-owner cannot archive recipe step", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, otherUserClient := createUserAndClientForTest(t)
		_, _, createdRecipe := createRecipeForTest(t, nil)

		createdRecipeStep := createdRecipe.Steps[0]

		_, err := otherUserClient.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStep.ID,
		})
		require.Error(t, err)
	})
}

/*

// TODO: uncomment me.
//func (s *TestSuite) TestRecipeSteps_ContentUploading() {
//	s.runTest("should be able to upload content for a recipe step", func(testClients *testClientWrapper) func() {
//		return func() {
//			t := s.T()
//
//			ctx, span := tracing.StartCustomSpan(s.ctx, t.Name())
//			defer span.End()
//
//			_, _, createdRecipe := createRecipeForTest(ctx, t, adminClient, adminClient, nil)
//
//			var createdRecipeStep *mealplanning.RecipeStep
//			for _, step := range createdRecipe.Steps {
//				createdRecipeStep = step
//				break
//			}
//
//			require.NotNil(t, createdRecipeStep)
//
//			_, img1Bytes := testutils.BuildArbitraryImagePNGBytes(200)
//			_, img2Bytes := testutils.BuildArbitraryImagePNGBytes(250)
//			_, img3Bytes := testutils.BuildArbitraryImagePNGBytes(300)
//
//			files := map[string][]byte{
//				"image_1.png": img1Bytes,
//				"image_2.png": img2Bytes,
//				"image_3.png": img3Bytes,
//			}
//
//			require.NoError(t, adminClient.UploadMediaForRecipeStep(ctx, files, createdRecipe.MealPlanTaskID, createdRecipeStep.MealPlanTaskID))
//
//			assert.NoError(t, adminClient.ArchiveRecipe(ctx, createdRecipe.MealPlanTaskID))
//		}
//	})
//}


*/
