package converters

// The conversions in this file are hand-written: each does something the generator in
// cmd/tools/codegen/converters does not produce — it fails, it fans one value out into many, it
// defaults something, it needs a second entity to make sense of the first. exceptions.go names
// each one and says why.
//
// Everything else in this package is generated. A conversion that is a field copy with a handful
// of exceptions belongs there, where no destination field can be silently forgotten.

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v11/identifiers"
)

// ConvertWebhookCreationRequestInputToWebhookDatabaseCreationInput creates a
// WebhookDatabaseCreationInput from a WebhookCreationRequestInput. CreatedByUser and
// BelongsToAccount are the caller's to set.
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
