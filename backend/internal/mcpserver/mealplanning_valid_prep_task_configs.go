package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v10/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	GetValidPrepTaskConfigInvocation struct {
		ValidPrepTaskConfigID string `jsonschema:"description=The prep task config ID"`
	}
)

var validPrepTaskConfigsSchema = map[string]any{
	"ID":                             stringField("The ID of the valid prep task config"),
	"CreatedAt":                      timestampField("When the valid prep task config was created"),
	"LastUpdatedAt":                  timestampField("When the valid prep task config was last updated"),
	"ArchivedAt":                     timestampField("When the valid prep task config was soft deleted"),
	"MinStorageDurationInSeconds":    uintField("Minimum storage duration in seconds (required)"),
	"MaxStorageDurationInSeconds":    uintField("Maximum storage duration in seconds (optional)"),
	"MinStorageTemperatureInCelsius": floatField("Minimum storage temperature in Celsius (optional)"),
	"MaxStorageTemperatureInCelsius": floatField("Maximum storage temperature in Celsius (optional)"),
	"StorageType":                    stringField("The type of storage container (e.g., covered, airtight, uncovered)"),
	"StorageInstructions":            stringField("Instructions for how to store the prepped ingredient"),
	"Notes":                          stringField("Additional notes about the prep task config"),
	"Source":                         stringField("The source of this prep task config information"),
	"Preparation":                    objectType(validPreparationsSchema),
	"Ingredient":                     objectType(validIngredientsSchema),
}

var getValidPrepTaskConfigTool = &mcp.Tool{
	Name:        "GetValidPrepTaskConfig",
	Description: "Get a valid prep task config by its ID. A prep task config defines how long a prepped ingredient can be stored under specific conditions.",
	InputSchema: schemaObject(map[string]any{
		"ValidPrepTaskConfigID": stringField("The ID of the valid prep task config to get"),
	}),
	OutputSchema: schemaObject(validPrepTaskConfigsSchema),
}

func (h *mcpToolManager) GetValidPrepTaskConfig() mcp.ToolHandlerFor[*GetValidPrepTaskConfigInvocation, *mealplanning.ValidPrepTaskConfig] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidPrepTaskConfigInvocation) (*mcp.CallToolResult, *mealplanning.ValidPrepTaskConfig, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		result, err := h.mealplanningRepo.GetValidPrepTaskConfig(ctx, x.ValidPrepTaskConfigID)
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

var getValidPrepTaskConfigsTool = &mcp.Tool{
	Name:        "GetValidPrepTaskConfigs",
	Description: "Get valid prep task configs with optional filtering. Prep task configs define how long prepped ingredients can be stored.",
	InputSchema: schemaObject(map[string]any{
		"Filter": filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(validPrepTaskConfigsSchema)),
	}),
}

func (h *mcpToolManager) GetValidPrepTaskConfigs() mcp.ToolHandlerFor[*GetValidPrepTaskConfigsInvocation, *GetValidPrepTaskConfigsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidPrepTaskConfigsInvocation) (*mcp.CallToolResult, *GetValidPrepTaskConfigsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetValidPrepTaskConfigs(ctx, x.Filter)
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

var getValidPrepTaskConfigsByIngredientTool = &mcp.Tool{
	Name:        "GetValidPrepTaskConfigsByIngredient",
	Description: "Get valid prep task configs for a specific ingredient. Use this to find storage information for a particular ingredient.",
	InputSchema: schemaObject(map[string]any{
		"ValidIngredientID": stringField("The ID of the ingredient to get prep task configs for"),
		"Filter":            filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(validPrepTaskConfigsSchema)),
	}),
}

func (h *mcpToolManager) GetValidPrepTaskConfigsByIngredient() mcp.ToolHandlerFor[*GetValidPrepTaskConfigsByIngredientInvocation, *GetValidPrepTaskConfigsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidPrepTaskConfigsByIngredientInvocation) (*mcp.CallToolResult, *GetValidPrepTaskConfigsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetValidPrepTaskConfigsForIngredient(ctx, x.ValidIngredientID, x.Filter)
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

var getValidPrepTaskConfigsByPreparationTool = &mcp.Tool{
	Name:        "GetValidPrepTaskConfigsByPreparation",
	Description: "Get valid prep task configs for a specific preparation method. Use this to find storage information for ingredients prepared a certain way.",
	InputSchema: schemaObject(map[string]any{
		"ValidPreparationID": stringField("The ID of the preparation to get prep task configs for"),
		"Filter":             filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(validPrepTaskConfigsSchema)),
	}),
}

func (h *mcpToolManager) GetValidPrepTaskConfigsByPreparation() mcp.ToolHandlerFor[*GetValidPrepTaskConfigsByPreparationInvocation, *GetValidPrepTaskConfigsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidPrepTaskConfigsByPreparationInvocation) (*mcp.CallToolResult, *GetValidPrepTaskConfigsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetValidPrepTaskConfigsForPreparation(ctx, x.ValidPreparationID, x.Filter)
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

var getValidPrepTaskConfigsByIngredientAndPreparationTool = &mcp.Tool{
	Name:        "GetValidPrepTaskConfigsByIngredientAndPreparation",
	Description: "Get valid prep task configs for a specific ingredient and preparation combination. Use this to find exactly how long a specific prepped ingredient (e.g., diced onions) can be stored.",
	InputSchema: schemaObject(map[string]any{
		"ValidIngredientID":  stringField("The ID of the ingredient"),
		"ValidPreparationID": stringField("The ID of the preparation"),
		"Filter":             filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		"Results": arrayType(schemaObject(validPrepTaskConfigsSchema)),
	}),
}

func (h *mcpToolManager) GetValidPrepTaskConfigsByIngredientAndPreparation() mcp.ToolHandlerFor[*GetValidPrepTaskConfigsByIngredientAndPreparationInvocation, *GetValidPrepTaskConfigsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetValidPrepTaskConfigsByIngredientAndPreparationInvocation) (*mcp.CallToolResult, *GetValidPrepTaskConfigsResult, error) {
		if _, err := h.userFromRequest(req); err != nil {
			return nil, nil, err
		}

		results, err := h.mealplanningRepo.GetValidPrepTaskConfigsForIngredientAndPreparation(ctx, x.ValidIngredientID, x.ValidPreparationID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		out := &GetValidPrepTaskConfigsResult{}
		out.Results = results.Data
		return nil, out, nil
	}
}
