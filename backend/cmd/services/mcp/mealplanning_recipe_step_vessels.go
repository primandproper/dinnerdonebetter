package main

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v3/database/filtering"

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
	"ID":                   stringField("The ID of the recipe step vessel"),
	"CreatedAt":            timestampField("When the recipe step vessel was created"),
	"LastUpdatedAt":        timestampField("When the recipe step vessel was last updated"),
	"ArchivedAt":           timestampField("When the recipe step vessel was soft deleted"),
	"BelongsToRecipeStep":  stringField("The ID of the recipe step this vessel belongs to"),
	"Name":                 stringField("Name of the vessel"),
	"Notes":                stringField("Notes about the vessel"),
	"Vessel":               objectType(validVesselsSchema),
	"RecipeStepProductID":  stringField("The ID of the recipe step product this vessel is associated with, if any"),
	"MinQuantity":          uintField("Minimum quantity of this vessel (required)"),
	"MaxQuantity":          uintField("Maximum quantity of this vessel (optional)"),
	"VesselPreposition":    stringField("The preposition to use with the vessel (e.g., 'in', 'on', 'over')"),
	"UnavailableAfterStep": boolField("Whether this vessel becomes unavailable after this step"),
}

var getRecipeStepVesselTool = &mcp.Tool{
	Name:        "GetRecipeStepVessel",
	Description: "Get a recipe step vessel by it's ID",
	InputSchema: schemaObject(map[string]any{
		"RecipeID":           stringField("The ID of the recipe"),
		"RecipeStepID":       stringField("The ID of the recipe step"),
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
		"RecipeID":     stringField("The ID of the recipe"),
		"RecipeStepID": stringField("The ID of the recipe step"),
		"Filter":       queryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(recipeStepVesselsSchema)),
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
