package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetValidIngredientPreparationInput is what GetValidIngredientPreparation takes.
type GetValidIngredientPreparationInput struct {
	ValidIngredientPreparationID string `json:"validIngredientPreparationID" jsonschema:"The identifier of the ingredient preparation to read"`
}

var getValidIngredientPreparationTool = &sdkmcp.Tool{
	Name:        "GetValidIngredientPreparation",
	Description: "Get a valid ingredient preparation by its ID",
	Annotations: readOnly(),
}

// GetValidIngredientPreparation reads one ingredient preparation from the catalog.
func (t *Tools) GetValidIngredientPreparation(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidIngredientPreparationInput) (*sdkmcp.CallToolResult, *mealplanning.ValidIngredientPreparation, error) {
	ctx, err := t.begin(ctx, req, getValidIngredientPreparationTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadValidIngredientPreparation(ctx, in.ValidIngredientPreparationID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

var getValidIngredientPreparationsTool = &sdkmcp.Tool{
	Name:        "GetValidIngredientPreparations",
	Description: "Page the valid ingredient preparations",
	Annotations: readOnly(),
}

// GetValidIngredientPreparations pages the ingredient preparations.
func (t *Tools) GetValidIngredientPreparations(ctx context.Context, req *sdkmcp.CallToolRequest, in FilterInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], error) {
	ctx, err := t.begin(ctx, req, getValidIngredientPreparationsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ListValidIngredientPreparations(ctx, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
