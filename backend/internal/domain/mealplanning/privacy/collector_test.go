package privacy

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
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

		// recipes span two pages: a full first page proves the collector keeps paging.
		fullPage := buildRecipePage(filtering.MaxQueryFilterLimit)
		lastPage := buildRecipePage(3)

		exampleMeal := fakes.BuildFakeMeal()
		exampleRating := fakes.BuildFakeRecipeRating()

		repo := &mocks.RepositoryMock{
			GetRecipesCreatedByUserFunc: func(_ context.Context, actualUserID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
				assert.Equal(t, userID, actualUserID)
				require.NotNil(t, filter)

				// the first page is requested without a cursor; the second resumes from the first page's cursor.
				if filter.Cursor == nil {
					return fullPage, nil
				}

				assert.Equal(t, fullPage.Cursor, *filter.Cursor)

				return lastPage, nil
			},
			GetMealsCreatedByUserFunc: func(_ context.Context, actualUserID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Meal], error) {
				assert.Equal(t, userID, actualUserID)

				return buildSingleItemPage(exampleMeal, func(m *mealplanning.Meal) string { return m.ID }), nil
			},
			GetUserIngredientPreferencesFunc: func(_ context.Context, actualUserID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.UserIngredientPreference], error) {
				assert.Equal(t, userID, actualUserID)

				return buildEmptyPage[mealplanning.UserIngredientPreference](), nil
			},
			GetRecipeRatingsForUserFunc: func(_ context.Context, actualUserID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeRating], error) {
				assert.Equal(t, userID, actualUserID)

				return buildSingleItemPage(exampleRating, func(r *mealplanning.RecipeRating) string { return r.ID }), nil
			},
		}

		collector := NewCollector(repo, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider())

		collection := &dataprivacy.UserDataCollection{}
		err := collector.CollectUserData(t.Context(), collection, userID)

		require.NoError(t, err)
		assert.Len(t, collection.MealPlanning.Recipes, filtering.MaxQueryFilterLimit+3, "both recipe pages must be collected")
		assert.Len(t, collection.MealPlanning.Meals, 1)
		assert.Empty(t, collection.MealPlanning.UserIngredientPreferences)
		assert.Len(t, collection.MealPlanning.RecipeRatings, 1)

		assert.Len(t, repo.GetRecipesCreatedByUserCalls(), 2)
		assert.Len(t, repo.GetMealsCreatedByUserCalls(), 1)
		assert.Len(t, repo.GetUserIngredientPreferencesCalls(), 1)
		assert.Len(t, repo.GetRecipeRatingsForUserCalls(), 1)
	})
}

func TestCollector_CollectAccountData(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		accountID := identifiers.New()

		exampleMealPlan := fakes.BuildFakeMealPlan()
		exampleOwnership := fakes.BuildFakeAccountInstrumentOwnership()

		repo := &mocks.RepositoryMock{
			GetMealPlansForAccountFunc: func(_ context.Context, actualAccountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlan], error) {
				assert.Equal(t, accountID, actualAccountID)

				return buildSingleItemPage(exampleMealPlan, func(mp *mealplanning.MealPlan) string { return mp.ID }), nil
			},
			GetAccountInstrumentOwnershipsFunc: func(_ context.Context, actualAccountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.AccountInstrumentOwnership], error) {
				assert.Equal(t, accountID, actualAccountID)

				return buildSingleItemPage(exampleOwnership, func(o *mealplanning.AccountInstrumentOwnership) string { return o.ID }), nil
			},
		}

		collector := NewCollector(repo, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider())

		collection := &dataprivacy.UserDataCollection{}
		err := collector.CollectAccountData(t.Context(), collection, accountID)

		require.NoError(t, err)
		assert.Len(t, collection.MealPlanning.MealPlans, 1)
		assert.Len(t, collection.MealPlanning.AccountInstrumentOwnerships, 1)

		assert.Len(t, repo.GetMealPlansForAccountCalls(), 1)
		assert.Len(t, repo.GetAccountInstrumentOwnershipsCalls(), 1)
	})
}
