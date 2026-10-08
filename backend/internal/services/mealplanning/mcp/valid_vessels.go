package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetValidVesselInput is what GetValidVessel takes.
type GetValidVesselInput struct {
	ValidVesselID string `json:"validVesselID" jsonschema:"The identifier of the vessel to read"`
}

var getValidVesselTool = &sdkmcp.Tool{
	Name:        "GetValidVessel",
	Description: "Get a valid vessel by its ID",
	Annotations: readOnly(),
}

// GetValidVessel reads one vessel from the catalog.
func (t *Tools) GetValidVessel(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidVesselInput) (*sdkmcp.CallToolResult, *mealplanning.ValidVessel, error) {
	ctx, err := t.begin(ctx, req, getValidVesselTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadValidVessel(ctx, in.ValidVesselID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

var searchForValidVesselsTool = &sdkmcp.Tool{
	Name:        "SearchForValidVessels",
	Description: "Search the valid vessels by name",
	Annotations: readOnly(),
}

// SearchForValidVessels pages the vessels matching a query.
func (t *Tools) SearchForValidVessels(ctx context.Context, req *sdkmcp.CallToolRequest, in SearchInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidVessel], error) {
	ctx, err := t.begin(ctx, req, searchForValidVesselsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.SearchValidVessels(ctx, in.Query, useDatabaseSearch, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
