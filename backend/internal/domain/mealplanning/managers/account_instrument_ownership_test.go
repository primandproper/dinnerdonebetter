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

func TestMealPlanningManager_ListAccountInstrumentOwnerships(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeAccountInstrumentOwnershipsList()
		exampleOwnerID := fakes.BuildFakeID()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetAccountInstrumentOwnerships), testutils.ContextMatcher, exampleOwnerID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.ListAccountInstrumentOwnerships(ctx, exampleOwnerID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_SearchValidInstrumentsNotOwnedByAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeValidInstrumentsList()
		exampleAccountID := fakes.BuildFakeID()
		exampleQuery := "knife"

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.SearchForValidInstrumentsNotOwnedByAccount), testutils.ContextMatcher, exampleAccountID, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.SearchValidInstrumentsNotOwnedByAccount(ctx, exampleAccountID, exampleQuery, false, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_CreateAccountInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		fakeOwnerID := fakes.BuildFakeID()
		expected := fakes.BuildFakeAccountInstrumentOwnership()
		fakeInput := fakes.BuildFakeAccountInstrumentOwnershipCreationRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.CreateAccountInstrumentOwnership), testutils.ContextMatcher, testutils.MatchType[*types.AccountInstrumentOwnershipDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.AccountInstrumentOwnershipCreatedServiceEventType: {mealplanningkeys.AccountInstrumentOwnershipIDKey},
			},
		)

		actual, err := mpm.CreateAccountInstrumentOwnership(ctx, fakeOwnerID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_ReadAccountInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		ownerID := fakes.BuildFakeID()
		expected := fakes.BuildFakeAccountInstrumentOwnership()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetAccountInstrumentOwnership), testutils.ContextMatcher, expected.ID, ownerID).Return(expected, nil)
			},
		)

		actual, err := mpm.ReadAccountInstrumentOwnership(ctx, ownerID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_UpdateAccountInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleAccountInstrumentOwnership := fakes.BuildFakeAccountInstrumentOwnership()
		ownerID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeAccountInstrumentOwnershipUpdateRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetAccountInstrumentOwnership), testutils.ContextMatcher, exampleAccountInstrumentOwnership.ID, ownerID).Return(exampleAccountInstrumentOwnership, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateAccountInstrumentOwnership), testutils.ContextMatcher, testutils.MatchType[*types.AccountInstrumentOwnership]()).Return(nil)
			},
			map[string][]string{
				types.AccountInstrumentOwnershipUpdatedServiceEventType: {
					mealplanningkeys.AccountInstrumentOwnershipIDKey,
				},
			},
		)

		assert.NoError(t, mpm.UpdateAccountInstrumentOwnership(ctx, exampleAccountInstrumentOwnership.ID, ownerID, exampleInput))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_ArchiveAccountInstrumentOwnership(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		ownershipID := fakes.BuildFakeID()
		expected := fakes.BuildFakeAccountInstrumentOwnership()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.ArchiveAccountInstrumentOwnership), testutils.ContextMatcher, expected.ID, ownershipID).Return(nil)
			},
			map[string][]string{
				types.AccountInstrumentOwnershipArchivedServiceEventType: {
					mealplanningkeys.AccountInstrumentOwnershipIDKey,
				},
			},
		)

		err := mpm.ArchiveAccountInstrumentOwnership(ctx, ownershipID, expected.ID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
