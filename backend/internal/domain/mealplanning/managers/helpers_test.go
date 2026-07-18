package managers

import (
	"context"
	"slices"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	mealplanningworkers "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers"

	"github.com/primandproper/platform-go/v5/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v5/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v5/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v5/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v5/observability/tracing/noop"
	textsearchcfg "github.com/primandproper/platform-go/v5/search/text/config"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func eventMatches(eventType string, keys []string) any {
	return mock.MatchedBy(func(message *audit.DataChangeMessage) bool {
		allContextKeys := []string{}
		for k := range message.Context {
			allContextKeys = append(allContextKeys, k)
		}

		slices.Sort(keys)
		slices.Sort(allContextKeys)
		allKeysMatch := slices.Equal(keys, allContextKeys)
		eventTypeMatches := message.EventType == eventType
		result := allKeysMatch && eventTypeMatches

		return result
	})
}

func buildMealPlanManagerForTest(t *testing.T) *mealPlanningManager {
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
		&mealplanningmock.Repository{},
		queueCfg,
		mpp,
		&recipeanalysis.MockRecipeAnalyzer{},
		&textsearchcfg.Config{},
		metricsnoop.NewMetricsProvider(),
		nil,
		nil,
	)
	require.NoError(t, err)

	return m.(*mealPlanningManager)
}

func buildMealPlanManagerForTestWithWorkers(t *testing.T, groceryWorker, taskWorker *mealplanningworkers.MockWorker) *mealPlanningManager {
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
		&mealplanningmock.Repository{},
		queueCfg,
		mpp,
		&recipeanalysis.MockRecipeAnalyzer{},
		&textsearchcfg.Config{},
		metricsnoop.NewMetricsProvider(),
		groceryWorker,
		taskWorker,
	)
	require.NoError(t, err)

	return m.(*mealPlanningManager)
}

func setupExpectationsForMealPlanningManager(
	manager *mealPlanningManager,
	dbSetupFunc func(db *mealplanningmock.Repository),
	eventTypeMaps ...map[string][]string,
) []any {
	db := &mealplanningmock.Repository{}
	if dbSetupFunc != nil {
		dbSetupFunc(db)
	}
	manager.db = db

	mp := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}
	manager.dataChangesPublisher = mp

	return []any{db}
}

func buildRecipeManagerForTest(t *testing.T) *mealPlanningManager {
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
		&mealplanningmock.Repository{},
		queueCfg,
		mpp,
		&recipeanalysis.MockRecipeAnalyzer{},
		&textsearchcfg.Config{},
		metricsnoop.NewMetricsProvider(),
		nil,
		nil,
	)
	require.NoError(t, err)

	return m.(*mealPlanningManager)
}

func setupExpectationsForRecipeManager(
	manager *mealPlanningManager,
	dbSetupFunc func(db *mealplanningmock.Repository),
	eventTypeMaps ...map[string][]string,
) []any {
	db := &mealplanningmock.Repository{}
	if dbSetupFunc != nil {
		dbSetupFunc(db)
	}
	manager.db = db

	mp := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}
	manager.dataChangesPublisher = mp

	return []any{db}
}

func setupExpectationsForRecipeManagerWithAnalyzer(
	manager *mealPlanningManager,
	dbSetupFunc func(db *mealplanningmock.Repository),
	analyzerSetupFunc func(analyzer *recipeanalysis.MockRecipeAnalyzer),
	eventTypeMaps ...map[string][]string,
) []any {
	expectations := setupExpectationsForRecipeManager(manager, dbSetupFunc, eventTypeMaps...)

	ra := &recipeanalysis.MockRecipeAnalyzer{}
	if analyzerSetupFunc != nil {
		analyzerSetupFunc(ra)
	}
	manager.recipeAnalyzer = ra

	return expectations
}

func buildValidEnumerationsManagerForTest(t *testing.T) *mealPlanningManager {
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
		&mealplanningmock.Repository{},
		queueCfg,
		mpp,
		&recipeanalysis.MockRecipeAnalyzer{},
		&textsearchcfg.Config{},
		metricsnoop.NewMetricsProvider(),
		nil,
		nil,
	)
	require.NoError(t, err)

	return m.(*mealPlanningManager)
}

func setupExpectationsForValidEnumerationManager(
	manager *mealPlanningManager,
	dbSetupFunc func(db *mealplanningmock.Repository),
	eventTypeMaps ...map[string][]string,
) []any {
	db := &mealplanningmock.Repository{}
	if dbSetupFunc != nil {
		dbSetupFunc(db)
	}
	manager.db = db

	mp := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}
	manager.dataChangesPublisher = mp

	return []any{db}
}
