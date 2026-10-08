package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetRecipeStepInstrumentInput is what GetRecipeStepInstrument takes.
type GetRecipeStepInstrumentInput struct {
	RecipeID               string `json:"recipeID"               jsonschema:"The identifier of the recipe"`
	RecipeStepID           string `json:"recipeStepID"           jsonschema:"The identifier of the step"`
	RecipeStepInstrumentID string `json:"recipeStepInstrumentID" jsonschema:"The identifier of the instrument to read"`
}

var getRecipeStepInstrumentTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepInstrument",
	Description: "Get a instrument of a recipe step by its ID",
	Annotations: readOnly(),
}

// GetRecipeStepInstrument reads one instrument of a recipe step.
func (t *Tools) GetRecipeStepInstrument(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipeStepInstrumentInput) (*sdkmcp.CallToolResult, *mealplanning.RecipeStepInstrument, error) {
	ctx, err := t.begin(ctx, req, getRecipeStepInstrumentTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadRecipeStepInstrument(ctx, in.RecipeID, in.RecipeStepID, in.RecipeStepInstrumentID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

// GetRecipeStepInstrumentsInput is what GetRecipeStepInstruments takes.
type GetRecipeStepInstrumentsInput struct {
	Filter       *filtering.QueryFilter `json:"filter,omitempty" jsonschema:"The page to read; absent reads the first page, oldest first"`
	RecipeID     string                 `json:"recipeID"         jsonschema:"The identifier of the recipe"`
	RecipeStepID string                 `json:"recipeStepID"     jsonschema:"The identifier of the step"`
}

var getRecipeStepInstrumentsTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepInstruments",
	Description: "Page the instruments of a recipe step",
	Annotations: readOnly(),
}

// GetRecipeStepInstruments pages the instruments of a recipe step.
func (t *Tools) GetRecipeStepInstruments(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipeStepInstrumentsInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.RecipeStepInstrument], error) {
	ctx, err := t.begin(ctx, req, getRecipeStepInstrumentsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ListRecipeStepInstruments(ctx, in.RecipeID, in.RecipeStepID, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
