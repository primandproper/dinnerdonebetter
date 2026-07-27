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

func TestMealPlanningManager_ListMealLists(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		ml := &types.MealList{
			ID:            fakes.BuildFakeID(),
			Name:          t.Name(),
			Description:   t.Name(),
			BelongsToUser: fakes.BuildFakeID(),
		}
		expected := &filtering.QueryFilteredResult[types.MealList]{Data: []*types.MealList{ml}}

		db := &mealplanningmock.RepositoryMock{
			GetMealListsFunc: func(_ context.Context, userID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealList], error) {
				assert.Equal(t, ml.BelongsToUser, userID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealLists(ctx, ml.BelongsToUser, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealListsCalls(), 1)
	})
}

func TestMealPlanningManager_CreateMealList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		userID := fakes.BuildFakeID()
		input := &types.MealListCreationRequestInput{
			Name:        t.Name(),
			Description: t.Name(),
		}
		expected := &types.MealList{
			ID:            fakes.BuildFakeID(),
			Name:          input.Name,
			Description:   input.Description,
			BelongsToUser: userID,
		}

		db := &mealplanningmock.RepositoryMock{
			CreateMealListFunc: func(_ context.Context, _ *types.MealListDatabaseCreationInput) (*types.MealList, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealList(ctx, userID, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealListCalls(), 1)
	})
}

func TestMealPlanningManager_ArchiveMealList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		listID := fakes.BuildFakeID()
		userID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			ArchiveMealListFunc: func(_ context.Context, mealListID string, actualUserID string) error {
				assert.Equal(t, listID, mealListID)
				assert.Equal(t, userID, actualUserID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		assert.NoError(t, mpm.ArchiveMealList(ctx, listID, userID))

		assert.Len(t, db.ArchiveMealListCalls(), 1)
	})
}

func TestMealPlanningManager_UpdateMealList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		listID := fakes.BuildFakeID()
		userID := fakes.BuildFakeID()
		name := t.Name()
		desc := "desc"
		input := &types.MealListUpdateRequestInput{
			Name:        &name,
			Description: &desc,
		}

		db := &mealplanningmock.RepositoryMock{
			UpdateMealListFunc: func(_ context.Context, _ *types.MealList) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		assert.NoError(t, mpm.UpdateMealList(ctx, listID, userID, input))

		assert.Len(t, db.UpdateMealListCalls(), 1)
	})
}
