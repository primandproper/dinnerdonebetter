package managers

import (
	"context"
	"testing"

	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	mealplanningworkers "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers"

	"github.com/primandproper/platform-go/v8/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v8/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v8/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v8/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v8/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v8/observability/tracing/noop"
	textsearchcfg "github.com/primandproper/platform-go/v8/search/text/config"

	"github.com/stretchr/testify/require"
)

// newManagerForTest constructs a manager wired to unconfigured mocks. Tests swap in their own
// configured repository via attachRepositoryToManager.
func newManagerForTest(t *testing.T, groceryWorker, taskWorker mealplanningworkers.Worker) *mealPlanningManager {
	t.Helper()

	queueCfg := &msgconfig.QueuesConfig{
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
		&textsearchcfg.Config{},
		metricsnoop.NewMetricsProvider(),
		groceryWorker,
		taskWorker,
	)
	require.NoError(t, err)

	return m.(*mealPlanningManager)
}

func buildMealPlanManagerForTest(t *testing.T) *mealPlanningManager {
	t.Helper()

	return newManagerForTest(t, nil, nil)
}

func buildMealPlanManagerForTestWithWorkers(t *testing.T, groceryWorker, taskWorker *mealplanningworkers.WorkerMock) *mealPlanningManager {
	t.Helper()

	return newManagerForTest(t, groceryWorker, taskWorker)
}

func buildRecipeManagerForTest(t *testing.T) *mealPlanningManager {
	t.Helper()

	return newManagerForTest(t, nil, nil)
}

func buildValidEnumerationsManagerForTest(t *testing.T) *mealPlanningManager {
	t.Helper()

	return newManagerForTest(t, nil, nil)
}

// attachRepositoryToManager wires a configured repository mock and a no-op data changes publisher
// into the manager under test.
func attachRepositoryToManager(manager *mealPlanningManager, db *mealplanningmock.RepositoryMock) {
	manager.db = db
	manager.dataChangesPublisher = &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}
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
