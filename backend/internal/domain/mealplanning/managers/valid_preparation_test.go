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

func TestValidEnumerationManager_SearchValidPreparations(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationsList()
		exampleQuery := fakes.BuildFakeID()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.SearchForValidPreparations), testutils.ContextMatcher, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)

				for _, prep := range expected.Data {
					db.On(reflection.GetMethodName(vem.db.GetPreparationMediaByPreparation), testutils.ContextMatcher, prep.ID).Return([]*types.PreparationMediaRow{}, nil)
				}
			},
		)

		actual, err := vem.SearchValidPreparations(ctx, exampleQuery, false, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ListValidPreparations(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparationsList()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidPreparations), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(expected, nil)

				for _, prep := range expected.Data {
					db.On(reflection.GetMethodName(vem.db.GetPreparationMediaByPreparation), testutils.ContextMatcher, prep.ID).Return([]*types.PreparationMediaRow{}, nil)
				}
			},
		)

		actual, err := vem.ListValidPreparations(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_CreateValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparation()
		fakeInput := fakes.BuildFakeValidPreparationCreationRequestInput()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.CreateValidPreparation), testutils.ContextMatcher, testutils.MatchType[*types.ValidPreparationDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.ValidPreparationCreatedServiceEventType: {mealplanningkeys.ValidPreparationIDKey},
			},
		)

		actual, err := vem.CreateValidPreparation(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ReadValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparation()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetValidPreparation), testutils.ContextMatcher, expected.ID).Return(expected, nil)

				db.On(reflection.GetMethodName(vem.db.GetPreparationMediaByPreparation), testutils.ContextMatcher, expected.ID).Return([]*types.PreparationMediaRow{}, nil)
			},
		)

		actual, err := vem.ReadValidPreparation(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_RandomValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparation()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.GetRandomValidPreparation), testutils.ContextMatcher).Return(expected, nil)

				db.On(reflection.GetMethodName(vem.db.GetPreparationMediaByPreparation), testutils.ContextMatcher, expected.ID).Return([]*types.PreparationMediaRow{}, nil)
			},
		)

		actual, err := vem.RandomValidPreparation(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_UpdateValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildValidEnumerationsManagerForTest(t)

		exampleValidPreparation := fakes.BuildFakeValidPreparation()
		exampleInput := fakes.BuildFakeValidPreparationUpdateRequestInput()

		expectations := setupExpectationsForValidEnumerationManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetValidPreparation), testutils.ContextMatcher, exampleValidPreparation.ID).Return(exampleValidPreparation, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateValidPreparation), testutils.ContextMatcher, testutils.MatchType[*types.ValidPreparation]()).Return(nil)

				db.On(reflection.GetMethodName(mpm.db.GetPreparationMediaByPreparation), testutils.ContextMatcher, exampleValidPreparation.ID).Return([]*types.PreparationMediaRow{}, nil)
			},
			map[string][]string{
				types.ValidPreparationUpdatedServiceEventType: {mealplanningkeys.ValidPreparationIDKey},
			},
		)

		result, err := mpm.UpdateValidPreparation(ctx, exampleValidPreparation.ID, exampleInput)
		assert.NotNil(t, result)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestValidEnumerationManager_ArchiveValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		vem := buildValidEnumerationsManagerForTest(t)

		expected := fakes.BuildFakeValidPreparation()

		expectations := setupExpectationsForValidEnumerationManager(
			vem,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(vem.db.ArchiveValidPreparation), testutils.ContextMatcher, expected.ID).Return(nil)
			},
			map[string][]string{
				types.ValidPreparationArchivedServiceEventType: {mealplanningkeys.ValidPreparationIDKey},
			},
		)

		assert.NoError(t, vem.ArchiveValidPreparation(ctx, expected.ID))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
