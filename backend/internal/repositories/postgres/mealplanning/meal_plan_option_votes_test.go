package mealplanning

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authorization"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v5/filtering"
	"github.com/primandproper/platform-go/v5/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMealPlanOptionVoteForTest(t *testing.T, ctx context.Context, mealPlanID, mealPlanEventID string, exampleMealPlanOptionVote *types.MealPlanOptionVote, dbc *repository) *types.MealPlanOptionVote {
	t.Helper()

	// create
	dbInput := converters.ConvertMealPlanOptionVoteToMealPlanOptionVotesDatabaseCreationInput(exampleMealPlanOptionVote)

	rawCreated, err := dbc.CreateMealPlanOptionVote(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, rawCreated)
	assert.Len(t, rawCreated, 1)
	created := rawCreated[0]

	exampleMealPlanOptionVote.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleMealPlanOptionVote, created)

	mealPlanOptionVote, err := dbc.GetMealPlanOptionVote(ctx, mealPlanID, mealPlanEventID, created.BelongsToMealPlanOption, created.ID)
	require.NoError(t, err)

	exampleMealPlanOptionVote.CreatedAt = mealPlanOptionVote.CreatedAt
	require.Equal(t, exampleMealPlanOptionVote.CreatedAt, mealPlanOptionVote.CreatedAt)
	require.Equal(t, exampleMealPlanOptionVote.LastUpdatedAt, mealPlanOptionVote.LastUpdatedAt)
	require.Equal(t, exampleMealPlanOptionVote.ID, mealPlanOptionVote.ID)

	assert.Equal(t, exampleMealPlanOptionVote, mealPlanOptionVote)

	return mealPlanOptionVote
}

func TestQuerier_Integration_MealPlanOptionVotes(t *testing.T) {
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
	mealPlanEvent := mealPlan.Events[0]
	mealPlanOption := mealPlanEvent.Options[0]

	exampleMealPlanOptionVote := fakes.BuildFakeMealPlanOptionVote()
	exampleMealPlanOptionVote.ByUser = user.ID
	exampleMealPlanOptionVote.BelongsToMealPlanOption = mealPlanOption.ID

	// create
	createdMealPlanOptionVotes := []*types.MealPlanOptionVote{}
	createdMealPlanOptionVotes = append(createdMealPlanOptionVotes, createMealPlanOptionVoteForTest(t, ctx, mealPlan.ID, mealPlanEvent.ID, exampleMealPlanOptionVote, dbc))

	// fetch as list
	mealPlanOptionVotes, err := dbc.GetMealPlanOptionVotes(ctx, mealPlan.ID, mealPlanEvent.ID, mealPlanOption.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, mealPlanOptionVotes)
	assert.Equal(t, len(createdMealPlanOptionVotes), len(mealPlanOptionVotes.Data))

	assert.NoError(t, dbc.UpdateMealPlanOptionVote(ctx, createdMealPlanOptionVotes[0]))

	// delete
	for _, mealPlanOptionVote := range createdMealPlanOptionVotes {
		assert.NoError(t, dbc.ArchiveMealPlanOptionVote(ctx, mealPlan.ID, mealPlanEvent.ID, mealPlanOption.ID, mealPlanOptionVote.ID))

		var exists bool
		exists, err = dbc.MealPlanOptionVoteExists(ctx, mealPlan.ID, mealPlanEvent.ID, mealPlanOption.ID, mealPlanOptionVote.ID)
		assert.NoError(t, err)
		assert.False(t, exists)
	}
}

func TestQuerier_MealPlanOptionVoteExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleMealPlanEventID := fakes.BuildFakeID()
		exampleMealPlanOptionID := fakes.BuildFakeID()
		exampleMealPlanOptionVote := fakes.BuildFakeMealPlanOptionVote()

		c := buildInertClientForTest(t)

		actual, err := c.MealPlanOptionVoteExists(ctx, "", exampleMealPlanEventID, exampleMealPlanOptionID, exampleMealPlanOptionVote.ID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid meal plan option MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		exampleMealPlanOptionVote := fakes.BuildFakeMealPlanOptionVote()

		c := buildInertClientForTest(t)

		actual, err := c.MealPlanOptionVoteExists(ctx, exampleMealPlanID, exampleMealPlanEventID, "", exampleMealPlanOptionVote.ID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid meal plan option vote MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		exampleMealPlanOptionID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.MealPlanOptionVoteExists(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetMealPlanOptionVote(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanEventID := fakes.BuildFakeID()
		exampleMealPlanOptionID := fakes.BuildFakeID()
		exampleMealPlanOptionVote := fakes.BuildFakeMealPlanOptionVote()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanOptionVote(ctx, "", exampleMealPlanEventID, exampleMealPlanOptionID, exampleMealPlanOptionVote.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid meal plan option MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		exampleMealPlanOptionVote := fakes.BuildFakeMealPlanOptionVote()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanOptionVote(ctx, exampleMealPlanID, exampleMealPlanEventID, "", exampleMealPlanOptionVote.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid meal plan option vote MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		exampleMealPlanOptionID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanOptionVote(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetMealPlanOptionVotes(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()
		exampleMealPlanEventID := fakes.BuildFakeID()
		exampleMealPlanOptionID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanOptionVotes(ctx, "", exampleMealPlanEventID, exampleMealPlanOptionID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid meal plan option MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()
		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanOptionVotes(ctx, exampleMealPlanID, exampleMealPlanEventID, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateMealPlanOptionVote(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateMealPlanOptionVote(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateMealPlanOptionVote(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateMealPlanOptionVote(ctx, nil))
	})
}

func TestQuerier_ArchiveMealPlanOptionVote(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan option MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		exampleMealPlanOptionVote := fakes.BuildFakeMealPlanOptionVote()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveMealPlanOptionVote(ctx, exampleMealPlanID, exampleMealPlanEventID, "", exampleMealPlanOptionVote.ID))
	})

	T.Run("with invalid meal plan option vote MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		exampleMealPlanOptionID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveMealPlanOptionVote(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, ""))
	})
}

func TestQuerier_Integration_MealPlanOptionVotes_CursorBasedPagination(t *testing.T) {
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

	// Follow the same setup pattern as TestQuerier_Integration_MealPlanOptionVotes
	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	recipe := createRecipeForTest(t, ctx, nil, dbc, true)
	buildMealForIntegrationTest(user.ID, recipe)
	meal := createMealForTest(t, ctx, nil, dbc)

	// Add extra non-voting users to the account to prevent the meal plan from being finalized
	// when all votes are received (we'll create 9 votes but have 10+ users in the account)
	addUserToAccountHelper := func(userID string) {
		_, execErr := dbc.writeDB.ExecContext(ctx,
			`INSERT INTO account_user_memberships (id, belongs_to_user, belongs_to_account, default_account) VALUES ($1, $2, $3, $4)`,
			identifiers.New(), userID, account.ID, false)
		require.NoError(t, execErr)
		_, execErr = dbc.writeDB.ExecContext(ctx,
			`INSERT INTO user_role_assignments (id, user_id, role_id, account_id) VALUES ($1, $2, $3, $4)`,
			identifiers.New(), userID, authorization.AccountMemberRoleID, account.ID)
		require.NoError(t, execErr)
	}
	// Add one extra non-voting user
	nonVotingUser := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	addUserToAccountHelper(nonVotingUser.ID)

	// We need to create a meal plan with 2 options so it stays in "awaiting_votes" status
	// Create a second meal for the second option
	recipe2 := createRecipeForTest(t, ctx, nil, dbc, true)
	meal2 := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe2), dbc)

	// Build meal plan with event containing 2 options - this forces "awaiting_votes" status
	exampleMealPlan := fakes.BuildFakeMealPlan()
	exampleMealPlan.CreatedByUser = user.ID
	exampleMealPlan.BelongsToAccount = account.ID

	exampleEvent := buildMealPlanEventForIntegrationTest(meal)
	// Add a second option to the event
	exampleOption2 := buildMealPlanOptionForIntegrationTest(meal2)
	exampleOption2.BelongsToMealPlanEvent = exampleEvent.ID
	exampleEvent.Options = append(exampleEvent.Options, exampleOption2)
	exampleMealPlan.Events = []*types.MealPlanEvent{exampleEvent}

	// Create directly to avoid status assertion issues
	mealPlanInput := converters.ConvertMealPlanToMealPlanDatabaseCreationInput(exampleMealPlan)
	mealPlan, err := dbc.CreateMealPlan(ctx, mealPlanInput)
	require.NoError(t, err)
	require.Equal(t, string(types.MealPlanStatusAwaitingVotes), mealPlan.Status)

	// Refetch the meal plan to ensure we have the full nested structure with proper IDs
	mealPlan, err = dbc.GetMealPlan(ctx, mealPlan.ID, account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, mealPlan.Events)
	require.NotEmpty(t, mealPlan.Events[0].Options)

	mealPlanEvent := mealPlan.Events[0]
	mealPlanOption := mealPlanEvent.Options[0]

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.MealPlanOptionVote]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "meal plan option vote",
		CreateItem: func(ctx context.Context, i int) *types.MealPlanOptionVote {
			// Create voting users and add them to the account
			votingUser := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
			addUserToAccountHelper(votingUser.ID)

			mealPlanOptionVote := fakes.BuildFakeMealPlanOptionVote()
			mealPlanOptionVote.ByUser = votingUser.ID
			mealPlanOptionVote.BelongsToMealPlanOption = mealPlanOption.ID
			return createMealPlanOptionVoteForTest(t, ctx, mealPlan.ID, mealPlanEvent.ID, mealPlanOptionVote, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanOptionVote], error) {
			return dbc.GetMealPlanOptionVotes(ctx, mealPlan.ID, mealPlanEvent.ID, mealPlanOption.ID, filter)
		},
		GetID: func(mealPlanOptionVote *types.MealPlanOptionVote) string {
			return mealPlanOptionVote.ID
		},
		CleanupItem: func(ctx context.Context, mealPlanOptionVote *types.MealPlanOptionVote) error {
			return dbc.ArchiveMealPlanOptionVote(ctx, mealPlan.ID, mealPlanEvent.ID, mealPlanOption.ID, mealPlanOptionVote.ID)
		},
	})
}
