package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetValidIngredientMeasurementUnitInput is what GetValidIngredientMeasurementUnit takes.
type GetValidIngredientMeasurementUnitInput struct {
	ValidIngredientMeasurementUnitID string `json:"validIngredientMeasurementUnitID" jsonschema:"The identifier of the ingredient measurement unit to read"`
}

var getValidIngredientMeasurementUnitTool = &sdkmcp.Tool{
	Name:        "GetValidIngredientMeasurementUnit",
	Description: "Get a valid ingredient measurement unit by its ID",
	Annotations: readOnly(),
}

// GetValidIngredientMeasurementUnit reads one ingredient measurement unit from the catalog.
func (t *Tools) GetValidIngredientMeasurementUnit(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidIngredientMeasurementUnitInput) (*sdkmcp.CallToolResult, *mealplanning.ValidIngredientMeasurementUnit, error) {
	ctx, err := t.begin(ctx, req, getValidIngredientMeasurementUnitTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadValidIngredientMeasurementUnit(ctx, in.ValidIngredientMeasurementUnitID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

var getValidIngredientMeasurementUnitsTool = &sdkmcp.Tool{
	Name:        "GetValidIngredientMeasurementUnits",
	Description: "Page the valid ingredient measurement units",
	Annotations: readOnly(),
}

// GetValidIngredientMeasurementUnits pages the ingredient measurement units.
func (t *Tools) GetValidIngredientMeasurementUnits(ctx context.Context, req *sdkmcp.CallToolRequest, in FilterInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit], error) {
	ctx, err := t.begin(ctx, req, getValidIngredientMeasurementUnitsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ListValidIngredientMeasurementUnits(ctx, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
