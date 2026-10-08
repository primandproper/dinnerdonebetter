package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetRecipeStepCompletionConditionInput is what GetRecipeStepCompletionCondition takes.
type GetRecipeStepCompletionConditionInput struct {
	RecipeID                        string `json:"recipeID"                        jsonschema:"The identifier of the recipe"`
	RecipeStepID                    string `json:"recipeStepID"                    jsonschema:"The identifier of the step"`
	RecipeStepCompletionConditionID string `json:"recipeStepCompletionConditionID" jsonschema:"The identifier of the completion condition to read"`
}

var getRecipeStepCompletionConditionTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepCompletionCondition",
	Description: "Get a completion condition of a recipe step by its ID",
	Annotations: readOnly(),
}

// GetRecipeStepCompletionCondition reads one completion condition of a recipe step.
func (t *Tools) GetRecipeStepCompletionCondition(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipeStepCompletionConditionInput) (*sdkmcp.CallToolResult, *mealplanning.RecipeStepCompletionCondition, error) {
	ctx, err := t.begin(ctx, req, getRecipeStepCompletionConditionTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadRecipeStepCompletionCondition(ctx, in.RecipeID, in.RecipeStepID, in.RecipeStepCompletionConditionID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

// GetRecipeStepCompletionConditionsInput is what GetRecipeStepCompletionConditions takes.
type GetRecipeStepCompletionConditionsInput struct {
	Filter       *filtering.QueryFilter `json:"filter,omitempty" jsonschema:"The page to read; absent reads the first page, oldest first"`
	RecipeID     string                 `json:"recipeID"         jsonschema:"The identifier of the recipe"`
	RecipeStepID string                 `json:"recipeStepID"     jsonschema:"The identifier of the step"`
}

var getRecipeStepCompletionConditionsTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepCompletionConditions",
	Description: "Page the completion conditions of a recipe step",
	Annotations: readOnly(),
}

// GetRecipeStepCompletionConditions pages the completion conditions of a recipe step.
func (t *Tools) GetRecipeStepCompletionConditions(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipeStepCompletionConditionsInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.RecipeStepCompletionCondition], error) {
	ctx, err := t.begin(ctx, req, getRecipeStepCompletionConditionsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ListRecipeStepCompletionConditions(ctx, in.RecipeID, in.RecipeStepID, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
