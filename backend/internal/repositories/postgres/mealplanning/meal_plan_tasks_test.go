package mealplanning

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v10/pointer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMealPlanTaskForTest(t *testing.T, ctx context.Context, exampleMealPlanTask *types.MealPlanTask, dbc *repository) *types.MealPlanTask {
	t.Helper()

	// create
	dbInput := converters.ConvertMealPlanTaskToMealPlanTaskDatabaseCreationInput(exampleMealPlanTask)

	created, err := dbc.CreateMealPlanTask(ctx, dbInput)
	require.NoError(t, err)
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
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

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
	require.NoError(t, err)
	assert.NotEmpty(t, mealPlanTasks)
	assert.Len(t, mealPlanTasks.Data, len(createdMealPlanTasks))

	// create an ad-hoc thaw task: no backing recipe prep task, so
	// belongs_to_recipe_prep_task is persisted as NULL.
	thawTaskInput := &types.MealPlanTaskDatabaseCreationInput{
		ID:                  fakes.BuildFakeID(),
		MealPlanOptionID:    mealPlan.Events[0].Options[0].ID,
		CreationExplanation: "frozen ingredient might need to be thawed",
		RecipePrepTaskID:    "",
	}
	createdThawTask, err := dbc.CreateMealPlanTask(ctx, thawTaskInput)
	require.NoError(t, err)
	require.NotNil(t, createdThawTask)
	assert.Empty(t, createdThawTask.RecipePrepTask.ID)
	createdMealPlanTasks = append(createdMealPlanTasks, createdThawTask)

	// the thaw task reads back individually and in the list despite its NULL prep-task reference.
	fetchedThawTask, err := dbc.GetMealPlanTask(ctx, createdThawTask.ID)
	require.NoError(t, err)
	assert.Empty(t, fetchedThawTask.RecipePrepTask.ID)
	assert.Equal(t, "frozen ingredient might need to be thawed", fetchedThawTask.CreationExplanation)

	mealPlanTasks, err = dbc.GetMealPlanTasksForMealPlan(ctx, mealPlan.ID, nil)
	require.NoError(t, err)
	assert.Len(t, mealPlanTasks.Data, len(createdMealPlanTasks))

	// a batch that fails partway through commits nothing, and leaves the plan unmarked. The task
	// creator re-selects unmarked plans and regenerates every task, so anything left behind by a
	// partial failure would be created a second time on the next run.
	goodTaskInput := func() *types.MealPlanTaskDatabaseCreationInput {
		return &types.MealPlanTaskDatabaseCreationInput{
			ID:                  fakes.BuildFakeID(),
			MealPlanOptionID:    mealPlan.Events[0].Options[0].ID,
			CreationExplanation: t.Name(),
		}
	}

	doomedTask := goodTaskInput()
	// a nonexistent meal plan option trips the foreign key, so this input fails after the one
	// before it has already been inserted.
	doomedTask.MealPlanOptionID = fakes.BuildFakeID()

	batched, err := dbc.CreateMealPlanTasksForMealPlan(ctx, mealPlan.ID, []*types.MealPlanTaskDatabaseCreationInput{goodTaskInput(), doomedTask})
	require.Error(t, err)
	assert.Nil(t, batched)

	mealPlanTasks, err = dbc.GetMealPlanTasksForMealPlan(ctx, mealPlan.ID, nil)
	require.NoError(t, err)
	assert.Len(t, mealPlanTasks.Data, len(createdMealPlanTasks), "the task preceding the failure must have rolled back with it")

	unmarkedMealPlan, err := dbc.GetMealPlan(ctx, mealPlan.ID, account.ID)
	require.NoError(t, err)
	assert.False(t, unmarkedMealPlan.TasksCreated)

	// a batch that succeeds writes every task and the flag together.
	taskBatch := []*types.MealPlanTaskDatabaseCreationInput{goodTaskInput(), goodTaskInput()}
	batched, err = dbc.CreateMealPlanTasksForMealPlan(ctx, mealPlan.ID, taskBatch)
	require.NoError(t, err)
	assert.Len(t, batched, len(taskBatch))
	createdMealPlanTasks = append(createdMealPlanTasks, batched...)

	mealPlanTasks, err = dbc.GetMealPlanTasksForMealPlan(ctx, mealPlan.ID, nil)
	require.NoError(t, err)
	assert.Len(t, mealPlanTasks.Data, len(createdMealPlanTasks))

	markedMealPlan, err := dbc.GetMealPlan(ctx, mealPlan.ID, account.ID)
	require.NoError(t, err)
	assert.True(t, markedMealPlan.TasksCreated)

	// delete
	for _, mealPlanTask := range createdMealPlanTasks {
		require.NoError(t, dbc.ChangeMealPlanTaskStatus(ctx, &types.MealPlanTaskStatusChangeRequestInput{
			Status:            pointer.To(types.MealPlanTaskStatusFinished),
			StatusExplanation: t.Name(),
			AssignedToUser:    &user.ID,
			MealPlanTaskID:    mealPlanTask.ID,
		}))

		var exists bool
		exists, err = dbc.MealPlanTaskExists(ctx, mealPlanTask.ID, account.ID)
		require.NoError(t, err)
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
		require.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid meal plan task ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleMealPlanID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.MealPlanTaskExists(ctx, exampleMealPlanID, "")
		require.Error(t, err)
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
		require.Error(t, err)
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
		require.Error(t, err)
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
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateMealPlanTasksForMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("with empty meal plan ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateMealPlanTasksForMealPlan(ctx, "", nil)
		require.Error(t, err)
		assert.Nil(t, actual)
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
