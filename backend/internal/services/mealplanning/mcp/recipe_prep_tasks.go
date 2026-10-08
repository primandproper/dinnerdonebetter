package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipePrepTaskInvocation struct {
		RecipeID         string `jsonschema:"description=The recipe ID"`
		RecipePrepTaskID string `jsonschema:"description=The recipe prep task ID"`
	}
)

var recipePrepTaskStepSchema = map[string]any{
	"ID":                      mcptools.StringField("The ID of the recipe prep task step"),
	fieldBelongsToRecipeStep:  mcptools.StringField("The ID of the recipe step this prep task step belongs to"),
	"BelongsToRecipePrepTask": mcptools.StringField("The ID of the recipe prep task this step belongs to"),
	"SatisfiesRecipeStep":     mcptools.BoolField("Whether this prep task step satisfies the recipe step"),
}

var recipePrepTasksSchema = map[string]any{
	"ID":                                 mcptools.StringField("The ID of the recipe prep task"),
	mcptools.FieldCreatedAt:              mcptools.TimestampField("When the recipe prep task was created"),
	mcptools.FieldLastUpdatedAt:          mcptools.TimestampField("When the recipe prep task was last updated"),
	mcptools.FieldArchivedAt:             mcptools.TimestampField("When the recipe prep task was soft deleted"),
	fieldBelongsToRecipe:                 mcptools.StringField("The ID of the recipe this prep task belongs to"),
	mcptools.FieldName:                   mcptools.StringField("Name of the prep task"),
	mcptools.FieldDescription:            mcptools.StringField("Description of the prep task"),
	fieldNotes:                           mcptools.StringField("Notes about the prep task"),
	"StorageType":                        mcptools.StringField("The storage type for the prep task (e.g., 'covered', 'uncovered', 'on a wire rack')"),
	"ExplicitStorageInstructions":        mcptools.StringField("Explicit storage instructions for the prep task"),
	fieldMinStorageTemperatureInCelsius:  mcptools.FloatField("Minimum storage temperature in Celsius (optional)"),
	fieldMaxStorageTemperatureInCelsius:  mcptools.FloatField("Maximum storage temperature in Celsius (optional)"),
	"MinTimeBufferBeforeRecipeInSeconds": mcptools.UintField("Minimum time buffer before recipe in seconds (required)"),
	"MaxTimeBufferBeforeRecipeInSeconds": mcptools.UintField("Maximum time buffer before recipe in seconds (optional)"),
	fieldOptional:                        mcptools.BoolField("Whether this prep task is optional"),
	"TaskSteps":                          mcptools.ArrayType(mcptools.SchemaObject(recipePrepTaskStepSchema)),
}

var getRecipePrepTaskTool = &sdkmcp.Tool{
	Name:        "GetRecipePrepTask",
	Description: "Get a recipe prep task by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID:      mcptools.StringField("The ID of the recipe"),
		"RecipePrepTaskID": mcptools.StringField("The ID of the recipe prep task to get"),
	}),
	OutputSchema: mcptools.SchemaObject(recipePrepTasksSchema),
}

func (t *Tools) GetRecipePrepTask() sdkmcp.ToolHandlerFor[*GetRecipePrepTaskInvocation, *mealplanning.RecipePrepTask] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipePrepTaskInvocation) (*sdkmcp.CallToolResult, *mealplanning.RecipePrepTask, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetRecipePrepTask(ctx, x.RecipeID, x.RecipePrepTaskID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetRecipePrepTasksInvocation struct {
		Filter   *filtering.QueryFilter
		RecipeID string
	}

	GetRecipePrepTasksResult struct {
		Results []*mealplanning.RecipePrepTask
	}
)

var getRecipePrepTasksTool = &sdkmcp.Tool{
	Name:        "GetRecipePrepTasks",
	Description: "Get recipe prep tasks with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID:        mcptools.StringField("The ID of the recipe"),
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(recipePrepTasksSchema)),
	}),
}

func (t *Tools) GetRecipePrepTasks() sdkmcp.ToolHandlerFor[*GetRecipePrepTasksInvocation, *GetRecipePrepTasksResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipePrepTasksInvocation) (*sdkmcp.CallToolResult, *GetRecipePrepTasksResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetRecipePrepTasks(ctx, x.RecipeID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipePrepTasksResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
