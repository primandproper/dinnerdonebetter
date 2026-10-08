package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetValidPreparationInstrumentInput is what GetValidPreparationInstrument takes.
type GetValidPreparationInstrumentInput struct {
	ValidPreparationInstrumentID string `json:"validPreparationInstrumentID" jsonschema:"The identifier of the preparation instrument to read"`
}

var getValidPreparationInstrumentTool = &sdkmcp.Tool{
	Name:        "GetValidPreparationInstrument",
	Description: "Get a valid preparation instrument by its ID",
	Annotations: readOnly(),
}

// GetValidPreparationInstrument reads one preparation instrument from the catalog.
func (t *Tools) GetValidPreparationInstrument(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidPreparationInstrumentInput) (*sdkmcp.CallToolResult, *mealplanning.ValidPreparationInstrument, error) {
	ctx, err := t.begin(ctx, req, getValidPreparationInstrumentTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadValidPreparationInstrument(ctx, in.ValidPreparationInstrumentID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

var getValidPreparationInstrumentsTool = &sdkmcp.Tool{
	Name:        "GetValidPreparationInstruments",
	Description: "Page the valid preparation instruments",
	Annotations: readOnly(),
}

// GetValidPreparationInstruments pages the preparation instruments.
func (t *Tools) GetValidPreparationInstruments(ctx context.Context, req *sdkmcp.CallToolRequest, in FilterInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument], error) {
	ctx, err := t.begin(ctx, req, getValidPreparationInstrumentsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ListValidPreparationInstruments(ctx, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
