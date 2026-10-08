package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetValidMeasurementUnitInput is what GetValidMeasurementUnit takes.
type GetValidMeasurementUnitInput struct {
	ValidMeasurementUnitID string `json:"validMeasurementUnitID" jsonschema:"The identifier of the measurement unit to read"`
}

var getValidMeasurementUnitTool = &sdkmcp.Tool{
	Name:        "GetValidMeasurementUnit",
	Description: "Get a valid measurement unit by its ID",
	Annotations: readOnly(),
}

// GetValidMeasurementUnit reads one measurement unit from the catalog.
func (t *Tools) GetValidMeasurementUnit(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidMeasurementUnitInput) (*sdkmcp.CallToolResult, *mealplanning.ValidMeasurementUnit, error) {
	ctx, err := t.begin(ctx, req, getValidMeasurementUnitTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadValidMeasurementUnit(ctx, in.ValidMeasurementUnitID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

var searchForValidMeasurementUnitsTool = &sdkmcp.Tool{
	Name:        "SearchForValidMeasurementUnits",
	Description: "Search the valid measurement units by name",
	Annotations: readOnly(),
}

// SearchForValidMeasurementUnits pages the measurement units matching a query.
func (t *Tools) SearchForValidMeasurementUnits(ctx context.Context, req *sdkmcp.CallToolRequest, in SearchInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit], error) {
	ctx, err := t.begin(ctx, req, searchForValidMeasurementUnitsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.SearchValidMeasurementUnits(ctx, in.Query, useDatabaseSearch, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
