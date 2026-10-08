package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidIngredientsInvocation struct {
		ValidIngredientID string `jsonschema:"description=The ingredient ID"`
	}
)

var validIngredientsSchema = map[string]any{
	"ID":                                mcptools.StringField("The ID of the valid ingredient"),
	mcptools.FieldCreatedAt:             mcptools.TimestampField("When the valid ingredient was created"),
	mcptools.FieldLastUpdatedAt:         mcptools.TimestampField("When the valid ingredient was last updated"),
	mcptools.FieldArchivedAt:            mcptools.TimestampField("When the valid ingredient was soft deleted"),
	mcptools.FieldName:                  mcptools.StringField("Name of the ingredient"),
	mcptools.FieldDescription:           mcptools.StringField("Description of the ingredient"),
	"Warning":                           mcptools.StringField("For things like allergen warnings"),
	fieldIconPath:                       mcptools.StringField("The URL for the icon for the item"),
	"ContainsDairy":                     mcptools.BoolField("Whether or not the valid ingredient contains dairy"),
	"ContainsPeanut":                    mcptools.BoolField("Whether or not the valid ingredient contains peanut"),
	"ContainsTreeNut":                   mcptools.BoolField("Whether or not the valid ingredient contains tree nut"),
	"ContainsEgg":                       mcptools.BoolField("Whether or not the valid ingredient contains egg"),
	"ContainsWheat":                     mcptools.BoolField("Whether or not the valid ingredient contains wheat"),
	"ContainsShellfish":                 mcptools.BoolField("Whether or not the valid ingredient contains shellfish"),
	"ContainsSesame":                    mcptools.BoolField("Whether or not the valid ingredient contains sesame"),
	"ContainsFish":                      mcptools.BoolField("Whether or not the valid ingredient contains fish"),
	"ContainsGluten":                    mcptools.BoolField("Whether or not the valid ingredient contains gluten"),
	"AnimalFlesh":                       mcptools.BoolField("Whether or not the valid ingredient is derived from animal flesh"),
	"IsLiquid":                          mcptools.BoolField("Whether or not the valid ingredient is a liquid"),
	"ContainsSoy":                       mcptools.BoolField("Whether or not the valid ingredient contains soy"),
	fieldPluralName:                     mcptools.StringField("The plural name for the ingredient. So for an ingredient named 'onion', this would be 'onions'"),
	"AnimalDerived":                     mcptools.BoolField("Whether or not the valid ingredient AnimalDerived"),
	"RestrictToPreparations":            mcptools.BoolField("Whether or not the valid ingredient is restrictToPreparations"),
	"ContaminatesEquipment":             mcptools.BoolField("Whether or not the valid ingredient contaminates equipment"),
	fieldMinStorageTemperatureInCelsius: mcptools.FloatField("Minimum storage temperature in Celsius (optional)"),
	fieldMaxStorageTemperatureInCelsius: mcptools.FloatField("Maximum storage temperature in Celsius (optional)"),
	fieldStorageInstructions:            mcptools.StringField("Instructions on how to store the item."),
	fieldSlug:                           mcptools.StringField("An easy-to-use URL slug for the ingredient"),
	"ContainsAlcohol":                   mcptools.BoolField("Whether or not the valid ingredient contains Alcohol"),
	"ShoppingSuggestions":               mcptools.StringField("Suggestions for the user to keep in mind when shopping for this ingredient"),
	"IsStarch":                          mcptools.BoolField("Whether or not the valid ingredient is a Starch"),
	"IsProtein":                         mcptools.BoolField("Whether or not the valid ingredient is a Protein"),
	"IsGrain":                           mcptools.BoolField("Whether or not the valid ingredient is a Grain"),
	"IsFruit":                           mcptools.BoolField("Whether or not the valid ingredient is a Fruit"),
	"IsSalt":                            mcptools.BoolField("Whether or not the valid ingredient is a Salt"),
	"IsFat":                             mcptools.BoolField("Whether or not the valid ingredient is a Fat"),
	"IsAcid":                            mcptools.BoolField("Whether or not the valid ingredient is a Acid"),
	"IsHeat":                            mcptools.BoolField("Whether or not the valid ingredient is a Heat"),
}

var getValidIngredientTool = &sdkmcp.Tool{
	Name:        "GetValidIngredient",
	Description: "Get a valid ingredient by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldValidIngredientID: mcptools.StringField("The ID of the valid ingredient to get"),
	}),
	OutputSchema: mcptools.SchemaObject(validIngredientsSchema),
}

func (t *Tools) GetValidIngredient() sdkmcp.ToolHandlerFor[*GetValidIngredientsInvocation, *mealplanning.ValidIngredient] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidIngredientsInvocation) (*sdkmcp.CallToolResult, *mealplanning.ValidIngredient, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetValidIngredient(ctx, x.ValidIngredientID)
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

var searchForValidIngredientsTool = &sdkmcp.Tool{
	Name:        "SearchForValidIngredients",
	Description: "Search for valid ingredients with optional filtering and query string",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
		mcptools.FieldQuery:  mcptools.StringField("The ingredient name query"),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validIngredientsSchema)),
	}),
}

func (t *Tools) SearchForValidIngredients() sdkmcp.ToolHandlerFor[*SearchValidIngredientsInvocation, *SearchValidIngredientsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *SearchValidIngredientsInvocation) (*sdkmcp.CallToolResult, *SearchValidIngredientsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.SearchForValidIngredients(ctx, x.Query, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &SearchValidIngredientsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}
