package managers

import (
	"errors"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRecipeManager_ListRecipes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipesList()
		status := types.RecipeStatusSubmitted

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipes), testutils.ContextMatcher, status, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := rm.ListRecipes(ctx, status, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		analyzer := &recipeanalysis.MockRecipeAnalyzer{}
		analyzer.On(reflection.GetMethodName(analyzer.ValidateRecipeCreationRequestInputIsDAG), testutils.ContextMatcher, testutils.MatchType[*types.RecipeCreationRequestInput]()).Return(nil)
		rm.recipeAnalyzer = analyzer

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.CreateRecipe), testutils.ContextMatcher, testutils.MatchType[*types.RecipeDatabaseCreationInput]()).Return(expected, nil)
				db.On(reflection.GetMethodName(rm.db.GetRecipe), testutils.ContextMatcher, expected.ID).Return(expected, nil)
			},
			map[string][]string{
				types.RecipeCreatedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
				},
			},
		)

		actual, err := rm.CreateRecipe(ctx, fakeCreatorID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, append(expectations, analyzer)...)
	})

	T.Run("with DAG error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		fakeCreatorID := fakes.BuildFakeID()
		fakeInput := fakes.BuildFakeRecipeCreationRequestInput()

		analyzer := &recipeanalysis.MockRecipeAnalyzer{}
		analyzer.On(reflection.GetMethodName(analyzer.ValidateRecipeCreationRequestInputIsDAG), testutils.ContextMatcher, testutils.MatchType[*types.RecipeCreationRequestInput]()).Return(errors.New("blah"))
		rm.recipeAnalyzer = analyzer

		actual, err := rm.CreateRecipe(ctx, fakeCreatorID, fakeInput)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, analyzer)
	})
}

func TestRecipeManager_ReadRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipe()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipe), testutils.ContextMatcher, expected.ID).Return(expected, nil)
			},
		)

		actual, err := rm.ReadRecipe(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.SearchForRecipes), testutils.ContextMatcher, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := rm.SearchRecipes(ctx, exampleQuery, false, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})

	T.Run("useSearchService true falls back to database when search returns empty", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipesList()
		exampleQuery := fakes.BuildFakeID()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.SearchForRecipes), testutils.ContextMatcher, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := rm.SearchRecipes(ctx, exampleQuery, true, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipe), testutils.ContextMatcher, exampleRecipe.ID).Return(exampleRecipe, nil)
				db.On(reflection.GetMethodName(rm.db.UpdateRecipe), testutils.ContextMatcher, testutils.MatchType[*types.Recipe]()).Return(nil)
			},
			map[string][]string{
				types.RecipeUpdatedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
				},
			},
		)

		assert.NoError(t, rm.UpdateRecipe(ctx, exampleRecipe.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.ArchiveRecipe), testutils.ContextMatcher, expected.ID, exampleOwnerID).Return(nil)
			},
			map[string][]string{
				types.RecipeArchivedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
				},
			},
		)

		assert.NoError(t, rm.ArchiveRecipe(ctx, expected.ID, exampleOwnerID))

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForRecipeManagerWithAnalyzer(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipe), testutils.ContextMatcher, exampleRecipe.ID).Return(exampleRecipe, nil)
			},
			func(analyzer *recipeanalysis.MockRecipeAnalyzer) {
				analyzer.On(reflection.GetMethodName(analyzer.GenerateMealPlanTasksForRecipe), testutils.ContextMatcher, "", testutils.MatchType[*types.Recipe]()).Return(expectedResults, nil)
			},
		)

		results, err := rm.RecipeEstimatedPrepSteps(ctx, exampleRecipe.ID)
		assert.NoError(t, err)

		assert.Equal(t, len(results), len(expectedResults))

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForRecipeManagerWithAnalyzer(
			rm,
			nil,
			func(analyzer *recipeanalysis.MockRecipeAnalyzer) {
				analyzer.On(reflection.GetMethodName(analyzer.RenderMermaidDiagramForMeal), testutils.ContextMatcher, testutils.MatchType[*types.Meal]()).Return(expectedResult, nil)
			},
		)

		result, err := rm.MealMermaid(ctx, exampleMeal)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForRecipeManagerWithAnalyzer(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipe), testutils.ContextMatcher, exampleRecipe.ID).Return(exampleRecipe, nil)
			},
			func(analyzer *recipeanalysis.MockRecipeAnalyzer) {
				analyzer.On(reflection.GetMethodName(analyzer.RenderMermaidDiagramForRecipe), testutils.ContextMatcher, testutils.MatchType[*types.Recipe]()).Return(expectedResult, nil)
			},
		)

		result, err := rm.RecipeMermaid(ctx, exampleRecipe.ID)
		assert.NoError(t, err)
		assert.Equal(t, result, expectedResult)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipe), testutils.ContextMatcher, expected.ID).Return(expected, nil)
				db.On(reflection.GetMethodName(rm.db.CreateRecipe), testutils.ContextMatcher, testutils.MatchType[*types.RecipeDatabaseCreationInput]()).Return(cloned, nil)
			},
			map[string][]string{
				types.RecipeClonedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
				},
			},
		)

		actual, err := rm.CloneRecipe(ctx, expected.ID, exampleOwnerID)
		assert.NoError(t, err)
		assert.Equal(t, cloned, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
