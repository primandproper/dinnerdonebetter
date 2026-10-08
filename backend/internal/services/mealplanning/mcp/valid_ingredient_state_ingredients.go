package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidIngredientStateIngredientInvocation struct {
		ValidIngredientStateIngredientID string `jsonschema:"description=The ingredient state ingredient ID"`
	}
)

var validIngredientStateIngredientsSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the valid ingredient state ingredient"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the valid ingredient state ingredient was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the valid ingredient state ingredient was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the valid ingredient state ingredient was soft deleted"),
	fieldNotes:                  mcptools.StringField("Notes about the ingredient state ingredient"),
	"IngredientState":           mcptools.ObjectType(validIngredientStatesSchema),
	fieldIngredient:             mcptools.ObjectType(validIngredientsSchema),
}

var getValidIngredientStateIngredientTool = &sdkmcp.Tool{
	Name:        "GetValidIngredientStateIngredient",
	Description: "Get a valid ingredient state ingredient by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		"ValidIngredientStateIngredientID": mcptools.StringField("The ID of the valid ingredient state ingredient to get"),
	}),
	OutputSchema: mcptools.SchemaObject(validIngredientStateIngredientsSchema),
}

func (t *Tools) GetValidIngredientStateIngredient() sdkmcp.ToolHandlerFor[*GetValidIngredientStateIngredientInvocation, *mealplanning.ValidIngredientStateIngredient] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidIngredientStateIngredientInvocation) (*sdkmcp.CallToolResult, *mealplanning.ValidIngredientStateIngredient, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetValidIngredientStateIngredient(ctx, x.ValidIngredientStateIngredientID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetValidIngredientStateIngredientsInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetValidIngredientStateIngredientsResult struct {
		Results []*mealplanning.ValidIngredientStateIngredient
	}
)

var getValidIngredientStateIngredientsTool = &sdkmcp.Tool{
	Name:        "GetValidIngredientStateIngredients",
	Description: "Get valid ingredient state ingredients with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validIngredientStateIngredientsSchema)),
	}),
}

func (t *Tools) GetValidIngredientStateIngredients() sdkmcp.ToolHandlerFor[*GetValidIngredientStateIngredientsInvocation, *GetValidIngredientStateIngredientsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidIngredientStateIngredientsInvocation) (*sdkmcp.CallToolResult, *GetValidIngredientStateIngredientsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetValidIngredientStateIngredients(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidIngredientStateIngredientsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
