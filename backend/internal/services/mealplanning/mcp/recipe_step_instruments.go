package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipeStepInstrumentInvocation struct {
		RecipeID               string `jsonschema:"description=The recipe ID"`
		RecipeStepID           string `jsonschema:"description=The recipe step ID"`
		RecipeStepInstrumentID string `jsonschema:"description=The recipe step instrument ID"`
	}
)

var recipeStepInstrumentsSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the recipe step instrument"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the recipe step instrument was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the recipe step instrument was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the recipe step instrument was soft deleted"),
	fieldBelongsToRecipeStep:    mcptools.StringField("The ID of the recipe step this instrument belongs to"),
	mcptools.FieldName:          mcptools.StringField("Name of the instrument"),
	fieldNotes:                  mcptools.StringField("Notes about the instrument"),
	"Instrument":                mcptools.ObjectType(validInstrumentsSchema),
	fieldRecipeStepProductID:    mcptools.StringField("The ID of the recipe step product this instrument is associated with, if any"),
	fieldMinQuantity:            mcptools.UintField("Minimum quantity of this instrument (required)"),
	fieldMaxQuantity:            mcptools.UintField("Maximum quantity of this instrument (optional)"),
	"OptionIndex":               mcptools.UintField("The option index for this instrument"),
	"PreferenceRank":            mcptools.UintField("The preference rank for this instrument (0-255)"),
	fieldOptional:               mcptools.BoolField("Whether this instrument is optional"),
}

var getRecipeStepInstrumentTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepInstrument",
	Description: "Get a recipe step instrument by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID:            mcptools.StringField("The ID of the recipe"),
		fieldRecipeStepID:        mcptools.StringField("The ID of the recipe step"),
		"RecipeStepInstrumentID": mcptools.StringField("The ID of the recipe step instrument to get"),
	}),
	OutputSchema: mcptools.SchemaObject(recipeStepInstrumentsSchema),
}

func (t *Tools) GetRecipeStepInstrument() sdkmcp.ToolHandlerFor[*GetRecipeStepInstrumentInvocation, *mealplanning.RecipeStepInstrument] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipeStepInstrumentInvocation) (*sdkmcp.CallToolResult, *mealplanning.RecipeStepInstrument, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetRecipeStepInstrument(ctx, x.RecipeID, x.RecipeStepID, x.RecipeStepInstrumentID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetRecipeStepInstrumentsInvocation struct {
		Filter       *filtering.QueryFilter
		RecipeID     string
		RecipeStepID string
	}

	GetRecipeStepInstrumentsResult struct {
		Results []*mealplanning.RecipeStepInstrument
	}
)

var getRecipeStepInstrumentsTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepInstruments",
	Description: "Get recipe step instruments with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID:        mcptools.StringField("The ID of the recipe"),
		fieldRecipeStepID:    mcptools.StringField("The ID of the recipe step"),
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(recipeStepInstrumentsSchema)),
	}),
}

func (t *Tools) GetRecipeStepInstruments() sdkmcp.ToolHandlerFor[*GetRecipeStepInstrumentsInvocation, *GetRecipeStepInstrumentsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipeStepInstrumentsInvocation) (*sdkmcp.CallToolResult, *GetRecipeStepInstrumentsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetRecipeStepInstruments(ctx, x.RecipeID, x.RecipeStepID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipeStepInstrumentsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
