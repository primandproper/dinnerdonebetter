/*
Package datachanges is the data change event: what a write enqueues on the outbox,
beside its audit entry and in the same transaction, for the async message handler,
search indexing and webhook dispatch to act on afterwards.

It is not the audit log, though the two are written together. An audit entry is the
durable record of what happened; a data change message is a request for something
else to happen because of it, and is gone once it has been handled.
*/
package datachanges

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	"github.com/primandproper/platform-go/v15/searchsync"
	"github.com/primandproper/primitives-go/v2/observability/logging"
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
	TestID    string         `json:"testID,omitempty"`
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
