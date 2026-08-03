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

func createValidPreparationInstrumentForTest(t *testing.T) (*types.ValidPreparation, *types.ValidInstrument, *types.ValidPreparationInstrument) {
	t.Helper()

	createdValidPreparation := createValidPreparationForTest(t)
	createdValidInstrument := createValidInstrumentForTest(t)

	return createdValidPreparation, createdValidInstrument, createValidPreparationInstrumentWithEntitiesForTest(t, createdValidPreparation, createdValidInstrument)
}

// createValidPreparationInstrumentWithEntitiesForTest creates a ValidPreparationInstrument with specific entities.
func createValidPreparationInstrumentWithEntitiesForTest(t *testing.T, preparation *types.ValidPreparation, instrument *types.ValidInstrument) *types.ValidPreparationInstrument {
	t.Helper()
	ctx := t.Context()

	exampleValidPreparationInstrument := fakes.BuildFakeValidPreparationInstrument()
	exampleValidPreparationInstrument.Instrument = *instrument
	exampleValidPreparationInstrument.Preparation = *preparation

	exampleValidPreparationInstrumentInput := mealplanningconverters.ConvertCreateValidPreparationInstrumentRequestToGRPCValidPreparationInstrumentCreationRequestInput(converters.ConvertValidPreparationInstrumentToValidPreparationInstrumentCreationRequestInput(exampleValidPreparationInstrument))
	createdValidPreparationInstrument, err := adminClient.CreateValidPreparationInstrument(ctx, &mealplanningsvc.CreateValidPreparationInstrumentRequest{Input: exampleValidPreparationInstrumentInput})
	require.NoError(t, err)
	require.NotNil(t, createdValidPreparationInstrument)

	validPrepInstrumentRes, err := adminClient.GetValidPreparationInstrument(ctx, &mealplanningsvc.GetValidPreparationInstrumentRequest{
		ValidPreparationInstrumentId: createdValidPreparationInstrument.Result.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, validPrepInstrumentRes.Result)

	return mealplanningconverters.ConvertGRPCValidPreparationInstrumentToValidPreparationInstrument(validPrepInstrumentRes.Result)
}

func TestValidPreparationInstruments_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createValidPreparationInstrumentForTest(t)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		exampleValidPreparationInstrument := fakes.BuildFakeValidPreparationInstrument()
		exampleValidPreparationInstrumentInput := mealplanningconverters.ConvertCreateValidPreparationInstrumentRequestToGRPCValidPreparationInstrumentCreationRequestInput(converters.ConvertValidPreparationInstrumentToValidPreparationInstrumentCreationRequestInput(exampleValidPreparationInstrument))
		exampleValidPreparationInstrumentInput.ValidInstrumentId = ""
		exampleValidPreparationInstrumentInput.ValidPreparationId = ""

		createdValidPreparationInstrument, err := adminClient.CreateValidPreparationInstrument(ctx, &mealplanningsvc.CreateValidPreparationInstrumentRequest{Input: exampleValidPreparationInstrumentInput})
		require.Error(t, err)
		require.Nil(t, createdValidPreparationInstrument)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreateValidPreparationInstrument(ctx, &mealplanningsvc.CreateValidPreparationInstrumentRequest{})
		assert.Error(t, err)
	})
}

func TestValidPreparationInstruments_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidPreparationInstrumentForTest(t)

		_, err := adminClient.ArchiveValidPreparationInstrument(ctx, &mealplanningsvc.ArchiveValidPreparationInstrumentRequest{ValidPreparationInstrumentId: created.ID})
		assert.NoError(t, err)

		x, err := adminClient.GetValidPreparationInstrument(ctx, &mealplanningsvc.GetValidPreparationInstrumentRequest{ValidPreparationInstrumentId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidPreparationInstrumentForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveValidPreparationInstrument(ctx, &mealplanningsvc.ArchiveValidPreparationInstrumentRequest{ValidPreparationInstrumentId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveValidPreparationInstrument(ctx, &mealplanningsvc.ArchiveValidPreparationInstrumentRequest{ValidPreparationInstrumentId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidPreparationInstrumentForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveValidPreparationInstrument(ctx, &mealplanningsvc.ArchiveValidPreparationInstrumentRequest{ValidPreparationInstrumentId: created.ID})
		assert.Error(t, err)
	})
}

func TestValidPreparationInstruments_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidPreparationInstrumentForTest(t)

		updateInput := fakes.BuildFakeValidPreparationInstrumentUpdateRequestInput()
		updateInput.ValidPreparationID = &created.Preparation.ID
		updateInput.ValidInstrumentID = &created.Instrument.ID

		response, err := adminClient.UpdateValidPreparationInstrument(ctx, &mealplanningsvc.UpdateValidPreparationInstrumentRequest{
			ValidPreparationInstrumentId: created.ID,
			Input:                        mealplanningconverters.ConvertValidPreparationInstrumentUpdateRequestInputToGRPCValidPreparationInstrumentUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)
		require.NotNil(t, response)
		require.NotNil(t, response.Result)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidPreparationInstrumentForTest(t)

		updateInput := fakes.BuildFakeValidPreparationInstrumentUpdateRequestInput()
		updateInput.ValidPreparationID = &created.Preparation.ID
		updateInput.ValidInstrumentID = &created.Instrument.ID

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateValidPreparationInstrument(ctx, &mealplanningsvc.UpdateValidPreparationInstrumentRequest{
			ValidPreparationInstrumentId: created.ID,
			Input:                        mealplanningconverters.ConvertValidPreparationInstrumentUpdateRequestInputToGRPCValidPreparationInstrumentUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from updating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		_, _, created := createValidPreparationInstrumentForTest(t)

		response, err := testClient.UpdateValidPreparationInstrument(ctx, &mealplanningsvc.UpdateValidPreparationInstrumentRequest{
			ValidPreparationInstrumentId: created.ID,
			Input: &mealplanningsvc.ValidPreparationInstrumentUpdateRequestInput{
				Notes: new("doesn't matter"),
			},
		})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestValidPreparationInstruments_Listing(T *testing.T) {
	T.Parallel()

	createdValidPreparationInstruments := []*types.ValidPreparationInstrument{}
	validPreparation, validInstrument, created := createValidPreparationInstrumentForTest(T)
	createdValidPreparationInstruments = append(createdValidPreparationInstruments, created)
	// Create more VPIs, each with unique (preparation, instrument) pairs to satisfy
	// idx_valid_preparation_instruments_prep_instrument_active unique constraint.
	for range exampleQuantity - 1 {
		_, _, vpi := createValidPreparationInstrumentForTest(T)
		createdValidPreparationInstruments = append(createdValidPreparationInstruments, vpi)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidPreparationInstruments(ctx, &mealplanningsvc.GetValidPreparationInstrumentsRequest{})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), len(createdValidPreparationInstruments))
	})

	T.Run("by Instrument", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidPreparationInstrumentsByInstrument(ctx, &mealplanningsvc.GetValidPreparationInstrumentsByInstrumentRequest{ValidInstrumentId: validInstrument.ID})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 1, "filter by instrument should return at least the created VPI")
	})

	T.Run("by preparation", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidPreparationInstrumentsByPreparation(ctx, &mealplanningsvc.GetValidPreparationInstrumentsByPreparationRequest{ValidPreparationId: validPreparation.ID})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 1, "filter by preparation should return at least the created VPI")
	})
}
