package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetRecipePrepTaskInput is what GetRecipePrepTask takes.
type GetRecipePrepTaskInput struct {
	RecipeID         string `json:"recipeID"         jsonschema:"The identifier of the recipe"`
	RecipePrepTaskID string `json:"recipePrepTaskID" jsonschema:"The identifier of the prep task to read"`
}

var getRecipePrepTaskTool = &sdkmcp.Tool{
	Name:        "GetRecipePrepTask",
	Description: "Get a recipe prep task by its ID",
	Annotations: readOnly(),
}

// GetRecipePrepTask reads one prep task of a recipe.
func (t *Tools) GetRecipePrepTask(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipePrepTaskInput) (*sdkmcp.CallToolResult, *mealplanning.RecipePrepTask, error) {
	ctx, err := t.begin(ctx, req, getRecipePrepTaskTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadRecipePrepTask(ctx, in.RecipeID, in.RecipePrepTaskID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

// GetRecipePrepTasksInput is what GetRecipePrepTasks takes.
type GetRecipePrepTasksInput struct {
	Filter   *filtering.QueryFilter `json:"filter,omitempty" jsonschema:"The page to read; absent reads the first page, oldest first"`
	RecipeID string                 `json:"recipeID"         jsonschema:"The identifier of the recipe"`
}

var getRecipePrepTasksTool = &sdkmcp.Tool{
	Name:        "GetRecipePrepTasks",
	Description: "Page the prep tasks of a recipe",
	Annotations: readOnly(),
}

// GetRecipePrepTasks pages the prep tasks of a recipe.
func (t *Tools) GetRecipePrepTasks(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipePrepTasksInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.RecipePrepTask], error) {
	ctx, err := t.begin(ctx, req, getRecipePrepTasksTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ListRecipePrepTask(ctx, in.RecipeID, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
