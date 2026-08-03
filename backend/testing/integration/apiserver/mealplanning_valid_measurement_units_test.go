package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkValidMeasurementUnitEquality(t *testing.T, expected, actual *mealplanning.ValidMeasurementUnit) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected ValidMeasurementUnit to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected ValidMeasurementUnit to have CreatedAt")

	assert.Equal(t, expected.Name, actual.Name, "expected ValidMeasurementUnit Name")
	assert.Equal(t, expected.Description, actual.Description, "expected ValidMeasurementUnit Description")
	assert.Equal(t, expected.Slug, actual.Slug, "expected ValidMeasurementUnit Slug")
	assert.Equal(t, expected.PluralName, actual.PluralName, "expected ValidMeasurementUnit PluralName")
	assert.Equal(t, expected.IconPath, actual.IconPath, "expected ValidMeasurementUnit IconPath")
	assert.Equal(t, expected.Volumetric, actual.Volumetric, "expected ValidMeasurementUnit Volumetric")
	assert.Equal(t, expected.Universal, actual.Universal, "expected ValidMeasurementUnit Universal")
	assert.Equal(t, expected.Metric, actual.Metric, "expected ValidMeasurementUnit Metric")
	assert.Equal(t, expected.Imperial, actual.Imperial, "expected ValidMeasurementUnit Imperial")
}

func createValidMeasurementUnitForTest(t *testing.T) *mealplanning.ValidMeasurementUnit {
	t.Helper()

	ctx := t.Context()

	exampleValidMeasurementUnit := fakes.BuildFakeValidMeasurementUnit()
	exampleValidMeasurementUnitInput := converters.ConvertValidMeasurementUnitToValidMeasurementUnitCreationRequestInput(exampleValidMeasurementUnit)
	created, err := adminClient.CreateValidMeasurementUnit(ctx, &mealplanningsvc.CreateValidMeasurementUnitRequest{
		Input: grpcconverters.ConvertValidMeasurementUnitCreationRequestInputToGRPCValidMeasurementUnitCreationRequestInput(exampleValidMeasurementUnitInput),
	})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCValidMeasurementUnitToValidMeasurementUnit(created.Result)
	checkValidMeasurementUnitEquality(t, exampleValidMeasurementUnit, converted)

	retrieved, err := adminClient.GetValidMeasurementUnit(ctx, &mealplanningsvc.GetValidMeasurementUnitRequest{
		ValidMeasurementUnitId: converted.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	validMeasurementUnit := grpcconverters.ConvertGRPCValidMeasurementUnitToValidMeasurementUnit(retrieved.Result)
	checkValidMeasurementUnitEquality(t, converted, validMeasurementUnit)

	return validMeasurementUnit
}

func TestValidMeasurementUnits_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createValidMeasurementUnitForTest(t)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeValidMeasurementUnitCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidMeasurementUnitCreationRequestInputToGRPCValidMeasurementUnitCreationRequestInput(creationRequestInput)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateValidMeasurementUnit(ctx, &mealplanningsvc.CreateValidMeasurementUnitRequest{
			Input: convertedInput,
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeValidMeasurementUnitCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidMeasurementUnitCreationRequestInputToGRPCValidMeasurementUnitCreationRequestInput(creationRequestInput)
		// this is not allowed
		convertedInput.Name = ""

		created, err := adminClient.CreateValidMeasurementUnit(ctx, &mealplanningsvc.CreateValidMeasurementUnitRequest{
			Input: convertedInput,
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		creationRequestInput := fakes.BuildFakeValidMeasurementUnitCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidMeasurementUnitCreationRequestInputToGRPCValidMeasurementUnitCreationRequestInput(creationRequestInput)

		created, err := testClient.CreateValidMeasurementUnit(ctx, &mealplanningsvc.CreateValidMeasurementUnitRequest{
			Input: convertedInput,
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestValidMeasurementUnits_Reading(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidMeasurementUnitForTest(t)

		retrieved, err := testClient.GetValidMeasurementUnit(ctx, &mealplanningsvc.GetValidMeasurementUnitRequest{ValidMeasurementUnitId: created.ID})
		require.NoError(t, err)

		converted := grpcconverters.ConvertGRPCValidMeasurementUnitToValidMeasurementUnit(retrieved.Result)

		assertRoughEquality(t, created, converted, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidMeasurementUnitForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetValidMeasurementUnit(ctx, &mealplanningsvc.GetValidMeasurementUnitRequest{ValidMeasurementUnitId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetValidMeasurementUnit(ctx, &mealplanningsvc.GetValidMeasurementUnitRequest{ValidMeasurementUnitId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestValidMeasurementUnits_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidMeasurementUnitForTest(t)

		updateInput := fakes.BuildFakeValidMeasurementUnitUpdateRequestInput()
		created.Update(updateInput)

		response, err := adminClient.UpdateValidMeasurementUnit(ctx, &mealplanningsvc.UpdateValidMeasurementUnitRequest{
			ValidMeasurementUnitId: created.ID,
			Input:                  grpcconverters.ConvertValidMeasurementUnitUpdateRequestInputToGRPCValidMeasurementUnitUpdateRequestInput(updateInput),
		})
		require.NoError(t, err)

		updated := grpcconverters.ConvertGRPCValidMeasurementUnitToValidMeasurementUnit(response.Result)
		// Ensure UpdatedAt was set
		require.NotNil(t, updated.LastUpdatedAt)

		assertRoughEquality(t, created, updated, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidMeasurementUnitForTest(t)

		updateInput := fakes.BuildFakeValidMeasurementUnitUpdateRequestInput()
		created.Update(updateInput)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateValidMeasurementUnit(ctx, &mealplanningsvc.UpdateValidMeasurementUnitRequest{
			ValidMeasurementUnitId: created.ID,
			Input:                  grpcconverters.ConvertValidMeasurementUnitUpdateRequestInputToGRPCValidMeasurementUnitUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()

		/*
			there's no way to provide invalid input to this method, but
			I want to make it explicit that tests should be written the moment that changes
		*/
	})

	T.Run("non-admin users are forbidden from updating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		created := createValidMeasurementUnitForTest(t)

		response, err := testClient.UpdateValidMeasurementUnit(ctx, &mealplanningsvc.UpdateValidMeasurementUnitRequest{
			ValidMeasurementUnitId: created.ID,
			Input: &mealplanningsvc.ValidMeasurementUnitUpdateRequestInput{
				Name: new("doesn't matter"),
			},
		})
		require.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestValidMeasurementUnits_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidMeasurementUnitForTest(t)

		_, err := adminClient.ArchiveValidMeasurementUnit(ctx, &mealplanningsvc.ArchiveValidMeasurementUnitRequest{ValidMeasurementUnitId: created.ID})
		require.NoError(t, err)

		x, err := adminClient.GetValidMeasurementUnit(ctx, &mealplanningsvc.GetValidMeasurementUnitRequest{ValidMeasurementUnitId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidMeasurementUnitForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveValidMeasurementUnit(ctx, &mealplanningsvc.ArchiveValidMeasurementUnitRequest{ValidMeasurementUnitId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveValidMeasurementUnit(ctx, &mealplanningsvc.ArchiveValidMeasurementUnitRequest{ValidMeasurementUnitId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidMeasurementUnitForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveValidMeasurementUnit(ctx, &mealplanningsvc.ArchiveValidMeasurementUnitRequest{ValidMeasurementUnitId: created.ID})
		assert.Error(t, err)
	})
}

/* TODO: I have this functionality, don't I?
func TestValidMeasurementUnits_GetRandom(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// in case we haven't already
		createValidMeasurementUnitForTest(t)

		response, err := testClient.GetRandomValidMeasurementUnit(ctx, &mealplanningsvc.GetRandomValidMeasurementUnitRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, response)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		response, err := c.GetRandomValidMeasurementUnit(ctx, &mealplanningsvc.GetRandomValidMeasurementUnitRequest{})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}
*/

func TestValidMeasurementUnits_Listing(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)
	createdValidMeasurementUnits := []*mealplanning.ValidMeasurementUnit{}
	for range exampleQuantity {
		created := createValidMeasurementUnitForTest(T)
		createdValidMeasurementUnits = append(createdValidMeasurementUnits, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.GetValidMeasurementUnits(ctx, &mealplanningsvc.GetValidMeasurementUnitsRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.GreaterOrEqual(t, len(retrieved.Results), len(createdValidMeasurementUnits))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetValidMeasurementUnits(ctx, &mealplanningsvc.GetValidMeasurementUnitsRequest{})
		assert.Error(t, err)
	})
}

func TestValidMeasurementUnits_Searching(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)
		created := createValidMeasurementUnitForTest(t)

		retrieved, err := testClient.SearchForValidMeasurementUnits(ctx, &mealplanningsvc.SearchForValidMeasurementUnitsRequest{
			Query: created.Name[:2],
		})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.SearchForValidMeasurementUnits(ctx, &mealplanningsvc.SearchForValidMeasurementUnitsRequest{})
		assert.Error(t, err)
	})
}
