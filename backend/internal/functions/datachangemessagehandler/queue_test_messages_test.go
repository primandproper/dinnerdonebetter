package datachangemessagehandler

import (
	"context"
	"errors"
	"testing"

	internalopsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops/mock"

	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func TestAsyncDataChangeMessageHandler_handleQueueTestMessage(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()
		_, span := tracing.NewTracerForTest(t.Name()).StartSpan(ctx)
		logger := loggingnoop.NewLogger()

		repo := &internalopsmock.InternalOpsDataManagerMock{}
		handler.internalOpsRepo = repo

		repo.AcknowledgeQueueTestMessageFunc = func(_ context.Context, id string) error {
			assert.Equal(t, "test-123", id)

			return nil
		}
		repo.PruneQueueTestMessagesFunc = func(_ context.Context, queueName string) error {
			assert.Equal(t, "data-changes", queueName)

			return nil
		}

		err := handler.handleQueueTestMessage(ctx, logger, span, "test-123", "data-changes")

		assert.NoError(t, err)
		assert.Len(t, repo.AcknowledgeQueueTestMessageCalls(), 1)
		assert.Len(t, repo.PruneQueueTestMessagesCalls(), 1)
	})

	t.Run("empty test_id", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()
		_, span := tracing.NewTracerForTest(t.Name()).StartSpan(ctx)
		logger := loggingnoop.NewLogger()

		err := handler.handleQueueTestMessage(ctx, logger, span, "", "data-changes")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test_id")
	})

	t.Run("empty topic_name skips prune", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()
		_, span := tracing.NewTracerForTest(t.Name()).StartSpan(ctx)
		logger := loggingnoop.NewLogger()

		repo := &internalopsmock.InternalOpsDataManagerMock{}
		handler.internalOpsRepo = repo

		repo.AcknowledgeQueueTestMessageFunc = func(_ context.Context, id string) error {
			assert.Equal(t, "test-123", id)

			return nil
		}

		err := handler.handleQueueTestMessage(ctx, logger, span, "test-123", "")

		assert.NoError(t, err)
		assert.Len(t, repo.AcknowledgeQueueTestMessageCalls(), 1)
	})

	t.Run("acknowledge error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()
		_, span := tracing.NewTracerForTest(t.Name()).StartSpan(ctx)
		logger := loggingnoop.NewLogger()

		repo := &internalopsmock.InternalOpsDataManagerMock{}
		handler.internalOpsRepo = repo

		repo.AcknowledgeQueueTestMessageFunc = func(_ context.Context, id string) error {
			assert.Equal(t, "test-123", id)

			return errors.New("db error")
		}

		err := handler.handleQueueTestMessage(ctx, logger, span, "test-123", "data-changes")

		assert.Error(t, err)
		assert.Len(t, repo.AcknowledgeQueueTestMessageCalls(), 1)
	})

	t.Run("prune error is not fatal", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()
		_, span := tracing.NewTracerForTest(t.Name()).StartSpan(ctx)
		logger := loggingnoop.NewLogger()

		repo := &internalopsmock.InternalOpsDataManagerMock{}
		handler.internalOpsRepo = repo

		repo.AcknowledgeQueueTestMessageFunc = func(_ context.Context, id string) error {
			assert.Equal(t, "test-123", id)

			return nil
		}
		repo.PruneQueueTestMessagesFunc = func(_ context.Context, queueName string) error {
			assert.Equal(t, "data-changes", queueName)

			return errors.New("prune error")
		}

		err := handler.handleQueueTestMessage(ctx, logger, span, "test-123", "data-changes")

		assert.NoError(t, err)
		assert.Len(t, repo.AcknowledgeQueueTestMessageCalls(), 1)
		assert.Len(t, repo.PruneQueueTestMessagesCalls(), 1)
	})
}
