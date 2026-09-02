package mcpserver

import (
	"context"

	ddbwaitlists "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"

	"github.com/primandproper/platform-go/v13/filtering"
	waitlists "github.com/primandproper/platform-go/v13/waitlists"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// The waitlist tools are the catalog and nothing else.
//
// The signups are deliberately not reachable from here, and that is a change
// this adoption forced rather than a gap. The table this replaced held a note
// and two ownership columns; the platform's holds the address the list writes
// to, so a tool that paged one list's signups would hand a model every
// signatory's email. Over gRPC that read is service-admin-only, and MCP has no
// equivalent — the token carries an account and no role — so there is nothing
// here to gate it with.
var waitlistSchema = map[string]any{
	"ID":               stringField("The ID of the waitlist"),
	fieldName:          stringField("The waitlist name"),
	fieldDescription:   stringField("The waitlist description"),
	"ClosesAt":         timestampField("When the waitlist stops taking signups"),
	fieldCreatedAt:     timestampField("When the waitlist was created"),
	fieldLastUpdatedAt: timestampField("When the waitlist was last updated"),
	fieldArchivedAt:    timestampField("When the waitlist was archived"),
}

var getWaitlistTool = &mcp.Tool{
	Name:        "GetWaitlist",
	Description: "Get a waitlist by its ID",
	InputSchema: schemaObject(map[string]any{
		fieldWaitlistID: stringField("The ID of the waitlist to get"),
	}),
	OutputSchema: schemaObject(waitlistSchema),
}

type GetWaitlistInvocation struct {
	WaitlistID string `jsonschema:"description=The waitlist ID"`
}

func (h *mcpToolManager) GetWaitlist() mcp.ToolHandlerFor[*GetWaitlistInvocation, *waitlists.List] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetWaitlistInvocation) (*mcp.CallToolResult, *waitlists.List, error) {
		result, err := h.waitlists.GetList(ctx, ddbwaitlists.Scope(), x.WaitlistID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

var getWaitlistsTool = &mcp.Tool{
	Name:        "GetWaitlists",
	Description: "Get waitlists with optional filtering",
	InputSchema: schemaObject(map[string]any{
		fieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		fieldResults: arrayType(schemaObject(waitlistSchema)),
	}),
}

type (
	GetWaitlistsInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetWaitlistsResult struct {
		Results []*waitlists.List
	}
)

func (h *mcpToolManager) GetWaitlists() mcp.ToolHandlerFor[*GetWaitlistsInvocation, *GetWaitlistsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetWaitlistsInvocation) (*mcp.CallToolResult, *GetWaitlistsResult, error) {
		results, err := h.waitlists.ListLists(ctx, ddbwaitlists.Scope(), x.Filter)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetWaitlistsResult{Results: results.Data}, nil
	}
}

var getOpenWaitlistsTool = &mcp.Tool{
	Name:        "GetOpenWaitlists",
	Description: "Get the waitlists that are still taking signups",
	InputSchema: schemaObject(map[string]any{
		fieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		fieldResults: arrayType(schemaObject(waitlistSchema)),
	}),
}

type (
	GetOpenWaitlistsInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetOpenWaitlistsResult struct {
		Results []*waitlists.List
	}
)

func (h *mcpToolManager) GetOpenWaitlists() mcp.ToolHandlerFor[*GetOpenWaitlistsInvocation, *GetOpenWaitlistsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetOpenWaitlistsInvocation) (*mcp.CallToolResult, *GetOpenWaitlistsResult, error) {
		results, err := h.waitlists.ListOpenLists(ctx, ddbwaitlists.Scope(), x.Filter)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetOpenWaitlistsResult{Results: results.Data}, nil
	}
}
