package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v11/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidIngredientStateInvocation struct {
		ValidIngredientStateID string `jsonschema:"description=The ingredient state ID"`
	}
)

var validIngredientStatesSchema = map[string]any{
	"ID":            stringField("The ID of the valid ingredient state"),
	"CreatedAt":     timestampField("When the valid ingredient state was created"),
	"LastUpdatedAt": timestampField("When the valid ingredient state was last updated"),
	"ArchivedAt":    timestampField("When the valid ingredient state was soft deleted"),
	"Name":          stringField("Name of the ingredient state"),
	"Description":   stringField("Description of the ingredient state"),
	"IconPath":      stringField("The URL for the icon for the item"),
	"Slug":          stringField("An easy-to-use URL slug for the ingredient state"),
	"PastTense":     stringField("The past tense form of the ingredient state name (e.g., 'chopped' for 'chop')"),
	"AttributeType": stringField("The attribute type of the ingredient state (texture, consistency, temperature, color, appearance, odor, taste, sound, or other)"),
}

var getValidIngredientStateTool = &mcp.Tool{
	Name:        "GetValidIngredientState",
	Description: "Get a valid ingredient state by it's ID",
	InputSchema: schemaObject(map[string]any{
		"ValidIngredientStateID": stringField("The ID of the valid ingredient state to get"),
	}),
	OutputSchema: schemaObject(validIngredientStatesSchema),
}

func (h *mcpToolManager) GetValidIngredientState() mcp.ToolHandlerFor[*GetValidIngredientStateInvocation, *mealplanning.ValidIngredientState] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidIngredientStateInvocation) (*mcp.CallToolResult, *mealplanning.ValidIngredientState, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetValidIngredientState(ctx, x.ValidIngredientStateID)
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

var searchForValidIngredientStatesTool = &mcp.Tool{
	Name:        "SearchForValidIngredientStates",
	Description: "Search for valid ingredient states with optional filtering and query string",
	InputSchema: schemaObject(map[string]any{
		"Filter": filtering.QueryFilterSchema(),
		"Query": map[string]any{
			"type":        strType,
			"description": "The ingredient state name query",
		},
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(validIngredientStatesSchema)),
	}),
}

func (h *mcpToolManager) SearchForValidIngredientStates() mcp.ToolHandlerFor[*SearchValidIngredientStatesInvocation, *SearchValidIngredientStatesResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *SearchValidIngredientStatesInvocation) (*mcp.CallToolResult, *SearchValidIngredientStatesResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.SearchForValidIngredientStates(ctx, x.Query, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &SearchValidIngredientStatesResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
