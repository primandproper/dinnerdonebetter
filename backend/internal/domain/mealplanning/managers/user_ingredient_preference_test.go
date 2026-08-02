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

func TestMealPlanningManager_ListUserIngredientPreferences(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeUserIngredientPreferencesList()
		exampleOwnerID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetUserIngredientPreferencesFunc: func(_ context.Context, userID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.UserIngredientPreference], error) {
				assert.Equal(t, exampleOwnerID, userID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListUserIngredientPreferences(ctx, exampleOwnerID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetUserIngredientPreferencesCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			CreateUserIngredientPreferenceFunc: func(_ context.Context, _ *types.UserIngredientPreferenceDatabaseCreationInput) ([]*types.UserIngredientPreference, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateUserIngredientPreference(ctx, userID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateUserIngredientPreferenceCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			GetUserIngredientPreferenceFunc: func(_ context.Context, userIngredientPreferenceID string, userID string) (*types.UserIngredientPreference, error) {
				assert.Equal(t, exampleUserIngredientPreference.ID, userIngredientPreferenceID)
				assert.Equal(t, ownerID, userID)

				return exampleUserIngredientPreference, nil
			},
			UpdateUserIngredientPreferenceFunc: func(_ context.Context, _ *types.UserIngredientPreference) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		assert.NoError(t, mpm.UpdateUserIngredientPreference(ctx, exampleUserIngredientPreference.ID, ownerID, exampleInput))

		assert.Len(t, db.GetUserIngredientPreferenceCalls(), 1)
		assert.Len(t, db.UpdateUserIngredientPreferenceCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			ArchiveUserIngredientPreferenceFunc: func(_ context.Context, userIngredientPreferenceID string, userID string) error {
				assert.Equal(t, expected.ID, userIngredientPreferenceID)
				assert.Equal(t, ownershipID, userID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveUserIngredientPreference(ctx, ownershipID, expected.ID)
		assert.NoError(t, err)

		assert.Len(t, db.ArchiveUserIngredientPreferenceCalls(), 1)
	})
}
