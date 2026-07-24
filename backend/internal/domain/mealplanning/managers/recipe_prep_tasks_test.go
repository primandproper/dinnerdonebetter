package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v6/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRecipeManager_ListRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipePrepTasksList()
		exampleRecipeID := fakes.BuildFakeID()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipePrepTasks), testutils.ContextMatcher, exampleRecipeID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := rm.ListRecipePrepTask(ctx, exampleRecipeID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_CreateRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipePrepTask()
		fakeInput := fakes.BuildFakeRecipePrepTaskCreationRequestInput()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.CreateRecipePrepTask), testutils.ContextMatcher, testutils.MatchType[*types.RecipePrepTaskDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.RecipePrepTaskCreatedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
					mealplanningkeys.RecipePrepTaskIDKey,
				},
			},
		)

		actual, err := rm.CreateRecipePrepTask(ctx, exampleRecipeID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_ReadRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipePrepTask()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipePrepTask), testutils.ContextMatcher, exampleRecipeID, expected.ID).Return(expected, nil)
			},
		)

		actual, err := rm.ReadRecipePrepTask(ctx, exampleRecipeID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_UpdateRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipePrepTask := fakes.BuildFakeRecipePrepTask()
		exampleInput := fakes.BuildFakeRecipePrepTaskUpdateRequestInput()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipePrepTask), testutils.ContextMatcher, exampleRecipeID, exampleRecipePrepTask.ID).Return(exampleRecipePrepTask, nil)
				db.On(reflection.GetMethodName(rm.db.UpdateRecipePrepTask), testutils.ContextMatcher, testutils.MatchType[*types.RecipePrepTask]()).Return(nil)
			},
			map[string][]string{
				types.RecipePrepTaskUpdatedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
					mealplanningkeys.RecipePrepTaskIDKey,
				},
			},
		)

		assert.NoError(t, rm.UpdateRecipePrepTask(ctx, exampleRecipeID, exampleRecipePrepTask.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_ArchiveRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipePrepTask()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.ArchiveRecipePrepTask), testutils.ContextMatcher, exampleRecipeID, expected.ID).Return(nil)
			},
			map[string][]string{
				types.RecipePrepTaskArchivedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
					mealplanningkeys.RecipePrepTaskIDKey,
				},
			},
		)

		assert.NoError(t, rm.ArchiveRecipePrepTask(ctx, exampleRecipeID, expected.ID))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
