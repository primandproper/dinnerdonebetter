package managers

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v11/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidEnumerationManager_ListValidIngredientStateIngredients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientStateIngredientsList()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientStateIngredientsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientStateIngredient], error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ListValidIngredientStateIngredients(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientStateIngredientsCalls(), 1)
	})
}

func TestValidEnumerationManager_CreateValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientStateIngredient()
		fakeInput := fakes.BuildFakeValidIngredientStateIngredientCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateValidIngredientStateIngredientFunc: func(_ context.Context, _ *types.ValidIngredientStateIngredientDatabaseCreationInput) (*types.ValidIngredientStateIngredient, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.CreateValidIngredientStateIngredient(ctx, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateValidIngredientStateIngredientCalls(), 1)
	})
}

func TestValidEnumerationManager_ReadValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientStateIngredient()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientStateIngredientFunc: func(_ context.Context, validIngredientPreparationID string) (*types.ValidIngredientStateIngredient, error) {
				assert.Equal(t, expected.ID, validIngredientPreparationID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ReadValidIngredientStateIngredient(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientStateIngredientCalls(), 1)
	})
}

func TestValidEnumerationManager_UpdateValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidIngredientStateIngredient := fakes.BuildFakeValidIngredientStateIngredient()
		exampleInput := fakes.BuildFakeValidIngredientStateIngredientUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientStateIngredientFunc: func(_ context.Context, validIngredientPreparationID string) (*types.ValidIngredientStateIngredient, error) {
				assert.Equal(t, exampleValidIngredientStateIngredient.ID, validIngredientPreparationID)

				return exampleValidIngredientStateIngredient, nil
			},
			UpdateValidIngredientStateIngredientFunc: func(_ context.Context, _ *types.ValidIngredientStateIngredient) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		result, err := mpm.UpdateValidIngredientStateIngredient(ctx, exampleValidIngredientStateIngredient.ID, exampleInput)
		assert.NotNil(t, result)
		require.NoError(t, err)

		assert.Len(t, db.GetValidIngredientStateIngredientCalls(), 2) // the manager re-reads the record after updating it
		assert.Len(t, db.UpdateValidIngredientStateIngredientCalls(), 1)
	})
}

func TestValidEnumerationManager_ArchiveValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientStateIngredient()

		db := &mealplanningmock.RepositoryMock{
			ArchiveValidIngredientStateIngredientFunc: func(_ context.Context, validIngredientPreparationID string) error {
				assert.Equal(t, expected.ID, validIngredientPreparationID)

				return nil
			},
		}
		attachRepositoryToManager(vem, db)

		require.NoError(t, vem.ArchiveValidIngredientStateIngredient(ctx, expected.ID))

		assert.Len(t, db.ArchiveValidIngredientStateIngredientCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidIngredientStateIngredientsByIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientStateIngredientsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientStateIngredientsForIngredientFunc: func(_ context.Context, ingredientID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientStateIngredient], error) {
				assert.Equal(t, exampleQuery, ingredientID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidIngredientStateIngredientsByIngredient(ctx, exampleQuery, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientStateIngredientsForIngredientCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidIngredientStateIngredientsByIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientStateIngredientsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientStateIngredientsForIngredientStateFunc: func(_ context.Context, ingredientStateID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientStateIngredient], error) {
				assert.Equal(t, exampleQuery, ingredientStateID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidIngredientStateIngredientsByIngredientState(ctx, exampleQuery, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientStateIngredientsForIngredientStateCalls(), 1)
	})
}
