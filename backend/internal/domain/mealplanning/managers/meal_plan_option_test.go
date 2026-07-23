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

func TestMealPlanningManager_ListMealPlanOptions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanOptionsList()
		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanOptions), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.ListMealPlanOptions(ctx, exampleMealPlanID, exampleMealPlanEventID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_CreateMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanOption()
		fakeInput := fakes.BuildFakeMealPlanOptionCreationRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.CreateMealPlanOption), testutils.ContextMatcher, testutils.MatchType[*types.MealPlanOptionDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.MealPlanOptionCreatedServiceEventType: {mealplanningkeys.MealPlanOptionIDKey},
			},
		)

		actual, err := mpm.CreateMealPlanOption(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})

	T.Run("with inline selections", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanOption()
		fakeInput := fakes.BuildFakeMealPlanOptionCreationRequestInput()
		fakeInput.Selections = []*types.MealPlanRecipeOptionSelectionCreationRequestInput{
			fakes.BuildFakeMealPlanRecipeOptionSelectionCreationRequestInput(),
			fakes.BuildFakeMealPlanRecipeOptionSelectionCreationRequestInput(),
		}

		createdSelection := fakes.BuildFakeMealPlanRecipeOptionSelection()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.CreateMealPlanOption), testutils.ContextMatcher, testutils.MatchType[*types.MealPlanOptionDatabaseCreationInput]()).Return(expected, nil)
				db.On(reflection.GetMethodName(mpm.db.CreateMealPlanRecipeOptionSelection), testutils.ContextMatcher, mock.MatchedBy(func(in *types.MealPlanRecipeOptionSelectionDatabaseCreationInput) bool {
					return in.BelongsToMealPlanOption == expected.ID
				})).Return(createdSelection, nil).Times(2)
			},
			map[string][]string{
				types.MealPlanOptionCreatedServiceEventType: {mealplanningkeys.MealPlanOptionIDKey},
			},
		)

		actual, err := mpm.CreateMealPlanOption(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_CreateMealPlanOptionWithEventID(T *testing.T) {
	T.Parallel()

	T.Run("returns ErrDuplicateMealPlanOption when MealExistsAsOptionInEvent returns true", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		eventID := fakes.BuildFakeID()
		fakeInput := fakes.BuildFakeMealPlanOptionCreationRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.MealExistsAsOptionInEvent), testutils.ContextMatcher, eventID, fakeInput.MealID).Return(true, nil)
			},
		)

		actual, err := mpm.CreateMealPlanOptionWithEventID(ctx, eventID, fakeInput)
		assert.ErrorIs(t, err, types.ErrDuplicateMealPlanOption)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_ReadMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanOption()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanOption), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID, expected.ID).Return(expected, nil)
			},
		)

		actual, err := mpm.ReadMealPlanOption(ctx, exampleMealPlanID, exampleMealPlanEventID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_MealPlanOptionBelongsToAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanOptionID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.MealPlanOptionBelongsToAccount), testutils.ContextMatcher, exampleMealPlanOptionID, exampleAccountID).Return(true, nil)
			},
		)

		belongs, err := mpm.MealPlanOptionBelongsToAccount(ctx, exampleMealPlanOptionID, exampleAccountID)
		assert.NoError(t, err)
		assert.True(t, belongs)

		mock.AssertExpectationsForObjects(t, expectations...)
	})

	T.Run("with option belonging to another account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanOptionID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.MealPlanOptionBelongsToAccount), testutils.ContextMatcher, exampleMealPlanOptionID, exampleAccountID).Return(false, nil)
			},
		)

		belongs, err := mpm.MealPlanOptionBelongsToAccount(ctx, exampleMealPlanOptionID, exampleAccountID)
		assert.NoError(t, err)
		assert.False(t, belongs)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_UpdateMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanOption := fakes.BuildFakeMealPlanOption()
		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanOptionUpdateRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanOption), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOption.ID).Return(exampleMealPlanOption, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateMealPlanOption), testutils.ContextMatcher, testutils.MatchType[*types.MealPlanOption]()).Return(nil)
			},
			map[string][]string{
				types.MealPlanOptionUpdatedServiceEventType: {
					mealplanningkeys.MealPlanIDKey,
					mealplanningkeys.MealPlanEventIDKey,
					mealplanningkeys.MealPlanOptionIDKey,
				},
			},
		)

		assert.NoError(t, mpm.UpdateMealPlanOption(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOption.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_ArchiveMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanID := fakes.BuildFakeID()
		mealPlanEventID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanOption()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.ArchiveMealPlanOption), testutils.ContextMatcher, mealPlanID, mealPlanEventID, expected.ID).Return(nil)
			},
			map[string][]string{
				types.MealPlanOptionArchivedServiceEventType: {
					mealplanningkeys.MealPlanIDKey,
					mealplanningkeys.MealPlanEventIDKey,
					mealplanningkeys.MealPlanOptionIDKey,
				},
			},
		)

		err := mpm.ArchiveMealPlanOption(ctx, mealPlanID, mealPlanEventID, expected.ID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
