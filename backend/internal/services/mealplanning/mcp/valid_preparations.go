package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidPreparationInvocation struct {
		ValidPreparationID string `jsonschema:"description=The preparation ID"`
	}
)

var validPreparationsSchema = map[string]any{
	"ID":                          mcptools.StringField("The ID of the valid preparation"),
	mcptools.FieldCreatedAt:       mcptools.TimestampField("When the valid preparation was created"),
	mcptools.FieldLastUpdatedAt:   mcptools.TimestampField("When the valid preparation was last updated"),
	mcptools.FieldArchivedAt:      mcptools.TimestampField("When the valid preparation was soft deleted"),
	mcptools.FieldName:            mcptools.StringField("Name of the preparation"),
	mcptools.FieldDescription:     mcptools.StringField("Description of the preparation"),
	fieldIconPath:                 mcptools.StringField("The URL for the icon for the item"),
	fieldSlug:                     mcptools.StringField("An easy-to-use URL slug for the preparation"),
	"PastTense":                   mcptools.StringField("The past tense form of the preparation name (e.g., 'chopped' for 'chop')"),
	"MinInstrumentCount":          mcptools.UintField("Minimum number of instruments required"),
	"MaxInstrumentCount":          mcptools.UintField("Maximum number of instruments allowed (optional)"),
	"MinIngredientCount":          mcptools.UintField("Minimum number of ingredients required"),
	"MaxIngredientCount":          mcptools.UintField("Maximum number of ingredients allowed (optional)"),
	"MinVesselCount":              mcptools.UintField("Minimum number of vessels required"),
	"MaxVesselCount":              mcptools.UintField("Maximum number of vessels allowed (optional)"),
	"RestrictToIngredients":       mcptools.BoolField("Whether or not the valid preparation is restricted to ingredients"),
	"TemperatureRequired":         mcptools.BoolField("Whether or not the valid preparation requires a temperature"),
	"TimeEstimateRequired":        mcptools.BoolField("Whether or not the valid preparation requires a time estimate"),
	"ConditionExpressionRequired": mcptools.BoolField("Whether or not the valid preparation requires a condition expression"),
	"ConsumesVessel":              mcptools.BoolField("Whether or not the valid preparation consumes a vessel"),
	"OnlyForVessels":              mcptools.BoolField("Whether or not the valid preparation is only for vessels"),
	"YieldsNothing":               mcptools.BoolField("Whether or not the valid preparation yields nothing"),
}

var getValidPreparationTool = &sdkmcp.Tool{
	Name:        "GetValidPreparation",
	Description: "Get a valid preparation by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldValidPreparationID: mcptools.StringField("The ID of the valid preparation to get"),
	}),
	OutputSchema: mcptools.SchemaObject(validPreparationsSchema),
}

func (t *Tools) GetValidPreparation() sdkmcp.ToolHandlerFor[*GetValidPreparationInvocation, *mealplanning.ValidPreparation] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidPreparationInvocation) (*sdkmcp.CallToolResult, *mealplanning.ValidPreparation, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetValidPreparation(ctx, x.ValidPreparationID)
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

var searchForValidPreparationsTool = &sdkmcp.Tool{
	Name:        "SearchForValidPreparations",
	Description: "Search for valid preparations with optional filtering and query string",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
		mcptools.FieldQuery:  mcptools.StringField("The preparation name query"),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validPreparationsSchema)),
	}),
}

func (t *Tools) SearchForValidPreparations() sdkmcp.ToolHandlerFor[*SearchValidPreparationsInvocation, *SearchValidPreparationsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *SearchValidPreparationsInvocation) (*sdkmcp.CallToolResult, *SearchValidPreparationsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.SearchForValidPreparations(ctx, x.Query, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &SearchValidPreparationsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
