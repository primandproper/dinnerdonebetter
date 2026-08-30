package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v13/filtering"

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
	"ID":                                stringField("The ID of the recipe step product"),
	fieldCreatedAt:                      timestampField("When the recipe step product was created"),
	fieldLastUpdatedAt:                  timestampField("When the recipe step product was last updated"),
	fieldArchivedAt:                     timestampField("When the recipe step product was soft deleted"),
	fieldBelongsToRecipeStep:            stringField("The ID of the recipe step this product belongs to"),
	fieldName:                           stringField("Name of the product"),
	"Type":                              stringField("The type of product (e.g., 'ingredient', 'waste', 'intermediate')"),
	"QuantityNotes":                     stringField("Notes about the quantity"),
	fieldStorageInstructions:            stringField("Storage instructions for the product"),
	fieldMeasurementUnit:                objectType(validMeasurementUnitsSchema),
	"MinMeasurementQuantity":            floatField("Minimum measurement quantity"),
	"MaxMeasurementQuantity":            floatField("Maximum measurement quantity"),
	"MinItemQuantity":                   floatField("Minimum item quantity"),
	"MaxItemQuantity":                   floatField("Maximum item quantity"),
	fieldMinStorageTemperatureInCelsius: floatField("Minimum storage temperature in celsius"),
	fieldMaxStorageTemperatureInCelsius: floatField("Maximum storage temperature in celsius"),
	"MinStorageDurationInSeconds":       uintField("Minimum storage duration in seconds"),
	"MaxStorageDurationInSeconds":       uintField("Maximum storage duration in seconds"),
	"ContainedInVesselIndex":            uintField("The index of the vessel this product is contained in, if any"),
	fieldIndex:                          uintField("The display index/order of this product"),
	"IsWaste":                           boolField("Whether this product is waste"),
	"IsLiquid":                          boolField("Whether this product is a liquid"),
	"Compostable":                       boolField("Whether this product is compostable"),
}

var getRecipeStepProductTool = &mcp.Tool{
	Name:        "GetRecipeStepProduct",
	Description: "Get a recipe step product by it's ID",
	InputSchema: schemaObject(map[string]any{
		fieldRecipeID:            stringField("The ID of the recipe"),
		fieldRecipeStepID:        stringField("The ID of the recipe step"),
		fieldRecipeStepProductID: stringField("The ID of the recipe step product to get"),
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
		fieldRecipeID:     stringField("The ID of the recipe"),
		fieldRecipeStepID: stringField("The ID of the recipe step"),
		fieldFilter:       filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		fieldResults: arrayType(schemaObject(recipeStepProductsSchema)),
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
