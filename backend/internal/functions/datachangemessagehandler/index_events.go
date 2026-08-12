package datachangemessagehandler

import (
	"context"

	platformerrors "github.com/primandproper/platform-go/v10/errors"
	searchsync "github.com/primandproper/platform-go/v10/search/sync"
)

// publishIndexEvent tells the index named by indexType that documentID changed.
//
// One topic per index, because that is how platform-go v10 says which index an event belongs
// to: a searchsync.Event carries a document ID and an operation and nothing else. The event is
// keyed by document ID, which is what buys per-document ordering — the broker admits a keyed
// message only while no older message with that key is pending, so at most one event per
// document is in flight however many replicas are publishing.
//
// The event names the document; it does not carry it. Whenever the Syncer applies this, and
// however many times, it reads the row back and indexes its current state — so redelivery and
// out-of-order delivery both converge, and an upsert whose row has since been deleted is
// applied as a delete rather than stranding a document nothing will mention again.
//
// This is still published from the data-change consumer rather than from the transaction that
// changed the row, which means it is still a dual write: a row can commit and its event fail to
// publish, and nothing detects the divergence until the next reindex. Closing that is a matter
// of calling outbox.Writer.Enqueue with the caller's executor inside each write transaction,
// which is what searchsync.Event.Message exists for — the events and the Syncers on the other
// end of them do not change.
func (a *AsyncDataChangeMessageHandler) publishIndexEvent(ctx context.Context, indexType, documentID string, deleted bool) error {
	publisher, ok := a.searchIndexPublishers[indexType]
	if !ok {
		return platformerrors.Newf("no publisher for search index %q", indexType)
	}

	op := searchsync.OpUpsert
	if deleted {
		op = searchsync.OpDelete
	}

	return publisher.Publish(ctx, searchsync.NewEvent(op, documentID))
}
