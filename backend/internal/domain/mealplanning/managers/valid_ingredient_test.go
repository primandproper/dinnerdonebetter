package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v4/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidEnumerationManager_SearchValidIngredients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientsList()
		exampleQuery := fakes.BuildFakeID()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.SearchForValidIngredients), testutils.ContextMatcher, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)

				for _, ing := range expected.Data {
					db.On(reflection.GetMethodName(vem.db.GetIngredientMediaByIngredient), testutils.ContextMatcher, ing.ID).Return([]*types.IngredientMediaRow{}, nil)
				}
			},
		)

		actual, err := vem.SearchValidIngredients(ctx, exampleQuery, false, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ListValidIngredients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientsList()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidIngredients), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(expected, nil)

				for _, ing := range expected.Data {
					db.On(reflection.GetMethodName(vem.db.GetIngredientMediaByIngredient), testutils.ContextMatcher, ing.ID).Return([]*types.IngredientMediaRow{}, nil)
				}
			},
		)

		actual, err := vem.ListValidIngredients(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_CreateValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredient()
		fakeInput := fakes.BuildFakeValidIngredientCreationRequestInput()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.CreateValidIngredient), testutils.ContextMatcher, testutils.MatchType[*types.ValidIngredientDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.ValidIngredientCreatedServiceEventType: {mealplanningkeys.ValidIngredientIDKey},
			},
		)

		actual, err := vem.CreateValidIngredient(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ReadValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredient()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidIngredient), testutils.ContextMatcher, expected.ID).Return(expected, nil)

				db.On(reflection.GetMethodName(vem.db.GetIngredientMediaByIngredient), testutils.ContextMatcher, expected.ID).Return([]*types.IngredientMediaRow{}, nil)
			},
		)

		actual, err := vem.ReadValidIngredient(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_RandomValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredient()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetRandomValidIngredient), testutils.ContextMatcher).Return(expected, nil)

				db.On(reflection.GetMethodName(vem.db.GetIngredientMediaByIngredient), testutils.ContextMatcher, expected.ID).Return([]*types.IngredientMediaRow{}, nil)
			},
		)

		actual, err := vem.RandomValidIngredient(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_UpdateValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidIngredient := fakes.BuildFakeValidIngredient()
		exampleInput := fakes.BuildFakeValidIngredientUpdateRequestInput()

		expectations := setupExpectationsForValidEnumerationManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetValidIngredient), testutils.ContextMatcher, exampleValidIngredient.ID).Return(exampleValidIngredient, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateValidIngredient), testutils.ContextMatcher, testutils.MatchType[*types.ValidIngredient]()).Return(nil)

				db.On(reflection.GetMethodName(mpm.db.GetIngredientMediaByIngredient), testutils.ContextMatcher, exampleValidIngredient.ID).Return([]*types.IngredientMediaRow{}, nil)
			},
			map[string][]string{
				types.ValidIngredientUpdatedServiceEventType: {mealplanningkeys.ValidIngredientIDKey},
			},
		)

		result, err := mpm.UpdateValidIngredient(ctx, exampleValidIngredient.ID, exampleInput)
		assert.NotNil(t, result)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ArchiveValidIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredient()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.ArchiveValidIngredient), testutils.ContextMatcher, expected.ID).Return(nil)
			},
			map[string][]string{
				types.ValidIngredientArchivedServiceEventType: {mealplanningkeys.ValidIngredientIDKey},
			},
		)

		assert.NoError(t, vem.ArchiveValidIngredient(ctx, expected.ID))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_SearchValidIngredientsByPreparationAndIngredientName(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientsList()
		preparationID := fakes.BuildFakeID()
		exampleQuery := fakes.BuildFakeID()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.SearchForValidIngredientsForPreparation), testutils.ContextMatcher, preparationID, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)

				for _, ing := range expected.Data {
					db.On(reflection.GetMethodName(vem.db.GetIngredientMediaByIngredient), testutils.ContextMatcher, ing.ID).Return([]*types.IngredientMediaRow{}, nil)
				}
			},
		)

		actual, err := vem.SearchValidIngredientsByPreparationAndIngredientName(ctx, preparationID, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
