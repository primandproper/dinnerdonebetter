package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipeStepIngredientInvocation struct {
		RecipeID               string `jsonschema:"description=The recipe ID"`
		RecipeStepID           string `jsonschema:"description=The recipe step ID"`
		RecipeStepIngredientID string `jsonschema:"description=The recipe step ingredient ID"`
	}
)

var recipeStepIngredientsSchema = map[string]any{
	"ID":                     stringField("The ID of the recipe step ingredient"),
	"CreatedAt":              timestampField("When the recipe step ingredient was created"),
	"LastUpdatedAt":          timestampField("When the recipe step ingredient was last updated"),
	"ArchivedAt":             timestampField("When the recipe step ingredient was soft deleted"),
	"BelongsToRecipeStep":    stringField("The ID of the recipe step this ingredient belongs to"),
	"Name":                   stringField("Name of the ingredient"),
	"QuantityNotes":          stringField("Notes about the quantity"),
	"IngredientNotes":        stringField("Notes about the ingredient"),
	"Ingredient":             objectType(validIngredientsSchema),
	"MeasurementUnit":        objectType(validMeasurementUnitsSchema),
	"MinQuantity":            floatField("Minimum quantity of this ingredient (required)"),
	"MaxQuantity":            floatField("Maximum quantity of this ingredient (optional)"),
	"RecipeStepProductID":    stringField("The ID of the recipe step product this ingredient is associated with, if any"),
	"ProductOfRecipeID":      stringField("The ID of the recipe that produces this ingredient, if any"),
	"ProductPercentageToUse": floatField("The percentage of the product to use, if any"),
	"VesselIndex":            uintField("The index of the vessel this ingredient is in, if any"),
	"OptionIndex":            uintField("The option index for this ingredient"),
	"Optional":               boolField("Whether this ingredient is optional"),
	"ToTaste":                boolField("Whether this ingredient is 'to taste'"),
}

var getRecipeStepIngredientTool = &mcp.Tool{
	Name:        "GetRecipeStepIngredient",
	Description: "Get a recipe step ingredient by it's ID",
	InputSchema: schemaObject(map[string]any{
		"RecipeID":               stringField("The ID of the recipe"),
		"RecipeStepID":           stringField("The ID of the recipe step"),
		"RecipeStepIngredientID": stringField("The ID of the recipe step ingredient to get"),
	}),
	OutputSchema: schemaObject(recipeStepIngredientsSchema),
}

func (h *mcpToolManager) GetRecipeStepIngredient() mcp.ToolHandlerFor[*GetRecipeStepIngredientInvocation, *mealplanning.RecipeStepIngredient] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipeStepIngredientInvocation) (*mcp.CallToolResult, *mealplanning.RecipeStepIngredient, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetRecipeStepIngredient(ctx, x.RecipeID, x.RecipeStepID, x.RecipeStepIngredientID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetRecipeStepIngredientsInvocation struct {
		Filter       *filtering.QueryFilter
		RecipeID     string
		RecipeStepID string
	}

	GetRecipeStepIngredientsResult struct {
		Results []*mealplanning.RecipeStepIngredient
	}
)

var getRecipeStepIngredientsTool = &mcp.Tool{
	Name:        "GetRecipeStepIngredients",
	Description: "Get recipe step ingredients with optional filtering",
	InputSchema: schemaObject(map[string]any{
		"RecipeID":     stringField("The ID of the recipe"),
		"RecipeStepID": stringField("The ID of the recipe step"),
		"Filter":       queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(recipeStepIngredientsSchema)),
	}),
}

func (h *mcpToolManager) GetRecipeStepIngredients() mcp.ToolHandlerFor[*GetRecipeStepIngredientsInvocation, *GetRecipeStepIngredientsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipeStepIngredientsInvocation) (*mcp.CallToolResult, *GetRecipeStepIngredientsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetRecipeStepIngredients(ctx, x.RecipeID, x.RecipeStepID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipeStepIngredientsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
