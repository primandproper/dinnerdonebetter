package main

// Conversions declared for the webhooks domain.
//
// A conversion with no Fields is a whole-struct copy: every field of the destination is
// filled from the field of the same name, and the generator fails rather than leave one
// empty. Fields carries a rule per destination field where that is not what happens, and
// the reason it carries is rendered into the generated source. See declaration.go for the
// rules, and converters_manual.go in the domain for the conversions this cannot express.

func init() {
	register("webhooks", []*Conversion{
		{Name: "ConvertWebhookToWebhookCreationRequestInput", From: Param{Name: "webhook", Type: "Webhook"}, To: "WebhookCreationRequestInput",
			Fields: map[string]Rule{
				"Events": Expr("webhook.EventTypes()", "A request lists event types; a webhook stores one trigger config per subscription. EventTypes flattens the configs back to the strings they subscribe to."),
			},
		},
		{Name: "ConvertWebhookToWebhookDatabaseCreationInput", From: Param{Name: "webhook", Type: "Webhook"}, To: "WebhookDatabaseCreationInput",
			Fields: map[string]Rule{
				"TriggerConfigs": MapSlice("ConvertWebhookTriggerConfigToWebhookTriggerConfigDatabaseCreationInput", FromField("TriggerConfigs")),
			},
		},
		{Name: "ConvertWebhookTriggerConfigToWebhookTriggerConfigCreationRequestInput", From: Param{Name: "cfg", Type: "WebhookTriggerConfig"}, To: "WebhookTriggerConfigCreationRequestInput"},
		{Name: "ConvertWebhookTriggerConfigCreationRequestInputToWebhookTriggerConfigDatabaseCreationInput", From: Param{Name: "input", Type: "WebhookTriggerConfigCreationRequestInput"}, To: "WebhookTriggerConfigDatabaseCreationInput",
			Fields: map[string]Rule{
				"ID": NewID(),
			},
		},
		{Name: "ConvertWebhookTriggerConfigToWebhookTriggerConfigDatabaseCreationInput", From: Param{Name: "cfg", Type: "WebhookTriggerConfig"}, To: "WebhookTriggerConfigDatabaseCreationInput"},
	})
}
