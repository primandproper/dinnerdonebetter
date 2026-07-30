package main

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v8/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidInstrumentInvocation struct {
		ValidInstrumentID string `jsonschema:"description=The instrument ID"`
	}
)

var validInstrumentsSchema = map[string]any{
	"ID":                             stringField("The ID of the valid instrument"),
	"CreatedAt":                      timestampField("When the valid instrument was created"),
	"LastUpdatedAt":                  timestampField("When the valid instrument was last updated"),
	"ArchivedAt":                     timestampField("When the valid instrument was soft deleted"),
	"Name":                           stringField("Name of the instrument"),
	"Description":                    stringField("Description of the instrument"),
	"IconPath":                       stringField("The URL for the icon for the item"),
	"PluralName":                     stringField("The plural name for the instrument. So for an instrument named 'knife', this would be 'knives'"),
	"Slug":                           stringField("An easy-to-use URL slug for the instrument"),
	"IncludeInGeneratedInstructions": boolField("Whether or not the valid instrument should be included in generated instructions"),
	"DisplayInSummaryLists":          boolField("Whether or not the valid instrument should be displayed in summary lists"),
	"UsableForStorage":               boolField("Whether or not the valid instrument is usable for storage"),
}

var getValidInstrumentTool = &mcp.Tool{
	Name:        "GetValidInstrument",
	Description: "Get a valid instrument by it's ID",
	InputSchema: schemaObject(map[string]any{
		"ValidInstrumentID": stringField("The ID of the valid instrument to get"),
	}),
	OutputSchema: schemaObject(validInstrumentsSchema),
}

func (h *mcpToolManager) GetValidInstrument() mcp.ToolHandlerFor[*GetValidInstrumentInvocation, *mealplanning.ValidInstrument] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidInstrumentInvocation) (*mcp.CallToolResult, *mealplanning.ValidInstrument, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetValidInstrument(ctx, x.ValidInstrumentID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	SearchValidInstrumentsInvocation struct {
		Filter *filtering.QueryFilter
		Query  string `jsonschema_description:"The instrument name query"`
	}

	SearchValidInstrumentsResult struct {
		Results []*mealplanning.ValidInstrument
	}
)

var searchForValidInstrumentsTool = &mcp.Tool{
	Name:        "SearchForValidInstruments",
	Description: "Search for valid instruments with optional filtering and query string",
	InputSchema: schemaObject(map[string]any{
		"Filter": queryFilterSchema(),
		"Query": map[string]any{
			"type":        strType,
			"description": "The instrument name query",
		},
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(validInstrumentsSchema)),
	}),
}

func (h *mcpToolManager) SearchForValidInstruments() mcp.ToolHandlerFor[*SearchValidInstrumentsInvocation, *SearchValidInstrumentsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *SearchValidInstrumentsInvocation) (*mcp.CallToolResult, *SearchValidInstrumentsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.SearchForValidInstruments(ctx, x.Query, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &SearchValidInstrumentsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
