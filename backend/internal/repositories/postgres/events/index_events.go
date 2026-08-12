package events

import (
	"context"

	"github.com/primandproper/platform-go/v10/database"
	searchsync "github.com/primandproper/platform-go/v10/search/sync"
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
*/

// WithIndexUpsert says this write left a row that the index named by topic should hold.
//
// topic is the index's name, which the indexing packages declare as their IndexType constants;
// documentID is the row's ID, which must be the ID the index holds the document under.
func WithIndexUpsert(topic, documentID string) EmitOption {
	return withIndexEvent(topic, documentID, searchsync.OpUpsert)
}

// WithIndexDelete says this write removed the row, so the index should stop holding it.
//
// Archival counts as removal: an archived row is one the search index must not return, and the
// Syncer applies a delete without reading anything back.
func WithIndexDelete(topic, documentID string) EmitOption {
	return withIndexEvent(topic, documentID, searchsync.OpDelete)
}

func withIndexEvent(topic, documentID string, op searchsync.Op) EmitOption {
	return func(c *emitConfig) {
		c.indexEvents = append(c.indexEvents, searchsync.NewEvent(op, documentID).Message(topic))
	}
}

// EmitIndex enqueues index events for a write that emits no data change event of its own.
//
// Emit is the usual path, because a write worth indexing is nearly always a write worth
// announcing, and passing the index event as an option to it keeps the two from being written
// apart. This exists for the writes where that is not true — where announcing the change would
// put an event on the wire that nothing puts there today, which is a decision about the public
// event stream rather than about the index.
//
// It takes the same EmitOptions as Emit and reads only the index ones, so a call site does not
// have to know which of the two it is using.
func (e *Emitter) EmitIndex(ctx context.Context, q database.SQLQueryExecutor, opts ...EmitOption) error {
	if e == nil {
		return nil
	}

	cfg := &emitConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	if len(cfg.indexEvents) == 0 {
		return nil
	}

	return e.writer.Enqueue(ctx, q, cfg.indexEvents...)
}
