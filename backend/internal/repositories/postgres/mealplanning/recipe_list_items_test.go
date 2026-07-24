package mealplanning

import (
	"database/sql"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v6/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildRecipeListItemForTest(listID, recipeID string) *mealplanning.RecipeListItemDatabaseCreationInput {
	return &mealplanning.RecipeListItemDatabaseCreationInput{
		ID:                  identifiers.New(),
		RecipeID:            recipeID,
		Notes:               "note1",
		BelongsToRecipeList: listID,
	}
}

func TestIntegration_RecipeListItems(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, _, container := buildDatabaseClientForTest(t)
	defer func() {
		assert.NoError(t, container.Terminate(ctx))
	}()

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	recipe := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)

	listInput := buildRecipeListForTest(user.ID)
	createdList, err := dbc.CreateRecipeList(ctx, listInput)
	require.NoError(t, err)

	itemInput := buildRecipeListItemForTest(createdList.ID, recipe.ID)
	createdItem, err := dbc.CreateRecipeListItem(ctx, itemInput)
	require.NoError(t, err)
	require.Equal(t, itemInput.ID, createdItem.ID)

	items, err := dbc.GetRecipeListItems(ctx, createdList.ID, nil)
	require.NoError(t, err)
	require.Len(t, items.Data, 1)
	assert.Equal(t, recipe.ID, items.Data[0].Recipe.ID)

	require.NoError(t, dbc.ArchiveRecipeListItem(ctx, createdItem.ID, createdList.ID))

	afterArchive, err := dbc.GetRecipeListItems(ctx, createdList.ID, nil)
	require.NoError(t, err)
	assert.Len(t, afterArchive.Data, 0)

	err = dbc.ArchiveRecipeListItem(ctx, createdItem.ID, createdList.ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
}
