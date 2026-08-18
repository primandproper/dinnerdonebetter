package mealplanning

import (
	"context"
	"testing"
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"resenje.org/schulze"
)

func buildMealPlanOptionForIntegrationTest(meal *types.Meal) *types.MealPlanOption {
	exampleMealPlanOption := fakes.BuildFakeMealPlanOption()
	exampleMealPlanOption.Meal = *meal
	exampleMealPlanOption.MealScale = 1
	exampleMealPlanOption.AssignedCook = nil
	exampleMealPlanOption.AssignedDishwasher = nil
	exampleMealPlanOption.Votes = []*types.MealPlanOptionVote{}

	return exampleMealPlanOption
}

func createMealPlanOptionForTest(t *testing.T, ctx context.Context, mealPlanID string, exampleMealPlanOption *types.MealPlanOption, dbc *repository) *types.MealPlanOption {
	t.Helper()

	// create
	dbInput := converters.ConvertMealPlanOptionToMealPlanOptionDatabaseCreationInput(exampleMealPlanOption)

	created, err := dbc.CreateMealPlanOption(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)

	exampleMealPlanOption.CreatedAt = created.CreatedAt
	require.Equal(t, exampleMealPlanOption.Meal.ID, created.Meal.ID)
	exampleMealPlanOption.Meal = created.Meal
	require.Len(t, created.Votes, len(exampleMealPlanOption.Votes))
	exampleMealPlanOption.Votes = created.Votes
	assert.Equal(t, exampleMealPlanOption, created)

	mealPlanOption, err := dbc.GetMealPlanOption(ctx, mealPlanID, created.BelongsToMealPlanEvent, created.ID)
	require.NoError(t, err)

	exampleMealPlanOption.CreatedAt = mealPlanOption.CreatedAt
	require.Equal(t, exampleMealPlanOption.Meal.ID, mealPlanOption.Meal.ID)
	exampleMealPlanOption.Meal = mealPlanOption.Meal
	require.Len(t, mealPlanOption.Votes, len(exampleMealPlanOption.Votes))
	exampleMealPlanOption.Votes = mealPlanOption.Votes
	require.Equal(t, exampleMealPlanOption.CreatedAt, mealPlanOption.CreatedAt)
	require.Equal(t, exampleMealPlanOption.LastUpdatedAt, mealPlanOption.LastUpdatedAt)
	require.Equal(t, exampleMealPlanOption.ID, mealPlanOption.ID)

	assert.Equal(t, exampleMealPlanOption, mealPlanOption)

	return mealPlanOption
}

func TestQuerier_Integration_MealPlanOptions(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	recipe := createRecipeForTest(t, ctx, nil, dbc, true)
	buildMealForIntegrationTest(user.ID, recipe)
	meal := createMealForTest(t, ctx, nil, dbc)

	exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
	exampleMealPlan.BelongsToAccount = account.ID
	mealPlan := createMealPlanForTest(t, ctx, exampleMealPlan, dbc)

	newMeal := createMealForTest(t, ctx, nil, dbc)
	exampleMealPlanOption := buildMealPlanOptionForIntegrationTest(newMeal)
	exampleMealPlanOption.BelongsToMealPlanEvent = mealPlan.Events[0].ID

	// create
	createdMealPlanOptions := []*types.MealPlanOption{
		mealPlan.Events[0].Options[0],
	}
	createdMealPlanOptions = append(createdMealPlanOptions, createMealPlanOptionForTest(t, ctx, mealPlan.ID, exampleMealPlanOption, dbc))

	// fetch as list
	mealPlanOptions, err := dbc.GetMealPlanOptions(ctx, mealPlan.ID, exampleMealPlanOption.BelongsToMealPlanEvent, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, mealPlanOptions)
	assert.Len(t, mealPlanOptions.Data, len(createdMealPlanOptions))

	require.NoError(t, dbc.UpdateMealPlanOption(ctx, createdMealPlanOptions[0]))

	byID, err := dbc.getMealPlanOptionByID(ctx, createdMealPlanOptions[0].ID)
	require.NoError(t, err)
	assert.Equal(t, createdMealPlanOptions[0].ID, byID.ID)

	_, err = dbc.FinalizeMealPlanOption(ctx, mealPlan.ID, mealPlan.Events[0].ID, createdMealPlanOptions[0].ID, account.ID)
	require.NoError(t, err)

	// delete
	for _, mealPlanOption := range createdMealPlanOptions {
		require.NoError(t, dbc.ArchiveMealPlanOption(ctx, mealPlan.ID, exampleMealPlanOption.BelongsToMealPlanEvent, mealPlanOption.ID))

		var exists bool
		exists, err = dbc.MealPlanOptionExists(ctx, mealPlan.ID, exampleMealPlanOption.BelongsToMealPlanEvent, mealPlanOption.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	}
}

func TestQuerier_Integration_MealExistsAsOptionInEvent(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)
	defer func() {
	}()

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	meal := createMealForTest(t, ctx, nil, dbc)

	exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
	exampleMealPlan.BelongsToAccount = account.ID
	mealPlan := createMealPlanForTest(t, ctx, exampleMealPlan, dbc)

	eventID := mealPlan.Events[0].ID
	existingOptionMealID := mealPlan.Events[0].Options[0].Meal.ID

	exists, err := dbc.MealExistsAsOptionInEvent(ctx, eventID, existingOptionMealID)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = dbc.MealExistsAsOptionInEvent(ctx, eventID, "nonexistent-meal-id")
	require.NoError(t, err)
	assert.False(t, exists)

	newMeal := createMealForTest(t, ctx, nil, dbc)
	exists, err = dbc.MealExistsAsOptionInEvent(ctx, eventID, newMeal.ID)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestQuerier_MealPlanOptionExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleMealPlanOption := fakes.BuildFakeMealPlanOption()

		c := buildInertClientForTest(t)

		actual, err := c.MealPlanOptionExists(ctx, "", "", exampleMealPlanOption.ID)
		require.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid meal plan option MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.MealPlanOptionExists(ctx, exampleMealPlanID, exampleMealPlanEventID, "")
		require.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanOption := fakes.BuildFakeMealPlanOption()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanOption(ctx, "", "", exampleMealPlanOption.ID)
		require.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid meal plan option MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanOption(ctx, exampleMealPlanID, exampleMealPlanEventID, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_getMealPlanOptionByID(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan option MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.getMealPlanOptionByID(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetMealPlanOptions(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanOptions(ctx, "", "", filter)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateMealPlanOption(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateMealPlanOption(ctx, nil))
	})
}

func TestQuerier_ArchiveMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanOption := fakes.BuildFakeMealPlanOption()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveMealPlanOption(ctx, "", "", exampleMealPlanOption.ID))
	})

	T.Run("with invalid meal plan option MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveMealPlanOption(ctx, exampleMealPlanID, exampleMealPlanEventID, ""))
	})
}

func Test_determineWinner(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c := buildInertClientForTest(t)

		expected := "blah blah blah"
		exampleWinners := []schulze.Result[string]{
			{
				Choice: t.Name(),
				Wins:   1,
			},
			{
				Choice: "",
				Wins:   2,
			},
			{
				Choice: expected,
				Wins:   3,
			},
		}

		actual := c.determineWinner(exampleWinners)

		assert.Equal(t, expected, actual)
	})

	T.Run("with tie", func(t *testing.T) {
		t.Parallel()

		c := buildInertClientForTest(t)

		expectedA := "blah blah blah"
		expectedB := "beeble beeble"
		exampleWinners := []schulze.Result[string]{
			{
				Choice: expectedA,
				Wins:   3,
			},
			{
				Choice: "",
				Wins:   1,
			},
			{
				Choice: expectedB,
				Wins:   3,
			},
		}

		actual := c.determineWinner(exampleWinners)

		assert.True(t, expectedA == actual || expectedB == actual)
	})
}

func Test_decideOptionWinner(T *testing.T) {
	T.Parallel()

	optionA := "eggs benedict"
	optionB := "scrambled eggs"
	optionC := "buttered toast"
	userID1 := fake.BuildFakeID()
	userID2 := fake.BuildFakeID()
	userID3 := fake.BuildFakeID()
	userID4 := fake.BuildFakeID()

	T.Run("with clear winner", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		expected := optionA
		exampleOptions := []*types.MealPlanOption{
			{
				ID: optionA,
				Votes: []*types.MealPlanOptionVote{
					{
						BelongsToMealPlanOption: optionA,
						Rank:                    0,
						ByUser:                  userID1,
					},
					{
						BelongsToMealPlanOption: optionA,
						Rank:                    0,
						ByUser:                  userID2,
					},
					{
						BelongsToMealPlanOption: optionA,
						Rank:                    1,
						ByUser:                  userID3,
					},
					{
						BelongsToMealPlanOption: optionA,
						Rank:                    2,
						ByUser:                  userID4,
					},
				},
			},
			{
				ID: optionB,
				Votes: []*types.MealPlanOptionVote{
					{
						BelongsToMealPlanOption: optionB,
						Rank:                    0,
						ByUser:                  userID3,
					},
					{
						BelongsToMealPlanOption: optionB,
						Rank:                    1,
						ByUser:                  userID2,
					},
					{
						BelongsToMealPlanOption: optionB,
						Rank:                    1,
						ByUser:                  userID4,
					},
					{
						BelongsToMealPlanOption: optionB,
						Rank:                    2,
						ByUser:                  userID1,
					},
				},
			},
			{
				ID: optionC,
				Votes: []*types.MealPlanOptionVote{
					{
						BelongsToMealPlanOption: optionC,
						Rank:                    0,
						ByUser:                  userID4,
					},

					{
						BelongsToMealPlanOption: optionC,
						Rank:                    1,
						ByUser:                  userID1,
					},
					{
						BelongsToMealPlanOption: optionC,
						Rank:                    2,
						ByUser:                  userID2,
					},
					{
						BelongsToMealPlanOption: optionC,
						Rank:                    2,
						ByUser:                  userID3,
					},
				},
			},
		}

		actual, tiebroken, chosen := c.decideOptionWinner(ctx, exampleOptions)
		assert.Equal(t, expected, actual)
		assert.False(t, tiebroken)
		assert.True(t, chosen)
	})

	T.Run("with tie", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		exampleOptions := []*types.MealPlanOption{
			{
				ID: optionA,
				Votes: []*types.MealPlanOptionVote{
					{
						BelongsToMealPlanOption: optionA,
						Rank:                    0,
						ByUser:                  userID1,
					},
				},
			},
			{
				ID: optionB,
				Votes: []*types.MealPlanOptionVote{
					{
						BelongsToMealPlanOption: optionB,
						Rank:                    0,
						ByUser:                  userID2,
					},
				},
			},
		}

		actual, tiebroken, chosen := c.decideOptionWinner(ctx, exampleOptions)
		assert.NotEmpty(t, actual)
		assert.True(t, tiebroken)
		assert.True(t, chosen)
	})

	T.Run("without enough votes", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		exampleOptions := []*types.MealPlanOption{
			{
				ID:    optionA,
				Votes: nil,
			},
		}

		actual, tiebroken, chosen := c.decideOptionWinner(ctx, exampleOptions)
		assert.Empty(t, actual)
		assert.False(t, tiebroken)
		assert.False(t, chosen)
	})
}

func TestQuerier_Integration_MealPlanOptions_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	// Create a meal plan with one event but no options in it
	// We create this directly without using createMealPlanForTest to avoid strict nil vs empty slice comparisons
	mealPlan := fakes.BuildFakeMealPlan()
	mealPlan.CreatedByUser = user.ID
	mealPlan.BelongsToAccount = account.ID

	// Create event without any options
	now := fake.BuildFakeTime()
	inTenMinutes := now.Add(10 * time.Minute)
	inOneWeek := now.Add(7 * 24 * time.Hour)
	mealPlanEvent := &types.MealPlanEvent{
		ID:                fake.BuildFakeID(),
		Notes:             fake.BuildFakeID(),
		StartsAt:          inTenMinutes,
		EndsAt:            inOneWeek,
		MealName:          types.BreakfastMealName,
		CreatedAt:         now,
		BelongsToMealPlan: mealPlan.ID,
		Options:           nil,
	}
	mealPlan.Events = []*types.MealPlanEvent{mealPlanEvent}

	// Create the meal plan directly
	dbInput := converters.ConvertMealPlanToMealPlanDatabaseCreationInput(mealPlan)
	createdMealPlan, err := dbc.CreateMealPlan(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, createdMealPlan)
	require.NotEmpty(t, createdMealPlan.Events)

	mealPlanEventID := createdMealPlan.Events[0].ID

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.MealPlanOption]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "meal plan option",
		CreateItem: func(ctx context.Context, i int) *types.MealPlanOption {
			// Create a distinct meal for each option (duplicate prevention forbids same meal twice per event)
			recipe := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)
			meal := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe), dbc)
			mealPlanOption := buildMealPlanOptionForIntegrationTest(meal)
			mealPlanOption.BelongsToMealPlanEvent = mealPlanEventID
			return createMealPlanOptionForTest(t, ctx, createdMealPlan.ID, mealPlanOption, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanOption], error) {
			return dbc.GetMealPlanOptions(ctx, createdMealPlan.ID, mealPlanEventID, filter)
		},
		GetID: func(mealPlanOption *types.MealPlanOption) string {
			return mealPlanOption.ID
		},
		CleanupItem: func(ctx context.Context, mealPlanOption *types.MealPlanOption) error {
			return dbc.ArchiveMealPlanOption(ctx, createdMealPlan.ID, mealPlanEventID, mealPlanOption.ID)
		},
	})
}
