package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetRecipeStepVesselInvocation struct {
		RecipeID           string `jsonschema:"description=The recipe ID"`
		RecipeStepID       string `jsonschema:"description=The recipe step ID"`
		RecipeStepVesselID string `jsonschema:"description=The recipe step vessel ID"`
	}
)

var recipeStepVesselsSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the recipe step vessel"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the recipe step vessel was created"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the recipe step vessel was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the recipe step vessel was soft deleted"),
	fieldBelongsToRecipeStep:    mcptools.StringField("The ID of the recipe step this vessel belongs to"),
	mcptools.FieldName:          mcptools.StringField("Name of the vessel"),
	fieldNotes:                  mcptools.StringField("Notes about the vessel"),
	"Vessel":                    mcptools.ObjectType(validVesselsSchema),
	fieldRecipeStepProductID:    mcptools.StringField("The ID of the recipe step product this vessel is associated with, if any"),
	fieldMinQuantity:            mcptools.UintField("Minimum quantity of this vessel (required)"),
	fieldMaxQuantity:            mcptools.UintField("Maximum quantity of this vessel (optional)"),
	"VesselPreposition":         mcptools.StringField("The preposition to use with the vessel (e.g., 'in', 'on', 'over')"),
	"UnavailableAfterStep":      mcptools.BoolField("Whether this vessel becomes unavailable after this step"),
}

var getRecipeStepVesselTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepVessel",
	Description: "Get a recipe step vessel by it's ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID:        mcptools.StringField("The ID of the recipe"),
		fieldRecipeStepID:    mcptools.StringField("The ID of the recipe step"),
		"RecipeStepVesselID": mcptools.StringField("The ID of the recipe step vessel to get"),
	}),
	OutputSchema: mcptools.SchemaObject(recipeStepVesselsSchema),
}

func (t *Tools) GetRecipeStepVessel() sdkmcp.ToolHandlerFor[*GetRecipeStepVesselInvocation, *mealplanning.RecipeStepVessel] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipeStepVesselInvocation) (*sdkmcp.CallToolResult, *mealplanning.RecipeStepVessel, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetRecipeStepVessel(ctx, x.RecipeID, x.RecipeStepID, x.RecipeStepVesselID)
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

var getRecipeStepVesselsTool = &sdkmcp.Tool{
	Name:        "GetRecipeStepVessels",
	Description: "Get recipe step vessels with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldRecipeID:        mcptools.StringField("The ID of the recipe"),
		fieldRecipeStepID:    mcptools.StringField("The ID of the recipe step"),
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(recipeStepVesselsSchema)),
	}),
}

func (t *Tools) GetRecipeStepVessels() sdkmcp.ToolHandlerFor[*GetRecipeStepVesselsInvocation, *GetRecipeStepVesselsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetRecipeStepVesselsInvocation) (*sdkmcp.CallToolResult, *GetRecipeStepVesselsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetRecipeStepVessels(ctx, x.RecipeID, x.RecipeStepID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetRecipeStepVesselsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

//
