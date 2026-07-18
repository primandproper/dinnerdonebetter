package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidEnumerationManager_ValidMeasurementUnitConversionsForMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidMeasurementUnitConversionsList()
		exampleQuery := fakes.BuildFakeID()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidMeasurementUnitConversionsForUnit), testutils.ContextMatcher, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := vem.ValidMeasurementUnitConversionsForMeasurementUnit(ctx, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_CreateValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidMeasurementUnitConversion()
		fakeInput := fakes.BuildFakeValidMeasurementUnitConversionCreationRequestInput()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.CreateValidMeasurementUnitConversion), testutils.ContextMatcher, testutils.MatchType[*types.ValidMeasurementUnitConversionDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.ValidMeasurementUnitConversionCreatedServiceEventType: {mealplanningkeys.ValidMeasurementUnitConversionIDKey},
			},
		)

		actual, err := vem.CreateValidMeasurementUnitConversion(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ReadValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidMeasurementUnitConversion()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidMeasurementUnitConversion), testutils.ContextMatcher, expected.ID).Return(expected, nil)
			},
		)

		actual, err := vem.ReadValidMeasurementUnitConversion(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_UpdateValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidMeasurementUnitConversion := fakes.BuildFakeValidMeasurementUnitConversion()
		exampleInput := fakes.BuildFakeValidMeasurementUnitConversionUpdateRequestInput()

		expectations := setupExpectationsForValidEnumerationManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetValidMeasurementUnitConversion), testutils.ContextMatcher, exampleValidMeasurementUnitConversion.ID).Return(exampleValidMeasurementUnitConversion, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateValidMeasurementUnitConversion), testutils.ContextMatcher, testutils.MatchType[*types.ValidMeasurementUnitConversion]()).Return(nil)
			},
			map[string][]string{
				types.ValidMeasurementUnitConversionUpdatedServiceEventType: {mealplanningkeys.ValidMeasurementUnitConversionIDKey},
			},
		)

		result, err := mpm.UpdateValidMeasurementUnitConversion(ctx, exampleValidMeasurementUnitConversion.ID, exampleInput)
		assert.NotNil(t, result)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ArchiveValidMeasurementUnitConversion(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidMeasurementUnitConversion()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.ArchiveValidMeasurementUnitConversion), testutils.ContextMatcher, expected.ID).Return(nil)
			},
			map[string][]string{
				types.ValidMeasurementUnitConversionArchivedServiceEventType: {mealplanningkeys.ValidMeasurementUnitConversionIDKey},
			},
		)

		assert.NoError(t, vem.ArchiveValidMeasurementUnitConversion(ctx, expected.ID))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
