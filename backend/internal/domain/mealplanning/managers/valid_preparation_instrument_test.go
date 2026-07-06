package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v3/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestValidEnumerationManager_ListValidPreparationInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationInstrumentsList()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidPreparationInstruments), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := vem.ListValidPreparationInstruments(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_CreateValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationInstrument()
		fakeInput := fakes.BuildFakeValidPreparationInstrumentCreationRequestInput()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.CreateValidPreparationInstrument), testutils.ContextMatcher, testutils.MatchType[*types.ValidPreparationInstrumentDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.ValidPreparationInstrumentCreatedServiceEventType: {mealplanningkeys.ValidPreparationInstrumentIDKey},
			},
		)

		actual, err := vem.CreateValidPreparationInstrument(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ReadValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationInstrument()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidPreparationInstrument), testutils.ContextMatcher, expected.ID).Return(expected, nil)
			},
		)

		actual, err := vem.ReadValidPreparationInstrument(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_UpdateValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidPreparationInstrument := fakes.BuildFakeValidPreparationInstrument()
		exampleInput := fakes.BuildFakeValidPreparationInstrumentUpdateRequestInput()

		expectations := setupExpectationsForValidEnumerationManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetValidPreparationInstrument), testutils.ContextMatcher, exampleValidPreparationInstrument.ID).Return(exampleValidPreparationInstrument, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateValidPreparationInstrument), testutils.ContextMatcher, testutils.MatchType[*types.ValidPreparationInstrument]()).Return(nil)
			},
			map[string][]string{
				types.ValidPreparationInstrumentUpdatedServiceEventType: {mealplanningkeys.ValidPreparationInstrumentIDKey},
			},
		)

		result, err := mpm.UpdateValidPreparationInstrument(ctx, exampleValidPreparationInstrument.ID, exampleInput)
		assert.NotNil(t, result)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ArchiveValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationInstrument()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.ArchiveValidPreparationInstrument), testutils.ContextMatcher, expected.ID).Return(nil)
			},
			map[string][]string{
				types.ValidPreparationInstrumentArchivedServiceEventType: {mealplanningkeys.ValidPreparationInstrumentIDKey},
			},
		)

		assert.NoError(t, vem.ArchiveValidPreparationInstrument(ctx, expected.ID))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_SearchValidPreparationInstrumentsByPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationInstrumentsList()
		exampleQuery := fakes.BuildFakeID()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidPreparationInstrumentsForPreparation), testutils.ContextMatcher, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := vem.SearchValidPreparationInstrumentsByPreparation(ctx, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_SearchValidPreparationInstrumentsByInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationInstrumentsList()
		exampleQuery := fakes.BuildFakeID()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidPreparationInstrumentsForInstrument), testutils.ContextMatcher, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := vem.SearchValidPreparationInstrumentsByInstrument(ctx, exampleQuery, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
