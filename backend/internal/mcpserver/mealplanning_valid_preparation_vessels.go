package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v13/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidPreparationVesselInvocation struct {
		ValidPreparationVesselID string `jsonschema:"description=The preparation vessel ID"`
	}
)

var validPreparationVesselsSchema = map[string]any{
	"ID":               stringField("The ID of the valid preparation vessel"),
	fieldCreatedAt:     timestampField("When the valid preparation vessel was created"),
	fieldLastUpdatedAt: timestampField("When the valid preparation vessel was last updated"),
	fieldArchivedAt:    timestampField("When the valid preparation vessel was soft deleted"),
	fieldNotes:         stringField("Notes about the preparation vessel"),
	"Vessel":           objectType(validVesselsSchema),
	fieldPreparation:   objectType(validPreparationsSchema),
}

var getValidPreparationVesselTool = &mcp.Tool{
	Name:        "GetValidPreparationVessel",
	Description: "Get a valid preparation vessel by it's ID",
	InputSchema: schemaObject(map[string]any{
		"ValidPreparationVesselID": stringField("The ID of the valid preparation vessel to get"),
	}),
	OutputSchema: schemaObject(validPreparationVesselsSchema),
}

func (h *mcpToolManager) GetValidPreparationVessel() mcp.ToolHandlerFor[*GetValidPreparationVesselInvocation, *mealplanning.ValidPreparationVessel] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidPreparationVesselInvocation) (*mcp.CallToolResult, *mealplanning.ValidPreparationVessel, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetValidPreparationVessel(ctx, x.ValidPreparationVesselID)
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

var getValidPreparationVesselsTool = &mcp.Tool{
	Name:        "GetValidPreparationVessels",
	Description: "Get valid preparation vessels with optional filtering",
	InputSchema: schemaObject(map[string]any{
		fieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		fieldResults: arrayType(schemaObject(validPreparationVesselsSchema)),
	}),
}

func (h *mcpToolManager) GetValidPreparationVessels() mcp.ToolHandlerFor[*GetValidPreparationVesselsInvocation, *GetValidPreparationVesselsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidPreparationVesselsInvocation) (*mcp.CallToolResult, *GetValidPreparationVesselsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetValidPreparationVessels(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidPreparationVesselsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
