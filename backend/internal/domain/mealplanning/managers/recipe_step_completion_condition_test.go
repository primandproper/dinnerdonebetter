package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v4/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRecipeManager_ListRecipeStepCompletionConditions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipeStepCompletionConditionsList()
		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipeStepCompletionConditions), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := rm.ListRecipeStepCompletionConditions(ctx, exampleRecipeID, exampleRecipeStepID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_CreateRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepCompletionCondition()
		fakeInput := fakes.BuildFakeRecipeStepCompletionConditionForExistingRecipeCreationRequestInput()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.CreateRecipeStepCompletionCondition), testutils.ContextMatcher, testutils.MatchType[*types.RecipeStepCompletionConditionDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.RecipeStepCompletionConditionCreatedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
					mealplanningkeys.RecipeStepIDKey,
					mealplanningkeys.RecipeStepCompletionConditionIDKey,
				},
			},
		)

		actual, err := rm.CreateRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_ReadRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepCompletionCondition()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipeStepCompletionCondition), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, expected.ID).Return(expected, nil)
			},
		)

		actual, err := rm.ReadRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_UpdateRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		exampleRecipeStepCompletionCondition := fakes.BuildFakeRecipeStepCompletionCondition()
		exampleInput := fakes.BuildFakeRecipeStepCompletionConditionUpdateRequestInput()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipeStepCompletionCondition), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepCompletionCondition.ID).Return(exampleRecipeStepCompletionCondition, nil)
				db.On(reflection.GetMethodName(rm.db.UpdateRecipeStepCompletionCondition), testutils.ContextMatcher, testutils.MatchType[*types.RecipeStepCompletionCondition]()).Return(nil)
			},
			map[string][]string{
				types.RecipeStepCompletionConditionUpdatedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
					mealplanningkeys.RecipeStepIDKey,
					mealplanningkeys.RecipeStepCompletionConditionIDKey,
				},
			},
		)

		assert.NoError(t, rm.UpdateRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepCompletionCondition.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_ArchiveRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepCompletionCondition()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.ArchiveRecipeStepCompletionCondition), testutils.ContextMatcher, exampleRecipeStepID, expected.ID).Return(nil)
			},
			map[string][]string{
				types.RecipeStepCompletionConditionArchivedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
					mealplanningkeys.RecipeStepIDKey,
					mealplanningkeys.RecipeStepCompletionConditionIDKey,
				},
			},
		)

		assert.NoError(t, rm.ArchiveRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
