package datachangemessagehandler

import (
	"context"
	"fmt"

	"github.com/primandproper/platform-go/v3/observability"
	"github.com/primandproper/platform-go/v3/observability/logging"
	"github.com/primandproper/platform-go/v3/observability/tracing"
)

func (a *AsyncDataChangeMessageHandler) handleQueueTestMessage(
	ctx context.Context,
	logger logging.Logger,
	span tracing.Span,
	testID,
	topicName string,
) error {
	if testID == "" {
		return fmt.Errorf("missing or invalid test_id in queue test message")
	}

	l := logger.WithValue("test_id", testID).WithValue("topic_name", topicName)
	l.Info("acknowledging queue test message")

	if err := a.internalOpsRepo.AcknowledgeQueueTestMessage(ctx, testID); err != nil {
		return observability.PrepareAndLogError(err, l, span, "acknowledging queue test message")
	}

	if topicName != "" {
		if err := a.internalOpsRepo.PruneQueueTestMessages(ctx, topicName); err != nil {
			observability.AcknowledgeError(err, l, span, "pruning queue test messages")
		}
	}

	return nil
}
