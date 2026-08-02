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

func checkValidVesselEquality(t *testing.T, expected, actual *mealplanning.ValidVessel) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected ValidVessel to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected ValidVessel to have CreatedAt")

	assert.Equal(t, expected.Name, actual.Name, "expected ValidVessel Name")
	assert.Equal(t, expected.Description, actual.Description, "expected ValidVessel Description")
	assert.Equal(t, expected.Slug, actual.Slug, "expected ValidVessel Slug")
	assert.Equal(t, expected.PluralName, actual.PluralName, "expected ValidVessel PluralName")
	assert.Equal(t, expected.IconPath, actual.IconPath, "expected ValidVessel IconPath")
	assert.Equal(t, expected.Shape, actual.Shape, "expected ValidVessel Shape")
	assert.Equal(t, expected.WidthInMillimeters, actual.WidthInMillimeters, "expected ValidVessel WidthInMillimeters")
	assert.Equal(t, expected.LengthInMillimeters, actual.LengthInMillimeters, "expected ValidVessel LengthInMillimeters")
	assert.Equal(t, expected.HeightInMillimeters, actual.HeightInMillimeters, "expected ValidVessel HeightInMillimeters")
	assert.Equal(t, expected.Capacity, actual.Capacity, "expected ValidVessel Capacity")
	assert.Equal(t, expected.DisplayInSummaryLists, actual.DisplayInSummaryLists, "expected ValidVessel DisplayInSummaryLists")
	assert.Equal(t, expected.IncludeInGeneratedInstructions, actual.IncludeInGeneratedInstructions, "expected ValidVessel IncludeInGeneratedInstructions")
	assert.Equal(t, expected.UsableForStorage, actual.UsableForStorage, "expected ValidVessel UsableForStorage")
}

func createValidVesselForTest(t *testing.T) *mealplanning.ValidVessel {
	t.Helper()

	ctx := t.Context()

	exampleValidVessel := fakes.BuildFakeValidVessel()
	exampleValidVesselInput := converters.ConvertValidVesselToValidVesselCreationRequestInput(exampleValidVessel)

	measurementUnit := createValidMeasurementUnitForTest(t)
	exampleValidVesselInput.CapacityUnitID = &measurementUnit.ID

	created, err := adminClient.CreateValidVessel(ctx, &mealplanningsvc.CreateValidVesselRequest{
		Input: grpcconverters.ConvertValidVesselCreationRequestInputToGRPCValidVesselCreationRequestInput(exampleValidVesselInput),
	})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCValidVesselToValidVessel(created.Result)
	checkValidVesselEquality(t, exampleValidVessel, converted)

	retrieved, err := adminClient.GetValidVessel(ctx, &mealplanningsvc.GetValidVesselRequest{
		ValidVesselId: converted.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	validVessel := grpcconverters.ConvertGRPCValidVesselToValidVessel(retrieved.Result)
	checkValidVesselEquality(t, converted, validVessel)

	return validVessel
}

func TestValidVessels_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createValidVesselForTest(t)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeValidVesselCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidVesselCreationRequestInputToGRPCValidVesselCreationRequestInput(creationRequestInput)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateValidVessel(ctx, &mealplanningsvc.CreateValidVesselRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeValidVesselCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidVesselCreationRequestInputToGRPCValidVesselCreationRequestInput(creationRequestInput)
		// this is not allowed
		convertedInput.Name = ""

		created, err := adminClient.CreateValidVessel(ctx, &mealplanningsvc.CreateValidVesselRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		creationRequestInput := fakes.BuildFakeValidVesselCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidVesselCreationRequestInputToGRPCValidVesselCreationRequestInput(creationRequestInput)

		created, err := testClient.CreateValidVessel(ctx, &mealplanningsvc.CreateValidVesselRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestValidVessels_Reading(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidVesselForTest(t)

		retrieved, err := testClient.GetValidVessel(ctx, &mealplanningsvc.GetValidVesselRequest{ValidVesselId: created.ID})
		assert.NoError(t, err)

		converted := grpcconverters.ConvertGRPCValidVesselToValidVessel(retrieved.Result)

		assertRoughEquality(t, created, converted, append(defaultIgnoredFields(), "CapacityUnit")...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidVesselForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetValidVessel(ctx, &mealplanningsvc.GetValidVesselRequest{ValidVesselId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetValidVessel(ctx, &mealplanningsvc.GetValidVesselRequest{ValidVesselId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestValidVessels_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidVesselForTest(t)

		updateInput := fakes.BuildFakeValidVesselUpdateRequestInput()
		updateInput.CapacityUnitID = nil
		created.Update(updateInput)

		response, err := adminClient.UpdateValidVessel(ctx, &mealplanningsvc.UpdateValidVesselRequest{
			ValidVesselId: created.ID,
			Input:         grpcconverters.ConvertValidVesselUpdateRequestInputToGRPCValidVesselUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)

		updated := grpcconverters.ConvertGRPCValidVesselToValidVessel(response.Result)
		// Ensure UpdatedAt was set
		require.NotNil(t, updated.LastUpdatedAt)

		assertRoughEquality(t, created, updated, append(defaultIgnoredFields(), "CapacityUnit")...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidVesselForTest(t)

		updateInput := fakes.BuildFakeValidVesselUpdateRequestInput()
		created.Update(updateInput)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateValidVessel(ctx, &mealplanningsvc.UpdateValidVesselRequest{
			ValidVesselId: created.ID,
			Input:         grpcconverters.ConvertValidVesselUpdateRequestInputToGRPCValidVesselUpdateRequestInput(updateInput),
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

		created := createValidVesselForTest(t)

		response, err := testClient.UpdateValidVessel(ctx, &mealplanningsvc.UpdateValidVesselRequest{
			ValidVesselId: created.ID,
			Input: &mealplanningsvc.ValidVesselUpdateRequestInput{
				Name: new("doesn't matter"),
			},
		})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestValidVessels_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidVesselForTest(t)

		_, err := adminClient.ArchiveValidVessel(ctx, &mealplanningsvc.ArchiveValidVesselRequest{ValidVesselId: created.ID})
		assert.NoError(t, err)

		x, err := adminClient.GetValidVessel(ctx, &mealplanningsvc.GetValidVesselRequest{ValidVesselId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidVesselForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveValidVessel(ctx, &mealplanningsvc.ArchiveValidVesselRequest{ValidVesselId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveValidVessel(ctx, &mealplanningsvc.ArchiveValidVesselRequest{ValidVesselId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidVesselForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveValidVessel(ctx, &mealplanningsvc.ArchiveValidVesselRequest{ValidVesselId: created.ID})
		assert.Error(t, err)
	})
}

func TestValidVessels_GetRandom(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// in case we haven't already
		createValidVesselForTest(t)

		response, err := testClient.GetRandomValidVessel(ctx, &mealplanningsvc.GetRandomValidVesselRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, response)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		response, err := c.GetRandomValidVessel(ctx, &mealplanningsvc.GetRandomValidVesselRequest{})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestValidVessels_Listing(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)
	createdValidVessels := []*mealplanning.ValidVessel{}
	for range exampleQuantity {
		created := createValidVesselForTest(T)
		createdValidVessels = append(createdValidVessels, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.GetValidVessels(ctx, &mealplanningsvc.GetValidVesselsRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.True(t, len(retrieved.Results) >= len(createdValidVessels))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetValidVessels(ctx, &mealplanningsvc.GetValidVesselsRequest{})
		assert.Error(t, err)
	})
}

func TestValidVessels_Searching(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)
		created := createValidVesselForTest(t)

		retrieved, err := testClient.SearchForValidVessels(ctx, &mealplanningsvc.SearchForValidVesselsRequest{
			Query: created.Name[:2],
		})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.SearchForValidVessels(ctx, &mealplanningsvc.SearchForValidVesselsRequest{})
		assert.Error(t, err)
	})
}
