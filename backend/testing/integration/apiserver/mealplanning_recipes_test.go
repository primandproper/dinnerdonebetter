package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mpconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkRecipeEquality(t *testing.T, expected, actual *mealplanning.Recipe) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected Name for recipe %s to be %v, but it was %v", expected.ID, expected.Name, actual.Name)
	assert.Equal(t, expected.InspiredByRecipeID, actual.InspiredByRecipeID, "expected InspiredByRecipeID for recipe %s to be %v, but it was %v", expected.ID, expected.InspiredByRecipeID, actual.InspiredByRecipeID)
	assert.Equal(t, expected.MinEstimatedPortions, actual.MinEstimatedPortions, "expected MinEstimatedPortions for recipe %s to be %v, but it was %v", expected.ID, expected.MinEstimatedPortions, actual.MinEstimatedPortions)
	assert.Equal(t, expected.MaxEstimatedPortions, actual.MaxEstimatedPortions, "expected MaxEstimatedPortions for recipe %s to be %v, but it was %v", expected.ID, expected.MaxEstimatedPortions, actual.MaxEstimatedPortions)
	assert.Equal(t, expected.PluralPortionName, actual.PluralPortionName, "expected PluralPortionName for recipe %s to be %v, but it was %v", expected.ID, expected.PluralPortionName, actual.PluralPortionName)
	assert.Equal(t, expected.Description, actual.Description, "expected Description for recipe %s to be %v, but it was %v", expected.ID, expected.Description, actual.Description)
	assert.Equal(t, expected.Name, actual.Name, "expected Name for recipe %s to be %v, but it was %v", expected.ID, expected.Name, actual.Name)
	assert.Equal(t, expected.PortionName, actual.PortionName, "expected PortionName for recipe %s to be %v, but it was %v", expected.ID, expected.PortionName, actual.PortionName)
	assert.NotZero(t, actual.CreatedByUser)
	assert.Equal(t, expected.Source, actual.Source, "expected Source for recipe %s to be %v, but it was %v", expected.ID, expected.Source, actual.Source)
	assert.Equal(t, expected.Slug, actual.Slug, "expected Slug for recipe %s to be %v, but it was %v", expected.ID, expected.Slug, actual.Slug)
	assert.Equal(t, expected.YieldsComponentType, actual.YieldsComponentType, "expected YieldsComponentType for recipe %s to be %v, but it was %v", expected.ID, expected.YieldsComponentType, actual.YieldsComponentType)
	checkRecipePrepTaskSliceEquality(t, expected.PrepTasks, actual.PrepTasks)
	checkRecipeLevelMediaSliceEquality(t, expected.Media, actual.Media)
	assert.Equal(t, expected.EligibleForMeals, actual.EligibleForMeals, "expected EligibleForMeals for recipe %s to be %v, but it was %v", expected.ID, expected.EligibleForMeals, actual.EligibleForMeals)

	for i, step := range expected.Steps {
		checkRecipeStepEquality(t, i, step, actual.Steps[i])
	}

	assert.NotZero(t, actual.CreatedAt)
}

// createRecipeForTestWithInstrument creates a recipe using the given instrument (required, optional=false).
// Use when testing SearchForRecipesWithInstrumentOwnership.
func createRecipeForTestWithInstrument(t *testing.T, instrument *mealplanning.ValidInstrument, recipe *mealplanning.Recipe) ([]*mealplanning.ValidIngredient, *mealplanning.ValidPreparation, *mealplanning.Recipe) {
	t.Helper()
	return createRecipeForTestWithInstrumentOrNew(t, instrument, recipe, nil)
}

func createRecipeForTest(t *testing.T, recipe *mealplanning.Recipe, inputFilter ...func(input *mealplanning.RecipeCreationRequestInput)) ([]*mealplanning.ValidIngredient, *mealplanning.ValidPreparation, *mealplanning.Recipe) {
	t.Helper()
	return createRecipeForTestWithInstrumentOrNew(t, nil, recipe, inputFilter)
}

func createRecipeForTestWithInstrumentOrNew(t *testing.T, instrument *mealplanning.ValidInstrument, recipe *mealplanning.Recipe, inputFilter []func(input *mealplanning.RecipeCreationRequestInput)) ([]*mealplanning.ValidIngredient, *mealplanning.ValidPreparation, *mealplanning.Recipe) {
	t.Helper()

	ctx := t.Context()

	createdValidPreparation := createValidPreparationForTest(t)
	createdValidMeasurementUnit := createValidMeasurementUnitForTest(t)
	createdValidInstrument := instrument
	if createdValidInstrument == nil {
		createdValidInstrument = createValidInstrumentForTest(t)
	}
	createdValidIngredientState := createValidIngredientStateForTest(t)
	createdValidVessel := createValidVesselForTest(t)

	// Create bridge table entries for preparation+instrument and preparation+vessel
	// These are shared across all steps since we use the same preparation, instrument, and vessel
	createdValidPreparationInstrument := createValidPreparationInstrumentWithEntitiesForTest(t, createdValidPreparation, createdValidInstrument)
	createdValidPreparationVessel := createValidPreparationVesselWithEntitiesForTest(t, createdValidPreparation, createdValidVessel)

	exampleRecipe := fakes.BuildFakeRecipe()
	if recipe != nil {
		exampleRecipe = recipe
	}
	exampleRecipe.Media = []*mealplanning.RecipeMedia{}

	createdValidIngredients := []*mealplanning.ValidIngredient{}
	ingredientPreparationIDs := make(map[string]string)
	ingredientMeasurementUnitIDs := make(map[string]string)

	for i, recipeStep := range exampleRecipe.Steps {
		for j := range recipeStep.Ingredients {
			createdValidIngredient := createValidIngredientForTest(t)
			createdValidIngredients = append(createdValidIngredients, createdValidIngredient)

			// Create bridge table entries for this ingredient
			createdVIP := createValidIngredientPreparationWithEntitiesForTest(t, createdValidIngredient, createdValidPreparation)
			createdVIMU := createValidIngredientMeasurementUnitWithEntitiesForTest(t, createdValidIngredient, createdValidMeasurementUnit)

			ingredientPreparationIDs[createdValidIngredient.ID] = createdVIP.ID
			ingredientMeasurementUnitIDs[createdValidIngredient.ID] = createdVIMU.ID

			exampleRecipe.Steps[i].Ingredients[j].Ingredient = createdValidIngredient
			exampleRecipe.Steps[i].Ingredients[j].MeasurementUnit = *createdValidMeasurementUnit
		}

		for j := range recipeStep.Products {
			exampleRecipe.Steps[i].Products[j].MeasurementUnit = createdValidMeasurementUnit
		}

		for j := range recipeStep.Instruments {
			recipeStep.Instruments[j].Instrument = createdValidInstrument
			if instrument != nil {
				recipeStep.Instruments[j].Optional = false // Required for SearchForRecipesWithInstrumentOwnership
			}
		}

		for j := range recipeStep.Vessels {
			recipeStep.Vessels[j].Vessel = createdValidVessel
		}

		for j := range recipeStep.CompletionConditions {
			recipeStep.CompletionConditions[j].IngredientState = *createdValidIngredientState
			for k := range recipeStep.CompletionConditions[j].Ingredients {
				recipeStep.CompletionConditions[j].Ingredients[k].RecipeStepIngredient = recipeStep.Ingredients[0].ID
			}
		}
	}

	exampleRecipeInput := mpconverters.ConvertRecipeToRecipeCreationRequestInput(exampleRecipe)
	exampleRecipeInput.AlsoCreateMeal = true
	// Set bridge table IDs
	for i := range exampleRecipeInput.Steps {
		exampleRecipeInput.Steps[i].PreparationID = createdValidPreparation.ID
		for j := range exampleRecipeInput.Steps[i].Ingredients {
			// Use the ingredient ID from the original recipe (which we've already set)
			if exampleRecipe.Steps[i].Ingredients[j].Ingredient != nil {
				ingredientID := exampleRecipe.Steps[i].Ingredients[j].Ingredient.ID
				if vipID, ok := ingredientPreparationIDs[ingredientID]; ok {
					exampleRecipeInput.Steps[i].Ingredients[j].ValidIngredientPreparationID = &vipID
				}
				if vimuID, ok := ingredientMeasurementUnitIDs[ingredientID]; ok {
					exampleRecipeInput.Steps[i].Ingredients[j].ValidIngredientMeasurementUnitID = &vimuID
				}
			}
		}

		for j := range exampleRecipeInput.Steps[i].Instruments {
			exampleRecipeInput.Steps[i].Instruments[j].ValidPreparationInstrumentID = &createdValidPreparationInstrument.ID
		}

		for j := range exampleRecipeInput.Steps[i].Vessels {
			exampleRecipeInput.Steps[i].Vessels[j].ValidPreparationVesselID = &createdValidPreparationVessel.ID
		}
	}

	examplePrepTask := fakes.BuildFakeRecipePrepTask()
	examplePrepTask.TaskSteps = []*mealplanning.RecipePrepTaskStep{
		{
			BelongsToRecipeStep: exampleRecipe.Steps[0].ID,
			SatisfiesRecipeStep: false,
		},
	}
	exampleRecipeInput.PrepTasks = []*mealplanning.RecipePrepTaskWithinRecipeCreationRequestInput{
		mpconverters.ConvertRecipePrepTaskToRecipePrepTaskWithinRecipeCreationRequestInput(exampleRecipe, examplePrepTask),
	}
	// Update the exampleRecipe to match what we're actually creating
	exampleRecipe.PrepTasks = []*mealplanning.RecipePrepTask{examplePrepTask}

	for _, filter := range inputFilter {
		filter(exampleRecipeInput)
	}

	createdRes, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(exampleRecipeInput)})
	require.NoError(t, err)

	createdRecipe, err := adminClient.GetRecipe(ctx, &mealplanninggrpc.GetRecipeRequest{RecipeId: createdRes.Created.Id})
	require.NoError(t, err)
	require.NotNil(t, createdRecipe)

	// Only do basic comparisons that we know should work
	converted := converters.ConvertGRPCRecipeToRecipe(createdRecipe.Result)
	require.NotEmpty(t, createdRecipe.Result.Steps, "created recipe must have steps")
	require.NotEmpty(t, converted.Steps, "converted recipe must have steps")

	// Verify that completion conditions are present (this was the original issue)
	for i, step := range converted.Steps {
		require.NotEmpty(t, step.CompletionConditions, "recipe step %d must have completion conditions", i)
		for j, condition := range step.CompletionConditions {
			require.NotEmpty(t, condition.Ingredients, "completion condition %d for step %d must have ingredients", j, i)
		}
	}

	return createdValidIngredients, createdValidPreparation, converted
}

func TestRecipes_Creating(T *testing.T) {
	T.Parallel()

	T.Run("realistic", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		soak := createValidPreparationForTest(t)
		mix := createValidPreparationForTest(t)
		grams := createValidMeasurementUnitForTest(t)
		cups := createValidMeasurementUnitForTest(t)
		pintoBeans := createValidIngredientForTest(t)
		water := createValidIngredientForTest(t)
		garlicPaste := createValidIngredientForTest(t)
		createdValidInstrument := createValidInstrumentForTest(t)

		// Create bridge table entries
		// ValidIngredientPreparations: ingredient+preparation combos
		vipPintoSoak := createValidIngredientPreparationWithEntitiesForTest(t, pintoBeans, soak)
		vipWaterSoak := createValidIngredientPreparationWithEntitiesForTest(t, water, soak)
		vipGarlicMix := createValidIngredientPreparationWithEntitiesForTest(t, garlicPaste, mix)

		// ValidIngredientMeasurementUnits: ingredient+unit combos
		vimuPintoGrams := createValidIngredientMeasurementUnitWithEntitiesForTest(t, pintoBeans, grams)
		vimuWaterCups := createValidIngredientMeasurementUnitWithEntitiesForTest(t, water, cups)
		vimuGarlicGrams := createValidIngredientMeasurementUnitWithEntitiesForTest(t, garlicPaste, grams)

		// ValidPreparationInstruments: preparation+instrument combos
		vpiSoakInstrument := createValidPreparationInstrumentWithEntitiesForTest(t, soak, createdValidInstrument)
		vpiMixInstrument := createValidPreparationInstrumentWithEntitiesForTest(t, mix, createdValidInstrument)

		expected := &mealplanning.Recipe{
			Name:                 "sopa de frijol",
			Slug:                 "sopa-de-frijol",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          t.Name(),
			PluralPortionName:    t.Name(),
			Media:                []*mealplanning.RecipeMedia{},
			PrepTasks:            []*mealplanning.RecipePrepTask{},
			MinEstimatedPortions: 1,

			MaxEstimatedPortions: nil,
			Steps: []*mealplanning.RecipeStep{
				{
					Products: []*mealplanning.RecipeStepProduct{
						{
							Name:                   "soaked pinto beans",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnit:        grams,
							QuantityNotes:          "",
							MinMeasurementQuantity: new(float32(1000)),
						},
					},
					Notes:       "first step",
					Preparation: *soak, // This will be updated after recipe creation
					Instruments: []*mealplanning.RecipeStepInstrument{
						{
							Name:        "whatever",
							Instrument:  createdValidInstrument,
							Index:       0, // Array index 0
							OptionIndex: 0,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Ingredient:      pintoBeans,
							Name:            "pinto beans",
							MeasurementUnit: *grams,
							MinQuantity:     500,
							Index:           0, // Array index 0
							OptionIndex:     0,
						},
						{
							Ingredient:      water,
							Name:            "water",
							MeasurementUnit: *cups,
							MinQuantity:     5,
							Index:           1, // Array index 1
							OptionIndex:     0,
						},
					},
					Index: 0,
				},
				{
					Products: []*mealplanning.RecipeStepProduct{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnit:        grams,
							QuantityNotes:          "",
							MinMeasurementQuantity: new(float32(1010)),
						},
					},
					Notes:       "second step",
					Preparation: *mix, // This will be updated after recipe creation
					Instruments: []*mealplanning.RecipeStepInstrument{
						{
							Name:        "whatever",
							Instrument:  createdValidInstrument,
							Index:       0, // Array index 0
							OptionIndex: 0,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Name:            "soaked pinto beans",
							MeasurementUnit: *grams,
							MinQuantity:     1000,
							Index:           0, // Array index 0
							OptionIndex:     0,
						},
						{
							Ingredient:      garlicPaste,
							Name:            "garlic paste",
							MeasurementUnit: *grams,
							MinQuantity:     10,
							Index:           1, // Array index 1
							OptionIndex:     0,
						},
					},
					Index: 1,
				},
			},
		}

		expectedInput := &mealplanning.RecipeCreationRequestInput{
			Name:                 expected.Name,
			Description:          expected.Description,
			Slug:                 expected.Slug,
			YieldsComponentType:  expected.YieldsComponentType,
			PortionName:          expected.PortionName,
			PluralPortionName:    expected.PluralPortionName,
			AlsoCreateMeal:       true,
			MinEstimatedPortions: expected.MinEstimatedPortions,

			MaxEstimatedPortions: expected.MaxEstimatedPortions,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					MinTemperatureInCelsius: expected.Steps[0].MinTemperatureInCelsius,
					MaxTemperatureInCelsius: expected.Steps[0].MaxTemperatureInCelsius,
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   expected.Steps[0].Products[0].Name,
							Type:                   expected.Steps[0].Products[0].Type,
							MeasurementUnitID:      &expected.Steps[0].Products[0].MeasurementUnit.ID,
							QuantityNotes:          expected.Steps[0].Products[0].QuantityNotes,
							MinMeasurementQuantity: expected.Steps[0].Products[0].MinMeasurementQuantity,
							MaxMeasurementQuantity: expected.Steps[0].Products[0].MaxMeasurementQuantity,
						},
					},
					Notes:         expected.Steps[0].Notes,
					PreparationID: expected.Steps[0].Preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "whatever",
							ValidPreparationInstrumentID: &vpiSoakInstrument.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             expected.Steps[0].Ingredients[0].Name,
							ValidIngredientPreparationID:     &vipPintoSoak.ID,
							ValidIngredientMeasurementUnitID: &vimuPintoGrams.ID,
							MinQuantity:                      expected.Steps[0].Ingredients[0].MinQuantity,

							MaxQuantity: nil,
						},
						{
							Name:                             expected.Steps[0].Ingredients[1].Name,
							ValidIngredientPreparationID:     &vipWaterSoak.ID,
							ValidIngredientMeasurementUnitID: &vimuWaterCups.ID,
							MinQuantity:                      expected.Steps[0].Ingredients[1].MinQuantity,

							MaxQuantity: nil,
						},
					},
					Index: expected.Steps[0].Index,
				},
				{
					MinTemperatureInCelsius: expected.Steps[1].MinTemperatureInCelsius,
					MaxTemperatureInCelsius: expected.Steps[1].MaxTemperatureInCelsius,
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   expected.Steps[1].Products[0].Name,
							Type:                   expected.Steps[1].Products[0].Type,
							MeasurementUnitID:      &expected.Steps[1].Products[0].MeasurementUnit.ID,
							QuantityNotes:          expected.Steps[1].Products[0].QuantityNotes,
							MinMeasurementQuantity: expected.Steps[1].Products[0].MinMeasurementQuantity,
							MaxMeasurementQuantity: expected.Steps[1].Products[0].MaxMeasurementQuantity,
						},
					},
					Notes:         expected.Steps[1].Notes,
					PreparationID: expected.Steps[1].Preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "whatever",
							ValidPreparationInstrumentID: &vpiMixInstrument.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							// This is a recipe step product (from step 0), no bridge table IDs needed
							Name:                            expected.Steps[1].Ingredients[0].Name,
							ProductOfRecipeStepIndex:        new(uint64(0)),
							ProductOfRecipeStepProductIndex: new(uint64(0)),
							MinQuantity:                     expected.Steps[1].Ingredients[0].MinQuantity,

							MaxQuantity: nil,
						},
						{
							Name:                             expected.Steps[1].Ingredients[1].Name,
							ValidIngredientPreparationID:     &vipGarlicMix.ID,
							ValidIngredientMeasurementUnitID: &vimuGarlicGrams.ID,
							MinQuantity:                      expected.Steps[1].Ingredients[1].MinQuantity,

							MaxQuantity: nil,
						},
					},
					Index: expected.Steps[1].Index,
				},
			},
		}

		createdRes, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(expectedInput)})
		require.NoError(t, err)

		created := converters.ConvertGRPCRecipeToRecipe(createdRes.Created)
		checkRecipeEquality(t, expected, created)

		recipeRes, err := adminClient.GetRecipe(ctx, &mealplanninggrpc.GetRecipeRequest{RecipeId: createdRes.Created.Id})
		require.NoError(t, err)
		created = converters.ConvertGRPCRecipeToRecipe(recipeRes.Result)

		assert.Equal(t, created.Status, mealplanning.RecipeStatusSubmitted)
		updateRes, err := adminClient.UpdateRecipeStatus(ctx, &mealplanninggrpc.UpdateRecipeStatusRequest{
			RecipeId:  createdRes.Created.Id,
			NewStatus: mealplanning.RecipeStatusApproved,
		})
		require.NoError(t, err)
		assert.Equal(t, updateRes.Updated.Status, mealplanning.RecipeStatusApproved)

		recipeStepProductIndex := -1
		for i, ingredient := range created.Steps[1].Ingredients {
			if ingredient.RecipeStepProductID != nil {
				recipeStepProductIndex = i
			}
		}

		require.NotEqual(t, -1, recipeStepProductIndex)
		require.NotNil(t, created.Steps[1].Ingredients[recipeStepProductIndex].RecipeStepProductID)
		assert.Equal(t, created.Steps[0].Products[0].ID, *created.Steps[1].Ingredients[recipeStepProductIndex].RecipeStepProductID)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		exampleRecipe := fakes.BuildFakeRecipe()
		exampleRecipeInput := mpconverters.ConvertRecipeToRecipeCreationRequestInput(exampleRecipe)
		convertedInput := converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(exampleRecipeInput)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		exampleRecipe := fakes.BuildFakeRecipe()
		exampleRecipeInput := mpconverters.ConvertRecipeToRecipeCreationRequestInput(exampleRecipe)
		convertedInput := converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(exampleRecipeInput)

		created, err := testClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestRecipes_Updating(T *testing.T) {
	T.Parallel()

	T.Run("should update recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		_, _, createdRecipe := createRecipeForTest(t, nil)

		// Store the original recipe data for comparison
		originalSteps := make([]*mealplanning.RecipeStep, len(createdRecipe.Steps))
		for i, step := range createdRecipe.Steps {
			originalSteps[i] = &mealplanning.RecipeStep{
				ID:                   step.ID,
				CompletionConditions: step.CompletionConditions,
			}
		}

		// Create update input with new basic fields
		newRecipe := fakes.BuildFakeRecipe()
		updateInput := mpconverters.ConvertRecipeToRecipeUpdateRequestInput(newRecipe)

		_, err := adminClient.UpdateRecipe(ctx, &mealplanninggrpc.UpdateRecipeRequest{
			RecipeId: createdRecipe.ID,
			Input:    converters.ConvertRecipeUpdateRequestInputToGRPCRecipeUpdateRequestInput(updateInput),
		})
		require.NoError(t, err)

		// Retrieve the updated recipe
		actual, err := adminClient.GetRecipe(ctx, &mealplanninggrpc.GetRecipeRequest{RecipeId: createdRecipe.ID})
		require.NoError(t, err)
		require.NotNil(t, actual)
		actualRecipe := converters.ConvertGRPCRecipeToRecipe(actual.Result)

		// Assert that basic fields were updated correctly
		assert.Equal(t, newRecipe.Name, actualRecipe.Name, "recipe name should be updated")
		assert.Equal(t, newRecipe.Slug, actualRecipe.Slug, "recipe slug should be updated")
		assert.Equal(t, newRecipe.Source, actualRecipe.Source, "recipe source should be updated")
		assert.Equal(t, newRecipe.Description, actualRecipe.Description, "recipe description should be updated")
		assert.Equal(t, newRecipe.InspiredByRecipeID, actualRecipe.InspiredByRecipeID, "recipe inspired by recipe ID should be updated")
		assert.Equal(t, newRecipe.MinEstimatedPortions, actualRecipe.MinEstimatedPortions, "recipe estimated portions should be updated")
		assert.Equal(t, newRecipe.MaxEstimatedPortions, actualRecipe.MaxEstimatedPortions, "recipe estimated portions should be updated")
		assert.Equal(t, newRecipe.PortionName, actualRecipe.PortionName, "recipe portion name should be updated")
		assert.Equal(t, newRecipe.PluralPortionName, actualRecipe.PluralPortionName, "recipe plural portion name should be updated")
		assert.Equal(t, newRecipe.EligibleForMeals, actualRecipe.EligibleForMeals, "recipe eligible for meals should be updated")
		assert.Equal(t, newRecipe.YieldsComponentType, actualRecipe.YieldsComponentType, "recipe yields component type should be updated")
		assert.NotNil(t, actual.Result.LastUpdatedAt, "recipe should have last updated timestamp")

		// Assert that steps and completion conditions remain unchanged (since UpdateRecipe only updates top-level fields)
		assert.Equal(t, len(originalSteps), len(actualRecipe.Steps), "number of recipe steps should remain the same")
		for i, originalStep := range originalSteps {
			actualStep := actualRecipe.Steps[i]
			assert.Equal(t, originalStep.ID, actualStep.ID, "recipe step ID should remain unchanged")
			assert.Equal(t, len(originalStep.CompletionConditions), len(actualStep.CompletionConditions), "number of completion conditions should remain the same")

			// Verify completion conditions are still present and working (this was the original issue)
			for j, originalCondition := range originalStep.CompletionConditions {
				actualCondition := actualStep.CompletionConditions[j]
				assert.Equal(t, originalCondition.ID, actualCondition.ID, "completion condition ID should remain unchanged")
				assert.Equal(t, originalCondition.Optional, actualCondition.Optional, "completion condition optional flag should remain unchanged")
				assert.Equal(t, originalCondition.Notes, actualCondition.Notes, "completion condition notes should remain unchanged")
				assert.Equal(t, len(originalCondition.Ingredients), len(actualCondition.Ingredients), "number of completion condition ingredients should remain the same")
			}
		}

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		newRecipe := fakes.BuildFakeRecipe()
		updateInput := mpconverters.ConvertRecipeToRecipeUpdateRequestInput(newRecipe)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdateRecipe(ctx, &mealplanninggrpc.UpdateRecipeRequest{
			RecipeId: createdRecipe.ID,
			Input:    converters.ConvertRecipeUpdateRequestInputToGRPCRecipeUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)
	})

	T.Run("nonexistent recipe", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		newRecipe := fakes.BuildFakeRecipe()
		updateInput := mpconverters.ConvertRecipeToRecipeUpdateRequestInput(newRecipe)

		_, err := adminClient.UpdateRecipe(ctx, &mealplanninggrpc.UpdateRecipeRequest{
			RecipeId: nonexistentID,
			Input:    converters.ConvertRecipeUpdateRequestInputToGRPCRecipeUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)
	})
}

func TestRecipes_Searching(T *testing.T) {
	T.Parallel()

	T.Run("should be searchable by name", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		exampleRecipe := fakes.BuildFakeRecipe()

		var expected []*mealplanning.Recipe
		for i := range 5 {
			exampleRecipe.Name = fmt.Sprintf("example%d", i)
			_, _, createdRecipe := createRecipeForTest(t, exampleRecipe)

			expected = append(expected, createdRecipe)
		}

		// assert recipe list equality
		actual, err := adminClient.SearchForRecipes(ctx, &mealplanninggrpc.SearchForRecipesRequest{
			Query: "example",
		})
		require.NoError(t, err)
		assert.True(
			t,
			len(expected) <= len(actual.Results),
			"expected %d to be <= %d",
			len(expected),
			len(actual.Results),
		)

		for _, createdRecipe := range expected {
			_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
			assert.NoError(t, err)
		}
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		results, err := c.SearchForRecipes(ctx, &mealplanninggrpc.SearchForRecipesRequest{
			Query: "test",
		})
		assert.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestRecipes_SearchForRecipesWithInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("returns recipes when account owns required instruments", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		instrument := createValidInstrumentForTest(t)
		createAccountInstrumentOwnershipForInstrumentForTest(t, testClient, instrument)

		exampleRecipe := fakes.BuildFakeRecipe()
		exampleRecipe.Name = "InstrumentOwnershipSearchTest Recipe"
		_, _, createdRecipe := createRecipeForTestWithInstrument(t, instrument, exampleRecipe)

		actual, err := testClient.SearchForRecipesWithInstrumentOwnership(ctx, &mealplanninggrpc.SearchForRecipesWithInstrumentOwnershipRequest{
			Query: "InstrumentOwnershipSearchTest",
		})
		require.NoError(t, err)
		require.NotEmpty(t, actual.Results)
		found := false
		for _, r := range actual.Results {
			if r.Id == createdRecipe.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "recipe with owned instrument should appear in search results")

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})

	T.Run("excludes recipes when account does not own required instruments", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		// Create recipe with instrument that the test user does NOT own
		unownedInstrument := createValidInstrumentForTest(t)
		exampleRecipe := fakes.BuildFakeRecipe()
		exampleRecipe.Name = "ExcludedInstrumentSearchTest Recipe"
		_, _, excludedRecipe := createRecipeForTestWithInstrument(t, unownedInstrument, exampleRecipe)

		actual, err := testClient.SearchForRecipesWithInstrumentOwnership(ctx, &mealplanninggrpc.SearchForRecipesWithInstrumentOwnershipRequest{
			Query: "ExcludedInstrumentSearchTest",
		})
		require.NoError(t, err)
		found := false
		for _, r := range actual.Results {
			if r.Id == excludedRecipe.ID {
				found = true
				break
			}
		}
		assert.False(t, found, "recipe with unowned instrument should not appear in search results")

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: excludedRecipe.ID})
		assert.NoError(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		results, err := c.SearchForRecipesWithInstrumentOwnership(ctx, &mealplanninggrpc.SearchForRecipesWithInstrumentOwnershipRequest{
			Query: "test",
		})
		assert.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestRecipes_Cloning(T *testing.T) {
	T.Parallel()

	T.Run("should CRUD", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		actual, err := adminClient.CloneRecipe(ctx, &mealplanninggrpc.CloneRecipeRequest{RecipeId: createdRecipe.ID})
		require.NoError(t, err)

		require.Equal(t, createdRecipe.Name, actual.Cloned.Name)
		require.Equal(t, len(createdRecipe.Steps), len(actual.Cloned.Steps))

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		c := buildUnauthenticatedGRPCClientForTest(t)
		cloned, err := c.CloneRecipe(ctx, &mealplanninggrpc.CloneRecipeRequest{RecipeId: createdRecipe.ID})
		assert.Error(t, err)
		assert.Nil(t, cloned)
	})

	T.Run("nonexistent recipe", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		cloned, err := adminClient.CloneRecipe(ctx, &mealplanninggrpc.CloneRecipeRequest{RecipeId: nonexistentID})
		assert.Error(t, err)
		assert.Nil(t, cloned)
	})

	T.Run("non-admin users can clone", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		_, _, createdRecipe := createRecipeForTest(t, nil)

		cloned, err := testClient.CloneRecipe(ctx, &mealplanninggrpc.CloneRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
		assert.NotNil(t, cloned)
	})
}

func TestRecipes_GetMealPlanTasksForRecipe(T *testing.T) {
	T.Parallel()

	T.Run("meal plan tasks with frozen chicken breast", func(t *testing.T) {
		t.Parallel()

		dice := createValidPreparationForTest(t)
		grams := createValidMeasurementUnitForTest(t)
		chickenBreast := createValidIngredientForTest(t)
		createdValidInstrument := createValidInstrumentForTest(t)
		sautee := createValidPreparationForTest(t)

		// Create bridge table entries
		vipChickenDice := createValidIngredientPreparationWithEntitiesForTest(t, chickenBreast, dice)
		vimuChickenGrams := createValidIngredientMeasurementUnitWithEntitiesForTest(t, chickenBreast, grams)
		vpiDiceInstrument := createValidPreparationInstrumentWithEntitiesForTest(t, dice, createdValidInstrument)
		vpiSauteeInstrument := createValidPreparationInstrumentWithEntitiesForTest(t, sautee, createdValidInstrument)

		expected := &mealplanning.Recipe{
			Name:                 "sopa de frijol",
			Slug:                 "whatever-who-cares-sopa-de-frijol",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          t.Name(),
			PluralPortionName:    t.Name(),
			MinEstimatedPortions: 1,
			EligibleForMeals:     true,

			MaxEstimatedPortions: nil,
			Steps: []*mealplanning.RecipeStep{
				{
					Products: []*mealplanning.RecipeStepProduct{
						{
							Name:                   "diced chicken breast",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnit:        grams,
							QuantityNotes:          "",
							MinMeasurementQuantity: new(float32(1000)),
						},
					},
					Notes:       "first step",
					Preparation: *dice,
					Instruments: []*mealplanning.RecipeStepInstrument{
						{
							Name:       "whatever",
							Instrument: createdValidInstrument,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							RecipeStepProductID: nil,
							Ingredient:          chickenBreast,
							Name:                "pinto beans",
							MeasurementUnit:     *grams,
							MinQuantity:         500,
						},
					},
					Index: 0,
				},
				{
					Products: []*mealplanning.RecipeStepProduct{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnit:        grams,
							QuantityNotes:          "",
							MinMeasurementQuantity: new(float32(1010)),
						},
					},
					Notes:       "second step",
					Preparation: *sautee,
					Instruments: []*mealplanning.RecipeStepInstrument{
						{
							Name:       "whatever",
							Instrument: createdValidInstrument,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Name:            "diced chicken breast",
							MeasurementUnit: *grams,
							MinQuantity:     1000,
						},
					},
					Index: 1,
				},
			},
		}

		expectedInput := &mealplanning.RecipeCreationRequestInput{
			Name:                 expected.Name,
			Slug:                 expected.Slug,
			YieldsComponentType:  expected.YieldsComponentType,
			PortionName:          expected.PortionName,
			PluralPortionName:    expected.PluralPortionName,
			MinEstimatedPortions: expected.MinEstimatedPortions,
			EligibleForMeals:     true,

			MaxEstimatedPortions: expected.MaxEstimatedPortions,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "diced chicken breast",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &grams.ID,
							QuantityNotes:          "",
							MinMeasurementQuantity: new(float32(1000)),
						},
					},
					Notes:         "first step",
					PreparationID: dice.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "whatever",
							ValidPreparationInstrumentID: &vpiDiceInstrument.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "pinto beans",
							ValidIngredientPreparationID:     &vipChickenDice.ID,
							ValidIngredientMeasurementUnitID: &vimuChickenGrams.ID,
							MinQuantity:                      500,
						},
					},
					Index: 0,
				},
				{
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &grams.ID,
							QuantityNotes:          "",
							MinMeasurementQuantity: new(float32(1010)),
						},
					},
					Notes:         "second step",
					PreparationID: sautee.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "whatever",
							ValidPreparationInstrumentID: &vpiSauteeInstrument.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							// This is a recipe step product (from step 0), no bridge table IDs needed
							Name:                            "diced chicken breast",
							MinQuantity:                     1000,
							ProductOfRecipeStepIndex:        new(uint64(0)),
							ProductOfRecipeStepProductIndex: new(uint64(0)),
						},
					},
					Index: 1,
				},
			},
		}

		ctx := t.Context()

		created, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(expectedInput)})
		require.NoError(t, err)
		createdRecipe := converters.ConvertGRPCRecipeToRecipe(created.Created)
		checkRecipeEquality(t, expected, createdRecipe)

		// Build a meal from the recipe and an auto-finalized single-option meal plan, whose
		// finalization saga generates the meal plan tasks. GetMealPlanTasks is account-scoped
		// (verifyMealPlanAccess), so it must be queried with a real meal plan ID owned by the
		// requester's account, not a recipe ID.
		_, userClient := createUserAndClientForTest(t)
		meal := createMealFromRecipe(t, createdRecipe, t.Name())
		require.NotNil(t, meal)

		now := time.Now()
		mealPlan := &mealplanning.MealPlan{
			Notes:                  t.Name(),
			Status:                 string(mealplanning.MealPlanStatusFinalized),
			GroceryListInitialized: true,
			TasksCreated:           true,
			// voting deadline must be before every event's start time; see MealPlanCreationRequestInput.ValidateWithContext.
			VotingDeadline: now.Add(1 * time.Hour),
			ElectionMethod: mealplanning.MealPlanElectionMethodSchulze,
			Events: []*mealplanning.MealPlanEvent{
				{
					StartsAt: now.Add(24 * time.Hour),
					EndsAt:   now.Add(72 * time.Hour),
					MealName: mealplanning.DinnerMealName,
					Options: []*mealplanning.MealPlanOption{
						{Meal: mealplanning.Meal{ID: meal.ID}},
					},
				},
			},
		}

		exampleMealPlanInput := mpconverters.ConvertMealPlanToMealPlanCreationRequestInput(mealPlan)
		createMealPlanRes, err := userClient.CreateMealPlan(ctx, &mealplanninggrpc.CreateMealPlanRequest{
			Input: converters.ConvertMealPlanCreationRequestInputToGRPCMealPlanCreationRequestInput(exampleMealPlanInput),
		})
		require.NoError(t, err)
		require.NotEmpty(t, createMealPlanRes.Created.Id)

		tasks := awaitMealPlanTasksCreated(t, ctx, userClient, createMealPlanRes.Created.Id)

		// H15: frozen thaw-task generation is intentionally deferred. meal_plan_tasks.belongs_to_recipe_prep_task
		// is NOT NULL and references recipe_prep_tasks(id), so ad-hoc thaw tasks (which have no backing prep task)
		// cannot be persisted until a migration makes that column nullable. Until then, NO thaw task is generated.
		// This mirrors the unit test "does not currently create frozen thawing steps".
		for _, task := range tasks {
			assert.NotContains(t, task.GetCreationExplanation(), "thawed",
				"frozen thaw-task generation is deferred (H15); no thaw task should be produced")
		}
	})
}

func TestRecipes_CreationWithDiscreteProducts(T *testing.T) {
	T.Parallel()

	T.Run("should create recipe with discrete and continuous products", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Create valid entities
		createdValidMeasurementUnit := createValidMeasurementUnitForTest(t)
		createdValidInstrument := createValidInstrumentForTest(t)

		// Create preparations (must be created before bridge table entries)
		shape := createValidPreparationForTest(t)
		assemble := createValidPreparationForTest(t)

		// Create ingredients
		groundBeef := createValidIngredientForTest(t)
		cheese := createValidIngredientForTest(t)

		// Create bridge table entries for ingredients (must match the preparation used in the step)
		vipGroundBeefShape := createValidIngredientPreparationWithEntitiesForTest(t, groundBeef, shape)
		vimuGroundBeefOunce := createValidIngredientMeasurementUnitWithEntitiesForTest(t, groundBeef, createdValidMeasurementUnit)

		vipCheeseSlice := createValidIngredientPreparationWithEntitiesForTest(t, cheese, assemble)
		vimuCheeseSlice := createValidIngredientMeasurementUnitWithEntitiesForTest(t, cheese, createdValidMeasurementUnit)

		// Create bridge table entries for preparations
		vpiShapeInstrument := createValidPreparationInstrumentWithEntitiesForTest(t, shape, createdValidInstrument)
		vpiAssembleInstrument := createValidPreparationInstrumentWithEntitiesForTest(t, assemble, createdValidInstrument)

		// Expected recipe with discrete product (beef patties) and continuous product (sauce)
		expected := &mealplanning.Recipe{
			Name:                 "Cheeseburgers",
			Slug:                 "cheeseburgers",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStep{
				{
					Products: []*mealplanning.RecipeStepProduct{
						{
							Name:            "beef patties",
							Type:            mealplanning.RecipeStepProductIngredientType,
							MeasurementUnit: createdValidMeasurementUnit,
							QuantityNotes:   "",
							// Discrete product: 4 patties, each 4 ounces
							MinItemQuantity:        new(float32(4)),
							MinMeasurementQuantity: new(float32(4)),
						},
					},
					Notes:       "Shape the meat into patties",
					Preparation: *shape,
					Instruments: []*mealplanning.RecipeStepInstrument{
						{
							Name:       "hands",
							Instrument: createdValidInstrument,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Ingredient:      groundBeef,
							Name:            "ground beef",
							MeasurementUnit: *createdValidMeasurementUnit,
						},
					},
					Index: 0,
				},
				{
					Products: []*mealplanning.RecipeStepProduct{
						{
							Name:            "special sauce",
							Type:            mealplanning.RecipeStepProductIngredientType,
							MeasurementUnit: createdValidMeasurementUnit,
							QuantityNotes:   "",
							// Continuous product: 8 ounces total

							MinMeasurementQuantity: new(float32(8)),
						},
					},
					Notes:       "Mix the sauce",
					Preparation: *assemble,
					Instruments: []*mealplanning.RecipeStepInstrument{
						{
							Name:       "spoon",
							Instrument: createdValidInstrument,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Ingredient:      cheese,
							Name:            "cheese",
							MeasurementUnit: *createdValidMeasurementUnit,
						},
					},
					Index: 1,
				},
			},
		}

		expectedInput := &mealplanning.RecipeCreationRequestInput{
			Name:                 expected.Name,
			Slug:                 expected.Slug,
			YieldsComponentType:  expected.YieldsComponentType,
			PortionName:          expected.PortionName,
			PluralPortionName:    expected.PluralPortionName,
			MinEstimatedPortions: expected.MinEstimatedPortions,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:              "beef patties",
							Type:              mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID: &createdValidMeasurementUnit.ID,
							QuantityNotes:     "",
							// Discrete product: 4 patties, each 4 ounces
							MinItemQuantity:        new(float32(4)),
							MinMeasurementQuantity: new(float32(4)),
						},
					},
					Notes:         "Shape the meat into patties",
					PreparationID: shape.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "hands",
							ValidPreparationInstrumentID: &vpiShapeInstrument.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "ground beef",
							ValidIngredientPreparationID:     &vipGroundBeefShape.ID,
							ValidIngredientMeasurementUnitID: &vimuGroundBeefOunce.ID,
							MinQuantity:                      16,
						},
					},
					Index: 0,
				},
				{
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:              "special sauce",
							Type:              mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID: &createdValidMeasurementUnit.ID,
							QuantityNotes:     "",
							// Continuous product: 8 ounces total (ItemQuantity not set)

							MinMeasurementQuantity: new(float32(8)),
						},
					},
					Notes:         "Mix the sauce",
					PreparationID: assemble.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpiAssembleInstrument.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "cheese",
							ValidIngredientPreparationID:     &vipCheeseSlice.ID,
							ValidIngredientMeasurementUnitID: &vimuCheeseSlice.ID,
							MinQuantity:                      4,
						},
					},
					Index: 1,
				},
			},
		}

		// Create the recipe
		created, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(expectedInput),
		})
		require.NoError(t, err)

		createdRecipe := converters.ConvertGRPCRecipeToRecipe(created.Created)

		// Verify the recipe was created correctly
		require.NotEmpty(t, createdRecipe.ID)
		assert.Equal(t, expected.Name, createdRecipe.Name)

		// Verify step 0 has discrete product (beef patties)
		require.Len(t, createdRecipe.Steps, len(expected.Steps))
		require.Len(t, createdRecipe.Steps[0].Products, 1)
		step0Product := createdRecipe.Steps[0].Products[0]
		assert.Equal(t, "beef patties", step0Product.Name)
		// Verify discrete product fields
		require.NotNil(t, step0Product.MinItemQuantity, "discrete product should have ItemQuantity.Min set")
		assert.Equal(t, float32(4), *step0Product.MinItemQuantity, "ItemQuantity should be 4 patties")
		assert.Equal(t, float32(4), *step0Product.MinMeasurementQuantity, "MeasurementQuantity should be 4 oz per patty")

		// Verify step 1 has continuous product (sauce)
		require.Len(t, createdRecipe.Steps[1].Products, 1)
		step1Product := createdRecipe.Steps[1].Products[0]
		assert.Equal(t, "special sauce", step1Product.Name)
		// Verify continuous product fields (ItemQuantity should be empty)
		assert.Nil(t, step1Product.MinItemQuantity, "continuous product should not have ItemQuantity.Min set")
		assert.Nil(t, step1Product.MaxItemQuantity, "continuous product should not have ItemQuantity.Max set")
		assert.Equal(t, float32(8), *step1Product.MinMeasurementQuantity, "MeasurementQuantity should be 8 oz total")

		// Verify we can retrieve the recipe and products are still correct
		retrieved, err := adminClient.GetRecipe(ctx, &mealplanninggrpc.GetRecipeRequest{
			RecipeId: createdRecipe.ID,
		})
		require.NoError(t, err)

		retrievedRecipe := converters.ConvertGRPCRecipeToRecipe(retrieved.Result)

		// Verify discrete product is still correct after retrieval
		require.Len(t, retrievedRecipe.Steps[0].Products, 1)
		retrievedStep0Product := retrievedRecipe.Steps[0].Products[0]
		assert.Equal(t, "beef patties", retrievedStep0Product.Name)
		require.NotNil(t, retrievedStep0Product.MinItemQuantity)
		assert.Equal(t, float32(4), *retrievedStep0Product.MinItemQuantity)
		assert.Equal(t, float32(4), *retrievedStep0Product.MinMeasurementQuantity)

		// Verify continuous product is still correct after retrieval
		require.Len(t, retrievedRecipe.Steps[1].Products, 1)
		retrievedStep1Product := retrievedRecipe.Steps[1].Products[0]
		assert.Equal(t, "special sauce", retrievedStep1Product.Name)
		assert.Nil(t, retrievedStep1Product.MinItemQuantity)
		assert.Nil(t, retrievedStep1Product.MaxItemQuantity)
		assert.Equal(t, float32(8), *retrievedStep1Product.MinMeasurementQuantity)

		// Cleanup
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{
			RecipeId: createdRecipe.ID,
		})
		assert.NoError(t, err)
	})
}

func TestRecipes_StepProducts_Discrete(T *testing.T) {
	T.Parallel()

	T.Run("discrete product with ItemQuantity.Min set", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)
		instrument := createValidInstrumentForTest(t)

		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation, instrument)

		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:              "cookies",
							Type:              mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID: &measurementUnit.ID,
							// Discrete product: ItemQuantity.Min set
							MinItemQuantity:        new(float32(12)),
							MinMeasurementQuantity: new(float32(2)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		created, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		require.NoError(t, err)

		createdRecipe := converters.ConvertGRPCRecipeToRecipe(created.Created)

		// Verify discrete product is correctly identified
		require.Len(t, createdRecipe.Steps[0].Products, 1)
		product := createdRecipe.Steps[0].Products[0]
		assert.Equal(t, "cookies", product.Name)
		// Verify ItemQuantity.Min is set (indicates discrete)
		require.NotNil(t, product.MinItemQuantity, "discrete product should have ItemQuantity.Min set")
		assert.Equal(t, float32(12), *product.MinItemQuantity, "ItemQuantity should be 12 cookies")
		// Verify MeasurementQuantity represents per-item measurement
		assert.Equal(t, float32(2), *product.MinMeasurementQuantity, "MeasurementQuantity should be 2 oz per cookie")

		// Cleanup
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{
			RecipeId: createdRecipe.ID,
		})
		assert.NoError(t, err)
	})

	T.Run("discrete product with ItemQuantity.Max set", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)
		instrument := createValidInstrumentForTest(t)

		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation, instrument)

		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:              "slices",
							Type:              mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID: &measurementUnit.ID,
							// Discrete product: ItemQuantity.Max set (range)
							MinItemQuantity:        new(float32(8)),
							MaxItemQuantity:        new(float32(10)),
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		created, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		require.NoError(t, err)

		createdRecipe := converters.ConvertGRPCRecipeToRecipe(created.Created)

		// Verify discrete product is correctly identified
		require.Len(t, createdRecipe.Steps[0].Products, 1)
		product := createdRecipe.Steps[0].Products[0]
		assert.Equal(t, "slices", product.Name)
		// Verify ItemQuantity.Max is set (indicates discrete with range)
		require.NotNil(t, product.MinItemQuantity, "discrete product should have ItemQuantity.Min set")
		require.NotNil(t, product.MaxItemQuantity, "discrete product with range should have ItemQuantity.Max set")
		assert.Equal(t, float32(8), *product.MinItemQuantity, "ItemQuantity.Min should be 8 slices")
		assert.Equal(t, float32(10), *product.MaxItemQuantity, "ItemQuantity.Max should be 10 slices")
		// Verify MeasurementQuantity represents per-item measurement
		assert.Equal(t, float32(1), *product.MinMeasurementQuantity, "MeasurementQuantity should be 1 oz per slice")

		// Cleanup
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{
			RecipeId: createdRecipe.ID,
		})
		assert.NoError(t, err)
	})
}

func TestRecipes_StepProducts_Continuous(T *testing.T) {
	T.Parallel()

	T.Run("continuous product with both ItemQuantity.Min and Max as null", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)
		instrument := createValidInstrumentForTest(t)

		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation, instrument)

		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:              "sauce",
							Type:              mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID: &measurementUnit.ID,
							// Continuous product: both ItemQuantity.Min and Max are null

							MinMeasurementQuantity: new(float32(16)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		created, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		require.NoError(t, err)

		createdRecipe := converters.ConvertGRPCRecipeToRecipe(created.Created)

		// Verify continuous product is correctly identified
		require.Len(t, createdRecipe.Steps[0].Products, 1)
		product := createdRecipe.Steps[0].Products[0]
		assert.Equal(t, "sauce", product.Name)
		// Verify both ItemQuantity.Min and Max are null (indicates continuous)
		assert.Nil(t, product.MinItemQuantity, "continuous product should not have ItemQuantity.Min set")
		assert.Nil(t, product.MaxItemQuantity, "continuous product should not have ItemQuantity.Max set")
		// Verify MeasurementQuantity represents total quantity
		assert.Equal(t, float32(16), *product.MinMeasurementQuantity, "MeasurementQuantity should be 16 oz total")

		// Cleanup
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{
			RecipeId: createdRecipe.ID,
		})
		assert.NoError(t, err)
	})
}

func TestRecipes_StepProducts_EdgeCases(T *testing.T) {
	T.Parallel()

	T.Run("product with ItemQuantity.Min set but Max null (discrete)", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)
		instrument := createValidInstrumentForTest(t)

		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation, instrument)

		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:              "patties",
							Type:              mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID: &measurementUnit.ID,
							// Discrete: ItemQuantity.Min set, Max null
							MinItemQuantity:        new(float32(4)),
							MinMeasurementQuantity: new(float32(4)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		created, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		require.NoError(t, err)

		createdRecipe := converters.ConvertGRPCRecipeToRecipe(created.Created)

		// Verify product is identified as discrete (ItemQuantity.Min is set)
		require.Len(t, createdRecipe.Steps[0].Products, 1)
		product := createdRecipe.Steps[0].Products[0]
		assert.Equal(t, "patties", product.Name)
		require.NotNil(t, product.MinItemQuantity, "discrete product should have ItemQuantity.Min set")
		assert.Nil(t, product.MaxItemQuantity, "ItemQuantity.Max can be null for discrete products")

		// Cleanup
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{
			RecipeId: createdRecipe.ID,
		})
		assert.NoError(t, err)
	})

	T.Run("product with ItemQuantity.Min null but Max set (discrete)", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)
		instrument := createValidInstrumentForTest(t)

		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation, instrument)

		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:              "pieces",
							Type:              mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID: &measurementUnit.ID,
							// Discrete: ItemQuantity.Min null, Max set
							MaxItemQuantity:        new(float32(6)),
							MinMeasurementQuantity: new(float32(3)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		created, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		require.NoError(t, err)

		createdRecipe := converters.ConvertGRPCRecipeToRecipe(created.Created)

		// Verify product is identified as discrete (ItemQuantity.Max is set)
		require.Len(t, createdRecipe.Steps[0].Products, 1)
		product := createdRecipe.Steps[0].Products[0]
		assert.Equal(t, "pieces", product.Name)
		assert.Nil(t, product.MinItemQuantity, "ItemQuantity.Min can be null when Max is set")
		require.NotNil(t, product.MaxItemQuantity, "discrete product should have ItemQuantity.Max set")

		// Cleanup
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{
			RecipeId: createdRecipe.ID,
		})
		assert.NoError(t, err)
	})

	T.Run("product with both ItemQuantity.Min and Max null (continuous)", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)
		instrument := createValidInstrumentForTest(t)

		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation, instrument)

		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:              "liquid",
							Type:              mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID: &measurementUnit.ID,
							// Continuous: both ItemQuantity.Min and Max are null

							MinMeasurementQuantity: new(float32(32)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		created, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		require.NoError(t, err)

		createdRecipe := converters.ConvertGRPCRecipeToRecipe(created.Created)

		// Verify product is identified as continuous (both ItemQuantity.Min and Max are null)
		require.Len(t, createdRecipe.Steps[0].Products, 1)
		product := createdRecipe.Steps[0].Products[0]
		assert.Equal(t, "liquid", product.Name)
		assert.Nil(t, product.MinItemQuantity, "continuous product should not have ItemQuantity.Min set")
		assert.Nil(t, product.MaxItemQuantity, "continuous product should not have ItemQuantity.Max set")
		// Verify MeasurementQuantity represents total quantity
		assert.Equal(t, float32(32), *product.MinMeasurementQuantity, "MeasurementQuantity should be 32 oz total")

		// Cleanup
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{
			RecipeId: createdRecipe.ID,
		})
		assert.NoError(t, err)
	})
}

func TestRecipes_Reading(T *testing.T) {
	T.Parallel()

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		c := buildUnauthenticatedGRPCClientForTest(t)
		recipe, err := c.GetRecipe(ctx, &mealplanninggrpc.GetRecipeRequest{RecipeId: createdRecipe.ID})
		assert.Error(t, err)
		assert.Nil(t, recipe)
	})

	T.Run("nonexistent recipe", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		recipe, err := adminClient.GetRecipe(ctx, &mealplanninggrpc.GetRecipeRequest{RecipeId: nonexistentID})
		assert.Error(t, err)
		assert.Nil(t, recipe)
	})
}

func TestRecipes_GetMermaidDiagramForRecipe(T *testing.T) {
	T.Parallel()

	T.Run("returns non-empty mermaid diagram", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		res, err := adminClient.GetMermaidDiagramForRecipe(ctx, &mealplanninggrpc.GetMermaidDiagramForRecipeRequest{RecipeId: createdRecipe.ID})
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.NotEmpty(t, res.Response, "mermaid diagram should not be empty")
	})
}

func TestRecipes_Validation(T *testing.T) {
	T.Parallel()

	T.Run("minimum steps requirement", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)
		instrument := createValidInstrumentForTest(t)

		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation, instrument)

		// Create recipe with only 1 step (should fail)
		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
			},
		}

		_, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 2 steps")
	})

	T.Run("step requirements - only instruments", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)
		instrument := createValidInstrumentForTest(t)

		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation, instrument)

		// Create recipe with step that has only instruments (should succeed)
		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		_, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		assert.NoError(t, err)
	})

	T.Run("step requirements - only vessels", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)
		vessel := createValidVesselForTest(t)

		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
		vpv := createValidPreparationVesselWithEntitiesForTest(t, preparation, vessel)

		// Create recipe with step that has only vessels (should succeed)
		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					Vessels: []*mealplanning.RecipeStepVesselCreationRequestInput{
						{
							Name:                     "pot",
							ValidPreparationVesselID: &vpv.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation.ID,
					Vessels: []*mealplanning.RecipeStepVesselCreationRequestInput{
						{
							Name:                     "pan",
							ValidPreparationVesselID: &vpv.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		_, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		assert.NoError(t, err)
	})

	T.Run("step requirements - neither instruments nor vessels", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)

		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)

		// Create recipe with step that has neither instruments nor vessels (should fail)
		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					// No instruments or vessels
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation.ID,
					// No instruments or vessels
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		_, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one instrument or vessel")
	})

	T.Run("bridge table validation - invalid ValidIngredientPreparationID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		instrument := createValidInstrumentForTest(t)

		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, createValidIngredientForTest(t), measurementUnit)
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation, instrument)

		invalidID := nonexistentID

		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &invalidID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		_, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step 0 ingredient 0")
		assert.Contains(t, err.Error(), "ValidIngredientPreparation")
		assert.Contains(t, err.Error(), "not found")
	})

	T.Run("bridge table validation - mismatched preparation for ingredient", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation1 := createValidPreparationForTest(t)
		preparation2 := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)
		instrument := createValidInstrumentForTest(t)

		// Create VIP for preparation1, but use it in step with preparation2
		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation1)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation2, instrument)

		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation2.ID, // Different preparation
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID, // VIP is for preparation1
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation2.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		_, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step 0 ingredient 0")
		assert.Contains(t, err.Error(), "is for preparation")
		assert.Contains(t, err.Error(), "but step uses preparation")
	})

	T.Run("bridge table validation - mismatched preparation for instrument", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation1 := createValidPreparationForTest(t)
		preparation2 := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)
		instrument := createValidInstrumentForTest(t)

		// Create VPI for preparation1, but use it in step with preparation2
		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation2)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation1, instrument)

		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation2.ID, // Different preparation
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID, // VPI is for preparation1
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation2.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		_, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step 0 instrument 0")
		assert.Contains(t, err.Error(), "is for preparation")
		assert.Contains(t, err.Error(), "but step uses preparation")
	})

	T.Run("bridge table validation - mismatched ingredient for ValidIngredientMeasurementUnit", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient1 := createValidIngredientForTest(t)
		ingredient2 := createValidIngredientForTest(t)
		instrument := createValidInstrumentForTest(t)

		// Create VIP for ingredient1, but VIMU for ingredient2
		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient1, preparation)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient2, measurementUnit) // Different ingredient
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation, instrument)

		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,  // ingredient1
							ValidIngredientMeasurementUnitID: &vimu.ID, // ingredient2 - mismatch!
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		_, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step 0 ingredient 0")
		assert.Contains(t, err.Error(), "ValidIngredientMeasurementUnit")
		assert.Contains(t, err.Error(), "is for ingredient")
		assert.Contains(t, err.Error(), "but ingredient")
	})

	T.Run("required fields - missing name", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)
		instrument := createValidInstrumentForTest(t)

		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation, instrument)

		input := &mealplanning.RecipeCreationRequestInput{
			// Name missing
			Slug:                 "test-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		_, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		assert.Error(t, err)
	})

	T.Run("required fields - missing slug", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)
		instrument := createValidInstrumentForTest(t)

		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation, instrument)

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "test recipe",
			// Slug missing
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		_, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
		})
		assert.Error(t, err)
	})

	T.Run("component type validation - invalid type", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		preparation := createValidPreparationForTest(t)
		measurementUnit := createValidMeasurementUnitForTest(t)
		ingredient := createValidIngredientForTest(t)
		instrument := createValidInstrumentForTest(t)

		vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation)
		vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
		vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation, instrument)

		input := &mealplanning.RecipeCreationRequestInput{
			Name:                 "test recipe",
			Slug:                 "test-recipe",
			YieldsComponentType:  "invalid-type", // Invalid component type
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "knife",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "test ingredient",
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		// Test validation at the domain layer directly, since gRPC conversion
		// normalizes invalid component types to "unspecified" before validation
		err := input.ValidateWithContext(ctx)
		require.Error(t, err, "should reject invalid component type")
		// The validation error uses camelCase field name
		assert.Contains(t, err.Error(), "yieldsComponentType", "error should mention yieldsComponentType")
	})

	T.Run("component type validation - all valid types", func(t *testing.T) {
		t.Parallel()

		validTypes := []string{
			mealplanning.MealComponentTypesUnspecified,
			mealplanning.MealComponentTypesAmuseBouche,
			mealplanning.MealComponentTypesAppetizer,
			mealplanning.MealComponentTypesSoup,
			mealplanning.MealComponentTypesMain,
			mealplanning.MealComponentTypesSalad,
			mealplanning.MealComponentTypesBeverage,
			mealplanning.MealComponentTypesSide,
			mealplanning.MealComponentTypesDessert,
		}

		for _, componentType := range validTypes {
			t.Run(componentType, func(t *testing.T) {
				t.Parallel()
				ctx := t.Context()

				preparation := createValidPreparationForTest(t)
				measurementUnit := createValidMeasurementUnitForTest(t)
				ingredient := createValidIngredientForTest(t)
				instrument := createValidInstrumentForTest(t)

				vip := createValidIngredientPreparationWithEntitiesForTest(t, ingredient, preparation)
				vimu := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient, measurementUnit)
				vpi := createValidPreparationInstrumentWithEntitiesForTest(t, preparation, instrument)

				input := &mealplanning.RecipeCreationRequestInput{
					Name:                 fmt.Sprintf("test recipe %s", componentType),
					Slug:                 fmt.Sprintf("test-recipe-%s", componentType),
					YieldsComponentType:  componentType,
					PortionName:          "serving",
					PluralPortionName:    "servings",
					MinEstimatedPortions: 4,
					Steps: []*mealplanning.RecipeStepCreationRequestInput{
						{
							PreparationID: preparation.ID,
							Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
								{
									Name:                         "knife",
									ValidPreparationInstrumentID: &vpi.ID,
								},
							},
							Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
								{
									Name:                             "test ingredient",
									ValidIngredientPreparationID:     &vip.ID,
									ValidIngredientMeasurementUnitID: &vimu.ID,
									MinQuantity:                      1,
								},
							},
							Products: []*mealplanning.RecipeStepProductCreationRequestInput{
								{
									Name:                   "output",
									Type:                   mealplanning.RecipeStepProductIngredientType,
									MeasurementUnitID:      &measurementUnit.ID,
									MinMeasurementQuantity: new(float32(1)),
								},
							},
							Index: 0,
						},
						{
							PreparationID: preparation.ID,
							Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
								{
									Name:                         "spoon",
									ValidPreparationInstrumentID: &vpi.ID,
								},
							},
							Products: []*mealplanning.RecipeStepProductCreationRequestInput{
								{
									Name:                   "final output",
									Type:                   mealplanning.RecipeStepProductIngredientType,
									MeasurementUnitID:      &measurementUnit.ID,
									MinMeasurementQuantity: new(float32(1)),
								},
							},
							Index: 1,
						},
					},
				}

				created, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
					Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input),
				})
				require.NoError(t, err, "component type %s should be valid", componentType)
				require.NotNil(t, created)

				// Cleanup
				_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: created.Created.Id})
				assert.NoError(t, err)
			})
		}
	})
}

func TestRecipes_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.Error(t, err)
	})

	T.Run("nonexistent recipe", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		_, _, createdRecipe := createRecipeForTest(t, nil)

		_, err := testClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.Error(t, err)
	})
}

func TestRecipes_OwnershipDenial(T *testing.T) {
	T.Parallel()

	T.Run("non-owner cannot update recipe", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, otherUserClient := createUserAndClientForTest(t)
		_, _, createdRecipe := createRecipeForTest(t, nil)

		name := "updated name"
		_, err := otherUserClient.UpdateRecipe(ctx, &mealplanninggrpc.UpdateRecipeRequest{
			RecipeId: createdRecipe.ID,
			Input: &mealplanninggrpc.RecipeUpdateRequestInput{
				Name: &name,
			},
		})
		require.Error(t, err)
	})

	T.Run("non-owner cannot archive recipe", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, otherUserClient := createUserAndClientForTest(t)
		_, _, createdRecipe := createRecipeForTest(t, nil)

		_, err := otherUserClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		require.Error(t, err)
	})
}

func TestRecipes_Listing(T *testing.T) {
	T.Parallel()

	T.Run("should be readable in paginated form", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		actual, err := adminClient.GetRecipes(ctx, &mealplanninggrpc.GetRecipesRequest{})
		require.NoError(t, err)
		require.NotNil(t, actual)

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		recipes, err := c.GetRecipes(ctx, &mealplanninggrpc.GetRecipesRequest{})
		assert.Error(t, err)
		assert.Nil(t, recipes)
	})
}

func TestRecipes_SearchForMealEligibleRecipes(T *testing.T) {
	T.Parallel()

	T.Run("should return meal-eligible recipes matching query", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// just verify the endpoint responds without error
		actual, err := adminClient.SearchForMealEligibleRecipes(ctx, &mealplanninggrpc.SearchForMealEligibleRecipesRequest{
			Query: "test",
		})
		require.NoError(t, err)
		require.NotNil(t, actual)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		results, err := c.SearchForMealEligibleRecipes(ctx, &mealplanninggrpc.SearchForMealEligibleRecipesRequest{
			Query: "test",
		})
		assert.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestRecipes_EstimateRecipePrepTasks(T *testing.T) {
	T.Parallel()

	T.Run("should return prep task estimates for a recipe", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		actual, err := adminClient.EstimateRecipePrepTasks(ctx, &mealplanninggrpc.EstimateRecipePrepTasksRequest{
			RecipeId: createdRecipe.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, actual)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		c := buildUnauthenticatedGRPCClientForTest(t)
		results, err := c.EstimateRecipePrepTasks(ctx, &mealplanninggrpc.EstimateRecipePrepTasksRequest{
			RecipeId: createdRecipe.ID,
		})
		assert.Error(t, err)
		assert.Nil(t, results)
	})

	T.Run("nonexistent recipe", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.EstimateRecipePrepTasks(ctx, &mealplanninggrpc.EstimateRecipePrepTasksRequest{
			RecipeId: nonexistentID,
		})
		assert.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestRecipes_AssociatedRecipes(T *testing.T) {
	T.Parallel()

	T.Run("recipe with cross-recipe product reference populates AssociatedRecipes", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Create valid entities for the first recipe
		preparation1 := createValidPreparationForTest(t)
		measurementUnit1 := createValidMeasurementUnitForTest(t)
		ingredient1 := createValidIngredientForTest(t)
		instrument1 := createValidInstrumentForTest(t)

		vip1 := createValidIngredientPreparationWithEntitiesForTest(t, ingredient1, preparation1)
		vimu1 := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient1, measurementUnit1)
		vpi1 := createValidPreparationInstrumentWithEntitiesForTest(t, preparation1, instrument1)

		// Create first recipe with a product in step 0
		firstRecipeInput := &mealplanning.RecipeCreationRequestInput{
			Name:                 "Base Sauce Recipe",
			Slug:                 "base-sauce-recipe",
			YieldsComponentType:  mealplanning.MealComponentTypesUnspecified,
			PortionName:          "cup",
			PluralPortionName:    "cups",
			MinEstimatedPortions: 1,
			EligibleForMeals:     false,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation1.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi1.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "base ingredient",
							ValidIngredientPreparationID:     &vip1.ID,
							ValidIngredientMeasurementUnitID: &vimu1.ID,
							MinQuantity:                      1,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "base sauce",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit1.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation1.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi1.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final sauce",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit1.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		// Create the first recipe
		firstRecipeCreated, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(firstRecipeInput),
		})
		require.NoError(t, err)
		firstRecipe := converters.ConvertGRPCRecipeToRecipe(firstRecipeCreated.Created)

		// Create valid entities for the second recipe
		preparation2 := createValidPreparationForTest(t)
		measurementUnit2 := createValidMeasurementUnitForTest(t)
		ingredient2 := createValidIngredientForTest(t)
		instrument2 := createValidInstrumentForTest(t)

		vip2 := createValidIngredientPreparationWithEntitiesForTest(t, ingredient2, preparation2)
		vimu2 := createValidIngredientMeasurementUnitWithEntitiesForTest(t, ingredient2, measurementUnit2)
		vpi2 := createValidPreparationInstrumentWithEntitiesForTest(t, preparation2, instrument2)

		// Create second recipe that references the product from the first recipe
		secondRecipeInput := &mealplanning.RecipeCreationRequestInput{
			Name:                 "Recipe Using Base Sauce",
			Slug:                 "recipe-using-base-sauce",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			EligibleForMeals:     true,
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					PreparationID: preparation2.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi2.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                             "other ingredient",
							ValidIngredientPreparationID:     &vip2.ID,
							ValidIngredientMeasurementUnitID: &vimu2.ID,
							MinQuantity:                      1,
						},
						{
							// Reference the product from the first recipe
							// The product "base sauce" is from step 0 (index 0), product index 0
							ProductOfRecipeStepIndex:        new(uint64(0)),
							ProductOfRecipeStepProductIndex: new(uint64(0)),
							RecipeStepProductRecipeID:       &firstRecipe.ID,
							Name:                            "base sauce",
							MinQuantity:                     0.5,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final dish",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit2.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					PreparationID: preparation2.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi2.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductCreationRequestInput{
						{
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit2.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		// Create the second recipe
		secondRecipeCreated, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{
			Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(secondRecipeInput),
		})
		require.NoError(t, err)
		secondRecipe := converters.ConvertGRPCRecipeToRecipe(secondRecipeCreated.Created)

		// Retrieve the second recipe and validate AssociatedRecipes
		retrievedSecond, err := adminClient.GetRecipe(ctx, &mealplanninggrpc.GetRecipeRequest{
			RecipeId: secondRecipe.ID,
		})
		require.NoError(t, err)
		retrievedSecondRecipe := converters.ConvertGRPCRecipeToRecipe(retrievedSecond.Result)

		// Validate that the first recipe is in the second's AssociatedRecipes
		require.Len(t, retrievedSecondRecipe.AssociatedRecipes, 1, "second recipe should have one associated recipe")
		assert.Equal(t, firstRecipe.ID, retrievedSecondRecipe.AssociatedRecipes[0].ID, "associated recipe should be the first recipe")
		assert.Equal(t, firstRecipe.Name, retrievedSecondRecipe.AssociatedRecipes[0].Name, "associated recipe should have the correct name")

		// Retrieve the first recipe and validate it has empty AssociatedRecipes
		retrievedFirst, err := adminClient.GetRecipe(ctx, &mealplanninggrpc.GetRecipeRequest{
			RecipeId: firstRecipe.ID,
		})
		require.NoError(t, err)
		retrievedFirstRecipe := converters.ConvertGRPCRecipeToRecipe(retrievedFirst.Result)

		assert.Empty(t, retrievedFirstRecipe.AssociatedRecipes, "first recipe should have no associated recipes")

		// Cleanup
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: secondRecipe.ID})
		assert.NoError(t, err)
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: firstRecipe.ID})
		assert.NoError(t, err)
	})
}
