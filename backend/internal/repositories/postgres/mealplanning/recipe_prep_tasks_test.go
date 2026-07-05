package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v3/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRecipePrepTaskForTest(t *testing.T, ctx context.Context, exampleRecipePrepTask *types.RecipePrepTask, dbc *repository) *types.RecipePrepTask {
	t.Helper()

	// create
	if exampleRecipePrepTask == nil {
		exampleRecipePrepTask = fakes.BuildFakeRecipePrepTask()
	}
	dbInput := converters.ConvertRecipePrepTaskToRecipePrepTaskDatabaseCreationInput(exampleRecipePrepTask)

	created, err := dbc.CreateRecipePrepTask(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleRecipePrepTask.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleRecipePrepTask, created)

	recipePrepTask, err := dbc.GetRecipePrepTask(ctx, created.BelongsToRecipe, created.ID)
	require.NoError(t, err)
	require.NotNil(t, recipePrepTask)
	exampleRecipePrepTask.CreatedAt = recipePrepTask.CreatedAt

	assert.Equal(t, exampleRecipePrepTask, recipePrepTask)

	return created
}

func TestQuerier_Integration_RecipePrepTasks(t *testing.T) {
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

	exampleRecipePrepTask := fakes.BuildFakeRecipePrepTask()
	exampleRecipePrepTask.BelongsToRecipe = createdRecipe.ID
	exampleRecipePrepTask.TaskSteps = []*types.RecipePrepTaskStep{
		{
			ID:                      identifiers.New(),
			BelongsToRecipeStep:     exampleRecipe.Steps[0].ID,
			BelongsToRecipePrepTask: exampleRecipePrepTask.ID,
			SatisfiesRecipeStep:     true,
		},
	}
	createdRecipePrepTasks := []*types.RecipePrepTask{
		exampleRecipe.PrepTasks[0],
	}

	// create
	createdRecipePrepTasks = append(createdRecipePrepTasks, createRecipePrepTaskForTest(t, ctx, exampleRecipePrepTask, dbc))

	// fetch as list
	recipePrepTasks, err := dbc.GetRecipePrepTasksForRecipe(ctx, exampleRecipe.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, recipePrepTasks)
	assert.Equal(t, len(createdRecipePrepTasks), len(recipePrepTasks))

	// delete
	for _, recipePrepTask := range createdRecipePrepTasks {
		assert.NoError(t, dbc.ArchiveRecipePrepTask(ctx, createdRecipe.ID, recipePrepTask.ID))

		var exists bool
		exists, err = dbc.RecipePrepTaskExists(ctx, createdRecipe.ID, recipePrepTask.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.RecipePrepTask
		y, err = dbc.GetRecipePrepTask(ctx, createdRecipe.ID, recipePrepTask.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_RecipePrepTaskExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRecipePrepTask := fakes.BuildFakeRecipePrepTask()

		c := buildInertClientForTest(t)

		actual, err := c.RecipePrepTaskExists(ctx, "", exampleRecipePrepTask.ID)
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRecipePrepTaskID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.GetRecipePrepTask(ctx, "", exampleRecipePrepTaskID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateRecipePrepTask(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetRecipePrepTasksForRecipe(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipePrepTasksForRecipe(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateRecipePrepTask(ctx, nil))
	})
}

func TestQuerier_ArchiveRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("with missing recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleRecipePrepTask := fakes.BuildFakeRecipePrepTask()

		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipePrepTask(ctx, "", exampleRecipePrepTask.ID))
	})
}
