package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v10/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidPreparationInvocation struct {
		ValidPreparationID string `jsonschema:"description=The preparation ID"`
	}
)

var validPreparationsSchema = map[string]any{
	"ID":                          stringField("The ID of the valid preparation"),
	"CreatedAt":                   timestampField("When the valid preparation was created"),
	"LastUpdatedAt":               timestampField("When the valid preparation was last updated"),
	"ArchivedAt":                  timestampField("When the valid preparation was soft deleted"),
	"Name":                        stringField("Name of the preparation"),
	"Description":                 stringField("Description of the preparation"),
	"IconPath":                    stringField("The URL for the icon for the item"),
	"Slug":                        stringField("An easy-to-use URL slug for the preparation"),
	"PastTense":                   stringField("The past tense form of the preparation name (e.g., 'chopped' for 'chop')"),
	"MinInstrumentCount":          uintField("Minimum number of instruments required"),
	"MaxInstrumentCount":          uintField("Maximum number of instruments allowed (optional)"),
	"MinIngredientCount":          uintField("Minimum number of ingredients required"),
	"MaxIngredientCount":          uintField("Maximum number of ingredients allowed (optional)"),
	"MinVesselCount":              uintField("Minimum number of vessels required"),
	"MaxVesselCount":              uintField("Maximum number of vessels allowed (optional)"),
	"RestrictToIngredients":       boolField("Whether or not the valid preparation is restricted to ingredients"),
	"TemperatureRequired":         boolField("Whether or not the valid preparation requires a temperature"),
	"TimeEstimateRequired":        boolField("Whether or not the valid preparation requires a time estimate"),
	"ConditionExpressionRequired": boolField("Whether or not the valid preparation requires a condition expression"),
	"ConsumesVessel":              boolField("Whether or not the valid preparation consumes a vessel"),
	"OnlyForVessels":              boolField("Whether or not the valid preparation is only for vessels"),
	"YieldsNothing":               boolField("Whether or not the valid preparation yields nothing"),
}

var getValidPreparationTool = &mcp.Tool{
	Name:        "GetValidPreparation",
	Description: "Get a valid preparation by it's ID",
	InputSchema: schemaObject(map[string]any{
		"ValidPreparationID": stringField("The ID of the valid preparation to get"),
	}),
	OutputSchema: schemaObject(validPreparationsSchema),
}

func (h *mcpToolManager) GetValidPreparation() mcp.ToolHandlerFor[*GetValidPreparationInvocation, *mealplanning.ValidPreparation] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidPreparationInvocation) (*mcp.CallToolResult, *mealplanning.ValidPreparation, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetValidPreparation(ctx, x.ValidPreparationID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	SearchValidPreparationsInvocation struct {
		Filter *filtering.QueryFilter
		Query  string `jsonschema_description:"The preparation name query"`
	}

	SearchValidPreparationsResult struct {
		Results []*mealplanning.ValidPreparation
	}
)

var searchForValidPreparationsTool = &mcp.Tool{
	Name:        "SearchForValidPreparations",
	Description: "Search for valid preparations with optional filtering and query string",
	InputSchema: schemaObject(map[string]any{
		"Filter": queryFilterSchema(),
		"Query": map[string]any{
			"type":        strType,
			"description": "The preparation name query",
		},
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(validPreparationsSchema)),
	}),
}

func (h *mcpToolManager) SearchForValidPreparations() mcp.ToolHandlerFor[*SearchValidPreparationsInvocation, *SearchValidPreparationsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *SearchValidPreparationsInvocation) (*mcp.CallToolResult, *SearchValidPreparationsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.SearchForValidPreparations(ctx, x.Query, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &SearchValidPreparationsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
