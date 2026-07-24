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

func TestValidEnumerationManager_SearchValidInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidInstrumentsList()
		exampleQuery := fakes.BuildFakeID()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.SearchForValidInstruments), testutils.ContextMatcher, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := vem.SearchValidInstruments(ctx, exampleQuery, false, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ListValidInstruments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidInstrumentsList()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidInstruments), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := vem.ListValidInstruments(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_CreateValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidInstrument()
		fakeInput := fakes.BuildFakeValidInstrumentCreationRequestInput()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.CreateValidInstrument), testutils.ContextMatcher, testutils.MatchType[*types.ValidInstrumentDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.ValidInstrumentCreatedServiceEventType: {mealplanningkeys.ValidInstrumentIDKey},
			},
		)

		actual, err := vem.CreateValidInstrument(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ReadValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidInstrument()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidInstrument), testutils.ContextMatcher, expected.ID).Return(expected, nil)
			},
		)

		actual, err := vem.ReadValidInstrument(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_RandomValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidInstrument()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetRandomValidInstrument), testutils.ContextMatcher).Return(expected, nil)
			},
		)

		actual, err := vem.RandomValidInstrument(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_UpdateValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidInstrument := fakes.BuildFakeValidInstrument()
		exampleInput := fakes.BuildFakeValidInstrumentUpdateRequestInput()

		expectations := setupExpectationsForValidEnumerationManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetValidInstrument), testutils.ContextMatcher, exampleValidInstrument.ID).Return(exampleValidInstrument, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateValidInstrument), testutils.ContextMatcher, testutils.MatchType[*types.ValidInstrument]()).Return(nil)
			},
			map[string][]string{
				types.ValidInstrumentUpdatedServiceEventType: {mealplanningkeys.ValidInstrumentIDKey},
			},
		)

		result, err := mpm.UpdateValidInstrument(ctx, exampleValidInstrument.ID, exampleInput)
		assert.NotNil(t, result)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ArchiveValidInstrument(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidInstrument()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.ArchiveValidInstrument), testutils.ContextMatcher, expected.ID).Return(nil)
			},
			map[string][]string{
				types.ValidInstrumentArchivedServiceEventType: {mealplanningkeys.ValidInstrumentIDKey},
			},
		)

		assert.NoError(t, vem.ArchiveValidInstrument(ctx, expected.ID))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
