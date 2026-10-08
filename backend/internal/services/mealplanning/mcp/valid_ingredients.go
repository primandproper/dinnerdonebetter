package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetValidIngredientInput is what GetValidIngredient takes.
type GetValidIngredientInput struct {
	ValidIngredientID string `json:"validIngredientID" jsonschema:"The identifier of the ingredient to read"`
}

var getValidIngredientTool = &sdkmcp.Tool{
	Name:        "GetValidIngredient",
	Description: "Get a valid ingredient by its ID",
	Annotations: readOnly(),
}

// GetValidIngredient reads one ingredient from the catalog.
func (t *Tools) GetValidIngredient(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidIngredientInput) (*sdkmcp.CallToolResult, *mealplanning.ValidIngredient, error) {
	ctx, err := t.begin(ctx, req, getValidIngredientTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadValidIngredient(ctx, in.ValidIngredientID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

var searchForValidIngredientsTool = &sdkmcp.Tool{
	Name:        "SearchForValidIngredients",
	Description: "Search the valid ingredients by name",
	Annotations: readOnly(),
}

// SearchForValidIngredients pages the ingredients matching a query.
func (t *Tools) SearchForValidIngredients(ctx context.Context, req *sdkmcp.CallToolRequest, in SearchInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidIngredient], error) {
	ctx, err := t.begin(ctx, req, searchForValidIngredientsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.SearchValidIngredients(ctx, in.Query, useDatabaseSearch, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
