package managers

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v10/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidEnumerationManager_ListValidIngredientMeasurementUnits(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientMeasurementUnitsList()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientMeasurementUnitsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientMeasurementUnit], error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ListValidIngredientMeasurementUnits(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientMeasurementUnitsCalls(), 1)
	})
}

func TestValidEnumerationManager_CreateValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientMeasurementUnit()
		fakeInput := fakes.BuildFakeValidIngredientMeasurementUnitCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateValidIngredientMeasurementUnitFunc: func(_ context.Context, _ *types.ValidIngredientMeasurementUnitDatabaseCreationInput) (*types.ValidIngredientMeasurementUnit, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.CreateValidIngredientMeasurementUnit(ctx, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateValidIngredientMeasurementUnitCalls(), 1)
	})
}

func TestValidEnumerationManager_ReadValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientMeasurementUnit()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientMeasurementUnitFunc: func(_ context.Context, validIngredientMeasurementUnitID string) (*types.ValidIngredientMeasurementUnit, error) {
				assert.Equal(t, expected.ID, validIngredientMeasurementUnitID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ReadValidIngredientMeasurementUnit(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientMeasurementUnitCalls(), 1)
	})
}

func TestValidEnumerationManager_UpdateValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidIngredientMeasurementUnit := fakes.BuildFakeValidIngredientMeasurementUnit()
		exampleInput := fakes.BuildFakeValidIngredientMeasurementUnitUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientMeasurementUnitFunc: func(_ context.Context, validIngredientMeasurementUnitID string) (*types.ValidIngredientMeasurementUnit, error) {
				assert.Equal(t, exampleValidIngredientMeasurementUnit.ID, validIngredientMeasurementUnitID)

				return exampleValidIngredientMeasurementUnit, nil
			},
			UpdateValidIngredientMeasurementUnitFunc: func(_ context.Context, _ *types.ValidIngredientMeasurementUnit) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		result, err := mpm.UpdateValidIngredientMeasurementUnit(ctx, exampleValidIngredientMeasurementUnit.ID, exampleInput)
		assert.NotNil(t, result)
		require.NoError(t, err)

		assert.Len(t, db.GetValidIngredientMeasurementUnitCalls(), 2) // the manager re-reads the record after updating it
		assert.Len(t, db.UpdateValidIngredientMeasurementUnitCalls(), 1)
	})
}

func TestValidEnumerationManager_ArchiveValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientMeasurementUnit()

		db := &mealplanningmock.RepositoryMock{
			ArchiveValidIngredientMeasurementUnitFunc: func(_ context.Context, validIngredientMeasurementUnitID string) error {
				assert.Equal(t, expected.ID, validIngredientMeasurementUnitID)

				return nil
			},
		}
		attachRepositoryToManager(vem, db)

		require.NoError(t, vem.ArchiveValidIngredientMeasurementUnit(ctx, expected.ID))

		assert.Len(t, db.ArchiveValidIngredientMeasurementUnitCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidIngredientMeasurementUnitsByIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientMeasurementUnitsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientMeasurementUnitsForIngredientFunc: func(_ context.Context, ingredientID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientMeasurementUnit], error) {
				assert.Equal(t, exampleQuery, ingredientID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidIngredientMeasurementUnitsByIngredient(ctx, exampleQuery, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientMeasurementUnitsForIngredientCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidIngredientMeasurementUnitsByMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientMeasurementUnitsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientMeasurementUnitsForMeasurementUnitFunc: func(_ context.Context, ingredientID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientMeasurementUnit], error) {
				assert.Equal(t, exampleQuery, ingredientID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidIngredientMeasurementUnitsByMeasurementUnit(ctx, exampleQuery, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientMeasurementUnitsForMeasurementUnitCalls(), 1)
	})
}
