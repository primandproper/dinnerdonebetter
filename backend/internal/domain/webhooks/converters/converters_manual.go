package converters

// The conversions in this file are hand-written: each does something the declaration format in
// cmd/tools/codegen/converters cannot express, and the note above each one says what. Everything
// else in this package is generated from those declarations into converters_generated.go.
//
// Adding a conversion here rather than declaring it is a decision, not a default. A conversion
// that is a field copy with a handful of exceptions belongs in the declaration, where the
// generator guarantees no destination field is silently forgotten.

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v11/identifiers"
)

// ConvertWebhookCreationRequestInputToWebhookDatabaseCreationInput creates a
// WebhookDatabaseCreationInput from a WebhookCreationRequestInput. CreatedByUser and
// BelongsToAccount are the caller's to set.
//
// Hand-written: it fans a flat list of event types out into one trigger config apiece, minting an
// ID for each and stamping the webhook's own — not a field copy.
func ConvertWebhookCreationRequestInputToWebhookDatabaseCreationInput(input *webhooks.WebhookCreationRequestInput) *webhooks.WebhookDatabaseCreationInput {
	webhookID := identifiers.New()
	x := &webhooks.WebhookDatabaseCreationInput{
		ID:             webhookID,
		Name:           input.Name,
		ContentType:    input.ContentType,
		URL:            input.URL,
		Method:         input.Method,
		TriggerConfigs: make([]*webhooks.WebhookTriggerConfigDatabaseCreationInput, 0, len(input.Events)),
	}

	// Every event type becomes a subscription. There is no longer a "resolve or create the
	// catalog row" step, and so no way for an event to be silently dropped here for lacking
	// one: event types are validated against the generated catalog before this runs.
	for _, eventType := range input.Events {
		x.TriggerConfigs = append(x.TriggerConfigs, &webhooks.WebhookTriggerConfigDatabaseCreationInput{
			ID:               identifiers.New(),
			BelongsToWebhook: webhookID,
			EventType:        eventType,
		})
	}

	return x
}
