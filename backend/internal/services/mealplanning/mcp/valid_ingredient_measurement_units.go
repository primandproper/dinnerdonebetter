package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidIngredientMeasurementUnitInvocation struct {
		ValidIngredientMeasurementUnitID string `jsonschema:"description=The ingredient measurement unit ID"`
	}
)

var validIngredientMeasurementUnitsSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the valid ingredient measurement unit"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the valid ingredient measurement unit was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the valid ingredient measurement unit was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the valid ingredient measurement unit was soft deleted"),
	fieldNotes:                  mcptools.StringField("Notes about the ingredient measurement unit"),
	"MinAllowableQuantity":      mcptools.FloatField("Minimum allowable quantity (required)"),
	"MaxAllowableQuantity":      mcptools.FloatField("Maximum allowable quantity (optional)"),
	fieldMeasurementUnit:        mcptools.ObjectType(validMeasurementUnitsSchema),
	fieldIngredient:             mcptools.ObjectType(validIngredientsSchema),
}

var getValidIngredientMeasurementUnitTool = &sdkmcp.Tool{
	Name:        "GetValidIngredientMeasurementUnit",
	Description: "Get a valid ingredient measurement unit by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		"ValidIngredientMeasurementUnitID": mcptools.StringField("The ID of the valid ingredient measurement unit to get"),
	}),
	OutputSchema: mcptools.SchemaObject(validIngredientMeasurementUnitsSchema),
}

func (t *Tools) GetValidIngredientMeasurementUnit() sdkmcp.ToolHandlerFor[*GetValidIngredientMeasurementUnitInvocation, *mealplanning.ValidIngredientMeasurementUnit] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidIngredientMeasurementUnitInvocation) (*sdkmcp.CallToolResult, *mealplanning.ValidIngredientMeasurementUnit, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetValidIngredientMeasurementUnit(ctx, x.ValidIngredientMeasurementUnitID)
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

var getValidIngredientMeasurementUnitsTool = &sdkmcp.Tool{
	Name:        "GetValidIngredientMeasurementUnits",
	Description: "Get valid ingredient measurement units with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validIngredientMeasurementUnitsSchema)),
	}),
}

func (t *Tools) GetValidIngredientMeasurementUnits() sdkmcp.ToolHandlerFor[*GetValidIngredientMeasurementUnitsInvocation, *GetValidIngredientMeasurementUnitsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidIngredientMeasurementUnitsInvocation) (*sdkmcp.CallToolResult, *GetValidIngredientMeasurementUnitsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetValidIngredientMeasurementUnits(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidIngredientMeasurementUnitsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
