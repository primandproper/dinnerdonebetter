package queuetest

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config/environments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"
	internalopsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops/mock"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildTestJob(t *testing.T, repo internalops.InternalOpsDataManager, queues queuescfg.Config) *Job {
	t.Helper()

	job, err := NewJob(repo, nil, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), &JobParams{Queues: queues})
	require.NoError(t, err)

	return job
}

// TestJob_topicNames pins every topic the probe may pick to one it can build a probe for, in
// every environment. A name it cannot build fails the run it is picked on, which for a random
// pick from four names was one run in four.
func TestJob_topicNames(T *testing.T) {
	T.Parallel()

	assertProbeable := func(t *testing.T, queues queuescfg.Config) {
		t.Helper()

		names := buildTestJob(t, &internalopsmock.InternalOpsDataManagerMock{}, queues).topicNames()
		require.NotEmpty(t, names)

		for _, name := range names {
			_, err := internalops.BuildQueueTestMessage(name, identifiers.New(), "")
			assert.NoError(t, err, "the probe can pick %q, and cannot build a message for it", name)
		}
	}

	T.Run("in production", func(t *testing.T) {
		t.Parallel()

		assertProbeable(t, environments.BuildProdConfig().Queues)
	})

	T.Run("in local development", func(t *testing.T) {
		t.Parallel()

		assertProbeable(t, environments.BuildLocalDevConfig().Queues)
	})

	T.Run("in the integration tests", func(t *testing.T) {
		t.Parallel()

		assertProbeable(t, environments.BuildIntegrationTestsConfig().Queues)
	})
}

func TestJob_Do(T *testing.T) {
	T.Parallel()

	T.Run("writes no probe row for a topic it cannot probe", func(t *testing.T) {
		t.Parallel()

		repo := &internalopsmock.InternalOpsDataManagerMock{
			CreateQueueTestMessageFunc: func(context.Context, string, string) error {
				t.Error("a probe row was written for a message that could not be built")
				return nil
			},
		}

		unprobeable := queuescfg.Config{
			DataChangesTopicName:         identifiers.New(),
			OutboundEmailsTopicName:      identifiers.New(),
			MobileNotificationsTopicName: identifiers.New(),
		}

		assert.Error(t, buildTestJob(t, repo, unprobeable).Do(t.Context()))
	})
}
