package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v8/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRecipeStepVesselForTest(t *testing.T, ctx context.Context, recipeID string, exampleRecipeStepVessel *types.RecipeStepVessel, dbc *repository) *types.RecipeStepVessel {
	t.Helper()

	// create
	if exampleRecipeStepVessel == nil {
		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
		exampleRecipe := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
		createdRecipe := createRecipeForTest(t, ctx, exampleRecipe, dbc, true)
		exampleRecipeStep := createdRecipe.Steps[0]

		exampleRecipeStepVessel = fakes.BuildFakeRecipeStepVessel()
		exampleRecipeStepVessel.BelongsToRecipeStep = exampleRecipeStep.ID
	}
	dbInput := converters.ConvertRecipeStepVesselToRecipeStepVesselDatabaseCreationInput(exampleRecipeStepVessel)

	created, err := dbc.CreateRecipeStepVessel(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleRecipeStepVessel.CreatedAt = created.CreatedAt
	exampleRecipeStepVessel.Vessel.CreatedAt = created.Vessel.CreatedAt
	exampleRecipeStepVessel.Vessel = created.Vessel
	assert.Equal(t, exampleRecipeStepVessel, created)

	recipeStepVessel, err := dbc.GetRecipeStepVessel(ctx, recipeID, exampleRecipeStepVessel.BelongsToRecipeStep, exampleRecipeStepVessel.ID)

	exampleRecipeStepVessel.CreatedAt = recipeStepVessel.CreatedAt
	exampleRecipeStepVessel.Vessel.CreatedAt = recipeStepVessel.Vessel.CreatedAt
	exampleRecipeStepVessel.Vessel = recipeStepVessel.Vessel

	require.Equal(t, exampleRecipeStepVessel, recipeStepVessel)

	assert.NoError(t, err)
	assert.Equal(t, recipeStepVessel, exampleRecipeStepVessel)

	return created
}

func TestQuerier_Integration_RecipeStepVessels(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	exampleRecipe := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
	createdRecipe := createRecipeForTest(t, ctx, exampleRecipe, dbc, true)
	exampleRecipeStep := createdRecipe.Steps[0]

	validVessel := createValidVesselForTest(t, ctx, nil, dbc)
	exampleRecipeStepVessel := fakes.BuildFakeRecipeStepVessel()
	exampleRecipeStepVessel.Vessel = validVessel
	exampleRecipeStepVessel.BelongsToRecipeStep = exampleRecipeStep.ID
	// Set unique index - first vessel from recipe creation has index 0, so start at 1
	exampleRecipeStepVessel.Index = 1
	exampleRecipeStepVessel.OptionIndex = 0
	createdRecipeStepVessels := []*types.RecipeStepVessel{
		exampleRecipeStep.Vessels[0],
	}

	// create
	createdRecipeStepVessels = append(createdRecipeStepVessels, createRecipeStepVesselForTest(t, ctx, exampleRecipe.ID, exampleRecipeStepVessel, dbc))

	// create more
	for i := range exampleQuantity {
		validVessel = createValidVesselForTest(t, ctx, nil, dbc)
		input := fakes.BuildFakeRecipeStepVessel()
		input.Vessel = validVessel
		input.BelongsToRecipeStep = exampleRecipeStep.ID
		// Set unique index - start from 2 since 0 and 1 are already taken
		input.Index = uint16(i + 2)
		input.OptionIndex = 0
		createdRecipeStepVessels = append(createdRecipeStepVessels, createRecipeStepVesselForTest(t, ctx, exampleRecipe.ID, input, dbc))
	}

	// fetch as list
	recipeStepVessels, err := dbc.GetRecipeStepVessels(ctx, exampleRecipe.ID, createdRecipeStepVessels[0].BelongsToRecipeStep, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, recipeStepVessels.Data)
	assert.Equal(t, len(createdRecipeStepVessels), len(recipeStepVessels.Data))

	// delete
	for _, recipeStepVessel := range createdRecipeStepVessels {
		assert.NoError(t, dbc.ArchiveRecipeStepVessel(ctx, exampleRecipeStep.ID, recipeStepVessel.ID))

		var exists bool
		exists, err = dbc.RecipeStepVesselExists(ctx, exampleRecipe.ID, recipeStepVessel.BelongsToRecipeStep, recipeStepVessel.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.RecipeStepVessel
		y, err = dbc.GetRecipeStepVessel(ctx, exampleRecipe.ID, recipeStepVessel.BelongsToRecipeStep, recipeStepVessel.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_RecipeStepVesselExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeStepID := fakes.BuildFakeID()
		exampleRecipeStepVessel := fakes.BuildFakeRecipeStepVessel()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepVesselExists(ctx, "", exampleRecipeStepID, exampleRecipeStepVessel.ID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepVessel := fakes.BuildFakeRecipeStepVessel()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepVesselExists(ctx, exampleRecipeID, "", exampleRecipeStepVessel.ID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid recipe step instrument MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepVesselExists(ctx, exampleRecipeID, exampleRecipeStepID, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepID := fakes.BuildFakeID()
		exampleRecipeStepVessel := fakes.BuildFakeRecipeStepVessel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepVessel(ctx, "", exampleRecipeStepID, exampleRecipeStepVessel.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepVessel := fakes.BuildFakeRecipeStepVessel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepVessel(ctx, exampleRecipeID, "", exampleRecipeStepVessel.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step instrument MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetRecipeStepVessels(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()
		exampleRecipeStepID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepVessels(ctx, "", exampleRecipeStepID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()
		exampleRecipeID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepVessels(ctx, exampleRecipeID, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateRecipeStepVessel(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateRecipeStepVessel(ctx, nil))
	})
}

func TestQuerier_ArchiveRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepVessel := fakes.BuildFakeRecipeStepVessel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeStepVessel(ctx, "", exampleRecipeStepVessel.ID))
	})

	T.Run("with invalid recipe step instrument MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeStepVessel(ctx, exampleRecipeStepID, ""))
	})
}

func TestQuerier_Integration_RecipeStepVessels_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	recipeStruct := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
	// Clear the default vessels from the step so we start fresh
	for _, step := range recipeStruct.Steps {
		step.Vessels = nil // Use nil instead of empty slice to match database behavior
	}
	recipe := createRecipeForTest(t, ctx, recipeStruct, dbc, false)
	recipeStep := recipe.Steps[0]

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.RecipeStepVessel]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "recipe step vessel",
		CreateItem: func(ctx context.Context, i int) *types.RecipeStepVessel {
			vessel := createValidVesselForTest(t, ctx, nil, dbc)
			recipeStepVessel := fakes.BuildFakeRecipeStepVessel()
			recipeStepVessel.BelongsToRecipeStep = recipeStep.ID
			recipeStepVessel.Vessel = vessel
			// Set unique index for each vessel to avoid constraint violations
			recipeStepVessel.Index = uint16(i)
			recipeStepVessel.OptionIndex = 0
			return createRecipeStepVesselForTest(t, ctx, recipe.ID, recipeStepVessel, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepVessel], error) {
			return dbc.GetRecipeStepVessels(ctx, recipe.ID, recipeStep.ID, filter)
		},
		GetID: func(recipeStepVessel *types.RecipeStepVessel) string {
			return recipeStepVessel.ID
		},
		CleanupItem: func(ctx context.Context, recipeStepVessel *types.RecipeStepVessel) error {
			return dbc.ArchiveRecipeStepVessel(ctx, recipeStep.ID, recipeStepVessel.ID)
		},
	})
}
