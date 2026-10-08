package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetValidPreparationVesselInput is what GetValidPreparationVessel takes.
type GetValidPreparationVesselInput struct {
	ValidPreparationVesselID string `json:"validPreparationVesselID" jsonschema:"The identifier of the preparation vessel to read"`
}

var getValidPreparationVesselTool = &sdkmcp.Tool{
	Name:        "GetValidPreparationVessel",
	Description: "Get a valid preparation vessel by its ID",
	Annotations: readOnly(),
}

// GetValidPreparationVessel reads one preparation vessel from the catalog.
func (t *Tools) GetValidPreparationVessel(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidPreparationVesselInput) (*sdkmcp.CallToolResult, *mealplanning.ValidPreparationVessel, error) {
	ctx, err := t.begin(ctx, req, getValidPreparationVesselTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadValidPreparationVessel(ctx, in.ValidPreparationVesselID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

var getValidPreparationVesselsTool = &sdkmcp.Tool{
	Name:        "GetValidPreparationVessels",
	Description: "Page the valid preparation vessels",
	Annotations: readOnly(),
}

// GetValidPreparationVessels pages the preparation vessels.
func (t *Tools) GetValidPreparationVessels(ctx context.Context, req *sdkmcp.CallToolRequest, in FilterInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel], error) {
	ctx, err := t.begin(ctx, req, getValidPreparationVesselsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ListValidPreparationVessels(ctx, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
