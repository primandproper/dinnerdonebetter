package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidMeasurementUnitInvocation struct {
		ValidMeasurementUnitID string `jsonschema:"description=The measurement unit ID"`
	}
)

var validMeasurementUnitsSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the valid measurement unit"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the valid measurement unit was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the valid measurement unit was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the valid measurement unit was soft deleted"),
	mcptools.FieldName:          mcptools.StringField("Name of the measurement unit"),
	mcptools.FieldDescription:   mcptools.StringField("Description of the measurement unit"),
	fieldIconPath:               mcptools.StringField("The URL for the icon for the item"),
	fieldPluralName:             mcptools.StringField("The plural name for the measurement unit. So for a unit named 'cup', this would be 'cups'"),
	fieldSlug:                   mcptools.StringField("An easy-to-use URL slug for the measurement unit"),
	"Volumetric":                mcptools.BoolField("Whether or not the valid measurement unit is volumetric"),
	"Universal":                 mcptools.BoolField("Whether or not the valid measurement unit is universal (valid for all ingredients). For instance, 'grams' is a universal measurement unit"),
	"Metric":                    mcptools.BoolField("Whether or not the valid measurement unit is metric"),
	"Imperial":                  mcptools.BoolField("Whether or not the valid measurement unit is imperial"),
}

var getValidMeasurementUnitTool = &sdkmcp.Tool{
	Name:        "GetValidMeasurementUnit",
	Description: "Get a valid measurement unit by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		"ValidMeasurementUnitID": mcptools.StringField("The ID of the valid measurement unit to get"),
	}),
	OutputSchema: mcptools.SchemaObject(validMeasurementUnitsSchema),
}

func (t *Tools) GetValidMeasurementUnit() sdkmcp.ToolHandlerFor[*GetValidMeasurementUnitInvocation, *mealplanning.ValidMeasurementUnit] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidMeasurementUnitInvocation) (*sdkmcp.CallToolResult, *mealplanning.ValidMeasurementUnit, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetValidMeasurementUnit(ctx, x.ValidMeasurementUnitID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	SearchValidMeasurementUnitsInvocation struct {
		Filter *filtering.QueryFilter
		Query  string `jsonschema_description:"The measurement unit name query"`
	}

	SearchValidMeasurementUnitsResult struct {
		Results []*mealplanning.ValidMeasurementUnit
	}
)

var searchForValidMeasurementUnitsTool = &sdkmcp.Tool{
	Name:        "SearchForValidMeasurementUnits",
	Description: "Search for valid measurement units with optional filtering and query string",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
		mcptools.FieldQuery:  mcptools.StringField("The measurement unit name query"),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validMeasurementUnitsSchema)),
	}),
}

func (t *Tools) SearchForValidMeasurementUnits() sdkmcp.ToolHandlerFor[*SearchValidMeasurementUnitsInvocation, *SearchValidMeasurementUnitsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *SearchValidMeasurementUnitsInvocation) (*sdkmcp.CallToolResult, *SearchValidMeasurementUnitsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.SearchForValidMeasurementUnits(ctx, x.Query, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &SearchValidMeasurementUnitsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
