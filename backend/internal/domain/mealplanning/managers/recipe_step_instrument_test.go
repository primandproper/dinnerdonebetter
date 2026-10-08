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

func TestRecipeManager_ListRecipeStepInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipeStepInstrumentsList()
		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepInstrumentsFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepInstrument], error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ListRecipeStepInstruments(ctx, exampleRecipeID, exampleRecipeStepID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeStepInstrumentsCalls(), 1)
	})
}

func TestRecipeManager_CreateRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepInstrument()
		fakeInput := fakes.BuildFakeRecipeStepInstrumentCreationRequestInput()
		fakeInput.Index = new(uint16(0))

		fakeValidPreparationInstrument := fakes.BuildFakeValidPreparationInstrument()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, true),
			GetValidPreparationInstrumentFunc: func(_ context.Context, validPreparationInstrumentID string) (*types.ValidPreparationInstrument, error) {
				assert.Equal(t, *fakeInput.ValidPreparationInstrumentID, validPreparationInstrumentID)

				return fakeValidPreparationInstrument, nil
			},
			CreateRecipeStepInstrumentFunc: func(_ context.Context, _ string, _ *types.RecipeStepInstrumentDatabaseCreationInput) (*types.RecipeStepInstrument, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPreparationInstrumentCalls(), 1)
		assert.Len(t, db.CreateRecipeStepInstrumentCalls(), 1)
	})

	T.Run("with a caller who does not own the recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeRecipeStepInstrumentCreationRequestInput()
		fakeInput.Index = new(uint16(0))

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc: recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, false),
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID, fakeInput)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.RecipeStepExistsCalls())
		assert.Empty(t, db.CreateRecipeStepInstrumentCalls())
	})

	T.Run("with a step from another recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeRecipeStepInstrumentCreationRequestInput()
		fakeInput.Index = new(uint16(0))

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, false),
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID, fakeInput)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.CreateRecipeStepInstrumentCalls())
	})
}

func TestRecipeManager_ReadRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepInstrument()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepInstrumentFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepInstrumentID string) (*types.RecipeStepInstrument, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, expected.ID, recipeStepInstrumentID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ReadRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeStepInstrumentCalls(), 1)
	})
}

func TestRecipeManager_UpdateRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleRecipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()
		exampleInput := fakes.BuildFakeRecipeStepInstrumentUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, true),
			GetRecipeStepInstrumentFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepInstrumentID string) (*types.RecipeStepInstrument, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, exampleRecipeStepInstrument.ID, recipeStepInstrumentID)

				return exampleRecipeStepInstrument, nil
			},
			UpdateRecipeStepInstrumentFunc: func(_ context.Context, _ string, _ *types.RecipeStepInstrument) error {
				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.UpdateRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepInstrument.ID, exampleOwnerID, exampleInput))

		assert.Len(t, db.GetRecipeStepInstrumentCalls(), 1)
		assert.Len(t, db.UpdateRecipeStepInstrumentCalls(), 1)
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

		err := rm.UpdateRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID, fakes.BuildFakeRecipeStepInstrumentUpdateRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.RecipeStepExistsCalls())
		assert.Empty(t, db.GetRecipeStepInstrumentCalls())
		assert.Empty(t, db.UpdateRecipeStepInstrumentCalls())
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

		err := rm.UpdateRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID, fakes.BuildFakeRecipeStepInstrumentUpdateRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.GetRecipeStepInstrumentCalls())
		assert.Empty(t, db.UpdateRecipeStepInstrumentCalls())
	})
}

func TestRecipeManager_ArchiveRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepInstrument()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, true),
			ArchiveRecipeStepInstrumentFunc: func(_ context.Context, _ string, recipeStepID string, recipeStepInstrumentID string) error {
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, expected.ID, recipeStepInstrumentID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.ArchiveRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID, exampleOwnerID))

		assert.Len(t, db.ArchiveRecipeStepInstrumentCalls(), 1)
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

		err := rm.ArchiveRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.RecipeStepExistsCalls())
		assert.Empty(t, db.ArchiveRecipeStepInstrumentCalls())
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

		err := rm.ArchiveRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ArchiveRecipeStepInstrumentCalls())
	})
}
