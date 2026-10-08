package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidInstrumentInvocation struct {
		ValidInstrumentID string `jsonschema:"description=The instrument ID"`
	}
)

var validInstrumentsSchema = map[string]any{
	"ID":                             mcptools.StringField("The ID of the valid instrument"),
	mcptools.FieldCreatedAt:          mcptools.TimestampField("When the valid instrument was created"),
	mcptools.FieldLastUpdatedAt:      mcptools.TimestampField("When the valid instrument was last updated"),
	mcptools.FieldArchivedAt:         mcptools.TimestampField("When the valid instrument was soft deleted"),
	mcptools.FieldName:               mcptools.StringField("Name of the instrument"),
	mcptools.FieldDescription:        mcptools.StringField("Description of the instrument"),
	fieldIconPath:                    mcptools.StringField("The URL for the icon for the item"),
	fieldPluralName:                  mcptools.StringField("The plural name for the instrument. So for an instrument named 'knife', this would be 'knives'"),
	fieldSlug:                        mcptools.StringField("An easy-to-use URL slug for the instrument"),
	"IncludeInGeneratedInstructions": mcptools.BoolField("Whether or not the valid instrument should be included in generated instructions"),
	"DisplayInSummaryLists":          mcptools.BoolField("Whether or not the valid instrument should be displayed in summary lists"),
	"UsableForStorage":               mcptools.BoolField("Whether or not the valid instrument is usable for storage"),
}

var getValidInstrumentTool = &sdkmcp.Tool{
	Name:        "GetValidInstrument",
	Description: "Get a valid instrument by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		"ValidInstrumentID": mcptools.StringField("The ID of the valid instrument to get"),
	}),
	OutputSchema: mcptools.SchemaObject(validInstrumentsSchema),
}

func (t *Tools) GetValidInstrument() sdkmcp.ToolHandlerFor[*GetValidInstrumentInvocation, *mealplanning.ValidInstrument] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidInstrumentInvocation) (*sdkmcp.CallToolResult, *mealplanning.ValidInstrument, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetValidInstrument(ctx, x.ValidInstrumentID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	SearchValidInstrumentsInvocation struct {
		Filter *filtering.QueryFilter
		Query  string `jsonschema_description:"The instrument name query"`
	}

	SearchValidInstrumentsResult struct {
		Results []*mealplanning.ValidInstrument
	}
)

var searchForValidInstrumentsTool = &sdkmcp.Tool{
	Name:        "SearchForValidInstruments",
	Description: "Search for valid instruments with optional filtering and query string",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
		mcptools.FieldQuery:  mcptools.StringField("The instrument name query"),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validInstrumentsSchema)),
	}),
}

func (t *Tools) SearchForValidInstruments() sdkmcp.ToolHandlerFor[*SearchValidInstrumentsInvocation, *SearchValidInstrumentsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *SearchValidInstrumentsInvocation) (*sdkmcp.CallToolResult, *SearchValidInstrumentsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.SearchForValidInstruments(ctx, x.Query, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &SearchValidInstrumentsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
