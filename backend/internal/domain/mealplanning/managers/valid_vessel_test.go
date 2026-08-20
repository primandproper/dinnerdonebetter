package managers

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidEnumerationManager_SearchValidVessels(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidVesselsList()
		exampleQuery := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			SearchForValidVesselsFunc: func(_ context.Context, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidVessel], error) {
				assert.Equal(t, exampleQuery, query)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidVessels(ctx, exampleQuery, false, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForValidVesselsCalls(), 1)
	})
}

func TestValidEnumerationManager_ListValidVessels(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidVesselsList()

		db := &mealplanningmock.RepositoryMock{
			GetValidVesselsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidVessel], error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ListValidVessels(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidVesselsCalls(), 1)
	})
}

func TestValidEnumerationManager_CreateValidVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidVessel()
		fakeInput := fakes.BuildFakeValidVesselCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateValidVesselFunc: func(_ context.Context, _ *types.ValidVesselDatabaseCreationInput) (*types.ValidVessel, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.CreateValidVessel(ctx, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateValidVesselCalls(), 1)
	})
}

func TestValidEnumerationManager_ReadValidVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidVessel()

		db := &mealplanningmock.RepositoryMock{
			GetValidVesselFunc: func(_ context.Context, validVesselID string) (*types.ValidVessel, error) {
				assert.Equal(t, expected.ID, validVesselID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ReadValidVessel(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidVesselCalls(), 1)
	})
}

func TestValidEnumerationManager_RandomValidVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidVessel()

		db := &mealplanningmock.RepositoryMock{
			GetRandomValidVesselFunc: func(_ context.Context) (*types.ValidVessel, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.RandomValidVessel(ctx)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRandomValidVesselCalls(), 1)
	})
}

func TestValidEnumerationManager_UpdateValidVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidVessel := fakes.BuildFakeValidVessel()
		exampleInput := fakes.BuildFakeValidVesselUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetValidVesselFunc: func(_ context.Context, validVesselID string) (*types.ValidVessel, error) {
				assert.Equal(t, exampleValidVessel.ID, validVesselID)

				return exampleValidVessel, nil
			},
			UpdateValidVesselFunc: func(_ context.Context, _ *types.ValidVessel) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		result, err := mpm.UpdateValidVessel(ctx, exampleValidVessel.ID, exampleInput)
		assert.NotNil(t, result)
		require.NoError(t, err)

		assert.Len(t, db.GetValidVesselCalls(), 2) // the manager re-reads the record after updating it
		assert.Len(t, db.UpdateValidVesselCalls(), 1)
	})
}

func TestValidEnumerationManager_ArchiveValidVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidVessel()

		db := &mealplanningmock.RepositoryMock{
			ArchiveValidVesselFunc: func(_ context.Context, validVesselID string) error {
				assert.Equal(t, expected.ID, validVesselID)

				return nil
			},
		}
		attachRepositoryToManager(vem, db)

		require.NoError(t, vem.ArchiveValidVessel(ctx, expected.ID))

		assert.Len(t, db.ArchiveValidVesselCalls(), 1)
	})
}
