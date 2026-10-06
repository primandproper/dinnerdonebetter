package events

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"

	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/primitives-go/v2/database"
)

/*
Search index events are enqueued the same way, and for the same reason, as the data change
events above them: through the executor of the transaction that changed the row.

Nothing here decides which write feeds which index. That is platform's searchsync side effect,
registered on the outbox Writer and built from the rules in internal/indexevents; it derives the
index events from the data change messages this Emitter already sends, so an index event is a
thing every write owes rather than a thing a call site could forget.
*/

// EmitIndex enqueues the index events a trigger implies, without announcing anything.
//
// Emit is the usual path, because a write worth indexing is nearly always a write worth
// announcing, and the side effect derives the index event from the announcement. This exists for
// the writes where that is not true — where putting the event on the wire would be a decision
// about the public event stream rather than about the index.
//
// It is the writer's EnqueueDerived over the same shape of message every announced write sends,
// so such a write reads out of the same table as every other and the rules it is matched against
// are the ones registered on the writer, not a copy kept here. The message itself is never
// enqueued; only what the side effects derive from it is, and a trigger that derives nothing
// writes nothing.
func (e *Emitter) EmitIndex(ctx context.Context, tx database.Tx, trigger string, metadata map[string]any) error {
	return e.writer.EnqueueDerived(ctx, tx, outbox.Message{
		Payload: &datachanges.Message{EventType: trigger, Context: metadata},
	})
}
