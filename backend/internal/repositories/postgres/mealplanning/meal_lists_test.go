package mealplanning

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v9/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildMealListForTest(userID string) *mealplanning.MealListDatabaseCreationInput {
	listID := identifiers.New()
	return &mealplanning.MealListDatabaseCreationInput{
		ID:            listID,
		Name:          "example meal list",
		Description:   "desc",
		BelongsToUser: userID,
	}
}

func TestIntegration_MealLists(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)
	defer func() {
	}()

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	listInput := buildMealListForTest(user.ID)
	createdList, err := dbc.CreateMealList(ctx, listInput)
	require.NoError(t, err)
	require.NotNil(t, createdList)

	res, err := dbc.GetMealLists(ctx, user.ID, nil)
	require.NoError(t, err)
	require.Len(t, res.Data, 1)
	require.Len(t, res.Data[0].Items, 0)

	updated := &mealplanning.MealList{
		ID:            createdList.ID,
		Name:          "updated meal list",
		Description:   "updated desc",
		BelongsToUser: user.ID,
	}
	require.NoError(t, dbc.UpdateMealList(ctx, updated))

	require.NoError(t, dbc.ArchiveMealList(ctx, createdList.ID, user.ID))

	resAfterArchive, err := dbc.GetMealLists(ctx, user.ID, nil)
	require.NoError(t, err)
	assert.Len(t, resAfterArchive.Data, 0)
}
