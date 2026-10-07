package managers

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"

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
		exampleOwnerID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepVessel()
		fakeInput := fakes.BuildFakeRecipeStepVesselCreationRequestInput()
		fakeInput.Index = new(uint16(0))

		fakeValidPreparationVessel := fakes.BuildFakeValidPreparationVessel()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, true),
			GetValidPreparationVesselFunc: func(_ context.Context, validPreparationVesselID string) (*types.ValidPreparationVessel, error) {
				assert.Equal(t, *fakeInput.ValidPreparationVesselID, validPreparationVesselID)

				return fakeValidPreparationVessel, nil
			},
			CreateRecipeStepVesselFunc: func(_ context.Context, _ string, _ *types.RecipeStepVesselDatabaseCreationInput) (*types.RecipeStepVessel, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPreparationVesselCalls(), 1)
		assert.Len(t, db.CreateRecipeStepVesselCalls(), 1)
	})

	T.Run("with a caller who does not own the recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeRecipeStepVesselCreationRequestInput()
		fakeInput.Index = new(uint16(0))

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc: recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, false),
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID, fakeInput)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.RecipeStepExistsCalls())
		assert.Empty(t, db.CreateRecipeStepVesselCalls())
	})

	T.Run("with a step from another recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeRecipeStepVesselCreationRequestInput()
		fakeInput.Index = new(uint16(0))

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, false),
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID, fakeInput)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.CreateRecipeStepVesselCalls())
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
		exampleOwnerID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleRecipeStepVessel := fakes.BuildFakeRecipeStepVessel()
		exampleInput := fakes.BuildFakeRecipeStepVesselUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, true),
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

		require.NoError(t, rm.UpdateRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepVessel.ID, exampleOwnerID, exampleInput))

		assert.Len(t, db.GetRecipeStepVesselCalls(), 1)
		assert.Len(t, db.UpdateRecipeStepVesselCalls(), 1)
	})

	T.Run("with a caller who does not own the recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc: recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, false),
		}
		attachRepositoryToManager(rm, db)

		err := rm.UpdateRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID, fakes.BuildFakeRecipeStepVesselUpdateRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.RecipeStepExistsCalls())
		assert.Empty(t, db.GetRecipeStepVesselCalls())
		assert.Empty(t, db.UpdateRecipeStepVesselCalls())
	})

	T.Run("with a step from another recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, false),
		}
		attachRepositoryToManager(rm, db)

		err := rm.UpdateRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID, fakes.BuildFakeRecipeStepVesselUpdateRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.GetRecipeStepVesselCalls())
		assert.Empty(t, db.UpdateRecipeStepVesselCalls())
	})
}

func TestRecipeManager_ArchiveRecipeStepVessel(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepVessel()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, true),
			ArchiveRecipeStepVesselFunc: func(_ context.Context, _ string, recipeStepID string, recipeStepInstrumentID string) error {
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, expected.ID, recipeStepInstrumentID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.ArchiveRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID, exampleOwnerID))

		assert.Len(t, db.ArchiveRecipeStepVesselCalls(), 1)
	})

	T.Run("with a caller who does not own the recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc: recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, false),
		}
		attachRepositoryToManager(rm, db)

		err := rm.ArchiveRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.RecipeStepExistsCalls())
		assert.Empty(t, db.ArchiveRecipeStepVesselCalls())
	})

	T.Run("with a step from another recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, false),
		}
		attachRepositoryToManager(rm, db)

		err := rm.ArchiveRecipeStepVessel(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ArchiveRecipeStepVesselCalls())
	})
}
