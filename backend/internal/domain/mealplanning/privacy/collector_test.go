package privacy

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v6/filtering"
	"github.com/primandproper/platform-go/v6/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v6/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v6/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildRecipePage(size int) *filtering.QueryFilteredResult[mealplanning.Recipe] {
	data := make([]*mealplanning.Recipe, 0, size)
	for range size {
		data = append(data, fakes.BuildFakeRecipe())
	}

	return filtering.NewQueryFilteredResult(
		data,
		uint64(size),
		uint64(size),
		func(r *mealplanning.Recipe) string { return r.ID },
		filtering.DefaultQueryFilter(),
	)
}

func buildEmptyPage[T any]() *filtering.QueryFilteredResult[T] {
	return filtering.NewQueryFilteredResult(
		[]*T{},
		0,
		0,
		func(*T) string { return "" },
		filtering.DefaultQueryFilter(),
	)
}

func buildSingleItemPage[T any](item *T, idExtractor func(*T) string) *filtering.QueryFilteredResult[T] {
	return filtering.NewQueryFilteredResult(
		[]*T{item},
		1,
		1,
		idExtractor,
		filtering.DefaultQueryFilter(),
	)
}

func TestCollector_CollectUserData(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		userID := identifiers.New()
		repo := &mocks.Repository{}
		collector := NewCollector(repo, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider())

		// recipes span two pages: a full first page proves the collector keeps paging.
		fullPage := buildRecipePage(filtering.MaxQueryFilterLimit)
		lastPage := buildRecipePage(3)

		repo.On("GetRecipesCreatedByUser", mock.Anything, userID, mock.MatchedBy(func(filter *filtering.QueryFilter) bool {
			return filter != nil && filter.Cursor == nil
		})).Return(fullPage, nil)
		repo.On("GetRecipesCreatedByUser", mock.Anything, userID, mock.MatchedBy(func(filter *filtering.QueryFilter) bool {
			return filter != nil && filter.Cursor != nil && *filter.Cursor == fullPage.Cursor
		})).Return(lastPage, nil)

		exampleMeal := fakes.BuildFakeMeal()
		repo.On("GetMealsCreatedByUser", mock.Anything, userID, mock.Anything).
			Return(buildSingleItemPage(exampleMeal, func(m *mealplanning.Meal) string { return m.ID }), nil)
		repo.On("GetUserIngredientPreferences", mock.Anything, userID, mock.Anything).
			Return(buildEmptyPage[mealplanning.UserIngredientPreference](), nil)

		exampleRating := fakes.BuildFakeRecipeRating()
		repo.On("GetRecipeRatingsForUser", mock.Anything, userID, mock.Anything).
			Return(buildSingleItemPage(exampleRating, func(r *mealplanning.RecipeRating) string { return r.ID }), nil)

		collection := &dataprivacy.UserDataCollection{}
		err := collector.CollectUserData(t.Context(), collection, userID)

		require.NoError(t, err)
		assert.Len(t, collection.MealPlanning.Recipes, filtering.MaxQueryFilterLimit+3, "both recipe pages must be collected")
		assert.Len(t, collection.MealPlanning.Meals, 1)
		assert.Empty(t, collection.MealPlanning.UserIngredientPreferences)
		assert.Len(t, collection.MealPlanning.RecipeRatings, 1)

		mock.AssertExpectationsForObjects(t, repo)
	})
}

func TestCollector_CollectAccountData(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		accountID := identifiers.New()
		repo := &mocks.Repository{}
		collector := NewCollector(repo, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider())

		exampleMealPlan := fakes.BuildFakeMealPlan()
		repo.On("GetMealPlansForAccount", mock.Anything, accountID, mock.Anything).
			Return(buildSingleItemPage(exampleMealPlan, func(mp *mealplanning.MealPlan) string { return mp.ID }), nil)

		exampleOwnership := fakes.BuildFakeAccountInstrumentOwnership()
		repo.On("GetAccountInstrumentOwnerships", mock.Anything, accountID, mock.Anything).
			Return(buildSingleItemPage(exampleOwnership, func(o *mealplanning.AccountInstrumentOwnership) string { return o.ID }), nil)

		collection := &dataprivacy.UserDataCollection{}
		err := collector.CollectAccountData(t.Context(), collection, accountID)

		require.NoError(t, err)
		assert.Len(t, collection.MealPlanning.MealPlans, 1)
		assert.Len(t, collection.MealPlanning.AccountInstrumentOwnerships, 1)

		mock.AssertExpectationsForObjects(t, repo)
	})
}
