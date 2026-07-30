package main

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v8/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidVesselInvocation struct {
		ValidVesselID string `jsonschema:"description=The vessel ID"`
	}
)

var validVesselsSchema = map[string]any{
	"ID":                             stringField("The ID of the valid vessel"),
	"CreatedAt":                      timestampField("When the valid vessel was created"),
	"LastUpdatedAt":                  timestampField("When the valid vessel was last updated"),
	"ArchivedAt":                     timestampField("When the valid vessel was soft deleted"),
	"Name":                           stringField("Name of the vessel"),
	"Description":                    stringField("Description of the vessel"),
	"IconPath":                       stringField("The URL for the icon for the item"),
	"PluralName":                     stringField("The plural name for the vessel. So for a vessel named 'pan', this would be 'pans'"),
	"Slug":                           stringField("An easy-to-use URL slug for the vessel"),
	"Shape":                          stringField("The shape of the vessel (hemisphere, rectangle, cone, pyramid, cylinder, sphere, cube, or other)"),
	"WidthInMillimeters":             floatField("Width of the vessel in millimeters"),
	"LengthInMillimeters":            floatField("Length of the vessel in millimeters"),
	"HeightInMillimeters":            floatField("Height of the vessel in millimeters"),
	"Capacity":                       floatField("Capacity of the vessel"),
	"IncludeInGeneratedInstructions": boolField("Whether or not the valid vessel should be included in generated instructions"),
	"DisplayInSummaryLists":          boolField("Whether or not the valid vessel should be displayed in summary lists"),
	"UsableForStorage":               boolField("Whether or not the valid vessel is usable for storage"),
	"CapacityUnit":                   objectType(validMeasurementUnitsSchema),
}

var getValidVesselTool = &mcp.Tool{
	Name:        "GetValidVessel",
	Description: "Get a valid vessel by it's ID",
	InputSchema: schemaObject(map[string]any{
		"ValidVesselID": stringField("The ID of the valid vessel to get"),
	}),
	OutputSchema: schemaObject(validVesselsSchema),
}

func (h *mcpToolManager) GetValidVessel() mcp.ToolHandlerFor[*GetValidVesselInvocation, *mealplanning.ValidVessel] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidVesselInvocation) (*mcp.CallToolResult, *mealplanning.ValidVessel, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetValidVessel(ctx, x.ValidVesselID)
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

var searchForValidVesselsTool = &mcp.Tool{
	Name:        "SearchForValidVessels",
	Description: "Search for valid vessels with optional filtering and query string",
	InputSchema: schemaObject(map[string]any{
		"Filter": queryFilterSchema(),
		"Query": map[string]any{
			"type":        strType,
			"description": "The vessel name query",
		},
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(validVesselsSchema)),
	}),
}

func (h *mcpToolManager) SearchForValidVessels() mcp.ToolHandlerFor[*SearchValidVesselsInvocation, *SearchValidVesselsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *SearchValidVesselsInvocation) (*mcp.CallToolResult, *SearchValidVesselsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.SearchForValidVessels(ctx, x.Query, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &SearchValidVesselsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
