package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidPrepTaskConfigInvocation struct {
		ValidPrepTaskConfigID string `jsonschema:"description=The prep task config ID"`
	}
)

var validPrepTaskConfigsSchema = map[string]any{
	"ID":                                mcptools.StringField("The ID of the valid prep task config"),
	mcptools.FieldCreatedAt:             mcptools.TimestampField("When the valid prep task config was created"),
	mcptools.FieldLastUpdatedAt:         mcptools.TimestampField("When the valid prep task config was last updated"),
	mcptools.FieldArchivedAt:            mcptools.TimestampField("When the valid prep task config was soft deleted"),
	"MinStorageDurationInSeconds":       mcptools.UintField("Minimum storage duration in seconds (required)"),
	"MaxStorageDurationInSeconds":       mcptools.UintField("Maximum storage duration in seconds (optional)"),
	fieldMinStorageTemperatureInCelsius: mcptools.FloatField("Minimum storage temperature in Celsius (optional)"),
	fieldMaxStorageTemperatureInCelsius: mcptools.FloatField("Maximum storage temperature in Celsius (optional)"),
	"StorageType":                       mcptools.StringField("The type of storage container (e.g., covered, airtight, uncovered)"),
	fieldStorageInstructions:            mcptools.StringField("Instructions for how to store the prepped ingredient"),
	fieldNotes:                          mcptools.StringField("Additional notes about the prep task config"),
	"Source":                            mcptools.StringField("The source of this prep task config information"),
	fieldPreparation:                    mcptools.ObjectType(validPreparationsSchema),
	fieldIngredient:                     mcptools.ObjectType(validIngredientsSchema),
}

var getValidPrepTaskConfigTool = &sdkmcp.Tool{
	Name:        "GetValidPrepTaskConfig",
	Description: "Get a valid prep task config by its ID. A prep task config defines how long a prepped ingredient can be stored under specific conditions.",
	InputSchema: mcptools.SchemaObject(map[string]any{
		"ValidPrepTaskConfigID": mcptools.StringField("The ID of the valid prep task config to get"),
	}),
	OutputSchema: mcptools.SchemaObject(validPrepTaskConfigsSchema),
}

func (t *Tools) GetValidPrepTaskConfig() sdkmcp.ToolHandlerFor[*GetValidPrepTaskConfigInvocation, *mealplanning.ValidPrepTaskConfig] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidPrepTaskConfigInvocation) (*sdkmcp.CallToolResult, *mealplanning.ValidPrepTaskConfig, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := t.repo.GetValidPrepTaskConfig(ctx, x.ValidPrepTaskConfigID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

type (
	GetValidPrepTaskConfigsInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetValidPrepTaskConfigsResult struct {
		Results []*mealplanning.ValidPrepTaskConfig
	}
)

var getValidPrepTaskConfigsTool = &sdkmcp.Tool{
	Name:        "GetValidPrepTaskConfigs",
	Description: "Get valid prep task configs with optional filtering. Prep task configs define how long prepped ingredients can be stored.",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validPrepTaskConfigsSchema)),
	}),
}

func (t *Tools) GetValidPrepTaskConfigs() sdkmcp.ToolHandlerFor[*GetValidPrepTaskConfigsInvocation, *GetValidPrepTaskConfigsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidPrepTaskConfigsInvocation) (*sdkmcp.CallToolResult, *GetValidPrepTaskConfigsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetValidPrepTaskConfigs(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidPrepTaskConfigsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

type (
	GetValidPrepTaskConfigsByIngredientInvocation struct {
		Filter            *filtering.QueryFilter
		ValidIngredientID string `jsonschema:"description=The ingredient ID to filter by"`
	}
)

var getValidPrepTaskConfigsByIngredientTool = &sdkmcp.Tool{
	Name:        "GetValidPrepTaskConfigsByIngredient",
	Description: "Get valid prep task configs for a specific ingredient. Use this to find storage information for a particular ingredient.",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldValidIngredientID: mcptools.StringField("The ID of the ingredient to get prep task configs for"),
		mcptools.FieldFilter:   filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validPrepTaskConfigsSchema)),
	}),
}

func (t *Tools) GetValidPrepTaskConfigsByIngredient() sdkmcp.ToolHandlerFor[*GetValidPrepTaskConfigsByIngredientInvocation, *GetValidPrepTaskConfigsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidPrepTaskConfigsByIngredientInvocation) (*sdkmcp.CallToolResult, *GetValidPrepTaskConfigsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetValidPrepTaskConfigsForIngredient(ctx, x.ValidIngredientID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidPrepTaskConfigsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

type (
	GetValidPrepTaskConfigsByPreparationInvocation struct {
		Filter             *filtering.QueryFilter
		ValidPreparationID string `jsonschema:"description=The preparation ID to filter by"`
	}
)

var getValidPrepTaskConfigsByPreparationTool = &sdkmcp.Tool{
	Name:        "GetValidPrepTaskConfigsByPreparation",
	Description: "Get valid prep task configs for a specific preparation method. Use this to find storage information for ingredients prepared a certain way.",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldValidPreparationID: mcptools.StringField("The ID of the preparation to get prep task configs for"),
		mcptools.FieldFilter:    filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validPrepTaskConfigsSchema)),
	}),
}

func (t *Tools) GetValidPrepTaskConfigsByPreparation() sdkmcp.ToolHandlerFor[*GetValidPrepTaskConfigsByPreparationInvocation, *GetValidPrepTaskConfigsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidPrepTaskConfigsByPreparationInvocation) (*sdkmcp.CallToolResult, *GetValidPrepTaskConfigsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetValidPrepTaskConfigsForPreparation(ctx, x.ValidPreparationID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidPrepTaskConfigsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}

type (
	GetValidPrepTaskConfigsByIngredientAndPreparationInvocation struct {
		Filter             *filtering.QueryFilter
		ValidIngredientID  string `jsonschema:"description=The ingredient ID to filter by"`
		ValidPreparationID string `jsonschema:"description=The preparation ID to filter by"`
	}
)

var getValidPrepTaskConfigsByIngredientAndPreparationTool = &sdkmcp.Tool{
	Name:        "GetValidPrepTaskConfigsByIngredientAndPreparation",
	Description: "Get valid prep task configs for a specific ingredient and preparation combination. Use this to find exactly how long a specific prepped ingredient (e.g., diced onions) can be stored.",
	InputSchema: mcptools.SchemaObject(map[string]any{
		fieldValidIngredientID:  mcptools.StringField("The ID of the ingredient"),
		fieldValidPreparationID: mcptools.StringField("The ID of the preparation"),
		mcptools.FieldFilter:    filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(validPrepTaskConfigsSchema)),
	}),
}

func (t *Tools) GetValidPrepTaskConfigsByIngredientAndPreparation() sdkmcp.ToolHandlerFor[*GetValidPrepTaskConfigsByIngredientAndPreparationInvocation, *GetValidPrepTaskConfigsResult] {
	return func(ctx context.Context, req *sdkmcp.CallToolRequest, x *GetValidPrepTaskConfigsByIngredientAndPreparationInvocation) (*sdkmcp.CallToolResult, *GetValidPrepTaskConfigsResult, error) {
		if _, err := mcptools.AccountFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := t.repo.GetValidPrepTaskConfigsForIngredientAndPreparation(ctx, x.ValidIngredientID, x.ValidPreparationID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidPrepTaskConfigsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}
