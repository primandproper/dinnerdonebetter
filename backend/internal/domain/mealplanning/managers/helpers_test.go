package managers

import (
	"context"
	"testing"

	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	eatingindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v13/messagequeue"
	mockpublishers "github.com/primandproper/platform-go/v13/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	textsearch "github.com/primandproper/platform-go/v13/search/text"
	textsearchcfg "github.com/primandproper/platform-go/v13/search/text/config"

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

// newManagerForTest constructs a manager wired to unconfigured mocks. Tests swap in their own
// configured repository via attachRepositoryToManager.
func newManagerForTest(t *testing.T, starter mealPlanFinalizationStarter) *mealPlanningManager {
	t.Helper()

	queueCfg := &queuescfg.Config{
		DataChangesTopicName: t.Name(),
	}

	mpp := &mockpublishers.PublisherProviderMock{
		NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{}, nil
		},
	}

	m, err := NewMealPlanningManager(
		t.Context(),
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		&mealplanningmock.RepositoryMock{},
		queueCfg,
		mpp,
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
func attachRecipeSearchIndexToManager(manager *mealPlanningManager, index textsearch.IndexSearcher[eatingindexing.RecipeSearchSubset]) {
	manager.recipeSearchIndex = index
}

// attachValidIngredientSearchIndexToManager swaps in a configured valid ingredient search index.
func attachValidIngredientSearchIndexToManager(manager *mealPlanningManager, index textsearch.IndexSearcher[eatingindexing.ValidIngredientSearchSubset]) {
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
