package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRecipeMediaForTest(t *testing.T, ctx context.Context, exampleRecipeMedia *types.RecipeMedia, dbc *repository) *types.RecipeMedia {
	t.Helper()

	// create
	if exampleRecipeMedia == nil {
		exampleRecipeMedia = fakes.BuildFakeRecipeMedia()
	}
	dbInput := converters.ConvertRecipeMediaToRecipeMediaDatabaseCreationInput(exampleRecipeMedia)

	created, err := dbc.CreateRecipeMedia(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleRecipeMedia.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleRecipeMedia, created)

	recipeMedia, err := dbc.GetRecipeMedia(ctx, created.ID)
	exampleRecipeMedia.CreatedAt = recipeMedia.CreatedAt

	assert.NoError(t, err)
	assert.Equal(t, recipeMedia, exampleRecipeMedia)

	return created
}

func TestQuerier_Integration_RecipeMedia(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	exampleRecipe := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
	createdRecipe := createRecipeForTest(t, ctx, exampleRecipe, dbc, true)

	exampleRecipeMedia := fakes.BuildFakeRecipeMedia()
	exampleRecipeMedia.BelongsToRecipe = &createdRecipe.ID
	createdRecipeMedias := []*types.RecipeMedia{}

	// create
	createdRecipeMedias = append(createdRecipeMedias, createRecipeMediaForTest(t, ctx, exampleRecipeMedia, dbc))

	// fetch as list
	recipeMediaList, err := dbc.getRecipeMediaForRecipe(ctx, exampleRecipe.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, recipeMediaList)
	assert.Equal(t, len(createdRecipeMedias), len(recipeMediaList))

	// delete
	for _, recipeMedia := range createdRecipeMedias {
		assert.NoError(t, dbc.ArchiveRecipeMedia(ctx, recipeMedia.ID))

		var exists bool
		exists, err = dbc.RecipeMediaExists(ctx, recipeMedia.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.RecipeMedia
		y, err = dbc.GetRecipeMedia(ctx, recipeMedia.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_RecipeMediaExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe media MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeMediaExists(ctx, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetRecipeMedia(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe media MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeMedia(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateRecipeMedia(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateRecipeMedia(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateRecipeMedia(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateRecipeMedia(ctx, nil))
	})
}

func TestQuerier_ArchiveRecipeMedia(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe media MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeMedia(ctx, ""))
	})
}
