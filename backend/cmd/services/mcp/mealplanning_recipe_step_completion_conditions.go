package main

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v7/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipeStepCompletionConditionInvocation struct {
		RecipeID                        string `jsonschema:"description=The recipe ID"`
		RecipeStepID                    string `jsonschema:"description=The recipe step ID"`
		RecipeStepCompletionConditionID string `jsonschema:"description=The recipe step completion condition ID"`
	}
)

var recipeStepCompletionConditionIngredientSchema = map[string]any{
	"ID":                                     stringField("The ID of the recipe step completion condition ingredient"),
	"CreatedAt":                              timestampField("When the recipe step completion condition ingredient was created"),
	"LastUpdatedAt":                          timestampField("When the recipe step completion condition ingredient was last updated"),
	"ArchivedAt":                             timestampField("When the recipe step completion condition ingredient was soft deleted"),
	"BelongsToRecipeStepCompletionCondition": stringField("The ID of the recipe step completion condition this ingredient belongs to"),
	"RecipeStepIngredient":                   stringField("The ID of the recipe step ingredient"),
}

var recipeStepCompletionConditionsSchema = map[string]any{
	"ID":                  stringField("The ID of the recipe step completion condition"),
	"CreatedAt":           timestampField("When the recipe step completion condition was created"),
	"LastUpdatedAt":       timestampField("When the recipe step completion condition was last updated"),
	"ArchivedAt":          timestampField("When the recipe step completion condition was soft deleted"),
	"BelongsToRecipeStep": stringField("The ID of the recipe step this completion condition belongs to"),
	"IngredientState":     objectType(validIngredientStatesSchema),
	"Notes":               stringField("Notes about the completion condition"),
	"Ingredients":         arrayType(schemaObject(recipeStepCompletionConditionIngredientSchema)),
	"Optional":            boolField("Whether this completion condition is optional"),
}

var getRecipeStepCompletionConditionTool = &mcp.Tool{
	Name:        "GetRecipeStepCompletionCondition",
	Description: "Get a recipe step completion condition by it's ID",
	InputSchema: schemaObject(map[string]any{
		"RecipeID":                        stringField("The ID of the recipe"),
		"RecipeStepID":                    stringField("The ID of the recipe step"),
		"RecipeStepCompletionConditionID": stringField("The ID of the recipe step completion condition to get"),
	}),
	OutputSchema: schemaObject(recipeStepCompletionConditionsSchema),
}

func (h *mcpToolManager) GetRecipeStepCompletionCondition() mcp.ToolHandlerFor[*GetRecipeStepCompletionConditionInvocation, *mealplanning.RecipeStepCompletionCondition] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipeStepCompletionConditionInvocation) (*mcp.CallToolResult, *mealplanning.RecipeStepCompletionCondition, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetRecipeStepCompletionCondition(ctx, x.RecipeID, x.RecipeStepID, x.RecipeStepCompletionConditionID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetRecipeStepCompletionConditionsInvocation struct {
		Filter       *filtering.QueryFilter
		RecipeID     string
		RecipeStepID string
	}

	GetRecipeStepCompletionConditionsResult struct {
		Results []*mealplanning.RecipeStepCompletionCondition
	}
)

var getRecipeStepCompletionConditionsTool = &mcp.Tool{
	Name:        "GetRecipeStepCompletionConditions",
	Description: "Get recipe step completion conditions with optional filtering",
	InputSchema: schemaObject(map[string]any{
		"RecipeID":     stringField("The ID of the recipe"),
		"RecipeStepID": stringField("The ID of the recipe step"),
		"Filter":       queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(recipeStepCompletionConditionsSchema)),
	}),
}

func (h *mcpToolManager) GetRecipeStepCompletionConditions() mcp.ToolHandlerFor[*GetRecipeStepCompletionConditionsInvocation, *GetRecipeStepCompletionConditionsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipeStepCompletionConditionsInvocation) (*mcp.CallToolResult, *GetRecipeStepCompletionConditionsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetRecipeStepCompletionConditions(ctx, x.RecipeID, x.RecipeStepID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipeStepCompletionConditionsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
