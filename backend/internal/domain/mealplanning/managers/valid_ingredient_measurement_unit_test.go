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

func TestValidEnumerationManager_ListValidIngredientMeasurementUnits(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientMeasurementUnitsList()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidIngredientMeasurementUnits), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := vem.ListValidIngredientMeasurementUnits(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_CreateValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientMeasurementUnit()
		fakeInput := fakes.BuildFakeValidIngredientMeasurementUnitCreationRequestInput()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.CreateValidIngredientMeasurementUnit), testutils.ContextMatcher, testutils.MatchType[*types.ValidIngredientMeasurementUnitDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.ValidIngredientMeasurementUnitCreatedServiceEventType: {mealplanningkeys.ValidIngredientMeasurementUnitIDKey},
			},
		)

		actual, err := vem.CreateValidIngredientMeasurementUnit(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ReadValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientMeasurementUnit()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidIngredientMeasurementUnit), testutils.ContextMatcher, expected.ID).Return(expected, nil)
			},
		)

		actual, err := vem.ReadValidIngredientMeasurementUnit(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_UpdateValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidIngredientMeasurementUnit := fakes.BuildFakeValidIngredientMeasurementUnit()
		exampleInput := fakes.BuildFakeValidIngredientMeasurementUnitUpdateRequestInput()

		expectations := setupExpectationsForValidEnumerationManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetValidIngredientMeasurementUnit), testutils.ContextMatcher, exampleValidIngredientMeasurementUnit.ID).Return(exampleValidIngredientMeasurementUnit, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateValidIngredientMeasurementUnit), testutils.ContextMatcher, testutils.MatchType[*types.ValidIngredientMeasurementUnit]()).Return(nil)
			},
			map[string][]string{
				types.ValidIngredientMeasurementUnitUpdatedServiceEventType: {mealplanningkeys.ValidIngredientMeasurementUnitIDKey},
			},
		)

		result, err := mpm.UpdateValidIngredientMeasurementUnit(ctx, exampleValidIngredientMeasurementUnit.ID, exampleInput)
		assert.NotNil(t, result)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ArchiveValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientMeasurementUnit()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.ArchiveValidIngredientMeasurementUnit), testutils.ContextMatcher, expected.ID).Return(nil)
			},
			map[string][]string{
				types.ValidIngredientMeasurementUnitArchivedServiceEventType: {mealplanningkeys.ValidIngredientMeasurementUnitIDKey},
			},
		)

		assert.NoError(t, vem.ArchiveValidIngredientMeasurementUnit(ctx, expected.ID))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_SearchValidIngredientMeasurementUnitsByIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientMeasurementUnitsList()
		exampleQuery := fakes.BuildFakeID()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidIngredientMeasurementUnitsForIngredient), testutils.ContextMatcher, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := vem.SearchValidIngredientMeasurementUnitsByIngredient(ctx, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_SearchValidIngredientMeasurementUnitsByMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidIngredientMeasurementUnitsList()
		exampleQuery := fakes.BuildFakeID()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidIngredientMeasurementUnitsForMeasurementUnit), testutils.ContextMatcher, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := vem.SearchValidIngredientMeasurementUnitsByMeasurementUnit(ctx, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
