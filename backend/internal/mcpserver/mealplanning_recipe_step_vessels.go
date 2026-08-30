package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v13/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipeStepVesselInvocation struct {
		RecipeID           string `jsonschema:"description=The recipe ID"`
		RecipeStepID       string `jsonschema:"description=The recipe step ID"`
		RecipeStepVesselID string `jsonschema:"description=The recipe step vessel ID"`
	}
)

var recipeStepVesselsSchema = map[string]any{
	"ID":                     stringField("The ID of the recipe step vessel"),
	fieldCreatedAt:           timestampField("When the recipe step vessel was created"),
	fieldLastUpdatedAt:       timestampField("When the recipe step vessel was last updated"),
	fieldArchivedAt:          timestampField("When the recipe step vessel was soft deleted"),
	fieldBelongsToRecipeStep: stringField("The ID of the recipe step this vessel belongs to"),
	fieldName:                stringField("Name of the vessel"),
	fieldNotes:               stringField("Notes about the vessel"),
	"Vessel":                 objectType(validVesselsSchema),
	fieldRecipeStepProductID: stringField("The ID of the recipe step product this vessel is associated with, if any"),
	fieldMinQuantity:         uintField("Minimum quantity of this vessel (required)"),
	fieldMaxQuantity:         uintField("Maximum quantity of this vessel (optional)"),
	"VesselPreposition":      stringField("The preposition to use with the vessel (e.g., 'in', 'on', 'over')"),
	"UnavailableAfterStep":   boolField("Whether this vessel becomes unavailable after this step"),
}

var getRecipeStepVesselTool = &mcp.Tool{
	Name:        "GetRecipeStepVessel",
	Description: "Get a recipe step vessel by it's ID",
	InputSchema: schemaObject(map[string]any{
		fieldRecipeID:        stringField("The ID of the recipe"),
		fieldRecipeStepID:    stringField("The ID of the recipe step"),
		"RecipeStepVesselID": stringField("The ID of the recipe step vessel to get"),
	}),
	OutputSchema: schemaObject(recipeStepVesselsSchema),
}

func (h *mcpToolManager) GetRecipeStepVessel() mcp.ToolHandlerFor[*GetRecipeStepVesselInvocation, *mealplanning.RecipeStepVessel] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipeStepVesselInvocation) (*mcp.CallToolResult, *mealplanning.RecipeStepVessel, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetRecipeStepVessel(ctx, x.RecipeID, x.RecipeStepID, x.RecipeStepVesselID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetRecipeStepVesselsInvocation struct {
		Filter       *filtering.QueryFilter
		RecipeID     string
		RecipeStepID string
	}

	GetRecipeStepVesselsResult struct {
		Results []*mealplanning.RecipeStepVessel
	}
)

var getRecipeStepVesselsTool = &mcp.Tool{
	Name:        "GetRecipeStepVessels",
	Description: "Get recipe step vessels with optional filtering",
	InputSchema: schemaObject(map[string]any{
		fieldRecipeID:     stringField("The ID of the recipe"),
		fieldRecipeStepID: stringField("The ID of the recipe step"),
		fieldFilter:       filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		fieldResults: arrayType(schemaObject(recipeStepVesselsSchema)),
	}),
}

func (h *mcpToolManager) GetRecipeStepVessels() mcp.ToolHandlerFor[*GetRecipeStepVesselsInvocation, *GetRecipeStepVesselsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetRecipeStepVesselsInvocation) (*mcp.CallToolResult, *GetRecipeStepVesselsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetRecipeStepVessels(ctx, x.RecipeID, x.RecipeStepID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipeStepVesselsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
