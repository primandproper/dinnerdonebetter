package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidPreparationInstrumentInvocation struct {
		ValidPreparationInstrumentID string `jsonschema:"description=The preparation instrument ID"`
	}
)

var validPreparationInstrumentsSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the valid preparation instrument"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the valid preparation instrument was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the valid preparation instrument was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the valid preparation instrument was soft deleted"),
	fieldNotes:                  mcptools.StringField("Notes about the preparation instrument"),
	"Instrument":                mcptools.ObjectType(validInstrumentsSchema),
	fieldPreparation:            mcptools.ObjectType(validPreparationsSchema),
}

var getValidPreparationInstrumentTool = &sdkmcp.Tool{
	Name:        "GetValidPreparationInstrument",
	Description: "Get a valid preparation instrument by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		"ValidPreparationInstrumentID": mcptools.StringField("The ID of the valid preparation instrument to get"),
	}),
	OutputSchema: mcptools.SchemaObject(validPreparationInstrumentsSchema),
}

func (t *Tools) GetValidPreparationInstrument() sdkmcp.ToolHandlerFor[*GetValidPreparationInstrumentInvocation, *mealplanning.ValidPreparationInstrument] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidPreparationInstrumentInvocation) (*sdkmcp.CallToolResult, *mealplanning.ValidPreparationInstrument, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetValidPreparationInstrument(ctx, x.ValidPreparationInstrumentID)
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

var getValidPreparationInstrumentsTool = &sdkmcp.Tool{
	Name:        "GetValidPreparationInstruments",
	Description: "Get valid preparation instruments with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validPreparationInstrumentsSchema)),
	}),
}

func (t *Tools) GetValidPreparationInstruments() sdkmcp.ToolHandlerFor[*GetValidPreparationInstrumentsInvocation, *GetValidPreparationInstrumentsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidPreparationInstrumentsInvocation) (*sdkmcp.CallToolResult, *GetValidPreparationInstrumentsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetValidPreparationInstruments(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidPreparationInstrumentsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
