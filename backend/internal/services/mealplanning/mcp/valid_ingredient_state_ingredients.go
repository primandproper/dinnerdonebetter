package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetValidIngredientStateIngredientInput is what GetValidIngredientStateIngredient takes.
type GetValidIngredientStateIngredientInput struct {
	ValidIngredientStateIngredientID string `json:"validIngredientStateIngredientID" jsonschema:"The identifier of the ingredient state ingredient to read"`
}

var getValidIngredientStateIngredientTool = &sdkmcp.Tool{
	Name:        "GetValidIngredientStateIngredient",
	Description: "Get a valid ingredient state ingredient by its ID",
	Annotations: readOnly(),
}

// GetValidIngredientStateIngredient reads one ingredient state ingredient from the catalog.
func (t *Tools) GetValidIngredientStateIngredient(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidIngredientStateIngredientInput) (*sdkmcp.CallToolResult, *mealplanning.ValidIngredientStateIngredient, error) {
	ctx, err := t.begin(ctx, req, getValidIngredientStateIngredientTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadValidIngredientStateIngredient(ctx, in.ValidIngredientStateIngredientID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

var getValidIngredientStateIngredientsTool = &sdkmcp.Tool{
	Name:        "GetValidIngredientStateIngredients",
	Description: "Page the valid ingredient state ingredients",
	Annotations: readOnly(),
}

// GetValidIngredientStateIngredients pages the ingredient state ingredients.
func (t *Tools) GetValidIngredientStateIngredients(ctx context.Context, req *sdkmcp.CallToolRequest, in FilterInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient], error) {
	ctx, err := t.begin(ctx, req, getValidIngredientStateIngredientsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ListValidIngredientStateIngredients(ctx, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
