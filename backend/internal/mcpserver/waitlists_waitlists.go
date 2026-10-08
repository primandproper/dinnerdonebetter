package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	waitlists "github.com/primandproper/platform-go/v15/waitlists"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"

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
	"ID":                        mcptools.StringField("The ID of the waitlist"),
	mcptools.FieldName:          mcptools.StringField("The waitlist name"),
	mcptools.FieldDescription:   mcptools.StringField("The waitlist description"),
	"ClosesAt":                  mcptools.TimestampField("When the waitlist stops taking signups"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the waitlist was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the waitlist was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the waitlist was archived"),
}

var getWaitlistTool = &mcp.Tool{
	Name:        "GetWaitlist",
	Description: "Get a waitlist by its ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldWaitlistID: mcptools.StringField("The ID of the waitlist to get"),
	}),
	OutputSchema: mcptools.SchemaObject(waitlistSchema),
}

type GetWaitlistInvocation struct {
	WaitlistID string `jsonschema:"description=The waitlist ID"`
}

func (h *mcpToolManager) GetWaitlist() mcp.ToolHandlerFor[*GetWaitlistInvocation, *waitlists.List] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetWaitlistInvocation) (*mcp.CallToolResult, *waitlists.List, error) {
		result, err := h.waitlists.GetList(ctx, h.reader, tenancy.Global(), x.WaitlistID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

var getWaitlistsTool = &mcp.Tool{
	Name:        "GetWaitlists",
	Description: "Get waitlists with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(waitlistSchema)),
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
		results, err := h.waitlists.ListLists(ctx, h.reader, tenancy.Global(), x.Filter)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetWaitlistsResult{Results: results.Data}, nil
	}
}

var getOpenWaitlistsTool = &mcp.Tool{
	Name:        "GetOpenWaitlists",
	Description: "Get the waitlists that are still taking signups",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(waitlistSchema)),
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
		results, err := h.waitlists.ListOpenLists(ctx, h.reader, tenancy.Global(), x.Filter)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetOpenWaitlistsResult{Results: results.Data}, nil
	}
}
