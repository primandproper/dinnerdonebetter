package events

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	"github.com/primandproper/platform-go/v12/database"
	"github.com/primandproper/platform-go/v12/outbox"
)

/*
Search index events are enqueued the same way, and for the same reason, as the data change
events above them: through the executor of the transaction that changed the row.

They used to be published by the data change consumer, which read an event off the broker,
picked the row ID out of its context map, and published a second event onto the index's topic.
That is a dual write with an extra hop in it — the row commits, the index event fails to
publish, and the index is wrong until the reindex backstop next runs, with nothing in between
able to tell that it is. Enqueued here, an index event lives or dies with the row it describes.

An event names a document; it does not carry one. Whenever the Syncer applies this, and however
many times, it reads the row back and indexes its current state, so redelivery and out-of-order
delivery both converge, and an upsert whose row has since been deleted is applied as a delete
rather than stranding a document nothing will mention again.

The topic is the index: platform-go says which index an event belongs to by where it arrived,
because a searchsync.Event carries a document ID and an operation and nothing else. The
document ID becomes the outbox key, which is what buys per-document ordering — at most one
event per document is ever in flight, however many relays are running.

# Where the events come from

Nothing here decides which write feeds which index. That is a registered outbox side effect,
supplied at construction, and it derives the index events from the data change messages this
Emitter already sends. It used to be an EmitOption every call site passed by hand, which made a
thing every write owes into a thing a call site could forget.
*/

// EmitIndex enqueues the index events a trigger implies, without announcing anything.
//
// Emit is the usual path, because a write worth indexing is nearly always a write worth
// announcing, and the side effect derives the index event from the announcement. This exists for
// the writes where that is not true — where putting the event on the wire would be a decision
// about the public event stream rather than about the index.
//
// It runs the same side effect over the same shape of message, so such a write reads out of the
// same table as every other. The message itself is never enqueued; only what the effect derives
// from it is.
func (e *Emitter) EmitIndex(ctx context.Context, q database.SQLQueryExecutor, trigger string, metadata map[string]any) error {
	if e == nil || e.sideEffect == nil {
		return nil
	}

	derived, err := e.sideEffect(ctx, q, []outbox.Message{{
		Topic:   e.topic,
		Payload: &audit.DataChangeMessage{EventType: trigger, Context: metadata},
	}})
	if err != nil {
		return err
	}

	if len(derived) == 0 {
		return nil
	}

	return e.writer.Enqueue(ctx, q, derived...)
}
