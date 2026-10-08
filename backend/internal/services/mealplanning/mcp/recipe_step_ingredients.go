package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipeStepIngredientInvocation struct {
		RecipeID               string `jsonschema:"description=The recipe ID"`
		RecipeStepID           string `jsonschema:"description=The recipe step ID"`
		RecipeStepIngredientID string `jsonschema:"description=The recipe step ingredient ID"`
	}
)

var recipeStepIngredientsSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the recipe step ingredient"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the recipe step ingredient was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the recipe step ingredient was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the recipe step ingredient was soft deleted"),
	fieldBelongsToRecipeStep:    mcptools.StringField("The ID of the recipe step this ingredient belongs to"),
	mcptools.FieldName:          mcptools.StringField("Name of the ingredient"),
	"QuantityNotes":             mcptools.StringField("Notes about the quantity"),
	"IngredientNotes":           mcptools.StringField("Notes about the ingredient"),
	fieldIngredient:             mcptools.ObjectType(validIngredientsSchema),
	fieldMeasurementUnit:        mcptools.ObjectType(validMeasurementUnitsSchema),
	fieldMinQuantity:            mcptools.FloatField("Minimum quantity of this ingredient (required)"),
	fieldMaxQuantity:            mcptools.FloatField("Maximum quantity of this ingredient (optional)"),
	fieldRecipeStepProductID:    mcptools.StringField("The ID of the recipe step product this ingredient is associated with, if any"),
	"ProductOfRecipeID":         mcptools.StringField("The ID of the recipe that produces this ingredient, if any"),
	"ProductPercentageToUse":    mcptools.FloatField("The percentage of the product to use, if any"),
	"VesselIndex":               mcptools.UintField("The index of the vessel this ingredient is in, if any"),
	"OptionIndex":               mcptools.UintField("The option index for this ingredient"),
	fieldOptional:               mcptools.BoolField("Whether this ingredient is optional"),
	"ToTaste":                   mcptools.BoolField("Whether this ingredient is 'to taste'"),
}

var getRecipeStepIngredientTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepIngredient",
	Description: "Get a recipe step ingredient by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID:            mcptools.StringField("The ID of the recipe"),
		fieldRecipeStepID:        mcptools.StringField("The ID of the recipe step"),
		"RecipeStepIngredientID": mcptools.StringField("The ID of the recipe step ingredient to get"),
	}),
	OutputSchema: mcptools.SchemaObject(recipeStepIngredientsSchema),
}

func (t *Tools) GetRecipeStepIngredient() sdkmcp.ToolHandlerFor[*GetRecipeStepIngredientInvocation, *mealplanning.RecipeStepIngredient] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipeStepIngredientInvocation) (*sdkmcp.CallToolResult, *mealplanning.RecipeStepIngredient, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetRecipeStepIngredient(ctx, x.RecipeID, x.RecipeStepID, x.RecipeStepIngredientID)
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

var getRecipeStepIngredientsTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepIngredients",
	Description: "Get recipe step ingredients with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID:        mcptools.StringField("The ID of the recipe"),
		fieldRecipeStepID:    mcptools.StringField("The ID of the recipe step"),
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(recipeStepIngredientsSchema)),
	}),
}

func (t *Tools) GetRecipeStepIngredients() sdkmcp.ToolHandlerFor[*GetRecipeStepIngredientsInvocation, *GetRecipeStepIngredientsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipeStepIngredientsInvocation) (*sdkmcp.CallToolResult, *GetRecipeStepIngredientsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetRecipeStepIngredients(ctx, x.RecipeID, x.RecipeStepID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipeStepIngredientsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
