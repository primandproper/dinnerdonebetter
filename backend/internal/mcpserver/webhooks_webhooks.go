package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"

	platformwebhooks "github.com/primandproper/platform-go/v14/webhooks"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// The webhook tools read platform's endpoints now rather than this repository's own
// webhooks table, which the store migration dropped.
//
// The shape a caller sees changed with it, and honestly so: a webhook was one row with a
// list of trigger configs hanging off it, and an endpoint is a row with a list of
// subscriptions. The fields that went — Method and CreatedByUser — went because platform
// has nowhere to put them: every delivery is a POST, and who registered an endpoint is the
// scope it was registered in rather than a person.
//
// The signing secret is absent for a better reason. It is not that platform has nowhere to
// put it; it is that a secret readable over a tool call is a secret anybody who can reach
// the tool can forge deliveries with.
var webhookSubscriptionSchema = map[string]any{
	"ID":                stringField("The ID of the subscription"),
	"BelongsToEndpoint": stringField("The ID of the endpoint this subscription belongs to"),
	"EventType":         stringField("The catalog event type this endpoint is subscribed to"),
	fieldCreatedAt:      timestampField("When the subscription was created"),
	fieldArchivedAt:     timestampField("When the subscription was archived"),
}

var webhookEndpointSchema = map[string]any{
	"ID":               stringField("The ID of the endpoint"),
	fieldName:          stringField("The endpoint name"),
	"URL":              stringField("The endpoint URL"),
	"ContentType":      stringField("The content type; always application/json"),
	"Disabled":         boolField("Whether deliveries to this endpoint are suspended"),
	"Subscriptions":    arrayType(schemaObject(webhookSubscriptionSchema)),
	fieldCreatedAt:     timestampField("When the endpoint was created"),
	fieldLastUpdatedAt: timestampField("When the endpoint was last updated"),
	fieldArchivedAt:    timestampField("When the endpoint was archived"),
}

var webhookEventTypeSchema = map[string]any{
	"Type":           stringField("The event type, as it appears in a webhook subscription"),
	fieldDescription: stringField("Prose explaining when the event fires"),
}

var getWebhookTool = &mcp.Tool{
	Name:        "GetWebhookEndpoint",
	Description: "Get a webhook endpoint by its ID",
	InputSchema: schemaObject(map[string]any{
		"EndpointID": stringField("The ID of the endpoint to get"),
	}),
	OutputSchema: schemaObject(webhookEndpointSchema),
}

type GetWebhookInvocation struct {
	EndpointID string `jsonschema:"description=The endpoint ID"`
}

func (h *mcpToolManager) GetWebhook() mcp.ToolHandlerFor[*GetWebhookInvocation, *platformwebhooks.Endpoint] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetWebhookInvocation) (*mcp.CallToolResult, *platformwebhooks.Endpoint, error) {
		accountID, err := h.userFromRequest(req)
		if err != nil {
			return nil, nil, err
		}

		// Scoped to the caller's account, which is what keeps one account's endpoints
		// out of another's reads. See internal/build/webhooks, where the gRPC surface
		// is handed the account-scoped principal for the same reason.
		result, err := h.webhooks.GetEndpoint(ctx, h.reader, tenancy.Of(accountID), x.EndpointID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

var getWebhooksTool = &mcp.Tool{
	Name:        "GetWebhookEndpoints",
	Description: "Get webhook endpoints with optional filtering",
	InputSchema: schemaObject(map[string]any{
		fieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		fieldResults: arrayType(schemaObject(webhookEndpointSchema)),
	}),
}

type (
	GetWebhooksInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetWebhooksResult struct {
		Results []*platformwebhooks.Endpoint
	}
)

func (h *mcpToolManager) GetWebhooks() mcp.ToolHandlerFor[*GetWebhooksInvocation, *GetWebhooksResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetWebhooksInvocation) (*mcp.CallToolResult, *GetWebhooksResult, error) {
		accountID, err := h.userFromRequest(req)
		if err != nil {
			return nil, nil, err
		}

		results, err := h.webhooks.ListEndpoints(ctx, h.reader, tenancy.Of(accountID), x.Filter)
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
		fieldResults: arrayType(schemaObject(webhookEventTypeSchema)),
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
// for every account, and constant for the lifetime of the deployment. It survives the store
// migration for that reason — the catalog is this application's, not platform's.
func (h *mcpToolManager) GetWebhookEventTypes() mcp.ToolHandlerFor[*GetWebhookEventTypesInvocation, *GetWebhookEventTypesResult] {
	return func(context.Context, *mcp.CallToolRequest, *GetWebhookEventTypesInvocation) (*mcp.CallToolResult, *GetWebhookEventTypesResult, error) {
		return nil, &GetWebhookEventTypesResult{Results: webhooks.EventTypeCatalog()}, nil
	}
}
