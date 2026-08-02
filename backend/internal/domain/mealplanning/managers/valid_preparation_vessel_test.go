package managers

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/stretchr/testify/assert"
)

func TestValidEnumerationManager_ListValidPreparationVessels(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationVesselsList()

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationVesselsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationVessel], error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ListValidPreparationVessels(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPreparationVesselsCalls(), 1)
	})
}

func TestValidEnumerationManager_CreateValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationVessel()
		fakeInput := fakes.BuildFakeValidPreparationVesselCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateValidPreparationVesselFunc: func(_ context.Context, _ *types.ValidPreparationVesselDatabaseCreationInput) (*types.ValidPreparationVessel, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.CreateValidPreparationVessel(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateValidPreparationVesselCalls(), 1)
	})
}

func TestValidEnumerationManager_ReadValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationVessel()

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationVesselFunc: func(_ context.Context, validPreparationVesselID string) (*types.ValidPreparationVessel, error) {
				assert.Equal(t, expected.ID, validPreparationVesselID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.ReadValidPreparationVessel(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPreparationVesselCalls(), 1)
	})
}

func TestValidEnumerationManager_UpdateValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidPreparationVessel := fakes.BuildFakeValidPreparationVessel()
		exampleInput := fakes.BuildFakeValidPreparationVesselUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationVesselFunc: func(_ context.Context, validPreparationVesselID string) (*types.ValidPreparationVessel, error) {
				assert.Equal(t, exampleValidPreparationVessel.ID, validPreparationVesselID)

				return exampleValidPreparationVessel, nil
			},
			UpdateValidPreparationVesselFunc: func(_ context.Context, _ *types.ValidPreparationVessel) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		result, err := mpm.UpdateValidPreparationVessel(ctx, exampleValidPreparationVessel.ID, exampleInput)
		assert.NotNil(t, result)
		assert.NoError(t, err)

		assert.Len(t, db.GetValidPreparationVesselCalls(), 2) // the manager re-reads the record after updating it
		assert.Len(t, db.UpdateValidPreparationVesselCalls(), 1)
	})
}

func TestValidEnumerationManager_ArchiveValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationVessel()

		db := &mealplanningmock.RepositoryMock{
			ArchiveValidPreparationVesselFunc: func(_ context.Context, validPreparationVesselID string) error {
				assert.Equal(t, expected.ID, validPreparationVesselID)

				return nil
			},
		}
		attachRepositoryToManager(vem, db)

		assert.NoError(t, vem.ArchiveValidPreparationVessel(ctx, expected.ID))

		assert.Len(t, db.ArchiveValidPreparationVesselCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidPreparationVesselsByPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationVesselsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationVesselsForPreparationFunc: func(_ context.Context, preparationID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationVessel], error) {
				assert.Equal(t, exampleQuery, preparationID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidPreparationVesselsByPreparation(ctx, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPreparationVesselsForPreparationCalls(), 1)
	})
}

func TestValidEnumerationManager_SearchValidPreparationVesselsByVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationVesselsList()
		exampleQuery := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationVesselsForVesselFunc: func(_ context.Context, instrumentID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationVessel], error) {
				assert.Equal(t, exampleQuery, instrumentID)

				return expected, nil
			},
		}
		attachRepositoryToManager(vem, db)

		actual, err := vem.SearchValidPreparationVesselsByVessel(ctx, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPreparationVesselsForVesselCalls(), 1)
	})
}
