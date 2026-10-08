package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetRecipeStepInput is what GetRecipeStep takes.
type GetRecipeStepInput struct {
	RecipeID     string `json:"recipeID"     jsonschema:"The identifier of the recipe"`
	RecipeStepID string `json:"recipeStepID" jsonschema:"The identifier of the step to read"`
}

var getRecipeStepTool = &sdkmcp.Tool{
	Name:        "GetRecipeStep",
	Description: "Get a recipe step by its ID",
	Annotations: readOnly(),
}

// GetRecipeStep reads one step of a recipe.
func (t *Tools) GetRecipeStep(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipeStepInput) (*sdkmcp.CallToolResult, *mealplanning.RecipeStep, error) {
	ctx, err := t.begin(ctx, req, getRecipeStepTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadRecipeStep(ctx, in.RecipeID, in.RecipeStepID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

// GetRecipeStepsInput is what GetRecipeSteps takes.
type GetRecipeStepsInput struct {
	Filter   *filtering.QueryFilter `json:"filter,omitempty" jsonschema:"The page to read; absent reads the first page, oldest first"`
	RecipeID string                 `json:"recipeID"         jsonschema:"The identifier of the recipe"`
}

var getRecipeStepsTool = &sdkmcp.Tool{
	Name:        "GetRecipeSteps",
	Description: "Page the steps of a recipe",
	Annotations: readOnly(),
}

// GetRecipeSteps pages the steps of a recipe.
func (t *Tools) GetRecipeSteps(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipeStepsInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.RecipeStep], error) {
	ctx, err := t.begin(ctx, req, getRecipeStepsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ListRecipeSteps(ctx, in.RecipeID, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
