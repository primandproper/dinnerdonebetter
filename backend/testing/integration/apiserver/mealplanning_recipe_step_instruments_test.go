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

func checkRecipeStepInstrumentSliceEquality(t *testing.T, stepIndex int, expected, actual []*mealplanning.RecipeStepInstrument) {
	t.Helper()
	require.Equal(t, len(expected), len(actual), "expected recipe step %d instruments length", stepIndex)
	for i := range expected {
		checkRecipeStepInstrumentEquality(t, stepIndex, i, expected[i], actual[i])
	}
}

func checkRecipeStepInstrumentEquality(t *testing.T, stepIndex, instrIndex int, expected, actual *mealplanning.RecipeStepInstrument) {
	t.Helper()
	assert.NotEmpty(t, actual.ID, "expected step %d instrument %d to have ID", stepIndex, instrIndex)
	assert.False(t, actual.CreatedAt.IsZero(), "expected step %d instrument %d to have CreatedAt", stepIndex, instrIndex)
	assert.NotEmpty(t, actual.BelongsToRecipeStep, "expected step %d instrument %d to have BelongsToRecipeStep", stepIndex, instrIndex)
	assert.Equal(t, expected.Name, actual.Name, "expected step %d instrument %d Name", stepIndex, instrIndex)
	assert.Equal(t, expected.Notes, actual.Notes, "expected step %d instrument %d Notes", stepIndex, instrIndex)
	assert.Equal(t, expected.MinQuantity, actual.MinQuantity, "expected step %d instrument %d MinQuantity", stepIndex, instrIndex)
	assert.Equal(t, expected.MaxQuantity, actual.MaxQuantity, "expected step %d instrument %d MaxQuantity", stepIndex, instrIndex)
	assert.Equal(t, expected.Index, actual.Index, "expected step %d instrument %d Index", stepIndex, instrIndex)
	assert.Equal(t, expected.OptionIndex, actual.OptionIndex, "expected step %d instrument %d OptionIndex", stepIndex, instrIndex)
	assert.Equal(t, expected.PreferenceRank, actual.PreferenceRank, "expected step %d instrument %d PreferenceRank", stepIndex, instrIndex)
	assert.Equal(t, expected.Optional, actual.Optional, "expected step %d instrument %d Optional", stepIndex, instrIndex)
	if expected.Instrument != nil {
		require.NotNil(t, actual.Instrument, "expected step %d instrument %d Instrument non-nil", stepIndex, instrIndex)
		assert.NotEmpty(t, actual.Instrument.ID, "expected step %d instrument %d Instrument.ID", stepIndex, instrIndex)
		assert.Equal(t, expected.Instrument.ID, actual.Instrument.ID, "expected step %d instrument %d Instrument.ID", stepIndex, instrIndex)
	}
}

func TestRecipeStepInstruments_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("CRUD", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		var createdRecipeStepID string
		var createdRecipeStepPreparationID string
		for _, step := range createdRecipe.Steps {
			createdRecipeStepID = step.ID
			createdRecipeStepPreparationID = step.Preparation.ID
			break
		}

		createdValidInstrument := createValidInstrumentForTest(t)

		// Create bridge table entry for preparation+instrument
		createdValidPreparation := &mealplanning.ValidPreparation{ID: createdRecipeStepPreparationID}
		createdValidPreparationInstrument := createValidPreparationInstrumentWithEntitiesForTest(t, createdValidPreparation, createdValidInstrument)

		// Get existing instruments to determine the next available index
		existingInstrumentsRes, err := adminClient.GetRecipeStepInstruments(ctx, &mealplanninggrpc.GetRecipeStepInstrumentsRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
		})
		require.NoError(t, err)
		existingInstruments := existingInstrumentsRes.Results
		nextIndex := uint16(len(existingInstruments)) // Start from the next index after existing ones

		exampleRecipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()
		exampleRecipeStepInstrument.BelongsToRecipeStep = createdRecipeStepID
		exampleRecipeStepInstrument.Instrument = &mealplanning.ValidInstrument{ID: createdValidInstrument.ID}
		// Set Index to match what we'll use in the creation input - use nextIndex to avoid conflicts
		exampleRecipeStepInstrument.Index = nextIndex
		exampleRecipeStepInstrumentInput := mpconverters.ConvertRecipeStepInstrumentToRecipeStepInstrumentCreationRequestInput(exampleRecipeStepInstrument)
		// Set bridge table ID (required)
		exampleRecipeStepInstrumentInput.ValidPreparationInstrumentID = &createdValidPreparationInstrument.ID
		// Set Index (required for individual creation) - use nextIndex to ensure uniqueness
		exampleRecipeStepInstrumentInput.Index = new(nextIndex)
		createdRecipeStepInstrumentRes, err := adminClient.CreateRecipeStepInstrument(ctx, &mealplanninggrpc.CreateRecipeStepInstrumentRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
			Input:        converters.ConvertRecipeStepInstrumentCreationRequestInputToGRPCRecipeStepInstrumentCreationRequestInput(exampleRecipeStepInstrumentInput),
		})
		require.NoError(t, err)

		createdRecipeStepInstrument := converters.ConvertGRPCRecipeStepInstrumentToRecipeStepInstrument(createdRecipeStepInstrumentRes.Created)
		checkRecipeStepInstrumentEquality(t, -1, -1, exampleRecipeStepInstrument, createdRecipeStepInstrument)

		retrievedRecipeStepInstrumentRes, err := adminClient.GetRecipeStepInstrument(ctx, &mealplanninggrpc.GetRecipeStepInstrumentRequest{
			RecipeId:               createdRecipe.ID,
			RecipeStepId:           createdRecipeStepID,
			RecipeStepInstrumentId: createdRecipeStepInstrument.ID,
		})
		require.NoError(t, err)

		createdRecipeStepInstrument = converters.ConvertGRPCRecipeStepInstrumentToRecipeStepInstrument(retrievedRecipeStepInstrumentRes.Result)
		checkRecipeStepInstrumentEquality(t, -1, -1, exampleRecipeStepInstrument, createdRecipeStepInstrument)

		require.Equal(t, createdRecipeStepID, createdRecipeStepInstrument.BelongsToRecipeStep)
		exampleRecipeStepInstrument.Instrument = createdValidInstrument
		exampleRecipeStepInstrument.Instrument.CreatedAt = createdRecipeStepInstrument.Instrument.CreatedAt

		checkRecipeStepInstrumentEquality(t, -1, -1, exampleRecipeStepInstrument, createdRecipeStepInstrument)

		newValidInstrument := createValidInstrumentForTest(t)

		newRecipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()
		newRecipeStepInstrument.BelongsToRecipeStep = createdRecipeStepID
		newRecipeStepInstrument.Instrument = newValidInstrument
		// Preserve the existing Index and OptionIndex to avoid conflicts
		newRecipeStepInstrument.Index = createdRecipeStepInstrument.Index
		newRecipeStepInstrument.OptionIndex = createdRecipeStepInstrument.OptionIndex
		updateInput := mpconverters.ConvertRecipeStepInstrumentToRecipeStepInstrumentUpdateRequestInput(newRecipeStepInstrument)
		createdRecipeStepInstrument.Update(updateInput)

		_, err = adminClient.UpdateRecipeStepInstrument(ctx, &mealplanninggrpc.UpdateRecipeStepInstrumentRequest{
			RecipeId:               createdRecipe.ID,
			RecipeStepId:           createdRecipeStepID,
			RecipeStepInstrumentId: createdRecipeStepInstrument.ID,
			Input:                  converters.ConvertRecipeStepInstrumentUpdateRequestInputToGRPCRecipeStepInstrumentUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)

		retrievedRecipeStepInstrumentRes, err = adminClient.GetRecipeStepInstrument(ctx, &mealplanninggrpc.GetRecipeStepInstrumentRequest{
			RecipeId:               createdRecipe.ID,
			RecipeStepId:           createdRecipeStepID,
			RecipeStepInstrumentId: createdRecipeStepInstrument.ID,
		})
		require.NoError(t, err)

		actual := converters.ConvertGRPCRecipeStepInstrumentToRecipeStepInstrument(retrievedRecipeStepInstrumentRes.Result)

		// assert recipe step instrument equality
		checkRecipeStepInstrumentEquality(t, -1, -1, newRecipeStepInstrument, actual)
		assert.NotNil(t, actual.LastUpdatedAt)

		_, err = adminClient.ArchiveRecipeStepInstrument(ctx, &mealplanninggrpc.ArchiveRecipeStepInstrumentRequest{
			RecipeId:               createdRecipe.ID,
			RecipeStepId:           createdRecipeStepID,
			RecipeStepInstrumentId: createdRecipeStepInstrument.ID,
		})
		assert.NoError(t, err)

		_, err = adminClient.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{RecipeId: createdRecipe.ID, RecipeStepId: createdRecipeStepID})
		assert.NoError(t, err)

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipeStepInstruments_AsRecipeStepProducts(T *testing.T) {
	T.Parallel()

	T.Run("should be able to use a recipe step instrument that was the product of a prior recipe step", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		knife := createValidInstrumentForTest(t)
		preheat := createValidPreparationForTest(t)
		cut := createValidPreparationForTest(t)
		stick := createValidMeasurementUnitForTest(t)
		unit := createValidMeasurementUnitForTest(t)
		butter := createValidIngredientForTest(t)

		// Create bridge table entries
		vpiKnifePreheat := createValidPreparationInstrumentWithEntitiesForTest(t, preheat, knife)
		vipButterCut := createValidIngredientPreparationWithEntitiesForTest(t, butter, cut)
		vimuButterStick := createValidIngredientMeasurementUnitWithEntitiesForTest(t, butter, stick)

		preheatedKnife := "preheated knife"

		expected := &mealplanning.Recipe{
			Name:                 t.Name(),
			Slug:                 "whatever-who-cares",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          t.Name(),
			PluralPortionName:    t.Name(),
			MinEstimatedPortions: 1,

			MaxEstimatedPortions: nil,
			Steps: []*mealplanning.RecipeStep{
				{
					Products: []*mealplanning.RecipeStepProduct{
						{
							Name:                   preheatedKnife,
							Type:                   mealplanning.RecipeStepProductInstrumentType,
							MeasurementUnit:        unit,
							QuantityNotes:          "",
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Notes:       "first step",
					Preparation: *preheat,
					Ingredients: []*mealplanning.RecipeStepIngredient{},
					Instruments: []*mealplanning.RecipeStepInstrument{
						{
							Instrument: knife,
						},
					},
					Index: 0,
				},
				{
					Preparation: *cut,
					Instruments: []*mealplanning.RecipeStepInstrument{
						{
							Name:       preheatedKnife,
							Instrument: nil,
						},
					},
					Products: []*mealplanning.RecipeStepProduct{
						{
							Name:                   "cut butter",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnit:        stick,
							QuantityNotes:          "",
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Notes: "second step",
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Ingredient:      butter,
							Name:            "butter",
							MeasurementUnit: *stick,
							MinQuantity:     1,

							MaxQuantity: nil,
						},
					},
					Index: 1,
				},
			},
		}

		exampleRecipeInput := mpconverters.ConvertRecipeToRecipeCreationRequestInput(expected)
		exampleRecipeInput.Steps[1].Instruments[0].ProductOfRecipeStepIndex = new(uint64(0))
		exampleRecipeInput.Steps[1].Instruments[0].ProductOfRecipeStepProductIndex = new(uint64(0))

		// Set bridge table IDs
		// Step 0: knife with preheat preparation
		exampleRecipeInput.Steps[0].Instruments[0].ValidPreparationInstrumentID = &vpiKnifePreheat.ID
		// Step 1: butter ingredient with cut preparation (the instrument is a recipe step product, no bridge ID needed)
		exampleRecipeInput.Steps[1].Ingredients[0].ValidIngredientPreparationID = &vipButterCut.ID
		exampleRecipeInput.Steps[1].Ingredients[0].ValidIngredientMeasurementUnitID = &vimuButterStick.ID

		createdRes, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(exampleRecipeInput)})
		require.NoError(t, err)

		created := converters.ConvertGRPCRecipeToRecipe(createdRes.Created)
		checkRecipeEquality(t, expected, created)

		fetchedRes, err := userClient.GetRecipe(ctx, &mealplanninggrpc.GetRecipeRequest{RecipeId: created.ID})
		require.NoError(t, err)

		created = converters.ConvertGRPCRecipeToRecipe(fetchedRes.Result)
		checkRecipeEquality(t, expected, created)

		recipeStepProductIndex := -1
		for i, ingredient := range created.Steps[1].Instruments {
			if ingredient.RecipeStepProductID != nil {
				recipeStepProductIndex = i
			}
		}

		require.NotEqual(t, -1, recipeStepProductIndex)
		require.NotNil(t, created.Steps[1].Instruments[recipeStepProductIndex].RecipeStepProductID)
		assert.Equal(t, created.Steps[0].Products[0].ID, *created.Steps[1].Instruments[recipeStepProductIndex].RecipeStepProductID)
	})
}

func TestRecipeStepInstruments_Listing(T *testing.T) {
	T.Parallel()

	T.Run("should be readable in paginated form", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		var createdRecipeStepID string
		var createdRecipeStepPreparationID string
		for _, step := range createdRecipe.Steps {
			createdRecipeStepID = step.ID
			createdRecipeStepPreparationID = step.Preparation.ID
			break
		}

		createdValidInstrument := createValidInstrumentForTest(t)

		// Create bridge table entry for preparation+instrument
		createdValidPreparation := &mealplanning.ValidPreparation{ID: createdRecipeStepPreparationID}
		createdValidPreparationInstrument := createValidPreparationInstrumentWithEntitiesForTest(t, createdValidPreparation, createdValidInstrument)

		// Get existing instruments to determine the next available index
		existingInstrumentsRes, err := adminClient.GetRecipeStepInstruments(ctx, &mealplanninggrpc.GetRecipeStepInstrumentsRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
		})
		require.NoError(t, err)
		existingInstruments := existingInstrumentsRes.Results
		nextIndex := uint16(len(existingInstruments)) // Start from the next index after existing ones

		var expected []*mealplanning.RecipeStepInstrument
		for i := range 5 {
			exampleRecipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()
			exampleRecipeStepInstrument.BelongsToRecipeStep = createdRecipeStepID
			exampleRecipeStepInstrument.Instrument = &mealplanning.ValidInstrument{ID: createdValidInstrument.ID}
			// Set Index to match what we'll use in the creation input - start from nextIndex to avoid conflicts
			instrumentIndex := nextIndex + uint16(i)
			exampleRecipeStepInstrument.Index = instrumentIndex
			exampleRecipeStepInstrumentInput := mpconverters.ConvertRecipeStepInstrumentToRecipeStepInstrumentCreationRequestInput(exampleRecipeStepInstrument)
			// Set bridge table ID (required)
			exampleRecipeStepInstrumentInput.ValidPreparationInstrumentID = &createdValidPreparationInstrument.ID
			// Set Index (required for individual creation) - use nextIndex + loop index to ensure uniqueness
			exampleRecipeStepInstrumentInput.Index = new(instrumentIndex)
			createdRecipeStepInstrumentRes, createErr := adminClient.CreateRecipeStepInstrument(ctx, &mealplanninggrpc.CreateRecipeStepInstrumentRequest{
				RecipeId:     createdRecipe.ID,
				RecipeStepId: createdRecipeStepID,
				Input:        converters.ConvertRecipeStepInstrumentCreationRequestInputToGRPCRecipeStepInstrumentCreationRequestInput(exampleRecipeStepInstrumentInput),
			})
			require.NoError(t, createErr)

			createdRecipeStepInstrument := converters.ConvertGRPCRecipeStepInstrumentToRecipeStepInstrument(createdRecipeStepInstrumentRes.Created)
			checkRecipeStepInstrumentEquality(t, i, i, exampleRecipeStepInstrument, createdRecipeStepInstrument)

			fetchedRecipeStepInstrumentRes, createErr := adminClient.GetRecipeStepInstrument(ctx, &mealplanninggrpc.GetRecipeStepInstrumentRequest{
				RecipeId:               createdRecipe.ID,
				RecipeStepId:           createdRecipeStepID,
				RecipeStepInstrumentId: createdRecipeStepInstrument.ID,
			})
			require.NoError(t, createErr)

			createdRecipeStepInstrument = converters.ConvertGRPCRecipeStepInstrumentToRecipeStepInstrument(fetchedRecipeStepInstrumentRes.Result)
			require.Equal(t, createdRecipeStepID, createdRecipeStepInstrument.BelongsToRecipeStep)

			expected = append(expected, createdRecipeStepInstrument)
		}

		// assert recipe step instrument list equality
		actual, err := adminClient.GetRecipeStepInstruments(ctx, &mealplanninggrpc.GetRecipeStepInstrumentsRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
		})
		require.NoError(t, err)
		assert.True(
			t,
			len(expected) <= len(actual.Results),
			"expected %d to be <= %d",
			len(expected),
			len(actual.Results),
		)

		for _, createdRecipeStepInstrument := range expected {
			_, err = adminClient.ArchiveRecipeStepInstrument(ctx, &mealplanninggrpc.ArchiveRecipeStepInstrumentRequest{
				RecipeId:               createdRecipe.ID,
				RecipeStepId:           createdRecipeStepID,
				RecipeStepInstrumentId: createdRecipeStepInstrument.ID,
			})
			assert.NoError(t, err)
		}

		_, err = adminClient.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{RecipeId: createdRecipe.ID, RecipeStepId: createdRecipeStepID})
		assert.NoError(t, err)

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipeStepInstruments_OwnershipDenial(T *testing.T) {
	T.Parallel()

	T.Run("Create", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, otherUserClient := createUserAndClientForTest(t)

		createdRecipeStep := createdRecipe.Steps[0]

		exampleRecipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()
		exampleRecipeStepInstrument.BelongsToRecipeStep = createdRecipeStep.ID
		exampleInput := mpconverters.ConvertRecipeStepInstrumentToRecipeStepInstrumentCreationRequestInput(exampleRecipeStepInstrument)

		_, err := otherUserClient.CreateRecipeStepInstrument(ctx, &mealplanninggrpc.CreateRecipeStepInstrumentRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStep.ID,
			Input:        converters.ConvertRecipeStepInstrumentCreationRequestInputToGRPCRecipeStepInstrumentCreationRequestInput(exampleInput),
		})
		require.Error(t, err)
	})

	T.Run("Update", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, otherUserClient := createUserAndClientForTest(t)

		createdRecipeStep := createdRecipe.Steps[0]
		existingInstrument := createdRecipeStep.Instruments[0]

		updateInput := mpconverters.ConvertRecipeStepInstrumentToRecipeStepInstrumentUpdateRequestInput(existingInstrument)

		_, err := otherUserClient.UpdateRecipeStepInstrument(ctx, &mealplanninggrpc.UpdateRecipeStepInstrumentRequest{
			RecipeId:               createdRecipe.ID,
			RecipeStepId:           createdRecipeStep.ID,
			RecipeStepInstrumentId: existingInstrument.ID,
			Input:                  converters.ConvertRecipeStepInstrumentUpdateRequestInputToGRPCRecipeStepInstrumentUpdateRequestInput(updateInput),
		})
		require.Error(t, err)
	})

	T.Run("Archive", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, otherUserClient := createUserAndClientForTest(t)

		createdRecipeStep := createdRecipe.Steps[0]
		existingInstrument := createdRecipeStep.Instruments[0]

		_, err := otherUserClient.ArchiveRecipeStepInstrument(ctx, &mealplanninggrpc.ArchiveRecipeStepInstrumentRequest{
			RecipeId:               createdRecipe.ID,
			RecipeStepId:           createdRecipeStep.ID,
			RecipeStepInstrumentId: existingInstrument.ID,
		})
		require.Error(t, err)
	})
}
