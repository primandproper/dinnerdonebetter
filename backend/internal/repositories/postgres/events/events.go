/*
Package events is this application's seam onto platform's recording spine: the outbox writer,
the webhooks Emitter that publishes a domain event and fans it out to subscribers on the caller's
transaction, and the recording.Recorder that writes an audit entry beside it.

Publishing an event after a repository commits is two operations against two systems that share
no commit: the row lands, the publish fails, and durable state and the event stream diverge with
nothing to detect it. Every write here goes through the executor of the transaction that wrote
the row, so the event lives or dies with it. That guarantee is platform's — outbox.Enqueue and
webhooks.Emitter.Emit both take the database.Tx — and nothing here re-states it.

What this package adds is the shape of this application's own events. A write to one of its
nouns announces itself as a *datachanges.Message — the event type, a context map, and the
session's user and account — because that is what the async message handler, the analytics
allowlist and the search index rules read. platform's Emitter takes any payload and never
interprets it; this is where that payload is built from the request, once, rather than at a
hundred and fifty call sites.

The platform-owned nouns do not come through here. Their stores' RecordingHooks record through
the same Recorder with platform's own payloads, which the broker cannot yet tell apart from these
(platform-go#1130); until it can, the stores whose events the async handler must recognize keep
their own hooks.

# Wire compatibility

The message is the same *datachanges.Message the handler has always decoded, marshaled the same
way. The relay republishes the stored bytes, so what reaches the broker is byte-identical to what
the local emitter this replaced produced.
*/
package events

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/platform-go/v15/outbox"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// Emitter publishes this application's data change events through platform's Emitter, and
// records the audit entry a write owes beside one through platform's Recorder.
//
// It is one type over the two because a write here owes both or one, never neither, and the
// payload it builds for the event is the same either way.
type Emitter struct {
	emitter  *webhooks.Emitter
	recorder *platformrecording.Recorder
	writer   *outbox.Writer
	// effect is the same side effect registered on the writer, kept so EmitIndex can run it
	// over a message it never enqueues. See index_events.go.
	effect outbox.SideEffect
}

// NewEmitter builds an Emitter over platform's.
//
// Every part is required. The emitter this replaced was nil-inert for a process with no topic,
// which made a process with no broker a process whose writes announced nothing: an event written
// to the outbox under the default topic is relayed by whichever worker has a broker, which is
// what an outbox is for.
func NewEmitter(emitter *webhooks.Emitter, recorder *platformrecording.Recorder, writer *outbox.Writer, effect outbox.SideEffect) (*Emitter, error) {
	switch {
	case emitter == nil:
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil webhooks emitter")
	case recorder == nil:
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil recording recorder")
	case writer == nil:
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil outbox writer")
	case effect == nil:
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil index side effect")
	}

	return &Emitter{emitter: emitter, recorder: recorder, writer: writer, effect: effect}, nil
}

// Recorder is the platform Recorder this Emitter records through, for a store whose
// RecordingHooks are built over it outside the injector — the local dev server and the
// integration suite's seeding build their stores by hand.
func (e *Emitter) Recorder() *platformrecording.Recorder { return e.recorder }

// EmitOption customizes one Emit.
type EmitOption func(*emitConfig)

type emitConfig struct {
	orderingKey string
	userID      string
}

// WithOrderingKey sets the ordering key for this event, overriding the default of the scope's
// own identifier — the account's.
//
// Deliveries sharing a key reach a given endpoint in dispatch order, and outbox messages sharing
// one publish in order, so this should be the subject resource's ID wherever the caller knows it:
// that is what stops a resource.updated overtaking the resource.created for the same resource.
// Callers that do not pass one get per-account ordering, which is correct but serializes an
// account's events more than it needs to.
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

// WithUserID attributes this event to the given user, overriding the one read from the context.
//
// It is for the writes made before anybody has a session: a sign-in is recorded inside the
// transaction that proves it, and the context there carries nobody yet, so the user the event
// is about has to be named rather than read.
func WithUserID(userID string) EmitOption {
	return func(c *emitConfig) {
		if userID != "" {
			c.userID = userID
		}
	}
}

// Emit publishes one data change event on the caller's transaction and fans it out to the
// account's webhook subscribers, so it commits with whatever else that transaction did.
//
// The user and account are read from the context. accountID overrides the account and should be
// passed whenever the repository knows it, because a background job has no session: the finalizer
// reaches the same repository method as a user request does, and on that path the context carries
// nobody. Pass "" only when the event genuinely has no account; it is then published in the global
// scope, where no endpoint lives, so it reaches the broker and no subscriber.
func (e *Emitter) Emit(ctx context.Context, tx database.Tx, logger logging.Logger, eventType, accountID string, metadata map[string]any, opts ...EmitOption) error {
	msg, event := e.event(ctx, logger, eventType, accountID, metadata, opts)

	return e.emitter.Emit(ctx, tx, scopeFor(msg.AccountID), event)
}

// Record writes entry to the audit log and publishes the event describing the same write, both
// on the caller's transaction, through platform's Recorder.
//
// The entry is one audit.NewEntry built: its Scope is the chain this application's attribution
// rule chose for it — the account's where there is one, the user's otherwise — and the event fans
// out within that same scope. Who did it is the principal on the context, which is platform's
// rule and the right one: an entry about a comment names the requester, not the comment's author.
// The one write that reaches here with no principal is the one that establishes who is acting —
// a sign-in, a registration — and audit.NewEntry has named the user on the entry instead; that is
// the case RecordAs exists for, and the only one it is used in.
func (e *Emitter) Record(
	ctx context.Context,
	tx database.Tx,
	logger logging.Logger,
	entry *platformaudit.Entry,
	eventType, accountID string,
	metadata map[string]any,
	opts ...EmitOption,
) error {
	if entry == nil {
		return platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil audit entry")
	}

	msg, event := e.event(ctx, logger, eventType, accountID, metadata, opts)

	scope := entry.Scope
	if scope == (tenancy.Scope{}) {
		scope = scopeFor(msg.AccountID)
	}

	recorded := &platformrecording.Entry{
		ResourceType: entry.ResourceType,
		ResourceID:   entry.ResourceID,
		EventType:    entry.EventType,
		Changes:      entry.Changes,
		Metadata:     entry.Metadata,
	}

	if _, present := sessions.PrincipalFromContext(ctx); !present && entry.Actor.Type == platformaudit.ActorUser && entry.Actor.ID != "" {
		return e.recorder.RecordAs(ctx, tx, scope, entry.Actor, event, recorded)
	}

	return e.recorder.Record(ctx, tx, scope, event, recorded)
}

// event builds the message a data change event carries and the platform event around it.
func (e *Emitter) event(ctx context.Context, logger logging.Logger, eventType, accountID string, metadata map[string]any, opts []EmitOption) (*datachanges.Message, *webhooks.Event) {
	msg := datachanges.MessageFromContext(ctx, logging.EnsureLogger(logger), eventType, metadata)
	if accountID != "" {
		msg.AccountID = accountID
	}

	cfg := &emitConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	if cfg.userID != "" {
		msg.UserID = cfg.userID
	}

	return msg, &webhooks.Event{
		EventType:   webhooks.EventType(eventType),
		OrderingKey: cfg.orderingKey,
		Payload:     msg,
	}
}

// scopeFor is the scope an event is published and fanned out in: the account's, or the global
// one for an event that happened in no account.
func scopeFor(accountID string) tenancy.Scope {
	if accountID == "" {
		return tenancy.Global()
	}

	return tenancy.Of(accountID)
}
