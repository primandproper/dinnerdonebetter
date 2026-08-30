package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v13/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidIngredientStateIngredientInvocation struct {
		ValidIngredientStateIngredientID string `jsonschema:"description=The ingredient state ingredient ID"`
	}
)

var validIngredientStateIngredientsSchema = map[string]any{
	"ID":               stringField("The ID of the valid ingredient state ingredient"),
	fieldCreatedAt:     timestampField("When the valid ingredient state ingredient was created"),
	fieldLastUpdatedAt: timestampField("When the valid ingredient state ingredient was last updated"),
	fieldArchivedAt:    timestampField("When the valid ingredient state ingredient was soft deleted"),
	fieldNotes:         stringField("Notes about the ingredient state ingredient"),
	"IngredientState":  objectType(validIngredientStatesSchema),
	fieldIngredient:    objectType(validIngredientsSchema),
}

var getValidIngredientStateIngredientTool = &mcp.Tool{
	Name:        "GetValidIngredientStateIngredient",
	Description: "Get a valid ingredient state ingredient by it's ID",
	InputSchema: schemaObject(map[string]any{
		"ValidIngredientStateIngredientID": stringField("The ID of the valid ingredient state ingredient to get"),
	}),
	OutputSchema: schemaObject(validIngredientStateIngredientsSchema),
}

func (h *mcpToolManager) GetValidIngredientStateIngredient() mcp.ToolHandlerFor[*GetValidIngredientStateIngredientInvocation, *mealplanning.ValidIngredientStateIngredient] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidIngredientStateIngredientInvocation) (*mcp.CallToolResult, *mealplanning.ValidIngredientStateIngredient, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetValidIngredientStateIngredient(ctx, x.ValidIngredientStateIngredientID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetValidIngredientStateIngredientsInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetValidIngredientStateIngredientsResult struct {
		Results []*mealplanning.ValidIngredientStateIngredient
	}
)

var getValidIngredientStateIngredientsTool = &mcp.Tool{
	Name:        "GetValidIngredientStateIngredients",
	Description: "Get valid ingredient state ingredients with optional filtering",
	InputSchema: schemaObject(map[string]any{
		fieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		fieldResults: arrayType(schemaObject(validIngredientStateIngredientsSchema)),
	}),
}

func (h *mcpToolManager) GetValidIngredientStateIngredients() mcp.ToolHandlerFor[*GetValidIngredientStateIngredientsInvocation, *GetValidIngredientStateIngredientsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidIngredientStateIngredientsInvocation) (*mcp.CallToolResult, *GetValidIngredientStateIngredientsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetValidIngredientStateIngredients(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidIngredientStateIngredientsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
