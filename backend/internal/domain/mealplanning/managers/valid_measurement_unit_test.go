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

func TestValidEnumerationManager_SearchValidMeasurementUnits(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidMeasurementUnitsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			SearchForValidMeasurementUnitsFunc: func(_ context.Context, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnit], error) {
				assert.Equal(t, exampleQuery, query)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidMeasurementUnits(ctx, exampleQuery, false, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForValidMeasurementUnitsCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidMeasurementUnitsByIngredientID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidMeasurementUnitsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			ValidMeasurementUnitsForIngredientIDFunc: func(_ context.Context, validIngredientID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnit], error) {
				assert.Equal(t, exampleQuery, validIngredientID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidMeasurementUnitsByIngredientID(ctx, exampleQuery, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.ValidMeasurementUnitsForIngredientIDCalls(), 1)
	})
}

func TestValidEnumerationManager_ListValidMeasurementUnits(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidMeasurementUnitsList()

		db := &mealplanningmock.RepositoryMock{
			GetValidMeasurementUnitsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnit], error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ListValidMeasurementUnits(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidMeasurementUnitsCalls(), 1)
	})
}

func TestValidEnumerationManager_CreateValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidMeasurementUnit()
		fakeInput := fakes.BuildFakeValidMeasurementUnitCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateValidMeasurementUnitFunc: func(_ context.Context, _ *types.ValidMeasurementUnitDatabaseCreationInput) (*types.ValidMeasurementUnit, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.CreateValidMeasurementUnit(ctx, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateValidMeasurementUnitCalls(), 1)
	})
}

func TestValidEnumerationManager_ReadValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidMeasurementUnit()

		db := &mealplanningmock.RepositoryMock{
			GetValidMeasurementUnitFunc: func(_ context.Context, validMeasurementUnitID string) (*types.ValidMeasurementUnit, error) {
				assert.Equal(t, expected.ID, validMeasurementUnitID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ReadValidMeasurementUnit(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidMeasurementUnitCalls(), 1)
	})
}

func TestValidEnumerationManager_UpdateValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidMeasurementUnit := fakes.BuildFakeValidMeasurementUnit()
		exampleInput := fakes.BuildFakeValidMeasurementUnitUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetValidMeasurementUnitFunc: func(_ context.Context, validMeasurementUnitID string) (*types.ValidMeasurementUnit, error) {
				assert.Equal(t, exampleValidMeasurementUnit.ID, validMeasurementUnitID)

				return exampleValidMeasurementUnit, nil
			},
			UpdateValidMeasurementUnitFunc: func(_ context.Context, _ *types.ValidMeasurementUnit) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		result, err := mpm.UpdateValidMeasurementUnit(ctx, exampleValidMeasurementUnit.ID, exampleInput)
		assert.NotNil(t, result)
		require.NoError(t, err)

		assert.Len(t, db.GetValidMeasurementUnitCalls(), 2) // the manager re-reads the record after updating it
		assert.Len(t, db.UpdateValidMeasurementUnitCalls(), 1)
	})
}

func TestValidEnumerationManager_ArchiveValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidMeasurementUnit()

		db := &mealplanningmock.RepositoryMock{
			ArchiveValidMeasurementUnitFunc: func(_ context.Context, validMeasurementUnitID string) error {
				assert.Equal(t, expected.ID, validMeasurementUnitID)

				return nil
			},
		}
		attachRepositoryToManager(vem, db)

		require.NoError(t, vem.ArchiveValidMeasurementUnit(ctx, expected.ID))

		assert.Len(t, db.ArchiveValidMeasurementUnitCalls(), 1)
	})
}
