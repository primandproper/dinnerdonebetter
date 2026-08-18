package mealplanning

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v11/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildRecipeListForTest(userID string) *mealplanning.RecipeListDatabaseCreationInput {
	listID := identifiers.New()
	return &mealplanning.RecipeListDatabaseCreationInput{
		ID:            listID,
		Name:          "example recipe list",
		Description:   "desc",
		BelongsToUser: userID,
	}
}

func TestIntegration_RecipeLists(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)
	defer func() {
	}()

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	listInput := buildRecipeListForTest(user.ID)
	createdList, err := dbc.CreateRecipeList(ctx, listInput)
	require.NoError(t, err)
	require.NotNil(t, createdList)

	res, err := dbc.GetRecipeLists(ctx, nil)
	require.NoError(t, err)
	require.Len(t, res.Data, 1)
	require.Empty(t, res.Data[0].Items)

	updated := &mealplanning.RecipeList{
		ID:            createdList.ID,
		Name:          "updated recipe list",
		Description:   "updated desc",
		BelongsToUser: user.ID,
	}
	require.NoError(t, dbc.UpdateRecipeList(ctx, updated))

	require.NoError(t, dbc.ArchiveRecipeList(ctx, createdList.ID, user.ID))

	resAfterArchive, err := dbc.GetRecipeLists(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, resAfterArchive.Data)
}
