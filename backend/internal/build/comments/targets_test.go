package comments

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningmgr "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers"
	mockmanagers "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers/mock"

	platformcomments "github.com/primandproper/platform-go/v15/comments"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// everyTargetType is the whole catalog, by domain.
var everyTargetType = []platformcomments.TargetType{
	issuereports.CommentTargetTypeIssueReports,
	mealplanning.CommentTargetTypeRecipes,
	mealplanning.CommentTargetTypeMeals,
	mealplanning.CommentTargetTypeMealPlans,
}

func TestCatalog(T *testing.T) {
	T.Parallel()

	T.Run("is every domain's targets, with no checks", func(t *testing.T) {
		t.Parallel()

		catalog, err := Catalog()
		require.NoError(t, err)
		assert.Len(t, catalog, len(everyTargetType))

		for _, targetType := range everyTargetType {
			require.True(t, catalog.Known(targetType), "%s is missing", targetType)
			assert.NotEmpty(t, catalog[targetType].Description)
			assert.Nil(t, catalog[targetType].Exists, "%s carries a check in the hookless catalog", targetType)
		}
	})
}

func TestCatalogWithChecks(T *testing.T) {
	T.Parallel()

	T.Run("carries each domain's checks over the same targets", func(t *testing.T) {
		t.Parallel()

		i := do.New()
		do.ProvideValue[mealplanningmgr.MealPlanningManager](i, &mockmanagers.MealPlanningManagerMock{
			ReadRecipeFunc: func(context.Context, string) (*mealplanning.Recipe, error) { return &mealplanning.Recipe{}, nil },
			ReadMealFunc:   func(context.Context, string) (*mealplanning.Meal, error) { return &mealplanning.Meal{}, nil },
		})

		catalog, err := CatalogWithChecks(i)
		require.NoError(t, err)
		assert.Len(t, catalog, len(everyTargetType))

		for _, targetType := range everyTargetType {
			require.True(t, catalog.Known(targetType), "%s is missing", targetType)
		}

		assert.NotNil(t, catalog[mealplanning.CommentTargetTypeRecipes].Exists)
		assert.NotNil(t, catalog[mealplanning.CommentTargetTypeMeals].Exists)
		assert.Nil(t, catalog[mealplanning.CommentTargetTypeMealPlans].Exists)
		assert.Nil(t, catalog[issuereports.CommentTargetTypeIssueReports].Exists)
	})
}

func TestMerge(T *testing.T) {
	T.Parallel()

	T.Run("refuses a type two domains both claim", func(t *testing.T) {
		t.Parallel()

		catalog := platformcomments.Targets{
			mealplanning.CommentTargetTypeRecipes: {Description: "A recipe."},
		}

		err := merge(catalog, platformcomments.Targets{
			mealplanning.CommentTargetTypeRecipes: {Description: "Also a recipe."},
		})
		require.Error(t, err)
		assert.Equal(t, "A recipe.", catalog[mealplanning.CommentTargetTypeRecipes].Description, "the first claim stands")
	})
}
