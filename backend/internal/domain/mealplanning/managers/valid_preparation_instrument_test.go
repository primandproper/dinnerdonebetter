package managers

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v6/filtering"

	"github.com/stretchr/testify/assert"
)

func TestValidEnumerationManager_ListValidPreparationInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationInstrumentsList()

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationInstrumentsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationInstrument], error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ListValidPreparationInstruments(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPreparationInstrumentsCalls(), 1)
	})
}

func TestValidEnumerationManager_CreateValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationInstrument()
		fakeInput := fakes.BuildFakeValidPreparationInstrumentCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateValidPreparationInstrumentFunc: func(_ context.Context, _ *types.ValidPreparationInstrumentDatabaseCreationInput) (*types.ValidPreparationInstrument, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.CreateValidPreparationInstrument(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateValidPreparationInstrumentCalls(), 1)
	})
}

func TestValidEnumerationManager_ReadValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationInstrument()

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationInstrumentFunc: func(_ context.Context, validPreparationInstrumentID string) (*types.ValidPreparationInstrument, error) {
				assert.Equal(t, expected.ID, validPreparationInstrumentID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ReadValidPreparationInstrument(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPreparationInstrumentCalls(), 1)
	})
}

func TestValidEnumerationManager_UpdateValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidPreparationInstrument := fakes.BuildFakeValidPreparationInstrument()
		exampleInput := fakes.BuildFakeValidPreparationInstrumentUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationInstrumentFunc: func(_ context.Context, validPreparationInstrumentID string) (*types.ValidPreparationInstrument, error) {
				assert.Equal(t, exampleValidPreparationInstrument.ID, validPreparationInstrumentID)

				return exampleValidPreparationInstrument, nil
			},
			UpdateValidPreparationInstrumentFunc: func(_ context.Context, _ *types.ValidPreparationInstrument) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		result, err := mpm.UpdateValidPreparationInstrument(ctx, exampleValidPreparationInstrument.ID, exampleInput)
		assert.NotNil(t, result)
		assert.NoError(t, err)

		assert.Len(t, db.GetValidPreparationInstrumentCalls(), 2) // the manager re-reads the record after updating it
		assert.Len(t, db.UpdateValidPreparationInstrumentCalls(), 1)
	})
}

func TestValidEnumerationManager_ArchiveValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationInstrument()

		db := &mealplanningmock.RepositoryMock{
			ArchiveValidPreparationInstrumentFunc: func(_ context.Context, validPreparationInstrumentID string) error {
				assert.Equal(t, expected.ID, validPreparationInstrumentID)

				return nil
			},
		}
		attachRepositoryToManager(vem, db)

		assert.NoError(t, vem.ArchiveValidPreparationInstrument(ctx, expected.ID))

		assert.Len(t, db.ArchiveValidPreparationInstrumentCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidPreparationInstrumentsByPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationInstrumentsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationInstrumentsForPreparationFunc: func(_ context.Context, preparationID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationInstrument], error) {
				assert.Equal(t, exampleQuery, preparationID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidPreparationInstrumentsByPreparation(ctx, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPreparationInstrumentsForPreparationCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidPreparationInstrumentsByInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationInstrumentsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationInstrumentsForInstrumentFunc: func(_ context.Context, instrumentID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationInstrument], error) {
				assert.Equal(t, exampleQuery, instrumentID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidPreparationInstrumentsByInstrument(ctx, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPreparationInstrumentsForInstrumentCalls(), 1)
	})
}
