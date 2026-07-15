package recipeanalysis

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	"github.com/primandproper/platform-go/v4/identifiers"
	"github.com/primandproper/platform-go/v4/observability"
	"github.com/primandproper/platform-go/v4/observability/logging"
	"github.com/primandproper/platform-go/v4/observability/tracing"

	"github.com/dustin/go-humanize/english"
	"github.com/hako/durafmt"
	"github.com/heimdalr/dag"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

var (
	_ graph.Node = (*recipeStepGraphNode)(nil)

	errNotAcyclic = errors.New("recipe is not acyclic")
)

// recipeStepGraphNode is a node in an implicit graph.
type recipeStepGraphNode struct {
	id int64
}

// newGraphNode returns a new recipeStepGraphNode.
func newGraphNode(id int64) graph.Node {
	return &recipeStepGraphNode{id: id}
}

func (g recipeStepGraphNode) ID() int64 {
	return g.id
}

var errRecipeStepIDNotFound = errors.New("recipe step ID not found")

const (
	// recipeGraphIDBase is the offset for associated recipe steps so they don't collide with main recipe steps (1..n).
	recipeGraphIDBase = 10000

	// mealGraphIDBase is the offset per meal component so recipe graphs don't collide when combined.
	mealGraphIDBase = 100000
)

// stepLocation identifies a step within a recipe (main or associated).
type stepLocation struct {
	recipeIndex int // 0 = main recipe, 1+ = associated recipe index
	stepIndex   int
}

// graphIDForStepLocation returns a unique int64 graph ID for a step.
// Main recipe steps use 1..n; associated recipe steps use recipeGraphIDBase + recipeIndex*1000 + stepIndex+1.
func graphIDForStepLocation(loc stepLocation) int64 {
	if loc.recipeIndex == 0 {
		return int64(loc.stepIndex + 1)
	}
	return int64(recipeGraphIDBase + loc.recipeIndex*1000 + (loc.stepIndex + 1))
}

func graphIDForStep(step *mealplanning.RecipeStep) int64 {
	return int64(step.Index + 1)
}

// graphIDForMainStep returns the graph ID for a main recipe step (1..n).
func graphIDForMainStep(step *mealplanning.RecipeStep) int64 {
	return int64(step.Index + 1)
}

func findStepLocationForRecipeStepProductID(recipe *mealplanning.Recipe, productID string, sourceRecipeID *string) (stepLocation, bool) {
	step, stepRecipe := recipe.FindStepForRecipeStepProductIDWithRecipe(productID, sourceRecipeID)
	if step == nil {
		return stepLocation{}, false
	}
	if stepRecipe == recipe {
		return stepLocation{recipeIndex: 0, stepIndex: recipe.FindStepIndexByID(step.ID)}, true
	}
	for i, assoc := range recipe.AssociatedRecipes {
		if assoc == stepRecipe {
			stepIdx := -1
			for j, s := range assoc.Steps {
				if s.ID == step.ID {
					stepIdx = j
					break
				}
			}
			if stepIdx >= 0 {
				return stepLocation{recipeIndex: i + 1, stepIndex: stepIdx}, true
			}
			break
		}
	}
	return stepLocation{}, false
}

func findStepIndexForRecipeStepProductIDWithSource(recipe *mealplanning.Recipe, recipeStepProductID string, sourceRecipeID *string) (int64, error) {
	if loc, ok := findStepLocationForRecipeStepProductID(recipe, recipeStepProductID, sourceRecipeID); ok {
		return graphIDForStepLocation(loc), nil
	}
	return -1, errRecipeStepIDNotFound
}

func findStepIDForRecipeStepProductIDWithSource(recipe *mealplanning.Recipe, recipeStepProductID string, sourceRecipeID *string) (string, error) {
	if loc, ok := findStepLocationForRecipeStepProductID(recipe, recipeStepProductID, sourceRecipeID); ok {
		return fmt.Sprintf("%d", graphIDForStepLocation(loc)), nil
	}
	return "", errRecipeStepIDNotFound
}

func findStepIndexForRecipeStepID(recipe *mealplanning.Recipe, recipeStepID string) (int64, error) {
	for _, step := range recipe.Steps {
		if step.ID == recipeStepID {
			return graphIDForMainStep(step), nil
		}
	}

	return -1, errRecipeStepIDNotFound
}

// allRecipeSteps returns all steps from main recipe and associated recipes with their locations.
func allRecipeSteps(recipe *mealplanning.Recipe) []struct {
	step *mealplanning.RecipeStep
	loc  stepLocation
} {
	var out []struct {
		step *mealplanning.RecipeStep
		loc  stepLocation
	}
	for i, step := range recipe.Steps {
		out = append(out, struct {
			step *mealplanning.RecipeStep
			loc  stepLocation
		}{step, stepLocation{recipeIndex: 0, stepIndex: i}})
	}
	for i, assoc := range recipe.AssociatedRecipes {
		for j, step := range assoc.Steps {
			out = append(out, struct {
				step *mealplanning.RecipeStep
				loc  stepLocation
			}{step, stepLocation{recipeIndex: i + 1, stepIndex: j}})
		}
	}
	return out
}

// RecipeAnalyzer analyzes recipes for insights (ugh).
type RecipeAnalyzer interface {
	MakeGraphForRecipe(ctx context.Context, recipe *mealplanning.Recipe) (*simple.DirectedGraph, error)
	MakeGraphForMeal(ctx context.Context, meal *mealplanning.Meal) (*simple.DirectedGraph, error)
	ValidateRecipeCreationRequestInputIsDAG(ctx context.Context, input *mealplanning.RecipeCreationRequestInput) error
	GenerateMealPlanTasksForRecipe(ctx context.Context, mealPlanOptionID string, recipe *mealplanning.Recipe) ([]*mealplanning.MealPlanTaskDatabaseCreationInput, error)
	RenderMermaidDiagramForRecipe(ctx context.Context, recipe *mealplanning.Recipe) string
	RenderMermaidDiagramForMeal(ctx context.Context, meal *mealplanning.Meal) string
	RenderGraphvizDiagramForRecipe(ctx context.Context, recipe *mealplanning.Recipe) string
	RenderGraphvizDiagramForMeal(ctx context.Context, meal *mealplanning.Meal) string
}

var _ RecipeAnalyzer = (*recipeAnalyzer)(nil)

// recipeAnalyzer creates graphs from recipes.
type recipeAnalyzer struct {
	tracer tracing.Tracer
	logger logging.Logger
}

// NewRecipeAnalyzer creates a recipeAnalyzer.
func NewRecipeAnalyzer(logger logging.Logger, tracerProvider tracing.TracerProvider) RecipeAnalyzer {
	return &recipeAnalyzer{
		logger: logging.NewNamedLogger(logger, "recipe_analyzer"),
		tracer: tracing.NewNamedTracer(tracerProvider, "recipe_grapher"),
	}
}

func (g *recipeAnalyzer) MakeGraphForRecipe(ctx context.Context, recipe *mealplanning.Recipe) (*simple.DirectedGraph, error) {
	_, span := g.tracer.StartSpan(ctx)
	defer span.End()

	recipeGraph := simple.NewDirectedGraph()

	allSteps := allRecipeSteps(recipe)
	for _, item := range allSteps {
		recipeGraph.AddNode(newGraphNode(graphIDForStepLocation(item.loc)))
	}

	for _, item := range allSteps {
		toGraphID := graphIDForStepLocation(item.loc)
		toStep := item.step

		for _, ingredient := range toStep.Ingredients {
			if ingredient.RecipeStepProductID == nil {
				continue
			}

			fromStep, err := findStepIndexForRecipeStepProductIDWithSource(recipe, *ingredient.RecipeStepProductID, ingredient.RecipeStepProductRecipeID)
			if err != nil {
				return nil, err
			}

			from := recipeGraph.Node(fromStep)
			to := recipeGraph.Node(toGraphID)
			recipeGraph.SetEdge(simple.Edge{F: from, T: to})
		}

		for _, instrument := range toStep.Instruments {
			if instrument.RecipeStepProductID == nil {
				continue
			}

			fromStep, err := findStepIndexForRecipeStepProductIDWithSource(recipe, *instrument.RecipeStepProductID, nil)
			if err != nil {
				return nil, err
			}

			from := recipeGraph.Node(fromStep)
			to := recipeGraph.Node(toGraphID)
			recipeGraph.SetEdge(simple.Edge{F: from, T: to})
		}

		for _, vessel := range toStep.Vessels {
			if vessel.RecipeStepProductID == nil {
				continue
			}

			fromStep, err := findStepIndexForRecipeStepProductIDWithSource(recipe, *vessel.RecipeStepProductID, nil)
			if err != nil {
				return nil, err
			}

			from := recipeGraph.Node(fromStep)
			to := recipeGraph.Node(toGraphID)
			recipeGraph.SetEdge(simple.Edge{F: from, T: to})
		}
	}

	directedCycles := topo.DirectedCyclesIn(recipeGraph)
	if len(directedCycles) > 0 {
		return nil, errNotAcyclic
	}

	if _, err := g.makeDAGForRecipe(ctx, recipe); err != nil {
		return nil, fmt.Errorf("parsing recipe as DAG: %w", err)
	}

	return recipeGraph, nil
}

// mealStepItem holds a step with its location within a meal (component index + step location within that recipe).
type mealStepItem struct {
	step           *mealplanning.RecipeStep
	recipe         *mealplanning.Recipe
	loc            stepLocation
	componentIndex int
}

// mealGraphID returns the graph ID for a step within a combined meal graph.
func mealGraphID(componentIndex int, loc stepLocation) int64 {
	return int64(componentIndex)*mealGraphIDBase + graphIDForStepLocation(loc)
}

// allMealSteps returns all steps from all meal components with their locations.
func allMealSteps(meal *mealplanning.Meal) []mealStepItem {
	var out []mealStepItem
	for compIdx, comp := range meal.Components {
		recipe := &comp.Recipe
		allSteps := allRecipeSteps(recipe)
		for _, item := range allSteps {
			out = append(out, mealStepItem{
				componentIndex: compIdx,
				loc:            item.loc,
				step:           item.step,
				recipe:         recipe,
			})
		}
	}
	return out
}

func (g *recipeAnalyzer) MakeGraphForMeal(ctx context.Context, meal *mealplanning.Meal) (*simple.DirectedGraph, error) {
	_, span := g.tracer.StartSpan(ctx)
	defer span.End()

	mealGraph := simple.NewDirectedGraph()

	allSteps := allMealSteps(meal)
	for i := range allSteps {
		mealGraph.AddNode(newGraphNode(mealGraphID(allSteps[i].componentIndex, allSteps[i].loc)))
	}

	for i := range allSteps {
		item := &allSteps[i]
		toGraphID := mealGraphID(item.componentIndex, item.loc)
		toStep := item.step
		recipe := item.recipe

		for _, ingredient := range toStep.Ingredients {
			if ingredient.RecipeStepProductID == nil {
				continue
			}

			fromLoc, ok := findStepLocationForRecipeStepProductID(recipe, *ingredient.RecipeStepProductID, ingredient.RecipeStepProductRecipeID)
			if !ok {
				return nil, errRecipeStepIDNotFound
			}

			fromGraphID := mealGraphID(item.componentIndex, fromLoc)
			from := mealGraph.Node(fromGraphID)
			to := mealGraph.Node(toGraphID)
			mealGraph.SetEdge(simple.Edge{F: from, T: to})
		}

		for _, instrument := range toStep.Instruments {
			if instrument.RecipeStepProductID == nil {
				continue
			}

			fromLoc, ok := findStepLocationForRecipeStepProductID(recipe, *instrument.RecipeStepProductID, nil)
			if !ok {
				return nil, errRecipeStepIDNotFound
			}

			fromGraphID := mealGraphID(item.componentIndex, fromLoc)
			from := mealGraph.Node(fromGraphID)
			to := mealGraph.Node(toGraphID)
			mealGraph.SetEdge(simple.Edge{F: from, T: to})
		}

		for _, vessel := range toStep.Vessels {
			if vessel.RecipeStepProductID == nil {
				continue
			}

			fromLoc, ok := findStepLocationForRecipeStepProductID(recipe, *vessel.RecipeStepProductID, nil)
			if !ok {
				return nil, errRecipeStepIDNotFound
			}

			fromGraphID := mealGraphID(item.componentIndex, fromLoc)
			from := mealGraph.Node(fromGraphID)
			to := mealGraph.Node(toGraphID)
			mealGraph.SetEdge(simple.Edge{F: from, T: to})
		}
	}

	directedCycles := topo.DirectedCyclesIn(mealGraph)
	if len(directedCycles) > 0 {
		return nil, errNotAcyclic
	}

	return mealGraph, nil
}

// ValidateRecipeCreationRequestInputIsDAG validates that a RecipeCreationRequestInput represents a valid DAG.
// It builds a graph from the request input and checks for cycles.
// ProductOfRecipeStepIndex is treated as an array index into the Steps slice.
func (g *recipeAnalyzer) ValidateRecipeCreationRequestInputIsDAG(ctx context.Context, input *mealplanning.RecipeCreationRequestInput) error {
	_, span := g.tracer.StartSpan(ctx)
	defer span.End()

	recipeGraph := simple.NewDirectedGraph()

	// Add nodes for each step (using array index+1 as the graph ID, similar to graphIDForStep)
	for stepIdx := range input.Steps {
		graphID := int64(stepIdx + 1)
		recipeGraph.AddNode(newGraphNode(graphID))
	}

	// Build edges based on ProductOfRecipeStepIndex references
	for stepIdx := range input.Steps {
		currentStepGraphID := int64(stepIdx + 1)
		step := input.Steps[stepIdx]

		// Check ingredients for references to other steps
		for _, ingredient := range step.Ingredients {
			if ingredient.ProductOfRecipeStepIndex == nil {
				continue
			}

			fromStepArrayIndex := *ingredient.ProductOfRecipeStepIndex

			// Skip cross-recipe references (they don't affect this recipe's DAG)
			// Check this first before validating index bounds
			// If RecipeStepProductRecipeID is set (even if empty), it indicates a cross-recipe reference
			// The empty string case can happen when getRecipeIDBySlug returns nil but we still want to indicate
			// this is a cross-recipe reference that will be resolved later
			if ingredient.RecipeStepProductRecipeID != nil {
				continue
			}

			// Validate that the referenced step array index exists
			if fromStepArrayIndex >= uint64(len(input.Steps)) {
				return fmt.Errorf("ingredient in step at array index %d references invalid step array index %d", stepIdx, fromStepArrayIndex)
			}

			// Prevent self-references (which would create a cycle)
			if fromStepArrayIndex == uint64(stepIdx) {
				return fmt.Errorf("%w: step at array index %d references itself", errNotAcyclic, stepIdx)
			}

			fromStepGraphID := int64(fromStepArrayIndex + 1)
			from := recipeGraph.Node(fromStepGraphID)
			to := recipeGraph.Node(currentStepGraphID)
			recipeGraph.SetEdge(simple.Edge{F: from, T: to})
		}

		// Check instruments for references to other steps
		for _, instrument := range step.Instruments {
			if instrument.ProductOfRecipeStepIndex == nil {
				continue
			}

			fromStepArrayIndex := *instrument.ProductOfRecipeStepIndex
			// Validate that the referenced step array index exists
			if fromStepArrayIndex >= uint64(len(input.Steps)) {
				return fmt.Errorf("instrument in step at array index %d references invalid step array index %d", stepIdx, fromStepArrayIndex)
			}

			// Prevent self-references (which would create a cycle)
			if fromStepArrayIndex == uint64(stepIdx) {
				return fmt.Errorf("%w: step at array index %d references itself", errNotAcyclic, stepIdx)
			}

			fromStepGraphID := int64(fromStepArrayIndex + 1)
			from := recipeGraph.Node(fromStepGraphID)
			to := recipeGraph.Node(currentStepGraphID)
			recipeGraph.SetEdge(simple.Edge{F: from, T: to})
		}

		// Check vessels for references to other steps
		for _, vessel := range step.Vessels {
			if vessel.ProductOfRecipeStepIndex == nil {
				continue
			}

			fromStepArrayIndex := *vessel.ProductOfRecipeStepIndex
			// Validate that the referenced step array index exists
			if fromStepArrayIndex >= uint64(len(input.Steps)) {
				return fmt.Errorf("vessel in step at array index %d references invalid step array index %d", stepIdx, fromStepArrayIndex)
			}

			// Prevent self-references (which would create a cycle)
			if fromStepArrayIndex == uint64(stepIdx) {
				return fmt.Errorf("%w: step at array index %d references itself", errNotAcyclic, stepIdx)
			}

			fromStepGraphID := int64(fromStepArrayIndex + 1)
			from := recipeGraph.Node(fromStepGraphID)
			to := recipeGraph.Node(currentStepGraphID)
			recipeGraph.SetEdge(simple.Edge{F: from, T: to})
		}
	}

	// Check for cycles
	directedCycles := topo.DirectedCyclesIn(recipeGraph)
	if len(directedCycles) > 0 {
		return fmt.Errorf("%w: recipe contains %d cycle(s)", errNotAcyclic, len(directedCycles))
	}

	return nil
}

type RecipeStepIdentifier struct {
	recipeStep *mealplanning.RecipeStep
	loc        *stepLocation // nil for backward compat with main-recipe-only callers
}

func (i *RecipeStepIdentifier) ID() string {
	if i.loc != nil {
		return fmt.Sprintf("%d", graphIDForStepLocation(*i.loc))
	}
	return fmt.Sprintf("%d", graphIDForStep(i.recipeStep))
}

// makeDAGForRecipe makes a proper DAG for the provided Recipe.
func (g *recipeAnalyzer) makeDAGForRecipe(ctx context.Context, recipe *mealplanning.Recipe) (*dag.DAG, error) {
	_, span := g.tracer.StartSpan(ctx)
	defer span.End()

	recipeGraph := dag.NewDAG()

	allSteps := allRecipeSteps(recipe)
	for _, item := range allSteps {
		loc := item.loc
		if _, err := recipeGraph.AddVertex(&RecipeStepIdentifier{recipeStep: item.step, loc: &loc}); err != nil {
			return nil, fmt.Errorf("adding step %v to graph: %w", loc, err)
		}
	}

	for _, item := range allSteps {
		consumerID := fmt.Sprintf("%d", graphIDForStepLocation(item.loc))
		step := item.step

		for _, ingredient := range step.Ingredients {
			if ingredient.RecipeStepProductID == nil {
				continue
			}

			producerStepID, err := findStepIDForRecipeStepProductIDWithSource(recipe, *ingredient.RecipeStepProductID, ingredient.RecipeStepProductRecipeID)
			if err != nil {
				return nil, fmt.Errorf("finding step ID for recipe step product ID: %w", err)
			}

			if err = recipeGraph.AddEdge(producerStepID, consumerID); err != nil {
				return nil, fmt.Errorf("adding recipe step edge: %w", err)
			}
		}

		for _, instrument := range step.Instruments {
			if instrument.RecipeStepProductID == nil {
				continue
			}

			producerStepID, err := findStepIDForRecipeStepProductIDWithSource(recipe, *instrument.RecipeStepProductID, nil)
			if err != nil {
				return nil, fmt.Errorf("finding step ID for recipe step instrument product ID: %w", err)
			}

			if err = recipeGraph.AddEdge(producerStepID, consumerID); err != nil {
				var dupeErr dag.EdgeDuplicateError
				if errors.As(err, &dupeErr) {
					continue
				}
				return nil, fmt.Errorf("adding instrument step edge: %w", err)
			}
		}

		for _, vessel := range step.Vessels {
			if vessel.RecipeStepProductID == nil {
				continue
			}

			producerStepID, err := findStepIDForRecipeStepProductIDWithSource(recipe, *vessel.RecipeStepProductID, nil)
			if err != nil {
				return nil, fmt.Errorf("finding step ID for recipe step vessel product ID: %w", err)
			}

			if err = recipeGraph.AddEdge(producerStepID, consumerID); err != nil {
				var dupeErr dag.EdgeDuplicateError
				if errors.As(err, &dupeErr) {
					continue
				}
				return nil, fmt.Errorf("adding vessel step edge: %w", err)
			}
		}
	}

	recipeGraph.ReduceTransitively()

	return recipeGraph, nil
}

// frozenIngredientDefrostStepsFilter iterates through a recipe and returns
// the list of ingredients within that are indicated as kept frozen.
func frozenIngredientDefrostStepsFilter(recipe *mealplanning.Recipe) map[string][]int {
	out := map[string][]int{}

	for _, recipeStep := range recipe.Steps {
		ingredientIndices := []int{}
		for i, ingredient := range recipeStep.Ingredients {
			// if it's a valid ingredient
			if ingredient.Ingredient != nil &&
				// if the ingredient has storage temperature set
				ingredient.Ingredient.MinStorageTemperatureInCelsius != nil &&
				// the ingredient's storage temperature is set to something about freezing temperature.
				*ingredient.Ingredient.MinStorageTemperatureInCelsius <= 3 {
				ingredientIndices = append(ingredientIndices, i)
			}
		}

		if len(ingredientIndices) > 0 {
			out[recipeStep.ID] = ingredientIndices
		}
	}

	return out
}

func buildThawStepCreationExplanation(recipeStepIndex int64, ingredientIndices ...int) string {
	if len(ingredientIndices) == 0 {
		return ""
	}

	stringIndices := []string{}
	for _, i := range ingredientIndices {
		stringIndices = append(stringIndices, fmt.Sprintf("#%d", i+1))
	}

	var d string
	if len(ingredientIndices) > 1 {
		d = "ingredients"
	} else if len(ingredientIndices) == 1 {
		d = "ingredient"
	}

	return fmt.Sprintf("frozen %s (%s) for step #%d might need to be thawed ahead of time", d, strings.Join(stringIndices, ", "), recipeStepIndex)
}

func (g *recipeAnalyzer) generateMealPlanTasksForFrozenIngredients(ctx context.Context, mealPlanOptionID string, recipe *mealplanning.Recipe) []*mealplanning.MealPlanTaskDatabaseCreationInput {
	_, span := g.tracer.StartSpan(ctx)
	defer span.End()

	logger := g.logger.Clone().WithValue(mealplanningkeys.RecipeIDKey, recipe.ID)

	frozenIngredientSteps := frozenIngredientDefrostStepsFilter(recipe)
	logger.WithValue("frozen_steps_qty", len(frozenIngredientSteps)).Info("creating frozen stepSet inputs")

	outputs := []*mealplanning.MealPlanTaskDatabaseCreationInput{}
	for stepID, ingredientIndices := range frozenIngredientSteps {
		stepIndex, err := findStepIndexForRecipeStepID(recipe, stepID)
		if err != nil {
			observability.AcknowledgeError(err, logger, span, "determining recipe stepSet index for stepSet ID")
			continue
		}

		explanation := buildThawStepCreationExplanation(stepIndex, ingredientIndices...)
		if explanation == "" {
			continue
		}

		outputs = append(outputs, &mealplanning.MealPlanTaskDatabaseCreationInput{
			ID:                  identifiers.New(),
			CreationExplanation: explanation,
			MealPlanOptionID:    mealPlanOptionID,
		})
	}

	return outputs
}

const recipeTaskStepCreationExplanation = "recipe prep task exists for steps"

func (g *recipeAnalyzer) GenerateMealPlanTasksForRecipe(ctx context.Context, mealPlanOptionID string, recipe *mealplanning.Recipe) ([]*mealplanning.MealPlanTaskDatabaseCreationInput, error) {
	ctx, span := g.tracer.StartSpan(ctx)
	defer span.End()

	inputs := g.generateMealPlanTasksForFrozenIngredients(ctx, mealPlanOptionID, recipe)

	for _, prepTask := range recipe.PrepTasks {
		inputs = append(inputs, &mealplanning.MealPlanTaskDatabaseCreationInput{
			AssignedToUser:      nil,
			CreationExplanation: recipeTaskStepCreationExplanation,
			StatusExplanation:   "",
			MealPlanOptionID:    mealPlanOptionID,

			RecipePrepTaskID: prepTask.ID,
			ID:               identifiers.New(),
		})
	}

	return inputs, nil
}

type provisionCount struct {
	ingredients, instruments, vessels uint
}

// stepProvidesWhatFromTo returns a description of what the from step provides to the to step.
func stepProvidesWhatFromTo(_ *mealplanning.Recipe, from, to *mealplanning.RecipeStep) string {
	provides := []string{}
	count := provisionCount{}

	for _, product := range from.Products {
		for _, ingredient := range to.Ingredients {
			if ingredient.RecipeStepProductID != nil && *ingredient.RecipeStepProductID == product.ID {
				count.ingredients++
			}
		}
		for _, instrument := range to.Instruments {
			if instrument.RecipeStepProductID != nil && *instrument.RecipeStepProductID == product.ID {
				count.instruments++
			}
		}
		for _, vessel := range to.Vessels {
			if vessel.RecipeStepProductID != nil && *vessel.RecipeStepProductID == product.ID {
				count.vessels++
			}
		}
	}

	renderCount := func(x uint, typ string) string {
		return strings.TrimSpace(fmt.Sprintf(" %s ", english.PluralWord(int(x), typ, fmt.Sprintf("%ss", typ))))
	}

	if count.ingredients > 0 {
		provides = append(provides, renderCount(count.ingredients, "ingredient"))
	}
	if count.instruments > 0 {
		provides = append(provides, renderCount(count.instruments, "instrument"))
	}
	if count.vessels > 0 {
		provides = append(provides, renderCount(count.vessels, "vessel"))
	}

	return english.OxfordWordSeries(provides, "and")
}

func (g *recipeAnalyzer) RenderMermaidDiagramForRecipe(ctx context.Context, recipe *mealplanning.Recipe) string {
	_, span := g.tracer.StartSpan(ctx)
	defer span.End()

	var mermaid strings.Builder
	mermaid.WriteString("flowchart TD;\n")

	allSteps := allRecipeSteps(recipe)
	for _, item := range allSteps {
		label := stepLabelForDiagram(item.loc, item.step, recipe)
		gid := graphIDForStepLocation(item.loc)
		if _, err := fmt.Fprintf(&mermaid, "\tStep%d[\"%s\"];\n", gid, label); err != nil {
			observability.AcknowledgeError(err, g.logger, span, "writing mermaid step node")
		}
	}

	for i := range allSteps {
		for j := range allSteps {
			if i == j {
				continue
			}
			provides := stepProvidesWhatFromTo(recipe, allSteps[i].step, allSteps[j].step)
			if provides != "" {
				if _, err := fmt.Fprintf(&mermaid, "\tStep%d -->|%s| Step%d;\n",
					graphIDForStepLocation(allSteps[i].loc), provides, graphIDForStepLocation(allSteps[j].loc)); err != nil {
					observability.AcknowledgeError(err, g.logger, span, "writing mermaid step edge")
				}
			}
		}
	}

	for i := range recipe.PrepTasks {
		prepTask := recipe.PrepTasks[i]

		if _, err := fmt.Fprintf(&mermaid, "subgraph %d [\"%s (prep task #%d)\"]\n", i, prepTask.Name, i+1); err != nil {
			observability.AcknowledgeError(err, g.logger, span, "writing mermaid subgraph header")
		}
		for j := range prepTask.TaskSteps {
			stepIdx := recipe.FindStepIndexByID(prepTask.TaskSteps[j].BelongsToRecipeStep)
			if stepIdx >= 0 {
				if _, err := fmt.Fprintf(&mermaid, "Step%d\n", stepIdx+1); err != nil {
					observability.AcknowledgeError(err, g.logger, span, "writing mermaid subgraph step")
				}
			}
		}
		mermaid.WriteString("end\n")
	}

	return mermaid.String()
}

func (g *recipeAnalyzer) RenderMermaidDiagramForMeal(ctx context.Context, meal *mealplanning.Meal) string {
	_, span := g.tracer.StartSpan(ctx)
	defer span.End()

	var mermaid strings.Builder
	mermaid.WriteString("flowchart TD;\n")

	allSteps := allMealSteps(meal)
	for i := range allSteps {
		label := stepLabelForMealDiagram(allSteps[i])
		gid := mealGraphID(allSteps[i].componentIndex, allSteps[i].loc)
		if _, err := fmt.Fprintf(&mermaid, "\tStep%d[\"%s\"];\n", gid, label); err != nil {
			observability.AcknowledgeError(err, g.logger, span, "writing mermaid meal step node")
		}
	}

	for i := range allSteps {
		for j := range allSteps {
			if i == j {
				continue
			}
			// Only draw edges within the same component (recipes don't share products across components)
			if allSteps[i].componentIndex != allSteps[j].componentIndex {
				continue
			}
			provides := stepProvidesWhatFromTo(allSteps[i].recipe, allSteps[i].step, allSteps[j].step)
			if provides != "" {
				if _, err := fmt.Fprintf(&mermaid, "\tStep%d -->|%s| Step%d;\n",
					mealGraphID(allSteps[i].componentIndex, allSteps[i].loc),
					provides,
					mealGraphID(allSteps[j].componentIndex, allSteps[j].loc)); err != nil {
					observability.AcknowledgeError(err, g.logger, span, "writing mermaid meal step edge")
				}
			}
		}
	}

	// Subgraphs per component
	for compIdx, comp := range meal.Components {
		recipe := &comp.Recipe
		subgraphLabel := fmt.Sprintf("%s: %s", comp.ComponentType, recipe.Name)
		if _, err := fmt.Fprintf(&mermaid, "subgraph comp%d [\"%s\"]\n", compIdx, subgraphLabel); err != nil {
			observability.AcknowledgeError(err, g.logger, span, "writing mermaid meal subgraph header")
		}
		for j := range allSteps {
			if allSteps[j].componentIndex == compIdx {
				if _, err := fmt.Fprintf(&mermaid, "Step%d\n", mealGraphID(allSteps[j].componentIndex, allSteps[j].loc)); err != nil {
					observability.AcknowledgeError(err, g.logger, span, "writing mermaid meal subgraph step")
				}
			}
		}
		mermaid.WriteString("end\n")
	}

	return mermaid.String()
}

// stepLabelForMealDiagram returns a display label for a step in meal diagrams.
func stepLabelForMealDiagram(item mealStepItem) string {
	baseLabel := stepLabelForDiagram(item.loc, item.step, item.recipe)
	if item.recipe.Name != "" {
		return fmt.Sprintf("%s: %s", item.recipe.Name, baseLabel)
	}
	return baseLabel
}

// stepLabelForDiagram returns a display label for a step in diagrams.
func stepLabelForDiagram(loc stepLocation, step *mealplanning.RecipeStep, recipe *mealplanning.Recipe) string {
	if loc.recipeIndex == 0 {
		return fmt.Sprintf("Step #%d (%s)", loc.stepIndex+1, step.Preparation.Name)
	}
	assoc := recipe.AssociatedRecipes[loc.recipeIndex-1]
	return fmt.Sprintf("%s: Step #%d (%s)", assoc.Name, loc.stepIndex+1, step.Preparation.Name)
}

var colorNames = []string{
	"CornflowerBlue",
	"purple",
	"Salmon",
	"PeachPuff",
	"Tomato",
	"Orange",
	"Chocolate",
	"LemonChiffon",
}

const graphvizPreamble = `strict digraph {
	fontname="Outfit,Helvetica,Arial,sans-serif";
	node [fontname="Helvetica,Arial,sans-serif" shape=egg];
	edge [fontname="Helvetica,Arial,sans-serif"];
    concentrate=true;
    compound=true;
	rankdir=TB;
	start=1;
`

func (g *recipeAnalyzer) RenderGraphvizDiagramForRecipe(ctx context.Context, recipe *mealplanning.Recipe) string {
	_, span := g.tracer.StartSpan(ctx)
	defer span.End()

	var graphViz strings.Builder
	graphViz.WriteString(graphvizPreamble + "\n")

	allSteps := allRecipeSteps(recipe)
	for _, item := range allSteps {
		label := stepLabelForDiagram(item.loc, item.step, recipe)
		gid := graphIDForStepLocation(item.loc)
		if _, err := fmt.Fprintf(&graphViz, "\tStep%d [label=%q];\n", gid, label); err != nil {
			observability.AcknowledgeError(err, g.logger, span, "writing graphviz step node")
		}
	}

	for i := range allSteps {
		for j := range allSteps {
			if i == j {
				continue
			}

			if provides := stepProvidesWhatFromTo(recipe, allSteps[i].step, allSteps[j].step); provides != "" {
				stepLabel := ""
				if allSteps[i].step.MinEstimatedTimeInSeconds != nil && *allSteps[i].step.MinEstimatedTimeInSeconds > 0 {
					stepLabel = durafmt.Parse(time.Duration(*allSteps[i].step.MinEstimatedTimeInSeconds) * time.Second).String()
				}

				if _, err := fmt.Fprintf(&graphViz, "\tStep%d -> Step%d [color=\"black\" label=%q];\n",
					graphIDForStepLocation(allSteps[i].loc), graphIDForStepLocation(allSteps[j].loc), stepLabel); err != nil {
					observability.AcknowledgeError(err, g.logger, span, "writing graphviz step edge")
				}
			}
		}
	}

	for i := range recipe.PrepTasks {
		prepTask := recipe.PrepTasks[i]

		if _, err := fmt.Fprintf(&graphViz, "\n\tsubgraph cluster_%d {\n\t\tnode [style=filled];\n\t\t", i); err != nil {
			observability.AcknowledgeError(err, g.logger, span, "writing graphviz subgraph header")
		}
		for j := range prepTask.TaskSteps {
			var arrow string
			if j != len(prepTask.TaskSteps)-1 {
				arrow = " -> "
			}
			stepIdx := recipe.FindStepIndexByID(prepTask.TaskSteps[j].BelongsToRecipeStep)
			if stepIdx >= 0 {
				if _, err := fmt.Fprintf(&graphViz, "Step%d%s", stepIdx+1, arrow); err != nil {
					observability.AcknowledgeError(err, g.logger, span, "writing graphviz subgraph step")
				}
			}
		}
		colorToUse := colorNames[i%len(colorNames)]

		if _, err := fmt.Fprintf(&graphViz, ";\n\t\tlabel = \"(prep task #%d)\";\n\t\tcolor=%s;\n\t}\n", i+1, colorToUse); err != nil {
			observability.AcknowledgeError(err, g.logger, span, "writing graphviz subgraph footer")
		}
	}

	graphViz.WriteString("}\n")

	return graphViz.String()
}

func (g *recipeAnalyzer) RenderGraphvizDiagramForMeal(ctx context.Context, meal *mealplanning.Meal) string {
	_, span := g.tracer.StartSpan(ctx)
	defer span.End()

	var graphViz strings.Builder
	graphViz.WriteString(graphvizPreamble + "\n")

	allSteps := allMealSteps(meal)
	for i := range allSteps {
		label := stepLabelForMealDiagram(allSteps[i])
		gid := mealGraphID(allSteps[i].componentIndex, allSteps[i].loc)
		if _, err := fmt.Fprintf(&graphViz, "\tStep%d [label=%q];\n", gid, label); err != nil {
			observability.AcknowledgeError(err, g.logger, span, "writing graphviz meal step node")
		}
	}

	for i := range allSteps {
		for j := range allSteps {
			if i == j {
				continue
			}
			if allSteps[i].componentIndex != allSteps[j].componentIndex {
				continue
			}
			if provides := stepProvidesWhatFromTo(allSteps[i].recipe, allSteps[i].step, allSteps[j].step); provides != "" {
				stepLabel := ""
				if allSteps[i].step.MinEstimatedTimeInSeconds != nil && *allSteps[i].step.MinEstimatedTimeInSeconds > 0 {
					stepLabel = durafmt.Parse(time.Duration(*allSteps[i].step.MinEstimatedTimeInSeconds) * time.Second).String()
				}
				if _, err := fmt.Fprintf(&graphViz, "\tStep%d -> Step%d [color=\"black\" label=%q];\n",
					mealGraphID(allSteps[i].componentIndex, allSteps[i].loc),
					mealGraphID(allSteps[j].componentIndex, allSteps[j].loc),
					stepLabel); err != nil {
					observability.AcknowledgeError(err, g.logger, span, "writing graphviz meal step edge")
				}
			}
		}
	}

	for compIdx, comp := range meal.Components {
		subgraphLabel := fmt.Sprintf("%s: %s", comp.ComponentType, comp.Recipe.Name)
		colorToUse := colorNames[compIdx%len(colorNames)]
		if _, err := fmt.Fprintf(&graphViz, "\n\tsubgraph cluster_comp%d {\n\t\tnode [style=filled];\n\t\t", compIdx); err != nil {
			observability.AcknowledgeError(err, g.logger, span, "writing graphviz meal subgraph header")
		}
		var stepIDs []string
		for j := range allSteps {
			if allSteps[j].componentIndex == compIdx {
				stepIDs = append(stepIDs, fmt.Sprintf("Step%d", mealGraphID(allSteps[j].componentIndex, allSteps[j].loc)))
			}
		}
		if _, err := fmt.Fprintf(&graphViz, "%s;\n\t\tlabel = %q;\n\t\tcolor=%s;\n\t}\n", strings.Join(stepIDs, " "), subgraphLabel, colorToUse); err != nil {
			observability.AcknowledgeError(err, g.logger, span, "writing graphviz meal subgraph footer")
		}
	}

	graphViz.WriteString("}\n")

	return graphViz.String()
}

func FindUnitigsByLength(g *simple.DirectedGraph) map[int][][]graph.Node {
	results := make(map[int][][]graph.Node)

	nodes := g.Nodes()
	for nodes.Next() {
		n := nodes.Node()

		// Start of a chain: In-degree is not 1, but Out-degree is exactly 1
		if g.To(n.ID()).Len() != 1 && g.From(n.ID()).Len() == 1 {
			chain := []graph.Node{n}
			curr := n

			for {
				outputs := g.From(curr.ID())
				if outputs.Len() != 1 {
					break
				}

				outputs.Next()
				next := outputs.Node()

				if g.To(next.ID()).Len() != 1 {
					chain = append(chain, next)
					break
				}

				curr = next
				chain = append(chain, curr)
			}

			// Filter by number of nodes
			nodeCount := len(chain)
			if nodeCount >= 3 {
				results[nodeCount] = append(results[nodeCount], chain)
			}
		}
	}

	for length, chains := range results {
		sort.Slice(chains, func(i, j int) bool {
			return chains[i][0].ID() < chains[j][0].ID()
		})
		results[length] = chains
	}

	return results
}
