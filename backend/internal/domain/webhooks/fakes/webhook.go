package fakes

import (
	"net/http"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/converters"

	"github.com/primandproper/platform-go/v10/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeWebhookEventType returns an event type from the real catalog.
//
// Not a random string: an event type outside the catalog is rejected at every boundary that
// takes one, so a fake that invented them would only ever exercise the rejection path.
func BuildFakeWebhookEventType() string {
	eventTypes := catalog.Catalog().EventTypes()

	return eventTypes[fake.Number(0, len(eventTypes)-1)]
}

// BuildFakeWebhook builds a faked Webhook.
func BuildFakeWebhook() *types.Webhook {
	webhookID := BuildFakeID()
	cfg := BuildFakeWebhookTriggerConfig()
	cfg.BelongsToWebhook = webhookID

	return &types.Webhook{
		ID:          webhookID,
		Name:        fake.UUID(),
		ContentType: "application/json",
		// A literal address from RFC 5737's TEST-NET-1, the range IANA reserves for exactly
		// this. Registration runs webhooks.CheckEndpointURL, which requires https and refuses
		// any host that is not globally routable — so a random fake domain is rejected for
		// failing to resolve, and a resolvable one would make every test that creates a
		// webhook depend on DNS. A literal address needs no lookup, and this one is nobody's.
		URL:              "https://192.0.2.1/webhook",
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

// BuildFakeWebhookTriggerConfig builds a faked WebhookTriggerConfig.
func BuildFakeWebhookTriggerConfig() *types.WebhookTriggerConfig {
	return &types.WebhookTriggerConfig{
		ID:               BuildFakeID(),
		BelongsToWebhook: BuildFakeID(),
		EventType:        BuildFakeWebhookEventType(),
		CreatedAt:        BuildFakeTime(),
		ArchivedAt:       nil,
	}
}

// BuildFakeWebhookCreationRequestInput builds a faked WebhookCreationRequestInput from a webhook.
func BuildFakeWebhookCreationRequestInput() *types.WebhookCreationRequestInput {
	webhook := BuildFakeWebhook()
	return converters.ConvertWebhookToWebhookCreationRequestInput(webhook)
}
