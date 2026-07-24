package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v6/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRecipeStepInstrumentForTest(t *testing.T, ctx context.Context, recipeID string, exampleRecipeStepInstrument *types.RecipeStepInstrument, dbc *repository) *types.RecipeStepInstrument {
	t.Helper()

	// create
	if exampleRecipeStepInstrument == nil {
		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
		exampleRecipe := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
		createdRecipe := createRecipeForTest(t, ctx, exampleRecipe, dbc, true)
		exampleRecipeStep := createdRecipe.Steps[0]

		exampleRecipeStepInstrument = fakes.BuildFakeRecipeStepInstrument()
		exampleRecipeStepInstrument.BelongsToRecipeStep = exampleRecipeStep.ID
	}
	dbInput := converters.ConvertRecipeStepInstrumentToRecipeStepInstrumentDatabaseCreationInput(exampleRecipeStepInstrument)

	created, err := dbc.CreateRecipeStepInstrument(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleRecipeStepInstrument.CreatedAt = created.CreatedAt
	exampleRecipeStepInstrument.Instrument.CreatedAt = created.Instrument.CreatedAt
	exampleRecipeStepInstrument.Instrument = created.Instrument
	assert.Equal(t, exampleRecipeStepInstrument, created)

	recipeStepInstrument, err := dbc.GetRecipeStepInstrument(ctx, recipeID, exampleRecipeStepInstrument.BelongsToRecipeStep, exampleRecipeStepInstrument.ID)

	exampleRecipeStepInstrument.CreatedAt = recipeStepInstrument.CreatedAt
	exampleRecipeStepInstrument.Instrument.CreatedAt = recipeStepInstrument.Instrument.CreatedAt
	exampleRecipeStepInstrument.Instrument = recipeStepInstrument.Instrument

	require.Equal(t, exampleRecipeStepInstrument, recipeStepInstrument)

	assert.NoError(t, err)
	assert.Equal(t, recipeStepInstrument, exampleRecipeStepInstrument)

	return created
}

func TestQuerier_Integration_RecipeStepInstruments(t *testing.T) {
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
	exampleRecipeStep := createdRecipe.Steps[0]

	validInstrument := createValidInstrumentForTest(t, ctx, nil, dbc)
	exampleRecipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()
	exampleRecipeStepInstrument.Instrument = validInstrument
	exampleRecipeStepInstrument.BelongsToRecipeStep = exampleRecipeStep.ID
	// Set unique index - first instrument from recipe creation has index 0, so start at 1
	exampleRecipeStepInstrument.Index = 1
	exampleRecipeStepInstrument.OptionIndex = 0
	createdRecipeStepInstruments := []*types.RecipeStepInstrument{
		exampleRecipeStep.Instruments[0],
	}

	// create
	createdRecipeStepInstruments = append(createdRecipeStepInstruments, createRecipeStepInstrumentForTest(t, ctx, exampleRecipe.ID, exampleRecipeStepInstrument, dbc))

	// create more
	for i := range exampleQuantity {
		validInstrument = createValidInstrumentForTest(t, ctx, nil, dbc)
		input := fakes.BuildFakeRecipeStepInstrument()
		input.Instrument = validInstrument
		input.BelongsToRecipeStep = exampleRecipeStep.ID
		// Set unique index - start from 2 since 0 and 1 are already taken
		input.Index = uint16(i + 2)
		input.OptionIndex = 0
		createdRecipeStepInstruments = append(createdRecipeStepInstruments, createRecipeStepInstrumentForTest(t, ctx, exampleRecipe.ID, input, dbc))
	}

	// fetch as list
	recipeStepInstruments, err := dbc.GetRecipeStepInstruments(ctx, exampleRecipe.ID, createdRecipeStepInstruments[0].BelongsToRecipeStep, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, recipeStepInstruments.Data)
	assert.Equal(t, len(createdRecipeStepInstruments), len(recipeStepInstruments.Data))

	// delete
	for _, recipeStepInstrument := range createdRecipeStepInstruments {
		assert.NoError(t, dbc.ArchiveRecipeStepInstrument(ctx, exampleRecipeStep.ID, recipeStepInstrument.ID))

		var exists bool
		exists, err = dbc.RecipeStepInstrumentExists(ctx, exampleRecipe.ID, recipeStepInstrument.BelongsToRecipeStep, recipeStepInstrument.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.RecipeStepInstrument
		y, err = dbc.GetRecipeStepInstrument(ctx, exampleRecipe.ID, recipeStepInstrument.BelongsToRecipeStep, recipeStepInstrument.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_RecipeStepInstrumentExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeStepID := fakes.BuildFakeID()
		exampleRecipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepInstrumentExists(ctx, "", exampleRecipeStepID, exampleRecipeStepInstrument.ID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepInstrumentExists(ctx, exampleRecipeID, "", exampleRecipeStepInstrument.ID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid recipe step instrument MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepInstrumentExists(ctx, exampleRecipeID, exampleRecipeStepID, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepID := fakes.BuildFakeID()
		exampleRecipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepInstrument(ctx, "", exampleRecipeStepID, exampleRecipeStepInstrument.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepInstrument(ctx, exampleRecipeID, "", exampleRecipeStepInstrument.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step instrument MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetRecipeStepInstruments(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()
		exampleRecipeStepID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepInstruments(ctx, "", exampleRecipeStepID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()
		exampleRecipeID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepInstruments(ctx, exampleRecipeID, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateRecipeStepInstrument(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateRecipeStepInstrument(ctx, nil))
	})
}

func TestQuerier_ArchiveRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeStepInstrument(ctx, "", exampleRecipeStepInstrument.ID))
	})

	T.Run("with invalid recipe step instrument MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeStepInstrument(ctx, exampleRecipeStepID, ""))
	})
}

func TestQuerier_Integration_RecipeStepInstruments_CursorBasedPagination(t *testing.T) {
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
	recipeStruct := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
	// Clear the default instruments from the step so we start fresh
	for _, step := range recipeStruct.Steps {
		step.Instruments = nil // Use nil instead of empty slice to match database behavior
	}
	recipe := createRecipeForTest(t, ctx, recipeStruct, dbc, false)
	recipeStep := recipe.Steps[0]

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.RecipeStepInstrument]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "recipe step instrument",
		CreateItem: func(ctx context.Context, i int) *types.RecipeStepInstrument {
			instrument := createValidInstrumentForTest(t, ctx, nil, dbc)
			recipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()
			recipeStepInstrument.BelongsToRecipeStep = recipeStep.ID
			recipeStepInstrument.Instrument = instrument
			// Set unique index for each instrument to avoid constraint violations
			recipeStepInstrument.Index = uint16(i)
			recipeStepInstrument.OptionIndex = 0
			return createRecipeStepInstrumentForTest(t, ctx, recipe.ID, recipeStepInstrument, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepInstrument], error) {
			return dbc.GetRecipeStepInstruments(ctx, recipe.ID, recipeStep.ID, filter)
		},
		GetID: func(recipeStepInstrument *types.RecipeStepInstrument) string {
			return recipeStepInstrument.ID
		},
		CleanupItem: func(ctx context.Context, recipeStepInstrument *types.RecipeStepInstrument) error {
			return dbc.ArchiveRecipeStepInstrument(ctx, recipeStep.ID, recipeStepInstrument.ID)
		},
	})
}
