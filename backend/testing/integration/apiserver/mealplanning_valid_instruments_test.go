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

func checkValidInstrumentEquality(t *testing.T, expected, actual *mealplanning.ValidInstrument) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected ValidInstrument to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected ValidInstrument to have CreatedAt")

	assert.Equal(t, expected.Name, actual.Name, "expected ValidInstrument Name")
	assert.Equal(t, expected.Description, actual.Description, "expected ValidInstrument Description")
	assert.Equal(t, expected.Slug, actual.Slug, "expected ValidInstrument Slug")
	assert.Equal(t, expected.PluralName, actual.PluralName, "expected ValidInstrument PluralName")
	assert.Equal(t, expected.IconPath, actual.IconPath, "expected ValidInstrument IconPath")
	assert.Equal(t, expected.DisplayInSummaryLists, actual.DisplayInSummaryLists, "expected ValidInstrument DisplayInSummaryLists")
	assert.Equal(t, expected.IncludeInGeneratedInstructions, actual.IncludeInGeneratedInstructions, "expected ValidInstrument IncludeInGeneratedInstructions")
	assert.Equal(t, expected.UsableForStorage, actual.UsableForStorage, "expected ValidInstrument UsableForStorage")
}

func createValidInstrumentForTest(t *testing.T) *mealplanning.ValidInstrument {
	t.Helper()

	ctx := t.Context()

	exampleValidInstrument := fakes.BuildFakeValidInstrument()
	exampleValidInstrumentInput := converters.ConvertValidInstrumentToValidInstrumentCreationRequestInput(exampleValidInstrument)
	created, err := adminClient.CreateValidInstrument(ctx, &mealplanningsvc.CreateValidInstrumentRequest{
		Input: grpcconverters.ConvertValidInstrumentCreationRequestInputToGRPCValidInstrumentCreationRequestInput(exampleValidInstrumentInput),
	})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCValidInstrumentToValidInstrument(created.Result)
	checkValidInstrumentEquality(t, exampleValidInstrument, converted)

	retrieved, err := adminClient.GetValidInstrument(ctx, &mealplanningsvc.GetValidInstrumentRequest{
		ValidInstrumentId: converted.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	validInstrument := grpcconverters.ConvertGRPCValidInstrumentToValidInstrument(retrieved.Result)
	checkValidInstrumentEquality(t, converted, validInstrument)

	return validInstrument
}

func TestValidInstruments_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createValidInstrumentForTest(t)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeValidInstrumentCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidInstrumentCreationRequestInputToGRPCValidInstrumentCreationRequestInput(creationRequestInput)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateValidInstrument(ctx, &mealplanningsvc.CreateValidInstrumentRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeValidInstrumentCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidInstrumentCreationRequestInputToGRPCValidInstrumentCreationRequestInput(creationRequestInput)
		// this is not allowed
		convertedInput.Name = ""

		created, err := adminClient.CreateValidInstrument(ctx, &mealplanningsvc.CreateValidInstrumentRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		creationRequestInput := fakes.BuildFakeValidInstrumentCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidInstrumentCreationRequestInputToGRPCValidInstrumentCreationRequestInput(creationRequestInput)

		created, err := testClient.CreateValidInstrument(ctx, &mealplanningsvc.CreateValidInstrumentRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestValidInstruments_Reading(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidInstrumentForTest(t)

		retrieved, err := testClient.GetValidInstrument(ctx, &mealplanningsvc.GetValidInstrumentRequest{ValidInstrumentId: created.ID})
		assert.NoError(t, err)

		converted := grpcconverters.ConvertGRPCValidInstrumentToValidInstrument(retrieved.Result)

		assertRoughEquality(t, created, converted, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidInstrumentForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetValidInstrument(ctx, &mealplanningsvc.GetValidInstrumentRequest{ValidInstrumentId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetValidInstrument(ctx, &mealplanningsvc.GetValidInstrumentRequest{ValidInstrumentId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestValidInstruments_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidInstrumentForTest(t)

		updateInput := fakes.BuildFakeValidInstrumentUpdateRequestInput()
		created.Update(updateInput)

		response, err := adminClient.UpdateValidInstrument(ctx, &mealplanningsvc.UpdateValidInstrumentRequest{
			ValidInstrumentId: created.ID,
			Input:             grpcconverters.ConvertValidInstrumentUpdateRequestInputToGRPCValidInstrumentUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)

		updated := grpcconverters.ConvertGRPCValidInstrumentToValidInstrument(response.Result)
		// Ensure UpdatedAt was set
		require.NotNil(t, updated.LastUpdatedAt)

		assertRoughEquality(t, created, updated, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidInstrumentForTest(t)

		updateInput := fakes.BuildFakeValidInstrumentUpdateRequestInput()
		created.Update(updateInput)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateValidInstrument(ctx, &mealplanningsvc.UpdateValidInstrumentRequest{
			ValidInstrumentId: created.ID,
			Input:             grpcconverters.ConvertValidInstrumentUpdateRequestInputToGRPCValidInstrumentUpdateRequestInput(updateInput),
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

		created := createValidInstrumentForTest(t)

		response, err := testClient.UpdateValidInstrument(ctx, &mealplanningsvc.UpdateValidInstrumentRequest{
			ValidInstrumentId: created.ID,
			Input: &mealplanningsvc.ValidInstrumentUpdateRequestInput{
				Name: new("doesn't matter"),
			},
		})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestValidInstruments_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidInstrumentForTest(t)

		_, err := adminClient.ArchiveValidInstrument(ctx, &mealplanningsvc.ArchiveValidInstrumentRequest{ValidInstrumentId: created.ID})
		assert.NoError(t, err)

		x, err := adminClient.GetValidInstrument(ctx, &mealplanningsvc.GetValidInstrumentRequest{ValidInstrumentId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidInstrumentForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveValidInstrument(ctx, &mealplanningsvc.ArchiveValidInstrumentRequest{ValidInstrumentId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveValidInstrument(ctx, &mealplanningsvc.ArchiveValidInstrumentRequest{ValidInstrumentId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidInstrumentForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveValidInstrument(ctx, &mealplanningsvc.ArchiveValidInstrumentRequest{ValidInstrumentId: created.ID})
		assert.Error(t, err)
	})
}

func TestValidInstruments_GetRandom(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// in case we haven't already
		createValidInstrumentForTest(t)

		response, err := testClient.GetRandomValidInstrument(ctx, &mealplanningsvc.GetRandomValidInstrumentRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, response)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		response, err := c.GetRandomValidInstrument(ctx, &mealplanningsvc.GetRandomValidInstrumentRequest{})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestValidInstruments_Listing(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)
	createdValidInstruments := []*mealplanning.ValidInstrument{}
	for range exampleQuantity {
		created := createValidInstrumentForTest(T)
		createdValidInstruments = append(createdValidInstruments, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.GetValidInstruments(ctx, &mealplanningsvc.GetValidInstrumentsRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.True(t, len(retrieved.Results) >= len(createdValidInstruments))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetValidInstruments(ctx, &mealplanningsvc.GetValidInstrumentsRequest{})
		assert.Error(t, err)
	})
}

func TestValidInstruments_Searching(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)
		created := createValidInstrumentForTest(t)

		retrieved, err := testClient.SearchForValidInstruments(ctx, &mealplanningsvc.SearchForValidInstrumentsRequest{
			Query: created.Name[:2],
		})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.SearchForValidInstruments(ctx, &mealplanningsvc.SearchForValidInstrumentsRequest{})
		assert.Error(t, err)
	})
}
