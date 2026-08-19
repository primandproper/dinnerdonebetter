/*
This file is the translation boundary between this application's webhook vocabulary and
platform-go's, and it is the only place either language is spoken in terms of the other.

The application's Webhook is an account-owned resource: a name, an owner, archival timestamps,
and identified trigger config rows. The platform's Endpoint is a delivery target: a URL, a
content type, a signing secret, and a flat list of subscription strings. Neither is derivable
from the other, and almost all of the gap between them is tenancy — see the divergence note on
Webhook in internal/domain/webhooks/webhook.go for why both types exist.

The rule that makes this a file rather than a struct literal in the middle of Register is the one
qualify encodes. The platform's Catalog is global by construction — two accounts subscribing to
"meal_plan_created" are one key to it — so the account has to be folded into the event type here
or the fan-out delivers every account's events to every account's endpoints. That is the same
class of decision as audit.ScopeFor, and it gets the same treatment: one implementation, called
from one place, because a rule applied at several call sites is a rule applied inconsistently.

# There is no fromPlatformEndpoint, and that is deliberate

Nothing reads an Endpoint back out as a Webhook. The two stores hold different things: the
platform's holds dispatch state — signing secret, retry counters, circuit breaker, attempt log —
and this application's webhooks and webhook_trigger_configs tables hold the account-facing
resource, including the trigger config identities and archival timestamps the API returns and the
platform has no column for. Every read path in internal/repositories/postgres/webhooks therefore
goes to this application's own tables, and reconstructing a Webhook from an Endpoint would mean
inventing the fields the generalization dropped.

The only reverse translation that exists is unqualify, and it recovers an event type for
setAccountSubscriptions' read-modify-write rather than for any read path.
*/

package webhookdispatch

import (
	"strings"

	platformerrors "github.com/primandproper/platform-go/v11/errors"
	"github.com/primandproper/platform-go/v11/tenancy"
	"github.com/primandproper/platform-go/v11/webhooks"
)

// accountSeparator joins an account to an event type in a subscription row.
//
// A colon, because neither half can contain one: account IDs are the alphanumeric identifiers
// identifiers.New mints, and event types are Go constants matching [a-z0-9_]+. That is what makes
// the join unambiguous — a separator either side could contain would let two different pairs
// render to the same subscription string, and events would reach the wrong account's endpoint.
const accountSeparator = ":"

// scope is the tenant dimension every stored record and every store call in this package
// carries.
//
// It is Global, and that is behavior-preserving rather than an omission. platform-go's webhooks
// tables grew an owner dimension in /v11, and this package had already solved the same problem a
// different way: the account travels inside the subscription's event type, as
// <accountID>:<eventType>, which is what the qualify/unqualify pair below is for. Two account
// filters would be one too many. Global matches only itself, so the added dimension selects
// every row this package writes and excludes nothing it would otherwise have read.
//
// Adopting the dimension properly means deleting qualify/unqualify, backfilling real account IDs
// into the scope column, and then deleting most of this package in favor of the platform's own
// Dispatcher — which is what the owner dimension was added for, and which is #1303. It is not
// this change.
func scope() tenancy.Scope {
	return tenancy.Global()
}

// qualify renders the subscription string for one account's interest in one event type.
func qualify(accountID, eventType string) webhooks.EventType {
	return webhooks.EventType(accountID + accountSeparator + eventType)
}

// unqualify recovers the event type from a subscription string, and reports whether it belonged
// to the given account. A subscription for another account returns false, which is what makes
// this safe to map over an endpoint's whole subscription set.
func unqualify(accountID string, subscription webhooks.EventType) (eventType string, ok bool) {
	return strings.CutPrefix(subscription.String(), accountID+accountSeparator)
}

// checkKnown rejects event types outside the application's catalog.
//
// The check is on the unqualified type, which is the one the catalog is about: the qualified
// string qualify produces can never be in a catalog, because a catalog holds unqualified types
// and would need one entry per account to hold these.
//
// It is separate from qualifyAll so that a caller which has to read before it can write — see
// Dispatcher.SetEventTypes — can refuse a bad request without paying for the query.
func (d *Dispatcher) checkKnown(eventTypes []string) error {
	for _, eventType := range eventTypes {
		if !d.catalog.Known(webhooks.EventType(eventType)) {
			return platformerrors.Wrapf(ErrUnknownEventType, "event type %q", eventType)
		}
	}

	return nil
}

// qualifyAll vets event types against the catalog and renders them for storage.
func (d *Dispatcher) qualifyAll(accountID string, eventTypes []string) ([]webhooks.EventType, error) {
	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	if err := d.checkKnown(eventTypes); err != nil {
		return nil, err
	}

	// An empty set is allowed here, producing an endpoint that is registered and subscribed to
	// nothing. That is a real state — a webhook adopted from before delivery worked starts in
	// it, and unsubscribing from the last event type returns to it — and it is inert rather
	// than dangerous, because fan-out reads the subscriptions table. The rule that a webhook is
	// created subscribed to something is a request-validation rule, and lives there.
	qualified := make([]webhooks.EventType, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		qualified = append(qualified, qualify(accountID, eventType))
	}

	return qualified, nil
}

// toPlatformEndpoint renders a registration in the platform's vocabulary.
//
// The endpoint comes back subscribed to nothing: setAccountSubscriptions is what puts event types
// on one, so that the qualification rule has a single implementation whether the endpoint is
// being created or amended. Registration's Name, owner, and audit fields have no counterpart
// here on purpose — the platform stores what it takes to deliver, and nothing about whose it is.
func toPlatformEndpoint(registration *Registration, secret []byte) *webhooks.Endpoint {
	endpoint := &webhooks.Endpoint{
		Scope:       scope(),
		ID:          registration.ID,
		URL:         registration.URL,
		ContentType: registration.ContentType,
		Secret:      webhooks.Secret{Current: secret},
	}
	endpoint.EnsureDefaults()

	return endpoint
}

// setAccountSubscriptions makes the account's subscriptions on an endpoint exactly eventTypes,
// leaving every other account's alone. It is the only writer of Endpoint.Events.
//
// The filter-then-append is not defensiveness: SaveEndpoint replaces an endpoint's whole
// subscription set, and this application never owns all of it, because the set is keyed by
// account. Rewriting it from one account's event types would drop any subscription belonging to
// another. In practice an endpoint has exactly one account — its webhook's — but relying on that
// here would make a future multi-account endpoint silently lose subscriptions.
func (d *Dispatcher) setAccountSubscriptions(endpoint *webhooks.Endpoint, accountID string, eventTypes []string) error {
	qualified, err := d.qualifyAll(accountID, eventTypes)
	if err != nil {
		return err
	}

	subscriptions := make([]webhooks.EventType, 0, len(endpoint.Events)+len(qualified))
	for _, subscription := range endpoint.Events {
		if _, mine := unqualify(accountID, subscription); !mine {
			subscriptions = append(subscriptions, subscription)
		}
	}

	subscriptions = append(subscriptions, qualified...)

	endpoint.Events = subscriptions

	return nil
}
