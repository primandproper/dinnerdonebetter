package managers

import (
	"context"
	"errors"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	eatingindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v10/filtering"
	textsearch "github.com/primandproper/platform-go/v10/search/text"
	mocksearch "github.com/primandproper/platform-go/v10/search/text/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipeManager_ListRecipes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipesList()
		status := types.RecipeStatusSubmitted

		db := &mealplanningmock.RepositoryMock{
			GetRecipesFunc: func(_ context.Context, recipeStatus string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Recipe], error) {
				assert.Equal(t, status, recipeStatus)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ListRecipes(ctx, status, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipesCalls(), 1)
	})
}

func TestRecipeManager_CreateRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		fakeCreatorID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipe()
		fakeInput := fakes.BuildFakeRecipeCreationRequestInput()

		analyzer := &recipeanalysis.RecipeAnalyzerMock{
			ValidateRecipeCreationRequestInputIsDAGFunc: func(_ context.Context, _ *types.RecipeCreationRequestInput) error {
				return nil
			},
		}

		db := &mealplanningmock.RepositoryMock{
			CreateRecipeFunc: func(_ context.Context, _ *types.RecipeDatabaseCreationInput) (*types.Recipe, error) {
				return expected, nil
			},
			GetRecipeFunc: func(_ context.Context, recipeID string) (*types.Recipe, error) {
				assert.Equal(t, expected.ID, recipeID)

				return expected, nil
			},
		}
		attachRepositoryAndAnalyzerToManager(rm, db, analyzer)

		actual, err := rm.CreateRecipe(ctx, fakeCreatorID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateRecipeCalls(), 1)
		assert.Len(t, db.GetRecipeCalls(), 1)
		assert.Len(t, analyzer.ValidateRecipeCreationRequestInputIsDAGCalls(), 1)
	})

	T.Run("with DAG error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		fakeCreatorID := fakes.BuildFakeID()
		fakeInput := fakes.BuildFakeRecipeCreationRequestInput()

		analyzer := &recipeanalysis.RecipeAnalyzerMock{
			ValidateRecipeCreationRequestInputIsDAGFunc: func(_ context.Context, _ *types.RecipeCreationRequestInput) error {
				return errors.New("blah")
			},
		}

		db := &mealplanningmock.RepositoryMock{}
		attachRepositoryAndAnalyzerToManager(rm, db, analyzer)

		actual, err := rm.CreateRecipe(ctx, fakeCreatorID, fakeInput)
		require.Error(t, err)
		assert.Nil(t, actual)

		assert.Len(t, analyzer.ValidateRecipeCreationRequestInputIsDAGCalls(), 1)
		// a recipe that fails DAG validation is never persisted.
		assert.Empty(t, db.CreateRecipeCalls())
	})
}

func TestRecipeManager_ReadRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipe()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeFunc: func(_ context.Context, recipeID string) (*types.Recipe, error) {
				assert.Equal(t, expected.ID, recipeID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ReadRecipe(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeCalls(), 1)
	})
}

func TestRecipeManager_SearchRecipes(T *testing.T) {
	T.Parallel()

	T.Run("useSearchService false uses database", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipesList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			SearchForRecipesFunc: func(_ context.Context, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Recipe], error) {
				assert.Equal(t, exampleQuery, query)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.SearchRecipes(ctx, exampleQuery, false, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForRecipesCalls(), 1)
	})

	T.Run("useSearchService true falls back to database when search returns empty", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipesList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			SearchForRecipesFunc: func(_ context.Context, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Recipe], error) {
				assert.Equal(t, exampleQuery, query)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.SearchRecipes(ctx, exampleQuery, true, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForRecipesCalls(), 1)
	})

	T.Run("useSearchService true asks the index for the filter's page", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleQuery := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipe()

		cursor := "cursor-from-a-previous-page"
		filter := filtering.DefaultQueryFilter()
		filter.MaxResponseSize = new(uint16(11))
		filter.Cursor = &cursor

		db := &mealplanningmock.RepositoryMock{
			GetRecipesWithIDsFunc: func(_ context.Context, ids []string) ([]*types.Recipe, error) {
				assert.Equal(t, []string{expected.ID}, ids)

				return []*types.Recipe{expected}, nil
			},
		}
		attachRepositoryToManager(rm, db)

		index := &mocksearch.IndexMock[eatingindexing.RecipeSearchSubset]{
			SearchFunc: func(_ context.Context, req textsearch.SearchRequest) (*textsearch.SearchResults[eatingindexing.RecipeSearchSubset], error) {
				assert.Equal(t, exampleQuery, req.Query)
				assert.Equal(t, 11, req.Limit)
				assert.Equal(t, textsearch.Cursor(cursor), req.Cursor)

				return &textsearch.SearchResults[eatingindexing.RecipeSearchSubset]{
					Hits:       []*eatingindexing.RecipeSearchSubset{{ID: expected.ID}},
					NextCursor: textsearch.Cursor("cursor-for-the-next-page"),
				}, nil
			},
		}
		attachRecipeSearchIndexToManager(rm, index)

		actual, err := rm.SearchRecipes(ctx, exampleQuery, true, filter)
		require.NoError(t, err)
		assert.Equal(t, []*types.Recipe{expected}, actual.Data)
		assert.Equal(t, "cursor-for-the-next-page", actual.Cursor)

		assert.Len(t, index.SearchCalls(), 1)
		assert.Empty(t, db.SearchForRecipesCalls())
	})

	T.Run("useSearchService true falls back to the database without the index's cursor", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleQuery := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipesList()

		cursor := "an-opaque-index-token"
		filter := filtering.DefaultQueryFilter()
		filter.Cursor = &cursor

		db := &mealplanningmock.RepositoryMock{
			SearchForRecipesFunc: func(_ context.Context, query string, fallbackFilter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Recipe], error) {
				assert.Equal(t, exampleQuery, query)
				// The database reads a cursor as the last row's ID, so an index token
				// cannot come along: it would match an arbitrary slice of the table.
				assert.Nil(t, fallbackFilter.Cursor)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		index := &mocksearch.IndexMock[eatingindexing.RecipeSearchSubset]{
			SearchFunc: func(_ context.Context, _ textsearch.SearchRequest) (*textsearch.SearchResults[eatingindexing.RecipeSearchSubset], error) {
				return nil, errors.New("elasticsearch is down")
			},
		}
		attachRecipeSearchIndexToManager(rm, index)

		actual, err := rm.SearchRecipes(ctx, exampleQuery, true, filter)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForRecipesCalls(), 1)
	})

	T.Run("useSearchService true surfaces a rejected cursor instead of falling back", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleQuery := fakes.BuildFakeID()

		cursor := "a-cursor-from-a-different-backend"
		filter := filtering.DefaultQueryFilter()
		filter.Cursor = &cursor

		db := &mealplanningmock.RepositoryMock{}
		attachRepositoryToManager(rm, db)

		index := &mocksearch.IndexMock[eatingindexing.RecipeSearchSubset]{
			SearchFunc: func(_ context.Context, _ textsearch.SearchRequest) (*textsearch.SearchResults[eatingindexing.RecipeSearchSubset], error) {
				return nil, textsearch.ErrInvalidCursor
			},
		}
		attachRecipeSearchIndexToManager(rm, index)

		actual, err := rm.SearchRecipes(ctx, exampleQuery, true, filter)
		assert.Nil(t, actual)
		require.ErrorIs(t, err, textsearch.ErrInvalidCursor)

		// Restarting the database from the first page would answer a question the
		// caller did not ask, so the refusal reaches them instead.
		assert.Empty(t, db.SearchForRecipesCalls())
	})

	T.Run("useSearchService true ends pagination on an empty page rather than restarting", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleQuery := fakes.BuildFakeID()

		cursor := "cursor-for-the-last-page"
		filter := filtering.DefaultQueryFilter()
		filter.Cursor = &cursor

		db := &mealplanningmock.RepositoryMock{}
		attachRepositoryToManager(rm, db)

		index := &mocksearch.IndexMock[eatingindexing.RecipeSearchSubset]{
			SearchFunc: func(_ context.Context, _ textsearch.SearchRequest) (*textsearch.SearchResults[eatingindexing.RecipeSearchSubset], error) {
				return &textsearch.SearchResults[eatingindexing.RecipeSearchSubset]{}, nil
			},
		}
		attachRecipeSearchIndexToManager(rm, index)

		actual, err := rm.SearchRecipes(ctx, exampleQuery, true, filter)
		require.NoError(t, err)
		assert.Empty(t, actual.Data)
		assert.Empty(t, actual.Cursor)

		assert.Empty(t, db.SearchForRecipesCalls())
	})
}

func TestRecipeManager_UpdateRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipe := fakes.BuildFakeRecipe()
		exampleInput := fakes.BuildFakeRecipeUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeFunc: func(_ context.Context, recipeID string) (*types.Recipe, error) {
				assert.Equal(t, exampleRecipe.ID, recipeID)

				return exampleRecipe, nil
			},
			UpdateRecipeFunc: func(_ context.Context, _ *types.Recipe) error {
				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.UpdateRecipe(ctx, exampleRecipe.ID, exampleInput))

		assert.Len(t, db.GetRecipeCalls(), 1)
		assert.Len(t, db.UpdateRecipeCalls(), 1)
	})
}

func TestRecipeManager_ArchiveRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipe()
		exampleOwnerID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			ArchiveRecipeFunc: func(_ context.Context, recipeID, accountID string) error {
				assert.Equal(t, expected.ID, recipeID)
				assert.Equal(t, exampleOwnerID, accountID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.ArchiveRecipe(ctx, expected.ID, exampleOwnerID))

		assert.Len(t, db.ArchiveRecipeCalls(), 1)
	})
}

func TestRecipeManager_RecipeEstimatedPrepSteps(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipe := fakes.BuildFakeRecipe()
		expectedResults := fakes.BuildFakeMealPlanTaskDatabaseCreationInputs()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeFunc: func(_ context.Context, recipeID string) (*types.Recipe, error) {
				assert.Equal(t, exampleRecipe.ID, recipeID)

				return exampleRecipe, nil
			},
		}

		analyzer := &recipeanalysis.RecipeAnalyzerMock{
			GenerateMealPlanTasksForRecipeFunc: func(_ context.Context, mealPlanOptionID string, _ *types.Recipe) ([]*types.MealPlanTaskDatabaseCreationInput, error) {
				assert.Empty(t, mealPlanOptionID)

				return expectedResults, nil
			},
		}

		attachRepositoryAndAnalyzerToManager(rm, db, analyzer)

		results, err := rm.RecipeEstimatedPrepSteps(ctx, exampleRecipe.ID)
		require.NoError(t, err)

		assert.Len(t, results, len(expectedResults))

		assert.Len(t, db.GetRecipeCalls(), 1)
		assert.Len(t, analyzer.GenerateMealPlanTasksForRecipeCalls(), 1)
	})
}

func TestRecipeManager_MealMermaid(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleMeal := fakes.BuildFakeMeal()
		expectedResult := "flowchart TD;\n\tStep1[\"Main\"];\n"

		db := &mealplanningmock.RepositoryMock{}

		analyzer := &recipeanalysis.RecipeAnalyzerMock{
			RenderMermaidDiagramForMealFunc: func(_ context.Context, _ *types.Meal) string {
				return expectedResult
			},
		}

		attachRepositoryAndAnalyzerToManager(rm, db, analyzer)

		result, err := rm.MealMermaid(ctx, exampleMeal)
		require.NoError(t, err)
		assert.Equal(t, expectedResult, result)

		assert.Len(t, analyzer.RenderMermaidDiagramForMealCalls(), 1)
	})
}

func TestRecipeManager_RecipeMermaid(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipe := fakes.BuildFakeRecipe()
		expectedResult := t.Name()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeFunc: func(_ context.Context, recipeID string) (*types.Recipe, error) {
				assert.Equal(t, exampleRecipe.ID, recipeID)

				return exampleRecipe, nil
			},
		}

		analyzer := &recipeanalysis.RecipeAnalyzerMock{
			RenderMermaidDiagramForRecipeFunc: func(_ context.Context, _ *types.Recipe) string {
				return expectedResult
			},
		}

		attachRepositoryAndAnalyzerToManager(rm, db, analyzer)

		result, err := rm.RecipeMermaid(ctx, exampleRecipe.ID)
		require.NoError(t, err)
		assert.Equal(t, expectedResult, result)

		assert.Len(t, db.GetRecipeCalls(), 1)
		assert.Len(t, analyzer.RenderMermaidDiagramForRecipeCalls(), 1)
	})
}

func TestRecipeManager_CloneRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipe()
		cloned := fakes.BuildFakeRecipe()
		exampleOwnerID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeFunc: func(_ context.Context, recipeID string) (*types.Recipe, error) {
				assert.Equal(t, expected.ID, recipeID)

				return expected, nil
			},
			CreateRecipeFunc: func(_ context.Context, _ *types.RecipeDatabaseCreationInput) (*types.Recipe, error) {
				return cloned, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CloneRecipe(ctx, expected.ID, exampleOwnerID)
		require.NoError(t, err)
		assert.Equal(t, cloned, actual)

		assert.Len(t, db.GetRecipeCalls(), 1)
		assert.Len(t, db.CreateRecipeCalls(), 1)
	})
}
