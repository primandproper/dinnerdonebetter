package internalops

import (
	"context"
	"time"
)

type (
	QueueTestMessage struct {
		CreatedAt      time.Time
		AcknowledgedAt *time.Time
		ID             string
		QueueName      string
	}

	InternalOpsDataManager interface {
		CreateQueueTestMessage(ctx context.Context, id, queueName string) error
		AcknowledgeQueueTestMessage(ctx context.Context, id string) error
		GetQueueTestMessage(ctx context.Context, id string) (*QueueTestMessage, error)
		PruneQueueTestMessages(ctx context.Context, queueName string) error
	}
)
