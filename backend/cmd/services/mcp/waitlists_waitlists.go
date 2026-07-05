package main

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists"

	"github.com/primandproper/platform-go/v3/database/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var waitlistSchema = map[string]any{
	"ID":            stringField("The ID of the waitlist"),
	"Name":          stringField("The waitlist name"),
	"Description":   stringField("The waitlist description"),
	"ValidUntil":    timestampField("When the waitlist expires"),
	"CreatedAt":     timestampField("When the waitlist was created"),
	"LastUpdatedAt": timestampField("When the waitlist was last updated"),
	"ArchivedAt":    timestampField("When the waitlist was archived"),
}

var waitlistSignupSchema = map[string]any{
	"ID":                stringField("The ID of the waitlist signup"),
	"Notes":             stringField("Notes about the signup"),
	"BelongsToWaitlist": stringField("The ID of the waitlist"),
	"BelongsToUser":     stringField("The ID of the user who signed up"),
	"BelongsToAccount":  stringField("The ID of the account"),
	"CreatedAt":         timestampField("When the signup was created"),
	"LastUpdatedAt":     timestampField("When the signup was last updated"),
	"ArchivedAt":        timestampField("When the signup was archived"),
}

var getWaitlistTool = &mcp.Tool{
	Name:        "GetWaitlist",
	Description: "Get a waitlist by its ID",
	InputSchema: schemaObject(map[string]any{
		"WaitlistID": stringField("The ID of the waitlist to get"),
	}),
	OutputSchema: schemaObject(waitlistSchema),
}

type GetWaitlistInvocation struct {
	WaitlistID string `jsonschema:"description=The waitlist ID"`
}

func (h *mcpToolManager) GetWaitlist() mcp.ToolHandlerFor[*GetWaitlistInvocation, *waitlists.Waitlist] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetWaitlistInvocation) (*mcp.CallToolResult, *waitlists.Waitlist, error) {
		result, err := h.waitlistsRepo.GetWaitlist(ctx, x.WaitlistID)
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
		"Filter": queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(waitlistSchema)),
	}),
}

type (
	GetWaitlistsInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetWaitlistsResult struct {
		Results []*waitlists.Waitlist
	}
)

func (h *mcpToolManager) GetWaitlists() mcp.ToolHandlerFor[*GetWaitlistsInvocation, *GetWaitlistsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetWaitlistsInvocation) (*mcp.CallToolResult, *GetWaitlistsResult, error) {
		results, err := h.waitlistsRepo.GetWaitlists(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetWaitlistsResult{Results: results.Data}, nil
	}
}

var getActiveWaitlistsTool = &mcp.Tool{
	Name:        "GetActiveWaitlists",
	Description: "Get waitlists that are currently active (not expired)",
	InputSchema: schemaObject(map[string]any{
		"Filter": queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(waitlistSchema)),
	}),
}

type (
	GetActiveWaitlistsInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetActiveWaitlistsResult struct {
		Results []*waitlists.Waitlist
	}
)

func (h *mcpToolManager) GetActiveWaitlists() mcp.ToolHandlerFor[*GetActiveWaitlistsInvocation, *GetActiveWaitlistsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetActiveWaitlistsInvocation) (*mcp.CallToolResult, *GetActiveWaitlistsResult, error) {
		results, err := h.waitlistsRepo.GetActiveWaitlists(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetActiveWaitlistsResult{Results: results.Data}, nil
	}
}

var getWaitlistSignupTool = &mcp.Tool{
	Name:        "GetWaitlistSignup",
	Description: "Get a specific waitlist signup",
	InputSchema: schemaObject(map[string]any{
		"WaitlistSignupID": stringField("The ID of the waitlist signup"),
		"WaitlistID":       stringField("The ID of the waitlist"),
	}),
	OutputSchema: schemaObject(waitlistSignupSchema),
}

type GetWaitlistSignupInvocation struct {
	WaitlistSignupID string `jsonschema:"description=The waitlist signup ID"`
	WaitlistID       string `jsonschema:"description=The waitlist ID"`
}

func (h *mcpToolManager) GetWaitlistSignup() mcp.ToolHandlerFor[*GetWaitlistSignupInvocation, *waitlists.WaitlistSignup] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetWaitlistSignupInvocation) (*mcp.CallToolResult, *waitlists.WaitlistSignup, error) {
		result, err := h.waitlistsRepo.GetWaitlistSignup(ctx, x.WaitlistSignupID, x.WaitlistID)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	}
}

var getWaitlistSignupsForWaitlistTool = &mcp.Tool{
	Name:        "GetWaitlistSignupsForWaitlist",
	Description: "Get all signups for a specific waitlist",
	InputSchema: schemaObject(map[string]any{
		"WaitlistID": stringField("The ID of the waitlist"),
		"Filter":     queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(waitlistSignupSchema)),
	}),
}

type (
	GetWaitlistSignupsForWaitlistInvocation struct {
		Filter     *filtering.QueryFilter
		WaitlistID string `jsonschema:"description=The waitlist ID"`
	}

	GetWaitlistSignupsForWaitlistResult struct {
		Results []*waitlists.WaitlistSignup
	}
)

func (h *mcpToolManager) GetWaitlistSignupsForWaitlist() mcp.ToolHandlerFor[*GetWaitlistSignupsForWaitlistInvocation, *GetWaitlistSignupsForWaitlistResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetWaitlistSignupsForWaitlistInvocation) (*mcp.CallToolResult, *GetWaitlistSignupsForWaitlistResult, error) {
		results, err := h.waitlistsRepo.GetWaitlistSignupsForWaitlist(ctx, x.WaitlistID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetWaitlistSignupsForWaitlistResult{Results: results.Data}, nil
	}
}
