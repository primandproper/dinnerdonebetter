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

func TestValidEnumerationManager_SearchValidIngredientGroups(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientGroupsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			SearchForValidIngredientGroupsFunc: func(_ context.Context, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientGroup], error) {
				assert.Equal(t, exampleQuery, query)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidIngredientGroups(ctx, exampleQuery, false, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForValidIngredientGroupsCalls(), 1)
	})
}

func TestValidEnumerationManager_ListValidIngredientGroups(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientGroupsList()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientGroupsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientGroup], error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ListValidIngredientGroups(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientGroupsCalls(), 1)
	})
}

func TestValidEnumerationManager_CreateValidIngredientGroup(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientGroup()
		fakeInput := fakes.BuildFakeValidIngredientGroupCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateValidIngredientGroupFunc: func(_ context.Context, _ *types.ValidIngredientGroupDatabaseCreationInput) (*types.ValidIngredientGroup, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.CreateValidIngredientGroup(ctx, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateValidIngredientGroupCalls(), 1)
	})
}

func TestValidEnumerationManager_ReadValidIngredientGroup(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientGroup()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientGroupFunc: func(_ context.Context, validIngredientID string) (*types.ValidIngredientGroup, error) {
				assert.Equal(t, expected.ID, validIngredientID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ReadValidIngredientGroup(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientGroupCalls(), 1)
	})
}

func TestValidEnumerationManager_UpdateValidIngredientGroup(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidIngredientGroup := fakes.BuildFakeValidIngredientGroup()
		exampleInput := fakes.BuildFakeValidIngredientGroupUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientGroupFunc: func(_ context.Context, validIngredientID string) (*types.ValidIngredientGroup, error) {
				assert.Equal(t, exampleValidIngredientGroup.ID, validIngredientID)

				return exampleValidIngredientGroup, nil
			},
			UpdateValidIngredientGroupFunc: func(_ context.Context, _ *types.ValidIngredientGroup) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		result, err := mpm.UpdateValidIngredientGroup(ctx, exampleValidIngredientGroup.ID, exampleInput)
		assert.NotNil(t, result)
		require.NoError(t, err)

		assert.Len(t, db.GetValidIngredientGroupCalls(), 2) // the manager re-reads the record after updating it
		assert.Len(t, db.UpdateValidIngredientGroupCalls(), 1)
	})
}

func TestValidEnumerationManager_ArchiveValidIngredientGroup(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientGroup()

		db := &mealplanningmock.RepositoryMock{
			ArchiveValidIngredientGroupFunc: func(_ context.Context, validIngredientID string) error {
				assert.Equal(t, expected.ID, validIngredientID)

				return nil
			},
		}
		attachRepositoryToManager(vem, db)

		require.NoError(t, vem.ArchiveValidIngredientGroup(ctx, expected.ID))

		assert.Len(t, db.ArchiveValidIngredientGroupCalls(), 1)
	})
}
