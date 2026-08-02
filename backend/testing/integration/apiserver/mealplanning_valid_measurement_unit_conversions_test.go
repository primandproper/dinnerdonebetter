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

func createValidMeasurementUnitConversionForTest(t *testing.T) (unit1, unit2 *types.ValidMeasurementUnit, conversion *types.ValidMeasurementUnitConversion) {
	t.Helper()
	ctx := t.Context()

	unit1 = createValidMeasurementUnitForTest(t)
	unit2 = createValidMeasurementUnitForTest(t)

	exampleValidMeasurementUnitConversion := fakes.BuildFakeValidMeasurementUnitConversion()
	exampleValidMeasurementUnitConversion.To = *unit1
	exampleValidMeasurementUnitConversion.From = *unit2

	exampleValidMeasurementUnitConversionInput := mealplanningconverters.ConvertCreateValidMeasurementUnitConversionRequestToGRPCValidMeasurementUnitConversionCreationRequestInput(converters.ConvertValidMeasurementUnitConversionToValidMeasurementUnitConversionCreationRequestInput(exampleValidMeasurementUnitConversion))
	createdValidMeasurementUnitConversion, err := adminClient.CreateValidMeasurementUnitConversion(ctx, &mealplanningsvc.CreateValidMeasurementUnitConversionRequest{Input: exampleValidMeasurementUnitConversionInput})
	require.NoError(t, err)
	require.NotNil(t, createdValidMeasurementUnitConversion)

	validPrepPreparationRes, err := adminClient.GetValidMeasurementUnitConversion(ctx, &mealplanningsvc.GetValidMeasurementUnitConversionRequest{
		ValidMeasurementUnitConversionId: createdValidMeasurementUnitConversion.Result.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, validPrepPreparationRes.Result)

	return unit1, unit2, mealplanningconverters.ConvertGRPCValidMeasurementUnitConversionToValidMeasurementUnitConversion(validPrepPreparationRes.Result)
}

func TestValidMeasurementUnitConversions_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createValidMeasurementUnitConversionForTest(t)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		exampleValidMeasurementUnitConversion := fakes.BuildFakeValidMeasurementUnitConversion()
		exampleValidMeasurementUnitConversionInput := mealplanningconverters.ConvertCreateValidMeasurementUnitConversionRequestToGRPCValidMeasurementUnitConversionCreationRequestInput(converters.ConvertValidMeasurementUnitConversionToValidMeasurementUnitConversionCreationRequestInput(exampleValidMeasurementUnitConversion))
		exampleValidMeasurementUnitConversionInput.To = ""
		exampleValidMeasurementUnitConversionInput.From = ""

		createdValidMeasurementUnitConversion, err := adminClient.CreateValidMeasurementUnitConversion(ctx, &mealplanningsvc.CreateValidMeasurementUnitConversionRequest{Input: exampleValidMeasurementUnitConversionInput})
		require.Error(t, err)
		require.Nil(t, createdValidMeasurementUnitConversion)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreateValidMeasurementUnitConversion(ctx, &mealplanningsvc.CreateValidMeasurementUnitConversionRequest{})
		assert.Error(t, err)
	})
}

func TestValidMeasurementUnitConversions_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidMeasurementUnitConversionForTest(t)

		_, err := adminClient.ArchiveValidMeasurementUnitConversion(ctx, &mealplanningsvc.ArchiveValidMeasurementUnitConversionRequest{ValidMeasurementUnitConversionId: created.ID})
		assert.NoError(t, err)

		x, err := adminClient.GetValidMeasurementUnitConversion(ctx, &mealplanningsvc.GetValidMeasurementUnitConversionRequest{ValidMeasurementUnitConversionId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidMeasurementUnitConversionForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveValidMeasurementUnitConversion(ctx, &mealplanningsvc.ArchiveValidMeasurementUnitConversionRequest{ValidMeasurementUnitConversionId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveValidMeasurementUnitConversion(ctx, &mealplanningsvc.ArchiveValidMeasurementUnitConversionRequest{ValidMeasurementUnitConversionId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidMeasurementUnitConversionForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveValidMeasurementUnitConversion(ctx, &mealplanningsvc.ArchiveValidMeasurementUnitConversionRequest{ValidMeasurementUnitConversionId: created.ID})
		assert.Error(t, err)
	})
}

func TestValidMeasurementUnitConversions_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidMeasurementUnitConversionForTest(t)

		updateInput := fakes.BuildFakeValidMeasurementUnitConversionUpdateRequestInput()
		updateInput.From = &created.From.ID
		updateInput.To = &created.To.ID
		updateInput.OnlyForIngredient = nil

		response, err := adminClient.UpdateValidMeasurementUnitConversion(ctx, &mealplanningsvc.UpdateValidMeasurementUnitConversionRequest{
			ValidMeasurementUnitConversionId: created.ID,
			Input:                            mealplanningconverters.ConvertValidMeasurementUnitConversionUpdateRequestInputToGRPCValidMeasurementUnitConversionUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)
		require.NotNil(t, response)
		require.NotNil(t, response.Result)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidMeasurementUnitConversionForTest(t)

		updateInput := fakes.BuildFakeValidMeasurementUnitConversionUpdateRequestInput()
		updateInput.From = &created.From.ID
		updateInput.To = &created.To.ID
		updateInput.OnlyForIngredient = nil

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateValidMeasurementUnitConversion(ctx, &mealplanningsvc.UpdateValidMeasurementUnitConversionRequest{
			ValidMeasurementUnitConversionId: created.ID,
			Input:                            mealplanningconverters.ConvertValidMeasurementUnitConversionUpdateRequestInputToGRPCValidMeasurementUnitConversionUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from updating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		_, _, created := createValidMeasurementUnitConversionForTest(t)

		response, err := testClient.UpdateValidMeasurementUnitConversion(ctx, &mealplanningsvc.UpdateValidMeasurementUnitConversionRequest{
			ValidMeasurementUnitConversionId: created.ID,
			Input: &mealplanningsvc.ValidMeasurementUnitConversionUpdateRequestInput{
				Notes: new("doesn't matter"),
			},
		})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestValidMeasurementUnitConversions_GetMismatches(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetMeasurementUnitConversionMismatches(ctx, &mealplanningsvc.GetMeasurementUnitConversionMismatchesRequest{})
		require.NoError(t, err)
		require.NotNil(t, results)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetMeasurementUnitConversionMismatches(ctx, &mealplanningsvc.GetMeasurementUnitConversionMismatchesRequest{})
		assert.Error(t, err)
	})

	T.Run("non-admin users can also access", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		response, err := testClient.GetMeasurementUnitConversionMismatches(ctx, &mealplanningsvc.GetMeasurementUnitConversionMismatchesRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, response)
	})
}

func TestValidMeasurementUnitConversions_Listing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		toUnit, fromUnit, created := createValidMeasurementUnitConversionForTest(T)
		createdValidMeasurementUnitConversions := []*types.ValidMeasurementUnitConversion{created}

		results, err := adminClient.GetValidMeasurementUnitConversionsForUnit(ctx, &mealplanningsvc.GetValidMeasurementUnitConversionsForUnitRequest{
			ValidMeasurementUnitId: toUnit.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Equal(t, len(results.Results), len(createdValidMeasurementUnitConversions))
		assert.Equal(t, results.Results[0].Id, createdValidMeasurementUnitConversions[0].ID)

		results, err = adminClient.GetValidMeasurementUnitConversionsForUnit(ctx, &mealplanningsvc.GetValidMeasurementUnitConversionsForUnitRequest{
			ValidMeasurementUnitId: fromUnit.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.Equal(t, len(results.Results), len(createdValidMeasurementUnitConversions))
		assert.Equal(t, results.Results[0].Id, createdValidMeasurementUnitConversions[0].ID)
	})
}

func TestValidMeasurementUnitConversions_ListingForIngredients(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		createdIngredient := createValidIngredientForTest(t)

		toUnit, _, _ := createValidMeasurementUnitConversionForTest(t)

		createValidIngredientMeasurementUnitWithEntitiesForTest(t, createdIngredient, toUnit)

		results, err := adminClient.GetValidMeasurementUnitConversionsForIngredients(ctx, &mealplanningsvc.GetValidMeasurementUnitConversionsForIngredientsRequest{
			ValidIngredientIds: []string{createdIngredient.ID},
		})
		require.NoError(t, err)
		require.NotNil(t, results)
		require.NotEmpty(t, results.Results)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetValidMeasurementUnitConversionsForIngredients(ctx, &mealplanningsvc.GetValidMeasurementUnitConversionsForIngredientsRequest{
			ValidIngredientIds: []string{"fake"},
		})
		assert.Error(t, err)
	})
}
