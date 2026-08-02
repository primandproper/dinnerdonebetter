package managers

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/stretchr/testify/assert"
)

func TestMealPlanningManager_ListMealPlanGroceryListItemsByMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanGroceryListItemsList()
		exampleMealPlanID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanGroceryListItemsForMealPlanFunc: func(_ context.Context, mealPlanID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanGroceryListItem], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlanGroceryListItemsByMealPlan(ctx, exampleMealPlanID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanGroceryListItemsForMealPlanCalls(), 1)
	})
}

func TestMealPlanningManager_CreateMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanGroceryListItem()
		fakeInput := fakes.BuildFakeMealPlanGroceryListItemCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateMealPlanGroceryListItemFunc: func(_ context.Context, _ *types.MealPlanGroceryListItemDatabaseCreationInput) (*types.MealPlanGroceryListItem, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanGroceryListItem(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanGroceryListItemCalls(), 1)
	})
}

func TestMealPlanningManager_ReadMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanGroceryListItem()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanGroceryListItemFunc: func(_ context.Context, mealPlanID string, mealPlanGroceryListItemID string) (*types.MealPlanGroceryListItem, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, expected.ID, mealPlanGroceryListItemID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlanGroceryListItem(ctx, exampleMealPlanID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanGroceryListItemCalls(), 1)
	})
}

func TestMealPlanningManager_UpdateMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanGroceryListItem := fakes.BuildFakeMealPlanGroceryListItem()
		exampleMealPlanID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanGroceryListItemUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanGroceryListItemFunc: func(_ context.Context, mealPlanID string, mealPlanGroceryListItemID string) (*types.MealPlanGroceryListItem, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanGroceryListItem.ID, mealPlanGroceryListItemID)

				return exampleMealPlanGroceryListItem, nil
			},
			UpdateMealPlanGroceryListItemFunc: func(_ context.Context, _ *types.MealPlanGroceryListItem) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		assert.NoError(t, mpm.UpdateMealPlanGroceryListItem(ctx, exampleMealPlanID, exampleMealPlanGroceryListItem.ID, exampleInput))

		assert.Len(t, db.GetMealPlanGroceryListItemCalls(), 1)
		assert.Len(t, db.UpdateMealPlanGroceryListItemCalls(), 1)
	})
}

func TestMealPlanningManager_ArchiveMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanGroceryListItem()

		db := &mealplanningmock.RepositoryMock{
			ArchiveMealPlanGroceryListItemFunc: func(_ context.Context, mealPlanGroceryListItemID string) error {
				assert.Equal(t, expected.ID, mealPlanGroceryListItemID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanGroceryListItem(ctx, mealPlanID, expected.ID)
		assert.NoError(t, err)

		assert.Len(t, db.ArchiveMealPlanGroceryListItemCalls(), 1)
	})
}
