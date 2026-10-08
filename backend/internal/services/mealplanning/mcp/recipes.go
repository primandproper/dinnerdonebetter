package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// useDatabaseSearch is what every search here passes for the manager's useSearchService: the
// text index is the API's to query, and a tool call reads the database the same as the tools
// over the repository did.
const useDatabaseSearch = false

// GetRecipeInput is what GetRecipe takes.
type GetRecipeInput struct {
	RecipeID string `json:"recipeID" jsonschema:"The identifier of the recipe to read"`
}

var getRecipeTool = &sdkmcp.Tool{
	Name:        "GetRecipe",
	Description: "Get a recipe by its ID, with its steps, prep tasks and media",
	Annotations: readOnly(),
}

// GetRecipe reads one recipe.
func (t *Tools) GetRecipe(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipeInput) (*sdkmcp.CallToolResult, *mealplanning.Recipe, error) {
	ctx, err := t.begin(ctx, req, getRecipeTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadRecipe(ctx, in.RecipeID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

var getRecipesTool = &sdkmcp.Tool{
	Name:        "GetRecipes",
	Description: "Page the recipes",
	Annotations: readOnly(),
}

// GetRecipes pages every recipe, whatever its status.
func (t *Tools) GetRecipes(ctx context.Context, req *sdkmcp.CallToolRequest, in FilterInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	ctx, err := t.begin(ctx, req, getRecipesTool)
	if err != nil {
		return nil, nil, err
	}

	results, err := t.manager.ListRecipes(ctx, "", in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, results, nil
}

var searchForRecipesTool = &sdkmcp.Tool{
	Name:        "SearchForRecipes",
	Description: "Search the recipes by name",
	Annotations: readOnly(),
}

// SearchForRecipes pages the recipes matching a query.
func (t *Tools) SearchForRecipes(ctx context.Context, req *sdkmcp.CallToolRequest, in SearchInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	ctx, err := t.begin(ctx, req, searchForRecipesTool)
	if err != nil {
		return nil, nil, err
	}

	results, err := t.manager.SearchRecipes(ctx, in.Query, useDatabaseSearch, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, results, nil
}
