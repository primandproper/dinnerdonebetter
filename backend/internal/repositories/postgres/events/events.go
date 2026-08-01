/*
Package events writes domain events into the outbox as part of the transaction that produced
them.

Publishing an event after a repository commits is two operations against two systems that share
no commit: the row lands, the publish fails, and durable state and the event stream diverge with
nothing to detect it. Every event published from a manager after a repository call has that gap.

The seam is the executor. outbox.Enqueue takes the database.SQLQueryExecutor that
database.Client.WithTransaction hands its callback, so the event is another statement in the
transaction that wrote the row and lives or dies with it. That executor only exists inside the
repository, which is why events are emitted there rather than from the manager — the same place,
and for the same reason, that audit log entries already are.

# Wire compatibility

The message is the same *audit.DataChangeMessage the managers publish directly, marshaled the
same way. The relay republishes the stored bytes as json.RawMessage, so what reaches the broker
is byte-identical to a direct Publish of the same value. Consumers need no change, and a domain
can be converted one method at a time while the rest still publish directly.
*/
package events

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"

	"github.com/primandproper/platform-go/v8/database"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/outbox"
)

// Emitter enqueues data change events into the outbox.
type Emitter struct {
	writer *outbox.Writer
	topic  string
}

// NewEmitter builds an Emitter that writes to the given topic.
//
// A nil writer yields a nil Emitter, which is inert: repositories constructed without an outbox
// keep working and simply emit nothing. That is what lets a repository be built in a test, or
// in a process with no publisher, without threading a mock through every call site.
func NewEmitter(writer *outbox.Writer, topic string) *Emitter {
	if writer == nil || topic == "" {
		return nil
	}

	return &Emitter{writer: writer, topic: topic}
}

// Emit enqueues one data change event using the caller's executor, so it commits with whatever
// else that transaction did.
//
// The user and account are read from the context exactly as the manager-side publish does, so a
// converted call site produces the same message it did before. accountID overrides the one from
// the context and should be passed whenever the repository knows it, because a background job
// has no session: the finalizer reaches the same repository method as a user request does, and
// on that path the context carries nobody. Pass "" only when the event genuinely has no account.
func (e *Emitter) Emit(ctx context.Context, q database.SQLQueryExecutor, logger logging.Logger, eventType, accountID string, metadata map[string]any) error {
	if e == nil {
		return nil
	}

	msg := audit.BuildDataChangeMessageFromContext(ctx, logging.EnsureLogger(logger), eventType, metadata)
	if accountID != "" {
		msg.AccountID = accountID
	}

	return e.writer.Enqueue(ctx, q, outbox.Message{
		Topic:   e.topic,
		Payload: msg,
		// Ordering is per account: two events for the same account publish in the order
		// they were written, and events for different accounts do not wait on each other.
		Key: msg.AccountID,
	})
}
