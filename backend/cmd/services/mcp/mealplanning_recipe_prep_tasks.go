package main

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v4/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipePrepTaskInvocation struct {
		RecipeID         string `jsonschema:"description=The recipe ID"`
		RecipePrepTaskID string `jsonschema:"description=The recipe prep task ID"`
	}
)

var recipePrepTaskStepSchema = map[string]any{
	"ID":                      stringField("The ID of the recipe prep task step"),
	"BelongsToRecipeStep":     stringField("The ID of the recipe step this prep task step belongs to"),
	"BelongsToRecipePrepTask": stringField("The ID of the recipe prep task this step belongs to"),
	"SatisfiesRecipeStep":     boolField("Whether this prep task step satisfies the recipe step"),
}

var recipePrepTasksSchema = map[string]any{
	"ID":                                 stringField("The ID of the recipe prep task"),
	"CreatedAt":                          timestampField("When the recipe prep task was created"),
	"LastUpdatedAt":                      timestampField("When the recipe prep task was last updated"),
	"ArchivedAt":                         timestampField("When the recipe prep task was soft deleted"),
	"BelongsToRecipe":                    stringField("The ID of the recipe this prep task belongs to"),
	"Name":                               stringField("Name of the prep task"),
	"Description":                        stringField("Description of the prep task"),
	"Notes":                              stringField("Notes about the prep task"),
	"StorageType":                        stringField("The storage type for the prep task (e.g., 'covered', 'uncovered', 'on a wire rack')"),
	"ExplicitStorageInstructions":        stringField("Explicit storage instructions for the prep task"),
	"MinStorageTemperatureInCelsius":     floatField("Minimum storage temperature in Celsius (optional)"),
	"MaxStorageTemperatureInCelsius":     floatField("Maximum storage temperature in Celsius (optional)"),
	"MinTimeBufferBeforeRecipeInSeconds": uintField("Minimum time buffer before recipe in seconds (required)"),
	"MaxTimeBufferBeforeRecipeInSeconds": uintField("Maximum time buffer before recipe in seconds (optional)"),
	"Optional":                           boolField("Whether this prep task is optional"),
	"TaskSteps":                          arrayType(schemaObject(recipePrepTaskStepSchema)),
}

var getRecipePrepTaskTool = &mcp.Tool{
	Name:        "GetRecipePrepTask",
	Description: "Get a recipe prep task by it's ID",
	InputSchema: schemaObject(map[string]any{
		"RecipeID":         stringField("The ID of the recipe"),
		"RecipePrepTaskID": stringField("The ID of the recipe prep task to get"),
	}),
	OutputSchema: schemaObject(recipePrepTasksSchema),
}

func (h *mcpToolManager) GetRecipePrepTask() mcp.ToolHandlerFor[*GetRecipePrepTaskInvocation, *mealplanning.RecipePrepTask] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipePrepTaskInvocation) (*mcp.CallToolResult, *mealplanning.RecipePrepTask, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetRecipePrepTask(ctx, x.RecipeID, x.RecipePrepTaskID)
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

var getRecipePrepTasksTool = &mcp.Tool{
	Name:        "GetRecipePrepTasks",
	Description: "Get recipe prep tasks with optional filtering",
	InputSchema: schemaObject(map[string]any{
		"RecipeID": stringField("The ID of the recipe"),
		"Filter":   queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(recipePrepTasksSchema)),
	}),
}

func (h *mcpToolManager) GetRecipePrepTasks() mcp.ToolHandlerFor[*GetRecipePrepTasksInvocation, *GetRecipePrepTasksResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipePrepTasksInvocation) (*mcp.CallToolResult, *GetRecipePrepTasksResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetRecipePrepTasks(ctx, x.RecipeID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipePrepTasksResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
