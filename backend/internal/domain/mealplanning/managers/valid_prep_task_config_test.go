package managers

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidEnumerationManager_ListValidPrepTaskConfigs(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPrepTaskConfigsList()

		db := &mealplanningmock.RepositoryMock{
			GetValidPrepTaskConfigsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPrepTaskConfig], error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ListValidPrepTaskConfigs(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPrepTaskConfigsCalls(), 1)
	})
}

func TestValidEnumerationManager_CreateValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPrepTaskConfig()
		fakeInput := fakes.BuildFakeValidPrepTaskConfigCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateValidPrepTaskConfigFunc: func(_ context.Context, _ *types.ValidPrepTaskConfigDatabaseCreationInput) (*types.ValidPrepTaskConfig, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.CreateValidPrepTaskConfig(ctx, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateValidPrepTaskConfigCalls(), 1)
	})
}

func TestValidEnumerationManager_ReadValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPrepTaskConfig()

		db := &mealplanningmock.RepositoryMock{
			GetValidPrepTaskConfigFunc: func(_ context.Context, validIngredientPreparationStorageConfigID string) (*types.ValidPrepTaskConfig, error) {
				assert.Equal(t, expected.ID, validIngredientPreparationStorageConfigID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ReadValidPrepTaskConfig(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPrepTaskConfigCalls(), 1)
	})
}

func TestValidEnumerationManager_UpdateValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidPrepTaskConfig := fakes.BuildFakeValidPrepTaskConfig()
		exampleInput := fakes.BuildFakeValidPrepTaskConfigUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetValidPrepTaskConfigFunc: func(_ context.Context, validIngredientPreparationStorageConfigID string) (*types.ValidPrepTaskConfig, error) {
				assert.Equal(t, exampleValidPrepTaskConfig.ID, validIngredientPreparationStorageConfigID)

				return exampleValidPrepTaskConfig, nil
			},
			UpdateValidPrepTaskConfigFunc: func(_ context.Context, _ *types.ValidPrepTaskConfig) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		result, err := mpm.UpdateValidPrepTaskConfig(ctx, exampleValidPrepTaskConfig.ID, exampleInput)
		assert.NotNil(t, result)
		require.NoError(t, err)

		assert.Len(t, db.GetValidPrepTaskConfigCalls(), 2) // the manager re-reads the record after updating it
		assert.Len(t, db.UpdateValidPrepTaskConfigCalls(), 1)
	})
}

func TestValidEnumerationManager_ArchiveValidPrepTaskConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPrepTaskConfig()

		db := &mealplanningmock.RepositoryMock{
			ArchiveValidPrepTaskConfigFunc: func(_ context.Context, validIngredientPreparationStorageConfigID string) error {
				assert.Equal(t, expected.ID, validIngredientPreparationStorageConfigID)

				return nil
			},
		}
		attachRepositoryToManager(vem, db)

		require.NoError(t, vem.ArchiveValidPrepTaskConfig(ctx, expected.ID))

		assert.Len(t, db.ArchiveValidPrepTaskConfigCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidPrepTaskConfigsByIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPrepTaskConfigsList()
		exampleIngredientID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetValidPrepTaskConfigsForIngredientFunc: func(_ context.Context, ingredientID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPrepTaskConfig], error) {
				assert.Equal(t, exampleIngredientID, ingredientID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidPrepTaskConfigsByIngredient(ctx, exampleIngredientID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPrepTaskConfigsForIngredientCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidPrepTaskConfigsByPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPrepTaskConfigsList()
		examplePreparationID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetValidPrepTaskConfigsForPreparationFunc: func(_ context.Context, preparationID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPrepTaskConfig], error) {
				assert.Equal(t, examplePreparationID, preparationID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidPrepTaskConfigsByPreparation(ctx, examplePreparationID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPrepTaskConfigsForPreparationCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidPrepTaskConfigsByIngredientAndPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPrepTaskConfigsList()
		exampleIngredientID := fakes.BuildFakeID()
		examplePreparationID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetValidPrepTaskConfigsForIngredientAndPreparationFunc: func(_ context.Context, ingredientID string, preparationID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPrepTaskConfig], error) {
				assert.Equal(t, exampleIngredientID, ingredientID)
				assert.Equal(t, examplePreparationID, preparationID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidPrepTaskConfigsByIngredientAndPreparation(ctx, exampleIngredientID, examplePreparationID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPrepTaskConfigsForIngredientAndPreparationCalls(), 1)
	})
}
