/*
Package mcp is the meal planning domain's Model Context Protocol tool surface: a read-only tool
over each of the catalog and recipe types, served by internal/mcpserver.

The server mounts this through mcptools.Toolset and names nothing in here; what it needs is one
entry in internal/build/services/mcp.

# The same rules as the gRPC surface, over another transport

Every tool reads through the domain manager, as every gRPC method does, so a rule the manager
enforces holds over both transports and the first write tool that arrives here is a write with
a rule in front of it. Every tool is named for its gRPC counterpart, and requires the permission
the gRPC permission table declares for that method — the one table, read rather than copied, so
a caller refused a read over gRPC is refused it here. The check is mcptools.Gate's, which turns
the call's verified token into this application's session before asking.

# The schema is the type's

What a model is told about a tool's arguments and its answer is reflected off the Go types by
the MCP SDK: property names from the json tags, descriptions from the jsonschema tags. The
hand-written schemas this replaced named every output property in a case the wire never used.
platform-go's mcptool adds each field's doc comment as its description; it is unexported at
v15.1.0 (platform-go#1162), and is what schemaFor becomes when it ships.
*/
package mcp

import (
	"context"

	mealplanningmgr "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/filtering"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	// ErrNilManager is a tool surface built over no manager.
	ErrNilManager = platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil meal planning manager for the MCP tools")

	// ErrNilGate is a tool surface built with nothing to check a caller with.
	ErrNilGate = platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil gate for the meal planning MCP tools")
)

// Tools is the surface: every handler is a method on it, reading through the manager once the
// gate has admitted the call.
type Tools struct {
	manager mealplanningmgr.MealPlanningManager
	gate    *mcptools.Gate
}

var _ mcptools.Toolset = (*Tools)(nil)

// NewTools builds the surface over the manager and the gate.
func NewTools(manager mealplanningmgr.MealPlanningManager, gate *mcptools.Gate) (*Tools, error) {
	if manager == nil {
		return nil, ErrNilManager
	}

	if gate == nil {
		return nil, ErrNilGate
	}

	return &Tools{manager: manager, gate: gate}, nil
}

// RegisterOn adds every tool to server. Every one of them is a read.
func (t *Tools) RegisterOn(server *sdkmcp.Server) {
	// Valid Ingredients
	add(server, getValidIngredientTool, t.GetValidIngredient)
	add(server, searchForValidIngredientsTool, t.SearchForValidIngredients)

	// Valid Preparations
	add(server, getValidPreparationTool, t.GetValidPreparation)
	add(server, searchForValidPreparationsTool, t.SearchForValidPreparations)

	// Valid Measurement Units
	add(server, getValidMeasurementUnitTool, t.GetValidMeasurementUnit)
	add(server, searchForValidMeasurementUnitsTool, t.SearchForValidMeasurementUnits)

	// Valid Ingredient Preparations
	add(server, getValidIngredientPreparationTool, t.GetValidIngredientPreparation)
	add(server, getValidIngredientPreparationsTool, t.GetValidIngredientPreparations)

	// Valid Prep Task Configs
	add(server, getValidPrepTaskConfigTool, t.GetValidPrepTaskConfig)
	add(server, getValidPrepTaskConfigsTool, t.GetValidPrepTaskConfigs)
	add(server, getValidPrepTaskConfigsByIngredientTool, t.GetValidPrepTaskConfigsByIngredient)
	add(server, getValidPrepTaskConfigsByPreparationTool, t.GetValidPrepTaskConfigsByPreparation)
	add(server, getValidPrepTaskConfigsByIngredientAndPreparationTool, t.GetValidPrepTaskConfigsByIngredientAndPreparation)

	// Valid Ingredient Measurement Units
	add(server, getValidIngredientMeasurementUnitTool, t.GetValidIngredientMeasurementUnit)
	add(server, getValidIngredientMeasurementUnitsTool, t.GetValidIngredientMeasurementUnits)

	// Valid Vessels
	add(server, getValidVesselTool, t.GetValidVessel)
	add(server, searchForValidVesselsTool, t.SearchForValidVessels)

	// Valid Measurement Unit Conversions
	add(server, getValidMeasurementUnitConversionTool, t.GetValidMeasurementUnitConversion)
	add(server, getValidMeasurementUnitConversionsForUnitTool, t.GetValidMeasurementUnitConversionsForUnit)
	add(server, getValidMeasurementUnitConversionsForIngredientsTool, t.GetValidMeasurementUnitConversionsForIngredients)

	// Valid Ingredient States
	add(server, getValidIngredientStateTool, t.GetValidIngredientState)
	add(server, searchForValidIngredientStatesTool, t.SearchForValidIngredientStates)

	// Valid Ingredient State Ingredients
	add(server, getValidIngredientStateIngredientTool, t.GetValidIngredientStateIngredient)
	add(server, getValidIngredientStateIngredientsTool, t.GetValidIngredientStateIngredients)

	// Valid Instruments
	add(server, getValidInstrumentTool, t.GetValidInstrument)
	add(server, searchForValidInstrumentsTool, t.SearchForValidInstruments)

	// Valid Preparation Instruments
	add(server, getValidPreparationInstrumentTool, t.GetValidPreparationInstrument)
	add(server, getValidPreparationInstrumentsTool, t.GetValidPreparationInstruments)

	// Valid Preparation Vessels
	add(server, getValidPreparationVesselTool, t.GetValidPreparationVessel)
	add(server, getValidPreparationVesselsTool, t.GetValidPreparationVessels)

	// Recipe Step Instruments
	add(server, getRecipeStepInstrumentTool, t.GetRecipeStepInstrument)
	add(server, getRecipeStepInstrumentsTool, t.GetRecipeStepInstruments)

	// Recipe Step Products
	add(server, getRecipeStepProductTool, t.GetRecipeStepProduct)
	add(server, getRecipeStepProductsTool, t.GetRecipeStepProducts)

	// Recipe Step Ingredients
	add(server, getRecipeStepIngredientTool, t.GetRecipeStepIngredient)
	add(server, getRecipeStepIngredientsTool, t.GetRecipeStepIngredients)

	// Recipe Prep Tasks
	add(server, getRecipePrepTaskTool, t.GetRecipePrepTask)
	add(server, getRecipePrepTasksTool, t.GetRecipePrepTasks)

	// Recipe Step Vessels
	add(server, getRecipeStepVesselTool, t.GetRecipeStepVessel)
	add(server, getRecipeStepVesselsTool, t.GetRecipeStepVessels)

	// Recipe Step Completion Conditions
	add(server, getRecipeStepCompletionConditionTool, t.GetRecipeStepCompletionCondition)
	add(server, getRecipeStepCompletionConditionsTool, t.GetRecipeStepCompletionConditions)

	// Recipe Steps
	add(server, getRecipeStepTool, t.GetRecipeStep)
	add(server, getRecipeStepsTool, t.GetRecipeSteps)

	// Recipes
	add(server, getRecipeTool, t.GetRecipe)
	add(server, getRecipesTool, t.GetRecipes)
	add(server, searchForRecipesTool, t.SearchForRecipes)
}

// begin is where every tool starts: the gate authenticates the call and checks that the caller
// holds what the tool's gRPC counterpart requires. The context returned is the one the manager
// is read through.
func (t *Tools) begin(ctx context.Context, req *sdkmcp.CallToolRequest, tool *sdkmcp.Tool) (context.Context, error) {
	required, err := permissionsFor(tool)
	if err != nil {
		return ctx, err
	}

	ctx, _, err = t.gate.Begin(ctx, req, required...)

	return ctx, err
}

// FilterInput is what a paged read takes.
type FilterInput struct {
	Filter *filtering.QueryFilter `json:"filter,omitempty" jsonschema:"The page to read; absent reads the first page, oldest first"`
}

// SearchInput is what a search takes: the page, and the text to match.
type SearchInput struct {
	Filter *filtering.QueryFilter `json:"filter,omitempty" jsonschema:"The page to read; absent reads the first page, oldest first"`
	Query  string                 `json:"query"            jsonschema:"The text to search for"`
}
