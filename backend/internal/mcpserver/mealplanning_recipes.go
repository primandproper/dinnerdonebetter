package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v10/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipeInvocation struct {
		RecipeID string `jsonschema:"description=The recipe ID"`
	}
)

var recipesSchema = map[string]any{
	"ID":                   stringField("The ID of the recipe"),
	"CreatedAt":            timestampField("When the recipe was created"),
	"LastUpdatedAt":        timestampField("When the recipe was last updated"),
	"ArchivedAt":           timestampField("When the recipe was soft deleted"),
	"InspiredByRecipeID":   stringField("The ID of the recipe this recipe was inspired by, if any"),
	"Name":                 stringField("Name of the recipe"),
	"Description":          stringField("Description of the recipe"),
	"Source":               stringField("Source of the recipe"),
	"SourceISBN":           stringField("ISBN of the recipe source book, if any"),
	"Slug":                 stringField("An easy-to-use URL slug for the recipe"),
	"CreatedByUser":        stringField("The ID of the user who created the recipe"),
	"PortionName":          stringField("Name for a single portion (e.g., 'serving', 'piece')"),
	"PluralPortionName":    stringField("Plural name for portions (e.g., 'servings', 'pieces')"),
	"YieldsComponentType":  stringField("The type of component this recipe yields"),
	"MinEstimatedPortions": floatField("Minimum estimated portions yielded (required)"),
	"MaxEstimatedPortions": floatField("Maximum estimated portions yielded (optional)"),
	"SealOfApproval":       boolField("Whether this recipe has the seal of approval"),
	"EligibleForMeals":     boolField("Whether this recipe is eligible for meals"),
	"PrepTasks":            arrayType(schemaObject(recipePrepTasksSchema)),
	"Steps":                arrayType(schemaObject(recipeStepsSchema)),
	"Media":                arrayType(schemaObject(recipeMediaSchema)),
}

var getRecipeTool = &mcp.Tool{
	Name:        "GetRecipe",
	Description: "Get a recipe by it's ID",
	InputSchema: schemaObject(map[string]any{
		"RecipeID": stringField("The ID of the recipe to get"),
	}),
	OutputSchema: schemaObject(recipesSchema),
}

func (h *mcpToolManager) GetRecipe() mcp.ToolHandlerFor[*GetRecipeInvocation, *mealplanning.Recipe] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipeInvocation) (*mcp.CallToolResult, *mealplanning.Recipe, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetRecipe(ctx, x.RecipeID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetRecipesInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetRecipesResult struct {
		Results []*mealplanning.Recipe
	}
)

var getRecipesTool = &mcp.Tool{
	Name:        "GetRecipes",
	Description: "Get recipes with optional filtering",
	InputSchema: schemaObject(map[string]any{
		"Filter": queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(recipesSchema)),
	}),
}

func (h *mcpToolManager) GetRecipes() mcp.ToolHandlerFor[*GetRecipesInvocation, *GetRecipesResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipesInvocation) (*mcp.CallToolResult, *GetRecipesResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetRecipes(ctx, "", x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipesResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

type (
	SearchForRecipesInvocation struct {
		Filter *filtering.QueryFilter
		Query  string
	}

	SearchForRecipesResult struct {
		Results []*mealplanning.Recipe
	}
)

var searchForRecipesTool = &mcp.Tool{
	Name:        "SearchForRecipes",
	Description: "Search for recipes with optional filtering",
	InputSchema: schemaObject(map[string]any{
		"Query":  stringField("The search query string"),
		"Filter": queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(recipesSchema)),
	}),
}

func (h *mcpToolManager) SearchForRecipes() mcp.ToolHandlerFor[*SearchForRecipesInvocation, *SearchForRecipesResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *SearchForRecipesInvocation) (*mcp.CallToolResult, *SearchForRecipesResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.SearchForRecipes(ctx, x.Query, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &SearchForRecipesResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
