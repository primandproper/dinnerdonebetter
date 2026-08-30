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

func TestRecipeManager_ListRecipeStepVessels(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipeStepVesselsList()
		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepVesselsFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepVessel], error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ListRecipeStepVessels(ctx, exampleRecipeID, exampleRecipeStepID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeStepVesselsCalls(), 1)
	})
}

func TestRecipeManager_CreateRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepVessel()
		fakeInput := fakes.BuildFakeRecipeStepVesselCreationRequestInput()
		fakeInput.Index = new(uint16(0))

		fakeValidPreparationVessel := fakes.BuildFakeValidPreparationVessel()

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationVesselFunc: func(_ context.Context, validPreparationVesselID string) (*types.ValidPreparationVessel, error) {
				assert.Equal(t, *fakeInput.ValidPreparationVesselID, validPreparationVesselID)

				return fakeValidPreparationVessel, nil
			},
			CreateRecipeStepVesselFunc: func(_ context.Context, _ string, _ *types.RecipeStepVesselDatabaseCreationInput) (*types.RecipeStepVessel, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPreparationVesselCalls(), 1)
		assert.Len(t, db.CreateRecipeStepVesselCalls(), 1)
	})
}

func TestRecipeManager_ReadRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepVessel()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepVesselFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepInstrumentID string) (*types.RecipeStepVessel, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, expected.ID, recipeStepInstrumentID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ReadRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeStepVesselCalls(), 1)
	})
}

func TestRecipeManager_UpdateRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleRecipeStepVessel := fakes.BuildFakeRecipeStepVessel()
		exampleInput := fakes.BuildFakeRecipeStepVesselUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepVesselFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepInstrumentID string) (*types.RecipeStepVessel, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, exampleRecipeStepVessel.ID, recipeStepInstrumentID)

				return exampleRecipeStepVessel, nil
			},
			UpdateRecipeStepVesselFunc: func(_ context.Context, _ string, _ *types.RecipeStepVessel) error {
				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.UpdateRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepVessel.ID, exampleInput))

		assert.Len(t, db.GetRecipeStepVesselCalls(), 1)
		assert.Len(t, db.UpdateRecipeStepVesselCalls(), 1)
	})
}

func TestRecipeManager_ArchiveRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepVessel()

		db := &mealplanningmock.RepositoryMock{
			ArchiveRecipeStepVesselFunc: func(_ context.Context, _ string, recipeStepID string, recipeStepInstrumentID string) error {
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, expected.ID, recipeStepInstrumentID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.ArchiveRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID))

		assert.Len(t, db.ArchiveRecipeStepVesselCalls(), 1)
	})
}
