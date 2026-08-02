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

func checkRecipeStepVesselSliceEquality(t *testing.T, stepIndex int, expected, actual []*mealplanning.RecipeStepVessel) {
	t.Helper()
	require.Equal(t, len(expected), len(actual), "expected recipe step %d vessels length", stepIndex)
	for i := range expected {
		checkRecipeStepVesselEquality(t, stepIndex, i, expected[i], actual[i])
	}
}

func checkRecipeStepVesselEquality(t *testing.T, stepIndex, vesselIndex int, expected, actual *mealplanning.RecipeStepVessel) {
	t.Helper()
	assert.NotEmpty(t, actual.ID, "expected step %d vessel %d to have ID", stepIndex, vesselIndex)
	assert.False(t, actual.CreatedAt.IsZero(), "expected step %d vessel %d to have CreatedAt", stepIndex, vesselIndex)
	assert.NotEmpty(t, actual.BelongsToRecipeStep, "expected step %d vessel %d to have BelongsToRecipeStep", stepIndex, vesselIndex)
	assert.Equal(t, expected.Name, actual.Name, "expected step %d vessel %d Name", stepIndex, vesselIndex)
	assert.Equal(t, expected.Notes, actual.Notes, "expected step %d vessel %d Notes", stepIndex, vesselIndex)
	assert.Equal(t, expected.MinQuantity, actual.MinQuantity, "expected step %d vessel %d MinQuantity", stepIndex, vesselIndex)
	assert.Equal(t, expected.MaxQuantity, actual.MaxQuantity, "expected step %d vessel %d MaxQuantity", stepIndex, vesselIndex)
	assert.Equal(t, expected.Index, actual.Index, "expected step %d vessel %d Index", stepIndex, vesselIndex)
	assert.Equal(t, expected.OptionIndex, actual.OptionIndex, "expected step %d vessel %d OptionIndex", stepIndex, vesselIndex)
	assert.Equal(t, expected.VesselPreposition, actual.VesselPreposition, "expected step %d vessel %d VesselPreposition", stepIndex, vesselIndex)
	assert.Equal(t, expected.UnavailableAfterStep, actual.UnavailableAfterStep, "expected step %d vessel %d UnavailableAfterStep", stepIndex, vesselIndex)
	if expected.Vessel != nil {
		require.NotNil(t, actual.Vessel, "expected step %d vessel %d Vessel non-nil", stepIndex, vesselIndex)
		assert.NotEmpty(t, actual.Vessel.ID, "expected step %d vessel %d Vessel.ID", stepIndex, vesselIndex)
		assert.Equal(t, expected.Vessel.ID, actual.Vessel.ID, "expected step %d vessel %d Vessel.ID", stepIndex, vesselIndex)
	}
}

func TestRecipeStepVessels_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("should be able to create a recipe step vessel", func(t *testing.T) {
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

		createdValidVessel := createValidVesselForTest(t)

		// Create bridge table entry for preparation+vessel
		createdValidPreparation := &mealplanning.ValidPreparation{ID: createdRecipeStepPreparationID}
		createdValidPreparationVessel := createValidPreparationVesselWithEntitiesForTest(t, createdValidPreparation, createdValidVessel)

		// Get existing vessels to determine the next available index
		existingVesselsRes, err := adminClient.GetRecipeStepVessels(ctx, &mealplanninggrpc.GetRecipeStepVesselsRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
		})
		require.NoError(t, err)
		existingVessels := existingVesselsRes.Results
		nextIndex := uint16(len(existingVessels)) // Start from the next index after existing ones

		exampleRecipeStepVessel := fakes.BuildFakeRecipeStepVessel()
		exampleRecipeStepVessel.BelongsToRecipeStep = createdRecipeStepID
		exampleRecipeStepVessel.Vessel = &mealplanning.ValidVessel{ID: createdValidVessel.ID}
		// Set Index to match what we'll use in the creation input - use nextIndex to avoid conflicts
		exampleRecipeStepVessel.Index = nextIndex
		exampleRecipeStepVesselInput := mpconverters.ConvertRecipeStepVesselToRecipeStepVesselCreationRequestInput(exampleRecipeStepVessel)
		// Set bridge table ID (required)
		exampleRecipeStepVesselInput.ValidPreparationVesselID = &createdValidPreparationVessel.ID
		// Set Index (required for individual creation) - use nextIndex to ensure uniqueness
		exampleRecipeStepVesselInput.Index = new(nextIndex)

		createdRecipeStepVesselRes, err := adminClient.CreateRecipeStepVessel(ctx, &mealplanninggrpc.CreateRecipeStepVesselRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
			Input:        converters.ConvertRecipeStepVesselCreationRequestInputToGRPCRecipeStepVesselCreationRequestInput(exampleRecipeStepVesselInput),
		})
		require.NoError(t, err)

		createdRecipeStepVessel := converters.ConvertGRPCRecipeStepVesselToRecipeStepVessel(createdRecipeStepVesselRes.Created)
		checkRecipeStepVesselEquality(t, -1, -1, exampleRecipeStepVessel, createdRecipeStepVessel)

		retrievedRecipeStepVesselRes, err := adminClient.GetRecipeStepVessel(ctx, &mealplanninggrpc.GetRecipeStepVesselRequest{
			RecipeId:           createdRecipe.ID,
			RecipeStepId:       createdRecipeStepID,
			RecipeStepVesselId: createdRecipeStepVessel.ID,
		})
		require.NoError(t, err)

		createdRecipeStepVessel = converters.ConvertGRPCRecipeStepVesselToRecipeStepVessel(retrievedRecipeStepVesselRes.Result)
		checkRecipeStepVesselEquality(t, -1, -1, exampleRecipeStepVessel, createdRecipeStepVessel)

		newVessel := createValidVesselForTest(t)

		newRecipeStepVessel := fakes.BuildFakeRecipeStepVessel()
		newRecipeStepVessel.BelongsToRecipeStep = createdRecipeStepID
		newRecipeStepVessel.Vessel = newVessel
		// Preserve the existing Index and OptionIndex to avoid conflicts
		newRecipeStepVessel.Index = createdRecipeStepVessel.Index
		newRecipeStepVessel.OptionIndex = createdRecipeStepVessel.OptionIndex
		updateInput := mpconverters.ConvertRecipeStepVesselToRecipeStepVesselUpdateRequestInput(newRecipeStepVessel)
		createdRecipeStepVessel.Update(updateInput)
		exampleRecipeStepVessel.Update(updateInput)

		_, err = adminClient.UpdateRecipeStepVessel(ctx, &mealplanninggrpc.UpdateRecipeStepVesselRequest{
			RecipeId:           createdRecipe.ID,
			RecipeStepId:       createdRecipeStepID,
			RecipeStepVesselId: createdRecipeStepVessel.ID,
			Input:              converters.ConvertRecipeStepVesselUpdateRequestInputToGRPCRecipeStepVesselUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)

		retrievedRecipeStepVesselRes, err = adminClient.GetRecipeStepVessel(ctx, &mealplanninggrpc.GetRecipeStepVesselRequest{
			RecipeId:           createdRecipe.ID,
			RecipeStepId:       createdRecipeStepID,
			RecipeStepVesselId: createdRecipeStepVessel.ID,
		})
		require.NoError(t, err)

		createdRecipeStepVessel = converters.ConvertGRPCRecipeStepVesselToRecipeStepVessel(retrievedRecipeStepVesselRes.Result)
		checkRecipeStepVesselEquality(t, -1, -1, exampleRecipeStepVessel, createdRecipeStepVessel)

		_, err = adminClient.ArchiveRecipeStepVessel(ctx, &mealplanninggrpc.ArchiveRecipeStepVesselRequest{
			RecipeId:           createdRecipe.ID,
			RecipeStepId:       createdRecipeStepID,
			RecipeStepVesselId: createdRecipeStepVessel.ID,
		})
		assert.NoError(t, err)

		_, err = adminClient.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
		})
		assert.NoError(t, err)

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipeStepVessels_AsRecipeStepProducts(T *testing.T) {
	T.Parallel()

	T.Run("should be able to use a recipe step vessel that was the product of a prior recipe step", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		line := createValidPreparationForTest(t)
		roast := createValidPreparationForTest(t)
		bakingSheet := createValidVesselForTest(t)
		sheets := createValidMeasurementUnitForTest(t)
		head := createValidMeasurementUnitForTest(t)
		unit := createValidMeasurementUnitForTest(t)
		aluminumFoil := createValidIngredientForTest(t)
		garlic := createValidIngredientForTest(t)

		// Create bridge table entries
		// Step 0: aluminumFoil ingredient with line preparation, bakingSheet vessel with line preparation
		vipAluminumFoilLine := createValidIngredientPreparationWithEntitiesForTest(t, aluminumFoil, line)
		vimuAluminumFoilSheets := createValidIngredientMeasurementUnitWithEntitiesForTest(t, aluminumFoil, sheets)
		vpvBakingSheetLine := createValidPreparationVesselWithEntitiesForTest(t, line, bakingSheet)
		// Step 1: garlic ingredient with roast preparation (vessel is a recipe step product, no bridge ID needed)
		vipGarlicRoast := createValidIngredientPreparationWithEntitiesForTest(t, garlic, roast)
		vimuGarlicHead := createValidIngredientMeasurementUnitWithEntitiesForTest(t, garlic, head)

		linedBakingSheetName := "lined baking sheet"

		expected := &mealplanning.Recipe{
			Name:                 t.Name(),
			Slug:                 "whatever-who-cares-yadda-yadda-vessels",
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          t.Name(),
			PluralPortionName:    t.Name(),
			MinEstimatedPortions: 1,

			MaxEstimatedPortions: nil,
			Steps: []*mealplanning.RecipeStep{
				{
					Products: []*mealplanning.RecipeStepProduct{
						{
							Name:                   linedBakingSheetName,
							Type:                   mealplanning.RecipeStepProductVesselType,
							MeasurementUnit:        unit,
							QuantityNotes:          "",
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Notes:       "first step",
					Preparation: *line,
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							RecipeStepProductID: nil,
							Ingredient:          aluminumFoil,
							Name:                "aluminum foil",
							MeasurementUnit:     *sheets,
							MinQuantity:         3,

							MaxQuantity: nil,
						},
					},
					Vessels: []*mealplanning.RecipeStepVessel{
						{
							Vessel:      bakingSheet,
							Name:        "baking sheet",
							MinQuantity: 1,

							MaxQuantity: nil,
						},
					},
					Index: 0,
				},
				{
					Preparation: *roast,
					Vessels: []*mealplanning.RecipeStepVessel{
						{
							Name:   linedBakingSheetName,
							Vessel: nil,
						},
					},
					Products: []*mealplanning.RecipeStepProduct{
						{
							Name:                   "roasted garlic",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnit:        head,
							QuantityNotes:          "",
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Notes: "second step",
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Ingredient:      garlic,
							Name:            "garlic",
							MeasurementUnit: *head,
							MinQuantity:     1,

							MaxQuantity: nil,
						},
					},
					Index: 1,
				},
			},
		}

		exampleRecipeInput := mpconverters.ConvertRecipeToRecipeCreationRequestInput(expected)
		exampleRecipeInput.Steps[1].Vessels[0].ProductOfRecipeStepIndex = new(uint64(0))
		exampleRecipeInput.Steps[1].Vessels[0].ProductOfRecipeStepProductIndex = new(uint64(0))

		// Set bridge table IDs
		// Step 0: aluminumFoil ingredient and bakingSheet vessel with line preparation
		exampleRecipeInput.Steps[0].Ingredients[0].ValidIngredientPreparationID = &vipAluminumFoilLine.ID
		exampleRecipeInput.Steps[0].Ingredients[0].ValidIngredientMeasurementUnitID = &vimuAluminumFoilSheets.ID
		exampleRecipeInput.Steps[0].Vessels[0].ValidPreparationVesselID = &vpvBakingSheetLine.ID
		// Step 1: garlic ingredient with roast preparation (vessel is a recipe step product, no bridge ID needed)
		exampleRecipeInput.Steps[1].Ingredients[0].ValidIngredientPreparationID = &vipGarlicRoast.ID
		exampleRecipeInput.Steps[1].Ingredients[0].ValidIngredientMeasurementUnitID = &vimuGarlicHead.ID

		createdRes, err := adminClient.CreateRecipe(ctx, &mealplanninggrpc.CreateRecipeRequest{Input: converters.ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(exampleRecipeInput)})
		require.NoError(t, err)

		created := converters.ConvertGRPCRecipeToRecipe(createdRes.Created)
		expected.Status = created.Status
		checkRecipeEquality(t, expected, created)

		retrievedRes, err := adminClient.GetRecipe(ctx, &mealplanninggrpc.GetRecipeRequest{RecipeId: created.ID})
		require.NotNil(t, retrievedRes)
		require.NoError(t, err)
		checkRecipeEquality(t, expected, created)

		recipeStepProductIndex := -1
		for i, vessel := range created.Steps[1].Vessels {
			if vessel.RecipeStepProductID != nil {
				recipeStepProductIndex = i
			}
		}

		require.NotEqual(t, -1, recipeStepProductIndex)
		require.NotNil(t, created.Steps[1].Vessels[recipeStepProductIndex].RecipeStepProductID)
		assert.Equal(t, created.Steps[0].Products[0].ID, *created.Steps[1].Vessels[recipeStepProductIndex].RecipeStepProductID)
	})
}

func TestRecipeStepVessels_Listing(T *testing.T) {
	T.Parallel()

	T.Run("should be able to list recipe step vessels", func(t *testing.T) {
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

		createdValidVessel := createValidVesselForTest(t)

		// Create bridge table entry for preparation+vessel
		createdValidPreparation := &mealplanning.ValidPreparation{ID: createdRecipeStepPreparationID}
		createdValidPreparationVessel := createValidPreparationVesselWithEntitiesForTest(t, createdValidPreparation, createdValidVessel)

		// Get existing vessels to determine the next available index
		existingVesselsRes, err := adminClient.GetRecipeStepVessels(ctx, &mealplanninggrpc.GetRecipeStepVesselsRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStepID,
		})
		require.NoError(t, err)
		existingVessels := existingVesselsRes.Results
		nextIndex := uint16(len(existingVessels)) // Start from the next index after existing ones

		var expected []*mealplanning.RecipeStepVessel
		for i := range 5 {
			exampleRecipeStepVessel := fakes.BuildFakeRecipeStepVessel()
			exampleRecipeStepVessel.BelongsToRecipeStep = createdRecipeStepID
			exampleRecipeStepVessel.Vessel = &mealplanning.ValidVessel{ID: createdValidVessel.ID}
			// Set Index to match what we'll use in the creation input - start from nextIndex to avoid conflicts
			vesselIndex := nextIndex + uint16(i)
			exampleRecipeStepVessel.Index = vesselIndex
			exampleRecipeStepVesselInput := mpconverters.ConvertRecipeStepVesselToRecipeStepVesselCreationRequestInput(exampleRecipeStepVessel)
			// Set bridge table ID (required)
			exampleRecipeStepVesselInput.ValidPreparationVesselID = &createdValidPreparationVessel.ID
			// Set Index (required for individual creation) - use nextIndex + loop index to ensure uniqueness
			exampleRecipeStepVesselInput.Index = new(vesselIndex)
			createdRecipeStepVesselRes, createErr := adminClient.CreateRecipeStepVessel(ctx, &mealplanninggrpc.CreateRecipeStepVesselRequest{
				RecipeId:     createdRecipe.ID,
				RecipeStepId: createdRecipeStepID,
				Input:        converters.ConvertRecipeStepVesselCreationRequestInputToGRPCRecipeStepVesselCreationRequestInput(exampleRecipeStepVesselInput),
			})
			require.NoError(t, createErr)

			createdRecipeStepVessel := converters.ConvertGRPCRecipeStepVesselToRecipeStepVessel(createdRecipeStepVesselRes.Created)
			checkRecipeStepVesselEquality(t, i, i, exampleRecipeStepVessel, createdRecipeStepVessel)

			retrievedRecipeStepVesselRes, getErr := adminClient.GetRecipeStepVessel(ctx, &mealplanninggrpc.GetRecipeStepVesselRequest{
				RecipeId:           createdRecipe.ID,
				RecipeStepId:       createdRecipeStepID,
				RecipeStepVesselId: createdRecipeStepVessel.ID,
			})
			require.NotNil(t, retrievedRecipeStepVesselRes)
			require.NoError(t, getErr)

			createdRecipeStepVessel = converters.ConvertGRPCRecipeStepVesselToRecipeStepVessel(retrievedRecipeStepVesselRes.Result)
			require.Equal(t, createdRecipeStepID, createdRecipeStepVessel.BelongsToRecipeStep)

			expected = append(expected, createdRecipeStepVessel)
		}

		// assert recipe step vessel list equality
		actual, err := adminClient.GetRecipeStepVessels(ctx, &mealplanninggrpc.GetRecipeStepVesselsRequest{
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

		for _, createdRecipeStepVessel := range expected {
			_, err = adminClient.ArchiveRecipeStepVessel(ctx, &mealplanninggrpc.ArchiveRecipeStepVesselRequest{
				RecipeId:           createdRecipe.ID,
				RecipeStepId:       createdRecipeStepID,
				RecipeStepVesselId: createdRecipeStepVessel.ID,
			})
			assert.NoError(t, err)
		}

		_, err = adminClient.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{RecipeId: createdRecipe.ID, RecipeStepId: createdRecipeStepID})
		assert.NoError(t, err)

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipeStepVessels_OwnershipDenial(T *testing.T) {
	T.Parallel()

	T.Run("non-owner cannot create recipe step vessel", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, otherUserClient := createUserAndClientForTest(t)
		_, _, createdRecipe := createRecipeForTest(t, nil)

		createdRecipeStep := createdRecipe.Steps[0]

		exampleRecipeStepVessel := fakes.BuildFakeRecipeStepVessel()
		exampleRecipeStepVesselInput := mpconverters.ConvertRecipeStepVesselToRecipeStepVesselCreationRequestInput(exampleRecipeStepVessel)
		_, err := otherUserClient.CreateRecipeStepVessel(ctx, &mealplanninggrpc.CreateRecipeStepVesselRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStep.ID,
			Input:        converters.ConvertRecipeStepVesselCreationRequestInputToGRPCRecipeStepVesselCreationRequestInput(exampleRecipeStepVesselInput),
		})
		require.Error(t, err)
	})

	T.Run("non-owner cannot update recipe step vessel", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, otherUserClient := createUserAndClientForTest(t)
		_, _, createdRecipe := createRecipeForTest(t, nil)

		createdRecipeStep := createdRecipe.Steps[0]
		createdRecipeStepVessel := createdRecipeStep.Vessels[0]

		newRecipeStepVessel := fakes.BuildFakeRecipeStepVessel()
		updateInput := mpconverters.ConvertRecipeStepVesselToRecipeStepVesselUpdateRequestInput(newRecipeStepVessel)
		_, err := otherUserClient.UpdateRecipeStepVessel(ctx, &mealplanninggrpc.UpdateRecipeStepVesselRequest{
			RecipeId:           createdRecipe.ID,
			RecipeStepId:       createdRecipeStep.ID,
			RecipeStepVesselId: createdRecipeStepVessel.ID,
			Input:              converters.ConvertRecipeStepVesselUpdateRequestInputToGRPCRecipeStepVesselUpdateRequestInput(updateInput),
		})
		require.Error(t, err)
	})

	T.Run("non-owner cannot archive recipe step vessel", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, otherUserClient := createUserAndClientForTest(t)
		_, _, createdRecipe := createRecipeForTest(t, nil)

		createdRecipeStep := createdRecipe.Steps[0]
		createdRecipeStepVessel := createdRecipeStep.Vessels[0]

		_, err := otherUserClient.ArchiveRecipeStepVessel(ctx, &mealplanninggrpc.ArchiveRecipeStepVesselRequest{
			RecipeId:           createdRecipe.ID,
			RecipeStepId:       createdRecipeStep.ID,
			RecipeStepVesselId: createdRecipeStepVessel.ID,
		})
		require.Error(t, err)
	})
}
