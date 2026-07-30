package managers

import (
	"context"
	"errors"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"

	"github.com/primandproper/platform-go/v8/filtering"

	"github.com/stretchr/testify/assert"
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
		assert.NoError(t, err)
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
		assert.NoError(t, err)
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
		assert.Error(t, err)
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
		assert.NoError(t, err)
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
		assert.NoError(t, err)
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
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForRecipesCalls(), 1)
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

		assert.NoError(t, rm.UpdateRecipe(ctx, exampleRecipe.ID, exampleInput))

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

		assert.NoError(t, rm.ArchiveRecipe(ctx, expected.ID, exampleOwnerID))

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
		assert.NoError(t, err)

		assert.Equal(t, len(results), len(expectedResults))

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
		assert.NoError(t, err)
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
		assert.NoError(t, err)
		assert.Equal(t, result, expectedResult)

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
		assert.NoError(t, err)
		assert.Equal(t, cloned, actual)

		assert.Len(t, db.GetRecipeCalls(), 1)
		assert.Len(t, db.CreateRecipeCalls(), 1)
	})
}
