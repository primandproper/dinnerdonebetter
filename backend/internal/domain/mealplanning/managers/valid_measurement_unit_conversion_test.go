package managers

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/stretchr/testify/assert"
)

func TestValidEnumerationManager_ValidMeasurementUnitConversionsForMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidMeasurementUnitConversionsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetValidMeasurementUnitConversionsForUnitFunc: func(_ context.Context, validMeasurementUnitID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnitConversion], error) {
				assert.Equal(t, exampleQuery, validMeasurementUnitID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ValidMeasurementUnitConversionsForMeasurementUnit(ctx, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidMeasurementUnitConversionsForUnitCalls(), 1)
	})
}

func TestValidEnumerationManager_CreateValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidMeasurementUnitConversion()
		fakeInput := fakes.BuildFakeValidMeasurementUnitConversionCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateValidMeasurementUnitConversionFunc: func(_ context.Context, _ *types.ValidMeasurementUnitConversionDatabaseCreationInput) (*types.ValidMeasurementUnitConversion, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.CreateValidMeasurementUnitConversion(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateValidMeasurementUnitConversionCalls(), 1)
	})
}

func TestValidEnumerationManager_ReadValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidMeasurementUnitConversion()

		db := &mealplanningmock.RepositoryMock{
			GetValidMeasurementUnitConversionFunc: func(_ context.Context, validMeasurementUnitConversionID string) (*types.ValidMeasurementUnitConversion, error) {
				assert.Equal(t, expected.ID, validMeasurementUnitConversionID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ReadValidMeasurementUnitConversion(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidMeasurementUnitConversionCalls(), 1)
	})
}

func TestValidEnumerationManager_UpdateValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidMeasurementUnitConversion := fakes.BuildFakeValidMeasurementUnitConversion()
		exampleInput := fakes.BuildFakeValidMeasurementUnitConversionUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetValidMeasurementUnitConversionFunc: func(_ context.Context, validMeasurementUnitConversionID string) (*types.ValidMeasurementUnitConversion, error) {
				assert.Equal(t, exampleValidMeasurementUnitConversion.ID, validMeasurementUnitConversionID)

				return exampleValidMeasurementUnitConversion, nil
			},
			UpdateValidMeasurementUnitConversionFunc: func(_ context.Context, _ *types.ValidMeasurementUnitConversion) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		result, err := mpm.UpdateValidMeasurementUnitConversion(ctx, exampleValidMeasurementUnitConversion.ID, exampleInput)
		assert.NotNil(t, result)
		assert.NoError(t, err)

		assert.Len(t, db.GetValidMeasurementUnitConversionCalls(), 2) // the manager re-reads the record after updating it
		assert.Len(t, db.UpdateValidMeasurementUnitConversionCalls(), 1)
	})
}

func TestValidEnumerationManager_ArchiveValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidMeasurementUnitConversion()

		db := &mealplanningmock.RepositoryMock{
			ArchiveValidMeasurementUnitConversionFunc: func(_ context.Context, validMeasurementUnitConversionID string) error {
				assert.Equal(t, expected.ID, validMeasurementUnitConversionID)

				return nil
			},
		}
		attachRepositoryToManager(vem, db)

		assert.NoError(t, vem.ArchiveValidMeasurementUnitConversion(ctx, expected.ID))

		assert.Len(t, db.ArchiveValidMeasurementUnitConversionCalls(), 1)
	})
}
