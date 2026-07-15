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

func TestRecipeManager_ListRecipeStepVessels(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipeStepVesselsList()
		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipeStepVessels), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := rm.ListRecipeStepVessels(ctx, exampleRecipeID, exampleRecipeStepID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_CreateRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepVessel()
		fakeInput := fakes.BuildFakeRecipeStepVesselCreationRequestInput()
		fakeInput.Index = new(uint16(0))

		fakeValidPreparationVessel := fakes.BuildFakeValidPreparationVessel()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetValidPreparationVessel), testutils.ContextMatcher, *fakeInput.ValidPreparationVesselID).Return(fakeValidPreparationVessel, nil)
				db.On(reflection.GetMethodName(rm.db.CreateRecipeStepVessel), testutils.ContextMatcher, testutils.MatchType[*types.RecipeStepVesselDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.RecipeStepVesselCreatedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
					mealplanningkeys.RecipeStepIDKey,
					mealplanningkeys.RecipeStepVesselIDKey,
				},
			},
		)

		actual, err := rm.CreateRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_ReadRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepVessel()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipeStepVessel), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, expected.ID).Return(expected, nil)
			},
		)

		actual, err := rm.ReadRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_UpdateRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		exampleRecipeStepVessel := fakes.BuildFakeRecipeStepVessel()
		exampleInput := fakes.BuildFakeRecipeStepVesselUpdateRequestInput()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipeStepVessel), testutils.ContextMatcher, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepVessel.ID).Return(exampleRecipeStepVessel, nil)
				db.On(reflection.GetMethodName(rm.db.UpdateRecipeStepVessel), testutils.ContextMatcher, testutils.MatchType[*types.RecipeStepVessel]()).Return(nil)
			},
			map[string][]string{
				types.RecipeStepVesselUpdatedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
					mealplanningkeys.RecipeStepIDKey,
					mealplanningkeys.RecipeStepVesselIDKey,
				},
			},
		)

		assert.NoError(t, rm.UpdateRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepVessel.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_ArchiveRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepVessel()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.ArchiveRecipeStepVessel), testutils.ContextMatcher, exampleRecipeStepID, expected.ID).Return(nil)
			},
			map[string][]string{
				types.RecipeStepVesselArchivedServiceEventType: {
					mealplanningkeys.RecipeIDKey,
					mealplanningkeys.RecipeStepIDKey,
					mealplanningkeys.RecipeStepVesselIDKey,
				},
			},
		)

		assert.NoError(t, rm.ArchiveRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
