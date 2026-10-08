/*
Package mcp is the meal planning domain's Model Context Protocol tool surface: a read-only tool
over each of the catalog and recipe types, served by internal/mcpserver.

The server mounts this through mcptools.Toolset and names nothing in here; what it needs is one
entry in internal/build/services/mcp. Every tool reads through the domain repository and checks
nothing but that the caller holds a token naming an account: the rows these read are the
catalog, which every account reads, and recipes, which the repository's reads already scope.
*/
package mcp

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	platformerrors "github.com/primandproper/primitives-go/v2/errors"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// The property names this domain's hand-written tool schemas repeat. The names shared with the
// other surfaces are mcptools'.
const (
	fieldBelongsToRecipe                = "BelongsToRecipe"
	fieldBelongsToRecipeStep            = "BelongsToRecipeStep"
	fieldCreatedByUser                  = "CreatedByUser"
	fieldIconPath                       = "IconPath"
	fieldIndex                          = "Index"
	fieldIngredient                     = "Ingredient"
	fieldMaxQuantity                    = "MaxQuantity"
	fieldMaxStorageTemperatureInCelsius = "MaxStorageTemperatureInCelsius"
	fieldMeasurementUnit                = "MeasurementUnit"
	fieldMinQuantity                    = "MinQuantity"
	fieldMinStorageTemperatureInCelsius = "MinStorageTemperatureInCelsius"
	fieldNotes                          = "Notes"
	fieldOptional                       = "Optional"
	fieldPluralName                     = "PluralName"
	fieldPreparation                    = "Preparation"
	fieldRecipeID                       = "RecipeID"
	fieldRecipeStepID                   = "RecipeStepID"
	fieldRecipeStepProductID            = "RecipeStepProductID"
	fieldSlug                           = "Slug"
	fieldStorageInstructions            = "StorageInstructions"
	fieldValidIngredientID              = "ValidIngredientID"
	fieldValidPreparationID             = "ValidPreparationID"
)

// Tools is the surface: every handler is a method on it, reading through repo.
type Tools struct {
	repo mealplanning.Repository
}

var _ mcptools.Toolset = (*Tools)(nil)

// NewTools builds the surface over repo.
func NewTools(repo mealplanning.Repository) (*Tools, error) {
	if repo == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil meal planning repository")
	}

	return &Tools{repo: repo}, nil
}

// RegisterOn adds every tool to server. Every one of them is a read.
func (t *Tools) RegisterOn(server *sdkmcp.Server) {
	// Valid Ingredients
	sdkmcp.AddTool(server, getValidIngredientTool, t.GetValidIngredient())
	sdkmcp.AddTool(server, searchForValidIngredientsTool, t.SearchForValidIngredients())

	// Valid Preparations
	sdkmcp.AddTool(server, getValidPreparationTool, t.GetValidPreparation())
	sdkmcp.AddTool(server, searchForValidPreparationsTool, t.SearchForValidPreparations())

	// Valid Measurement Units
	sdkmcp.AddTool(server, getValidMeasurementUnitTool, t.GetValidMeasurementUnit())
	sdkmcp.AddTool(server, searchForValidMeasurementUnitsTool, t.SearchForValidMeasurementUnits())

	// Valid Ingredient Preparations
	sdkmcp.AddTool(server, getValidIngredientPreparationTool, t.GetValidIngredientPreparation())
	sdkmcp.AddTool(server, getValidIngredientPreparationsTool, t.GetValidIngredientPreparations())

	// Valid Prep Task Configs
	sdkmcp.AddTool(server, getValidPrepTaskConfigTool, t.GetValidPrepTaskConfig())
	sdkmcp.AddTool(server, getValidPrepTaskConfigsTool, t.GetValidPrepTaskConfigs())
	sdkmcp.AddTool(server, getValidPrepTaskConfigsByIngredientTool, t.GetValidPrepTaskConfigsByIngredient())
	sdkmcp.AddTool(server, getValidPrepTaskConfigsByPreparationTool, t.GetValidPrepTaskConfigsByPreparation())
	sdkmcp.AddTool(server, getValidPrepTaskConfigsByIngredientAndPreparationTool, t.GetValidPrepTaskConfigsByIngredientAndPreparation())

	// Valid Ingredient Measurement Units
	sdkmcp.AddTool(server, getValidIngredientMeasurementUnitTool, t.GetValidIngredientMeasurementUnit())
	sdkmcp.AddTool(server, getValidIngredientMeasurementUnitsTool, t.GetValidIngredientMeasurementUnits())

	// Valid Vessels
	sdkmcp.AddTool(server, getValidVesselTool, t.GetValidVessel())
	sdkmcp.AddTool(server, searchForValidVesselsTool, t.SearchForValidVessels())

	// Valid Measurement Unit Conversions
	sdkmcp.AddTool(server, getValidMeasurementUnitConversionTool, t.GetValidMeasurementUnitConversion())
	sdkmcp.AddTool(server, getValidMeasurementUnitConversionsForUnitTool, t.GetValidMeasurementUnitConversionsForUnit())
	sdkmcp.AddTool(server, getValidMeasurementUnitConversionsForIngredientsTool, t.GetValidMeasurementUnitConversionsForIngredients())

	// Valid Ingredient States
	sdkmcp.AddTool(server, getValidIngredientStateTool, t.GetValidIngredientState())
	sdkmcp.AddTool(server, searchForValidIngredientStatesTool, t.SearchForValidIngredientStates())

	// Valid Ingredient State Ingredients
	sdkmcp.AddTool(server, getValidIngredientStateIngredientTool, t.GetValidIngredientStateIngredient())
	sdkmcp.AddTool(server, getValidIngredientStateIngredientsTool, t.GetValidIngredientStateIngredients())

	// Valid Instruments
	sdkmcp.AddTool(server, getValidInstrumentTool, t.GetValidInstrument())
	sdkmcp.AddTool(server, searchForValidInstrumentsTool, t.SearchForValidInstruments())

	// Valid Preparation Instruments
	sdkmcp.AddTool(server, getValidPreparationInstrumentTool, t.GetValidPreparationInstrument())
	sdkmcp.AddTool(server, getValidPreparationInstrumentsTool, t.GetValidPreparationInstruments())

	// Valid Preparation Vessels
	sdkmcp.AddTool(server, getValidPreparationVesselTool, t.GetValidPreparationVessel())
	sdkmcp.AddTool(server, getValidPreparationVesselsTool, t.GetValidPreparationVessels())

	// Recipe Step Instruments
	sdkmcp.AddTool(server, getRecipeStepInstrumentTool, t.GetRecipeStepInstrument())
	sdkmcp.AddTool(server, getRecipeStepInstrumentsTool, t.GetRecipeStepInstruments())

	// Recipe Step Products
	sdkmcp.AddTool(server, getRecipeStepProductTool, t.GetRecipeStepProduct())
	sdkmcp.AddTool(server, getRecipeStepProductsTool, t.GetRecipeStepProducts())

	// Recipe Step Ingredients
	sdkmcp.AddTool(server, getRecipeStepIngredientTool, t.GetRecipeStepIngredient())
	sdkmcp.AddTool(server, getRecipeStepIngredientsTool, t.GetRecipeStepIngredients())

	// Recipe Prep Tasks
	sdkmcp.AddTool(server, getRecipePrepTaskTool, t.GetRecipePrepTask())
	sdkmcp.AddTool(server, getRecipePrepTasksTool, t.GetRecipePrepTasks())

	// Recipe Step Vessels
	sdkmcp.AddTool(server, getRecipeStepVesselTool, t.GetRecipeStepVessel())
	sdkmcp.AddTool(server, getRecipeStepVesselsTool, t.GetRecipeStepVessels())

	// Recipe Step Completion Conditions
	sdkmcp.AddTool(server, getRecipeStepCompletionConditionTool, t.GetRecipeStepCompletionCondition())
	sdkmcp.AddTool(server, getRecipeStepCompletionConditionsTool, t.GetRecipeStepCompletionConditions())

	// Recipe Steps
	sdkmcp.AddTool(server, getRecipeStepTool, t.GetRecipeStep())
	sdkmcp.AddTool(server, getRecipeStepsTool, t.GetRecipeSteps())

	// Recipes
	sdkmcp.AddTool(server, getRecipeTool, t.GetRecipe())
	sdkmcp.AddTool(server, getRecipesTool, t.GetRecipes())
	sdkmcp.AddTool(server, searchForRecipesTool, t.SearchForRecipes())
}
