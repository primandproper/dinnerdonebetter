package managers

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v8/filtering"

	"github.com/stretchr/testify/assert"
)

func TestValidEnumerationManager_SearchValidInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidInstrumentsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			SearchForValidInstrumentsFunc: func(_ context.Context, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidInstrument], error) {
				assert.Equal(t, exampleQuery, query)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidInstruments(ctx, exampleQuery, false, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForValidInstrumentsCalls(), 1)
	})
}

func TestValidEnumerationManager_ListValidInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidInstrumentsList()

		db := &mealplanningmock.RepositoryMock{
			GetValidInstrumentsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidInstrument], error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ListValidInstruments(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidInstrumentsCalls(), 1)
	})
}

func TestValidEnumerationManager_CreateValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidInstrument()
		fakeInput := fakes.BuildFakeValidInstrumentCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateValidInstrumentFunc: func(_ context.Context, _ *types.ValidInstrumentDatabaseCreationInput) (*types.ValidInstrument, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.CreateValidInstrument(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateValidInstrumentCalls(), 1)
	})
}

func TestValidEnumerationManager_ReadValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidInstrument()

		db := &mealplanningmock.RepositoryMock{
			GetValidInstrumentFunc: func(_ context.Context, validInstrumentID string) (*types.ValidInstrument, error) {
				assert.Equal(t, expected.ID, validInstrumentID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ReadValidInstrument(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidInstrumentCalls(), 1)
	})
}

func TestValidEnumerationManager_RandomValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidInstrument()

		db := &mealplanningmock.RepositoryMock{
			GetRandomValidInstrumentFunc: func(_ context.Context) (*types.ValidInstrument, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.RandomValidInstrument(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRandomValidInstrumentCalls(), 1)
	})
}

func TestValidEnumerationManager_UpdateValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidInstrument := fakes.BuildFakeValidInstrument()
		exampleInput := fakes.BuildFakeValidInstrumentUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetValidInstrumentFunc: func(_ context.Context, validInstrumentID string) (*types.ValidInstrument, error) {
				assert.Equal(t, exampleValidInstrument.ID, validInstrumentID)

				return exampleValidInstrument, nil
			},
			UpdateValidInstrumentFunc: func(_ context.Context, _ *types.ValidInstrument) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		result, err := mpm.UpdateValidInstrument(ctx, exampleValidInstrument.ID, exampleInput)
		assert.NotNil(t, result)
		assert.NoError(t, err)

		assert.Len(t, db.GetValidInstrumentCalls(), 2) // the manager re-reads the record after updating it
		assert.Len(t, db.UpdateValidInstrumentCalls(), 1)
	})
}

func TestValidEnumerationManager_ArchiveValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidInstrument()

		db := &mealplanningmock.RepositoryMock{
			ArchiveValidInstrumentFunc: func(_ context.Context, validInstrumentID string) error {
				assert.Equal(t, expected.ID, validInstrumentID)

				return nil
			},
		}
		attachRepositoryToManager(vem, db)

		assert.NoError(t, vem.ArchiveValidInstrument(ctx, expected.ID))

		assert.Len(t, db.ArchiveValidInstrumentCalls(), 1)
	})
}
