package webhooks

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"

	platformerrors "github.com/primandproper/platform-go/v12/errors"
	"github.com/primandproper/platform-go/v12/tenancy"
	"github.com/primandproper/platform-go/v12/webhooks"
)

// The delivery side of a webhook: the endpoint it is delivered to, its signing secret, and the
// subscriptions fan-out reads.
//
// It is platform-go's, whole. The endpoint carries the account as its tenancy.Scope, so
// EndpointsForEvent resolves subscribers within one account and the event type stored against an
// endpoint is the unqualified catalog type — the same string that reaches a subscriber in
// X-Platform-Event. Registration and dispatch both go through webhooks.Dispatcher, which is what
// applies the catalog gate and the SSRF policy.
//
// What still goes to the Store directly is the pair of operations Dispatcher does not offer:
// replacing an endpoint's subscription set, and retiring it. Both are reads and writes of an
// endpoint this account already owns, resolved within its own scope.

// secretBytes is the size of a generated endpoint signing secret. 32 bytes matches the output
// width of the SHA-256 the signature is computed with; a longer key buys nothing against HMAC.
const secretBytes = 32

var (
	// ErrNoDispatcher indicates a webhook operation attempted in a repository built with no
	// delivery wiring. Registration fails on it rather than silently doing nothing, because a
	// user who asked for a webhook and was told it was created would otherwise never learn it
	// was not.
	ErrNoDispatcher = platformerrors.New("no webhook dispatcher configured")

	// ErrEndpointNotRegistered indicates an operation on a webhook that has no delivery
	// endpoint. That is the state a webhook whose row committed and whose registration then
	// failed is left in — see CreateWebhook for why the two are ordered that way. Rotating the
	// secret registers it.
	ErrEndpointNotRegistered = platformerrors.New("webhook has no delivery endpoint; rotate its signing secret to register one")
)

// registerEndpoint registers a webhook as a delivery endpoint, returning the signing secret.
//
// The secret is generated here and returned exactly once: it is stored to sign with, never read
// back out, and there is no endpoint on which a caller can ask for it again. RotateWebhookSecret
// is how a lost one is replaced.
func (r *repository) registerEndpoint(ctx context.Context, webhook *types.Webhook) (string, error) {
	if r.dispatcher == nil {
		return "", ErrNoDispatcher
	}

	secret, err := newSigningSecret()
	if err != nil {
		return "", err
	}

	if err = r.dispatcher.Register(ctx, &webhooks.Endpoint{
		// The account, as the endpoint's tenant. tenancy.Of refuses to name nobody, so a
		// webhook that lost its account fails to register rather than registering into the
		// global scope and receiving every account's events.
		Scope: tenancy.Of(webhook.BelongsToAccount),
		// Deliberately the webhook's own ID rather than a separate identifier. One row per
		// webhook in each of two tables, joined by nothing but a shared key, is what keeps
		// "the endpoint for this webhook" from being a lookup that can return the wrong
		// answer.
		ID:          webhook.ID,
		URL:         webhook.URL,
		ContentType: webhook.ContentType,
		Secret:      webhooks.Secret{Current: secret},
		Events:      subscriptions(webhook),
	}); err != nil {
		return "", err
	}

	return hex.EncodeToString(secret), nil
}

// rotateEndpointSecret mints a new signing secret, demoting the current one to previous.
//
// Both keys sign every delivery while Previous is set, so a subscriber accepts either signature
// for as long as it needs to switch. Clearing the old one is a second, later rotation — calling
// this twice retires the original key.
//
// A webhook with no endpoint is registered rather than refused, which makes "adopt a webhook
// whose registration failed" and "replace a secret I lost" the same operation. It is also the
// only operation whose response can hand the owner a secret.
func (r *repository) rotateEndpointSecret(ctx context.Context, webhook *types.Webhook) (string, error) {
	if r.dispatcher == nil {
		return "", ErrNoDispatcher
	}

	endpoint, err := r.endpoints.GetEndpoint(ctx, tenancy.Of(webhook.BelongsToAccount), webhook.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return r.registerEndpoint(ctx, webhook)
		}

		return "", err
	}

	secret, err := newSigningSecret()
	if err != nil {
		return "", err
	}

	endpoint.Secret = webhooks.Secret{Current: secret, Previous: endpoint.Secret.Current}

	if err = r.endpoints.SaveEndpoint(ctx, endpoint); err != nil {
		return "", err
	}

	return hex.EncodeToString(secret), nil
}

// setSubscriptions replaces an endpoint's subscription set with the webhook's live trigger
// configs.
//
// It is a read-modify-write rather than a fresh Endpoint because SaveEndpoint replaces the whole
// row, and the URL and signing secret it would otherwise overwrite are not this caller's to
// supply. The event types are replaced outright: the endpoint is scoped to one account, so every
// subscription on it belongs to the webhook being synced.
//
// It goes to the Store rather than through Dispatcher.Register because an endpoint subscribed to
// nothing is a state this application reaches — unsubscribing from the last event type — and
// Register refuses it, reasonably, as a registration that cannot have been meant. Here it is not
// a registration, and it is inert rather than dangerous: fan-out reads the subscriptions table.
func (r *repository) setSubscriptions(ctx context.Context, webhook *types.Webhook) error {
	if r.dispatcher == nil {
		return ErrNoDispatcher
	}

	scope := tenancy.Of(webhook.BelongsToAccount)

	endpoint, err := r.endpoints.GetEndpoint(ctx, scope, webhook.ID)
	if err != nil {
		// Distinguished from every other read failure, because it is the one an operator can
		// act on: an unregistered webhook needs its secret rotated before it can be
		// subscribed to anything, and "no rows" alone does not say that.
		if errors.Is(err, sql.ErrNoRows) {
			return platformerrors.Wrapf(ErrEndpointNotRegistered, "webhook %q", webhook.ID)
		}

		return err
	}

	endpoint.Events = subscriptions(webhook)

	return r.endpoints.SaveEndpoint(ctx, endpoint)
}

// archiveEndpoint retires an endpoint, stopping delivery without discarding its attempt history.
func (r *repository) archiveEndpoint(ctx context.Context, webhookID, accountID string) error {
	if r.dispatcher == nil {
		return ErrNoDispatcher
	}

	return r.endpoints.ArchiveEndpoint(ctx, tenancy.Of(accountID), webhookID)
}

// subscriptions renders a webhook's live trigger configs as catalog event types.
//
// They are unqualified — the account is on the endpoint, not smuggled into the event type — so
// what is stored here is the same string the catalog holds and the same one a subscriber reads
// out of X-Platform-Event.
func subscriptions(webhook *types.Webhook) []webhooks.EventType {
	eventTypes := webhook.EventTypes()

	events := make([]webhooks.EventType, 0, len(eventTypes))
	for _, eventType := range eventTypes {
		events = append(events, webhooks.EventType(eventType))
	}

	return events
}

// newSigningSecret mints the bytes an endpoint's deliveries are signed with.
func newSigningSecret() ([]byte, error) {
	secret := make([]byte, secretBytes)
	if _, err := rand.Read(secret); err != nil {
		return nil, platformerrors.Wrap(err, "generating webhook signing secret")
	}

	return secret, nil
}
