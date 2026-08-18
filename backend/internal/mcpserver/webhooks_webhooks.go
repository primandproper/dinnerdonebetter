package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v11/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var webhookTriggerConfigSchema = map[string]any{
	"ID":               stringField("The ID of the trigger config"),
	"BelongsToWebhook": stringField("The ID of the webhook this config belongs to"),
	"EventType":        stringField("The catalog event type this webhook is subscribed to"),
	"CreatedAt":        timestampField("When the trigger config was created"),
	"ArchivedAt":       timestampField("When the trigger config was archived"),
}

var webhookSchema = map[string]any{
	"ID":               stringField("The ID of the webhook"),
	"Name":             stringField("The webhook name"),
	"URL":              stringField("The webhook URL"),
	"Method":           stringField("The HTTP method deliveries are made with; always POST"),
	"ContentType":      stringField("The content type; always application/json"),
	"BelongsToAccount": stringField("The ID of the account this webhook belongs to"),
	"CreatedByUser":    stringField("The ID of the user who created this webhook"),
	"TriggerConfigs":   arrayType(schemaObject(webhookTriggerConfigSchema)),
	"CreatedAt":        timestampField("When the webhook was created"),
	"LastUpdatedAt":    timestampField("When the webhook was last updated"),
	"ArchivedAt":       timestampField("When the webhook was archived"),
}

var webhookEventTypeSchema = map[string]any{
	"Type":        stringField("The event type, as it appears in a webhook subscription"),
	"Description": stringField("Prose explaining when the event fires"),
}

var getWebhookTool = &mcp.Tool{
	Name:        "GetWebhook",
	Description: "Get a webhook by its ID",
	InputSchema: schemaObject(map[string]any{
		"WebhookID": stringField("The ID of the webhook to get"),
	}),
	OutputSchema: schemaObject(webhookSchema),
}

type GetWebhookInvocation struct {
	WebhookID string `jsonschema:"description=The webhook ID"`
}

func (h *mcpToolManager) GetWebhook() mcp.ToolHandlerFor[*GetWebhookInvocation, *webhooks.Webhook] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetWebhookInvocation) (*mcp.CallToolResult, *webhooks.Webhook, error) {
		accountID, err := h.userFromRequest(req)
		if err != nil {
			return nil, nil, err
		}

		result, err := h.webhooksRepo.GetWebhook(ctx, x.WebhookID, accountID)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	}
}

var getWebhooksTool = &mcp.Tool{
	Name:        "GetWebhooks",
	Description: "Get webhooks with optional filtering",
	InputSchema: schemaObject(map[string]any{
		"Filter": filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(webhookSchema)),
	}),
}

type (
	GetWebhooksInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetWebhooksResult struct {
		Results []*webhooks.Webhook
	}
)

func (h *mcpToolManager) GetWebhooks() mcp.ToolHandlerFor[*GetWebhooksInvocation, *GetWebhooksResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetWebhooksInvocation) (*mcp.CallToolResult, *GetWebhooksResult, error) {
		accountID, err := h.userFromRequest(req)
		if err != nil {
			return nil, nil, err
		}

		results, err := h.webhooksRepo.GetWebhooks(ctx, accountID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetWebhooksResult{Results: results.Data}, nil
	}
}

var getWebhookEventTypesTool = &mcp.Tool{
	Name:        "GetWebhookEventTypes",
	Description: "Get the event types a webhook can subscribe to",
	InputSchema: schemaObject(map[string]any{}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(webhookEventTypeSchema)),
	}),
}

type (
	GetWebhookEventTypesInvocation struct{}

	GetWebhookEventTypesResult struct {
		Results []*webhooks.WebhookEventType
	}
)

// GetWebhookEventTypes reads the generated catalog rather than the database.
//
// It takes no filter because there is nothing to filter against: the catalog is Go, identical
// for every account, and constant for the lifetime of the deployment.
func (h *mcpToolManager) GetWebhookEventTypes() mcp.ToolHandlerFor[*GetWebhookEventTypesInvocation, *GetWebhookEventTypesResult] {
	return func(context.Context, *mcp.CallToolRequest, *GetWebhookEventTypesInvocation) (*mcp.CallToolResult, *GetWebhookEventTypesResult, error) {
		return nil, &GetWebhookEventTypesResult{Results: webhooks.EventTypeCatalog()}, nil
	}
}
