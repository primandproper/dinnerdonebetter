package datachangemessagehandler

import (
	"errors"
	"testing"

	internalopsmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/internalops/mock"

	loggingnoop "github.com/primandproper/platform-go/v3/observability/logging/noop"
	"github.com/primandproper/platform-go/v3/observability/tracing"
	"github.com/primandproper/platform-go/v3/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAsyncDataChangeMessageHandler_handleQueueTestMessage(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()
		_, span := tracing.NewTracerForTest(t.Name()).StartSpan(ctx)
		logger := loggingnoop.NewLogger()

		repo := &internalopsmock.InternalOpsDataManager{}
		handler.internalOpsRepo = repo

		repo.On(reflection.GetMethodName(repo.AcknowledgeQueueTestMessage), mock.Anything, "test-123").Return(nil).Once()
		repo.On(reflection.GetMethodName(repo.PruneQueueTestMessages), mock.Anything, "data-changes").Return(nil).Once()

		err := handler.handleQueueTestMessage(ctx, logger, span, "test-123", "data-changes")

		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, repo)
	})

	t.Run("empty test_id", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()
		_, span := tracing.NewTracerForTest(t.Name()).StartSpan(ctx)
		logger := loggingnoop.NewLogger()

		err := handler.handleQueueTestMessage(ctx, logger, span, "", "data-changes")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test_id")
	})

	t.Run("empty topic_name skips prune", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()
		_, span := tracing.NewTracerForTest(t.Name()).StartSpan(ctx)
		logger := loggingnoop.NewLogger()

		repo := &internalopsmock.InternalOpsDataManager{}
		handler.internalOpsRepo = repo

		repo.On(reflection.GetMethodName(repo.AcknowledgeQueueTestMessage), mock.Anything, "test-123").Return(nil).Once()

		err := handler.handleQueueTestMessage(ctx, logger, span, "test-123", "")

		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, repo)
	})

	t.Run("acknowledge error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()
		_, span := tracing.NewTracerForTest(t.Name()).StartSpan(ctx)
		logger := loggingnoop.NewLogger()

		repo := &internalopsmock.InternalOpsDataManager{}
		handler.internalOpsRepo = repo

		repo.On(reflection.GetMethodName(repo.AcknowledgeQueueTestMessage), mock.Anything, "test-123").Return(errors.New("db error")).Once()

		err := handler.handleQueueTestMessage(ctx, logger, span, "test-123", "data-changes")

		assert.Error(t, err)
		mock.AssertExpectationsForObjects(t, repo)
	})

	t.Run("prune error is not fatal", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()
		_, span := tracing.NewTracerForTest(t.Name()).StartSpan(ctx)
		logger := loggingnoop.NewLogger()

		repo := &internalopsmock.InternalOpsDataManager{}
		handler.internalOpsRepo = repo

		repo.On(reflection.GetMethodName(repo.AcknowledgeQueueTestMessage), mock.Anything, "test-123").Return(nil).Once()
		repo.On(reflection.GetMethodName(repo.PruneQueueTestMessages), mock.Anything, "data-changes").Return(errors.New("prune error")).Once()

		err := handler.handleQueueTestMessage(ctx, logger, span, "test-123", "data-changes")

		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, repo)
	})
}
