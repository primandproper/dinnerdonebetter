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

func createValidIngredientPreparationForTest(t *testing.T) (*types.ValidIngredient, *types.ValidPreparation, *types.ValidIngredientPreparation) {
	t.Helper()

	createdValidIngredient := createValidIngredientForTest(t)
	createdValidPreparation := createValidPreparationForTest(t)

	return createdValidIngredient, createdValidPreparation, createValidIngredientPreparationWithEntitiesForTest(t, createdValidIngredient, createdValidPreparation)
}

// createValidIngredientPreparationWithEntitiesForTest creates a ValidIngredientPreparation with specific entities.
func createValidIngredientPreparationWithEntitiesForTest(t *testing.T, ingredient *types.ValidIngredient, preparation *types.ValidPreparation) *types.ValidIngredientPreparation {
	t.Helper()
	ctx := t.Context()

	exampleValidIngredientPreparation := fakes.BuildFakeValidIngredientPreparation()
	exampleValidIngredientPreparation.Preparation = *preparation
	exampleValidIngredientPreparation.Ingredient = *ingredient

	exampleValidIngredientPreparationInput := mealplanningconverters.ConvertCreateValidIngredientPreparationRequestToGRPCValidIngredientPreparationCreationRequestInput(converters.ConvertValidIngredientPreparationToValidIngredientPreparationCreationRequestInput(exampleValidIngredientPreparation))
	createdValidIngredientPreparation, err := adminClient.CreateValidIngredientPreparation(ctx, &mealplanningsvc.CreateValidIngredientPreparationRequest{Input: exampleValidIngredientPreparationInput})
	require.NoError(t, err)
	require.NotNil(t, createdValidIngredientPreparation)

	validPrepPreparationRes, err := adminClient.GetValidIngredientPreparation(ctx, &mealplanningsvc.GetValidIngredientPreparationRequest{
		ValidIngredientPreparationId: createdValidIngredientPreparation.Result.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, validPrepPreparationRes.Result)

	return mealplanningconverters.ConvertGRPCValidIngredientPreparationToValidIngredientPreparation(validPrepPreparationRes.Result)
}

func TestValidIngredientPreparations_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createValidIngredientPreparationForTest(t)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		exampleValidIngredientPreparation := fakes.BuildFakeValidIngredientPreparation()
		exampleValidIngredientPreparationInput := mealplanningconverters.ConvertCreateValidIngredientPreparationRequestToGRPCValidIngredientPreparationCreationRequestInput(converters.ConvertValidIngredientPreparationToValidIngredientPreparationCreationRequestInput(exampleValidIngredientPreparation))
		exampleValidIngredientPreparationInput.ValidPreparationId = ""
		exampleValidIngredientPreparationInput.ValidIngredientId = ""

		createdValidIngredientPreparation, err := adminClient.CreateValidIngredientPreparation(ctx, &mealplanningsvc.CreateValidIngredientPreparationRequest{Input: exampleValidIngredientPreparationInput})
		require.Error(t, err)
		require.Nil(t, createdValidIngredientPreparation)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreateValidIngredientPreparation(ctx, &mealplanningsvc.CreateValidIngredientPreparationRequest{})
		assert.Error(t, err)
	})
}

func TestValidIngredientPreparations_Listing(T *testing.T) {
	T.Parallel()

	createdValidIngredientPreparations := []*types.ValidIngredientPreparation{}
	validIngredient, validPreparation, created := createValidIngredientPreparationForTest(T)
	createdValidIngredientPreparations = append(createdValidIngredientPreparations, created)
	// Create additional VIPs with unique (prep, ingredient) pairs - use same prep, different ingredients for "by Preparation" filter
	for range exampleQuantity - 1 {
		extraIngredient := createValidIngredientForTest(T)
		createdVIP := createValidIngredientPreparationWithEntitiesForTest(T, extraIngredient, validPreparation)
		createdValidIngredientPreparations = append(createdValidIngredientPreparations, createdVIP)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidIngredientPreparations(ctx, &mealplanningsvc.GetValidIngredientPreparationsRequest{})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), len(createdValidIngredientPreparations))
	})

	T.Run("by Preparation", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidIngredientPreparationsByPreparation(ctx, &mealplanningsvc.GetValidIngredientPreparationsByPreparationRequest{ValidPreparationId: validPreparation.ID})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), len(createdValidIngredientPreparations))
	})

	T.Run("by ingredient", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidIngredientPreparationsByIngredient(ctx, &mealplanningsvc.GetValidIngredientPreparationsByIngredientRequest{ValidIngredientId: validIngredient.ID})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 1, "at least one VIP for this ingredient")
	})
}

func TestValidIngredientPreparations_SearchByPreparation(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		createdValidIngredient, createdValidPreparation, _ := createValidIngredientPreparationForTest(t)

		results, err := adminClient.SearchValidIngredientsByPreparation(ctx, &mealplanningsvc.SearchValidIngredientsByPreparationRequest{
			ValidPreparationId: createdValidPreparation.ID,
			Query:              createdValidIngredient.Name[:2],
		})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 1, "expected at least one result when searching by preparation")
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.SearchValidIngredientsByPreparation(ctx, &mealplanningsvc.SearchValidIngredientsByPreparationRequest{})
		assert.Error(t, err)
	})
}

func TestIntegration_UpdateValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientPreparationForTest(t)

		updateInput := fakes.BuildFakeValidIngredientPreparationUpdateRequestInput()
		updateInput.ValidIngredientID = &created.Ingredient.ID
		updateInput.ValidPreparationID = &created.Preparation.ID
		created.Update(updateInput)

		response, err := adminClient.UpdateValidIngredientPreparation(ctx, &mealplanningsvc.UpdateValidIngredientPreparationRequest{
			ValidIngredientPreparationId: created.ID,
			Input:                        mealplanningconverters.ConvertValidIngredientPreparationUpdateRequestInputToGRPCValidIngredientPreparationUpdateRequestInput(updateInput),
		})
		require.NoError(t, err)

		updated := mealplanningconverters.ConvertGRPCValidIngredientPreparationToValidIngredientPreparation(response.Result)
		require.NotNil(t, updated.LastUpdatedAt)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientPreparationForTest(t)

		updateInput := fakes.BuildFakeValidIngredientPreparationUpdateRequestInput()
		created.Update(updateInput)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateValidIngredientPreparation(ctx, &mealplanningsvc.UpdateValidIngredientPreparationRequest{
			ValidIngredientPreparationId: created.ID,
			Input:                        mealplanningconverters.ConvertValidIngredientPreparationUpdateRequestInputToGRPCValidIngredientPreparationUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from updating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		_, _, created := createValidIngredientPreparationForTest(t)

		updateInput := fakes.BuildFakeValidIngredientPreparationUpdateRequestInput()

		response, err := testClient.UpdateValidIngredientPreparation(ctx, &mealplanningsvc.UpdateValidIngredientPreparationRequest{
			ValidIngredientPreparationId: created.ID,
			Input:                        mealplanningconverters.ConvertValidIngredientPreparationUpdateRequestInputToGRPCValidIngredientPreparationUpdateRequestInput(updateInput),
		})
		require.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestIntegration_ArchiveValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientPreparationForTest(t)

		_, err := adminClient.ArchiveValidIngredientPreparation(ctx, &mealplanningsvc.ArchiveValidIngredientPreparationRequest{ValidIngredientPreparationId: created.ID})
		require.NoError(t, err)

		x, err := adminClient.GetValidIngredientPreparation(ctx, &mealplanningsvc.GetValidIngredientPreparationRequest{ValidIngredientPreparationId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientPreparationForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveValidIngredientPreparation(ctx, &mealplanningsvc.ArchiveValidIngredientPreparationRequest{ValidIngredientPreparationId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveValidIngredientPreparation(ctx, &mealplanningsvc.ArchiveValidIngredientPreparationRequest{ValidIngredientPreparationId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientPreparationForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveValidIngredientPreparation(ctx, &mealplanningsvc.ArchiveValidIngredientPreparationRequest{ValidIngredientPreparationId: created.ID})
		assert.Error(t, err)
	})
}
