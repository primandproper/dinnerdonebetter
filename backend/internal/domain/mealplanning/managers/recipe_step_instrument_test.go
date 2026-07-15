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

func TestRecipeManager_ListRecipeStepInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipeStepInstrumentsList()
		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipeStepInstruments), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := rm.ListRecipeStepInstruments(ctx, exampleRecipeID, exampleRecipeStepID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_CreateRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepInstrument()
		fakeInput := fakes.BuildFakeRecipeStepInstrumentCreationRequestInput()
		fakeInput.Index = new(uint16(0))

		fakeValidPreparationInstrument := fakes.BuildFakeValidPreparationInstrument()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetValidPreparationInstrument), testutils.ContextMatcher, *fakeInput.ValidPreparationInstrumentID).Return(fakeValidPreparationInstrument, nil)
				db.On(reflection.GetMethodName(rm.db.CreateRecipeStepInstrument), testutils.ContextMatcher, testutils.MatchType[*types.RecipeStepInstrumentDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.RecipeStepInstrumentCreatedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
					mealplanningkeys.RecipeStepIDKey,
					mealplanningkeys.RecipeStepInstrumentIDKey,
				},
			},
		)

		actual, err := rm.CreateRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_ReadRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepInstrument()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipeStepInstrument), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, expected.ID).Return(expected, nil)
			},
		)

		actual, err := rm.ReadRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_UpdateRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		exampleRecipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()
		exampleInput := fakes.BuildFakeRecipeStepInstrumentUpdateRequestInput()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipeStepInstrument), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepInstrument.ID).Return(exampleRecipeStepInstrument, nil)
				db.On(reflection.GetMethodName(rm.db.UpdateRecipeStepInstrument), testutils.ContextMatcher, testutils.MatchType[*types.RecipeStepInstrument]()).Return(nil)
			},
			map[string][]string{
				types.RecipeStepInstrumentUpdatedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
					mealplanningkeys.RecipeStepIDKey,
					mealplanningkeys.RecipeStepInstrumentIDKey,
				},
			},
		)

		assert.NoError(t, rm.UpdateRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepInstrument.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_ArchiveRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepInstrument()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.ArchiveRecipeStepInstrument), testutils.ContextMatcher, exampleRecipeStepID, expected.ID).Return(nil)
			},
			map[string][]string{
				types.RecipeStepInstrumentArchivedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
					mealplanningkeys.RecipeStepIDKey,
					mealplanningkeys.RecipeStepInstrumentIDKey,
				},
			},
		)

		assert.NoError(t, rm.ArchiveRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
