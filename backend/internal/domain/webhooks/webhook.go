package webhooks

import (
	"context"
	"net/http"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"

	"github.com/primandproper/platform-go/v11/encoding"
	"github.com/primandproper/platform-go/v11/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const (
	// WebhookCreatedServiceEventType indicates a webhook was created.
	WebhookCreatedServiceEventType = "webhook_created"
	// WebhookArchivedServiceEventType indicates a webhook was archived.
	WebhookArchivedServiceEventType = "webhook_archived"
	// WebhookTriggerConfigCreatedServiceEventType indicates a webhook trigger config was created.
	WebhookTriggerConfigCreatedServiceEventType = "webhook_trigger_config_created"
	// WebhookTriggerConfigArchivedServiceEventType indicates a webhook trigger config was archived.
	WebhookTriggerConfigArchivedServiceEventType = "webhook_trigger_config_archived"

	// DeliveryMethod is the only HTTP method a webhook is delivered with.
	//
	// The delivery worker POSTs, always. A webhook is a message being handed to a subscriber,
	// and the other methods this model used to accept described requests nobody was making:
	// a GET carries no body, so there is nothing to sign or to receive.
	DeliveryMethod = http.MethodPost
)

type (
	// Webhook represents a webhook listener, an endpoint to send an HTTP request to upon an event.
	//
	// The signing secret is deliberately absent. It is returned once, from the call that creates
	// or rotates it, and never read back: a secret an API will hand out on request is one an
	// attacker with read access can hand out to themselves.
	//
	// # Divergence from webhooks.Endpoint
	//
	// The platform speaks of an Endpoint, which is deliberately general: it is a delivery target
	// — a URL, a content type, a signing secret, and Events, a flat list of subscription strings
	// — and it says nothing about whose it is, because tenancy depth is an application's
	// decision. Ours is an account's webhook, and this type keeps that.
	//
	// Two fields are why both types exist, and neither survives a translation to Endpoint:
	//
	//   - BelongsToAccount is the filter on every read and write of this resource. Endpoint has
	//     no owner to filter on, so the account travels inside the subscription string instead;
	//     see webhookdispatch.qualify.
	//   - TriggerConfigs are identified, individually archivable rows, and the API creates and
	//     archives them one at a time. Endpoint.Events is a bare []string: it can express the
	//     set, not the identity or the archival timestamp of any member of it.
	//
	// Deleting this type in favor of Endpoint would therefore not be an internal cleanup — it
	// would drop the fields the permission model and the API surface are built on, and rename
	// the rest, because the platform's JSON tags are not these. webhookdispatch/conversion.go
	// is where the translation lives, and it is the only place it happens.
	Webhook struct {
		_ struct{} `json:"-"`

		CreatedAt        time.Time               `json:"createdAt"`
		ArchivedAt       *time.Time              `json:"archivedAt"`
		LastUpdatedAt    *time.Time              `json:"lastUpdatedAt"`
		Name             string                  `json:"name"`
		URL              string                  `json:"url"`
		Method           string                  `json:"method"`
		ID               string                  `json:"id"`
		BelongsToAccount string                  `json:"belongsToAccount"`
		CreatedByUser    string                  `json:"createdByUser"`
		ContentType      string                  `json:"contentType"`
		TriggerConfigs   []*WebhookTriggerConfig `json:"triggerConfigs"`
	}

	// WebhookTriggerConfig represents a webhook's subscription to one event type.
	WebhookTriggerConfig struct {
		_ struct{} `json:"-"`

		CreatedAt        time.Time  `json:"createdAt"`
		ArchivedAt       *time.Time `json:"archivedAt"`
		ID               string     `json:"id"`
		BelongsToWebhook string     `json:"belongsToWebhook"`
		// EventType is a catalog event type — one of the strings the application publishes.
		// It was a foreign key into a webhook_trigger_events table whose IDs were random,
		// which is why no webhook ever matched an event: the fan-out compared those IDs
		// against event type strings. The catalog is now generated Go, and this holds the
		// event type itself.
		EventType string `json:"eventType"`
	}

	// WebhookCreationRequestInput represents what a User could set as input for creating a webhook.
	WebhookCreationRequestInput struct {
		_ struct{} `json:"-"`

		Name        string `json:"name"`
		ContentType string `json:"contentType"`
		URL         string `json:"url"`
		Method      string `json:"method"`
		// Events are catalog event types this webhook subscribes to.
		Events []string `json:"events"`
	}

	// WebhookCreationResponse pairs a created webhook with its signing secret.
	//
	// The secret is separate from Webhook rather than a field on it precisely so that it
	// cannot be returned by accident: every read path deals in Webhook, and no read path can
	// populate a field that type does not have.
	WebhookCreationResponse struct {
		_ struct{} `json:"-"`

		Webhook *Webhook `json:"webhook"`
		// Secret is the hex-encoded HMAC signing key, shown exactly once. See
		// webhooks.Verify for what a subscriber does with it.
		Secret string `json:"secret"`
	}

	// WebhookDatabaseCreationInput is used for creating a webhook.
	WebhookDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID               string                                       `json:"-"`
		Name             string                                       `json:"-"`
		ContentType      string                                       `json:"-"`
		URL              string                                       `json:"-"`
		Method           string                                       `json:"-"`
		CreatedByUser    string                                       `json:"-"`
		BelongsToAccount string                                       `json:"-"`
		TriggerConfigs   []*WebhookTriggerConfigDatabaseCreationInput `json:"-"`
	}

	// WebhookTriggerConfigCreationRequestInput represents what a User could set as input for adding a trigger config.
	WebhookTriggerConfigCreationRequestInput struct {
		_ struct{} `json:"-"`

		BelongsToWebhook string `json:"belongsToWebhook"`
		EventType        string `json:"eventType"`
	}

	// WebhookTriggerConfigDatabaseCreationInput is used for creating a webhook trigger config.
	WebhookTriggerConfigDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID               string `json:"-"`
		BelongsToWebhook string `json:"-"`
		EventType        string `json:"-"`
	}

	// WebhookEventType is one subscribable event type, as an API surface renders it.
	//
	// It replaces the webhook_trigger_events table, whose rows carried random IDs that the
	// fan-out then compared against event type strings — which is why nothing ever matched. The
	// event type is now its own identity, and the set of them is generated from the constants
	// the domains declare rather than stored.
	WebhookEventType struct {
		_ struct{} `json:"-"`

		// Type is the event type itself, and what a trigger config stores.
		Type string `json:"type"`
		// Description is prose explaining when the event fires.
		Description string `json:"description"`
	}

	// WebhookDataManager describes a structure capable of storing and retrieving webhooks.
	WebhookDataManager interface {
		WebhookExists(ctx context.Context, webhookID, accountID string) (bool, error)
		GetWebhook(ctx context.Context, webhookID, accountID string) (*Webhook, error)
		GetWebhooks(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Webhook], error)
		CreateWebhook(ctx context.Context, input *WebhookDatabaseCreationInput) (*WebhookCreationResponse, error)
		ArchiveWebhook(ctx context.Context, webhookID, accountID string) error
		AddWebhookTriggerConfig(ctx context.Context, accountID string, input *WebhookTriggerConfigDatabaseCreationInput) (*WebhookTriggerConfig, error)
		ArchiveWebhookTriggerConfig(ctx context.Context, webhookID, accountID, configID string) error
		RotateWebhookSecret(ctx context.Context, webhookID, accountID string) (string, error)
	}
)

// validEventType rejects an event type the application does not publish.
//
// The catalog is generated from the domains' own constants, so this is the same set the
// dispatcher gates on. Checking it at the API boundary is what turns a typo into a 400 with the
// offending value in it, rather than a webhook that is accepted, stored, and never fires.
var validEventType = validation.By(func(value any) error {
	eventType, ok := value.(string)
	if !ok || !catalog.Known(eventType) {
		return validation.NewError("validation_unknown_event_type", "must be a known webhook event type")
	}

	return nil
})

// validContentType and validMethod pin the two fields the delivery worker no longer varies.
//
// Both used to be per-webhook and neither is honored any more: the worker POSTs, and a delivery
// carries one payload shared by every subscriber, so a per-endpoint XML rendering would mean
// dispatching the same event twice. Rejecting the other values is better than accepting and
// ignoring them — a webhook configured for XML that silently receives JSON is a subscriber
// parsing failure with nothing pointing at the cause.
var (
	validContentType = validation.In(encoding.ContentTypeJSON.String())
	validMethod      = validation.In(DeliveryMethod)
)

var _ validation.ValidatableWithContext = (*WebhookCreationRequestInput)(nil)

// ValidateWithContext validates a WebhookCreationRequestInput.
func (w *WebhookCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, w,
		validation.Field(&w.Name, validation.Required),
		validation.Field(&w.URL, validation.Required, is.URL),
		validation.Field(&w.Method, validation.Required, validMethod),
		validation.Field(&w.ContentType, validation.Required, validContentType),
		validation.Field(&w.Events, validation.Required, validation.Length(1, 100), validation.Each(validEventType)),
	)
}

var _ validation.ValidatableWithContext = (*WebhookTriggerConfigCreationRequestInput)(nil)

// ValidateWithContext validates a WebhookTriggerConfigCreationRequestInput.
func (w *WebhookTriggerConfigCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, w,
		validation.Field(&w.BelongsToWebhook, validation.Required),
		validation.Field(&w.EventType, validation.Required, validEventType),
	)
}

var _ validation.ValidatableWithContext = (*WebhookDatabaseCreationInput)(nil)

// ValidateWithContext validates a WebhookDatabaseCreationInput.
func (w *WebhookDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, w,
		validation.Field(&w.ID, validation.Required),
		validation.Field(&w.Name, validation.Required),
		validation.Field(&w.URL, validation.Required, is.URL),
		validation.Field(&w.Method, validation.Required, validMethod),
		validation.Field(&w.ContentType, validation.Required, validContentType),
		validation.Field(&w.TriggerConfigs, validation.Required),
		validation.Field(&w.BelongsToAccount, validation.Required),
		validation.Field(&w.CreatedByUser, validation.Required),
	)
}

// EventTypeCatalog returns every subscribable event type, sorted by type.
//
// It reads the generated catalog rather than a table, which is what makes "the events this
// application publishes" and "the events a webhook may subscribe to" the same list by
// construction instead of by an admin remembering to keep two of them aligned.
func EventTypeCatalog() []*WebhookEventType {
	known := catalog.Catalog()

	eventTypes := make([]*WebhookEventType, 0, len(known))
	for _, eventType := range known.EventTypes() {
		eventTypes = append(eventTypes, &WebhookEventType{
			Type:        eventType.String(),
			Description: known[eventType].Description,
		})
	}

	return eventTypes
}

// EventTypes returns the event types a webhook is subscribed to, in trigger config order.
//
// It is the shape the dispatcher wants — subscriptions are replaced as a set, never edited in
// place — and having one place render it keeps the two callers that need it from disagreeing
// about whether archived configs count. They do not.
func (w *Webhook) EventTypes() []string {
	if w == nil {
		return nil
	}

	eventTypes := make([]string, 0, len(w.TriggerConfigs))
	for _, cfg := range w.TriggerConfigs {
		if cfg == nil || cfg.ArchivedAt != nil {
			continue
		}

		eventTypes = append(eventTypes, cfg.EventType)
	}

	return eventTypes
}
