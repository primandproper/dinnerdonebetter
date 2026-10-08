package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidIngredientPreparationInvocation struct {
		ValidIngredientPreparationID string `jsonschema:"description=The ingredient preparation ID"`
	}
)

var validIngredientPreparationsSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the valid ingredient preparation"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the valid ingredient preparation was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the valid ingredient preparation was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the valid ingredient preparation was soft deleted"),
	fieldNotes:                  mcptools.StringField("Notes about the ingredient preparation"),
	fieldPreparation:            mcptools.ObjectType(validPreparationsSchema),
	fieldIngredient:             mcptools.ObjectType(validIngredientsSchema),
}

var getValidIngredientPreparationTool = &sdkmcp.Tool{
	Name:        "GetValidIngredientPreparation",
	Description: "Get a valid ingredient preparation by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		"ValidIngredientPreparationID": mcptools.StringField("The ID of the valid ingredient preparation to get"),
	}),
	OutputSchema: mcptools.SchemaObject(validIngredientPreparationsSchema),
}

func (t *Tools) GetValidIngredientPreparation() sdkmcp.ToolHandlerFor[*GetValidIngredientPreparationInvocation, *mealplanning.ValidIngredientPreparation] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidIngredientPreparationInvocation) (*sdkmcp.CallToolResult, *mealplanning.ValidIngredientPreparation, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetValidIngredientPreparation(ctx, x.ValidIngredientPreparationID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetValidIngredientPreparationsInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetValidIngredientPreparationsResult struct {
		Results []*mealplanning.ValidIngredientPreparation
	}
)

var getValidIngredientPreparationsTool = &sdkmcp.Tool{
	Name:        "GetValidIngredientPreparations",
	Description: "Get valid ingredient preparations with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validIngredientPreparationsSchema)),
	}),
}

func (t *Tools) GetValidIngredientPreparations() sdkmcp.ToolHandlerFor[*GetValidIngredientPreparationsInvocation, *GetValidIngredientPreparationsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidIngredientPreparationsInvocation) (*sdkmcp.CallToolResult, *GetValidIngredientPreparationsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetValidIngredientPreparations(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidIngredientPreparationsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
