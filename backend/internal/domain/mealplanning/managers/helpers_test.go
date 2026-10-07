package managers

import (
	"context"
	"testing"

	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/searchindex"

	"github.com/primandproper/primitives-go/v2/clock"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	textsearch "github.com/primandproper/primitives-go/v2/search/text"
	textsearchcfg "github.com/primandproper/primitives-go/v2/search/text/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeFinalizationStarter records the plans it was asked to enter into the finalization
// pipeline, so a test can assert that finalizing one started its saga.
//
// Hand-written rather than generated: the interface it satisfies is unexported, because nothing
// outside this package needs to name it.
type fakeFinalizationStarter struct {
	err   error
	calls []string
}

func (f *fakeFinalizationStarter) EnsureStarted(_ context.Context, mealPlanID, _ string) error {
	f.calls = append(f.calls, mealPlanID)

	return f.err
}

// fakeElectorate answers every roster read with the same members.
type fakeElectorate struct {
	err     error
	members []string
}

func (f *fakeElectorate) MembersOfAccount(context.Context, string) ([]string, error) {
	return f.members, f.err
}

// newManagerForTest constructs a manager wired to unconfigured mocks and the wall clock. Tests
// swap in their own configured repository via attachRepositoryToManager, and a test that needs
// to stand at a particular moment uses newManagerForTestWithClock.
func newManagerForTest(t *testing.T, starter mealPlanFinalizationStarter) *mealPlanningManager {
	t.Helper()

	return newManagerForTestWithClock(t, starter, clock.NewClock())
}

func newManagerForTestWithClock(t *testing.T, starter mealPlanFinalizationStarter, wallClock clock.Clock) *mealPlanningManager {
	t.Helper()

	m, err := NewMealPlanningManager(
		t.Context(),
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		&mealplanningmock.RepositoryMock{},
		&fakeElectorate{},
		wallClock,
		&recipeanalysis.RecipeAnalyzerMock{},
		&textsearchcfg.Config{Provider: textsearchcfg.ProviderNoop},
		metricsnoop.NewMetricsProvider(),
		starter,
	)
	require.NoError(t, err)

	return m.(*mealPlanningManager)
}

func buildMealPlanManagerForTest(t *testing.T) *mealPlanningManager {
	t.Helper()

	return newManagerForTest(t, nil)
}

func buildMealPlanManagerForTestWithStarter(t *testing.T, starter *fakeFinalizationStarter) *mealPlanningManager {
	t.Helper()

	return newManagerForTest(t, starter)
}

func buildRecipeManagerForTest(t *testing.T) *mealPlanningManager {
	t.Helper()

	return newManagerForTest(t, nil)
}

func buildValidEnumerationsManagerForTest(t *testing.T) *mealPlanningManager {
	t.Helper()

	return newManagerForTest(t, nil)
}

// attachRepositoryToManager wires a configured repository mock into the manager under test.
//
// There is no publisher to wire any more: data change events are enqueued into the outbox by the
// repository, inside the transaction that writes the row they describe.
func attachRepositoryToManager(manager *mealPlanningManager, db *mealplanningmock.RepositoryMock) {
	manager.db = db
}

// attachRecipeSearchIndexToManager swaps in a configured recipe search index. The manager is
// otherwise built against the noop index, which answers every query with no hits and no cursor.
func attachRecipeSearchIndexToManager(manager *mealPlanningManager, index textsearch.IndexSearcher[searchindex.RecipeSearchSubset]) {
	manager.recipeSearchIndex = index
}

// attachValidIngredientSearchIndexToManager swaps in a configured valid ingredient search index.
func attachValidIngredientSearchIndexToManager(manager *mealPlanningManager, index textsearch.IndexSearcher[searchindex.ValidIngredientSearchSubset]) {
	manager.validIngredientSearchIndex = index
}

// attachRepositoryAndAnalyzerToManager additionally swaps in a configured recipe analyzer. A nil
// analyzer gets an unconfigured mock, which panics if any of its methods are called.
func attachRepositoryAndAnalyzerToManager(manager *mealPlanningManager, db *mealplanningmock.RepositoryMock, analyzer *recipeanalysis.RecipeAnalyzerMock) {
	attachRepositoryToManager(manager, db)

	if analyzer == nil {
		analyzer = &recipeanalysis.RecipeAnalyzerMock{}
	}
	manager.recipeAnalyzer = analyzer
}

// The stubs below answer the ownership checks the manager makes before it touches the repository.
// Each asserts that it was asked about exactly the IDs the test passed in, and answers as told.

func recipeIsOwnedByStub(t *testing.T, expectedRecipeID, expectedOwnerID string, owned bool) func(context.Context, string, string) (bool, error) {
	t.Helper()

	return func(_ context.Context, recipeID, userID string) (bool, error) {
		assert.Equal(t, expectedRecipeID, recipeID)
		assert.Equal(t, expectedOwnerID, userID)

		return owned, nil
	}
}

func recipeStepExistsStub(t *testing.T, expectedRecipeID, expectedRecipeStepID string, exists bool) func(context.Context, string, string) (bool, error) {
	t.Helper()

	return func(_ context.Context, recipeID, recipeStepID string) (bool, error) {
		assert.Equal(t, expectedRecipeID, recipeID)
		assert.Equal(t, expectedRecipeStepID, recipeStepID)

		return exists, nil
	}
}

func mealPlanExistsStub(t *testing.T, expectedMealPlanID, expectedAccountID string, exists bool) func(context.Context, string, string) (bool, error) {
	t.Helper()

	return func(_ context.Context, mealPlanID, accountID string) (bool, error) {
		assert.Equal(t, expectedMealPlanID, mealPlanID)
		assert.Equal(t, expectedAccountID, accountID)

		return exists, nil
	}
}

func mealPlanEventExistsStub(t *testing.T, expectedMealPlanID, expectedMealPlanEventID string, exists bool) func(context.Context, string, string) (bool, error) {
	t.Helper()

	return func(_ context.Context, mealPlanID, mealPlanEventID string) (bool, error) {
		assert.Equal(t, expectedMealPlanID, mealPlanID)
		assert.Equal(t, expectedMealPlanEventID, mealPlanEventID)

		return exists, nil
	}
}

func mealPlanTaskExistsStub(t *testing.T, expectedMealPlanID, expectedMealPlanTaskID string, exists bool) func(context.Context, string, string) (bool, error) {
	t.Helper()

	return func(_ context.Context, mealPlanID, mealPlanTaskID string) (bool, error) {
		assert.Equal(t, expectedMealPlanID, mealPlanID)
		assert.Equal(t, expectedMealPlanTaskID, mealPlanTaskID)

		return exists, nil
	}
}

func mealPlanGroceryListItemExistsStub(t *testing.T, expectedMealPlanID, expectedItemID string, exists bool) func(context.Context, string, string) (bool, error) {
	t.Helper()

	return func(_ context.Context, mealPlanID, itemID string) (bool, error) {
		assert.Equal(t, expectedMealPlanID, mealPlanID)
		assert.Equal(t, expectedItemID, itemID)

		return exists, nil
	}
}

func mealPlanOptionExistsStub(t *testing.T, expectedMealPlanID, expectedMealPlanEventID, expectedMealPlanOptionID string, exists bool) func(context.Context, string, string, string) (bool, error) {
	t.Helper()

	return func(_ context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) (bool, error) {
		assert.Equal(t, expectedMealPlanID, mealPlanID)
		assert.Equal(t, expectedMealPlanEventID, mealPlanEventID)
		assert.Equal(t, expectedMealPlanOptionID, mealPlanOptionID)

		return exists, nil
	}
}

func mealPlanOptionBelongsToAccountStub(t *testing.T, expectedMealPlanOptionID, expectedAccountID string, belongs bool) func(context.Context, string, string) (bool, error) {
	t.Helper()

	return func(_ context.Context, mealPlanOptionID, accountID string) (bool, error) {
		assert.Equal(t, expectedMealPlanOptionID, mealPlanOptionID)
		assert.Equal(t, expectedAccountID, accountID)

		return belongs, nil
	}
}
