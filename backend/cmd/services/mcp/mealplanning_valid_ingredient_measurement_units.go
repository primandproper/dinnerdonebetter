package main

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v8/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidIngredientMeasurementUnitInvocation struct {
		ValidIngredientMeasurementUnitID string `jsonschema:"description=The ingredient measurement unit ID"`
	}
)

var validIngredientMeasurementUnitsSchema = map[string]any{
	"ID":                   stringField("The ID of the valid ingredient measurement unit"),
	"CreatedAt":            timestampField("When the valid ingredient measurement unit was created"),
	"LastUpdatedAt":        timestampField("When the valid ingredient measurement unit was last updated"),
	"ArchivedAt":           timestampField("When the valid ingredient measurement unit was soft deleted"),
	"Notes":                stringField("Notes about the ingredient measurement unit"),
	"MinAllowableQuantity": floatField("Minimum allowable quantity (required)"),
	"MaxAllowableQuantity": floatField("Maximum allowable quantity (optional)"),
	"MeasurementUnit":      objectType(validMeasurementUnitsSchema),
	"Ingredient":           objectType(validIngredientsSchema),
}

var getValidIngredientMeasurementUnitTool = &mcp.Tool{
	Name:        "GetValidIngredientMeasurementUnit",
	Description: "Get a valid ingredient measurement unit by it's ID",
	InputSchema: schemaObject(map[string]any{
		"ValidIngredientMeasurementUnitID": stringField("The ID of the valid ingredient measurement unit to get"),
	}),
	OutputSchema: schemaObject(validIngredientMeasurementUnitsSchema),
}

func (h *mcpToolManager) GetValidIngredientMeasurementUnit() mcp.ToolHandlerFor[*GetValidIngredientMeasurementUnitInvocation, *mealplanning.ValidIngredientMeasurementUnit] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidIngredientMeasurementUnitInvocation) (*mcp.CallToolResult, *mealplanning.ValidIngredientMeasurementUnit, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetValidIngredientMeasurementUnit(ctx, x.ValidIngredientMeasurementUnitID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetValidIngredientMeasurementUnitsInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetValidIngredientMeasurementUnitsResult struct {
		Results []*mealplanning.ValidIngredientMeasurementUnit
	}
)

var getValidIngredientMeasurementUnitsTool = &mcp.Tool{
	Name:        "GetValidIngredientMeasurementUnits",
	Description: "Get valid ingredient measurement units with optional filtering",
	InputSchema: schemaObject(map[string]any{
		"Filter": queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(validIngredientMeasurementUnitsSchema)),
	}),
}

func (h *mcpToolManager) GetValidIngredientMeasurementUnits() mcp.ToolHandlerFor[*GetValidIngredientMeasurementUnitsInvocation, *GetValidIngredientMeasurementUnitsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidIngredientMeasurementUnitsInvocation) (*mcp.CallToolResult, *GetValidIngredientMeasurementUnitsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetValidIngredientMeasurementUnits(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidIngredientMeasurementUnitsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
