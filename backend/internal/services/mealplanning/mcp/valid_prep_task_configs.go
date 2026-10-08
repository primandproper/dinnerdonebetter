package mcp

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetValidPrepTaskConfigInput is what GetValidPrepTaskConfig takes.
type GetValidPrepTaskConfigInput struct {
	ValidPrepTaskConfigID string `json:"validPrepTaskConfigID" jsonschema:"The identifier of the prep task config to read"`
}

var getValidPrepTaskConfigTool = &sdkmcp.Tool{
	Name:        "GetValidPrepTaskConfig",
	Description: "Get a valid prep task config by its ID. A prep task config defines how long a prepped ingredient can be stored under specific conditions.",
	Annotations: readOnly(),
}

// GetValidPrepTaskConfig reads one prep task config.
func (t *Tools) GetValidPrepTaskConfig(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidPrepTaskConfigInput) (*sdkmcp.CallToolResult, *mealplanning.ValidPrepTaskConfig, error) {
	ctx, err := t.begin(ctx, req, getValidPrepTaskConfigTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ReadValidPrepTaskConfig(ctx, in.ValidPrepTaskConfigID)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

var getValidPrepTaskConfigsTool = &sdkmcp.Tool{
	Name:        "GetValidPrepTaskConfigs",
	Description: "Page the valid prep task configs. Prep task configs define how long prepped ingredients can be stored.",
	Annotations: readOnly(),
}

// GetValidPrepTaskConfigs pages the prep task configs.
func (t *Tools) GetValidPrepTaskConfigs(ctx context.Context, req *sdkmcp.CallToolRequest, in FilterInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
	ctx, err := t.begin(ctx, req, getValidPrepTaskConfigsTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.ListValidPrepTaskConfigs(ctx, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

// GetValidPrepTaskConfigsByIngredientInput is what GetValidPrepTaskConfigsByIngredient takes.
type GetValidPrepTaskConfigsByIngredientInput struct {
	Filter            *filtering.QueryFilter `json:"filter,omitempty"  jsonschema:"The page to read; absent reads the first page, oldest first"`
	ValidIngredientID string                 `json:"validIngredientID" jsonschema:"The identifier of the ingredient"`
}

var getValidPrepTaskConfigsByIngredientTool = &sdkmcp.Tool{
	Name:        "GetValidPrepTaskConfigsByIngredient",
	Description: "Page the valid prep task configs for an ingredient. Use this to find storage information for a particular ingredient.",
	Annotations: readOnly(),
}

// GetValidPrepTaskConfigsByIngredient pages the prep task configs for one ingredient.
func (t *Tools) GetValidPrepTaskConfigsByIngredient(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidPrepTaskConfigsByIngredientInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
	ctx, err := t.begin(ctx, req, getValidPrepTaskConfigsByIngredientTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.SearchValidPrepTaskConfigsByIngredient(ctx, in.ValidIngredientID, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

// GetValidPrepTaskConfigsByPreparationInput is what GetValidPrepTaskConfigsByPreparation takes.
type GetValidPrepTaskConfigsByPreparationInput struct {
	Filter             *filtering.QueryFilter `json:"filter,omitempty"   jsonschema:"The page to read; absent reads the first page, oldest first"`
	ValidPreparationID string                 `json:"validPreparationID" jsonschema:"The identifier of the preparation"`
}

var getValidPrepTaskConfigsByPreparationTool = &sdkmcp.Tool{
	Name:        "GetValidPrepTaskConfigsByPreparation",
	Description: "Page the valid prep task configs for a preparation. Use this to find storage information for ingredients prepared a certain way.",
	Annotations: readOnly(),
}

// GetValidPrepTaskConfigsByPreparation pages the prep task configs for one preparation.
func (t *Tools) GetValidPrepTaskConfigsByPreparation(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidPrepTaskConfigsByPreparationInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
	ctx, err := t.begin(ctx, req, getValidPrepTaskConfigsByPreparationTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.SearchValidPrepTaskConfigsByPreparation(ctx, in.ValidPreparationID, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}

// GetValidPrepTaskConfigsByIngredientAndPreparationInput is what GetValidPrepTaskConfigsByIngredientAndPreparation takes.
type GetValidPrepTaskConfigsByIngredientAndPreparationInput struct {
	Filter             *filtering.QueryFilter `json:"filter,omitempty"   jsonschema:"The page to read; absent reads the first page, oldest first"`
	ValidIngredientID  string                 `json:"validIngredientID"  jsonschema:"The identifier of the ingredient"`
	ValidPreparationID string                 `json:"validPreparationID" jsonschema:"The identifier of the preparation"`
}

var getValidPrepTaskConfigsByIngredientAndPreparationTool = &sdkmcp.Tool{
	Name:        "GetValidPrepTaskConfigsByIngredientAndPreparation",
	Description: "Page the valid prep task configs for an ingredient and preparation combination. Use this to find exactly how long a specific prepped ingredient (e.g., diced onions) can be stored.",
	Annotations: readOnly(),
}

// GetValidPrepTaskConfigsByIngredientAndPreparation pages the prep task configs for one ingredient prepared one way.
func (t *Tools) GetValidPrepTaskConfigsByIngredientAndPreparation(ctx context.Context, req *sdkmcp.CallToolRequest, in GetValidPrepTaskConfigsByIngredientAndPreparationInput) (*sdkmcp.CallToolResult, *filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
	ctx, err := t.begin(ctx, req, getValidPrepTaskConfigsByIngredientAndPreparationTool)
	if err != nil {
		return nil, nil, err
	}

	result, err := t.manager.SearchValidPrepTaskConfigsByIngredientAndPreparation(ctx, in.ValidIngredientID, in.ValidPreparationID, in.Filter)
	if err != nil {
		return nil, nil, err
	}

	return nil, result, nil
}
