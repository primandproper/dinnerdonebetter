package main

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidMeasurementUnitConversionInvocation struct {
		ValidMeasurementUnitConversionID string `jsonschema:"description=The measurement unit conversion ID"`
	}
)

var validMeasurementUnitConversionsSchema = map[string]any{
	"ID":                stringField("The ID of the valid measurement unit conversion"),
	"CreatedAt":         timestampField("When the valid measurement unit conversion was created"),
	"LastUpdatedAt":     timestampField("When the valid measurement unit conversion was last updated"),
	"ArchivedAt":        timestampField("When the valid measurement unit conversion was soft deleted"),
	"Notes":             stringField("Notes about the measurement unit conversion"),
	"Modifier":          floatField("The conversion modifier (multiplier to convert from 'From' unit to 'To' unit)"),
	"From":              objectType(validMeasurementUnitsSchema),
	"To":                objectType(validMeasurementUnitsSchema),
	"OnlyForIngredient": objectType(validIngredientsSchema),
}

var getValidMeasurementUnitConversionTool = &mcp.Tool{
	Name:        "GetValidMeasurementUnitConversion",
	Description: "Get a valid measurement unit conversion by it's ID",
	InputSchema: schemaObject(map[string]any{
		"ValidMeasurementUnitConversionID": stringField("The ID of the valid measurement unit conversion to get"),
	}),
	OutputSchema: schemaObject(validMeasurementUnitConversionsSchema),
}

func (h *mcpToolManager) GetValidMeasurementUnitConversion() mcp.ToolHandlerFor[*GetValidMeasurementUnitConversionInvocation, *mealplanning.ValidMeasurementUnitConversion] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidMeasurementUnitConversionInvocation) (*mcp.CallToolResult, *mealplanning.ValidMeasurementUnitConversion, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetValidMeasurementUnitConversion(ctx, x.ValidMeasurementUnitConversionID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetValidMeasurementUnitConversionsForUnitInvocation struct {
		Filter                 *filtering.QueryFilter
		ValidMeasurementUnitID string `jsonschema:"description=The measurement unit ID"`
	}

	GetValidMeasurementUnitConversionsForUnitResult struct {
		Results []*mealplanning.ValidMeasurementUnitConversion
	}
)

var getValidMeasurementUnitConversionsForUnitTool = &mcp.Tool{
	Name:        "GetValidMeasurementUnitConversionsForUnit",
	Description: "Get valid measurement unit conversions for a specific measurement unit with optional filtering",
	InputSchema: schemaObject(map[string]any{
		"Filter":                 queryFilterSchema(),
		"ValidMeasurementUnitID": stringField("The ID of the valid measurement unit"),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(validMeasurementUnitConversionsSchema)),
	}),
}

func (h *mcpToolManager) GetValidMeasurementUnitConversionsForUnit() mcp.ToolHandlerFor[*GetValidMeasurementUnitConversionsForUnitInvocation, *GetValidMeasurementUnitConversionsForUnitResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidMeasurementUnitConversionsForUnitInvocation) (*mcp.CallToolResult, *GetValidMeasurementUnitConversionsForUnitResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetValidMeasurementUnitConversionsForUnit(ctx, x.ValidMeasurementUnitID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidMeasurementUnitConversionsForUnitResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

type (
	GetValidMeasurementUnitConversionsForIngredientsInvocation struct {
		ValidIngredientIDs []string `jsonschema:"description=The valid ingredient IDs to fetch conversions for"`
	}

	GetValidMeasurementUnitConversionsForIngredientsResult struct {
		Results []*mealplanning.ValidMeasurementUnitConversion
	}
)

var getValidMeasurementUnitConversionsForIngredientsTool = &mcp.Tool{
	Name:        "GetValidMeasurementUnitConversionsForIngredients",
	Description: "Get all valid measurement unit conversions applicable to the given ingredient IDs (universal conversions plus ingredient-specific ones)",
	InputSchema: schemaObject(map[string]any{
		"ValidIngredientIDs": arrayType(stringField("A valid ingredient ID")),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(validMeasurementUnitConversionsSchema)),
	}),
}

func (h *mcpToolManager) GetValidMeasurementUnitConversionsForIngredients() mcp.ToolHandlerFor[*GetValidMeasurementUnitConversionsForIngredientsInvocation, *GetValidMeasurementUnitConversionsForIngredientsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidMeasurementUnitConversionsForIngredientsInvocation) (*mcp.CallToolResult, *GetValidMeasurementUnitConversionsForIngredientsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetValidMeasurementUnitConversionsForIngredients(ctx, x.ValidIngredientIDs)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidMeasurementUnitConversionsForIngredientsResult{}
		out.Results = results
		return nil, out, nil
	}
}

//
