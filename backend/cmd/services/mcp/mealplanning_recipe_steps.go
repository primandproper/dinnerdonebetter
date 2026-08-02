package main

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipeStepInvocation struct {
		RecipeID     string `jsonschema:"description=The recipe ID"`
		RecipeStepID string `jsonschema:"description=The recipe step ID"`
	}
)

var recipeMediaSchema = map[string]any{
	"ID":                  stringField("The ID of the recipe media"),
	"CreatedAt":           timestampField("When the recipe media was created"),
	"LastUpdatedAt":       timestampField("When the recipe media was last updated"),
	"ArchivedAt":          timestampField("When the recipe media was soft deleted"),
	"BelongsToRecipe":     stringField("The ID of the recipe this media belongs to, if any"),
	"BelongsToRecipeStep": stringField("The ID of the recipe step this media belongs to, if any"),
	"MimeType":            stringField("The MIME type of the media"),
	"InternalPath":        stringField("The internal path to the media file"),
	"ExternalPath":        stringField("The external path to the media file"),
	"Index":               uintField("The index of the media"),
}

var recipeStepsSchema = map[string]any{
	"ID":                        stringField("The ID of the recipe step"),
	"CreatedAt":                 timestampField("When the recipe step was created"),
	"LastUpdatedAt":             timestampField("When the recipe step was last updated"),
	"ArchivedAt":                timestampField("When the recipe step was soft deleted"),
	"BelongsToRecipe":           stringField("The ID of the recipe this step belongs to"),
	"ConditionExpression":       stringField("The condition expression for this step"),
	"Notes":                     stringField("Notes about the step"),
	"ExplicitInstructions":      stringField("Explicit instructions for this step"),
	"Media":                     arrayType(schemaObject(recipeMediaSchema)),
	"Products":                  arrayType(schemaObject(recipeStepProductsSchema)),
	"Instruments":               arrayType(schemaObject(recipeStepInstrumentsSchema)),
	"Vessels":                   arrayType(schemaObject(recipeStepVesselsSchema)),
	"CompletionConditions":      arrayType(schemaObject(recipeStepCompletionConditionsSchema)),
	"Ingredients":               arrayType(schemaObject(recipeStepIngredientsSchema)),
	"Preparation":               objectType(validPreparationsSchema),
	"Index":                     uintField("The index of the step within the recipe"),
	"Optional":                  boolField("Whether this step is optional"),
	"StartTimerAutomatically":   boolField("Whether to start a timer automatically for this step"),
	"MinEstimatedTimeInSeconds": uintField("Minimum estimated time in seconds (optional)"),
	"MaxEstimatedTimeInSeconds": uintField("Maximum estimated time in seconds (optional)"),
	"MinTemperatureInCelsius":   floatField("Minimum temperature in Celsius (optional)"),
	"MaxTemperatureInCelsius":   floatField("Maximum temperature in Celsius (optional)"),
}

var getRecipeStepTool = &mcp.Tool{
	Name:        "GetRecipeStep",
	Description: "Get a recipe step by it's ID",
	InputSchema: schemaObject(map[string]any{
		"RecipeID":     stringField("The ID of the recipe"),
		"RecipeStepID": stringField("The ID of the recipe step to get"),
	}),
	OutputSchema: schemaObject(recipeStepsSchema),
}

func (h *mcpToolManager) GetRecipeStep() mcp.ToolHandlerFor[*GetRecipeStepInvocation, *mealplanning.RecipeStep] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipeStepInvocation) (*mcp.CallToolResult, *mealplanning.RecipeStep, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetRecipeStep(ctx, x.RecipeID, x.RecipeStepID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetRecipeStepsInvocation struct {
		Filter   *filtering.QueryFilter
		RecipeID string
	}

	GetRecipeStepsResult struct {
		Results []*mealplanning.RecipeStep
	}
)

var getRecipeStepsTool = &mcp.Tool{
	Name:        "GetRecipeSteps",
	Description: "Get recipe steps with optional filtering",
	InputSchema: schemaObject(map[string]any{
		"RecipeID": stringField("The ID of the recipe"),
		"Filter":   queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(recipeStepsSchema)),
	}),
}

func (h *mcpToolManager) GetRecipeSteps() mcp.ToolHandlerFor[*GetRecipeStepsInvocation, *GetRecipeStepsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipeStepsInvocation) (*mcp.CallToolResult, *GetRecipeStepsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetRecipeSteps(ctx, x.RecipeID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipeStepsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
