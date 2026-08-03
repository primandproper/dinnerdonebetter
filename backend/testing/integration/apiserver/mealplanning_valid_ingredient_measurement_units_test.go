package integration

import (
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	mealplanningconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createValidIngredientMeasurementUnitForTest(t *testing.T) (*types.ValidIngredient, *types.ValidMeasurementUnit, *types.ValidIngredientMeasurementUnit) {
	t.Helper()

	createdValidIngredient := createValidIngredientForTest(t)
	createdValidMeasurementUnit := createValidMeasurementUnitForTest(t)

	return createdValidIngredient, createdValidMeasurementUnit, createValidIngredientMeasurementUnitWithEntitiesForTest(t, createdValidIngredient, createdValidMeasurementUnit)
}

// createValidIngredientMeasurementUnitWithEntitiesForTest creates a ValidIngredientMeasurementUnit with specific entities.
func createValidIngredientMeasurementUnitWithEntitiesForTest(t *testing.T, ingredient *types.ValidIngredient, measurementUnit *types.ValidMeasurementUnit) *types.ValidIngredientMeasurementUnit {
	t.Helper()
	ctx := t.Context()

	exampleValidIngredientMeasurementUnit := fakes.BuildFakeValidIngredientMeasurementUnit()
	exampleValidIngredientMeasurementUnit.MeasurementUnit = *measurementUnit
	exampleValidIngredientMeasurementUnit.Ingredient = *ingredient

	exampleValidIngredientMeasurementUnitInput := mealplanningconverters.ConvertCreateValidIngredientMeasurementUnitRequestToGRPCValidIngredientMeasurementUnitCreationRequestInput(converters.ConvertValidIngredientMeasurementUnitToValidIngredientMeasurementUnitCreationRequestInput(exampleValidIngredientMeasurementUnit))
	createdValidIngredientMeasurementUnit, err := adminClient.CreateValidIngredientMeasurementUnit(ctx, &mealplanningsvc.CreateValidIngredientMeasurementUnitRequest{Input: exampleValidIngredientMeasurementUnitInput})
	require.NoError(t, err)
	require.NotNil(t, createdValidIngredientMeasurementUnit)

	validPrepMeasurementUnitRes, err := adminClient.GetValidIngredientMeasurementUnit(ctx, &mealplanningsvc.GetValidIngredientMeasurementUnitRequest{
		ValidIngredientMeasurementUnitId: createdValidIngredientMeasurementUnit.Result.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, validPrepMeasurementUnitRes.Result)

	return mealplanningconverters.ConvertGRPCValidIngredientMeasurementUnitToValidIngredientMeasurementUnit(validPrepMeasurementUnitRes.Result)
}

func TestValidIngredientMeasurementUnits_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createValidIngredientMeasurementUnitForTest(t)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		exampleValidIngredientMeasurementUnit := fakes.BuildFakeValidIngredientMeasurementUnit()
		exampleValidIngredientMeasurementUnitInput := mealplanningconverters.ConvertCreateValidIngredientMeasurementUnitRequestToGRPCValidIngredientMeasurementUnitCreationRequestInput(converters.ConvertValidIngredientMeasurementUnitToValidIngredientMeasurementUnitCreationRequestInput(exampleValidIngredientMeasurementUnit))
		exampleValidIngredientMeasurementUnitInput.ValidMeasurementUnitId = ""
		exampleValidIngredientMeasurementUnitInput.ValidIngredientId = ""

		createdValidIngredientMeasurementUnit, err := adminClient.CreateValidIngredientMeasurementUnit(ctx, &mealplanningsvc.CreateValidIngredientMeasurementUnitRequest{Input: exampleValidIngredientMeasurementUnitInput})
		require.Error(t, err)
		require.Nil(t, createdValidIngredientMeasurementUnit)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreateValidIngredientMeasurementUnit(ctx, &mealplanningsvc.CreateValidIngredientMeasurementUnitRequest{})
		assert.Error(t, err)
	})
}

func TestValidIngredientMeasurementUnits_Listing(T *testing.T) {
	T.Parallel()

	createdValidIngredientMeasurementUnits := []*types.ValidIngredientMeasurementUnit{}
	validIngredient, validMeasurementUnit, created := createValidIngredientMeasurementUnitForTest(T)
	createdValidIngredientMeasurementUnits = append(createdValidIngredientMeasurementUnits, created)
	// Create additional VIMUs with unique (ingredient, unit) pairs - use same unit, different ingredients for "by MeasurementUnit" filter
	for range exampleQuantity - 1 {
		extraIngredient := createValidIngredientForTest(T)
		createdVIMU := createValidIngredientMeasurementUnitWithEntitiesForTest(T, extraIngredient, validMeasurementUnit)
		createdValidIngredientMeasurementUnits = append(createdValidIngredientMeasurementUnits, createdVIMU)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidIngredientMeasurementUnits(ctx, &mealplanningsvc.GetValidIngredientMeasurementUnitsRequest{})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), len(createdValidIngredientMeasurementUnits))
	})

	T.Run("by MeasurementUnit", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidIngredientMeasurementUnitsByMeasurementUnit(ctx, &mealplanningsvc.GetValidIngredientMeasurementUnitsByMeasurementUnitRequest{ValidMeasurementUnitId: validMeasurementUnit.ID})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), len(createdValidIngredientMeasurementUnits))
	})

	T.Run("by ingredient", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidIngredientMeasurementUnitsByIngredient(ctx, &mealplanningsvc.GetValidIngredientMeasurementUnitsByIngredientRequest{ValidIngredientId: validIngredient.ID})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 1, "at least one VIMU for this ingredient")
	})
}

func TestValidIngredientMeasurementUnits_SearchByIngredient(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		createdValidIngredient, createdValidMeasurementUnit, _ := createValidIngredientMeasurementUnitForTest(t)

		results, err := adminClient.SearchValidMeasurementUnitsByIngredient(ctx, &mealplanningsvc.SearchValidMeasurementUnitsByIngredientRequest{
			ValidIngredientId: createdValidIngredient.ID,
			Query:             createdValidMeasurementUnit.Name[:2],
		})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 1, "expected at least one result when searching by ingredient")
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.SearchValidMeasurementUnitsByIngredient(ctx, &mealplanningsvc.SearchValidMeasurementUnitsByIngredientRequest{})
		assert.Error(t, err)
	})
}

func TestIntegration_UpdateValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientMeasurementUnitForTest(t)

		updateInput := fakes.BuildFakeValidIngredientMeasurementUnitUpdateRequestInput()
		updateInput.ValidIngredientID = &created.Ingredient.ID
		updateInput.ValidMeasurementUnitID = &created.MeasurementUnit.ID
		created.Update(updateInput)

		response, err := adminClient.UpdateValidIngredientMeasurementUnit(ctx, &mealplanningsvc.UpdateValidIngredientMeasurementUnitRequest{
			ValidIngredientMeasurementUnitId: created.ID,
			Input:                            mealplanningconverters.ConvertValidIngredientMeasurementUnitUpdateRequestInputToGRPCValidIngredientMeasurementUnitUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)

		updated := mealplanningconverters.ConvertGRPCValidIngredientMeasurementUnitToValidIngredientMeasurementUnit(response.Result)
		require.NotNil(t, updated.LastUpdatedAt)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientMeasurementUnitForTest(t)

		updateInput := fakes.BuildFakeValidIngredientMeasurementUnitUpdateRequestInput()
		created.Update(updateInput)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateValidIngredientMeasurementUnit(ctx, &mealplanningsvc.UpdateValidIngredientMeasurementUnitRequest{
			ValidIngredientMeasurementUnitId: created.ID,
			Input:                            mealplanningconverters.ConvertValidIngredientMeasurementUnitUpdateRequestInputToGRPCValidIngredientMeasurementUnitUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from updating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		_, _, created := createValidIngredientMeasurementUnitForTest(t)

		updateInput := fakes.BuildFakeValidIngredientMeasurementUnitUpdateRequestInput()

		response, err := testClient.UpdateValidIngredientMeasurementUnit(ctx, &mealplanningsvc.UpdateValidIngredientMeasurementUnitRequest{
			ValidIngredientMeasurementUnitId: created.ID,
			Input:                            mealplanningconverters.ConvertValidIngredientMeasurementUnitUpdateRequestInputToGRPCValidIngredientMeasurementUnitUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestIntegration_ArchiveValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientMeasurementUnitForTest(t)

		_, err := adminClient.ArchiveValidIngredientMeasurementUnit(ctx, &mealplanningsvc.ArchiveValidIngredientMeasurementUnitRequest{ValidIngredientMeasurementUnitId: created.ID})
		assert.NoError(t, err)

		x, err := adminClient.GetValidIngredientMeasurementUnit(ctx, &mealplanningsvc.GetValidIngredientMeasurementUnitRequest{ValidIngredientMeasurementUnitId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientMeasurementUnitForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveValidIngredientMeasurementUnit(ctx, &mealplanningsvc.ArchiveValidIngredientMeasurementUnitRequest{ValidIngredientMeasurementUnitId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveValidIngredientMeasurementUnit(ctx, &mealplanningsvc.ArchiveValidIngredientMeasurementUnitRequest{ValidIngredientMeasurementUnitId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientMeasurementUnitForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveValidIngredientMeasurementUnit(ctx, &mealplanningsvc.ArchiveValidIngredientMeasurementUnitRequest{ValidIngredientMeasurementUnitId: created.ID})
		assert.Error(t, err)
	})
}
