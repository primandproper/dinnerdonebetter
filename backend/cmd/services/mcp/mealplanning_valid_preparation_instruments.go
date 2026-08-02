package main

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidPreparationInstrumentInvocation struct {
		ValidPreparationInstrumentID string `jsonschema:"description=The preparation instrument ID"`
	}
)

var validPreparationInstrumentsSchema = map[string]any{
	"ID":            stringField("The ID of the valid preparation instrument"),
	"CreatedAt":     timestampField("When the valid preparation instrument was created"),
	"LastUpdatedAt": timestampField("When the valid preparation instrument was last updated"),
	"ArchivedAt":    timestampField("When the valid preparation instrument was soft deleted"),
	"Notes":         stringField("Notes about the preparation instrument"),
	"Instrument":    objectType(validInstrumentsSchema),
	"Preparation":   objectType(validPreparationsSchema),
}

var getValidPreparationInstrumentTool = &mcp.Tool{
	Name:        "GetValidPreparationInstrument",
	Description: "Get a valid preparation instrument by it's ID",
	InputSchema: schemaObject(map[string]any{
		"ValidPreparationInstrumentID": stringField("The ID of the valid preparation instrument to get"),
	}),
	OutputSchema: schemaObject(validPreparationInstrumentsSchema),
}

func (h *mcpToolManager) GetValidPreparationInstrument() mcp.ToolHandlerFor[*GetValidPreparationInstrumentInvocation, *mealplanning.ValidPreparationInstrument] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidPreparationInstrumentInvocation) (*mcp.CallToolResult, *mealplanning.ValidPreparationInstrument, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetValidPreparationInstrument(ctx, x.ValidPreparationInstrumentID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetValidPreparationInstrumentsInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetValidPreparationInstrumentsResult struct {
		Results []*mealplanning.ValidPreparationInstrument
	}
)

var getValidPreparationInstrumentsTool = &mcp.Tool{
	Name:        "GetValidPreparationInstruments",
	Description: "Get valid preparation instruments with optional filtering",
	InputSchema: schemaObject(map[string]any{
		"Filter": queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(validPreparationInstrumentsSchema)),
	}),
}

func (h *mcpToolManager) GetValidPreparationInstruments() mcp.ToolHandlerFor[*GetValidPreparationInstrumentsInvocation, *GetValidPreparationInstrumentsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidPreparationInstrumentsInvocation) (*mcp.CallToolResult, *GetValidPreparationInstrumentsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetValidPreparationInstruments(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidPreparationInstrumentsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
