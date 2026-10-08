package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipeStepCompletionConditionInvocation struct {
		RecipeID                        string `jsonschema:"description=The recipe ID"`
		RecipeStepID                    string `jsonschema:"description=The recipe step ID"`
		RecipeStepCompletionConditionID string `jsonschema:"description=The recipe step completion condition ID"`
	}
)

var recipeStepCompletionConditionIngredientSchema = map[string]any{
	"ID":                                     mcptools.StringField("The ID of the recipe step completion condition ingredient"),
	mcptools.FieldCreatedAt:                  mcptools.TimestampField("When the recipe step completion condition ingredient was created"),
	mcptools.FieldLastUpdatedAt:              mcptools.TimestampField("When the recipe step completion condition ingredient was last updated"),
	mcptools.FieldArchivedAt:                 mcptools.TimestampField("When the recipe step completion condition ingredient was soft deleted"),
	"BelongsToRecipeStepCompletionCondition": mcptools.StringField("The ID of the recipe step completion condition this ingredient belongs to"),
	"RecipeStepIngredient":                   mcptools.StringField("The ID of the recipe step ingredient"),
}

var recipeStepCompletionConditionsSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the recipe step completion condition"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the recipe step completion condition was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the recipe step completion condition was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the recipe step completion condition was soft deleted"),
	fieldBelongsToRecipeStep:    mcptools.StringField("The ID of the recipe step this completion condition belongs to"),
	"IngredientState":           mcptools.ObjectType(validIngredientStatesSchema),
	fieldNotes:                  mcptools.StringField("Notes about the completion condition"),
	"Ingredients":               mcptools.ArrayType(mcptools.SchemaObject(recipeStepCompletionConditionIngredientSchema)),
	fieldOptional:               mcptools.BoolField("Whether this completion condition is optional"),
}

var getRecipeStepCompletionConditionTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepCompletionCondition",
	Description: "Get a recipe step completion condition by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID:                     mcptools.StringField("The ID of the recipe"),
		fieldRecipeStepID:                 mcptools.StringField("The ID of the recipe step"),
		"RecipeStepCompletionConditionID": mcptools.StringField("The ID of the recipe step completion condition to get"),
	}),
	OutputSchema: mcptools.SchemaObject(recipeStepCompletionConditionsSchema),
}

func (t *Tools) GetRecipeStepCompletionCondition() sdkmcp.ToolHandlerFor[*GetRecipeStepCompletionConditionInvocation, *mealplanning.RecipeStepCompletionCondition] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipeStepCompletionConditionInvocation) (*sdkmcp.CallToolResult, *mealplanning.RecipeStepCompletionCondition, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetRecipeStepCompletionCondition(ctx, x.RecipeID, x.RecipeStepID, x.RecipeStepCompletionConditionID)
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

var getRecipeStepCompletionConditionsTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepCompletionConditions",
	Description: "Get recipe step completion conditions with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID:        mcptools.StringField("The ID of the recipe"),
		fieldRecipeStepID:    mcptools.StringField("The ID of the recipe step"),
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(recipeStepCompletionConditionsSchema)),
	}),
}

func (t *Tools) GetRecipeStepCompletionConditions() sdkmcp.ToolHandlerFor[*GetRecipeStepCompletionConditionsInvocation, *GetRecipeStepCompletionConditionsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipeStepCompletionConditionsInvocation) (*sdkmcp.CallToolResult, *GetRecipeStepCompletionConditionsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetRecipeStepCompletionConditions(ctx, x.RecipeID, x.RecipeStepID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipeStepCompletionConditionsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
