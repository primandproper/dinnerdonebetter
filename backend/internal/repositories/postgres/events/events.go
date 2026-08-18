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
	"encoding/json"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhookdispatch"

	"github.com/primandproper/platform-go/v10/database"
	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/outbox"
)

// Emitter enqueues data change events into the outbox and fans them out to webhooks.
type Emitter struct {
	writer     *outbox.Writer
	dispatcher *webhookdispatch.Dispatcher
	// sideEffect is the same effect registered on the writer, kept so EmitIndex can run it
	// over a message it never enqueues. See index_events.go.
	sideEffect outbox.SideEffect
	topic      string
}

// NewEmitter builds an Emitter that writes to the given topic.
//
// A nil writer yields a nil Emitter, which is inert: repositories constructed without an outbox
// keep working and simply emit nothing. That is what lets a repository be built in a test, or
// in a process with no publisher, without threading a mock through every call site.
//
// The dispatcher is separately optional and separately nil-safe, for the same reason at a
// different granularity: a process with an outbox but no webhook tables still emits events.
func NewEmitter(writer *outbox.Writer, topic string, dispatcher *webhookdispatch.Dispatcher, sideEffect outbox.SideEffect) *Emitter {
	if writer == nil || topic == "" {
		return nil
	}

	return &Emitter{writer: writer, topic: topic, dispatcher: dispatcher, sideEffect: sideEffect}
}

// EmitOption customizes one Emit.
type EmitOption func(*emitConfig)

type emitConfig struct {
	orderingKey string
}

// WithOrderingKey sets the webhook ordering key for this event, overriding the default of the
// account ID.
//
// Deliveries sharing a key reach a given endpoint in dispatch order, so this should be the
// subject resource's ID wherever the caller knows it: that is what stops a resource.updated
// overtaking the resource.created for the same resource. Callers that do not pass one get
// per-account ordering, which is correct but serializes an account's deliveries to a subscriber
// more than it needs to.
//
// It is an option rather than a parameter because roughly a hundred and fifty call sites emit
// events and only some of them know their subject.
func WithOrderingKey(key string) EmitOption {
	return func(c *emitConfig) {
		if key != "" {
			c.orderingKey = key
		}
	}
}

// Emit enqueues one data change event using the caller's executor, so it commits with whatever
// else that transaction did.
//
// The user and account are read from the context exactly as the manager-side publish does, so a
// converted call site produces the same message it did before. accountID overrides the one from
// the context and should be passed whenever the repository knows it, because a background job
// has no session: the finalizer reaches the same repository method as a user request does, and
// on that path the context carries nobody. Pass "" only when the event genuinely has no account.
func (e *Emitter) Emit(ctx context.Context, q database.SQLQueryExecutor, logger logging.Logger, eventType, accountID string, metadata map[string]any, opts ...EmitOption) error {
	if e == nil {
		return nil
	}

	msg := audit.BuildDataChangeMessageFromContext(ctx, logging.EnsureLogger(logger), eventType, metadata)
	if accountID != "" {
		msg.AccountID = accountID
	}

	cfg := &emitConfig{
		// Per-account webhook ordering by default, matching the outbox key below: an
		// account's deliveries to one endpoint arrive in the order they were written, and
		// different accounts never wait on each other.
		orderingKey: msg.AccountID,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	// The data change event is enqueued inside the caller's transaction, and the registered
	// side effect derives this write's index events from it in the same statement. That is what
	// keeps a search index from diverging from the row the way it could when index events were
	// published by a consumer downstream of the broker: there, the row committed and the index
	// event was a second, unrelated write that could fail on its own, and nothing noticed until
	// the next reindex.
	msg2 := outbox.Message{
		Topic:   e.topic,
		Payload: msg,
		// Ordering is per account: two events for the same account publish in the order
		// they were written, and events for different accounts do not wait on each other.
		Key: msg.AccountID,
	}

	if err := e.writer.Enqueue(ctx, q, msg2); err != nil {
		return err
	}

	return e.dispatchWebhooks(ctx, q, msg, cfg)
}

// dispatchWebhooks fans the same message out to the account's webhook subscribers, through the
// caller's executor.
//
// This is where webhook delivery became transactional. It used to happen in the async message
// handler, downstream of the broker: the row committed, the event was published, a consumer
// resolved subscribers and published one execution request per subscriber, and each of those
// steps could fail independently of the write that caused them. Now the dispatch rows are
// further statements in the transaction that wrote the row, so a delivery and the state change
// it describes commit together or not at all.
//
// The cost is on the same ledger: a webhook table failure now fails the business transaction. It
// is the same trade the outbox already makes one line above, and for the same reason — the
// alternative is durable state and delivery diverging with nothing able to detect it.
func (e *Emitter) dispatchWebhooks(ctx context.Context, q database.SQLQueryExecutor, msg *audit.DataChangeMessage, cfg *emitConfig) error {
	// The payload is the same *audit.DataChangeMessage the broker carries, marshaled once. A
	// subscriber and a queue consumer therefore see byte-identical bodies, and the bytes signed
	// are the bytes sent — re-marshaling between dispatch and delivery is exactly how a
	// signature comes to cover something other than the request body.
	payload, err := json.Marshal(msg)
	if err != nil {
		return platformerrors.Wrap(err, "marshaling webhook payload")
	}

	return e.dispatcher.Dispatch(ctx, q, msg.AccountID, msg.EventType, cfg.orderingKey, payload)
}
