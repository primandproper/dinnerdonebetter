package mealplanning

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v6/identifiers"

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
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, _, container := buildDatabaseClientForTest(t)
	defer func() {
		assert.NoError(t, container.Terminate(ctx))
	}()

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	listInput := buildRecipeListForTest(user.ID)
	createdList, err := dbc.CreateRecipeList(ctx, listInput)
	require.NoError(t, err)
	require.NotNil(t, createdList)

	res, err := dbc.GetRecipeLists(ctx, nil)
	require.NoError(t, err)
	require.Len(t, res.Data, 1)
	require.Len(t, res.Data[0].Items, 0)

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
	assert.Len(t, resAfterArchive.Data, 0)
}
