package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetValidIngredientStateInput is what GetValidIngredientState takes.
type GetValidIngredientStateInput struct {
	ValidIngredientStateID string `json:"validIngredientStateID" jsonschema:"The identifier of the ingredient state to read"`
}

var getValidIngredientStateTool = &sdkmcp.Tool{
	Name:        "GetValidIngredientState",
	Description: "Get a valid ingredient state by its ID",
	Annotations: readOnly(),
}

// GetValidIngredientState reads one ingredient state from the catalog.
func (t *Tools) GetValidIngredientState(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidIngredientStateInput) (*sdkmcp.CallToolResult, *mealplanning.ValidIngredientState, error) {
	ctx, err := t.begin(ctx, req, getValidIngredientStateTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadValidIngredientState(ctx, in.ValidIngredientStateID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

var searchForValidIngredientStatesTool = &sdkmcp.Tool{
	Name:        "SearchForValidIngredientStates",
	Description: "Search the valid ingredient states by name",
	Annotations: readOnly(),
}

// SearchForValidIngredientStates pages the ingredient states matching a query.
func (t *Tools) SearchForValidIngredientStates(ctx context.Context, req *sdkmcp.CallToolRequest, in SearchInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidIngredientState], error) {
	ctx, err := t.begin(ctx, req, searchForValidIngredientStatesTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.SearchValidIngredientStates(ctx, in.Query, useDatabaseSearch, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
