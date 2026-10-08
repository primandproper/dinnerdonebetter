/*
Package analytics decides which of the application's events are worth reporting to the product
analytics platform, and what each one says when it is.

Every data change event carrying a user ID used to be reported, because the data changes topic
carried every event the application emitted and the consumer had no reason to be selective. That
was never a decision anyone made; it was the absence of one. It sent roughly a hundred and sixty
event types to a third-party vendor, of which all but a dozen were create/update/archive traffic
on catalog tables that no product question is ever asked of, and it did so unfiltered — a
password change and a two-factor secret rotation went the same way a recipe creation did.

# Why an allowlist rather than a denylist

The webhook catalog is a denylist, deliberately: a new domain event nobody remembered to
classify should be deliverable, because that is what webhooks are for, and the events that must
not leak are a known, enumerable set.

The default has to run the other way here. A new event nobody remembered to classify should not
silently become a metric: the cost of forgetting to add one is a question you cannot answer
until you add it, which is recoverable, while the cost of forgetting to leave one out is vendor
spend and a dashboard full of noise that nobody notices is noise. Unclassified means unreported.

# Why this is not the webhook exclusion list

The two answer different questions against different threats. A webhook endpoint is a URL an
account member supplied, so shipping an account's authentication activity there hands a live
security feed to whoever has a foothold in that account. The analytics platform is a single
vendor the operator chose and configured. So the sets are related but neither contains the
other: a registration and an archival are excluded from webhooks and are the two most important
numbers the product has.

# Why each entry says how to read its payload

The events on the data changes topic do not share a shape. This application's own travel as
datachanges.Message, with the actor and the account on the message; platform's travel as the
payload type each platform package declares — identity.UserEvent names its user as UserID,
identity.InvitationEvent names the sender as FromUser and the acceptor as ToUser,
billing.SubscriptionEvent names an account and nobody. The allowlist therefore pairs every event
with the reading that turns its payload into a report, so the one place that decides an event
is worth reporting is also the one place that decides what is said about it, and nothing the
vendor is told is a field nobody chose to send.

Events are named by their constants rather than their strings, so deleting an event type from a
domain fails this package's build rather than leaving a metric that silently stops arriving.
*/
package analytics

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v15/billing"
	"github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/webhooks"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
)

// The property names a report carries. They are platform's JSON field names for the same facts,
// so a property on a platform event's report reads the way the payload did.
const (
	propertyAccountID    = "accountID"
	propertyInvitationID = "invitationID"
	propertyProductID    = "productID"
	propertyStatus       = "status"
)

// Report is what the analytics platform is told about one event.
type Report struct {
	_ struct{} `json:"-"`

	// Properties travel with the event. Never a secret.
	Properties map[string]any
	// UserID is who the event is attributed to. Empty for an event about an account and
	// nobody in particular, where AccountID says which; the reporter then decides who stands
	// in for the account.
	UserID string
	// AccountID is the account the event happened in, where the payload says.
	AccountID string
}

// reading turns one event's payload into its report.
type reading func(payload json.RawMessage) (*Report, error)

// platformEvents is the allowlist's platform half: each of platform's events worth reporting,
// paired with the reading of its payload type.
var platformEvents = map[webhooks.EventType]reading{
	identity.EventUserRegistered:    userEvent,
	identity.EventUserArchived:      userEvent,
	identity.EventAccountCreated:    accountEvent,
	identity.EventInvitationCreated: invitationSent,
	// Both ways an invitation is answered yes: platform announces an accept by an existing
	// user as its own event, and a registration through an invitation as a registration
	// naming the invitation — which userEvent carries as a property.
	identity.EventInvitationAccepted: invitationAccepted,

	billing.EventSubscriptionCreated: subscriptionEvent,
}

// ownEvents is the allowlist's application half: each domain's own events worth reporting,
// contributed by the domain. Every one travels as a datachanges.Message and is read as one.
func ownEvents() [][]string {
	return [][]string{
		// Domain: mealplanning
		mealplanning.AnalyticsEventTypes(),
	}
}

// reportable is the allowlist, each event paired with the reading of its payload.
var reportable = allowlist()

// allowlist merges the two halves. A domain's event that platform also names would be read two
// ways, so the merge refuses one rather than letting map order decide; the event names are
// constants on both sides, so that is a build-time collision caught at package init.
func allowlist() map[webhooks.EventType]reading {
	merged := make(map[webhooks.EventType]reading, len(platformEvents))
	maps.Copy(merged, platformEvents)

	for _, domain := range ownEvents() {
		for _, eventType := range domain {
			if _, taken := merged[webhooks.EventType(eventType)]; taken {
				panic(fmt.Sprintf("analytics: event type %q is allowlisted twice", eventType))
			}

			merged[webhooks.EventType(eventType)] = ownMessage
		}
	}

	return merged
}

// Reportable reports whether eventType is one the analytics platform is told about.
func Reportable(eventType string) bool {
	_, found := reportable[webhooks.EventType(eventType)]

	return found
}

// Read turns an event off the data changes topic into its report. It reports false, and
// nothing else, for an event not on the allowlist; an allowlisted event whose payload does not
// decode as the type its name promises is an error, because that is a wiring fault rather than
// a decision.
func Read(event *webhooks.Envelope) (*Report, bool, error) {
	if event == nil {
		return nil, false, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil event")
	}

	read, found := reportable[event.EventType]
	if !found {
		return nil, false, nil
	}

	report, err := read(event.Payload)
	if err != nil {
		return nil, false, platformerrors.Wrapf(err, "reading a %q event for analytics", event.EventType)
	}

	return report, true, nil
}

// ownMessage reads this application's own event: the actor and account are on the message, and
// the context is the properties, as it has always been.
func ownMessage(payload json.RawMessage) (*Report, error) {
	var msg datachanges.Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil, err
	}

	return &Report{UserID: msg.UserID, AccountID: msg.AccountID, Properties: msg.Context}, nil
}

// userEvent reads an event about a user.
func userEvent(payload json.RawMessage) (*Report, error) {
	var e identity.UserEvent
	if err := json.Unmarshal(payload, &e); err != nil {
		return nil, err
	}

	properties := map[string]any{}
	if e.AccountID != "" {
		properties[propertyAccountID] = e.AccountID
	}

	if e.InvitationID != "" {
		properties[propertyInvitationID] = e.InvitationID
	}

	return &Report{UserID: e.UserID, AccountID: e.AccountID, Properties: properties}, nil
}

// accountEvent reads an event about an account, attributed to its owner.
func accountEvent(payload json.RawMessage) (*Report, error) {
	var e identity.AccountEvent
	if err := json.Unmarshal(payload, &e); err != nil {
		return nil, err
	}

	return &Report{
		UserID:     e.OwnerUserID,
		AccountID:  e.AccountID,
		Properties: map[string]any{propertyAccountID: e.AccountID},
	}, nil
}

// invitationSent reads an invitation issued, attributed to whoever sent it.
func invitationSent(payload json.RawMessage) (*Report, error) {
	e, err := invitationEvent(payload)
	if err != nil {
		return nil, err
	}

	return &Report{UserID: e.FromUser, AccountID: e.AccountID, Properties: invitationProperties(e)}, nil
}

// invitationAccepted reads an invitation answered yes, attributed to whoever answered.
func invitationAccepted(payload json.RawMessage) (*Report, error) {
	e, err := invitationEvent(payload)
	if err != nil {
		return nil, err
	}

	var userID string
	if e.ToUser != nil {
		userID = *e.ToUser
	}

	return &Report{UserID: userID, AccountID: e.AccountID, Properties: invitationProperties(e)}, nil
}

func invitationEvent(payload json.RawMessage) (*identity.InvitationEvent, error) {
	var e identity.InvitationEvent
	if err := json.Unmarshal(payload, &e); err != nil {
		return nil, err
	}

	return &e, nil
}

func invitationProperties(e *identity.InvitationEvent) map[string]any {
	return map[string]any{propertyAccountID: e.AccountID, propertyInvitationID: e.InvitationID}
}

// subscriptionEvent reads a subscription opened. It names the account and nobody: a
// subscription is an account's, and who to attribute it to is the reporter's to decide.
func subscriptionEvent(payload json.RawMessage) (*Report, error) {
	var e billing.SubscriptionEvent
	if err := json.Unmarshal(payload, &e); err != nil {
		return nil, err
	}

	return &Report{
		AccountID: e.AccountID,
		Properties: map[string]any{
			propertyAccountID: e.AccountID,
			propertyProductID: e.ProductID,
			propertyStatus:    string(e.Status),
		},
	}, nil
}
