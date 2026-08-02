package main

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipeStepProductInvocation struct {
		RecipeID            string `jsonschema:"description=The recipe ID"`
		RecipeStepID        string `jsonschema:"description=The recipe step ID"`
		RecipeStepProductID string `jsonschema:"description=The recipe step product ID"`
	}
)

var recipeStepProductsSchema = map[string]any{
	"ID":                             stringField("The ID of the recipe step product"),
	"CreatedAt":                      timestampField("When the recipe step product was created"),
	"LastUpdatedAt":                  timestampField("When the recipe step product was last updated"),
	"ArchivedAt":                     timestampField("When the recipe step product was soft deleted"),
	"BelongsToRecipeStep":            stringField("The ID of the recipe step this product belongs to"),
	"Name":                           stringField("Name of the product"),
	"Type":                           stringField("The type of product (e.g., 'ingredient', 'waste', 'intermediate')"),
	"QuantityNotes":                  stringField("Notes about the quantity"),
	"StorageInstructions":            stringField("Storage instructions for the product"),
	"MeasurementUnit":                objectType(validMeasurementUnitsSchema),
	"MinMeasurementQuantity":         floatField("Minimum measurement quantity"),
	"MaxMeasurementQuantity":         floatField("Maximum measurement quantity"),
	"MinItemQuantity":                floatField("Minimum item quantity"),
	"MaxItemQuantity":                floatField("Maximum item quantity"),
	"MinStorageTemperatureInCelsius": floatField("Minimum storage temperature in celsius"),
	"MaxStorageTemperatureInCelsius": floatField("Maximum storage temperature in celsius"),
	"MinStorageDurationInSeconds":    uintField("Minimum storage duration in seconds"),
	"MaxStorageDurationInSeconds":    uintField("Maximum storage duration in seconds"),
	"ContainedInVesselIndex":         uintField("The index of the vessel this product is contained in, if any"),
	"Index":                          uintField("The display index/order of this product"),
	"IsWaste":                        boolField("Whether this product is waste"),
	"IsLiquid":                       boolField("Whether this product is a liquid"),
	"Compostable":                    boolField("Whether this product is compostable"),
}

var getRecipeStepProductTool = &mcp.Tool{
	Name:        "GetRecipeStepProduct",
	Description: "Get a recipe step product by it's ID",
	InputSchema: schemaObject(map[string]any{
		"RecipeID":            stringField("The ID of the recipe"),
		"RecipeStepID":        stringField("The ID of the recipe step"),
		"RecipeStepProductID": stringField("The ID of the recipe step product to get"),
	}),
	OutputSchema: schemaObject(recipeStepProductsSchema),
}

func (h *mcpToolManager) GetRecipeStepProduct() mcp.ToolHandlerFor[*GetRecipeStepProductInvocation, *mealplanning.RecipeStepProduct] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipeStepProductInvocation) (*mcp.CallToolResult, *mealplanning.RecipeStepProduct, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetRecipeStepProduct(ctx, x.RecipeID, x.RecipeStepID, x.RecipeStepProductID)
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

var getRecipeStepProductsTool = &mcp.Tool{
	Name:        "GetRecipeStepProducts",
	Description: "Get recipe step products with optional filtering",
	InputSchema: schemaObject(map[string]any{
		"RecipeID":     stringField("The ID of the recipe"),
		"RecipeStepID": stringField("The ID of the recipe step"),
		"Filter":       queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(recipeStepProductsSchema)),
	}),
}

func (h *mcpToolManager) GetRecipeStepProducts() mcp.ToolHandlerFor[*GetRecipeStepProductsInvocation, *GetRecipeStepProductsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipeStepProductsInvocation) (*mcp.CallToolResult, *GetRecipeStepProductsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetRecipeStepProducts(ctx, x.RecipeID, x.RecipeStepID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipeStepProductsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
