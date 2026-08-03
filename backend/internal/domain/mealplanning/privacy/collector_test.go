package privacy

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	platformdataprivacy "github.com/primandproper/platform-go/v9/dataprivacy"
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

func TestCollector_Collect(T *testing.T) {
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

		collection := collect(t, repo, noAccounts, userID)

		assert.Len(t, collection.Recipes, filtering.MaxQueryFilterLimit+3, "both recipe pages must be collected")
		assert.Len(t, collection.Meals, 1)
		assert.Empty(t, collection.UserIngredientPreferences)
		assert.Len(t, collection.RecipeRatings, 1)

		assert.Len(t, repo.GetRecipesCreatedByUserCalls(), 2)
		assert.Len(t, repo.GetMealsCreatedByUserCalls(), 1)
		assert.Len(t, repo.GetUserIngredientPreferencesCalls(), 1)
		assert.Len(t, repo.GetRecipeRatingsForUserCalls(), 1)
	})

	T.Run("nothing held is reported as no section rather than an empty one", func(t *testing.T) {
		t.Parallel()

		repo := &mocks.RepositoryMock{
			GetRecipesCreatedByUserFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
				return buildEmptyPage[mealplanning.Recipe](), nil
			},
			GetMealsCreatedByUserFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Meal], error) {
				return buildEmptyPage[mealplanning.Meal](), nil
			},
			GetUserIngredientPreferencesFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.UserIngredientPreference], error) {
				return buildEmptyPage[mealplanning.UserIngredientPreference](), nil
			},
			GetRecipeRatingsForUserFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeRating], error) {
				return buildEmptyPage[mealplanning.RecipeRating](), nil
			},
		}

		collector := NewCollector(repo, noAccounts, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider())

		fragment, err := collector.Collect(t.Context(), subjectFor(identifiers.New()))

		require.NoError(t, err)
		// nil, not an encoded empty object: the section is then omitted from the artifact,
		// so an export's sections are the domains that actually held something.
		assert.Nil(t, fragment)
	})
}

func TestCollector_Collect_accountScoped(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		accountID := identifiers.New()

		exampleMealPlan := fakes.BuildFakeMealPlan()
		exampleOwnership := fakes.BuildFakeAccountInstrumentOwnership()

		repo := &mocks.RepositoryMock{
			GetRecipesCreatedByUserFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
				return buildEmptyPage[mealplanning.Recipe](), nil
			},
			GetMealsCreatedByUserFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Meal], error) {
				return buildEmptyPage[mealplanning.Meal](), nil
			},
			GetUserIngredientPreferencesFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.UserIngredientPreference], error) {
				return buildEmptyPage[mealplanning.UserIngredientPreference](), nil
			},
			GetRecipeRatingsForUserFunc: func(context.Context, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeRating], error) {
				return buildEmptyPage[mealplanning.RecipeRating](), nil
			},
			GetMealPlansForAccountFunc: func(_ context.Context, actualAccountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlan], error) {
				assert.Equal(t, accountID, actualAccountID)

				return buildSingleItemPage(exampleMealPlan, func(mp *mealplanning.MealPlan) string { return mp.ID }), nil
			},
			GetAccountInstrumentOwnershipsFunc: func(_ context.Context, actualAccountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.AccountInstrumentOwnership], error) {
				assert.Equal(t, accountID, actualAccountID)

				return buildSingleItemPage(exampleOwnership, func(o *mealplanning.AccountInstrumentOwnership) string { return o.ID }), nil
			},
		}

		collection := collect(t, repo, resolvesTo(accountID), identifiers.New())

		assert.Len(t, collection.MealPlans, 1)
		assert.Len(t, collection.AccountInstrumentOwnerships, 1)

		assert.Len(t, repo.GetMealPlansForAccountCalls(), 1)
		assert.Len(t, repo.GetAccountInstrumentOwnershipsCalls(), 1)
	})
}

// noAccounts is the resolver for a subject who is in no accounts, which is what the
// user-scoped assertions want: the account-scoped queries are then never reached and
// cannot contribute to the counts being checked.
func noAccounts(context.Context, string) ([]string, error) { return nil, nil }

// resolvesTo builds a resolver that reports the subject as a member of accountIDs.
func resolvesTo(accountIDs ...string) dataprivacy.AccountIDResolver {
	return func(context.Context, string) ([]string, error) { return accountIDs, nil }
}

func subjectFor(userID string) platformdataprivacy.Subject {
	return platformdataprivacy.Subject{ID: userID, Type: platformdataprivacy.SubjectUser}
}

// collect runs the collector and decodes the fragment it produced.
//
// Decoding rather than reaching into the collector is the point of the interface: what a
// domain contributes to an export is opaque encoded JSON, so a test that asserted on an
// intermediate struct would be checking something the artifact never sees.
func collect(
	t *testing.T,
	repo mealplanning.Repository,
	resolveAccounts dataprivacy.AccountIDResolver,
	userID string,
) *mealplanning.UserDataCollection {
	t.Helper()

	collector := NewCollector(repo, resolveAccounts, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider())

	fragment, err := collector.Collect(t.Context(), subjectFor(userID))
	require.NoError(t, err)
	require.NotNil(t, fragment)

	var collection mealplanning.UserDataCollection
	require.NoError(t, json.Unmarshal(fragment, &collection))

	return &collection
}
