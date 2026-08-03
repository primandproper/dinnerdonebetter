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

func createValidIngredientStateIngredientForTest(t *testing.T) (*types.ValidIngredientState, *types.ValidIngredient, *types.ValidIngredientStateIngredient) {
	t.Helper()
	ctx := t.Context()

	createdValidIngredientState := createValidIngredientStateForTest(t)
	createdValidIngredient := createValidIngredientForTest(t)

	exampleValidIngredientStateIngredient := fakes.BuildFakeValidIngredientStateIngredient()
	exampleValidIngredientStateIngredient.Ingredient = *createdValidIngredient
	exampleValidIngredientStateIngredient.IngredientState = *createdValidIngredientState

	exampleValidIngredientStateIngredientInput := mealplanningconverters.ConvertCreateValidIngredientStateIngredientRequestToGRPCValidIngredientStateIngredientCreationRequestInput(converters.ConvertValidIngredientStateIngredientToValidIngredientStateIngredientCreationRequestInput(exampleValidIngredientStateIngredient))
	createdValidIngredientStateIngredient, err := adminClient.CreateValidIngredientStateIngredient(ctx, &mealplanningsvc.CreateValidIngredientStateIngredientRequest{Input: exampleValidIngredientStateIngredientInput})
	require.NoError(t, err)
	require.NotNil(t, createdValidIngredientStateIngredient)

	validPrepIngredientRes, err := adminClient.GetValidIngredientStateIngredient(ctx, &mealplanningsvc.GetValidIngredientStateIngredientRequest{
		ValidIngredientStateIngredientId: createdValidIngredientStateIngredient.Result.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, validPrepIngredientRes.Result)

	return createdValidIngredientState, createdValidIngredient, mealplanningconverters.ConvertGRPCValidIngredientStateIngredientToValidIngredientStateIngredient(validPrepIngredientRes.Result)
}

func TestValidIngredientStateIngredients_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		createValidIngredientStateIngredientForTest(t)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		exampleValidIngredientStateIngredient := fakes.BuildFakeValidIngredientStateIngredient()
		exampleValidIngredientStateIngredientInput := mealplanningconverters.ConvertCreateValidIngredientStateIngredientRequestToGRPCValidIngredientStateIngredientCreationRequestInput(converters.ConvertValidIngredientStateIngredientToValidIngredientStateIngredientCreationRequestInput(exampleValidIngredientStateIngredient))
		exampleValidIngredientStateIngredientInput.ValidIngredientId = ""
		exampleValidIngredientStateIngredientInput.ValidIngredientStateId = ""

		createdValidIngredientStateIngredient, err := adminClient.CreateValidIngredientStateIngredient(ctx, &mealplanningsvc.CreateValidIngredientStateIngredientRequest{Input: exampleValidIngredientStateIngredientInput})
		require.Error(t, err)
		require.Nil(t, createdValidIngredientStateIngredient)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreateValidIngredientStateIngredient(ctx, &mealplanningsvc.CreateValidIngredientStateIngredientRequest{})
		assert.Error(t, err)
	})
}

func TestValidIngredientStateIngredients_Listing(T *testing.T) {
	T.Parallel()

	createdValidIngredientStateIngredients := []*types.ValidIngredientStateIngredient{}
	validIngredientState, validIngredient, created := createValidIngredientStateIngredientForTest(T)
	createdValidIngredientStateIngredients = append(createdValidIngredientStateIngredients, created)
	// Create additional VSI entries with unique (ingredient, state) pairs - use same state, different ingredients for "by ingredient state" filter
	for range exampleQuantity - 1 {
		extraIngredient := createValidIngredientForTest(T)
		exampleValidIngredientStateIngredient := fakes.BuildFakeValidIngredientStateIngredient()
		exampleValidIngredientStateIngredient.Ingredient = *extraIngredient
		exampleValidIngredientStateIngredient.IngredientState = *validIngredientState
		exampleValidIngredientStateIngredientInput := mealplanningconverters.ConvertCreateValidIngredientStateIngredientRequestToGRPCValidIngredientStateIngredientCreationRequestInput(converters.ConvertValidIngredientStateIngredientToValidIngredientStateIngredientCreationRequestInput(exampleValidIngredientStateIngredient))
		createdValidIngredientStateIngredient, err := adminClient.CreateValidIngredientStateIngredient(T.Context(), &mealplanningsvc.CreateValidIngredientStateIngredientRequest{Input: exampleValidIngredientStateIngredientInput})
		require.NoError(T, err)
		require.NotNil(T, createdValidIngredientStateIngredient)
		createdValidIngredientStateIngredients = append(createdValidIngredientStateIngredients, mealplanningconverters.ConvertGRPCValidIngredientStateIngredientToValidIngredientStateIngredient(createdValidIngredientStateIngredient.Result))
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidIngredientStateIngredients(ctx, &mealplanningsvc.GetValidIngredientStateIngredientsRequest{})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), len(createdValidIngredientStateIngredients))
	})

	T.Run("by ingredient", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidIngredientStateIngredientsByIngredient(ctx, &mealplanningsvc.GetValidIngredientStateIngredientsByIngredientRequest{ValidIngredientId: validIngredient.ID})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), 1, "at least one VSI for this ingredient")
	})

	T.Run("by ingredient state", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		results, err := adminClient.GetValidIngredientStateIngredientsByIngredientState(ctx, &mealplanningsvc.GetValidIngredientStateIngredientsByIngredientStateRequest{ValidIngredientStateId: validIngredientState.ID})
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), len(createdValidIngredientStateIngredients))
	})
}

func TestIntegration_UpdateValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientStateIngredientForTest(t)

		updateInput := fakes.BuildFakeValidIngredientStateIngredientUpdateRequestInput()
		updateInput.ValidIngredientStateID = &created.IngredientState.ID
		updateInput.ValidIngredientID = &created.Ingredient.ID
		created.Update(updateInput)

		response, err := adminClient.UpdateValidIngredientStateIngredient(ctx, &mealplanningsvc.UpdateValidIngredientStateIngredientRequest{
			ValidIngredientStateIngredientId: created.ID,
			Input:                            mealplanningconverters.ConvertValidIngredientStateIngredientUpdateRequestInputToGRPCValidIngredientStateIngredientUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)

		updated := mealplanningconverters.ConvertGRPCValidIngredientStateIngredientToValidIngredientStateIngredient(response.Result)
		require.NotNil(t, updated.LastUpdatedAt)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientStateIngredientForTest(t)

		updateInput := fakes.BuildFakeValidIngredientStateIngredientUpdateRequestInput()
		created.Update(updateInput)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateValidIngredientStateIngredient(ctx, &mealplanningsvc.UpdateValidIngredientStateIngredientRequest{
			ValidIngredientStateIngredientId: created.ID,
			Input:                            mealplanningconverters.ConvertValidIngredientStateIngredientUpdateRequestInputToGRPCValidIngredientStateIngredientUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from updating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		_, _, created := createValidIngredientStateIngredientForTest(t)

		updateInput := fakes.BuildFakeValidIngredientStateIngredientUpdateRequestInput()

		response, err := testClient.UpdateValidIngredientStateIngredient(ctx, &mealplanningsvc.UpdateValidIngredientStateIngredientRequest{
			ValidIngredientStateIngredientId: created.ID,
			Input:                            mealplanningconverters.ConvertValidIngredientStateIngredientUpdateRequestInputToGRPCValidIngredientStateIngredientUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestIntegration_ArchiveValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientStateIngredientForTest(t)

		_, err := adminClient.ArchiveValidIngredientStateIngredient(ctx, &mealplanningsvc.ArchiveValidIngredientStateIngredientRequest{ValidIngredientStateIngredientId: created.ID})
		assert.NoError(t, err)

		x, err := adminClient.GetValidIngredientStateIngredient(ctx, &mealplanningsvc.GetValidIngredientStateIngredientRequest{ValidIngredientStateIngredientId: created.ID})
		assert.Nil(t, x)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientStateIngredientForTest(t)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveValidIngredientStateIngredient(ctx, &mealplanningsvc.ArchiveValidIngredientStateIngredientRequest{ValidIngredientStateIngredientId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.ArchiveValidIngredientStateIngredient(ctx, &mealplanningsvc.ArchiveValidIngredientStateIngredientRequest{ValidIngredientStateIngredientId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("non-admin users are forbidden from archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, created := createValidIngredientStateIngredientForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveValidIngredientStateIngredient(ctx, &mealplanningsvc.ArchiveValidIngredientStateIngredientRequest{ValidIngredientStateIngredientId: created.ID})
		assert.Error(t, err)
	})
}
