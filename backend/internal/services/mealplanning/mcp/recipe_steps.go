package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipeStepInvocation struct {
		RecipeID     string `jsonschema:"description=The recipe ID"`
		RecipeStepID string `jsonschema:"description=The recipe step ID"`
	}
)

var recipeMediaSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the recipe media"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the recipe media was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the recipe media was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the recipe media was soft deleted"),
	fieldBelongsToRecipe:        mcptools.StringField("The ID of the recipe this media belongs to, if any"),
	fieldBelongsToRecipeStep:    mcptools.StringField("The ID of the recipe step this media belongs to, if any"),
	"MimeType":                  mcptools.StringField("The MIME type of the media"),
	"InternalPath":              mcptools.StringField("The internal path to the media file"),
	"ExternalPath":              mcptools.StringField("The external path to the media file"),
	fieldIndex:                  mcptools.UintField("The index of the media"),
}

var recipeStepsSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the recipe step"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the recipe step was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the recipe step was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the recipe step was soft deleted"),
	fieldBelongsToRecipe:        mcptools.StringField("The ID of the recipe this step belongs to"),
	"ConditionExpression":       mcptools.StringField("The condition expression for this step"),
	fieldNotes:                  mcptools.StringField("Notes about the step"),
	"ExplicitInstructions":      mcptools.StringField("Explicit instructions for this step"),
	"Media":                     mcptools.ArrayType(mcptools.SchemaObject(recipeMediaSchema)),
	"Products":                  mcptools.ArrayType(mcptools.SchemaObject(recipeStepProductsSchema)),
	"Instruments":               mcptools.ArrayType(mcptools.SchemaObject(recipeStepInstrumentsSchema)),
	"Vessels":                   mcptools.ArrayType(mcptools.SchemaObject(recipeStepVesselsSchema)),
	"CompletionConditions":      mcptools.ArrayType(mcptools.SchemaObject(recipeStepCompletionConditionsSchema)),
	"Ingredients":               mcptools.ArrayType(mcptools.SchemaObject(recipeStepIngredientsSchema)),
	fieldPreparation:            mcptools.ObjectType(validPreparationsSchema),
	fieldIndex:                  mcptools.UintField("The index of the step within the recipe"),
	fieldOptional:               mcptools.BoolField("Whether this step is optional"),
	"StartTimerAutomatically":   mcptools.BoolField("Whether to start a timer automatically for this step"),
	"MinEstimatedTimeInSeconds": mcptools.UintField("Minimum estimated time in seconds (optional)"),
	"MaxEstimatedTimeInSeconds": mcptools.UintField("Maximum estimated time in seconds (optional)"),
	"MinTemperatureInCelsius":   mcptools.FloatField("Minimum temperature in Celsius (optional)"),
	"MaxTemperatureInCelsius":   mcptools.FloatField("Maximum temperature in Celsius (optional)"),
}

var getRecipeStepTool = &sdkmcp.Tool{
	Name:        "GetRecipeStep",
	Description: "Get a recipe step by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID:     mcptools.StringField("The ID of the recipe"),
		fieldRecipeStepID: mcptools.StringField("The ID of the recipe step to get"),
	}),
	OutputSchema: mcptools.SchemaObject(recipeStepsSchema),
}

func (t *Tools) GetRecipeStep() sdkmcp.ToolHandlerFor[*GetRecipeStepInvocation, *mealplanning.RecipeStep] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipeStepInvocation) (*sdkmcp.CallToolResult, *mealplanning.RecipeStep, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetRecipeStep(ctx, x.RecipeID, x.RecipeStepID)
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

var getRecipeStepsTool = &sdkmcp.Tool{
	Name:        "GetRecipeSteps",
	Description: "Get recipe steps with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID:        mcptools.StringField("The ID of the recipe"),
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(recipeStepsSchema)),
	}),
}

func (t *Tools) GetRecipeSteps() sdkmcp.ToolHandlerFor[*GetRecipeStepsInvocation, *GetRecipeStepsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipeStepsInvocation) (*sdkmcp.CallToolResult, *GetRecipeStepsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetRecipeSteps(ctx, x.RecipeID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipeStepsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
