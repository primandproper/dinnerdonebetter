package managers

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v11/filtering"

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
		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()

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

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepInstrument()
		fakeInput := fakes.BuildFakeRecipeStepInstrumentCreationRequestInput()
		fakeInput.Index = new(uint16(0))

		fakeValidPreparationInstrument := fakes.BuildFakeValidPreparationInstrument()

		db := &mealplanningmock.RepositoryMock{
			GetValidPreparationInstrumentFunc: func(_ context.Context, validPreparationInstrumentID string) (*types.ValidPreparationInstrument, error) {
				assert.Equal(t, *fakeInput.ValidPreparationInstrumentID, validPreparationInstrumentID)

				return fakeValidPreparationInstrument, nil
			},
			CreateRecipeStepInstrumentFunc: func(_ context.Context, _ string, _ *types.RecipeStepInstrumentDatabaseCreationInput) (*types.RecipeStepInstrument, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidPreparationInstrumentCalls(), 1)
		assert.Len(t, db.CreateRecipeStepInstrumentCalls(), 1)
	})
}

func TestRecipeManager_ReadRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
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

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		exampleRecipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()
		exampleInput := fakes.BuildFakeRecipeStepInstrumentUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
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

		require.NoError(t, rm.UpdateRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepInstrument.ID, exampleInput))

		assert.Len(t, db.GetRecipeStepInstrumentCalls(), 1)
		assert.Len(t, db.UpdateRecipeStepInstrumentCalls(), 1)
	})
}

func TestRecipeManager_ArchiveRecipeStepInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepInstrument()

		db := &mealplanningmock.RepositoryMock{
			ArchiveRecipeStepInstrumentFunc: func(_ context.Context, _ string, recipeStepID string, recipeStepInstrumentID string) error {
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, expected.ID, recipeStepInstrumentID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.ArchiveRecipeStepInstrument(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID))

		assert.Len(t, db.ArchiveRecipeStepInstrumentCalls(), 1)
	})
}
