package managers

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	eatingindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v10/filtering"
	textsearch "github.com/primandproper/platform-go/v10/search/text"
	"github.com/primandproper/platform-go/v10/search/text/elasticsearch"
	mocksearch "github.com/primandproper/platform-go/v10/search/text/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidEnumerationManager_SearchValidIngredients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientsList()
		exampleQuery := fakes.BuildFakeID()

		// media is looked up once per returned record.
		expectedIDs := map[string]bool{}
		for _, ing := range expected.Data {
			expectedIDs[ing.ID] = true
		}

		db := &mealplanningmock.RepositoryMock{
			SearchForValidIngredientsFunc: func(_ context.Context, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredient], error) {
				assert.Equal(t, exampleQuery, query)

				return expected, nil
			},
			GetIngredientMediaByIngredientFunc: func(_ context.Context, id string) ([]*types.IngredientMediaRow, error) {
				assert.True(t, expectedIDs[id], "unexpected media lookup for %s", id)

				return []*types.IngredientMediaRow{}, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidIngredients(ctx, exampleQuery, false, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForValidIngredientsCalls(), 1)
		assert.Len(t, db.GetIngredientMediaByIngredientCalls(), len(expected.Data))
	})

	T.Run("with search service asking the index for the filter's page", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		exampleQuery := fakes.BuildFakeID()
		expected := fakes.BuildFakeValidIngredient()

		cursor := "cursor-from-a-previous-page"
		filter := filtering.DefaultQueryFilter()
		filter.MaxResponseSize = new(uint16(11))
		filter.Cursor = &cursor

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientsWithIDsFunc: func(_ context.Context, ids []string) ([]*types.ValidIngredient, error) {
				assert.Equal(t, []string{expected.ID}, ids)

				return []*types.ValidIngredient{expected}, nil
			},
			GetIngredientMediaByIngredientFunc: func(_ context.Context, id string) ([]*types.IngredientMediaRow, error) {
				assert.Equal(t, expected.ID, id)

				return []*types.IngredientMediaRow{}, nil
			},
		}
		attachRepositoryToManager(vem, db)

		index := &mocksearch.IndexMock[eatingindexing.ValidIngredientSearchSubset]{
			SearchFunc: func(_ context.Context, req textsearch.SearchRequest) (*textsearch.SearchResults[eatingindexing.ValidIngredientSearchSubset], error) {
				assert.Equal(t, exampleQuery, req.Query)
				assert.Equal(t, 11, req.Limit)
				assert.Equal(t, textsearch.Cursor(cursor), req.Cursor)

				return &textsearch.SearchResults[eatingindexing.ValidIngredientSearchSubset]{
					Hits:       []*eatingindexing.ValidIngredientSearchSubset{{ID: expected.ID}},
					NextCursor: textsearch.Cursor("cursor-for-the-next-page"),
				}, nil
			},
		}
		attachValidIngredientSearchIndexToManager(vem, index)

		actual, err := vem.SearchValidIngredients(ctx, exampleQuery, true, filter)
		require.NoError(t, err)
		assert.Equal(t, []*types.ValidIngredient{expected}, actual.Data)
		assert.Equal(t, "cursor-for-the-next-page", actual.Cursor)

		assert.Len(t, index.SearchCalls(), 1)
		assert.Empty(t, db.SearchForValidIngredientsCalls())
	})

	T.Run("with search service refusing to page deeper", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		exampleQuery := fakes.BuildFakeID()

		cursor := "cursor-past-the-result-window"
		filter := filtering.DefaultQueryFilter()
		filter.Cursor = &cursor

		db := &mealplanningmock.RepositoryMock{}
		attachRepositoryToManager(vem, db)

		index := &mocksearch.IndexMock[eatingindexing.ValidIngredientSearchSubset]{
			SearchFunc: func(_ context.Context, _ textsearch.SearchRequest) (*textsearch.SearchResults[eatingindexing.ValidIngredientSearchSubset], error) {
				return nil, elasticsearch.ErrResultWindowExceeded
			},
		}
		attachValidIngredientSearchIndexToManager(vem, index)

		actual, err := vem.SearchValidIngredients(ctx, exampleQuery, true, filter)
		assert.Nil(t, actual)
		// The refusal survives wrapping so the service layer can turn it into a status
		// the client can act on, rather than a generic internal error.
		assert.ErrorIs(t, err, elasticsearch.ErrResultWindowExceeded)
	})
}

func TestValidEnumerationManager_ListValidIngredients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientsList()

		// media is looked up once per returned record.
		expectedIDs := map[string]bool{}
		for _, ing := range expected.Data {
			expectedIDs[ing.ID] = true
		}

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredient], error) {
				return expected, nil
			},
			GetIngredientMediaByIngredientFunc: func(_ context.Context, id string) ([]*types.IngredientMediaRow, error) {
				assert.True(t, expectedIDs[id], "unexpected media lookup for %s", id)

				return []*types.IngredientMediaRow{}, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ListValidIngredients(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientsCalls(), 1)
		assert.Len(t, db.GetIngredientMediaByIngredientCalls(), len(expected.Data))
	})
}

func TestValidEnumerationManager_CreateValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredient()
		fakeInput := fakes.BuildFakeValidIngredientCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateValidIngredientFunc: func(_ context.Context, _ *types.ValidIngredientDatabaseCreationInput) (*types.ValidIngredient, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.CreateValidIngredient(ctx, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateValidIngredientCalls(), 1)
	})
}

func TestValidEnumerationManager_ReadValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredient()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientFunc: func(_ context.Context, validIngredientID string) (*types.ValidIngredient, error) {
				assert.Equal(t, expected.ID, validIngredientID)

				return expected, nil
			},
			GetIngredientMediaByIngredientFunc: func(_ context.Context, validIngredientID string) ([]*types.IngredientMediaRow, error) {
				assert.Equal(t, expected.ID, validIngredientID)

				return []*types.IngredientMediaRow{}, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ReadValidIngredient(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientCalls(), 1)
		assert.Len(t, db.GetIngredientMediaByIngredientCalls(), 1)
	})
}

func TestValidEnumerationManager_RandomValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredient()

		db := &mealplanningmock.RepositoryMock{
			GetRandomValidIngredientFunc: func(_ context.Context) (*types.ValidIngredient, error) {
				return expected, nil
			},
			GetIngredientMediaByIngredientFunc: func(_ context.Context, validIngredientID string) ([]*types.IngredientMediaRow, error) {
				assert.Equal(t, expected.ID, validIngredientID)

				return []*types.IngredientMediaRow{}, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.RandomValidIngredient(ctx)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRandomValidIngredientCalls(), 1)
		assert.Len(t, db.GetIngredientMediaByIngredientCalls(), 1)
	})
}

func TestValidEnumerationManager_UpdateValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidIngredient := fakes.BuildFakeValidIngredient()
		exampleInput := fakes.BuildFakeValidIngredientUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientFunc: func(_ context.Context, validIngredientID string) (*types.ValidIngredient, error) {
				assert.Equal(t, exampleValidIngredient.ID, validIngredientID)

				return exampleValidIngredient, nil
			},
			UpdateValidIngredientFunc: func(_ context.Context, _ *types.ValidIngredient) error {
				return nil
			},
			GetIngredientMediaByIngredientFunc: func(_ context.Context, validIngredientID string) ([]*types.IngredientMediaRow, error) {
				assert.Equal(t, exampleValidIngredient.ID, validIngredientID)

				return []*types.IngredientMediaRow{}, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		result, err := mpm.UpdateValidIngredient(ctx, exampleValidIngredient.ID, exampleInput)
		assert.NotNil(t, result)
		require.NoError(t, err)

		assert.Len(t, db.GetValidIngredientCalls(), 2) // the manager re-reads the record after updating it
		assert.Len(t, db.UpdateValidIngredientCalls(), 1)
		assert.Len(t, db.GetIngredientMediaByIngredientCalls(), 1)
	})
}

func TestValidEnumerationManager_ArchiveValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredient()

		db := &mealplanningmock.RepositoryMock{
			ArchiveValidIngredientFunc: func(_ context.Context, validIngredientID string) error {
				assert.Equal(t, expected.ID, validIngredientID)

				return nil
			},
		}
		attachRepositoryToManager(vem, db)

		require.NoError(t, vem.ArchiveValidIngredient(ctx, expected.ID))

		assert.Len(t, db.ArchiveValidIngredientCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidIngredientsByPreparationAndIngredientName(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientsList()
		preparationID := fakes.BuildFakeID()
		exampleQuery := fakes.BuildFakeID()

		// media is looked up once per returned record.
		expectedIDs := map[string]bool{}
		for _, ing := range expected.Data {
			expectedIDs[ing.ID] = true
		}

		db := &mealplanningmock.RepositoryMock{
			SearchForValidIngredientsForPreparationFunc: func(_ context.Context, prepID string, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredient], error) {
				assert.Equal(t, preparationID, prepID)
				assert.Equal(t, exampleQuery, query)

				return expected, nil
			},
			GetIngredientMediaByIngredientFunc: func(_ context.Context, id string) ([]*types.IngredientMediaRow, error) {
				assert.True(t, expectedIDs[id], "unexpected media lookup for %s", id)

				return []*types.IngredientMediaRow{}, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidIngredientsByPreparationAndIngredientName(ctx, preparationID, exampleQuery, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForValidIngredientsForPreparationCalls(), 1)
		assert.Len(t, db.GetIngredientMediaByIngredientCalls(), len(expected.Data))
	})
}
