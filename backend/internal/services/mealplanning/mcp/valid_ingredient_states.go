package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidIngredientStateInvocation struct {
		ValidIngredientStateID string `jsonschema:"description=The ingredient state ID"`
	}
)

var validIngredientStatesSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the valid ingredient state"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the valid ingredient state was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the valid ingredient state was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the valid ingredient state was soft deleted"),
	mcptools.FieldName:          mcptools.StringField("Name of the ingredient state"),
	mcptools.FieldDescription:   mcptools.StringField("Description of the ingredient state"),
	fieldIconPath:               mcptools.StringField("The URL for the icon for the item"),
	fieldSlug:                   mcptools.StringField("An easy-to-use URL slug for the ingredient state"),
	"PastTense":                 mcptools.StringField("The past tense form of the ingredient state name (e.g., 'chopped' for 'chop')"),
	"AttributeType":             mcptools.StringField("The attribute type of the ingredient state (texture, consistency, temperature, color, appearance, odor, taste, sound, or other)"),
}

var getValidIngredientStateTool = &sdkmcp.Tool{
	Name:        "GetValidIngredientState",
	Description: "Get a valid ingredient state by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		"ValidIngredientStateID": mcptools.StringField("The ID of the valid ingredient state to get"),
	}),
	OutputSchema: mcptools.SchemaObject(validIngredientStatesSchema),
}

func (t *Tools) GetValidIngredientState() sdkmcp.ToolHandlerFor[*GetValidIngredientStateInvocation, *mealplanning.ValidIngredientState] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidIngredientStateInvocation) (*sdkmcp.CallToolResult, *mealplanning.ValidIngredientState, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetValidIngredientState(ctx, x.ValidIngredientStateID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	SearchValidIngredientStatesInvocation struct {
		Filter *filtering.QueryFilter
		Query  string `jsonschema_description:"The ingredient state name query"`
	}

	SearchValidIngredientStatesResult struct {
		Results []*mealplanning.ValidIngredientState
	}
)

var searchForValidIngredientStatesTool = &sdkmcp.Tool{
	Name:        "SearchForValidIngredientStates",
	Description: "Search for valid ingredient states with optional filtering and query string",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
		mcptools.FieldQuery:  mcptools.StringField("The ingredient state name query"),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validIngredientStatesSchema)),
	}),
}

func (t *Tools) SearchForValidIngredientStates() sdkmcp.ToolHandlerFor[*SearchValidIngredientStatesInvocation, *SearchValidIngredientStatesResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *SearchValidIngredientStatesInvocation) (*sdkmcp.CallToolResult, *SearchValidIngredientStatesResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.SearchForValidIngredientStates(ctx, x.Query, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &SearchValidIngredientStatesResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
