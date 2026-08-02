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

func createValidPreparationVesselForTest(t *testing.T) (*types.ValidPreparation, *types.ValidVessel, *types.ValidPreparationVessel) {
	t.Helper()

	createdValidPreparation := createValidPreparationForTest(t)
	createdValidVessel := createValidVesselForTest(t)

	return createdValidPreparation, createdValidVessel, createValidPreparationVesselWithEntitiesForTest(t, createdValidPreparation, createdValidVessel)
}

// createValidPreparationVesselWithEntitiesForTest creates a ValidPreparationVessel with specific entities.
func createValidPreparationVesselWithEntitiesForTest(t *testing.T, preparation *types.ValidPreparation, vessel *types.ValidVessel) *types.ValidPreparationVessel {
	t.Helper()
	ctx := t.Context()

	exampleValidPreparationVessel := fakes.BuildFakeValidPreparationVessel()
	exampleValidPreparationVessel.Vessel = *vessel
	exampleValidPreparationVessel.Preparation = *preparation

	exampleValidPreparationVesselInput := mealplanningconverters.ConvertCreateValidPreparationVesselRequestToGRPCValidPreparationVesselCreationRequestInput(converters.ConvertValidPreparationVesselToValidPreparationVesselCreationRequestInput(exampleValidPreparationVessel))
	createdValidPreparationVessel, err := adminClient.CreateValidPreparationVessel(ctx, &mealplanningsvc.CreateValidPreparationVesselRequest{Input: exampleValidPreparationVesselInput})
	require.NoError(t, err)
	require.NotNil(t, createdValidPreparationVessel)

	validPrepVesselRes, err := adminClient.GetValidPreparationVessel(ctx, &mealplanningsvc.GetValidPreparationVesselRequest{
		ValidPreparationVesselId: createdValidPreparationVessel.Result.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, validPrepVesselRes.Result)

	return mealplanningconverters.ConvertGRPCValidPreparationVesselToValidPreparationVessel(validPrepVesselRes.Result)
}

func TestValidPreparationVessels_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createValidPreparationVesselForTest(t)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		exampleValidPreparationVessel := fakes.BuildFakeValidPreparationVessel()
		exampleValidPreparationVesselInput := mealplanningconverters.ConvertCreateValidPreparationVesselRequestToGRPCValidPreparationVesselCreationRequestInput(converters.ConvertValidPreparationVesselToValidPreparationVesselCreationRequestInput(exampleValidPreparationVessel))
		exampleValidPreparationVesselInput.ValidVesselId = ""
		exampleValidPreparationVesselInput.ValidPreparationId = ""

		createdValidPreparationVessel, err := adminClient.CreateValidPreparationVessel(ctx, &mealplanningsvc.CreateValidPreparationVesselRequest{Input: exampleValidPreparationVesselInput})
		require.Error(t, err)
		require.Nil(t, createdValidPreparationVessel)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreateValidPreparationVessel(ctx, &mealplanningsvc.CreateValidPreparationVesselRequest{})
		assert.Error(t, err)
	})
}

func TestValidPreparationVessels_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidPreparationVesselForTest(t)

		_, err := adminClient.ArchiveValidPreparationVessel(ctx, &mealplanningsvc.ArchiveValidPreparationVesselRequest{ValidPreparationVesselId: created.ID})
		assert.NoError(t, err)

		x, err := adminClient.GetValidPreparationVessel(ctx, &mealplanningsvc.GetValidPreparationVesselRequest{ValidPreparationVesselId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidPreparationVesselForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveValidPreparationVessel(ctx, &mealplanningsvc.ArchiveValidPreparationVesselRequest{ValidPreparationVesselId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveValidPreparationVessel(ctx, &mealplanningsvc.ArchiveValidPreparationVesselRequest{ValidPreparationVesselId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidPreparationVesselForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveValidPreparationVessel(ctx, &mealplanningsvc.ArchiveValidPreparationVesselRequest{ValidPreparationVesselId: created.ID})
		assert.Error(t, err)
	})
}

func TestValidPreparationVessels_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidPreparationVesselForTest(t)

		updateInput := fakes.BuildFakeValidPreparationVesselUpdateRequestInput()
		updateInput.ValidPreparationID = &created.Preparation.ID
		updateInput.ValidVesselID = &created.Vessel.ID

		response, err := adminClient.UpdateValidPreparationVessel(ctx, &mealplanningsvc.UpdateValidPreparationVesselRequest{
			ValidPreparationVesselId: created.ID,
			Input:                    mealplanningconverters.ConvertValidPreparationVesselUpdateRequestInputToGRPCValidPreparationVesselUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)
		require.NotNil(t, response)
		require.NotNil(t, response.Result)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidPreparationVesselForTest(t)

		updateInput := fakes.BuildFakeValidPreparationVesselUpdateRequestInput()
		updateInput.ValidPreparationID = &created.Preparation.ID
		updateInput.ValidVesselID = &created.Vessel.ID

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateValidPreparationVessel(ctx, &mealplanningsvc.UpdateValidPreparationVesselRequest{
			ValidPreparationVesselId: created.ID,
			Input:                    mealplanningconverters.ConvertValidPreparationVesselUpdateRequestInputToGRPCValidPreparationVesselUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from updating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		_, _, created := createValidPreparationVesselForTest(t)

		response, err := testClient.UpdateValidPreparationVessel(ctx, &mealplanningsvc.UpdateValidPreparationVesselRequest{
			ValidPreparationVesselId: created.ID,
			Input: &mealplanningsvc.ValidPreparationVesselUpdateRequestInput{
				Notes: new("doesn't matter"),
			},
		})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestValidPreparationVessels_Listing(T *testing.T) {
	T.Parallel()

	createdValidPreparationVessels := []*types.ValidPreparationVessel{}
	validPreparation, validVessel, created := createValidPreparationVesselForTest(T)
	createdValidPreparationVessels = append(createdValidPreparationVessels, created)
	// Create additional VPVs with unique (prep, vessel) pairs - use same vessel, different preparations for "by vessel" filter
	for range exampleQuantity - 1 {
		extraPreparation := createValidPreparationForTest(T)
		createdVPV := createValidPreparationVesselWithEntitiesForTest(T, extraPreparation, validVessel)
		createdValidPreparationVessels = append(createdValidPreparationVessels, createdVPV)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidPreparationVessels(ctx, &mealplanningsvc.GetValidPreparationVesselsRequest{})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.True(t, len(results.Results) >= len(createdValidPreparationVessels))
	})

	T.Run("by vessel", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidPreparationVesselsByVessel(ctx, &mealplanningsvc.GetValidPreparationVesselsByVesselRequest{ValidVesselId: validVessel.ID})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.True(t, len(results.Results) >= len(createdValidPreparationVessels))
	})

	T.Run("by preparation", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidPreparationVesselsByPreparation(ctx, &mealplanningsvc.GetValidPreparationVesselsByPreparationRequest{ValidPreparationId: validPreparation.ID})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.True(t, len(results.Results) >= 1, "at least one VPV for this preparation")
	})
}
