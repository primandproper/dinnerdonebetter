package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetValidPreparationInput is what GetValidPreparation takes.
type GetValidPreparationInput struct {
	ValidPreparationID string `json:"validPreparationID" jsonschema:"The identifier of the preparation to read"`
}

var getValidPreparationTool = &sdkmcp.Tool{
	Name:        "GetValidPreparation",
	Description: "Get a valid preparation by its ID",
	Annotations: readOnly(),
}

// GetValidPreparation reads one preparation from the catalog.
func (t *Tools) GetValidPreparation(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidPreparationInput) (*sdkmcp.CallToolResult, *mealplanning.ValidPreparation, error) {
	ctx, err := t.begin(ctx, req, getValidPreparationTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadValidPreparation(ctx, in.ValidPreparationID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

var searchForValidPreparationsTool = &sdkmcp.Tool{
	Name:        "SearchForValidPreparations",
	Description: "Search the valid preparations by name",
	Annotations: readOnly(),
}

// SearchForValidPreparations pages the preparations matching a query.
func (t *Tools) SearchForValidPreparations(ctx context.Context, req *sdkmcp.CallToolRequest, in SearchInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidPreparation], error) {
	ctx, err := t.begin(ctx, req, searchForValidPreparationsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.SearchValidPreparations(ctx, in.Query, useDatabaseSearch, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
