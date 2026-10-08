package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidPreparationVesselInvocation struct {
		ValidPreparationVesselID string `jsonschema:"description=The preparation vessel ID"`
	}
)

var validPreparationVesselsSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the valid preparation vessel"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the valid preparation vessel was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the valid preparation vessel was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the valid preparation vessel was soft deleted"),
	fieldNotes:                  mcptools.StringField("Notes about the preparation vessel"),
	"Vessel":                    mcptools.ObjectType(validVesselsSchema),
	fieldPreparation:            mcptools.ObjectType(validPreparationsSchema),
}

var getValidPreparationVesselTool = &sdkmcp.Tool{
	Name:        "GetValidPreparationVessel",
	Description: "Get a valid preparation vessel by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		"ValidPreparationVesselID": mcptools.StringField("The ID of the valid preparation vessel to get"),
	}),
	OutputSchema: mcptools.SchemaObject(validPreparationVesselsSchema),
}

func (t *Tools) GetValidPreparationVessel() sdkmcp.ToolHandlerFor[*GetValidPreparationVesselInvocation, *mealplanning.ValidPreparationVessel] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidPreparationVesselInvocation) (*sdkmcp.CallToolResult, *mealplanning.ValidPreparationVessel, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetValidPreparationVessel(ctx, x.ValidPreparationVesselID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetValidPreparationVesselsInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetValidPreparationVesselsResult struct {
		Results []*mealplanning.ValidPreparationVessel
	}
)

var getValidPreparationVesselsTool = &sdkmcp.Tool{
	Name:        "GetValidPreparationVessels",
	Description: "Get valid preparation vessels with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validPreparationVesselsSchema)),
	}),
}

func (t *Tools) GetValidPreparationVessels() sdkmcp.ToolHandlerFor[*GetValidPreparationVesselsInvocation, *GetValidPreparationVesselsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidPreparationVesselsInvocation) (*sdkmcp.CallToolResult, *GetValidPreparationVesselsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetValidPreparationVessels(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidPreparationVesselsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
