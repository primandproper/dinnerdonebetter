package fakes

import (
	"net/http"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/converters"

	"github.com/primandproper/platform-go/v5/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeWebhook builds a faked Webhook.
func BuildFakeWebhook() *types.Webhook {
	webhookID := BuildFakeID()
	cfg := BuildFakeWebhookTriggerConfig()
	cfg.BelongsToWebhook = webhookID

	return &types.Webhook{
		ID:               webhookID,
		Name:             fake.UUID(),
		ContentType:      "application/json",
		URL:              fake.URL(),
		Method:           http.MethodPost,
		TriggerConfigs:   []*types.WebhookTriggerConfig{cfg},
		CreatedAt:        BuildFakeTime(),
		ArchivedAt:       nil,
		BelongsToAccount: fake.UUID(),
		CreatedByUser:    fake.UUID(),
	}
}

// BuildFakeWebhooksList builds a faked WebhookList.
func BuildFakeWebhooksList() *filtering.QueryFilteredResult[types.Webhook] {
	var examples []*types.Webhook
	for range exampleQuantity {
		examples = append(examples, BuildFakeWebhook())
	}

	return &filtering.QueryFilteredResult[types.Webhook]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeWebhookTriggerConfig builds a faked WebhookTriggerConfig (join table).
func BuildFakeWebhookTriggerConfig() *types.WebhookTriggerConfig {
	return &types.WebhookTriggerConfig{
		ID:               BuildFakeID(),
		BelongsToWebhook: BuildFakeID(),
		TriggerEventID:   BuildFakeID(),
		CreatedAt:        BuildFakeTime(),
		ArchivedAt:       nil,
	}
}

// BuildFakeWebhookTriggerEvent builds a faked catalog WebhookTriggerEvent.
func BuildFakeWebhookTriggerEvent() *types.WebhookTriggerEvent {
	return &types.WebhookTriggerEvent{
		ID:            BuildFakeID(),
		Name:          fake.UUID(),
		Description:   fake.LoremIpsumSentence(5),
		CreatedAt:     BuildFakeTime(),
		LastUpdatedAt: nil,
		ArchivedAt:    nil,
	}
}

// BuildFakeWebhookTriggerEventList builds a faked WebhookList.
func BuildFakeWebhookTriggerEventList() *filtering.QueryFilteredResult[types.WebhookTriggerEvent] {
	var examples []*types.WebhookTriggerEvent
	for range exampleQuantity {
		examples = append(examples, BuildFakeWebhookTriggerEvent())
	}

	return &filtering.QueryFilteredResult[types.WebhookTriggerEvent]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

func BuildFakeWebhookTriggerEventCreationRequestInput() *types.WebhookTriggerEventCreationRequestInput {
	return &types.WebhookTriggerEventCreationRequestInput{
		Name:        fake.UUID(),
		Description: fake.LoremIpsumSentence(5),
	}
}

// BuildFakeWebhookCreationRequestInput builds a faked WebhookCreationRequestInput from a webhook.
func BuildFakeWebhookCreationRequestInput() *types.WebhookCreationRequestInput {
	webhook := BuildFakeWebhook()
	return converters.ConvertWebhookToWebhookCreationRequestInput(webhook)
}
