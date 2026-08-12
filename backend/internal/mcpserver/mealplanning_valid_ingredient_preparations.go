package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v10/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidIngredientPreparationInvocation struct {
		ValidIngredientPreparationID string `jsonschema:"description=The ingredient preparation ID"`
	}
)

var validIngredientPreparationsSchema = map[string]any{
	"ID":            stringField("The ID of the valid ingredient preparation"),
	"CreatedAt":     timestampField("When the valid ingredient preparation was created"),
	"LastUpdatedAt": timestampField("When the valid ingredient preparation was last updated"),
	"ArchivedAt":    timestampField("When the valid ingredient preparation was soft deleted"),
	"Notes":         stringField("Notes about the ingredient preparation"),
	"Preparation":   objectType(validPreparationsSchema),
	"Ingredient":    objectType(validIngredientsSchema),
}

var getValidIngredientPreparationTool = &mcp.Tool{
	Name:        "GetValidIngredientPreparation",
	Description: "Get a valid ingredient preparation by it's ID",
	InputSchema: schemaObject(map[string]any{
		"ValidIngredientPreparationID": stringField("The ID of the valid ingredient preparation to get"),
	}),
	OutputSchema: schemaObject(validIngredientPreparationsSchema),
}

func (h *mcpToolManager) GetValidIngredientPreparation() mcp.ToolHandlerFor[*GetValidIngredientPreparationInvocation, *mealplanning.ValidIngredientPreparation] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidIngredientPreparationInvocation) (*mcp.CallToolResult, *mealplanning.ValidIngredientPreparation, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetValidIngredientPreparation(ctx, x.ValidIngredientPreparationID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetValidIngredientPreparationsInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetValidIngredientPreparationsResult struct {
		Results []*mealplanning.ValidIngredientPreparation
	}
)

var getValidIngredientPreparationsTool = &mcp.Tool{
	Name:        "GetValidIngredientPreparations",
	Description: "Get valid ingredient preparations with optional filtering",
	InputSchema: schemaObject(map[string]any{
		"Filter": queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(validIngredientPreparationsSchema)),
	}),
}

func (h *mcpToolManager) GetValidIngredientPreparations() mcp.ToolHandlerFor[*GetValidIngredientPreparationsInvocation, *GetValidIngredientPreparationsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidIngredientPreparationsInvocation) (*mcp.CallToolResult, *GetValidIngredientPreparationsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetValidIngredientPreparations(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidIngredientPreparationsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
