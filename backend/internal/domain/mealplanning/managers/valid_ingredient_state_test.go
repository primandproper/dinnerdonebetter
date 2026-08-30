package managers

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidEnumerationManager_SearchValidIngredientStates(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientStatesList()
		exampleQuery := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			SearchForValidIngredientStatesFunc: func(_ context.Context, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientState], error) {
				assert.Equal(t, exampleQuery, query)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidIngredientStates(ctx, exampleQuery, false, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForValidIngredientStatesCalls(), 1)
	})
}

func TestValidEnumerationManager_ListValidIngredientStates(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientStatesList()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientStatesFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientState], error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ListValidIngredientStates(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientStatesCalls(), 1)
	})
}

func TestValidEnumerationManager_CreateValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientState()
		fakeInput := fakes.BuildFakeValidIngredientStateCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateValidIngredientStateFunc: func(_ context.Context, _ *types.ValidIngredientStateDatabaseCreationInput) (*types.ValidIngredientState, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.CreateValidIngredientState(ctx, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateValidIngredientStateCalls(), 1)
	})
}

func TestValidEnumerationManager_ReadValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientState()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientStateFunc: func(_ context.Context, validIngredientState string) (*types.ValidIngredientState, error) {
				assert.Equal(t, expected.ID, validIngredientState)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ReadValidIngredientState(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientStateCalls(), 1)
	})
}

func TestValidEnumerationManager_UpdateValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidIngredientState := fakes.BuildFakeValidIngredientState()
		exampleInput := fakes.BuildFakeValidIngredientStateUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientStateFunc: func(_ context.Context, validIngredientState string) (*types.ValidIngredientState, error) {
				assert.Equal(t, exampleValidIngredientState.ID, validIngredientState)

				return exampleValidIngredientState, nil
			},
			UpdateValidIngredientStateFunc: func(_ context.Context, _ *types.ValidIngredientState) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		result, err := mpm.UpdateValidIngredientState(ctx, exampleValidIngredientState.ID, exampleInput)
		assert.NotNil(t, result)
		require.NoError(t, err)

		assert.Len(t, db.GetValidIngredientStateCalls(), 2) // the manager re-reads the record after updating it
		assert.Len(t, db.UpdateValidIngredientStateCalls(), 1)
	})
}

func TestValidEnumerationManager_ArchiveValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientState()

		db := &mealplanningmock.RepositoryMock{
			ArchiveValidIngredientStateFunc: func(_ context.Context, validIngredientState string) error {
				assert.Equal(t, expected.ID, validIngredientState)

				return nil
			},
		}
		attachRepositoryToManager(vem, db)

		require.NoError(t, vem.ArchiveValidIngredientState(ctx, expected.ID))

		assert.Len(t, db.ArchiveValidIngredientStateCalls(), 1)
	})
}
