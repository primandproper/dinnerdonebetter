package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v4/filtering"
	"github.com/primandproper/platform-go/v4/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealLists), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.ListMealLists(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.CreateMealList), testutils.ContextMatcher, testutils.MatchType[*types.MealListDatabaseCreationInput]()).Return(expected, nil)
			},
		)

		actual, err := mpm.CreateMealList(ctx, userID, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.ArchiveMealList), testutils.ContextMatcher, listID, userID).Return(nil)
			},
		)

		assert.NoError(t, mpm.ArchiveMealList(ctx, listID, userID))

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.UpdateMealList), testutils.ContextMatcher, testutils.MatchType[*types.MealList]()).Return(nil)
			},
		)

		assert.NoError(t, mpm.UpdateMealList(ctx, listID, userID, input))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
