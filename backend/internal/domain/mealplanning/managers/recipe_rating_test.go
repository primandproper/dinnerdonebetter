package managers

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v7/filtering"

	"github.com/stretchr/testify/assert"
)

func TestRecipeManager_ListRecipeRatings(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipeRatingsList()
		exampleRecipeID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeRatingsForRecipeFunc: func(_ context.Context, recipeID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeRating], error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ListRecipeRatings(ctx, exampleRecipeID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeRatingsForRecipeCalls(), 1)
	})
}

func TestRecipeManager_ReadRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeRating()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeRatingFunc: func(_ context.Context, recipeID string, recipeRatingID string) (*types.RecipeRating, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, expected.ID, recipeRatingID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ReadRecipeRating(ctx, exampleRecipeID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeRatingCalls(), 1)
	})
}

func TestRecipeManager_CreateRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeRating()
		fakeInput := fakes.BuildFakeRecipeRatingCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateRecipeRatingFunc: func(_ context.Context, _ *types.RecipeRatingDatabaseCreationInput) (*types.RecipeRating, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeRating(ctx, exampleRecipeID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateRecipeRatingCalls(), 1)
	})
}

func TestRecipeManager_UpdateRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeRating := fakes.BuildFakeRecipeRating()
		exampleInput := fakes.BuildFakeRecipeRatingUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeRatingFunc: func(_ context.Context, recipeID string, recipeRatingID string) (*types.RecipeRating, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeRating.ID, recipeRatingID)

				return exampleRecipeRating, nil
			},
			UpdateRecipeRatingFunc: func(_ context.Context, _ *types.RecipeRating) error {
				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		assert.NoError(t, rm.UpdateRecipeRating(ctx, exampleRecipeID, exampleRecipeRating.ID, exampleInput))

		assert.Len(t, db.GetRecipeRatingCalls(), 1)
		assert.Len(t, db.UpdateRecipeRatingCalls(), 1)
	})
}

func TestRecipeManager_ArchiveRecipeRating(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeRating()

		db := &mealplanningmock.RepositoryMock{
			ArchiveRecipeRatingFunc: func(_ context.Context, recipeID string, recipeRatingID string) error {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, expected.ID, recipeRatingID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		assert.NoError(t, rm.ArchiveRecipeRating(ctx, exampleRecipeID, expected.ID))

		assert.Len(t, db.ArchiveRecipeRatingCalls(), 1)
	})
}
