package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v6/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidEnumerationManager_ListValidIngredientPreparations(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientPreparationsList()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidIngredientPreparations), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := vem.ListValidIngredientPreparations(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_CreateValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientPreparation()
		fakeInput := fakes.BuildFakeValidIngredientPreparationCreationRequestInput()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.CreateValidIngredientPreparation), testutils.ContextMatcher, testutils.MatchType[*types.ValidIngredientPreparationDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.ValidIngredientPreparationCreatedServiceEventType: {mealplanningkeys.ValidIngredientPreparationIDKey},
			},
		)

		actual, err := vem.CreateValidIngredientPreparation(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ReadValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientPreparation()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidIngredientPreparation), testutils.ContextMatcher, expected.ID).Return(expected, nil)
			},
		)

		actual, err := vem.ReadValidIngredientPreparation(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_UpdateValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidIngredientPreparation := fakes.BuildFakeValidIngredientPreparation()
		exampleInput := fakes.BuildFakeValidIngredientPreparationUpdateRequestInput()

		expectations := setupExpectationsForValidEnumerationManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetValidIngredientPreparation), testutils.ContextMatcher, exampleValidIngredientPreparation.ID).Return(exampleValidIngredientPreparation, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateValidIngredientPreparation), testutils.ContextMatcher, testutils.MatchType[*types.ValidIngredientPreparation]()).Return(nil)
			},
			map[string][]string{
				types.ValidIngredientPreparationUpdatedServiceEventType: {mealplanningkeys.ValidIngredientPreparationIDKey},
			},
		)

		result, err := mpm.UpdateValidIngredientPreparation(ctx, exampleValidIngredientPreparation.ID, exampleInput)
		assert.NotNil(t, result)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ArchiveValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientPreparation()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.ArchiveValidIngredientPreparation), testutils.ContextMatcher, expected.ID).Return(nil)
			},
			map[string][]string{
				types.ValidIngredientPreparationArchivedServiceEventType: {mealplanningkeys.ValidIngredientPreparationIDKey},
			},
		)

		assert.NoError(t, vem.ArchiveValidIngredientPreparation(ctx, expected.ID))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_SearchValidIngredientPreparationsByIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientPreparationsList()
		exampleQuery := fakes.BuildFakeID()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidIngredientPreparationsForIngredient), testutils.ContextMatcher, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := vem.SearchValidIngredientPreparationsByIngredient(ctx, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_SearchValidIngredientPreparationsByPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientPreparationsList()
		exampleQuery := fakes.BuildFakeID()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidIngredientPreparationsForPreparation), testutils.ContextMatcher, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := vem.SearchValidIngredientPreparationsByPreparation(ctx, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
