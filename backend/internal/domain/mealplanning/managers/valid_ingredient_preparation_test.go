package managers

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v7/filtering"

	"github.com/stretchr/testify/assert"
)

func TestValidEnumerationManager_ListValidIngredientPreparations(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientPreparationsList()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientPreparationsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientPreparation], error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ListValidIngredientPreparations(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientPreparationsCalls(), 1)
	})
}

func TestValidEnumerationManager_CreateValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientPreparation()
		fakeInput := fakes.BuildFakeValidIngredientPreparationCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateValidIngredientPreparationFunc: func(_ context.Context, _ *types.ValidIngredientPreparationDatabaseCreationInput) (*types.ValidIngredientPreparation, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.CreateValidIngredientPreparation(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateValidIngredientPreparationCalls(), 1)
	})
}

func TestValidEnumerationManager_ReadValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientPreparation()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientPreparationFunc: func(_ context.Context, validIngredientPreparationID string) (*types.ValidIngredientPreparation, error) {
				assert.Equal(t, expected.ID, validIngredientPreparationID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ReadValidIngredientPreparation(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientPreparationCalls(), 1)
	})
}

func TestValidEnumerationManager_UpdateValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidIngredientPreparation := fakes.BuildFakeValidIngredientPreparation()
		exampleInput := fakes.BuildFakeValidIngredientPreparationUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientPreparationFunc: func(_ context.Context, validIngredientPreparationID string) (*types.ValidIngredientPreparation, error) {
				assert.Equal(t, exampleValidIngredientPreparation.ID, validIngredientPreparationID)

				return exampleValidIngredientPreparation, nil
			},
			UpdateValidIngredientPreparationFunc: func(_ context.Context, _ *types.ValidIngredientPreparation) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		result, err := mpm.UpdateValidIngredientPreparation(ctx, exampleValidIngredientPreparation.ID, exampleInput)
		assert.NotNil(t, result)
		assert.NoError(t, err)

		assert.Len(t, db.GetValidIngredientPreparationCalls(), 2) // the manager re-reads the record after updating it
		assert.Len(t, db.UpdateValidIngredientPreparationCalls(), 1)
	})
}

func TestValidEnumerationManager_ArchiveValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientPreparation()

		db := &mealplanningmock.RepositoryMock{
			ArchiveValidIngredientPreparationFunc: func(_ context.Context, validIngredientPreparationID string) error {
				assert.Equal(t, expected.ID, validIngredientPreparationID)

				return nil
			},
		}
		attachRepositoryToManager(vem, db)

		assert.NoError(t, vem.ArchiveValidIngredientPreparation(ctx, expected.ID))

		assert.Len(t, db.ArchiveValidIngredientPreparationCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidIngredientPreparationsByIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientPreparationsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientPreparationsForIngredientFunc: func(_ context.Context, ingredientID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientPreparation], error) {
				assert.Equal(t, exampleQuery, ingredientID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidIngredientPreparationsByIngredient(ctx, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientPreparationsForIngredientCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidIngredientPreparationsByPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientPreparationsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientPreparationsForPreparationFunc: func(_ context.Context, preparationID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientPreparation], error) {
				assert.Equal(t, exampleQuery, preparationID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidIngredientPreparationsByPreparation(ctx, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientPreparationsForPreparationCalls(), 1)
	})
}
