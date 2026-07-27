package converters

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v7/identifiers"
)

// ConvertWebhookToWebhookCreationRequestInput builds a WebhookCreationRequestInput from a Webhook.
func ConvertWebhookToWebhookCreationRequestInput(webhook *types.Webhook) *types.WebhookCreationRequestInput {
	events := make([]*types.WebhookTriggerEventCreationRequestInput, 0, len(webhook.TriggerConfigs))
	for _, cfg := range webhook.TriggerConfigs {
		events = append(events, &types.WebhookTriggerEventCreationRequestInput{ID: cfg.TriggerEventID})
	}
	return &types.WebhookCreationRequestInput{
		Name:        webhook.Name,
		ContentType: webhook.ContentType,
		URL:         webhook.URL,
		Method:      webhook.Method,
		Events:      events,
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
		TriggerEventID:   cfg.TriggerEventID,
	}
}

// ConvertWebhookTriggerConfigCreationRequestInputToWebhookTriggerConfigDatabaseCreationInput builds a WebhookTriggerConfigDatabaseCreationInput from a WebhookTriggerConfigCreationRequestInput.
func ConvertWebhookTriggerConfigCreationRequestInputToWebhookTriggerConfigDatabaseCreationInput(input *types.WebhookTriggerConfigCreationRequestInput) *types.WebhookTriggerConfigDatabaseCreationInput {
	return &types.WebhookTriggerConfigDatabaseCreationInput{
		ID:               identifiers.New(),
		BelongsToWebhook: input.BelongsToWebhook,
		TriggerEventID:   input.TriggerEventID,
	}
}

// ConvertWebhookTriggerConfigToWebhookTriggerConfigDatabaseCreationInput builds a WebhookTriggerConfigDatabaseCreationInput from a WebhookTriggerConfig.
func ConvertWebhookTriggerConfigToWebhookTriggerConfigDatabaseCreationInput(cfg *types.WebhookTriggerConfig) *types.WebhookTriggerConfigDatabaseCreationInput {
	return &types.WebhookTriggerConfigDatabaseCreationInput{
		ID:               cfg.ID,
		BelongsToWebhook: cfg.BelongsToWebhook,
		TriggerEventID:   cfg.TriggerEventID,
	}
}

// ConvertWebhookCreationRequestInputToWebhookDatabaseCreationInput creates a WebhookDatabaseCreationInput from a WebhookCreationRequestInput (without CreatedByUser; caller must set it).
// Only events with ID set are added to TriggerConfigs; create-new (Name/Description) must be resolved by the manager.
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
	for _, ev := range input.Events {
		if ev == nil || ev.ID == "" {
			continue
		}
		x.TriggerConfigs = append(x.TriggerConfigs, &types.WebhookTriggerConfigDatabaseCreationInput{
			ID:               identifiers.New(),
			BelongsToWebhook: webhookID,
			TriggerEventID:   ev.ID,
		})
	}
	return x
}
