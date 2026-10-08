package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipeStepProductInvocation struct {
		RecipeID            string `jsonschema:"description=The recipe ID"`
		RecipeStepID        string `jsonschema:"description=The recipe step ID"`
		RecipeStepProductID string `jsonschema:"description=The recipe step product ID"`
	}
)

var recipeStepProductsSchema = map[string]any{
	"ID":                                mcptools.StringField("The ID of the recipe step product"),
	mcptools.FieldCreatedAt:             mcptools.TimestampField("When the recipe step product was created"),
	mcptools.FieldLastUpdatedAt:         mcptools.TimestampField("When the recipe step product was last updated"),
	mcptools.FieldArchivedAt:            mcptools.TimestampField("When the recipe step product was soft deleted"),
	fieldBelongsToRecipeStep:            mcptools.StringField("The ID of the recipe step this product belongs to"),
	mcptools.FieldName:                  mcptools.StringField("Name of the product"),
	"Type":                              mcptools.StringField("The type of product (e.g., 'ingredient', 'waste', 'intermediate')"),
	"QuantityNotes":                     mcptools.StringField("Notes about the quantity"),
	fieldStorageInstructions:            mcptools.StringField("Storage instructions for the product"),
	fieldMeasurementUnit:                mcptools.ObjectType(validMeasurementUnitsSchema),
	"MinMeasurementQuantity":            mcptools.FloatField("Minimum measurement quantity"),
	"MaxMeasurementQuantity":            mcptools.FloatField("Maximum measurement quantity"),
	"MinItemQuantity":                   mcptools.FloatField("Minimum item quantity"),
	"MaxItemQuantity":                   mcptools.FloatField("Maximum item quantity"),
	fieldMinStorageTemperatureInCelsius: mcptools.FloatField("Minimum storage temperature in celsius"),
	fieldMaxStorageTemperatureInCelsius: mcptools.FloatField("Maximum storage temperature in celsius"),
	"MinStorageDurationInSeconds":       mcptools.UintField("Minimum storage duration in seconds"),
	"MaxStorageDurationInSeconds":       mcptools.UintField("Maximum storage duration in seconds"),
	"ContainedInVesselIndex":            mcptools.UintField("The index of the vessel this product is contained in, if any"),
	fieldIndex:                          mcptools.UintField("The display index/order of this product"),
	"IsWaste":                           mcptools.BoolField("Whether this product is waste"),
	"IsLiquid":                          mcptools.BoolField("Whether this product is a liquid"),
	"Compostable":                       mcptools.BoolField("Whether this product is compostable"),
}

var getRecipeStepProductTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepProduct",
	Description: "Get a recipe step product by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID:            mcptools.StringField("The ID of the recipe"),
		fieldRecipeStepID:        mcptools.StringField("The ID of the recipe step"),
		fieldRecipeStepProductID: mcptools.StringField("The ID of the recipe step product to get"),
	}),
	OutputSchema: mcptools.SchemaObject(recipeStepProductsSchema),
}

func (t *Tools) GetRecipeStepProduct() sdkmcp.ToolHandlerFor[*GetRecipeStepProductInvocation, *mealplanning.RecipeStepProduct] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipeStepProductInvocation) (*sdkmcp.CallToolResult, *mealplanning.RecipeStepProduct, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetRecipeStepProduct(ctx, x.RecipeID, x.RecipeStepID, x.RecipeStepProductID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetRecipeStepProductsInvocation struct {
		Filter       *filtering.QueryFilter
		RecipeID     string
		RecipeStepID string
	}

	GetRecipeStepProductsResult struct {
		Results []*mealplanning.RecipeStepProduct
	}
)

var getRecipeStepProductsTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepProducts",
	Description: "Get recipe step products with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID:        mcptools.StringField("The ID of the recipe"),
		fieldRecipeStepID:    mcptools.StringField("The ID of the recipe step"),
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(recipeStepProductsSchema)),
	}),
}

func (t *Tools) GetRecipeStepProducts() sdkmcp.ToolHandlerFor[*GetRecipeStepProductsInvocation, *GetRecipeStepProductsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipeStepProductsInvocation) (*sdkmcp.CallToolResult, *GetRecipeStepProductsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetRecipeStepProducts(ctx, x.RecipeID, x.RecipeStepID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipeStepProductsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
