package mealplanning

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildMealPlanEventForIntegrationTest(meal *types.Meal) *types.MealPlanEvent {
	exampleMealPlanEvent := fakes.BuildFakeMealPlanEvent()

	exampleMealPlanOption := buildMealPlanOptionForIntegrationTest(meal)
	exampleMealPlanOption.BelongsToMealPlanEvent = exampleMealPlanEvent.ID
	exampleMealPlanEvent.Options = []*types.MealPlanOption{
		exampleMealPlanOption,
	}

	return exampleMealPlanEvent
}

func createMealPlanEventForTest(t *testing.T, ctx context.Context, exampleMealPlanEvent *types.MealPlanEvent, dbc *repository) *types.MealPlanEvent {
	t.Helper()

	// create
	dbInput := converters.ConvertMealPlanEventToMealPlanEventDatabaseCreationInput(exampleMealPlanEvent)

	created, err := dbc.CreateMealPlanEvent(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleMealPlanEvent.CreatedAt = created.CreatedAt
	for i, option := range created.Options {
		exampleMealPlanEvent.Options[i].CreatedAt = option.CreatedAt
		require.Equal(t, exampleMealPlanEvent.Options[i].Meal.ID, option.Meal.ID)
		exampleMealPlanEvent.Options[i].Meal = option.Meal
	}
	assert.Equal(t, exampleMealPlanEvent, created)

	mealPlanEvent, err := dbc.GetMealPlanEvent(ctx, created.BelongsToMealPlan, created.ID)
	require.NoError(t, err)

	exampleMealPlanEvent.CreatedAt = mealPlanEvent.CreatedAt
	exampleMealPlanEvent.StartsAt = mealPlanEvent.StartsAt
	exampleMealPlanEvent.EndsAt = mealPlanEvent.EndsAt
	for i, option := range mealPlanEvent.Options {
		exampleMealPlanEvent.Options[i].CreatedAt = option.CreatedAt
		require.Equal(t, exampleMealPlanEvent.Options[i].Meal.ID, option.Meal.ID)
		exampleMealPlanEvent.Options[i].Meal = option.Meal
		exampleMealPlanEvent.Options[i].Chosen = option.Chosen
	}
	require.Equal(t, exampleMealPlanEvent.CreatedAt, mealPlanEvent.CreatedAt)
	require.Equal(t, exampleMealPlanEvent.LastUpdatedAt, mealPlanEvent.LastUpdatedAt)
	require.Equal(t, exampleMealPlanEvent.ID, mealPlanEvent.ID)

	assert.Equal(t, exampleMealPlanEvent, mealPlanEvent)

	return mealPlanEvent
}

func TestQuerier_Integration_MealPlanEvents(t *testing.T) {
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
	buildMealForIntegrationTest(user.ID, recipe)
	meal := createMealForTest(t, ctx, nil, dbc)

	exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
	exampleMealPlan.BelongsToAccount = account.ID
	mealPlan := createMealPlanForTest(t, ctx, exampleMealPlan, dbc)

	newMeal := createMealForTest(t, ctx, nil, dbc)
	newEvent := buildMealPlanEventForIntegrationTest(newMeal)
	newEvent.BelongsToMealPlan = mealPlan.ID

	exampleMealPlanEvent := fakes.BuildFakeMealPlanEvent()
	exampleMealPlanEvent.BelongsToMealPlan = mealPlan.ID
	exampleMealPlanEvent.Options = newEvent.Options
	for i := range exampleMealPlanEvent.Options {
		exampleMealPlanEvent.Options[i].BelongsToMealPlanEvent = exampleMealPlanEvent.ID
	}

	// create
	createdMealPlanEvents := []*types.MealPlanEvent{
		mealPlan.Events[0],
	}
	createdMealPlanEvents = append(createdMealPlanEvents, createMealPlanEventForTest(t, ctx, exampleMealPlanEvent, dbc))

	// fetch as list
	mealPlanEvents, err := dbc.GetMealPlanEvents(ctx, mealPlan.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, mealPlanEvents)
	assert.Equal(t, len(createdMealPlanEvents), len(mealPlanEvents.Data))

	assert.NoError(t, dbc.UpdateMealPlanEvent(ctx, createdMealPlanEvents[0]))

	_, err = dbc.MealPlanEventIsEligibleForVoting(ctx, mealPlan.ID, createdMealPlanEvents[0].ID)
	assert.NoError(t, err)

	// delete
	for _, mealPlanEvent := range createdMealPlanEvents {
		assert.NoError(t, dbc.ArchiveMealPlanEvent(ctx, mealPlan.ID, mealPlanEvent.ID))

		var exists bool
		exists, err = dbc.MealPlanEventExists(ctx, mealPlanEvent.ID, account.ID)
		assert.NoError(t, err)
		assert.False(t, exists)
	}
}

func TestQuerier_MealPlanEventExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleMealPlanEventID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.MealPlanEventExists(ctx, "", exampleMealPlanEventID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid meal plan event ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleMealPlanID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.MealPlanEventExists(ctx, exampleMealPlanID, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan ID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanEventID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanEvent(ctx, "", exampleMealPlanEventID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid meal plan event ID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlan := fakes.BuildFakeMealPlan()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanEvent(ctx, exampleMealPlan.ID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_getMealPlanEventsForMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.getMealPlanEventsForMealPlan(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetMealPlanEvents(T *testing.T) {
	T.Parallel()

	T.Run("with invalid ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanEvents(ctx, "", nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateMealPlanEvent(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.UpdateMealPlanEvent(ctx, nil)
		assert.Error(t, err)
	})
}

func TestQuerier_ArchiveMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan event ID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlan := fakes.BuildFakeMealPlan()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveMealPlanEvent(ctx, "", exampleMealPlan.ID))
	})

	T.Run("with invalid meal plan ID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanEventID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveMealPlanEvent(ctx, exampleMealPlanEventID, ""))
	})
}

func TestQuerier_Integration_MealPlanEvents_CursorBasedPagination(t *testing.T) {
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
	recipe := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)
	meal := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe), dbc)

	// Create an empty meal plan without events - bypass createMealPlanForTest since it has assertions
	// that don't work with empty event lists
	mealPlanInput := converters.ConvertMealPlanToMealPlanDatabaseCreationInput(fakes.BuildFakeMealPlan())
	mealPlanInput.CreatedByUser = user.ID
	mealPlanInput.BelongsToAccount = account.ID
	mealPlanInput.Events = []*types.MealPlanEventDatabaseCreationInput{}

	createdMealPlan, err := dbc.CreateMealPlan(ctx, mealPlanInput)
	require.NoError(t, err)
	require.NotNil(t, createdMealPlan)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.MealPlanEvent]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "meal plan event",
		CreateItem: func(ctx context.Context, i int) *types.MealPlanEvent {
			mealPlanEvent := buildMealPlanEventForIntegrationTest(meal)
			mealPlanEvent.BelongsToMealPlan = createdMealPlan.ID
			return createMealPlanEventForTest(t, ctx, mealPlanEvent, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanEvent], error) {
			return dbc.GetMealPlanEvents(ctx, createdMealPlan.ID, filter)
		},
		GetID: func(mealPlanEvent *types.MealPlanEvent) string {
			return mealPlanEvent.ID
		},
		CleanupItem: func(ctx context.Context, mealPlanEvent *types.MealPlanEvent) error {
			return dbc.ArchiveMealPlanEvent(ctx, createdMealPlan.ID, mealPlanEvent.ID)
		},
	})
}
