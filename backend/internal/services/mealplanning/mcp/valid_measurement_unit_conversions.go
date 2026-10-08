package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidMeasurementUnitConversionInvocation struct {
		ValidMeasurementUnitConversionID string `jsonschema:"description=The measurement unit conversion ID"`
	}
)

var validMeasurementUnitConversionsSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the valid measurement unit conversion"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the valid measurement unit conversion was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the valid measurement unit conversion was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the valid measurement unit conversion was soft deleted"),
	fieldNotes:                  mcptools.StringField("Notes about the measurement unit conversion"),
	"Modifier":                  mcptools.FloatField("The conversion modifier (multiplier to convert from 'From' unit to 'To' unit)"),
	"From":                      mcptools.ObjectType(validMeasurementUnitsSchema),
	"To":                        mcptools.ObjectType(validMeasurementUnitsSchema),
	"OnlyForIngredient":         mcptools.ObjectType(validIngredientsSchema),
}

var getValidMeasurementUnitConversionTool = &sdkmcp.Tool{
	Name:        "GetValidMeasurementUnitConversion",
	Description: "Get a valid measurement unit conversion by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		"ValidMeasurementUnitConversionID": mcptools.StringField("The ID of the valid measurement unit conversion to get"),
	}),
	OutputSchema: mcptools.SchemaObject(validMeasurementUnitConversionsSchema),
}

func (t *Tools) GetValidMeasurementUnitConversion() sdkmcp.ToolHandlerFor[*GetValidMeasurementUnitConversionInvocation, *mealplanning.ValidMeasurementUnitConversion] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidMeasurementUnitConversionInvocation) (*sdkmcp.CallToolResult, *mealplanning.ValidMeasurementUnitConversion, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetValidMeasurementUnitConversion(ctx, x.ValidMeasurementUnitConversionID)
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

var getValidMeasurementUnitConversionsForUnitTool = &sdkmcp.Tool{
	Name:        "GetValidMeasurementUnitConversionsForUnit",
	Description: "Get valid measurement unit conversions for a specific measurement unit with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter:     filtering.QueryFilterSchema(),
		"ValidMeasurementUnitID": mcptools.StringField("The ID of the valid measurement unit"),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validMeasurementUnitConversionsSchema)),
	}),
}

func (t *Tools) GetValidMeasurementUnitConversionsForUnit() sdkmcp.ToolHandlerFor[*GetValidMeasurementUnitConversionsForUnitInvocation, *GetValidMeasurementUnitConversionsForUnitResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidMeasurementUnitConversionsForUnitInvocation) (*sdkmcp.CallToolResult, *GetValidMeasurementUnitConversionsForUnitResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetValidMeasurementUnitConversionsForUnit(ctx, x.ValidMeasurementUnitID, x.Filter)
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

var getValidMeasurementUnitConversionsForIngredientsTool = &sdkmcp.Tool{
	Name:        "GetValidMeasurementUnitConversionsForIngredients",
	Description: "Get all valid measurement unit conversions applicable to the given ingredient IDs (universal conversions plus ingredient-specific ones)",
	InputSchema: mcptools.SchemaObject(map[string]any{
		"ValidIngredientIDs": mcptools.ArrayType(mcptools.StringField("A valid ingredient ID")),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validMeasurementUnitConversionsSchema)),
	}),
}

func (t *Tools) GetValidMeasurementUnitConversionsForIngredients() sdkmcp.ToolHandlerFor[*GetValidMeasurementUnitConversionsForIngredientsInvocation, *GetValidMeasurementUnitConversionsForIngredientsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidMeasurementUnitConversionsForIngredientsInvocation) (*sdkmcp.CallToolResult, *GetValidMeasurementUnitConversionsForIngredientsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetValidMeasurementUnitConversionsForIngredients(ctx, x.ValidIngredientIDs)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidMeasurementUnitConversionsForIngredientsResult{}
		out.Results = results
		return nil, out, nil
	}
}

//
