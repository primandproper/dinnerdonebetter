package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipeInvocation struct {
		RecipeID string `jsonschema:"description=The recipe ID"`
	}
)

var recipesSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the recipe"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the recipe was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the recipe was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the recipe was soft deleted"),
	"InspiredByRecipeID":        mcptools.StringField("The ID of the recipe this recipe was inspired by, if any"),
	mcptools.FieldName:          mcptools.StringField("Name of the recipe"),
	mcptools.FieldDescription:   mcptools.StringField("Description of the recipe"),
	"Source":                    mcptools.StringField("Source of the recipe"),
	"SourceISBN":                mcptools.StringField("ISBN of the recipe source book, if any"),
	fieldSlug:                   mcptools.StringField("An easy-to-use URL slug for the recipe"),
	fieldCreatedByUser:          mcptools.StringField("The ID of the user who created the recipe"),
	"PortionName":               mcptools.StringField("Name for a single portion (e.g., 'serving', 'piece')"),
	"PluralPortionName":         mcptools.StringField("Plural name for portions (e.g., 'servings', 'pieces')"),
	"YieldsComponentType":       mcptools.StringField("The type of component this recipe yields"),
	"MinEstimatedPortions":      mcptools.FloatField("Minimum estimated portions yielded (required)"),
	"MaxEstimatedPortions":      mcptools.FloatField("Maximum estimated portions yielded (optional)"),
	"SealOfApproval":            mcptools.BoolField("Whether this recipe has the seal of approval"),
	"EligibleForMeals":          mcptools.BoolField("Whether this recipe is eligible for meals"),
	"PrepTasks":                 mcptools.ArrayType(mcptools.SchemaObject(recipePrepTasksSchema)),
	"Steps":                     mcptools.ArrayType(mcptools.SchemaObject(recipeStepsSchema)),
	"Media":                     mcptools.ArrayType(mcptools.SchemaObject(recipeMediaSchema)),
}

var getRecipeTool = &sdkmcp.Tool{
	Name:        "GetRecipe",
	Description: "Get a recipe by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID: mcptools.StringField("The ID of the recipe to get"),
	}),
	OutputSchema: mcptools.SchemaObject(recipesSchema),
}

func (t *Tools) GetRecipe() sdkmcp.ToolHandlerFor[*GetRecipeInvocation, *mealplanning.Recipe] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipeInvocation) (*sdkmcp.CallToolResult, *mealplanning.Recipe, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetRecipe(ctx, x.RecipeID)
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

var getRecipesTool = &sdkmcp.Tool{
	Name:        "GetRecipes",
	Description: "Get recipes with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(recipesSchema)),
	}),
}

func (t *Tools) GetRecipes() sdkmcp.ToolHandlerFor[*GetRecipesInvocation, *GetRecipesResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipesInvocation) (*sdkmcp.CallToolResult, *GetRecipesResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetRecipes(ctx, "", x.Filter)
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

var searchForRecipesTool = &sdkmcp.Tool{
	Name:        "SearchForRecipes",
	Description: "Search for recipes with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldQuery:  mcptools.StringField("The search query string"),
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(recipesSchema)),
	}),
}

func (t *Tools) SearchForRecipes() sdkmcp.ToolHandlerFor[*SearchForRecipesInvocation, *SearchForRecipesResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *SearchForRecipesInvocation) (*sdkmcp.CallToolResult, *SearchForRecipesResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.SearchForRecipes(ctx, x.Query, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &SearchForRecipesResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
