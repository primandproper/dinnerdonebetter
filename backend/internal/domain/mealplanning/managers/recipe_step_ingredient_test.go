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

func TestRecipeManager_ListRecipeStepIngredients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipeStepIngredientsList()
		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepIngredientsFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepIngredient], error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ListRecipeStepIngredients(ctx, exampleRecipeID, exampleRecipeStepID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeStepIngredientsCalls(), 1)
	})
}

func TestRecipeManager_CreateRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepIngredient()
		fakeInput := fakes.BuildFakeRecipeStepIngredientCreationRequestInput()
		fakeInput.Index = new(uint16(0))

		fakeValidIngredientPreparation := fakes.BuildFakeValidIngredientPreparation()
		fakeValidIngredientMeasurementUnit := fakes.BuildFakeValidIngredientMeasurementUnit()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, true),
			GetValidIngredientPreparationFunc: func(_ context.Context, validIngredientPreparationID string) (*types.ValidIngredientPreparation, error) {
				assert.Equal(t, *fakeInput.ValidIngredientPreparationID, validIngredientPreparationID)

				return fakeValidIngredientPreparation, nil
			},
			GetValidIngredientMeasurementUnitFunc: func(_ context.Context, validIngredientMeasurementUnitID string) (*types.ValidIngredientMeasurementUnit, error) {
				assert.Equal(t, *fakeInput.ValidIngredientMeasurementUnitID, validIngredientMeasurementUnitID)

				return fakeValidIngredientMeasurementUnit, nil
			},
			CreateRecipeStepIngredientFunc: func(_ context.Context, _ string, _ *types.RecipeStepIngredientDatabaseCreationInput) (*types.RecipeStepIngredient, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientPreparationCalls(), 1)
		assert.Len(t, db.GetValidIngredientMeasurementUnitCalls(), 1)
		assert.Len(t, db.CreateRecipeStepIngredientCalls(), 1)
	})

	T.Run("with a caller who does not own the recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeRecipeStepIngredientCreationRequestInput()
		fakeInput.Index = new(uint16(0))

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc: recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, false),
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID, fakeInput)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.RecipeStepExistsCalls())
		assert.Empty(t, db.CreateRecipeStepIngredientCalls())
	})

	T.Run("with a step from another recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeRecipeStepIngredientCreationRequestInput()
		fakeInput.Index = new(uint16(0))

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, false),
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID, fakeInput)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.CreateRecipeStepIngredientCalls())
	})

	T.Run("with a product from a recipe that draws on this one", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipe := fakes.BuildFakeRecipe()
		exampleRecipeStepID := exampleRecipe.Steps[0].ID
		exampleOwnerID := fake.BuildFakeID()

		// The recipe the new ingredient draws its product from already draws on this one.
		productRecipe := fakes.BuildFakeRecipe()
		productRecipe.Steps[0].Ingredients[0].RecipeStepProductRecipeID = &exampleRecipe.ID

		fakeInput := fakes.BuildFakeRecipeStepIngredientCreationRequestInputForRecipeStepProduct()
		fakeInput.Index = new(uint16(0))
		fakeInput.RecipeStepProductRecipeID = &productRecipe.ID

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipe.ID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipe.ID, exampleRecipeStepID, true),
			GetRecipeFunc: func(_ context.Context, recipeID string) (*types.Recipe, error) {
				switch recipeID {
				case exampleRecipe.ID:
					return exampleRecipe, nil
				case productRecipe.ID:
					return productRecipe, nil
				default:
					t.Errorf("unexpected recipe read: %s", recipeID)
					return nil, sql.ErrNoRows
				}
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepIngredient(ctx, exampleRecipe.ID, exampleRecipeStepID, exampleOwnerID, fakeInput)
		require.ErrorIs(t, err, types.ErrInvalidRecipeInput)
		require.ErrorContains(t, err, exampleRecipe.ID)
		assert.Nil(t, actual)

		assert.Len(t, db.GetRecipeCalls(), 2)
		assert.Empty(t, db.CreateRecipeStepIngredientCalls())
	})

	T.Run("with a product from a recipe that does not draw on this one", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipe := fakes.BuildFakeRecipe()
		exampleRecipeStepID := exampleRecipe.Steps[0].ID
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepIngredient()

		productRecipe := fakes.BuildFakeRecipe()

		fakeInput := fakes.BuildFakeRecipeStepIngredientCreationRequestInputForRecipeStepProduct()
		fakeInput.Index = new(uint16(0))
		fakeInput.RecipeStepProductRecipeID = &productRecipe.ID

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipe.ID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipe.ID, exampleRecipeStepID, true),
			GetRecipeFunc: func(_ context.Context, recipeID string) (*types.Recipe, error) {
				switch recipeID {
				case exampleRecipe.ID:
					return exampleRecipe, nil
				case productRecipe.ID:
					return productRecipe, nil
				default:
					t.Errorf("unexpected recipe read: %s", recipeID)
					return nil, sql.ErrNoRows
				}
			},
			CreateRecipeStepIngredientFunc: func(_ context.Context, recipeID string, input *types.RecipeStepIngredientDatabaseCreationInput) (*types.RecipeStepIngredient, error) {
				assert.Equal(t, exampleRecipe.ID, recipeID)
				assert.Equal(t, fakeInput.RecipeStepProductRecipeID, input.RecipeStepProductRecipeID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepIngredient(ctx, exampleRecipe.ID, exampleRecipeStepID, exampleOwnerID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeCalls(), 2)
		assert.Len(t, db.CreateRecipeStepIngredientCalls(), 1)
	})
}

func TestRecipeManager_ReadRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepIngredient()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepIngredientFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepIngredientID string) (*types.RecipeStepIngredient, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, expected.ID, recipeStepIngredientID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ReadRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeStepIngredientCalls(), 1)
	})
}

func TestRecipeManager_UpdateRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleRecipeStepIngredient := fakes.BuildFakeRecipeStepIngredient()
		exampleInput := fakes.BuildFakeRecipeStepIngredientUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, true),
			GetRecipeStepIngredientFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepIngredientID string) (*types.RecipeStepIngredient, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, exampleRecipeStepIngredient.ID, recipeStepIngredientID)

				return exampleRecipeStepIngredient, nil
			},
			UpdateRecipeStepIngredientFunc: func(_ context.Context, _ string, _ *types.RecipeStepIngredient) error {
				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.UpdateRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepIngredient.ID, exampleOwnerID, exampleInput))

		assert.Len(t, db.GetRecipeStepIngredientCalls(), 1)
		assert.Len(t, db.UpdateRecipeStepIngredientCalls(), 1)
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

		err := rm.UpdateRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID, fakes.BuildFakeRecipeStepIngredientUpdateRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.RecipeStepExistsCalls())
		assert.Empty(t, db.GetRecipeStepIngredientCalls())
		assert.Empty(t, db.UpdateRecipeStepIngredientCalls())
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

		err := rm.UpdateRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID, fakes.BuildFakeRecipeStepIngredientUpdateRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.GetRecipeStepIngredientCalls())
		assert.Empty(t, db.UpdateRecipeStepIngredientCalls())
	})

	T.Run("with a product from a recipe that draws on this one", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipe := fakes.BuildFakeRecipe()
		exampleRecipeStep := exampleRecipe.Steps[0]
		exampleRecipeStepIngredient := exampleRecipeStep.Ingredients[0]
		exampleOwnerID := fake.BuildFakeID()

		// The recipe the ingredient is changed to draw its product from already draws on this one.
		productRecipe := fakes.BuildFakeRecipe()
		productRecipe.Steps[0].Ingredients[0].RecipeStepProductRecipeID = &exampleRecipe.ID

		exampleInput := fakes.BuildFakeRecipeStepIngredientUpdateRequestInput()
		exampleInput.RecipeStepProductRecipeID = &productRecipe.ID

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipe.ID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipe.ID, exampleRecipeStep.ID, true),
			GetRecipeStepIngredientFunc: func(_ context.Context, recipeID, recipeStepID, recipeStepIngredientID string) (*types.RecipeStepIngredient, error) {
				assert.Equal(t, exampleRecipe.ID, recipeID)
				assert.Equal(t, exampleRecipeStep.ID, recipeStepID)
				assert.Equal(t, exampleRecipeStepIngredient.ID, recipeStepIngredientID)

				return exampleRecipeStepIngredient, nil
			},
			GetRecipeFunc: func(_ context.Context, recipeID string) (*types.Recipe, error) {
				switch recipeID {
				case exampleRecipe.ID:
					return exampleRecipe, nil
				case productRecipe.ID:
					return productRecipe, nil
				default:
					t.Errorf("unexpected recipe read: %s", recipeID)
					return nil, sql.ErrNoRows
				}
			},
		}
		attachRepositoryToManager(rm, db)

		err := rm.UpdateRecipeStepIngredient(ctx, exampleRecipe.ID, exampleRecipeStep.ID, exampleRecipeStepIngredient.ID, exampleOwnerID, exampleInput)
		require.ErrorIs(t, err, types.ErrInvalidRecipeInput)
		require.ErrorContains(t, err, exampleRecipe.ID)

		assert.Empty(t, db.UpdateRecipeStepIngredientCalls())
	})
}

func TestRecipeManager_ArchiveRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepIngredient()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, true),
			ArchiveRecipeStepIngredientFunc: func(_ context.Context, _ string, recipeStepID string, recipeStepIngredientID string) error {
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, expected.ID, recipeStepIngredientID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.ArchiveRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID, exampleOwnerID))

		assert.Len(t, db.ArchiveRecipeStepIngredientCalls(), 1)
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

		err := rm.ArchiveRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.RecipeStepExistsCalls())
		assert.Empty(t, db.ArchiveRecipeStepIngredientCalls())
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

		err := rm.ArchiveRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ArchiveRecipeStepIngredientCalls())
	})
}
