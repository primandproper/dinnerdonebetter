package fakes

import (
	"net/http"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/converters"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeWebhookEventType returns an event type from the real catalog.
//
// Not a random string: an event type outside the catalog is rejected at every boundary that
// takes one, so a fake that invented them would only ever exercise the rejection path.
func BuildFakeWebhookEventType() string {
	eventTypes := catalog.Catalog().EventTypes()

	return eventTypes[gofakeit.Number(0, len(eventTypes)-1)].String()
}

// BuildFakeWebhook builds a faked Webhook.
func BuildFakeWebhook() *types.Webhook {
	webhook := fake.BuildFakeRecord[types.Webhook]()

	webhook.ContentType = "application/json"
	webhook.Method = http.MethodPost

	// A literal address from RFC 5737's TEST-NET-1, the range IANA reserves for exactly
	// this. Registration runs webhooks.CheckEndpointURL, which requires https and refuses
	// any host that is not globally routable — so a random fake domain is rejected for
	// failing to resolve, and a resolvable one would make every test that creates a
	// webhook depend on DNS. A literal address needs no lookup, and this one is nobody's.
	webhook.URL = "https://192.0.2.1/webhook"

	cfg := BuildFakeWebhookTriggerConfig()
	cfg.BelongsToWebhook = webhook.ID
	webhook.TriggerConfigs = []*types.WebhookTriggerConfig{cfg}

	return webhook
}

// BuildFakeWebhooksList builds a faked WebhookList.
func BuildFakeWebhooksList() *filtering.QueryFilteredResult[types.Webhook] {
	return fake.BuildFakePage(BuildFakeWebhook)
}

// BuildFakeWebhookTriggerConfig builds a faked WebhookTriggerConfig.
func BuildFakeWebhookTriggerConfig() *types.WebhookTriggerConfig {
	cfg := fake.BuildFakeRecord[types.WebhookTriggerConfig]()
	cfg.EventType = BuildFakeWebhookEventType()

	return cfg
}

// BuildFakeWebhookCreationRequestInput builds a faked WebhookCreationRequestInput from a webhook.
func BuildFakeWebhookCreationRequestInput() *types.WebhookCreationRequestInput {
	webhook := BuildFakeWebhook()

	return converters.ConvertWebhookToWebhookCreationRequestInput(webhook)
}
