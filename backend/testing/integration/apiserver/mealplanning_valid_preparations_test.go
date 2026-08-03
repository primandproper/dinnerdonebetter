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

func createValidPreparationForTest(t *testing.T) *mealplanning.ValidPreparation {
	t.Helper()

	ctx := t.Context()

	exampleValidPreparation := fakes.BuildFakeValidPreparation()
	exampleValidPreparationInput := converters.ConvertValidPreparationToValidPreparationCreationRequestInput(exampleValidPreparation)
	created, err := adminClient.CreateValidPreparation(ctx, &mealplanningsvc.CreateValidPreparationRequest{
		Input: grpcconverters.ConvertValidPreparationCreationRequestInputToGRPCValidPreparationCreationRequestInput(exampleValidPreparationInput),
	})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCValidPreparationToValidPreparation(created.Result)
	checkValidPreparationEquality(t, 0, exampleValidPreparation, converted)

	retrieved, err := adminClient.GetValidPreparation(ctx, &mealplanningsvc.GetValidPreparationRequest{
		ValidPreparationId: converted.ID,
	})
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	validPreparation := grpcconverters.ConvertGRPCValidPreparationToValidPreparation(retrieved.Result)
	checkValidPreparationEquality(t, 0, converted, validPreparation)

	return validPreparation
}

func checkValidPreparationEquality(t *testing.T, i int, expected, actual *mealplanning.ValidPreparation) {
	t.Helper()

	assert.NotZero(t, actual.CreatedAt, "expected recipe step %d preparation to have CreatedAt", i)
	assert.Equal(t, expected.MinInstrumentCount, actual.MinInstrumentCount, "expected recipe step %d preparation MinInstrumentCount to be %v, but it was %v", i, expected.MinInstrumentCount, actual.MinInstrumentCount)
	assert.Equal(t, expected.MaxInstrumentCount, actual.MaxInstrumentCount, "expected recipe step %d preparation MaxInstrumentCount to be %v, but it was %v", i, expected.MaxInstrumentCount, actual.MaxInstrumentCount)
	assert.Equal(t, expected.MinIngredientCount, actual.MinIngredientCount, "expected recipe step %d preparation MinIngredientCount to be %v, but it was %v", i, expected.MinIngredientCount, actual.MinIngredientCount)
	assert.Equal(t, expected.MaxIngredientCount, actual.MaxIngredientCount, "expected recipe step %d preparation MaxIngredientCount to be %v, but it was %v", i, expected.MaxIngredientCount, actual.MaxIngredientCount)
	assert.Equal(t, expected.MinVesselCount, actual.MinVesselCount, "expected recipe step %d preparation MinVesselCount to be %v, but it was %v", i, expected.MinVesselCount, actual.MinVesselCount)
	assert.Equal(t, expected.MaxVesselCount, actual.MaxVesselCount, "expected recipe step %d preparation MaxVesselCount to be %v, but it was %v", i, expected.MaxVesselCount, actual.MaxVesselCount)
	assert.Equal(t, expected.IconPath, actual.IconPath, "expected recipe step %d preparation IconPath to be %v, but it was %v", i, expected.IconPath, actual.IconPath)
	assert.Equal(t, expected.PastTense, actual.PastTense, "expected recipe step %d preparation PastTense to be %v, but it was %v", i, expected.PastTense, actual.PastTense)
	assert.NotEmpty(t, actual.ID)
	assert.Equal(t, expected.Name, actual.Name, "expected recipe step %d preparation Name to be %v, but it was %v", i, expected.Name, actual.Name)
	assert.Equal(t, expected.Description, actual.Description, "expected recipe step %d preparation Description to be %v, but it was %v", i, expected.Description, actual.Description)
	assert.Equal(t, expected.Slug, actual.Slug, "expected recipe step %d preparation Slug to be %v, but it was %v", i, expected.Slug, actual.Slug)
	assert.Equal(t, expected.RestrictToIngredients, actual.RestrictToIngredients, "expected recipe step %d preparation RestrictToIngredients to be %v, but it was %v", i, expected.RestrictToIngredients, actual.RestrictToIngredients)
	assert.Equal(t, expected.TemperatureRequired, actual.TemperatureRequired, "expected recipe step %d preparation TemperatureRequired to be %v, but it was %v", i, expected.TemperatureRequired, actual.TemperatureRequired)
	assert.Equal(t, expected.TimeEstimateRequired, actual.TimeEstimateRequired, "expected recipe step %d preparation TimeEstimateRequired to be %v, but it was %v", i, expected.TimeEstimateRequired, actual.TimeEstimateRequired)
	assert.Equal(t, expected.ConditionExpressionRequired, actual.ConditionExpressionRequired, "expected recipe step %d preparation ConditionExpressionRequired to be %v, but it was %v", i, expected.ConditionExpressionRequired, actual.ConditionExpressionRequired)
	assert.Equal(t, expected.ConsumesVessel, actual.ConsumesVessel, "expected recipe step %d preparation ConsumesVessel to be %v, but it was %v", i, expected.ConsumesVessel, actual.ConsumesVessel)
	assert.Equal(t, expected.OnlyForVessels, actual.OnlyForVessels, "expected recipe step %d preparation OnlyForVessels to be %v, but it was %v", i, expected.OnlyForVessels, actual.OnlyForVessels)
	assert.Equal(t, expected.YieldsNothing, actual.YieldsNothing, "expected recipe step %d preparation YieldsNothing to be %v, but it was %v", i, expected.YieldsNothing, actual.YieldsNothing)
}

func TestValidPreparations_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createValidPreparationForTest(t)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeValidPreparationCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidPreparationCreationRequestInputToGRPCValidPreparationCreationRequestInput(creationRequestInput)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateValidPreparation(ctx, &mealplanningsvc.CreateValidPreparationRequest{
			Input: convertedInput,
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		creationRequestInput := fakes.BuildFakeValidPreparationCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidPreparationCreationRequestInputToGRPCValidPreparationCreationRequestInput(creationRequestInput)
		// this is not allowed
		convertedInput.Name = ""

		created, err := adminClient.CreateValidPreparation(ctx, &mealplanningsvc.CreateValidPreparationRequest{
			Input: convertedInput,
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		creationRequestInput := fakes.BuildFakeValidPreparationCreationRequestInput()
		convertedInput := grpcconverters.ConvertValidPreparationCreationRequestInputToGRPCValidPreparationCreationRequestInput(creationRequestInput)

		created, err := testClient.CreateValidPreparation(ctx, &mealplanningsvc.CreateValidPreparationRequest{
			Input: convertedInput,
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestValidPreparations_Reading(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidPreparationForTest(t)

		retrieved, err := testClient.GetValidPreparation(ctx, &mealplanningsvc.GetValidPreparationRequest{ValidPreparationId: created.ID})
		require.NoError(t, err)

		converted := grpcconverters.ConvertGRPCValidPreparationToValidPreparation(retrieved.Result)

		assertRoughEquality(t, created, converted, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidPreparationForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetValidPreparation(ctx, &mealplanningsvc.GetValidPreparationRequest{ValidPreparationId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetValidPreparation(ctx, &mealplanningsvc.GetValidPreparationRequest{ValidPreparationId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestValidPreparations_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidPreparationForTest(t)

		updateInput := fakes.BuildFakeValidPreparationUpdateRequestInput()
		created.Update(updateInput)

		response, err := adminClient.UpdateValidPreparation(ctx, &mealplanningsvc.UpdateValidPreparationRequest{
			ValidPreparationId: created.ID,
			Input:              grpcconverters.ConvertValidPreparationUpdateRequestInputToGRPCValidPreparationUpdateRequestInput(updateInput),
		})
		require.NoError(t, err)

		updated := grpcconverters.ConvertGRPCValidPreparationToValidPreparation(response.Result)
		// Ensure UpdatedAt was set
		require.NotNil(t, updated.LastUpdatedAt)

		assertRoughEquality(t, created, updated, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidPreparationForTest(t)

		updateInput := fakes.BuildFakeValidPreparationUpdateRequestInput()
		created.Update(updateInput)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateValidPreparation(ctx, &mealplanningsvc.UpdateValidPreparationRequest{
			ValidPreparationId: created.ID,
			Input:              grpcconverters.ConvertValidPreparationUpdateRequestInputToGRPCValidPreparationUpdateRequestInput(updateInput),
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

		created := createValidPreparationForTest(t)

		response, err := testClient.UpdateValidPreparation(ctx, &mealplanningsvc.UpdateValidPreparationRequest{
			ValidPreparationId: created.ID,
			Input: &mealplanningsvc.ValidPreparationUpdateRequestInput{
				Name: new("doesn't matter"),
			},
		})
		require.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestValidPreparations_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidPreparationForTest(t)

		_, err := adminClient.ArchiveValidPreparation(ctx, &mealplanningsvc.ArchiveValidPreparationRequest{ValidPreparationId: created.ID})
		require.NoError(t, err)

		x, err := adminClient.GetValidPreparation(ctx, &mealplanningsvc.GetValidPreparationRequest{ValidPreparationId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidPreparationForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveValidPreparation(ctx, &mealplanningsvc.ArchiveValidPreparationRequest{ValidPreparationId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveValidPreparation(ctx, &mealplanningsvc.ArchiveValidPreparationRequest{ValidPreparationId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createValidPreparationForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveValidPreparation(ctx, &mealplanningsvc.ArchiveValidPreparationRequest{ValidPreparationId: created.ID})
		assert.Error(t, err)
	})
}

func TestValidPreparations_GetRandom(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// in case we haven't already
		createValidPreparationForTest(t)

		response, err := testClient.GetRandomValidPreparation(ctx, &mealplanningsvc.GetRandomValidPreparationRequest{})
		require.NoError(t, err)
		assert.NotNil(t, response)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		response, err := c.GetRandomValidPreparation(ctx, &mealplanningsvc.GetRandomValidPreparationRequest{})
		require.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestValidPreparations_Listing(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)
	createdValidPreparations := []*mealplanning.ValidPreparation{}
	for range exampleQuantity {
		created := createValidPreparationForTest(T)
		createdValidPreparations = append(createdValidPreparations, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.GetValidPreparations(ctx, &mealplanningsvc.GetValidPreparationsRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.GreaterOrEqual(t, len(retrieved.Results), len(createdValidPreparations))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetValidPreparations(ctx, &mealplanningsvc.GetValidPreparationsRequest{})
		assert.Error(t, err)
	})
}

func TestValidPreparations_Searching(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)
		created := createValidPreparationForTest(t)

		retrieved, err := testClient.SearchForValidPreparations(ctx, &mealplanningsvc.SearchForValidPreparationsRequest{
			Query: created.Name[:2],
		})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.SearchForValidPreparations(ctx, &mealplanningsvc.SearchForValidPreparationsRequest{})
		assert.Error(t, err)
	})
}
