package main

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidIngredientsInvocation struct {
		ValidIngredientID string `jsonschema:"description=The ingredient ID"`
	}
)

var validIngredientsSchema = map[string]any{
	"ID":                             stringField("The ID of the valid ingredient"),
	"CreatedAt":                      timestampField("When the valid ingredient was created"),
	"LastUpdatedAt":                  timestampField("When the valid ingredient was last updated"),
	"ArchivedAt":                     timestampField("When the valid ingredient was soft deleted"),
	"Name":                           stringField("Name of the ingredient"),
	"Description":                    stringField("Description of the ingredient"),
	"Warning":                        stringField("For things like allergen warnings"),
	"IconPath":                       stringField("The URL for the icon for the item"),
	"ContainsDairy":                  boolField("Whether or not the valid ingredient contains dairy"),
	"ContainsPeanut":                 boolField("Whether or not the valid ingredient contains peanut"),
	"ContainsTreeNut":                boolField("Whether or not the valid ingredient contains tree nut"),
	"ContainsEgg":                    boolField("Whether or not the valid ingredient contains egg"),
	"ContainsWheat":                  boolField("Whether or not the valid ingredient contains wheat"),
	"ContainsShellfish":              boolField("Whether or not the valid ingredient contains shellfish"),
	"ContainsSesame":                 boolField("Whether or not the valid ingredient contains sesame"),
	"ContainsFish":                   boolField("Whether or not the valid ingredient contains fish"),
	"ContainsGluten":                 boolField("Whether or not the valid ingredient contains gluten"),
	"AnimalFlesh":                    boolField("Whether or not the valid ingredient is derived from animal flesh"),
	"IsLiquid":                       boolField("Whether or not the valid ingredient is a liquid"),
	"ContainsSoy":                    boolField("Whether or not the valid ingredient contains soy"),
	"PluralName":                     stringField("The plural name for the ingredient. So for an ingredient named 'onion', this would be 'onions'"),
	"AnimalDerived":                  boolField("Whether or not the valid ingredient AnimalDerived"),
	"RestrictToPreparations":         boolField("Whether or not the valid ingredient is restrictToPreparations"),
	"ContaminatesEquipment":          boolField("Whether or not the valid ingredient contaminates equipment"),
	"MinStorageTemperatureInCelsius": floatField("Minimum storage temperature in Celsius (optional)"),
	"MaxStorageTemperatureInCelsius": floatField("Maximum storage temperature in Celsius (optional)"),
	"StorageInstructions":            stringField("Instructions on how to store the item."),
	"Slug":                           stringField("An easy-to-use URL slug for the ingredient"),
	"ContainsAlcohol":                boolField("Whether or not the valid ingredient contains Alcohol"),
	"ShoppingSuggestions":            stringField("Suggestions for the user to keep in mind when shopping for this ingredient"),
	"IsStarch":                       boolField("Whether or not the valid ingredient is a Starch"),
	"IsProtein":                      boolField("Whether or not the valid ingredient is a Protein"),
	"IsGrain":                        boolField("Whether or not the valid ingredient is a Grain"),
	"IsFruit":                        boolField("Whether or not the valid ingredient is a Fruit"),
	"IsSalt":                         boolField("Whether or not the valid ingredient is a Salt"),
	"IsFat":                          boolField("Whether or not the valid ingredient is a Fat"),
	"IsAcid":                         boolField("Whether or not the valid ingredient is a Acid"),
	"IsHeat":                         boolField("Whether or not the valid ingredient is a Heat"),
}

var getValidIngredientTool = &mcp.Tool{
	Name:        "GetValidIngredient",
	Description: "Get a valid ingredient by it's ID",
	InputSchema: schemaObject(map[string]any{
		"ValidIngredientID": stringField("The ID of the valid ingredient to get"),
	}),
	OutputSchema: schemaObject(validIngredientsSchema),
}

func (h *mcpToolManager) GetValidIngredient() mcp.ToolHandlerFor[*GetValidIngredientsInvocation, *mealplanning.ValidIngredient] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidIngredientsInvocation) (*mcp.CallToolResult, *mealplanning.ValidIngredient, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetValidIngredient(ctx, x.ValidIngredientID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	SearchValidIngredientsInvocation struct {
		Filter *filtering.QueryFilter
		Query  string `jsonschema_description:"The ingredient name query"`
	}

	SearchValidIngredientsResult struct {
		Results []*mealplanning.ValidIngredient
	}
)

var searchForValidIngredientsTool = &mcp.Tool{
	Name:        "SearchForValidIngredients",
	Description: "Search for valid ingredients with optional filtering and query string",
	InputSchema: schemaObject(map[string]any{
		"Filter": queryFilterSchema(),
		"Query": map[string]any{
			"type":        strType,
			"description": "The ingredient name query",
		},
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(validIngredientsSchema)),
	}),
}

func (h *mcpToolManager) SearchForValidIngredients() mcp.ToolHandlerFor[*SearchValidIngredientsInvocation, *SearchValidIngredientsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *SearchValidIngredientsInvocation) (*mcp.CallToolResult, *SearchValidIngredientsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.SearchForValidIngredients(ctx, x.Query, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &SearchValidIngredientsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}
