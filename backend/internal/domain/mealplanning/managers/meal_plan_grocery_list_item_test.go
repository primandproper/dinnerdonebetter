package managers

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMealPlanningManager_ListMealPlanGroceryListItemsByMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanGroceryListItemsList()
		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			GetMealPlanGroceryListItemsForMealPlanFunc: func(_ context.Context, mealPlanID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanGroceryListItem], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlanGroceryListItemsByMealPlan(ctx, exampleMealPlanID, exampleOwnerID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanGroceryListItemsForMealPlanCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlanGroceryListItemsByMealPlan(ctx, exampleMealPlanID, exampleOwnerID, nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.GetMealPlanGroceryListItemsForMealPlanCalls())
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
		require.NoError(t, err)
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

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlanGroceryListItem()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			GetMealPlanGroceryListItemFunc: func(_ context.Context, mealPlanID string, mealPlanGroceryListItemID string) (*types.MealPlanGroceryListItem, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, expected.ID, mealPlanGroceryListItemID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlanGroceryListItem(ctx, exampleMealPlanID, expected.ID, exampleOwnerID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanGroceryListItemCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlanGroceryListItem(ctx, exampleMealPlanID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.GetMealPlanGroceryListItemCalls())
	})
}

func TestMealPlanningManager_UpdateMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanGroceryListItem := fakes.BuildFakeMealPlanGroceryListItem()
		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanGroceryListItemUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
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

		require.NoError(t, mpm.UpdateMealPlanGroceryListItem(ctx, exampleMealPlanID, exampleMealPlanGroceryListItem.ID, exampleOwnerID, exampleInput))

		assert.Len(t, db.GetMealPlanGroceryListItemCalls(), 1)
		assert.Len(t, db.UpdateMealPlanGroceryListItemCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.UpdateMealPlanGroceryListItem(ctx, exampleMealPlanID, fake.BuildFakeID(), exampleOwnerID, fakes.BuildFakeMealPlanGroceryListItemUpdateRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.GetMealPlanGroceryListItemCalls())
		assert.Empty(t, db.UpdateMealPlanGroceryListItemCalls())
	})
}

func TestMealPlanningManager_ArchiveMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlanGroceryListItem()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:                mealPlanExistsStub(t, mealPlanID, exampleOwnerID, true),
			MealPlanGroceryListItemExistsFunc: mealPlanGroceryListItemExistsStub(t, mealPlanID, expected.ID, true),
			ArchiveMealPlanGroceryListItemFunc: func(_ context.Context, mealPlanGroceryListItemID string) error {
				assert.Equal(t, expected.ID, mealPlanGroceryListItemID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanGroceryListItem(ctx, mealPlanID, expected.ID, exampleOwnerID)
		require.NoError(t, err)

		assert.Len(t, db.ArchiveMealPlanGroceryListItemCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanGroceryListItem(ctx, exampleMealPlanID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ArchiveMealPlanGroceryListItemCalls())
	})

	// The archive is keyed by the item alone, so the item has to be shown to be the plan's.
	T.Run("with a grocery list item from another meal plan", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleItemID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:                mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			MealPlanGroceryListItemExistsFunc: mealPlanGroceryListItemExistsStub(t, exampleMealPlanID, exampleItemID, false),
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanGroceryListItem(ctx, exampleMealPlanID, exampleItemID, exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ArchiveMealPlanGroceryListItemCalls())
	})
}
