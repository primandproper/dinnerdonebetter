package mealplanning

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v4/pointer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMealPlanTaskForTest(t *testing.T, ctx context.Context, exampleMealPlanTask *types.MealPlanTask, dbc *repository) *types.MealPlanTask {
	t.Helper()

	// create
	dbInput := converters.ConvertMealPlanTaskToMealPlanTaskDatabaseCreationInput(exampleMealPlanTask)

	created, err := dbc.CreateMealPlanTask(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleMealPlanTask.CreatedAt = created.CreatedAt
	require.Equal(t, exampleMealPlanTask.RecipePrepTask.ID, created.RecipePrepTask.ID)
	exampleMealPlanTask.RecipePrepTask = created.RecipePrepTask
	require.Equal(t, exampleMealPlanTask.MealPlanOption.ID, created.MealPlanOption.ID)
	exampleMealPlanTask.MealPlanOption = created.MealPlanOption
	assert.Equal(t, exampleMealPlanTask, created)

	mealPlanTask, err := dbc.GetMealPlanTask(ctx, created.ID)
	require.NoError(t, err)

	exampleMealPlanTask.CreatedAt = mealPlanTask.CreatedAt
	exampleMealPlanTask.RecipePrepTask = mealPlanTask.RecipePrepTask
	exampleMealPlanTask.MealPlanOption = mealPlanTask.MealPlanOption
	require.Equal(t, exampleMealPlanTask.CreatedAt, mealPlanTask.CreatedAt)
	require.Equal(t, exampleMealPlanTask.LastUpdatedAt, mealPlanTask.LastUpdatedAt)
	require.Equal(t, exampleMealPlanTask.ID, mealPlanTask.ID)
	require.Equal(t, exampleMealPlanTask.Status, mealPlanTask.Status)

	assert.Equal(t, exampleMealPlanTask, mealPlanTask)

	return mealPlanTask
}

func TestQuerier_Integration_MealPlanTasks(t *testing.T) {
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
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	recipe := createRecipeForTest(t, ctx, nil, dbc, true)
	meal := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe), dbc)

	exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
	exampleMealPlan.BelongsToAccount = account.ID
	mealPlan := createMealPlanForTest(t, ctx, exampleMealPlan, dbc)

	exampleMealPlanTask := fakes.BuildFakeMealPlanTask()
	exampleMealPlanTask.RecipePrepTask = *recipe.PrepTasks[0]
	exampleMealPlanTask.MealPlanOption = *mealPlan.Events[0].Options[0]

	// create
	createdMealPlanTasks := []*types.MealPlanTask{}
	createdMealPlanTasks = append(createdMealPlanTasks, createMealPlanTaskForTest(t, ctx, exampleMealPlanTask, dbc))

	// fetch as list
	mealPlanTasks, err := dbc.GetMealPlanTasksForMealPlan(ctx, mealPlan.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, mealPlanTasks)
	assert.Equal(t, len(createdMealPlanTasks), len(mealPlanTasks.Data))

	// delete
	for _, mealPlanTask := range createdMealPlanTasks {
		assert.NoError(t, dbc.ChangeMealPlanTaskStatus(ctx, &types.MealPlanTaskStatusChangeRequestInput{
			Status:            pointer.To(types.MealPlanTaskStatusFinished),
			StatusExplanation: t.Name(),
			AssignedToUser:    &user.ID,
			MealPlanTaskID:    mealPlanTask.ID,
		}))

		var exists bool
		exists, err = dbc.MealPlanTaskExists(ctx, mealPlanTask.ID, account.ID)
		assert.NoError(t, err)
		assert.False(t, exists)
	}
}

func TestQuerier_MealPlanTaskExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleMealPlanTaskID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.MealPlanTaskExists(ctx, "", exampleMealPlanTaskID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid meal plan task ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleMealPlanID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.MealPlanTaskExists(ctx, exampleMealPlanID, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetMealPlanTask(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan task MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanTask(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateMealPlanTask(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateMealPlanTask(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetMealPlanTasksForMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("with missing meal plan MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanTasksForMealPlan(ctx, "", nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_MarkMealPlanAsHavingTasksCreated(T *testing.T) {
	T.Parallel()

	T.Run("with empty meal plan MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.MarkMealPlanAsHavingTasksCreated(ctx, ""))
	})
}

func TestQuerier_MarkMealPlanAsHavingGroceryListInitialized(T *testing.T) {
	T.Parallel()

	T.Run("with empty meal plan MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.MarkMealPlanAsHavingGroceryListInitialized(ctx, ""))
	})
}

func TestQuerier_ChangeMealPlanTaskStatus(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ChangeMealPlanTaskStatus(ctx, nil))
	})
}
