package mealplanfinalizer

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v7/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v7/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v7/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v7/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v7/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v7/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildNewMealPlanFinalizerForTest(t *testing.T) *Worker {
	t.Helper()

	ctx := t.Context()
	cfg := &msgconfig.QueuesConfig{DataChangesTopicName: "data_changes"}

	pp := &mockpublishers.PublisherProviderMock{
		NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{
				PublishFunc:      func(_ context.Context, _ any) error { return nil },
				PublishAsyncFunc: func(_ context.Context, _ any) {},
				StopFunc:         func() {},
			}, nil
		},
	}

	x, err := NewMealPlanFinalizer(
		ctx,
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		&mealplanningmock.RepositoryMock{},
		pp,
		metricsnoop.NewMetricsProvider(),
		cfg,
	)
	require.NoError(t, err)

	return x
}

func TestWorker_Work(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleMealPlans := fakes.BuildFakeMealPlansList().Data

		// every returned meal plan must be finalized with its own ID and owning account.
		expectedFinalizations := map[string]string{}
		for _, mealPlan := range exampleMealPlans {
			expectedFinalizations[mealPlan.ID] = mealPlan.BelongsToAccount
		}

		dbm := &mealplanningmock.RepositoryMock{
			GetUnfinalizedMealPlansWithExpiredVotingPeriodsFunc: func(context.Context) ([]*mealplanning.MealPlan, error) {
				return exampleMealPlans, nil
			},
			AttemptToFinalizeMealPlanFunc: func(_ context.Context, mealPlanID, accountID string) (bool, error) {
				expectedAccountID, ok := expectedFinalizations[mealPlanID]
				assert.True(t, ok, "unexpected meal plan finalized: %s", mealPlanID)
				assert.Equal(t, expectedAccountID, accountID)

				return true, nil
			},
		}

		pup := &mockpublishers.PublisherMock{
			PublishFunc: func(_ context.Context, _ any) error { return nil },
		}

		worker := buildNewMealPlanFinalizerForTest(t)
		worker.dataManager = dbm
		worker.postUpdatesPublisher = pup

		expected := int64(len(exampleMealPlans))

		actual, err := worker.Work(ctx)
		assert.Equal(t, expected, actual)
		assert.NoError(t, err)

		assert.Len(t, dbm.GetUnfinalizedMealPlansWithExpiredVotingPeriodsCalls(), 1)
		assert.Len(t, dbm.AttemptToFinalizeMealPlanCalls(), len(exampleMealPlans))
	})
}
