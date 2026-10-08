package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetRecipeStepVesselInput is what GetRecipeStepVessel takes.
type GetRecipeStepVesselInput struct {
	RecipeID           string `json:"recipeID"           jsonschema:"The identifier of the recipe"`
	RecipeStepID       string `json:"recipeStepID"       jsonschema:"The identifier of the step"`
	RecipeStepVesselID string `json:"recipeStepVesselID" jsonschema:"The identifier of the vessel to read"`
}

var getRecipeStepVesselTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepVessel",
	Description: "Get a vessel of a recipe step by its ID",
	Annotations: readOnly(),
}

// GetRecipeStepVessel reads one vessel of a recipe step.
func (t *Tools) GetRecipeStepVessel(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipeStepVesselInput) (*sdkmcp.CallToolResult, *mealplanning.RecipeStepVessel, error) {
	ctx, err := t.begin(ctx, req, getRecipeStepVesselTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadRecipeStepVessel(ctx, in.RecipeID, in.RecipeStepID, in.RecipeStepVesselID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

// GetRecipeStepVesselsInput is what GetRecipeStepVessels takes.
type GetRecipeStepVesselsInput struct {
	Filter       *filtering.QueryFilter `json:"filter,omitempty" jsonschema:"The page to read; absent reads the first page, oldest first"`
	RecipeID     string                 `json:"recipeID"         jsonschema:"The identifier of the recipe"`
	RecipeStepID string                 `json:"recipeStepID"     jsonschema:"The identifier of the step"`
}

var getRecipeStepVesselsTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepVessels",
	Description: "Page the vessels of a recipe step",
	Annotations: readOnly(),
}

// GetRecipeStepVessels pages the vessels of a recipe step.
func (t *Tools) GetRecipeStepVessels(ctx context.Context, req *sdkmcp.CallToolRequest, in GetRecipeStepVesselsInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.RecipeStepVessel], error) {
	ctx, err := t.begin(ctx, req, getRecipeStepVesselsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ListRecipeStepVessels(ctx, in.RecipeID, in.RecipeStepID, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
