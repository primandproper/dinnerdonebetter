package converters

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v11/identifiers"
)

// ConvertWebhookToWebhookCreationRequestInput builds a WebhookCreationRequestInput from a Webhook.
func ConvertWebhookToWebhookCreationRequestInput(webhook *types.Webhook) *types.WebhookCreationRequestInput {
	return &types.WebhookCreationRequestInput{
		Name:        webhook.Name,
		ContentType: webhook.ContentType,
		URL:         webhook.URL,
		Method:      webhook.Method,
		Events:      webhook.EventTypes(),
	}
}

// ConvertWebhookToWebhookDatabaseCreationInput builds a WebhookDatabaseCreationInput from a Webhook.
func ConvertWebhookToWebhookDatabaseCreationInput(webhook *types.Webhook) *types.WebhookDatabaseCreationInput {
	configs := make([]*types.WebhookTriggerConfigDatabaseCreationInput, 0, len(webhook.TriggerConfigs))
	for _, cfg := range webhook.TriggerConfigs {
		configs = append(configs, ConvertWebhookTriggerConfigToWebhookTriggerConfigDatabaseCreationInput(cfg))
	}
	return &types.WebhookDatabaseCreationInput{
		ID:               webhook.ID,
		Name:             webhook.Name,
		ContentType:      webhook.ContentType,
		URL:              webhook.URL,
		Method:           webhook.Method,
		CreatedByUser:    webhook.CreatedByUser,
		BelongsToAccount: webhook.BelongsToAccount,
		TriggerConfigs:   configs,
	}
}

// ConvertWebhookTriggerConfigToWebhookTriggerConfigCreationRequestInput builds a WebhookTriggerConfigCreationRequestInput from a WebhookTriggerConfig.
func ConvertWebhookTriggerConfigToWebhookTriggerConfigCreationRequestInput(cfg *types.WebhookTriggerConfig) *types.WebhookTriggerConfigCreationRequestInput {
	return &types.WebhookTriggerConfigCreationRequestInput{
		BelongsToWebhook: cfg.BelongsToWebhook,
		EventType:        cfg.EventType,
	}
}

// ConvertWebhookTriggerConfigCreationRequestInputToWebhookTriggerConfigDatabaseCreationInput builds a WebhookTriggerConfigDatabaseCreationInput from a WebhookTriggerConfigCreationRequestInput.
func ConvertWebhookTriggerConfigCreationRequestInputToWebhookTriggerConfigDatabaseCreationInput(input *types.WebhookTriggerConfigCreationRequestInput) *types.WebhookTriggerConfigDatabaseCreationInput {
	return &types.WebhookTriggerConfigDatabaseCreationInput{
		ID:               identifiers.New(),
		BelongsToWebhook: input.BelongsToWebhook,
		EventType:        input.EventType,
	}
}

// ConvertWebhookTriggerConfigToWebhookTriggerConfigDatabaseCreationInput builds a WebhookTriggerConfigDatabaseCreationInput from a WebhookTriggerConfig.
func ConvertWebhookTriggerConfigToWebhookTriggerConfigDatabaseCreationInput(cfg *types.WebhookTriggerConfig) *types.WebhookTriggerConfigDatabaseCreationInput {
	return &types.WebhookTriggerConfigDatabaseCreationInput{
		ID:               cfg.ID,
		BelongsToWebhook: cfg.BelongsToWebhook,
		EventType:        cfg.EventType,
	}
}

// ConvertWebhookCreationRequestInputToWebhookDatabaseCreationInput creates a
// WebhookDatabaseCreationInput from a WebhookCreationRequestInput. CreatedByUser and
// BelongsToAccount are the caller's to set.
func ConvertWebhookCreationRequestInputToWebhookDatabaseCreationInput(input *types.WebhookCreationRequestInput) *types.WebhookDatabaseCreationInput {
	webhookID := identifiers.New()
	x := &types.WebhookDatabaseCreationInput{
		ID:             webhookID,
		Name:           input.Name,
		ContentType:    input.ContentType,
		URL:            input.URL,
		Method:         input.Method,
		TriggerConfigs: make([]*types.WebhookTriggerConfigDatabaseCreationInput, 0, len(input.Events)),
	}

	// Every event type becomes a subscription. There is no longer a "resolve or create the
	// catalog row" step, and so no way for an event to be silently dropped here for lacking
	// one: event types are validated against the generated catalog before this runs.
	for _, eventType := range input.Events {
		x.TriggerConfigs = append(x.TriggerConfigs, &types.WebhookTriggerConfigDatabaseCreationInput{
			ID:               identifiers.New(),
			BelongsToWebhook: webhookID,
			EventType:        eventType,
		})
	}

	return x
}
