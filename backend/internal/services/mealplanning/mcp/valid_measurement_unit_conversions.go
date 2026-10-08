package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetValidMeasurementUnitConversionInput is what GetValidMeasurementUnitConversion takes.
type GetValidMeasurementUnitConversionInput struct {
	ValidMeasurementUnitConversionID string `json:"validMeasurementUnitConversionID" jsonschema:"The identifier of the conversion to read"`
}

var getValidMeasurementUnitConversionTool = &sdkmcp.Tool{
	Name:        "GetValidMeasurementUnitConversion",
	Description: "Get a valid measurement unit conversion by its ID",
	Annotations: readOnly(),
}

// GetValidMeasurementUnitConversion reads one conversion.
func (t *Tools) GetValidMeasurementUnitConversion(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidMeasurementUnitConversionInput) (*sdkmcp.CallToolResult, *mealplanning.ValidMeasurementUnitConversion, error) {
	ctx, err := t.begin(ctx, req, getValidMeasurementUnitConversionTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadValidMeasurementUnitConversion(ctx, in.ValidMeasurementUnitConversionID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

// GetValidMeasurementUnitConversionsForUnitInput is what GetValidMeasurementUnitConversionsForUnit takes.
type GetValidMeasurementUnitConversionsForUnitInput struct {
	Filter                 *filtering.QueryFilter `json:"filter,omitempty"       jsonschema:"The page to read; absent reads the first page, oldest first"`
	ValidMeasurementUnitID string                 `json:"validMeasurementUnitID" jsonschema:"The identifier of the measurement unit converted from"`
}

var getValidMeasurementUnitConversionsForUnitTool = &sdkmcp.Tool{
	Name:        "GetValidMeasurementUnitConversionsForUnit",
	Description: "Page the valid measurement unit conversions from a measurement unit",
	Annotations: readOnly(),
}

// GetValidMeasurementUnitConversionsForUnit pages the conversions from one measurement unit.
func (t *Tools) GetValidMeasurementUnitConversionsForUnit(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidMeasurementUnitConversionsForUnitInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnitConversion], error) {
	ctx, err := t.begin(ctx, req, getValidMeasurementUnitConversionsForUnitTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ValidMeasurementUnitConversionsForMeasurementUnit(ctx, in.ValidMeasurementUnitID, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

// GetValidMeasurementUnitConversionsForIngredientsInput is what GetValidMeasurementUnitConversionsForIngredients takes.
type GetValidMeasurementUnitConversionsForIngredientsInput struct {
	ValidIngredientIDs []string `json:"validIngredientIDs" jsonschema:"The identifiers of the ingredients to find conversions for"`
}

// ValidMeasurementUnitConversions is what GetValidMeasurementUnitConversionsForIngredients answers with.
type ValidMeasurementUnitConversions struct {
	Results []*mealplanning.ValidMeasurementUnitConversion `json:"results"`
}

var getValidMeasurementUnitConversionsForIngredientsTool = &sdkmcp.Tool{
	Name:        "GetValidMeasurementUnitConversionsForIngredients",
	Description: "List every valid measurement unit conversion applicable to the given ingredients: the universal conversions plus the ingredient-specific ones",
	Annotations: readOnly(),
}

// GetValidMeasurementUnitConversionsForIngredients lists the conversions that apply to the ingredients named.
func (t *Tools) GetValidMeasurementUnitConversionsForIngredients(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidMeasurementUnitConversionsForIngredientsInput) (*sdkmcp.CallToolResult, *ValidMeasurementUnitConversions, error) {
	ctx, err := t.begin(ctx, req, getValidMeasurementUnitConversionsForIngredientsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.GetValidMeasurementUnitConversionsForIngredients(ctx, in.ValidIngredientIDs)
	if err != nil {
		return nil, nil, err
	}

	return nil, &ValidMeasurementUnitConversions{Results: result}, nil
}
