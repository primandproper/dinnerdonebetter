/*
Package webhookdispatch is this application's write side for outbound webhooks: it registers
endpoints and fans events out to them, on top of platform-go's webhooks Store.

# Why not webhooks.Dispatcher

platform-go ships a Dispatcher that does exactly this, and it is not usable here for one reason:
it has no tenant dimension. Its Dispatch resolves subscribers with EndpointsForEvent(eventType)
and delivers to every endpoint subscribed to that type, which is right for an application whose
events are global and wrong for one where a webhook belongs to an account. Used as shipped, one
account's meal_plan_created would be delivered to every other account's endpoints.

The account therefore travels in the subscription's event type. A subscription row is written as

	<accountID>:<eventType>

so EndpointsForEvent is account-scoped by construction rather than by a filter someone can forget
to apply, and the platform's subscriptions table remains the single answer to "who wants this
event". That string cannot pass the Catalog gate webhooks.Dispatcher applies — the catalog holds
unqualified types and cannot enumerate one per account — so this package drives the Store
directly and applies the gate itself, to the unqualified type, which is the one the gate is
about.

What that costs is Dispatcher's own two instruments, re-emitted here under the same names.
Everything the migration was for lives below the Store: signing, per-endpoint retry state,
ordering, circuit breaking, the attempt log, replay, and reaping.

# Registration is not transactional; dispatch is

Dispatch takes the caller's executor, so deliveries commit with the state change that caused
them. Registration cannot: Store.SaveEndpoint and Store.ArchiveEndpoint own their statements and
offer no seam to pass an executor through.

So registration is ordered to fail safe instead, and the order is opposite for the two
directions. See Register and Archive.
*/
package webhookdispatch

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/primandproper/platform-go/v9/clock"
	"github.com/primandproper/platform-go/v9/database"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/identifiers"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	"github.com/primandproper/platform-go/v9/webhooks"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const o11yName = "webhook_dispatcher"

// accountSeparator joins an account to an event type in a subscription row.
//
// A colon, because neither half can contain one: account IDs are the alphanumeric identifiers
// identifiers.New mints, and event types are Go constants matching [a-z0-9_]+. That is what makes
// the join unambiguous — a separator either side could contain would let two different pairs
// render to the same subscription string, and events would reach the wrong account's endpoint.
const accountSeparator = ":"

// secretBytes is the size of a generated endpoint signing secret. 32 bytes matches the output
// width of the SHA-256 the signature is computed with; a longer key buys nothing against HMAC.
const secretBytes = 32

var (
	// ErrNoDispatcher indicates a webhook operation attempted in a process with no webhook
	// wiring. Registration fails on it rather than silently doing nothing, because a user who
	// asked for a webhook and was told it was created would otherwise never learn it was not.
	ErrNoDispatcher = platformerrors.New("no webhook dispatcher configured")

	// ErrUnknownEventType indicates an event type outside the application's catalog. It wraps
	// the platform's error of the same meaning, so a caller may check either.
	ErrUnknownEventType = platformerrors.Wrap(webhooks.ErrUnknownEventType, "event type is not in the webhook catalog")

	// ErrEndpointNotRegistered indicates an operation on a webhook that has no delivery
	// endpoint. That is the state a webhook created before delivery worked is in: the
	// migration that adopted these tables backfilled no endpoints, because minting a signing
	// secret in SQL would have meant a non-CSPRNG one. Rotating the secret registers it.
	ErrEndpointNotRegistered = platformerrors.New("webhook has no delivery endpoint; rotate its signing secret to register one")
)

// Registration is one webhook as a delivery target: where it goes, and what it wants.
//
// It carries no name, owner, or audit fields. Those belong to the application's own webhooks
// table, which is the account-facing record; this is only what is needed to deliver.
type Registration struct {
	_ struct{} `json:"-"`

	// ID is the endpoint's ID, and is deliberately the webhook's own ID rather than a
	// separate identifier. One row per webhook in each of two tables, joined by nothing but a
	// shared key, is what keeps "the endpoint for this webhook" from being a lookup that can
	// return the wrong answer.
	ID string
	// AccountID scopes every subscription this endpoint gets.
	AccountID string
	// URL is the https:// target. Vetted by webhooks.CheckEndpointURL before it is stored.
	URL string
	// ContentType is the request's Content-Type.
	ContentType string
	// EventTypes are unqualified catalog event types. Qualified on the way to storage.
	EventTypes []string
}

// Dispatcher registers webhook endpoints and fans events out to them.
type Dispatcher struct {
	store    webhooks.Store
	catalog  webhooks.Catalog
	clock    clock.Clock
	checkURL webhooks.URLChecker
	tracer   tracing.Tracer
	logger   logging.Logger

	dispatchedCounter metrics.Int64Counter
	fanoutHist        metrics.Float64Histogram
}

// Option customizes a Dispatcher.
type Option func(*Dispatcher)

// WithURLChecker replaces the SSRF policy applied at registration.
//
// It exists because webhooks.CheckEndpointURL resolves the host through DNS and refuses anything
// that is not globally routable, which is exactly right in production and unusable in a test:
// an integration test delivers to an httptest server on loopback, and a unit test would
// otherwise depend on the resolver.
//
// Replacing it means owning the SSRF question yourself. The replacement is the only thing between
// a user-supplied URL and an authenticated request from inside your network, so outside a test it
// should be an allowlist of hosts you operate rather than a function that returns nil. Pair it
// with the worker's WithWorkerURLChecker: an endpoint accepted here and refused at delivery sits
// in the backlog until it dies.
func WithURLChecker(checker webhooks.URLChecker) Option {
	return func(d *Dispatcher) {
		if checker != nil {
			d.checkURL = checker
		}
	}
}

// NewDispatcher builds a Dispatcher over the given Store and catalog.
func NewDispatcher(
	store webhooks.Store,
	catalog webhooks.Catalog,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	opts ...Option,
) (*Dispatcher, error) {
	if store == nil {
		return nil, webhooks.ErrNilStore
	}

	if len(catalog) == 0 {
		// An empty catalog rejects every event type, so a Dispatcher built with one would
		// accept no subscriptions and dispatch nothing — a total outage presenting as a
		// series of individually plausible rejections.
		return nil, platformerrors.New("webhook catalog is empty")
	}

	d := &Dispatcher{
		store:    store,
		catalog:  catalog,
		clock:    clock.NewClock(),
		checkURL: webhooks.CheckEndpointURL,
		tracer:   tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:   logging.NewNamedLogger(logger, o11yName),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(d)
		}
	}

	mp := metrics.EnsureMetricsProvider(metricsProvider)

	var err error
	if d.dispatchedCounter, err = mp.NewInt64Counter("webhooks_deliveries_dispatched"); err != nil {
		return nil, platformerrors.Wrap(err, "creating deliveries dispatched counter")
	}

	// Named to match what webhooks.Dispatcher would have emitted, so a dashboard written
	// against the platform's instruments reads this application without knowing the write side
	// was replaced.
	if d.fanoutHist, err = mp.NewFloat64Histogram("webhooks_dispatch_fanout"); err != nil {
		return nil, platformerrors.Wrap(err, "creating dispatch fanout histogram")
	}

	return d, nil
}

// qualify renders the subscription string for one account's interest in one event type.
func qualify(accountID, eventType string) string {
	return accountID + accountSeparator + eventType
}

// unqualify recovers the event type from a subscription string, and reports whether it belonged
// to the given account. A subscription for another account returns false, which is what makes
// this safe to map over an endpoint's whole subscription set.
func unqualify(accountID, subscription string) (eventType string, ok bool) {
	return strings.CutPrefix(subscription, accountID+accountSeparator)
}

// Register stores an endpoint and its subscriptions, returning the signing secret.
//
// The secret is generated here and returned exactly once: it is stored to sign with, never read
// back out, and there is no endpoint on which a caller can ask for it again. RotateSecret is how
// a lost one is replaced.
//
// # Ordering
//
// Callers must commit the webhook row *before* calling this. A failure here then leaves a
// webhook that exists and does not yet deliver, and the caller reports the error. The reverse
// order fails in the direction that matters: an endpoint whose subscriptions are live but whose
// webhook row was rolled back would receive an account's events with nothing in the API that
// says it exists, and no way for the account to find or remove it.
func (d *Dispatcher) Register(ctx context.Context, registration *Registration) (secret string, err error) {
	if d == nil {
		return "", ErrNoDispatcher
	}

	ctx, span := d.tracer.StartSpan(ctx)
	defer span.End()

	if registration == nil {
		return "", platformerrors.ErrNilInputProvided
	}

	logger := d.logger.WithValue("webhook_id", registration.ID)

	events, err := d.qualifyAll(registration.AccountID, registration.EventTypes)
	if err != nil {
		return "", observability.PrepareAndLogError(err, logger, span, "qualifying webhook event types")
	}

	rawSecret := make([]byte, secretBytes)
	if _, err = rand.Read(rawSecret); err != nil {
		return "", observability.PrepareAndLogError(err, logger, span, "generating webhook signing secret")
	}

	endpoint := &webhooks.Endpoint{
		ID:          registration.ID,
		URL:         registration.URL,
		ContentType: registration.ContentType,
		Secret:      webhooks.Secret{Current: rawSecret},
		Events:      events,
	}
	endpoint.EnsureDefaults()

	// Vetted before it is stored, not only before it is dialed. A rejection reported at
	// registration reaches whoever submitted the URL; one discovered at delivery reaches a log
	// line, days later, after the dispatch has died.
	if err = d.checkURL(ctx, endpoint.URL); err != nil {
		return "", observability.PrepareAndLogError(err, logger, span, "validating webhook endpoint URL")
	}

	if err = d.store.SaveEndpoint(ctx, endpoint); err != nil {
		return "", observability.PrepareAndLogError(err, logger, span, "saving webhook endpoint")
	}

	return hex.EncodeToString(rawSecret), nil
}

// SetEventTypes replaces the account's subscriptions for one endpoint.
//
// Read-modify-write, because SaveEndpoint replaces an endpoint's whole subscription set and this
// application never owns all of it: the set stored against an endpoint is keyed by account, and
// rewriting it from the event types alone would drop any subscription belonging to a different
// account. In practice an endpoint has exactly one account — its webhook's — but relying on that
// here would make a future multi-account endpoint silently lose subscriptions.
func (d *Dispatcher) SetEventTypes(ctx context.Context, endpointID, accountID string, eventTypes []string) error {
	if d == nil {
		return ErrNoDispatcher
	}

	ctx, span := d.tracer.StartSpan(ctx)
	defer span.End()

	logger := d.logger.WithValue("webhook_id", endpointID)

	qualified, err := d.qualifyAll(accountID, eventTypes)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "qualifying webhook event types")
	}

	endpoint, err := d.store.GetEndpoint(ctx, endpointID)
	if err != nil {
		// Distinguished from every other read failure, because it is the one an operator can
		// act on: an unadopted webhook needs its secret rotated before it can be subscribed
		// to anything, and "no rows" alone does not say that.
		if errors.Is(err, sql.ErrNoRows) {
			return platformerrors.Wrapf(ErrEndpointNotRegistered, "webhook %q", endpointID)
		}

		return observability.PrepareAndLogError(err, logger, span, "reading webhook endpoint")
	}

	subscriptions := make([]string, 0, len(endpoint.Events)+len(qualified))
	for _, subscription := range endpoint.Events {
		if _, mine := unqualify(accountID, subscription); !mine {
			subscriptions = append(subscriptions, subscription)
		}
	}

	subscriptions = append(subscriptions, qualified...)
	endpoint.Events = subscriptions

	if err = d.store.SaveEndpoint(ctx, endpoint); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "saving webhook endpoint subscriptions")
	}

	return nil
}

// Archive retires an endpoint, stopping delivery without discarding its attempt history.
//
// Callers must call this *before* archiving the webhook row, which is the opposite of Register's
// ordering and for the same reason: each direction is ordered so the failure leaves the system
// delivering less than intended rather than more. A webhook row archived first, with this call
// then failing, is a subscriber that keeps receiving an account's events after the account
// removed it.
func (d *Dispatcher) Archive(ctx context.Context, endpointID string) error {
	if d == nil {
		return ErrNoDispatcher
	}

	ctx, span := d.tracer.StartSpan(ctx)
	defer span.End()

	if err := d.store.ArchiveEndpoint(ctx, endpointID); err != nil {
		return observability.PrepareAndLogError(err, d.logger.WithValue("webhook_id", endpointID), span, "archiving webhook endpoint")
	}

	return nil
}

// RotateSecret mints a new signing secret, demoting the current one to previous, and returns it.
//
// Both keys sign every delivery while Previous is set, so a subscriber accepts either signature
// for as long as it needs to switch. Clearing the old one is a second, later rotation — calling
// this twice retires the original key.
// registration supplies what an endpoint needs if it has to be created rather than rotated. It
// may be nil, in which case a missing endpoint is an error rather than something to repair.
func (d *Dispatcher) RotateSecret(ctx context.Context, endpointID string, registration *Registration) (secret string, err error) {
	if d == nil {
		return "", ErrNoDispatcher
	}

	ctx, span := d.tracer.StartSpan(ctx)
	defer span.End()

	logger := d.logger.WithValue("webhook_id", endpointID)

	endpoint, err := d.store.GetEndpoint(ctx, endpointID)
	if err != nil {
		// A webhook created before delivery existed has no endpoint, because the migration
		// that adopted these tables deliberately backfilled none: minting a signing secret in
		// SQL would have meant either requiring pgcrypto everywhere or using random(), and a
		// non-CSPRNG secret would start signing real deliveries the moment its owner
		// subscribed it to anything.
		//
		// Registering it here makes "adopt a webhook that predates delivery" and "replace a
		// secret I lost" the same operation, which is also the only operation whose response
		// can hand the owner a secret.
		if registration != nil && errors.Is(err, sql.ErrNoRows) {
			logger.Info("registering webhook endpoint on secret rotation")

			return d.Register(ctx, registration)
		}

		return "", observability.PrepareAndLogError(err, logger, span, "reading webhook endpoint")
	}

	rawSecret := make([]byte, secretBytes)
	if _, err = rand.Read(rawSecret); err != nil {
		return "", observability.PrepareAndLogError(err, logger, span, "generating webhook signing secret")
	}

	endpoint.Secret = webhooks.Secret{Current: rawSecret, Previous: endpoint.Secret.Current}

	if err = d.store.SaveEndpoint(ctx, endpoint); err != nil {
		return "", observability.PrepareAndLogError(err, logger, span, "saving rotated webhook signing secret")
	}

	logger.Info("webhook signing secret rotated")

	return hex.EncodeToString(rawSecret), nil
}

// Dispatch fans an event out to the account's subscribed endpoints, through the caller's
// executor.
//
// Taking the executor is the whole transactional guarantee: the dispatch rows are further
// statements in the transaction that wrote the row the event describes, so they commit with it
// or not at all. This is the same seam outbox.Enqueue uses, for the same reason.
//
// An event nobody subscribes to writes nothing and is not an error. That is the common case for
// most event types most of the time, and making it an error would have every caller branch on it.
func (d *Dispatcher) Dispatch(
	ctx context.Context,
	q database.SQLQueryExecutor,
	accountID,
	eventType,
	orderingKey string,
	payload json.RawMessage,
) error {
	// A nil Dispatcher is inert, so a process wired without webhooks still writes rows and
	// emits events. This is the opposite of Register's behavior on nil deliberately:
	// registering is a user asking for a webhook and must fail loudly, whereas dispatching is
	// a side effect of an unrelated write that must not fail it.
	if d == nil {
		return nil
	}

	ctx, span := d.tracer.StartSpan(ctx)
	defer span.End()

	if q == nil {
		return webhooks.ErrNilExecutor
	}

	if accountID == "" {
		// An event with no account belongs to no subscriber. Background jobs emit these.
		return nil
	}

	// An event type that is not subscribable is skipped, not rejected.
	//
	// Two things land here. Most are the deliberate exclusions — sign-ins, two-factor changes,
	// OAuth2 client lifecycle — which the application publishes and which no webhook may
	// receive. The rest would be an event type nothing publishes, which is a programming error.
	//
	// Neither may fail the caller. Dispatch runs inside the transaction that wrote the row the
	// event describes, so returning an error here does not fail a webhook: it fails the meal
	// plan. Registration is where a typo'd event type is caught, because that is where a human
	// types one; a constant that drifts out of the catalog is caught by the catalog's own test,
	// at build time rather than by taking down a write at runtime.
	if !d.catalog.Known(eventType) {
		return nil
	}

	logger := d.logger.WithValues(map[string]any{
		"account_id":            accountID,
		"webhooks.event_type":   eventType,
		"webhooks.ordering_key": orderingKey,
	})

	endpoints, err := d.store.EndpointsForEvent(ctx, q, qualify(accountID, eventType))
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "resolving webhook endpoints for event")
	}

	eventTypeAttr := metric.WithAttributes(attribute.String("webhooks.event_type", eventType))
	d.fanoutHist.Record(ctx, float64(len(endpoints)), eventTypeAttr)

	if len(endpoints) == 0 {
		return nil
	}

	endpointIDs := make([]string, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpointIDs = append(endpointIDs, endpoint.ID)
	}

	delivery := &webhooks.Delivery{
		// Generated here because Store.Enqueue does not: webhooks.Dispatcher mints this,
		// and bypassing it means owning the bits it did. Without one, every delivery shares
		// the empty string as its primary key — the second one in any transaction collides —
		// and the DeliveryIDHeader a subscriber deduplicates on carries nothing.
		ID:        identifiers.New(),
		EventType: eventType,
		// The ordering key is not qualified. It is compared only against other dispatches
		// for the same endpoint, and an endpoint belongs to one account, so the account adds
		// nothing but width to an index that is already on the claim path.
		OrderingKey: orderingKey,
		Payload:     payload,
	}

	if err = d.store.Enqueue(ctx, q, delivery, endpointIDs, d.clock.Now().UTC()); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "enqueuing webhook delivery")
	}

	// Counted once the statements succeed, but the transaction can still roll back after
	// this — so it counts intent to deliver rather than committed rows. The gap against
	// webhooks_deliveries_sent is the rollback rate.
	d.dispatchedCounter.Add(ctx, int64(len(endpointIDs)), eventTypeAttr)

	return nil
}

// qualifyAll vets event types against the catalog and renders them for storage.
//
// The catalog check happens here rather than being left to the platform, because the qualified
// string this produces can never be in a catalog: it holds unqualified types and would need one
// entry per account to hold these. Checking the unqualified type is checking the thing the
// catalog is actually about.
func (d *Dispatcher) qualifyAll(accountID string, eventTypes []string) ([]string, error) {
	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	// An empty set is allowed here, producing an endpoint that is registered and subscribed to
	// nothing. That is a real state — a webhook adopted from before delivery worked starts in
	// it, and unsubscribing from the last event type returns to it — and it is inert rather
	// than dangerous, because fan-out reads the subscriptions table. The rule that a webhook is
	// created subscribed to something is a request-validation rule, and lives there.
	qualified := make([]string, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		if !d.catalog.Known(eventType) {
			return nil, platformerrors.Wrapf(ErrUnknownEventType, "event type %q", eventType)
		}

		qualified = append(qualified, qualify(accountID, eventType))
	}

	return qualified, nil
}
