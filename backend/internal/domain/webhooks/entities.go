package webhooks

import "github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

// Entities declares this domain's entities for code generation.
var Entities = entitydecl.Domain{
	Entities: []entitydecl.Entity{
		{
			Type: Webhook{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "Name", Expr: `fake.UUID()`},
					{
						Name: "ContentType",
						Expr: `"application/json"`,
						Why:  "Validation pins this to encoding.ContentTypeJSON; the delivery worker no longer varies it, so anything else is rejected at registration.",
					},
					{
						Name: "URL",
						Expr: `"https://192.0.2.1/webhook"`,
						Why:  "A literal address from RFC 5737's TEST-NET-1, the range IANA reserves for exactly this. Registration runs webhooks.CheckEndpointURL, which requires https and refuses any host that is not globally routable — so a random fake domain is rejected for failing to resolve, and a resolvable one would make every test that creates a webhook depend on DNS. A literal address needs no lookup, and this one is nobody's.",
					},
					{
						Name: "Method",
						Expr: `types.DeliveryMethod`,
						Why:  "The only method a webhook is delivered with, and the only one validation accepts. Named rather than spelled so the fake follows the constant if it ever moves.",
					},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: WebhookCreationRequestInput{}, Converter: "ConvertWebhookToWebhookCreationRequestInput"},
				},
			},
		},
		{
			Type: WebhookTriggerConfig{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{
						Name: "EventType",
						Expr: `BuildFakeWebhookEventType()`,
						Why:  "An event type from the real catalog. An event type outside it is rejected at every boundary that takes one, so a fake that invented them would only ever exercise the rejection path.",
					},
				},
			},
		},
	},
}
