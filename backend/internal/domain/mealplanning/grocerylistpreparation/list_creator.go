package grocerylistpreparation

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	"github.com/primandproper/platform-go/v7/identifiers"
	"github.com/primandproper/platform-go/v7/numbers"
	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/tracing"

	"github.com/shopspring/decimal"
)

// groceryListExcludedIngredientSlugs contains ingredient slugs that should not appear on grocery lists
// (e.g. water, which is assumed to be available from the tap).
var groceryListExcludedIngredientSlugs = map[string]struct{}{
	"water": {},
}

// GroceryListCreator creates meal plan grocery lists for a given meal plan.
type GroceryListCreator interface {
	GenerateGroceryListInputs(ctx context.Context, mealPlan *mealplanning.MealPlan) ([]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput, error)
}

type groceryListCreator struct {
	logger logging.Logger
	tracer tracing.Tracer
}

func NewGroceryListCreator(logger logging.Logger, tracerProvider tracing.TracerProvider) GroceryListCreator {
	return &groceryListCreator{
		logger: logging.NewNamedLogger(logger, "grocery_list_creator"),
		tracer: tracing.NewNamedTracer(tracerProvider, "grocery_list_creator"),
	}
}

// mergeQuantityNeeded folds an incoming ingredient quantity into an existing grocery list item.
// A missing max on either side is treated as equal to that side's min, so the merged max is always
// summed alongside the min and can never end up below the merged min (which the old "sum max only
// when both sides have one" logic could produce).
func mergeQuantityNeeded(existing *mealplanning.MealPlanGroceryListItemDatabaseCreationInput, minQty float32, maxQty *float32) {
	existingMax := existing.MinQuantityNeeded
	if existing.MaxQuantityNeeded != nil {
		existingMax = *existing.MaxQuantityNeeded
	}

	incomingMax := minQty
	if maxQty != nil {
		incomingMax = *maxQty
	}

	existing.MinQuantityNeeded += minQty
	mergedMax := existingMax + incomingMax
	existing.MaxQuantityNeeded = &mergedMax
}

// clearMismatchedAttribution nulls the source-attribution fields on a consolidated item when a
// merged contribution comes from a different option/recipe/step, so the item no longer claims to
// belong solely to its first contributor. The columns are nullable precisely for this case.
func clearMismatchedAttribution(existing *mealplanning.MealPlanGroceryListItemDatabaseCreationInput, optionID, recipeID, stepID string) {
	if existing.BelongsToMealPlanOption != nil && *existing.BelongsToMealPlanOption != optionID {
		existing.BelongsToMealPlanOption = nil
	}
	if existing.RecipeID != nil && *existing.RecipeID != recipeID {
		existing.RecipeID = nil
	}
	if existing.RecipeStepID != nil && *existing.RecipeStepID != stepID {
		existing.RecipeStepID = nil
	}
}

// processRecipeIngredients processes ingredients from a recipe (main or associated) and adds them to the grocery list.
func (g *groceryListCreator) processRecipeIngredients(
	recipe *mealplanning.Recipe,
	recipeScale decimal.Decimal,
	optionID string,
	mealPlanID string,
	optionGroups map[string]int,
	selectionLookup map[string]uint16,
	aggregatedInputs map[string]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput,
	optionInputs *[]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput,
	logger logging.Logger,
) {
	for _, step := range recipe.Steps {
		logger = logger.WithValue(mealplanningkeys.RecipeStepIDKey, step.ID)
		for _, ingredient := range step.Ingredients {
			if ingredient.Ingredient == nil || (ingredient.RecipeStepProductID != nil && *ingredient.RecipeStepProductID != "") {
				continue
			}
			if _, excluded := groceryListExcludedIngredientSlugs[ingredient.Ingredient.Slug]; excluded {
				continue
			}

			logger = logger.WithValue(mealplanningkeys.RecipeStepIngredientIDKey, ingredient.ID)

			scaleFactor := ingredient.ScaleFactor
			if scaleFactor <= 0 {
				scaleFactor = 1.0
			}
			effectiveScale := recipeScale.Mul(decimal.NewFromFloat32(scaleFactor))
			minQty := float32(effectiveScale.Mul(decimal.NewFromFloat32(ingredient.MinQuantity)).Truncate(2).InexactFloat64())
			var maxQty *float32
			if ingredient.MaxQuantity != nil {
				maximum := float32(effectiveScale.Mul(decimal.NewFromFloat32(*ingredient.MaxQuantity)).Truncate(2).InexactFloat64())
				maxQty = &maximum
			}

			optionGroupKey := fmt.Sprintf("%s:%d", step.ID, ingredient.Index)
			isOptionGroup := optionGroups[optionGroupKey] > 1

			if isOptionGroup {
				// This ingredient is part of an option group
				// Check if user made a selection for this option group
				selectionKey := fmt.Sprintf("%s:%d:%s", step.ID, ingredient.Index, mealplanning.MealPlanRecipeOptionSelectionTypeIngredient)
				selectedOptionIndex, hasSelection := selectionLookup[selectionKey]

				// Determine which option index to use:
				// - If user made a selection, use it
				// - If no selection, default to optionIndex=0 (the first/default option)
				targetOptionIndex := uint16(0)
				if hasSelection {
					targetOptionIndex = selectedOptionIndex
				}

				// Only add this ingredient if it matches the target option index
				if ingredient.OptionIndex != targetOptionIndex {
					continue
				}

				// Aggregate with existing item if same ingredient+unit (e.g. vegetable oil in step 6 and step 7)
				aggregationKey := fmt.Sprintf("%s:%s", ingredient.Ingredient.ID, ingredient.MeasurementUnit.ID)
				if existing, ok := aggregatedInputs[aggregationKey]; ok {
					mergeQuantityNeeded(existing, minQty, maxQty)
					clearMismatchedAttribution(existing, optionID, recipe.ID, step.ID)
					continue
				}
				// Check optionInputs for same ingredient+unit within this option
				var merged bool
				for _, existing := range *optionInputs {
					if existing.ValidIngredientID == ingredient.Ingredient.ID &&
						existing.ValidMeasurementUnitID == ingredient.MeasurementUnit.ID &&
						existing.BelongsToMealPlanOption != nil && *existing.BelongsToMealPlanOption == optionID {
						mergeQuantityNeeded(existing, minQty, maxQty)
						merged = true
						break
					}
				}
				if merged {
					continue
				}

				// Create grocery list item for the selected option with recipe context
				ingredientIndex := ingredient.Index
				optionIndex := ingredient.OptionIndex
				*optionInputs = append(*optionInputs, &mealplanning.MealPlanGroceryListItemDatabaseCreationInput{
					Status:                  mealplanning.MealPlanGroceryListItemStatusNeeds,
					ValidMeasurementUnitID:  ingredient.MeasurementUnit.ID,
					ValidIngredientID:       ingredient.Ingredient.ID,
					BelongsToMealPlan:       mealPlanID,
					BelongsToMealPlanOption: &optionID,
					RecipeID:                &recipe.ID,
					RecipeStepID:            &step.ID,
					IngredientIndex:         &ingredientIndex,
					OptionIndex:             &optionIndex,
					ID:                      identifiers.New(),
					MinQuantityNeeded:       minQty,
					MaxQuantityNeeded:       maxQty,
				})
			} else {
				// This ingredient is not part of an option group - aggregate by (ingredient, unit)
				// Same ingredient with different units (e.g., salt in tsp vs grams) becomes separate line items
				aggregationKey := fmt.Sprintf("%s:%s", ingredient.Ingredient.ID, ingredient.MeasurementUnit.ID)
				if existing, ok := aggregatedInputs[aggregationKey]; !ok {
					aggregatedInputs[aggregationKey] = &mealplanning.MealPlanGroceryListItemDatabaseCreationInput{
						BelongsToMealPlanOption: &optionID,
						RecipeID:                &recipe.ID,
						RecipeStepID:            &step.ID,
						Status:                  mealplanning.MealPlanGroceryListItemStatusNeeds,
						ValidMeasurementUnitID:  ingredient.MeasurementUnit.ID,
						ValidIngredientID:       ingredient.Ingredient.ID,
						BelongsToMealPlan:       mealPlanID,
						ID:                      identifiers.New(),
						MinQuantityNeeded:       minQty,
						MaxQuantityNeeded:       maxQty,
					}
				} else {
					mergeQuantityNeeded(existing, minQty, maxQty)
					clearMismatchedAttribution(existing, optionID, recipe.ID, step.ID)
				}
			}
		}
	}
}

func (g *groceryListCreator) GenerateGroceryListInputs(ctx context.Context, mealPlan *mealplanning.MealPlan) ([]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput, error) {
	_, span := g.tracer.StartSpan(ctx)
	defer span.End()

	// Map to track option groups: key is (recipeStepID, ingredientIndex), value is count of options
	optionGroups := make(map[string]int)
	// Map for aggregated items (non-option items): key is ingredientID
	aggregatedInputs := make(map[string]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput)
	// Slice for option items (items with alternatives): these are not aggregated
	optionInputs := []*mealplanning.MealPlanGroceryListItemDatabaseCreationInput{}

	logger := g.logger.Clone().WithValue(mealplanningkeys.MealPlanIDKey, mealPlan.ID)

	// Build a lookup map for user selections: key is (recipeStepID, ingredientIndex, selectionType), value is selectedOptionIndex
	selectionLookup := make(map[string]uint16)
	for _, selection := range mealPlan.Selections {
		if selection.SelectionType == mealplanning.MealPlanRecipeOptionSelectionTypeIngredient {
			selectionKey := fmt.Sprintf("%s:%d:%s", selection.RecipeStepID, selection.IngredientIndex, selection.SelectionType)
			selectionLookup[selectionKey] = selection.SelectedOptionIndex
		}
	}

	// First pass: identify option groups (ingredients with multiple options at the same index)
	// This includes both main recipes and their associated recipes
	for _, event := range mealPlan.Events {
		for _, option := range event.Options {
			if option.Chosen {
				for _, component := range option.Meal.Components {
					// Process main recipe
					for _, step := range component.Recipe.Steps {
						// Track how many options exist for each (stepID, index) combination
						indexCounts := make(map[uint16]int)
						for _, ingredient := range step.Ingredients {
							if ingredient.Ingredient != nil {
								indexCounts[ingredient.Index]++
							}
						}
						// Mark groups with multiple options
						for index, count := range indexCounts {
							if count > 1 {
								optionGroupKey := fmt.Sprintf("%s:%d", step.ID, index)
								optionGroups[optionGroupKey] = count
							}
						}
					}
					// Process associated recipes
					for _, associatedRecipe := range component.Recipe.AssociatedRecipes {
						for _, step := range associatedRecipe.Steps {
							// Track how many options exist for each (stepID, index) combination
							indexCounts := make(map[uint16]int)
							for _, ingredient := range step.Ingredients {
								if ingredient.Ingredient != nil {
									indexCounts[ingredient.Index]++
								}
							}
							// Mark groups with multiple options
							for index, count := range indexCounts {
								if count > 1 {
									optionGroupKey := fmt.Sprintf("%s:%d", step.ID, index)
									optionGroups[optionGroupKey] = count
								}
							}
						}
					}
				}
			}
		}
	}

	// Second pass: process ingredients from main recipes and associated recipes
	for _, event := range mealPlan.Events {
		logger = logger.WithValue(mealplanningkeys.MealPlanEventIDKey, event.ID)
		for _, option := range event.Options {
			if option.Chosen {
				mealScale := decimal.NewFromFloat32(option.MealScale)
				logger = logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, option.ID)
				for _, component := range option.Meal.Components {
					recipeScale := decimal.NewFromFloat32(component.RecipeScale).Mul(mealScale)
					logger = logger.WithValue(mealplanningkeys.RecipeIDKey, component.Recipe.ID)

					// Process main recipe ingredients
					g.processRecipeIngredients(
						&component.Recipe,
						recipeScale,
						option.ID,
						mealPlan.ID,
						optionGroups,
						selectionLookup,
						aggregatedInputs,
						&optionInputs,
						logger,
					)

					// Process associated recipe ingredients
					for _, associatedRecipe := range component.Recipe.AssociatedRecipes {
						logger = logger.WithValue(mealplanningkeys.RecipeIDKey, associatedRecipe.ID)
						g.processRecipeIngredients(
							associatedRecipe,
							recipeScale,
							option.ID,
							mealPlan.ID,
							optionGroups,
							selectionLookup,
							aggregatedInputs,
							&optionInputs,
							logger,
						)
					}
				}
			}
		}
	}

	// Combine aggregated items and option items
	dbInputs := make([]*mealplanning.MealPlanGroceryListItemDatabaseCreationInput, 0, len(aggregatedInputs)+len(optionInputs))
	for _, i := range aggregatedInputs {
		dbInputs = append(dbInputs, i)
	}
	dbInputs = append(dbInputs, optionInputs...)

	// Round quantities to the nearest tenth for cleaner grocery list display
	for _, item := range dbInputs {
		item.MinQuantityNeeded = numbers.RoundToDecimalPlaces(item.MinQuantityNeeded, 1)
		if item.MaxQuantityNeeded != nil {
			rounded := numbers.RoundToDecimalPlaces(*item.MaxQuantityNeeded, 1)
			item.MaxQuantityNeeded = &rounded
		}
	}

	return dbInputs, nil
}
