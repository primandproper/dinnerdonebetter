package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMealPlanningManager_ListUserIngredientPreferences(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeUserIngredientPreferencesList()
		exampleOwnerID := fakes.BuildFakeID()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetUserIngredientPreferences), testutils.ContextMatcher, exampleOwnerID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.ListUserIngredientPreferences(ctx, exampleOwnerID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_CreateUserIngredientPreference(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeUserIngredientPreferencesList().Data
		userID := fakes.BuildFakeID()
		fakeInput := fakes.BuildFakeUserIngredientPreferenceCreationRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.CreateUserIngredientPreference), testutils.ContextMatcher, testutils.MatchType[*types.UserIngredientPreferenceDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.UserIngredientPreferenceCreatedServiceEventType: {mealplanningkeys.ValidIngredientGroupIDKey, mealplanningkeys.ValidIngredientIDKey, "created"},
			},
		)

		actual, err := mpm.CreateUserIngredientPreference(ctx, userID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_UpdateUserIngredientPreference(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleUserIngredientPreference := fakes.BuildFakeUserIngredientPreference()
		ownerID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeUserIngredientPreferenceUpdateRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetUserIngredientPreference), testutils.ContextMatcher, exampleUserIngredientPreference.ID, ownerID).Return(exampleUserIngredientPreference, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateUserIngredientPreference), testutils.ContextMatcher, testutils.MatchType[*types.UserIngredientPreference]()).Return(nil)
			},
			map[string][]string{
				types.UserIngredientPreferenceUpdatedServiceEventType: {
					mealplanningkeys.UserIngredientPreferenceIDKey,
				},
			},
		)

		assert.NoError(t, mpm.UpdateUserIngredientPreference(ctx, exampleUserIngredientPreference.ID, ownerID, exampleInput))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_ArchiveUserIngredientPreference(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		ownershipID := fakes.BuildFakeID()
		expected := fakes.BuildFakeUserIngredientPreference()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.ArchiveUserIngredientPreference), testutils.ContextMatcher, expected.ID, ownershipID).Return(nil)
			},
			map[string][]string{
				types.UserIngredientPreferenceArchivedServiceEventType: {
					mealplanningkeys.UserIngredientPreferenceIDKey,
				},
			},
		)

		err := mpm.ArchiveUserIngredientPreference(ctx, ownershipID, expected.ID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
