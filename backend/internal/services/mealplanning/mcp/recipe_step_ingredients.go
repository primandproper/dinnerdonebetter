package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetRecipeStepIngredientInput is what GetRecipeStepIngredient takes.
type GetRecipeStepIngredientInput struct {
	RecipeID               string `json:"recipeID"               jsonschema:"The identifier of the recipe"`
	RecipeStepID           string `json:"recipeStepID"           jsonschema:"The identifier of the step"`
	RecipeStepIngredientID string `json:"recipeStepIngredientID" jsonschema:"The identifier of the ingredient to read"`
}

var getRecipeStepIngredientTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepIngredient",
	Description: "Get a ingredient of a recipe step by its ID",
	Annotations: readOnly(),
}

// GetRecipeStepIngredient reads one ingredient of a recipe step.
func (t *Tools) GetRecipeStepIngredient(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipeStepIngredientInput) (*sdkmcp.CallToolResult, *mealplanning.RecipeStepIngredient, error) {
	ctx, err := t.begin(ctx, req, getRecipeStepIngredientTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadRecipeStepIngredient(ctx, in.RecipeID, in.RecipeStepID, in.RecipeStepIngredientID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

// GetRecipeStepIngredientsInput is what GetRecipeStepIngredients takes.
type GetRecipeStepIngredientsInput struct {
	Filter       *filtering.QueryFilter `json:"filter,omitempty" jsonschema:"The page to read; absent reads the first page, oldest first"`
	RecipeID     string                 `json:"recipeID"         jsonschema:"The identifier of the recipe"`
	RecipeStepID string                 `json:"recipeStepID"     jsonschema:"The identifier of the step"`
}

var getRecipeStepIngredientsTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepIngredients",
	Description: "Page the ingredients of a recipe step",
	Annotations: readOnly(),
}

// GetRecipeStepIngredients pages the ingredients of a recipe step.
func (t *Tools) GetRecipeStepIngredients(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipeStepIngredientsInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.RecipeStepIngredient], error) {
	ctx, err := t.begin(ctx, req, getRecipeStepIngredientsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ListRecipeStepIngredients(ctx, in.RecipeID, in.RecipeStepID, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
