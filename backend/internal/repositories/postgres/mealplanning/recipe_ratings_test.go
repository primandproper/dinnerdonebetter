package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRecipeRatingForTest(t *testing.T, ctx context.Context, exampleRecipeRating *types.RecipeRating, dbc *repository) *types.RecipeRating {
	t.Helper()

	// create
	if exampleRecipeRating == nil {
		exampleRecipeRating = fakes.BuildFakeRecipeRating()
	}
	dbInput := converters.ConvertRecipeRatingToRecipeRatingDatabaseCreationInput(exampleRecipeRating)

	created, err := dbc.CreateRecipeRating(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleRecipeRating.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleRecipeRating, created)

	recipeRating, err := dbc.GetRecipeRating(ctx, created.BelongsToRecipe, created.ID)
	exampleRecipeRating.CreatedAt = recipeRating.CreatedAt

	assert.NoError(t, err)
	assert.Equal(t, recipeRating, exampleRecipeRating)

	return created
}

func TestQuerier_Integration_RecipeRatings(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, _, container := buildDatabaseClientForTest(t)

	databaseURI, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, databaseURI)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	exampleRecipe := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
	createdRecipe := createRecipeForTest(t, ctx, exampleRecipe, dbc, true)

	exampleRecipeRating := fakes.BuildFakeRecipeRating()
	exampleRecipeRating.CreatedByUser = user.ID
	exampleRecipeRating.BelongsToRecipe = createdRecipe.ID
	createdRecipeRatings := []*types.RecipeRating{}

	// create
	createdRecipeRatings = append(createdRecipeRatings, createRecipeRatingForTest(t, ctx, exampleRecipeRating, dbc))

	// fetch as list
	recipeRatings, err := dbc.GetRecipeRatingsForRecipe(ctx, createdRecipe.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, recipeRatings.Data)
	assert.Equal(t, len(createdRecipeRatings), len(recipeRatings.Data))

	// delete
	for _, recipeRating := range createdRecipeRatings {
		assert.NoError(t, dbc.ArchiveRecipeRating(ctx, createdRecipe.ID, recipeRating.ID))

		var exists bool
		exists, err = dbc.RecipeRatingExists(ctx, createdRecipe.ID, recipeRating.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.RecipeRating
		y, err = dbc.GetRecipeRating(ctx, createdRecipe.ID, recipeRating.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_RecipeRatingExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.RecipeRatingExists(ctx, "", t.Name())
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid recipe rating ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.RecipeRatingExists(ctx, t.Name(), "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeRating(ctx, "", t.Name())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe rating ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeRating(ctx, t.Name(), "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateRecipeRating(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateRecipeRating(ctx, nil))
	})
}

func TestQuerier_ArchiveRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeRating(ctx, "", t.Name()))
	})

	T.Run("with invalid recipe rating ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeRating(ctx, t.Name(), ""))
	})
}

func TestQuerier_Integration_RecipeRatings_CursorBasedPagination(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, _, container := buildDatabaseClientForTest(t)

	databaseURI, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, databaseURI)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	recipe := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.RecipeRating]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "recipe rating",
		CreateItem: func(ctx context.Context, i int) *types.RecipeRating {
			// Create a unique user for each rating since there's a unique constraint on (by_user, recipe_id)
			ratingUser := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
			recipeRating := fakes.BuildFakeRecipeRating()
			recipeRating.BelongsToRecipe = recipe.ID
			recipeRating.CreatedByUser = ratingUser.ID
			return createRecipeRatingForTest(t, ctx, recipeRating, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeRating], error) {
			return dbc.GetRecipeRatingsForRecipe(ctx, recipe.ID, filter)
		},
		GetID: func(recipeRating *types.RecipeRating) string {
			return recipeRating.ID
		},
		CleanupItem: func(ctx context.Context, recipeRating *types.RecipeRating) error {
			return dbc.ArchiveRecipeRating(ctx, recipe.ID, recipeRating.ID)
		},
	})
}
