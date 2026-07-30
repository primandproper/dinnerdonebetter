package main

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"

	"github.com/primandproper/platform-go/v8/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var webhookTriggerConfigSchema = map[string]any{
	"ID":               stringField("The ID of the trigger config"),
	"BelongsToWebhook": stringField("The ID of the webhook this config belongs to"),
	"TriggerEventID":   stringField("The ID of the trigger event"),
	"CreatedAt":        timestampField("When the trigger config was created"),
	"ArchivedAt":       timestampField("When the trigger config was archived"),
}

var webhookSchema = map[string]any{
	"ID":               stringField("The ID of the webhook"),
	"Name":             stringField("The webhook name"),
	"URL":              stringField("The webhook URL"),
	"Method":           stringField("The HTTP method (GET, POST, PUT, PATCH, DELETE)"),
	"ContentType":      stringField("The content type (application/json or application/xml)"),
	"BelongsToAccount": stringField("The ID of the account this webhook belongs to"),
	"CreatedByUser":    stringField("The ID of the user who created this webhook"),
	"TriggerConfigs":   arrayType(schemaObject(webhookTriggerConfigSchema)),
	"CreatedAt":        timestampField("When the webhook was created"),
	"LastUpdatedAt":    timestampField("When the webhook was last updated"),
	"ArchivedAt":       timestampField("When the webhook was archived"),
}

var webhookTriggerEventSchema = map[string]any{
	"ID":            stringField("The ID of the trigger event"),
	"Name":          stringField("The trigger event name"),
	"Description":   stringField("The trigger event description"),
	"CreatedAt":     timestampField("When the trigger event was created"),
	"LastUpdatedAt": timestampField("When the trigger event was last updated"),
	"ArchivedAt":    timestampField("When the trigger event was archived"),
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
		"Filter": queryFilterSchema(),
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

var getWebhookTriggerEventsTool = &mcp.Tool{
	Name:        "GetWebhookTriggerEvents",
	Description: "Get webhook trigger events (the catalog of available event types)",
	InputSchema: schemaObject(map[string]any{
		"Filter": queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(webhookTriggerEventSchema)),
	}),
}

type (
	GetWebhookTriggerEventsInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetWebhookTriggerEventsResult struct {
		Results []*webhooks.WebhookTriggerEvent
	}
)

func (h *mcpToolManager) GetWebhookTriggerEvents() mcp.ToolHandlerFor[*GetWebhookTriggerEventsInvocation, *GetWebhookTriggerEventsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetWebhookTriggerEventsInvocation) (*mcp.CallToolResult, *GetWebhookTriggerEventsResult, error) {
		results, err := h.webhooksRepo.GetWebhookTriggerEvents(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetWebhookTriggerEventsResult{Results: results.Data}, nil
	}
}
