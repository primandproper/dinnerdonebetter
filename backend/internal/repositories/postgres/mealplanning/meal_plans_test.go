package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildMealPlanForIntegrationTest(userID string, meal *types.Meal) *types.MealPlan {
	exampleMealPlan := fakes.BuildFakeMealPlan()
	exampleMealPlan.CreatedByUser = userID

	exampleMealPlan.Events = []*types.MealPlanEvent{
		buildMealPlanEventForIntegrationTest(meal),
	}

	// only one event means it's immediately finalized
	exampleMealPlan.Status = string(types.MealPlanStatusFinalized)

	return exampleMealPlan
}

func createMealPlanForTest(t *testing.T, ctx context.Context, exampleMealPlan *types.MealPlan, dbc *repository) *types.MealPlan {
	t.Helper()

	// create
	dbInput := converters.ConvertMealPlanToMealPlanDatabaseCreationInput(exampleMealPlan)

	created, err := dbc.CreateMealPlan(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleMealPlan.CreatedAt = created.CreatedAt
	for i := range created.Events {
		assert.Equal(t, created.Events[i].ID, exampleMealPlan.Events[i].ID)
		exampleMealPlan.Events[i].CreatedAt = created.Events[i].CreatedAt
		exampleMealPlan.Events[i].StartsAt = created.Events[i].StartsAt
		exampleMealPlan.Events[i].EndsAt = created.Events[i].EndsAt
		exampleMealPlan.Events[i].BelongsToMealPlan = created.Events[i].BelongsToMealPlan

		for j := range created.Events[i].Options {
			assert.Equal(t, created.Events[i].Options[j].ID, exampleMealPlan.Events[i].Options[j].ID)
			assert.Equal(t, created.Events[i].Options[j].Meal.ID, exampleMealPlan.Events[i].Options[j].Meal.ID)
			exampleMealPlan.Events[i].Options[j] = created.Events[i].Options[j]
		}

		assert.Equal(t, created.Events[i].Options, exampleMealPlan.Events[i].Options)
	}
	assert.Equal(t, exampleMealPlan, created)

	mealPlan, err := dbc.GetMealPlan(ctx, created.ID, created.BelongsToAccount)
	require.NoError(t, err)

	exampleMealPlan.CreatedAt = mealPlan.CreatedAt
	exampleMealPlan.VotingDeadline = mealPlan.VotingDeadline
	for i := range mealPlan.Events {
		assert.Equal(t, mealPlan.Events[i].ID, exampleMealPlan.Events[i].ID)
		exampleMealPlan.Events[i].CreatedAt = mealPlan.Events[i].CreatedAt
		exampleMealPlan.Events[i].StartsAt = mealPlan.Events[i].StartsAt
		exampleMealPlan.Events[i].EndsAt = mealPlan.Events[i].EndsAt
		exampleMealPlan.Events[i].BelongsToMealPlan = mealPlan.Events[i].BelongsToMealPlan

		for j := range mealPlan.Events[i].Options {
			assert.Equal(t, mealPlan.Events[i].Options[j].ID, exampleMealPlan.Events[i].Options[j].ID)
			assert.Equal(t, mealPlan.Events[i].Options[j].Meal.ID, exampleMealPlan.Events[i].Options[j].Meal.ID)
			exampleMealPlan.Events[i].Options[j] = mealPlan.Events[i].Options[j]
		}

		assert.Equal(t, mealPlan.Events[i].Options, exampleMealPlan.Events[i].Options)
	}

	require.Equal(t, exampleMealPlan.CreatedAt, mealPlan.CreatedAt)
	require.Equal(t, exampleMealPlan.VotingDeadline, mealPlan.VotingDeadline)
	require.Equal(t, exampleMealPlan.ArchivedAt, mealPlan.ArchivedAt)
	require.Equal(t, exampleMealPlan.LastUpdatedAt, mealPlan.LastUpdatedAt)
	require.Equal(t, exampleMealPlan.ID, mealPlan.ID)
	require.Equal(t, exampleMealPlan.Status, mealPlan.Status)
	require.Equal(t, exampleMealPlan.Notes, mealPlan.Notes)
	require.Equal(t, exampleMealPlan.ElectionMethod, mealPlan.ElectionMethod)
	require.Equal(t, exampleMealPlan.BelongsToAccount, mealPlan.BelongsToAccount)
	require.Equal(t, exampleMealPlan.CreatedByUser, mealPlan.CreatedByUser)
	require.Equal(t, exampleMealPlan.GroceryListInitialized, mealPlan.GroceryListInitialized)
	require.Equal(t, exampleMealPlan.TasksCreated, mealPlan.TasksCreated)

	for i := range mealPlan.Events {
		require.Equal(t, exampleMealPlan.Events[i].CreatedAt, mealPlan.Events[i].CreatedAt)
		require.Equal(t, exampleMealPlan.Events[i].StartsAt, mealPlan.Events[i].StartsAt)
		require.Equal(t, exampleMealPlan.Events[i].EndsAt, mealPlan.Events[i].EndsAt)
		require.Equal(t, exampleMealPlan.Events[i].ArchivedAt, mealPlan.Events[i].ArchivedAt)
		require.Equal(t, exampleMealPlan.Events[i].LastUpdatedAt, mealPlan.Events[i].LastUpdatedAt)
		require.Equal(t, exampleMealPlan.Events[i].MealName, mealPlan.Events[i].MealName)
		require.Equal(t, exampleMealPlan.Events[i].Notes, mealPlan.Events[i].Notes)
		require.Equal(t, exampleMealPlan.Events[i].BelongsToMealPlan, mealPlan.Events[i].BelongsToMealPlan)
		require.Equal(t, exampleMealPlan.Events[i].ID, mealPlan.Events[i].ID)

		for j := range mealPlan.Events[i].Options {
			require.Equal(t, exampleMealPlan.Events[i].Options[j].CreatedAt, mealPlan.Events[i].Options[j].CreatedAt)
			require.Equal(t, exampleMealPlan.Events[i].Options[j].LastUpdatedAt, mealPlan.Events[i].Options[j].LastUpdatedAt)
			require.Equal(t, exampleMealPlan.Events[i].Options[j].AssignedCook, mealPlan.Events[i].Options[j].AssignedCook)
			require.Equal(t, exampleMealPlan.Events[i].Options[j].ArchivedAt, mealPlan.Events[i].Options[j].ArchivedAt)
			require.Equal(t, exampleMealPlan.Events[i].Options[j].AssignedDishwasher, mealPlan.Events[i].Options[j].AssignedDishwasher)
			require.Equal(t, exampleMealPlan.Events[i].Options[j].Notes, mealPlan.Events[i].Options[j].Notes)
			require.Equal(t, exampleMealPlan.Events[i].Options[j].BelongsToMealPlanEvent, mealPlan.Events[i].Options[j].BelongsToMealPlanEvent)
			require.Equal(t, exampleMealPlan.Events[i].Options[j].ID, mealPlan.Events[i].Options[j].ID)
			require.Equal(t, exampleMealPlan.Events[i].Options[j].Votes, mealPlan.Events[i].Options[j].Votes)
			require.Equal(t, exampleMealPlan.Events[i].Options[j].Meal, mealPlan.Events[i].Options[j].Meal)
			require.Equal(t, exampleMealPlan.Events[i].Options[j].MealScale, mealPlan.Events[i].Options[j].MealScale)
			require.Equal(t, exampleMealPlan.Events[i].Options[j].Chosen, mealPlan.Events[i].Options[j].Chosen)
			require.Equal(t, exampleMealPlan.Events[i].Options[j].TieBroken, mealPlan.Events[i].Options[j].TieBroken)
		}
	}

	assert.Equal(t, exampleMealPlan, mealPlan)

	return mealPlan
}

func TestQuerier_Integration_MealPlans(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)
	accountID := account.ID

	recipe := createRecipeForTest(t, ctx, nil, dbc, true)
	meal := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe), dbc)

	exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
	exampleMealPlan.BelongsToAccount = accountID
	createdMealPlans := []*types.MealPlan{}

	// create
	createdMealPlans = append(createdMealPlans, createMealPlanForTest(t, ctx, exampleMealPlan, dbc))

	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, accountID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeMealPlans, RelevantID: createdMealPlans[0].ID},
	})

	// create more
	for range exampleQuantity {
		input := buildMealPlanForIntegrationTest(user.ID, meal)
		input.BelongsToAccount = accountID
		createdMealPlans = append(createdMealPlans, createMealPlanForTest(t, ctx, input, dbc))
	}

	// fetch as list
	mealPlans, err := dbc.GetMealPlansForAccount(ctx, accountID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, mealPlans.Data)
	assert.Len(t, mealPlans.Data, len(createdMealPlans))

	_, err = dbc.GetMealPlansAwaitingFinalizationSaga(ctx, 10)
	assert.NoError(t, err)
	_, err = dbc.GetFinalizedMealPlanOptionsForMealPlan(ctx, createdMealPlans[0].ID)
	assert.NoError(t, err)
	_, err = dbc.FetchMissingVotesForMealPlan(ctx, createdMealPlans[0].ID, accountID)
	assert.NoError(t, err)

	// delete
	for _, mealPlan := range createdMealPlans {
		_, err = dbc.AttemptToFinalizeMealPlan(ctx, mealPlan.ID, accountID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrAlreadyFinalized)
		assert.NoError(t, dbc.ArchiveMealPlan(ctx, mealPlan.ID, accountID))

		pgtesting.AssertAuditLogContains(t, ctx, auditRepo, accountID, []*audit.AuditLogEntry{
			{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeMealPlans, RelevantID: mealPlan.ID},
		})

		var exists bool
		exists, err = dbc.MealPlanExists(ctx, mealPlan.ID, accountID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.MealPlan
		y, err = dbc.GetMealPlan(ctx, mealPlan.ID, accountID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_MealPlanExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleAccountID := fakes.BuildFakeID()
		c := buildInertClientForTest(t)

		actual, err := c.MealPlanExists(ctx, "", exampleAccountID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid account MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleMealPlanID := fakes.BuildFakeID()
		c := buildInertClientForTest(t)

		actual, err := c.MealPlanExists(ctx, exampleMealPlanID, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlan(ctx, "", exampleAccountID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid account MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlanID := fakes.BuildFakeID()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlan(ctx, exampleMealPlanID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateMealPlan(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateMealPlan(ctx, nil))
	})
}

func TestQuerier_ArchiveMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveMealPlan(ctx, "", exampleAccountID))
	})

	T.Run("with invalid account MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlan := fakes.BuildFakeMealPlan()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveMealPlan(ctx, exampleMealPlan.ID, ""))
	})
}

func TestQuerier_AttemptToFinalizeCompleteMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()
		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.AttemptToFinalizeMealPlan(ctx, "", exampleAccountID)
		assert.False(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid account MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleMealPlan := fakes.BuildFakeMealPlan()
		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.AttemptToFinalizeMealPlan(ctx, exampleMealPlan.ID, "")
		assert.False(t, actual)
		assert.Error(t, err)
	})
}

func TestQuerier_FetchMissingVotesForMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("with missing meal plan MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)
		exampleAccountID := fakes.BuildFakeID()

		actual, err := c.FetchMissingVotesForMealPlan(ctx, "", exampleAccountID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with missing account MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)
		exampleMealPlan := fakes.BuildFakeMealPlan()

		actual, err := c.FetchMissingVotesForMealPlan(ctx, exampleMealPlan.ID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_Integration_MealPlans_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.MealPlan]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "meal plan",
		CreateItem: func(ctx context.Context, i int) *types.MealPlan {
			// Create a unique recipe and meal for each meal plan
			recipe := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)
			meal := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe), dbc)
			mealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
			mealPlan.BelongsToAccount = account.ID
			return createMealPlanForTest(t, ctx, mealPlan, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlan], error) {
			return dbc.GetMealPlansForAccount(ctx, account.ID, filter)
		},
		GetID: func(mealPlan *types.MealPlan) string {
			return mealPlan.ID
		},
		CleanupItem: func(ctx context.Context, mealPlan *types.MealPlan) error {
			return dbc.ArchiveMealPlan(ctx, mealPlan.ID, account.ID)
		},
	})
}

func TestQuerier_Integration_MealPlans_WithSelections(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)
	accountID := account.ID

	// Create recipe with a step that has ingredients
	recipe := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)
	require.NotEmpty(t, recipe.Steps)
	recipeStep := recipe.Steps[0]

	// Create meal with the recipe as a component
	meal := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe), dbc)

	t.Run("creates meal plan with matching selections", func(t *testing.T) {
		// Build meal plan with selections
		exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
		exampleMealPlan.BelongsToAccount = accountID

		dbInput := converters.ConvertMealPlanToMealPlanDatabaseCreationInput(exampleMealPlan)

		// Add a selection that should match the recipe in the meal
		dbInput.Selections = []*types.MealPlanRecipeOptionSelectionDatabaseCreationInput{
			{
				ID:                  fakes.BuildFakeID(),
				RecipeID:            recipe.ID,
				RecipeStepID:        recipeStep.ID,
				IngredientIndex:     0,
				SelectedOptionIndex: 0,
				SelectionType:       types.MealPlanRecipeOptionSelectionTypeIngredient,
			},
		}

		created, createErr := dbc.CreateMealPlan(ctx, dbInput)
		assert.NoError(t, createErr)
		require.NotNil(t, created)
		require.NotEmpty(t, created.Events)
		require.NotEmpty(t, created.Events[0].Options)

		// Verify the selection was created by querying for it
		optionID := created.Events[0].Options[0].ID
		selections, selectionErr := dbc.GetSelectionsForMealPlanOption(ctx, optionID, nil)
		assert.NoError(t, selectionErr)
		require.NotEmpty(t, selections.Data)
		assert.Equal(t, recipe.ID, selections.Data[0].RecipeID)
		assert.Equal(t, recipeStep.ID, selections.Data[0].RecipeStepID)
		assert.Equal(t, uint16(0), selections.Data[0].IngredientIndex)
		assert.Equal(t, uint16(0), selections.Data[0].SelectedOptionIndex)
		assert.Equal(t, types.MealPlanRecipeOptionSelectionTypeIngredient, selections.Data[0].SelectionType)

		// Cleanup
		assert.NoError(t, dbc.ArchiveMealPlan(ctx, created.ID, accountID))
	})

	t.Run("creates meal plan with multiple selections", func(t *testing.T) {
		// Build meal plan with selections
		exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
		exampleMealPlan.BelongsToAccount = accountID

		dbInput := converters.ConvertMealPlanToMealPlanDatabaseCreationInput(exampleMealPlan)

		// Add multiple selections for the same recipe
		dbInput.Selections = []*types.MealPlanRecipeOptionSelectionDatabaseCreationInput{
			{
				ID:                  fakes.BuildFakeID(),
				RecipeID:            recipe.ID,
				RecipeStepID:        recipeStep.ID,
				IngredientIndex:     0,
				SelectedOptionIndex: 1,
				SelectionType:       types.MealPlanRecipeOptionSelectionTypeIngredient,
			},
			{
				ID:                  fakes.BuildFakeID(),
				RecipeID:            recipe.ID,
				RecipeStepID:        recipeStep.ID,
				IngredientIndex:     1,
				SelectedOptionIndex: 0,
				SelectionType:       types.MealPlanRecipeOptionSelectionTypeInstrument,
			},
			{
				ID:                  fakes.BuildFakeID(),
				RecipeID:            recipe.ID,
				RecipeStepID:        recipeStep.ID,
				IngredientIndex:     2,
				SelectedOptionIndex: 0,
				SelectionType:       types.MealPlanRecipeOptionSelectionTypeVessel,
			},
		}

		created, createErr := dbc.CreateMealPlan(ctx, dbInput)
		assert.NoError(t, createErr)
		require.NotNil(t, created)
		require.NotEmpty(t, created.Events)
		require.NotEmpty(t, created.Events[0].Options)

		// Verify all selections were created
		optionID := created.Events[0].Options[0].ID
		selections, selectionErr := dbc.GetSelectionsForMealPlanOption(ctx, optionID, nil)
		assert.NoError(t, selectionErr)
		assert.Len(t, selections.Data, 3)

		// Verify selection types
		selectionTypes := make(map[string]bool)
		for _, sel := range selections.Data {
			selectionTypes[sel.SelectionType] = true
		}
		assert.True(t, selectionTypes[types.MealPlanRecipeOptionSelectionTypeIngredient])
		assert.True(t, selectionTypes[types.MealPlanRecipeOptionSelectionTypeInstrument])
		assert.True(t, selectionTypes[types.MealPlanRecipeOptionSelectionTypeVessel])

		// Cleanup
		assert.NoError(t, dbc.ArchiveMealPlan(ctx, created.ID, accountID))
	})

	t.Run("creates meal plan and skips non-matching selections", func(t *testing.T) {
		// Build meal plan with selections
		exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
		exampleMealPlan.BelongsToAccount = accountID

		dbInput := converters.ConvertMealPlanToMealPlanDatabaseCreationInput(exampleMealPlan)

		// Add a selection with a non-existent recipe ID - should be skipped
		dbInput.Selections = []*types.MealPlanRecipeOptionSelectionDatabaseCreationInput{
			{
				ID:                  fakes.BuildFakeID(),
				RecipeID:            fakes.BuildFakeID(), // Non-existent recipe
				RecipeStepID:        fakes.BuildFakeID(),
				IngredientIndex:     0,
				SelectedOptionIndex: 0,
				SelectionType:       types.MealPlanRecipeOptionSelectionTypeIngredient,
			},
		}

		// Should succeed even though selection doesn't match
		created, createErr := dbc.CreateMealPlan(ctx, dbInput)
		assert.NoError(t, createErr)
		require.NotNil(t, created)
		require.NotEmpty(t, created.Events)
		require.NotEmpty(t, created.Events[0].Options)

		// Verify no selections were created (since recipe didn't match)
		optionID := created.Events[0].Options[0].ID
		selections, selectionErr := dbc.GetSelectionsForMealPlanOption(ctx, optionID, nil)
		assert.NoError(t, selectionErr)
		assert.Empty(t, selections.Data)

		// Cleanup
		assert.NoError(t, dbc.ArchiveMealPlan(ctx, created.ID, accountID))
	})

	t.Run("creates meal plan with mixed matching and non-matching selections", func(t *testing.T) {
		// Build meal plan with selections
		exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
		exampleMealPlan.BelongsToAccount = accountID

		dbInput := converters.ConvertMealPlanToMealPlanDatabaseCreationInput(exampleMealPlan)

		// Add one matching and one non-matching selection
		dbInput.Selections = []*types.MealPlanRecipeOptionSelectionDatabaseCreationInput{
			{
				ID:                  fakes.BuildFakeID(),
				RecipeID:            recipe.ID, // Matching recipe
				RecipeStepID:        recipeStep.ID,
				IngredientIndex:     0,
				SelectedOptionIndex: 0,
				SelectionType:       types.MealPlanRecipeOptionSelectionTypeIngredient,
			},
			{
				ID:                  fakes.BuildFakeID(),
				RecipeID:            fakes.BuildFakeID(), // Non-matching recipe
				RecipeStepID:        fakes.BuildFakeID(),
				IngredientIndex:     1,
				SelectedOptionIndex: 0,
				SelectionType:       types.MealPlanRecipeOptionSelectionTypeIngredient,
			},
		}

		created, createErr := dbc.CreateMealPlan(ctx, dbInput)
		assert.NoError(t, createErr)
		require.NotNil(t, created)
		require.NotEmpty(t, created.Events)
		require.NotEmpty(t, created.Events[0].Options)

		// Verify only the matching selection was created
		optionID := created.Events[0].Options[0].ID
		selections, selectionErr := dbc.GetSelectionsForMealPlanOption(ctx, optionID, nil)
		assert.NoError(t, selectionErr)
		assert.Len(t, selections.Data, 1)
		assert.Equal(t, recipe.ID, selections.Data[0].RecipeID)

		// Cleanup
		assert.NoError(t, dbc.ArchiveMealPlan(ctx, created.ID, accountID))
	})

	t.Run("creates meal plan without selections", func(t *testing.T) {
		// Build meal plan without selections
		exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
		exampleMealPlan.BelongsToAccount = accountID

		dbInput := converters.ConvertMealPlanToMealPlanDatabaseCreationInput(exampleMealPlan)
		// Ensure no selections
		dbInput.Selections = nil

		created, createErr := dbc.CreateMealPlan(ctx, dbInput)
		assert.NoError(t, createErr)
		require.NotNil(t, created)
		require.NotEmpty(t, created.Events)
		require.NotEmpty(t, created.Events[0].Options)

		// Verify no selections were created
		optionID := created.Events[0].Options[0].ID
		selections, selectionErr := dbc.GetSelectionsForMealPlanOption(ctx, optionID, nil)
		assert.NoError(t, selectionErr)
		assert.Empty(t, selections.Data)

		// Cleanup
		assert.NoError(t, dbc.ArchiveMealPlan(ctx, created.ID, accountID))
	})

	t.Run("GetMealPlan returns selections in Selections field", func(t *testing.T) {
		// Build meal plan with selections
		exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
		exampleMealPlan.BelongsToAccount = accountID

		dbInput := converters.ConvertMealPlanToMealPlanDatabaseCreationInput(exampleMealPlan)

		// Add multiple selections
		dbInput.Selections = []*types.MealPlanRecipeOptionSelectionDatabaseCreationInput{
			{
				ID:                  fakes.BuildFakeID(),
				RecipeID:            recipe.ID,
				RecipeStepID:        recipeStep.ID,
				IngredientIndex:     0,
				SelectedOptionIndex: 1,
				SelectionType:       types.MealPlanRecipeOptionSelectionTypeIngredient,
			},
			{
				ID:                  fakes.BuildFakeID(),
				RecipeID:            recipe.ID,
				RecipeStepID:        recipeStep.ID,
				IngredientIndex:     1,
				SelectedOptionIndex: 0,
				SelectionType:       types.MealPlanRecipeOptionSelectionTypeInstrument,
			},
		}

		created, createErr := dbc.CreateMealPlan(ctx, dbInput)
		assert.NoError(t, createErr)
		require.NotNil(t, created)

		// Now fetch the meal plan via GetMealPlan
		fetched, fetchErr := dbc.GetMealPlan(ctx, created.ID, accountID)
		assert.NoError(t, fetchErr)
		require.NotNil(t, fetched)

		// Verify the Selections field is populated
		require.NotNil(t, fetched.Selections)
		assert.Len(t, fetched.Selections, 2)

		// Verify selection contents
		selectionTypes := make(map[string]bool)
		for _, sel := range fetched.Selections {
			assert.Equal(t, recipe.ID, sel.RecipeID)
			assert.Equal(t, recipeStep.ID, sel.RecipeStepID)
			selectionTypes[sel.SelectionType] = true
		}
		assert.True(t, selectionTypes[types.MealPlanRecipeOptionSelectionTypeIngredient])
		assert.True(t, selectionTypes[types.MealPlanRecipeOptionSelectionTypeInstrument])

		// Cleanup
		assert.NoError(t, dbc.ArchiveMealPlan(ctx, created.ID, accountID))
	})

	t.Run("GetMealPlan returns nil Selections when none exist", func(t *testing.T) {
		// Build meal plan without selections
		exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
		exampleMealPlan.BelongsToAccount = accountID

		dbInput := converters.ConvertMealPlanToMealPlanDatabaseCreationInput(exampleMealPlan)
		dbInput.Selections = nil

		created, createErr := dbc.CreateMealPlan(ctx, dbInput)
		assert.NoError(t, createErr)
		require.NotNil(t, created)

		// Now fetch the meal plan via GetMealPlan
		fetched, fetchErr := dbc.GetMealPlan(ctx, created.ID, accountID)
		assert.NoError(t, fetchErr)
		require.NotNil(t, fetched)

		// Verify the Selections field is nil (not an empty slice)
		assert.Nil(t, fetched.Selections)

		// Cleanup
		assert.NoError(t, dbc.ArchiveMealPlan(ctx, created.ID, accountID))
	})
}
