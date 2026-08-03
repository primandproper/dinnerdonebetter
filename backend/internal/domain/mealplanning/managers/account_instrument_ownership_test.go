package managers

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMealPlanningManager_ListAccountInstrumentOwnerships(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeAccountInstrumentOwnershipsList()
		exampleOwnerID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetAccountInstrumentOwnershipsFunc: func(_ context.Context, accountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.AccountInstrumentOwnership], error) {
				assert.Equal(t, exampleOwnerID, accountID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListAccountInstrumentOwnerships(ctx, exampleOwnerID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetAccountInstrumentOwnershipsCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			SearchForValidInstrumentsNotOwnedByAccountFunc: func(_ context.Context, accountID string, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidInstrument], error) {
				assert.Equal(t, exampleAccountID, accountID)
				assert.Equal(t, exampleQuery, query)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.SearchValidInstrumentsNotOwnedByAccount(ctx, exampleAccountID, exampleQuery, false, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForValidInstrumentsNotOwnedByAccountCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			CreateAccountInstrumentOwnershipFunc: func(_ context.Context, _ *types.AccountInstrumentOwnershipDatabaseCreationInput) (*types.AccountInstrumentOwnership, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateAccountInstrumentOwnership(ctx, fakeOwnerID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateAccountInstrumentOwnershipCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			GetAccountInstrumentOwnershipFunc: func(_ context.Context, accountInstrumentOwnershipID string, accountID string) (*types.AccountInstrumentOwnership, error) {
				assert.Equal(t, expected.ID, accountInstrumentOwnershipID)
				assert.Equal(t, ownerID, accountID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadAccountInstrumentOwnership(ctx, ownerID, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetAccountInstrumentOwnershipCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			GetAccountInstrumentOwnershipFunc: func(_ context.Context, accountInstrumentOwnershipID string, accountID string) (*types.AccountInstrumentOwnership, error) {
				assert.Equal(t, exampleAccountInstrumentOwnership.ID, accountInstrumentOwnershipID)
				assert.Equal(t, ownerID, accountID)

				return exampleAccountInstrumentOwnership, nil
			},
			UpdateAccountInstrumentOwnershipFunc: func(_ context.Context, _ *types.AccountInstrumentOwnership) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		require.NoError(t, mpm.UpdateAccountInstrumentOwnership(ctx, exampleAccountInstrumentOwnership.ID, ownerID, exampleInput))

		assert.Len(t, db.GetAccountInstrumentOwnershipCalls(), 1)
		assert.Len(t, db.UpdateAccountInstrumentOwnershipCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			ArchiveAccountInstrumentOwnershipFunc: func(_ context.Context, accountInstrumentOwnershipID string, accountID string) error {
				assert.Equal(t, expected.ID, accountInstrumentOwnershipID)
				assert.Equal(t, ownershipID, accountID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveAccountInstrumentOwnership(ctx, ownershipID, expected.ID)
		require.NoError(t, err)

		assert.Len(t, db.ArchiveAccountInstrumentOwnershipCalls(), 1)
	})
}
