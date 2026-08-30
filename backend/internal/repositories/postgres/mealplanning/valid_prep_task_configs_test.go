package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v13/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createValidPrepTaskConfigForTest(t *testing.T, ctx context.Context, exampleValidPrepTaskConfig *types.ValidPrepTaskConfig, dbc *repository) *types.ValidPrepTaskConfig {
	t.Helper()

	// create
	if exampleValidPrepTaskConfig == nil {
		exampleValidIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
		exampleValidPreparation := createValidPreparationForTest(t, ctx, nil, dbc)
		exampleValidPrepTaskConfig = fakes.BuildFakeValidPrepTaskConfig()
		exampleValidPrepTaskConfig.Ingredient = *exampleValidIngredient
		exampleValidPrepTaskConfig.Preparation = *exampleValidPreparation
	}

	dbInput := converters.ConvertValidPrepTaskConfigToValidPrepTaskConfigDatabaseCreationInput(exampleValidPrepTaskConfig)

	created, err := dbc.CreateValidPrepTaskConfig(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)
	exampleValidPrepTaskConfig.CreatedAt = created.CreatedAt
	assertValidPrepTaskConfigsEqual(t, exampleValidPrepTaskConfig, created)

	validPrepTaskConfig, err := dbc.GetValidPrepTaskConfig(ctx, created.ID)
	require.NotNil(t, validPrepTaskConfig)
	exampleValidPrepTaskConfig.CreatedAt = validPrepTaskConfig.CreatedAt
	exampleValidPrepTaskConfig.Preparation = validPrepTaskConfig.Preparation
	exampleValidPrepTaskConfig.Ingredient = validPrepTaskConfig.Ingredient

	require.NoError(t, err)
	assertValidPrepTaskConfigsEqual(t, exampleValidPrepTaskConfig, validPrepTaskConfig)

	return created
}

func assertValidPrepTaskConfigsEqual(t *testing.T, expected, actual *types.ValidPrepTaskConfig) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.MinStorageDurationInSeconds, actual.MinStorageDurationInSeconds)
	assert.Equal(t, expected.StorageType, actual.StorageType)
	assert.Equal(t, expected.StorageInstructions, actual.StorageInstructions)
	assert.Equal(t, expected.Notes, actual.Notes)
	assert.Equal(t, expected.Source, actual.Source)
}

func TestQuerier_Integration_ValidPrepTaskConfigs(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	exampleValidIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
	exampleValidPreparation := createValidPreparationForTest(t, ctx, nil, dbc)
	exampleValidPrepTaskConfig := fakes.BuildFakeValidPrepTaskConfig()
	exampleValidPrepTaskConfig.Preparation = *exampleValidPreparation
	exampleValidPrepTaskConfig.Ingredient = *exampleValidIngredient
	createdValidPrepTaskConfigs := []*types.ValidPrepTaskConfig{}

	// create
	createdValidPrepTaskConfigs = append(createdValidPrepTaskConfigs, createValidPrepTaskConfigForTest(t, ctx, exampleValidPrepTaskConfig, dbc))

	// update
	updatedValidPrepTaskConfig := fakes.BuildFakeValidPrepTaskConfig()
	updatedValidPrepTaskConfig.ID = createdValidPrepTaskConfigs[0].ID
	updatedValidPrepTaskConfig.Preparation = createdValidPrepTaskConfigs[0].Preparation
	updatedValidPrepTaskConfig.Ingredient = createdValidPrepTaskConfigs[0].Ingredient
	require.NoError(t, dbc.UpdateValidPrepTaskConfig(ctx, updatedValidPrepTaskConfig))

	// create more
	for range exampleQuantity {
		input := fakes.BuildFakeValidPrepTaskConfig()
		input.Preparation = createdValidPrepTaskConfigs[0].Preparation
		input.Ingredient = createdValidPrepTaskConfigs[0].Ingredient
		createdValidPrepTaskConfigs = append(createdValidPrepTaskConfigs, createValidPrepTaskConfigForTest(t, ctx, input, dbc))
	}

	// fetch as list
	validPrepTaskConfigs, err := dbc.GetValidPrepTaskConfigs(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, validPrepTaskConfigs.Data)
	assert.Len(t, validPrepTaskConfigs.Data, len(createdValidPrepTaskConfigs))

	forIngredient, err := dbc.GetValidPrepTaskConfigsForIngredient(ctx, createdValidPrepTaskConfigs[0].Ingredient.ID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, forIngredient.Data)

	forPreparation, err := dbc.GetValidPrepTaskConfigsForPreparation(ctx, createdValidPrepTaskConfigs[0].Preparation.ID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, forPreparation.Data)

	forIngredientAndPreparation, err := dbc.GetValidPrepTaskConfigsForIngredientAndPreparation(ctx, createdValidPrepTaskConfigs[0].Ingredient.ID, createdValidPrepTaskConfigs[0].Preparation.ID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, forIngredientAndPreparation.Data)

	// delete
	for _, validPrepTaskConfig := range createdValidPrepTaskConfigs {
		require.NoError(t, dbc.ArchiveValidPrepTaskConfig(ctx, validPrepTaskConfig.ID))

		var exists bool
		exists, err = dbc.ValidPrepTaskConfigExists(ctx, validPrepTaskConfig.ID)
		require.NoError(t, err)
		assert.False(t, exists)

		var y *types.ValidPrepTaskConfig
		y, err = dbc.GetValidPrepTaskConfig(ctx, validPrepTaskConfig.ID)
		assert.Nil(t, y)
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_ValidPrepTaskConfigExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid prep task config MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.ValidPrepTaskConfigExists(ctx, "")
		require.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid prep task config MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidPrepTaskConfig(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetValidPrepTaskConfigsForIngredient(T *testing.T) {
	T.Parallel()

	T.Run("with invalid ingredient MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidPrepTaskConfigsForIngredient(ctx, "", nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetValidPrepTaskConfigsForPreparation(T *testing.T) {
	T.Parallel()

	T.Run("with invalid preparation MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidPrepTaskConfigsForPreparation(ctx, "", nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetValidPrepTaskConfigsForIngredientAndPreparation(T *testing.T) {
	T.Parallel()

	T.Run("with invalid ingredient MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidPrepTaskConfigsForIngredientAndPreparation(ctx, "", "valid-preparation-id", nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid preparation MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidPrepTaskConfigsForIngredientAndPreparation(ctx, "valid-ingredient-id", "", nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateValidPrepTaskConfig(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateValidPrepTaskConfig(ctx, nil))
	})
}

func TestQuerier_ArchiveValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid prep task config MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveValidPrepTaskConfig(ctx, ""))
	})
}

func TestQuerier_Integration_ValidPrepTaskConfigs_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.ValidPrepTaskConfig]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "valid prep task config",
		CreateItem: func(ctx context.Context, i int) *types.ValidPrepTaskConfig {
			// Create unique ingredient and preparation for each item
			exampleValidIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
			exampleValidPreparation := createValidPreparationForTest(t, ctx, nil, dbc)
			exampleValidPrepTaskConfig := fakes.BuildFakeValidPrepTaskConfig()
			exampleValidPrepTaskConfig.Ingredient = *exampleValidIngredient
			exampleValidPrepTaskConfig.Preparation = *exampleValidPreparation
			return createValidPrepTaskConfigForTest(t, ctx, exampleValidPrepTaskConfig, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPrepTaskConfig], error) {
			return dbc.GetValidPrepTaskConfigs(ctx, filter)
		},
		GetID: func(validPrepTaskConfig *types.ValidPrepTaskConfig) string {
			return validPrepTaskConfig.ID
		},
		CleanupItem: func(ctx context.Context, validPrepTaskConfig *types.ValidPrepTaskConfig) error {
			return dbc.ArchiveValidPrepTaskConfig(ctx, validPrepTaskConfig.ID)
		},
	})
}
