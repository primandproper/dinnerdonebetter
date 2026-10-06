/*
Package datachanges is this application's own data change event: what a write to one of
its own tables enqueues on the outbox, beside its audit entry and in the same
transaction, for the async message handler, search indexing and webhook dispatch to
act on afterwards.

It is not the audit log, though the two are written together. An audit entry is the
durable record of what happened; a data change message is a request for something
else to happen because of it, and is gone once it has been handled.

It is also not the only payload on the data changes topic. Every event platform's
stores record travels under platform's own payload types — identity.UserEvent,
billing.SubscriptionEvent and the rest — and every event, this application's or
platform's, reaches the broker inside a webhooks.Envelope that names it. A consumer
reads the envelope's event type and decodes the payload into the type that event
names; for this application's own events that type is Message.
*/
package datachanges

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	"github.com/primandproper/platform-go/v15/searchsync"
	"github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// A Message is what platform's search index bridge reads off the outbox: the
// event type names the rule and the context holds the document's ID under the
// key the rule names. See internal/indexevents for the rules.
var _ searchsync.Change = (*Message)(nil)

// Message is one data change event, as it travels through the outbox.
type Message struct {
	_ struct{} `json:"-"`

	EventType string         `json:"eventType"`
	Context   map[string]any `json:"context,omitempty"`
	UserID    string         `json:"userID"`
	AccountID string         `json:"accountID,omitempty"`
}

// IndexEventType is the event type an index rule matches on.
func (m *Message) IndexEventType() string { return m.EventType }

// IndexDocumentID reads the document ID an index rule names out of the context.
// A value under the key that is not a string, or no value at all, is reported as
// absent, which the bridge refuses rather than indexing nothing.
func (m *Message) IndexDocumentID(key string) (string, bool) {
	if m.Context == nil {
		return "", false
	}

	id, ok := m.Context[key].(string)

	return id, ok
}

// MessageFromContext builds a message attributed to the request's session.
func MessageFromContext(ctx context.Context, logger logging.Logger, eventType string, metadata map[string]any) *Message {
	sessionContext := sessions.FromContext(ctx)
	if sessionContext == nil {
		logger.WithValue("event_type", eventType).Info("failed to extract session data from context")
	}

	// The getters are nil-safe, so an absent session yields empty attribution rather than a panic.
	return &Message{
		EventType: eventType,
		Context:   metadata,
		UserID:    sessionContext.GetUserID(),
		AccountID: sessionContext.GetActiveAccountID(),
	}
}

// Event builds the platform event one of this application's writes announces itself with: a
// Message attributed to the request's session, under eventType, for platform's webhooks.Emitter
// to publish and fan out on the writer's transaction.
//
// accountID overrides the session's account and should be passed whenever the writer knows it,
// because a background job has no session: the finalizer reaches the same repository method a
// user request does, and on that path the context carries nobody. Pass "" only when the event
// genuinely has no account.
//
// This is the whole of what this application adds to platform's Emitter. The payload is the
// shape the async message handler, the analytics allowlist and the search index rules read, and
// building it here is what keeps the hundred-odd writes that announce themselves from each
// building it again.
func Event(ctx context.Context, logger logging.Logger, eventType, accountID string, metadata map[string]any) (*webhooks.Event, *Message) {
	msg := MessageFromContext(ctx, logging.EnsureLogger(logger), eventType, metadata)
	if accountID != "" {
		msg.AccountID = accountID
	}

	return &webhooks.Event{EventType: webhooks.EventType(eventType), Payload: msg}, msg
}

// Scope is the scope an event is published and fanned out in: the account's, or the global one
// for an event that happened in no account — where no endpoint lives, so it reaches the broker
// and no subscriber.
func Scope(accountID string) tenancy.Scope {
	if accountID == "" {
		return tenancy.Global()
	}

	return tenancy.Of(accountID)
}
