package managers

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipeManager_ListRecipeStepProducts(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipeStepProductsList()
		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepProductsFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepProduct], error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ListRecipeStepProducts(ctx, exampleRecipeID, exampleRecipeStepID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeStepProductsCalls(), 1)
	})
}

func TestRecipeManager_CreateRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepProduct()
		fakeInput := fakes.BuildFakeRecipeStepProductCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateRecipeStepProductFunc: func(_ context.Context, _ string, _ *types.RecipeStepProductDatabaseCreationInput) (*types.RecipeStepProduct, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepProduct(ctx, exampleRecipeID, exampleRecipeStepID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateRecipeStepProductCalls(), 1)
	})
}

func TestRecipeManager_ReadRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepProduct()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepProductFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepProductID string) (*types.RecipeStepProduct, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, expected.ID, recipeStepProductID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ReadRecipeStepProduct(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeStepProductCalls(), 1)
	})
}

func TestRecipeManager_UpdateRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		exampleRecipeStepProduct := fakes.BuildFakeRecipeStepProduct()
		exampleInput := fakes.BuildFakeRecipeStepProductUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepProductFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepProductID string) (*types.RecipeStepProduct, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, exampleRecipeStepProduct.ID, recipeStepProductID)

				return exampleRecipeStepProduct, nil
			},
			UpdateRecipeStepProductFunc: func(_ context.Context, _ string, _ *types.RecipeStepProduct) error {
				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.UpdateRecipeStepProduct(ctx, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepProduct.ID, exampleInput))

		assert.Len(t, db.GetRecipeStepProductCalls(), 1)
		assert.Len(t, db.UpdateRecipeStepProductCalls(), 1)
	})
}

func TestRecipeManager_ArchiveRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepProduct()

		db := &mealplanningmock.RepositoryMock{
			ArchiveRecipeStepProductFunc: func(_ context.Context, _ string, recipeStepID string, recipeStepProductID string) error {
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, expected.ID, recipeStepProductID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.ArchiveRecipeStepProduct(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID))

		assert.Len(t, db.ArchiveRecipeStepProductCalls(), 1)
	})
}
