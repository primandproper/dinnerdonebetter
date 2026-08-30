package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v13/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipeStepInstrumentInvocation struct {
		RecipeID               string `jsonschema:"description=The recipe ID"`
		RecipeStepID           string `jsonschema:"description=The recipe step ID"`
		RecipeStepInstrumentID string `jsonschema:"description=The recipe step instrument ID"`
	}
)

var recipeStepInstrumentsSchema = map[string]any{
	"ID":                     stringField("The ID of the recipe step instrument"),
	fieldCreatedAt:           timestampField("When the recipe step instrument was created"),
	fieldLastUpdatedAt:       timestampField("When the recipe step instrument was last updated"),
	fieldArchivedAt:          timestampField("When the recipe step instrument was soft deleted"),
	fieldBelongsToRecipeStep: stringField("The ID of the recipe step this instrument belongs to"),
	fieldName:                stringField("Name of the instrument"),
	fieldNotes:               stringField("Notes about the instrument"),
	"Instrument":             objectType(validInstrumentsSchema),
	fieldRecipeStepProductID: stringField("The ID of the recipe step product this instrument is associated with, if any"),
	fieldMinQuantity:         uintField("Minimum quantity of this instrument (required)"),
	fieldMaxQuantity:         uintField("Maximum quantity of this instrument (optional)"),
	"OptionIndex":            uintField("The option index for this instrument"),
	"PreferenceRank":         uintField("The preference rank for this instrument (0-255)"),
	fieldOptional:            boolField("Whether this instrument is optional"),
}

var getRecipeStepInstrumentTool = &mcp.Tool{
	Name:        "GetRecipeStepInstrument",
	Description: "Get a recipe step instrument by it's ID",
	InputSchema: schemaObject(map[string]any{
		fieldRecipeID:            stringField("The ID of the recipe"),
		fieldRecipeStepID:        stringField("The ID of the recipe step"),
		"RecipeStepInstrumentID": stringField("The ID of the recipe step instrument to get"),
	}),
	OutputSchema: schemaObject(recipeStepInstrumentsSchema),
}

func (h *mcpToolManager) GetRecipeStepInstrument() mcp.ToolHandlerFor[*GetRecipeStepInstrumentInvocation, *mealplanning.RecipeStepInstrument] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipeStepInstrumentInvocation) (*mcp.CallToolResult, *mealplanning.RecipeStepInstrument, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetRecipeStepInstrument(ctx, x.RecipeID, x.RecipeStepID, x.RecipeStepInstrumentID)
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

var getRecipeStepInstrumentsTool = &mcp.Tool{
	Name:        "GetRecipeStepInstruments",
	Description: "Get recipe step instruments with optional filtering",
	InputSchema: schemaObject(map[string]any{
		fieldRecipeID:     stringField("The ID of the recipe"),
		fieldRecipeStepID: stringField("The ID of the recipe step"),
		fieldFilter:       filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		fieldResults: arrayType(schemaObject(recipeStepInstrumentsSchema)),
	}),
}

func (h *mcpToolManager) GetRecipeStepInstruments() mcp.ToolHandlerFor[*GetRecipeStepInstrumentsInvocation, *GetRecipeStepInstrumentsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipeStepInstrumentsInvocation) (*mcp.CallToolResult, *GetRecipeStepInstrumentsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetRecipeStepInstruments(ctx, x.RecipeID, x.RecipeStepID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipeStepInstrumentsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
