/*
Package webhooks is this application's half of platform-go's webhooks: the data
change events a webhook write emits, and the event catalog rendered for a
surface that lists it.

The model is platform's — an endpoint, and the subscriptions under it — and so
are the store, the dispatcher and the gRPC surface; see
internal/repositories/postgres/webhooksstore and internal/build/webhooks. What
is not platform's is which events exist, which is generated into catalog from
the constants every domain declares.
*/
package webhooks

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"
)

// The descriptions below are subscriber-facing: `make webhook_catalog` renders
// each doc comment into the catalog a client reads to decide what to subscribe
// to, so they say what happened and nothing about why the constant exists.
const (
	// WebhookCreatedServiceEventType indicates a webhook was created.
	WebhookCreatedServiceEventType = "webhook_created"
	// WebhookUpdatedServiceEventType indicates a webhook's name, URL or content type changed.
	WebhookUpdatedServiceEventType = "webhook_updated"
	// WebhookArchivedServiceEventType indicates a webhook was archived.
	WebhookArchivedServiceEventType = "webhook_archived"
	// WebhookSecretRotatedServiceEventType indicates a webhook's signing secret changed, and carries no secret.
	WebhookSecretRotatedServiceEventType = "webhook_secret_rotated"
	// WebhookTriggerConfigCreatedServiceEventType indicates a webhook trigger config was created.
	WebhookTriggerConfigCreatedServiceEventType = "webhook_trigger_config_created"
	// WebhookTriggerConfigArchivedServiceEventType indicates a webhook trigger config was archived.
	WebhookTriggerConfigArchivedServiceEventType = "webhook_trigger_config_archived"
)

// WebhookEventType is one subscribable event type, as the MCP server renders it.
type WebhookEventType struct {
	_ struct{} `json:"-"`

	// Type is the event type itself, and what a subscription stores.
	Type string `json:"type"`
	// Description is prose explaining when the event fires.
	Description string `json:"description"`
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
