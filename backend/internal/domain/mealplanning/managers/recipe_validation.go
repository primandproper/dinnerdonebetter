package managers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipevalidator"
)

// checkRecipeForCreation is the gate every new recipe passes before it is written, whether a
// member wrote it or cloned it. It holds the rules that need a read: that the bridge rows a step
// names agree with the step (and the fields they imply, which it fills in), and that the recipes
// this one draws products from do not, somewhere down the line, draw on it.
//
// Every refusal wraps ErrInvalidRecipeInput, which is what the error mapper reads to answer
// InvalidArgument; the rule's own message is joined beside it, for logs and traces.
func (m *mealPlanningManager) checkRecipeForCreation(ctx context.Context, input *types.RecipeDatabaseCreationInput) error {
	if err := input.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("%w: %w", types.ErrInvalidRecipeInput, err)
	}

	if err := m.checkRecipeBridges(ctx, input); err != nil {
		return err
	}

	return types.CheckRecipeDependencies(ctx, input.ID, input.CrossRecipeDependencies(), m.recipeDependencies)
}

// checkRecipeBridges validates the bridge rows a recipe's steps name against the steps naming
// them, and populates the fields they imply. A recipe naming no bridge rows has nothing to check.
func (m *mealPlanningManager) checkRecipeBridges(ctx context.Context, input *types.RecipeDatabaseCreationInput) error {
	vipIDs := input.GetAllValidIngredientPreparationIDs()
	vimuIDs := input.GetAllValidIngredientMeasurementUnitIDs()
	vpiIDs := input.GetAllValidPreparationInstrumentIDs()
	vpvIDs := input.GetAllValidPreparationVesselIDs()

	if len(vipIDs) == 0 && len(vimuIDs) == 0 && len(vpiIDs) == 0 && len(vpvIDs) == 0 {
		return nil
	}

	vipMap, err := m.db.GetValidIngredientPreparationsByIDs(ctx, vipIDs)
	if err != nil {
		return fmt.Errorf("fetching valid ingredient preparations: %w", err)
	}

	vimuMap, err := m.db.GetValidIngredientMeasurementUnitsByIDs(ctx, vimuIDs)
	if err != nil {
		return fmt.Errorf("fetching valid ingredient measurement units: %w", err)
	}

	vpiMap, err := m.db.GetValidPreparationInstrumentsByIDs(ctx, vpiIDs)
	if err != nil {
		return fmt.Errorf("fetching valid preparation instruments: %w", err)
	}

	vpvMap, err := m.db.GetValidPreparationVesselsByIDs(ctx, vpvIDs)
	if err != nil {
		return fmt.Errorf("fetching valid preparation vessels: %w", err)
	}

	if err = recipevalidator.NewRecipeValidator(vipMap, vimuMap, vpiMap, vpvMap).ValidateAndPopulate(input); err != nil {
		return fmt.Errorf("%w: %w", types.ErrInvalidRecipeInput, err)
	}

	return nil
}

// checkIngredientDependency refuses an ingredient whose product comes from a recipe that would,
// once the ingredient is in place, lead back to the recipe holding it. An ingredient drawing on
// its own recipe, or on none, has nothing to check.
func (m *mealPlanningManager) checkIngredientDependency(ctx context.Context, recipeID string, productRecipeID *string) error {
	if productRecipeID == nil || *productRecipeID == "" || *productRecipeID == recipeID {
		return nil
	}

	dependencies, err := m.recipeDependencies(ctx, recipeID)
	if err != nil {
		return fmt.Errorf("reading the dependencies of recipe %s: %w", recipeID, err)
	}
	dependencies[*productRecipeID] = true

	return types.CheckRecipeDependencies(ctx, recipeID, dependencies, m.recipeDependencies)
}

// recipeDependencies reads the recipes a stored recipe draws products from. A recipe that is not
// there draws on nothing, so a reference to one cannot close a cycle.
func (m *mealPlanningManager) recipeDependencies(ctx context.Context, recipeID string) (map[string]bool, error) {
	recipe, err := m.db.GetRecipe(ctx, recipeID)
	if errors.Is(err, sql.ErrNoRows) {
		return map[string]bool{}, nil
	}
	if err != nil {
		return nil, err
	}

	return recipe.CrossRecipeDependencies(), nil
}
