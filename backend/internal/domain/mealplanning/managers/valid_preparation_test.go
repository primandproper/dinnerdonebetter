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

func TestValidEnumerationManager_SearchValidPreparations(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationsList()
		exampleQuery := fakes.BuildFakeID()

		// media is looked up once per returned record.
		expectedIDs := map[string]bool{}
		for _, prep := range expected.Data {
			expectedIDs[prep.ID] = true
		}

		db := &mealplanningmock.RepositoryMock{
			SearchForValidPreparationsFunc: func(_ context.Context, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparation], error) {
				assert.Equal(t, exampleQuery, query)

				return expected, nil
			},
			GetPreparationMediaByPreparationFunc: func(_ context.Context, id string) ([]*types.PreparationMediaRow, error) {
				assert.True(t, expectedIDs[id], "unexpected media lookup for %s", id)

				return []*types.PreparationMediaRow{}, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidPreparations(ctx, exampleQuery, false, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForValidPreparationsCalls(), 1)
		assert.Len(t, db.GetPreparationMediaByPreparationCalls(), len(expected.Data))
	})
}

func TestValidEnumerationManager_ListValidPreparations(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationsList()

		// media is looked up once per returned record.
		expectedIDs := map[string]bool{}
		for _, prep := range expected.Data {
			expectedIDs[prep.ID] = true
		}

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparation], error) {
				return expected, nil
			},
			GetPreparationMediaByPreparationFunc: func(_ context.Context, id string) ([]*types.PreparationMediaRow, error) {
				assert.True(t, expectedIDs[id], "unexpected media lookup for %s", id)

				return []*types.PreparationMediaRow{}, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ListValidPreparations(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPreparationsCalls(), 1)
		assert.Len(t, db.GetPreparationMediaByPreparationCalls(), len(expected.Data))
	})
}

func TestValidEnumerationManager_CreateValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparation()
		fakeInput := fakes.BuildFakeValidPreparationCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateValidPreparationFunc: func(_ context.Context, _ *types.ValidPreparationDatabaseCreationInput) (*types.ValidPreparation, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.CreateValidPreparation(ctx, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateValidPreparationCalls(), 1)
	})
}

func TestValidEnumerationManager_ReadValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparation()

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationFunc: func(_ context.Context, validPreparationID string) (*types.ValidPreparation, error) {
				assert.Equal(t, expected.ID, validPreparationID)

				return expected, nil
			},
			GetPreparationMediaByPreparationFunc: func(_ context.Context, validPreparationID string) ([]*types.PreparationMediaRow, error) {
				assert.Equal(t, expected.ID, validPreparationID)

				return []*types.PreparationMediaRow{}, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ReadValidPreparation(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPreparationCalls(), 1)
		assert.Len(t, db.GetPreparationMediaByPreparationCalls(), 1)
	})
}

func TestValidEnumerationManager_RandomValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparation()

		db := &mealplanningmock.RepositoryMock{
			GetRandomValidPreparationFunc: func(_ context.Context) (*types.ValidPreparation, error) {
				return expected, nil
			},
			GetPreparationMediaByPreparationFunc: func(_ context.Context, validPreparationID string) ([]*types.PreparationMediaRow, error) {
				assert.Equal(t, expected.ID, validPreparationID)

				return []*types.PreparationMediaRow{}, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.RandomValidPreparation(ctx)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRandomValidPreparationCalls(), 1)
		assert.Len(t, db.GetPreparationMediaByPreparationCalls(), 1)
	})
}

func TestValidEnumerationManager_UpdateValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidPreparation := fakes.BuildFakeValidPreparation()
		exampleInput := fakes.BuildFakeValidPreparationUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationFunc: func(_ context.Context, validPreparationID string) (*types.ValidPreparation, error) {
				assert.Equal(t, exampleValidPreparation.ID, validPreparationID)

				return exampleValidPreparation, nil
			},
			UpdateValidPreparationFunc: func(_ context.Context, _ *types.ValidPreparation) error {
				return nil
			},
			GetPreparationMediaByPreparationFunc: func(_ context.Context, validPreparationID string) ([]*types.PreparationMediaRow, error) {
				assert.Equal(t, exampleValidPreparation.ID, validPreparationID)

				return []*types.PreparationMediaRow{}, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		result, err := mpm.UpdateValidPreparation(ctx, exampleValidPreparation.ID, exampleInput)
		assert.NotNil(t, result)
		require.NoError(t, err)

		assert.Len(t, db.GetValidPreparationCalls(), 2) // the manager re-reads the record after updating it
		assert.Len(t, db.UpdateValidPreparationCalls(), 1)
		assert.Len(t, db.GetPreparationMediaByPreparationCalls(), 1)
	})
}

func TestValidEnumerationManager_ArchiveValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparation()

		db := &mealplanningmock.RepositoryMock{
			ArchiveValidPreparationFunc: func(_ context.Context, validPreparationID string) error {
				assert.Equal(t, expected.ID, validPreparationID)

				return nil
			},
		}
		attachRepositoryToManager(vem, db)

		require.NoError(t, vem.ArchiveValidPreparation(ctx, expected.ID))

		assert.Len(t, db.ArchiveValidPreparationCalls(), 1)
	})
}
