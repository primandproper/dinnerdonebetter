package integration

import (
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mpconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRecipeRatingForTest(t *testing.T, recipeID string, clientToUse client.Client) *types.RecipeRating {
	t.Helper()
	ctx := t.Context()

	exampleRecipeRating := fakes.BuildFakeRecipeRating()
	exampleRecipeRating.BelongsToRecipe = recipeID
	exampleRecipeRatingInput := mpconverters.ConvertRecipeRatingToRecipeRatingCreationRequestInput(exampleRecipeRating)
	convertedInput := converters.ConvertRecipeRatingCreationRequestInputToGRPCRecipeRatingCreationRequestInput(exampleRecipeRatingInput)

	createdRecipeRating, err := clientToUse.CreateRecipeRating(ctx, &mealplanninggrpc.CreateRecipeRatingRequest{
		RecipeId: recipeID,
		Input:    convertedInput,
	})
	require.NoError(t, err)
	require.NotNil(t, createdRecipeRating)

	converted := converters.ConvertGRPCRecipeRatingToRecipeRating(createdRecipeRating.Created)
	assertRoughEquality(t, exampleRecipeRating, converted, defaultIgnoredFields("ID", "CreatedByUser")...)

	return converted
}

func checkRecipeRatingEquality(t *testing.T, expected, actual *types.RecipeRating) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Notes, actual.Notes, "expected Notes for recipe rating %s to be %v, but it was %v", expected.ID, expected.Notes, actual.Notes)
	assert.Equal(t, expected.BelongsToRecipe, actual.BelongsToRecipe, "expected RecipeID for recipe rating %s to be %v, but it was %v", expected.ID, expected.BelongsToRecipe, actual.BelongsToRecipe)
	assert.Equal(t, expected.Taste, actual.Taste, "expected Taste for recipe rating %s to be %v, but it was %v", expected.ID, expected.Taste, actual.Taste)
	assert.Equal(t, expected.Instructions, actual.Instructions, "expected Instructions for recipe rating %s to be %v, but it was %v", expected.ID, expected.Instructions, actual.Instructions)
	assert.Equal(t, expected.Overall, actual.Overall, "expected Overall for recipe rating %s to be %v, but it was %v", expected.ID, expected.Overall, actual.Overall)
	assert.Equal(t, expected.Cleanup, actual.Cleanup, "expected Cleanup for recipe rating %s to be %v, but it was %v", expected.ID, expected.Cleanup, actual.Cleanup)
	assert.Equal(t, expected.Difficulty, actual.Difficulty, "expected Difficulty for recipe rating %s to be %v, but it was %v", expected.ID, expected.Difficulty, actual.Difficulty)
	assert.NotZero(t, actual.CreatedAt)
}

func TestRecipeRatings_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)
		exampleRecipeRating := createRecipeRatingForTest(t, createdRecipe.ID, testClient)

		newRecipeRating := fakes.BuildFakeRecipeRating()
		newRecipeRating.BelongsToRecipe = createdRecipe.ID
		newRecipeRating.CreatedByUser = exampleRecipeRating.CreatedByUser

		updateInput := mpconverters.ConvertRecipeRatingToRecipeRatingUpdateRequestInput(newRecipeRating)
		exampleRecipeRating.Update(updateInput)

		_, err := testClient.UpdateRecipeRating(ctx, &mealplanninggrpc.UpdateRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: exampleRecipeRating.ID,
			Input:          converters.ConvertRecipeRatingUpdateRequestInputToGRPCRecipeRatingUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)

		actualRes, err := testClient.GetRecipeRating(ctx, &mealplanninggrpc.GetRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: exampleRecipeRating.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, actualRes)

		actual := converters.ConvertGRPCRecipeRatingToRecipeRating(actualRes.Result)

		// assert recipe rating equality
		checkRecipeRatingEquality(t, newRecipeRating, actual)
		assert.NotNil(t, actual.LastUpdatedAt)

		_, err = testClient.ArchiveRecipeRating(ctx, &mealplanninggrpc.ArchiveRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: exampleRecipeRating.ID,
		})
		assert.NoError(t, err)
	})

	T.Run("requires auth for creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		exampleRecipeRating := fakes.BuildFakeRecipeRating()
		exampleRecipeRating.BelongsToRecipe = createdRecipe.ID
		exampleRecipeRatingInput := mpconverters.ConvertRecipeRatingToRecipeRatingCreationRequestInput(exampleRecipeRating)
		convertedInput := converters.ConvertRecipeRatingCreationRequestInputToGRPCRecipeRatingCreationRequestInput(exampleRecipeRatingInput)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateRecipeRating(ctx, &mealplanninggrpc.CreateRecipeRatingRequest{
			RecipeId: createdRecipe.ID,
			Input:    convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)

		// Clean up
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})

	T.Run("requires auth for updating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)
		exampleRecipeRating := createRecipeRatingForTest(t, createdRecipe.ID, testClient)

		newRecipeRating := fakes.BuildFakeRecipeRating()
		newRecipeRating.BelongsToRecipe = createdRecipe.ID
		updateInput := mpconverters.ConvertRecipeRatingToRecipeRatingUpdateRequestInput(newRecipeRating)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdateRecipeRating(ctx, &mealplanninggrpc.UpdateRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: exampleRecipeRating.ID,
			Input:          converters.ConvertRecipeRatingUpdateRequestInputToGRPCRecipeRatingUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)

		// Clean up
		_, err = testClient.ArchiveRecipeRating(ctx, &mealplanninggrpc.ArchiveRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: exampleRecipeRating.ID,
		})
		assert.NoError(t, err)
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})

	T.Run("requires auth for archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)
		exampleRecipeRating := createRecipeRatingForTest(t, createdRecipe.ID, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveRecipeRating(ctx, &mealplanninggrpc.ArchiveRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: exampleRecipeRating.ID,
		})
		assert.Error(t, err)

		// Clean up
		_, err = testClient.ArchiveRecipeRating(ctx, &mealplanninggrpc.ArchiveRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: exampleRecipeRating.ID,
		})
		assert.NoError(t, err)
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})

	T.Run("nonexistent recipe rating for updating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)

		newRecipeRating := fakes.BuildFakeRecipeRating()
		newRecipeRating.BelongsToRecipe = createdRecipe.ID
		updateInput := mpconverters.ConvertRecipeRatingToRecipeRatingUpdateRequestInput(newRecipeRating)

		_, err := testClient.UpdateRecipeRating(ctx, &mealplanninggrpc.UpdateRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: nonexistentID,
			Input:          converters.ConvertRecipeRatingUpdateRequestInputToGRPCRecipeRatingUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)

		// Clean up
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})

	T.Run("nonexistent recipe rating for archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.ArchiveRecipeRating(ctx, &mealplanninggrpc.ArchiveRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: nonexistentID,
		})
		assert.Error(t, err)

		// Clean up
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipeRatings_Listing(T *testing.T) {
	T.Parallel()

	T.Run("should be readable in paginated form", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)
		exampleRecipeRating := createRecipeRatingForTest(t, createdRecipe.ID, testClient)

		// assert recipe rating list equality
		actualRes, err := testClient.GetRecipeRatingsForRecipe(ctx, &mealplanninggrpc.GetRecipeRatingsForRecipeRequest{
			RecipeId: createdRecipe.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, actualRes)

		assert.Equal(t, len(actualRes.Results), 1, "expected %d to be <= %d", len(actualRes.Results), 1)

		_, err = testClient.ArchiveRecipeRating(ctx, &mealplanninggrpc.ArchiveRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: exampleRecipeRating.ID,
		})

		assert.NoError(t, err)
	})

	T.Run("requires auth for listing", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		c := buildUnauthenticatedGRPCClientForTest(t)
		ratings, err := c.GetRecipeRatingsForRecipe(ctx, &mealplanninggrpc.GetRecipeRatingsForRecipeRequest{
			RecipeId: createdRecipe.ID,
		})
		assert.Error(t, err)
		assert.Nil(t, ratings)

		// Clean up
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipeRatings_Reading(T *testing.T) {
	T.Parallel()

	T.Run("requires auth for reading", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)
		exampleRecipeRating := createRecipeRatingForTest(t, createdRecipe.ID, testClient)

		c := buildUnauthenticatedGRPCClientForTest(t)
		rating, err := c.GetRecipeRating(ctx, &mealplanninggrpc.GetRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: exampleRecipeRating.ID,
		})
		assert.Error(t, err)
		assert.Nil(t, rating)

		// Clean up
		_, err = testClient.ArchiveRecipeRating(ctx, &mealplanninggrpc.ArchiveRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: exampleRecipeRating.ID,
		})
		assert.NoError(t, err)
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})

	T.Run("nonexistent recipe rating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)

		rating, err := testClient.GetRecipeRating(ctx, &mealplanninggrpc.GetRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: nonexistentID,
		})
		assert.Error(t, err)
		assert.Nil(t, rating)

		// Clean up
		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}
