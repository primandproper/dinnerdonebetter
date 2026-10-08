package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetValidInstrumentInput is what GetValidInstrument takes.
type GetValidInstrumentInput struct {
	ValidInstrumentID string `json:"validInstrumentID" jsonschema:"The identifier of the instrument to read"`
}

var getValidInstrumentTool = &sdkmcp.Tool{
	Name:        "GetValidInstrument",
	Description: "Get a valid instrument by its ID",
	Annotations: readOnly(),
}

// GetValidInstrument reads one instrument from the catalog.
func (t *Tools) GetValidInstrument(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidInstrumentInput) (*sdkmcp.CallToolResult, *mealplanning.ValidInstrument, error) {
	ctx, err := t.begin(ctx, req, getValidInstrumentTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadValidInstrument(ctx, in.ValidInstrumentID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

var searchForValidInstrumentsTool = &sdkmcp.Tool{
	Name:        "SearchForValidInstruments",
	Description: "Search the valid instruments by name",
	Annotations: readOnly(),
}

// SearchForValidInstruments pages the instruments matching a query.
func (t *Tools) SearchForValidInstruments(ctx context.Context, req *sdkmcp.CallToolRequest, in SearchInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidInstrument], error) {
	ctx, err := t.begin(ctx, req, searchForValidInstrumentsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.SearchValidInstruments(ctx, in.Query, useDatabaseSearch, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
