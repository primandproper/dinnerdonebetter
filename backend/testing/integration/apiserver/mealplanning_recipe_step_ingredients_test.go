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

func checkRecipeStepIngredientSliceEquality(t *testing.T, stepIndex int, expected, actual []*mealplanning.RecipeStepIngredient) {
	t.Helper()
	require.Equal(t, len(expected), len(actual), "expected recipe step %d ingredients length", stepIndex)
	for i := range expected {
		checkRecipeStepIngredientEquality(t, stepIndex, i, expected[i], actual[i])
	}
}

func checkRecipeStepIngredientEquality(t *testing.T, stepIndex, ingIndex int, expected, actual *mealplanning.RecipeStepIngredient) {
	t.Helper()
	assert.NotEmpty(t, actual.ID, "expected step %d ingredient %d to have ID", stepIndex, ingIndex)
	assert.False(t, actual.CreatedAt.IsZero(), "expected step %d ingredient %d to have CreatedAt", stepIndex, ingIndex)
	assert.NotEmpty(t, actual.BelongsToRecipeStep, "expected step %d ingredient %d to have BelongsToRecipeStep", stepIndex, ingIndex)
	assert.Equal(t, expected.Name, actual.Name, "expected step %d ingredient %d Name", stepIndex, ingIndex)
	assert.Equal(t, expected.MinQuantity, actual.MinQuantity, "expected step %d ingredient %d MinQuantity", stepIndex, ingIndex)
	assert.Equal(t, expected.MaxQuantity, actual.MaxQuantity, "expected step %d ingredient %d MaxQuantity", stepIndex, ingIndex)
	assert.Equal(t, expected.QuantityNotes, actual.QuantityNotes, "expected step %d ingredient %d QuantityNotes", stepIndex, ingIndex)
	assert.Equal(t, expected.IngredientNotes, actual.IngredientNotes, "expected step %d ingredient %d IngredientNotes", stepIndex, ingIndex)
	assert.Equal(t, expected.Index, actual.Index, "expected step %d ingredient %d Index", stepIndex, ingIndex)
	assert.Equal(t, expected.OptionIndex, actual.OptionIndex, "expected step %d ingredient %d OptionIndex", stepIndex, ingIndex)
	assert.Equal(t, expected.Optional, actual.Optional, "expected step %d ingredient %d Optional", stepIndex, ingIndex)
	assert.Equal(t, expected.ToTaste, actual.ToTaste, "expected step %d ingredient %d ToTaste", stepIndex, ingIndex)
	if expected.VesselIndex != nil {
		require.NotNil(t, actual.VesselIndex, "expected step %d ingredient %d VesselIndex non-nil", stepIndex, ingIndex)
		assert.Equal(t, *expected.VesselIndex, *actual.VesselIndex, "expected step %d ingredient %d VesselIndex", stepIndex, ingIndex)
	}
	if expected.ProductPercentageToUse != nil {
		require.NotNil(t, actual.ProductPercentageToUse, "expected step %d ingredient %d ProductPercentageToUse non-nil", stepIndex, ingIndex)
		assert.Equal(t, *expected.ProductPercentageToUse, *actual.ProductPercentageToUse, "expected step %d ingredient %d ProductPercentageToUse", stepIndex, ingIndex)
	}
	// MeasurementUnit comparison by ID (and ranges already compared above)
	assert.Equal(t, expected.MeasurementUnit.ID, actual.MeasurementUnit.ID, "expected step %d ingredient %d MeasurementUnit.ID", stepIndex, ingIndex)
	// Ingredient pointer may be nil if this ingredient refers to a product of a prior step
	if expected.Ingredient != nil {
		require.NotNil(t, actual.Ingredient, "expected step %d ingredient %d Ingredient non-nil", stepIndex, ingIndex)
		assert.Equal(t, expected.Ingredient.ID, actual.Ingredient.ID, "expected step %d ingredient %d Ingredient.ID", stepIndex, ingIndex)
	}
}

func TestRecipeStepIngredients_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("should CRUD", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		firstStep := createdRecipe.Steps[0]
		createdRecipeStepID := firstStep.ID
		createdRecipeStepIngredientID := firstStep.Ingredients[0].ID

		require.NotEmpty(t, createdRecipeStepID, "created recipe step ID must not be empty")
		require.NotEmpty(t, createdRecipeStepIngredientID, "created recipe step ingredient ID must not be empty")

		createdRecipeStepIngredientRes, err := adminClient.GetRecipeStepIngredient(ctx, &mealplanninggrpc.GetRecipeStepIngredientRequest{
			RecipeId:               createdRecipe.ID,
			RecipeStepId:           createdRecipeStepID,
			RecipeStepIngredientId: createdRecipeStepIngredientID,
		})
		require.NotNil(t, createdRecipeStepIngredientRes)
		require.NoError(t, err)

		createdRecipeStepIngredient := converters.ConvertGRPCRecipeStepIngredientToRecipeStepIngredient(createdRecipeStepIngredientRes.Result)

		createdValidIngredient := createValidIngredientForTest(t)

		newRecipeStepIngredient := fakes.BuildFakeRecipeStepIngredient()
		newRecipeStepIngredient.BelongsToRecipeStep = createdRecipeStepID
		newRecipeStepIngredient.ID = createdRecipeStepIngredientID
		newRecipeStepIngredient.Ingredient = createdValidIngredient
		newRecipeStepIngredient.MeasurementUnit = createdRecipeStepIngredient.MeasurementUnit
		// Preserve the existing Index and OptionIndex to avoid conflicts
		newRecipeStepIngredient.Index = createdRecipeStepIngredient.Index
		newRecipeStepIngredient.OptionIndex = createdRecipeStepIngredient.OptionIndex

		updateInput := mpconverters.ConvertRecipeStepIngredientToRecipeStepIngredientUpdateRequestInput(newRecipeStepIngredient)
		createdRecipeStepIngredient.Update(updateInput)

		_, err = adminClient.UpdateRecipeStepIngredient(ctx, &mealplanninggrpc.UpdateRecipeStepIngredientRequest{
			RecipeId:               createdRecipe.ID,
			RecipeStepId:           createdRecipeStepID,
			RecipeStepIngredientId: createdRecipeStepIngredientID,
			Input:                  converters.ConvertRecipeStepIngredientUpdateRequestInputToGRPCRecipeStepIngredientUpdateRequestInput(updateInput),
		})
		require.NoError(t, err)

		retrievedRes, err := adminClient.GetRecipeStepIngredient(ctx, &mealplanninggrpc.GetRecipeStepIngredientRequest{
			RecipeId:               createdRecipe.ID,
			RecipeStepId:           createdRecipeStepID,
			RecipeStepIngredientId: createdRecipeStepIngredientID,
		})
		require.NotNil(t, retrievedRes)
		require.NoError(t, err)

		actual := converters.ConvertGRPCRecipeStepIngredientToRecipeStepIngredient(retrievedRes.Result)

		// assert recipe step ingredient equality
		checkRecipeStepIngredientEquality(t, -1, -1, newRecipeStepIngredient, actual)
		assert.NotNil(t, actual.LastUpdatedAt)

		_, err = adminClient.ArchiveRecipeStepIngredient(ctx, &mealplanninggrpc.ArchiveRecipeStepIngredientRequest{
			RecipeId:               createdRecipe.ID,
			RecipeStepId:           createdRecipeStepID,
			RecipeStepIngredientId: createdRecipeStepIngredientID,
		})
		assert.NoError(t, err)

		_, err = adminClient.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{RecipeId: createdRecipe.ID, RecipeStepId: createdRecipeStepID})
		assert.NoError(t, err)

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipeStepIngredients_Listing(T *testing.T) {
	T.Parallel()

	T.Run("should be readable in paginated form", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		var (
			createdRecipeStepID            string
			createdRecipeStepPreparationID string
		)
		for _, step := range createdRecipe.Steps {
			createdRecipeStepID = step.ID
			createdRecipeStepPreparationID = step.Preparation.ID
			break
		}

		createdValidMeasurementUnit := createValidMeasurementUnitForTest(t)

		// Get existing ingredients to determine the next available index
		existingIngredientsRes, err := adminClient.GetRecipeStepIngredients(ctx, &mealplanninggrpc.GetRecipeStepIngredientsRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
		})
		require.NoError(t, err)
		existingIngredients := existingIngredientsRes.Results
		nextIndex := uint16(len(existingIngredients)) // Start from the next index after existing ones

		var expected []*mealplanning.RecipeStepIngredient
		for i := range 5 {
			x, _, _ := createRecipeForTest(t, nil)

			// Create bridge table entries for this ingredient
			createdValidPreparation := &mealplanning.ValidPreparation{ID: createdRecipeStepPreparationID}
			createdValidIngredient := &mealplanning.ValidIngredient{ID: x[0].ID}
			createdVIP := createValidIngredientPreparationWithEntitiesForTest(t, createdValidIngredient, createdValidPreparation)
			createdVIMU := createValidIngredientMeasurementUnitWithEntitiesForTest(t, createdValidIngredient, createdValidMeasurementUnit)

			exampleRecipeStepIngredient := fakes.BuildFakeRecipeStepIngredient()
			exampleRecipeStepIngredient.BelongsToRecipeStep = createdRecipeStepID
			exampleRecipeStepIngredient.Ingredient = createdValidIngredient
			exampleRecipeStepIngredient.MeasurementUnit = mealplanning.ValidMeasurementUnit{ID: createdValidMeasurementUnit.ID}
			// Set Index to match what we'll use in the creation input - start from nextIndex to avoid conflicts
			ingredientIndex := nextIndex + uint16(i)
			exampleRecipeStepIngredient.Index = ingredientIndex

			exampleRecipeStepIngredientInput := mpconverters.ConvertRecipeStepIngredientToRecipeStepIngredientCreationRequestInput(exampleRecipeStepIngredient)
			// Set bridge table IDs (required)
			exampleRecipeStepIngredientInput.ValidIngredientPreparationID = &createdVIP.ID
			exampleRecipeStepIngredientInput.ValidIngredientMeasurementUnitID = &createdVIMU.ID
			// Set Index (required for individual creation) - use nextIndex + loop index to ensure uniqueness
			exampleRecipeStepIngredientInput.Index = new(ingredientIndex)
			createdRecipeStepIngredientRes, createErr := adminClient.CreateRecipeStepIngredient(ctx, &mealplanninggrpc.CreateRecipeStepIngredientRequest{
				RecipeId:     createdRecipe.ID,
				RecipeStepId: createdRecipeStepID,
				Input:        converters.ConvertRecipeStepIngredientCreationRequestInputToGRPCRecipeStepIngredientCreationRequestInput(exampleRecipeStepIngredientInput),
			})
			require.NoError(t, createErr)

			createdRecipeStepIngredient := converters.ConvertGRPCRecipeStepIngredientToRecipeStepIngredient(createdRecipeStepIngredientRes.Created)
			checkRecipeStepIngredientEquality(t, -1, -1, exampleRecipeStepIngredient, createdRecipeStepIngredient)

			retrievedRecipeStepIngredientRes, getErr := adminClient.GetRecipeStepIngredient(ctx, &mealplanninggrpc.GetRecipeStepIngredientRequest{
				RecipeId:               createdRecipe.ID,
				RecipeStepId:           createdRecipeStepID,
				RecipeStepIngredientId: createdRecipeStepIngredient.ID,
			})
			require.NotNil(t, retrievedRecipeStepIngredientRes.Result)
			require.NoError(t, getErr)
			require.Equal(t, createdRecipeStepID, createdRecipeStepIngredient.BelongsToRecipeStep)

			expected = append(expected, createdRecipeStepIngredient)
		}

		// assert recipe step ingredient list equality
		actual, err := adminClient.GetRecipeStepIngredients(ctx, &mealplanninggrpc.GetRecipeStepIngredientsRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
		})
		require.NotNil(t, actual)
		require.NoError(t, err)
		assert.True(
			t,
			len(expected) <= len(actual.Results),
			"expected %d to be <= %d",
			len(expected),
			len(actual.Results),
		)

		for _, createdRecipeStepIngredient := range expected {
			_, err = adminClient.ArchiveRecipeStepIngredient(ctx, &mealplanninggrpc.ArchiveRecipeStepIngredientRequest{
				RecipeId:               createdRecipe.ID,
				RecipeStepId:           createdRecipeStepID,
				RecipeStepIngredientId: createdRecipeStepIngredient.ID,
			})
			assert.NoError(t, err)
		}

		_, err = adminClient.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
		})
		assert.NoError(t, err)

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipeStepIngredients_OwnershipDenial(T *testing.T) {
	T.Parallel()

	T.Run("Create", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, otherUserClient := createUserAndClientForTest(t)

		createdRecipeStep := createdRecipe.Steps[0]

		exampleRecipeStepIngredient := fakes.BuildFakeRecipeStepIngredient()
		exampleRecipeStepIngredient.BelongsToRecipeStep = createdRecipeStep.ID
		exampleInput := mpconverters.ConvertRecipeStepIngredientToRecipeStepIngredientCreationRequestInput(exampleRecipeStepIngredient)

		_, err := otherUserClient.CreateRecipeStepIngredient(ctx, &mealplanninggrpc.CreateRecipeStepIngredientRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStep.ID,
			Input:        converters.ConvertRecipeStepIngredientCreationRequestInputToGRPCRecipeStepIngredientCreationRequestInput(exampleInput),
		})
		require.Error(t, err)
	})

	T.Run("Update", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, otherUserClient := createUserAndClientForTest(t)

		createdRecipeStep := createdRecipe.Steps[0]
		existingIngredient := createdRecipeStep.Ingredients[0]

		updateInput := mpconverters.ConvertRecipeStepIngredientToRecipeStepIngredientUpdateRequestInput(existingIngredient)

		_, err := otherUserClient.UpdateRecipeStepIngredient(ctx, &mealplanninggrpc.UpdateRecipeStepIngredientRequest{
			RecipeId:               createdRecipe.ID,
			RecipeStepId:           createdRecipeStep.ID,
			RecipeStepIngredientId: existingIngredient.ID,
			Input:                  converters.ConvertRecipeStepIngredientUpdateRequestInputToGRPCRecipeStepIngredientUpdateRequestInput(updateInput),
		})
		require.Error(t, err)
	})

	T.Run("Archive", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, otherUserClient := createUserAndClientForTest(t)

		createdRecipeStep := createdRecipe.Steps[0]
		existingIngredient := createdRecipeStep.Ingredients[0]

		_, err := otherUserClient.ArchiveRecipeStepIngredient(ctx, &mealplanninggrpc.ArchiveRecipeStepIngredientRequest{
			RecipeId:               createdRecipe.ID,
			RecipeStepId:           createdRecipeStep.ID,
			RecipeStepIngredientId: existingIngredient.ID,
		})
		require.Error(t, err)
	})
}
