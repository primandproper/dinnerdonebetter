package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v13/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidMeasurementUnitInvocation struct {
		ValidMeasurementUnitID string `jsonschema:"description=The measurement unit ID"`
	}
)

var validMeasurementUnitsSchema = map[string]any{
	"ID":               stringField("The ID of the valid measurement unit"),
	fieldCreatedAt:     timestampField("When the valid measurement unit was created"),
	fieldLastUpdatedAt: timestampField("When the valid measurement unit was last updated"),
	fieldArchivedAt:    timestampField("When the valid measurement unit was soft deleted"),
	fieldName:          stringField("Name of the measurement unit"),
	fieldDescription:   stringField("Description of the measurement unit"),
	fieldIconPath:      stringField("The URL for the icon for the item"),
	fieldPluralName:    stringField("The plural name for the measurement unit. So for a unit named 'cup', this would be 'cups'"),
	fieldSlug:          stringField("An easy-to-use URL slug for the measurement unit"),
	"Volumetric":       boolField("Whether or not the valid measurement unit is volumetric"),
	"Universal":        boolField("Whether or not the valid measurement unit is universal (valid for all ingredients). For instance, 'grams' is a universal measurement unit"),
	"Metric":           boolField("Whether or not the valid measurement unit is metric"),
	"Imperial":         boolField("Whether or not the valid measurement unit is imperial"),
}

var getValidMeasurementUnitTool = &mcp.Tool{
	Name:        "GetValidMeasurementUnit",
	Description: "Get a valid measurement unit by it's ID",
	InputSchema: schemaObject(map[string]any{
		"ValidMeasurementUnitID": stringField("The ID of the valid measurement unit to get"),
	}),
	OutputSchema: schemaObject(validMeasurementUnitsSchema),
}

func (h *mcpToolManager) GetValidMeasurementUnit() mcp.ToolHandlerFor[*GetValidMeasurementUnitInvocation, *mealplanning.ValidMeasurementUnit] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidMeasurementUnitInvocation) (*mcp.CallToolResult, *mealplanning.ValidMeasurementUnit, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetValidMeasurementUnit(ctx, x.ValidMeasurementUnitID)
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

var searchForValidMeasurementUnitsTool = &mcp.Tool{
	Name:        "SearchForValidMeasurementUnits",
	Description: "Search for valid measurement units with optional filtering and query string",
	InputSchema: schemaObject(map[string]any{
		fieldFilter: filtering.QueryFilterSchema(),
		fieldQuery: map[string]any{
			keyType:        strType,
			keyDescription: "The measurement unit name query",
		},
	}),
	OutputSchema: schemaObject(map[string]any{
		fieldResults: arrayType(schemaObject(validMeasurementUnitsSchema)),
	}),
}

func (h *mcpToolManager) SearchForValidMeasurementUnits() mcp.ToolHandlerFor[*SearchValidMeasurementUnitsInvocation, *SearchValidMeasurementUnitsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *SearchValidMeasurementUnitsInvocation) (*mcp.CallToolResult, *SearchValidMeasurementUnitsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.SearchForValidMeasurementUnits(ctx, x.Query, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &SearchValidMeasurementUnitsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
