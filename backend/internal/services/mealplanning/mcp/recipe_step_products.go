package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetRecipeStepProductInput is what GetRecipeStepProduct takes.
type GetRecipeStepProductInput struct {
	RecipeID            string `json:"recipeID"            jsonschema:"The identifier of the recipe"`
	RecipeStepID        string `json:"recipeStepID"        jsonschema:"The identifier of the step"`
	RecipeStepProductID string `json:"recipeStepProductID" jsonschema:"The identifier of the product to read"`
}

var getRecipeStepProductTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepProduct",
	Description: "Get a product of a recipe step by its ID",
	Annotations: readOnly(),
}

// GetRecipeStepProduct reads one product of a recipe step.
func (t *Tools) GetRecipeStepProduct(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipeStepProductInput) (*sdkmcp.CallToolResult, *mealplanning.RecipeStepProduct, error) {
	ctx, err := t.begin(ctx, req, getRecipeStepProductTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadRecipeStepProduct(ctx, in.RecipeID, in.RecipeStepID, in.RecipeStepProductID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

// GetRecipeStepProductsInput is what GetRecipeStepProducts takes.
type GetRecipeStepProductsInput struct {
	Filter       *filtering.QueryFilter `json:"filter,omitempty" jsonschema:"The page to read; absent reads the first page, oldest first"`
	RecipeID     string                 `json:"recipeID"         jsonschema:"The identifier of the recipe"`
	RecipeStepID string                 `json:"recipeStepID"     jsonschema:"The identifier of the step"`
}

var getRecipeStepProductsTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepProducts",
	Description: "Page the products of a recipe step",
	Annotations: readOnly(),
}

// GetRecipeStepProducts pages the products of a recipe step.
func (t *Tools) GetRecipeStepProducts(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipeStepProductsInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.RecipeStepProduct], error) {
	ctx, err := t.begin(ctx, req, getRecipeStepProductsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ListRecipeStepProducts(ctx, in.RecipeID, in.RecipeStepID, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
