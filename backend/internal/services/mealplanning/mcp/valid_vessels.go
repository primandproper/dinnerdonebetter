package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidVesselInvocation struct {
		ValidVesselID string `jsonschema:"description=The vessel ID"`
	}
)

var validVesselsSchema = map[string]any{
	"ID":                             mcptools.StringField("The ID of the valid vessel"),
	mcptools.FieldCreatedAt:          mcptools.TimestampField("When the valid vessel was created"),
	mcptools.FieldLastUpdatedAt:      mcptools.TimestampField("When the valid vessel was last updated"),
	mcptools.FieldArchivedAt:         mcptools.TimestampField("When the valid vessel was soft deleted"),
	mcptools.FieldName:               mcptools.StringField("Name of the vessel"),
	mcptools.FieldDescription:        mcptools.StringField("Description of the vessel"),
	fieldIconPath:                    mcptools.StringField("The URL for the icon for the item"),
	fieldPluralName:                  mcptools.StringField("The plural name for the vessel. So for a vessel named 'pan', this would be 'pans'"),
	fieldSlug:                        mcptools.StringField("An easy-to-use URL slug for the vessel"),
	"Shape":                          mcptools.StringField("The shape of the vessel (hemisphere, rectangle, cone, pyramid, cylinder, sphere, cube, or other)"),
	"WidthInMillimeters":             mcptools.FloatField("Width of the vessel in millimeters"),
	"LengthInMillimeters":            mcptools.FloatField("Length of the vessel in millimeters"),
	"HeightInMillimeters":            mcptools.FloatField("Height of the vessel in millimeters"),
	"Capacity":                       mcptools.FloatField("Capacity of the vessel"),
	"IncludeInGeneratedInstructions": mcptools.BoolField("Whether or not the valid vessel should be included in generated instructions"),
	"DisplayInSummaryLists":          mcptools.BoolField("Whether or not the valid vessel should be displayed in summary lists"),
	"UsableForStorage":               mcptools.BoolField("Whether or not the valid vessel is usable for storage"),
	"CapacityUnit":                   mcptools.ObjectType(validMeasurementUnitsSchema),
}

var getValidVesselTool = &sdkmcp.Tool{
	Name:        "GetValidVessel",
	Description: "Get a valid vessel by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		"ValidVesselID": mcptools.StringField("The ID of the valid vessel to get"),
	}),
	OutputSchema: mcptools.SchemaObject(validVesselsSchema),
}

func (t *Tools) GetValidVessel() sdkmcp.ToolHandlerFor[*GetValidVesselInvocation, *mealplanning.ValidVessel] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidVesselInvocation) (*sdkmcp.CallToolResult, *mealplanning.ValidVessel, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetValidVessel(ctx, x.ValidVesselID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	SearchValidVesselsInvocation struct {
		Filter *filtering.QueryFilter
		Query  string `jsonschema_description:"The vessel name query"`
	}

	SearchValidVesselsResult struct {
		Results []*mealplanning.ValidVessel
	}
)

var searchForValidVesselsTool = &sdkmcp.Tool{
	Name:        "SearchForValidVessels",
	Description: "Search for valid vessels with optional filtering and query string",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
		mcptools.FieldQuery:  mcptools.StringField("The vessel name query"),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validVesselsSchema)),
	}),
}

func (t *Tools) SearchForValidVessels() sdkmcp.ToolHandlerFor[*SearchValidVesselsInvocation, *SearchValidVesselsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *SearchValidVesselsInvocation) (*sdkmcp.CallToolResult, *SearchValidVesselsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.SearchForValidVessels(ctx, x.Query, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &SearchValidVesselsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
