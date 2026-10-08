package mealplanning

import (
	"context"
	"fmt"
	"maps"
	"slices"
)

// RecipeDependencyReader returns the IDs of the recipes a recipe draws a product from. A recipe
// that does not exist has none.
type RecipeDependencyReader func(ctx context.Context, recipeID string) (map[string]bool, error)

// CrossRecipeDependencies returns the IDs of the other recipes this one's ingredients draw a
// product from.
func (x *RecipeDatabaseCreationInput) CrossRecipeDependencies() map[string]bool {
	dependencies := map[string]bool{}
	for _, step := range x.Steps {
		for _, ingredient := range step.Ingredients {
			addCrossRecipeDependency(dependencies, x.ID, ingredient.RecipeStepProductRecipeID)
		}
	}

	return dependencies
}

// CrossRecipeDependencies returns the IDs of the other recipes this one's ingredients draw a
// product from.
func (x *Recipe) CrossRecipeDependencies() map[string]bool {
	dependencies := map[string]bool{}
	for _, step := range x.Steps {
		for _, ingredient := range step.Ingredients {
			addCrossRecipeDependency(dependencies, x.ID, ingredient.RecipeStepProductRecipeID)
		}
	}

	return dependencies
}

func addCrossRecipeDependency(dependencies map[string]bool, recipeID string, dependencyID *string) {
	if dependencyID != nil && *dependencyID != "" && *dependencyID != recipeID {
		dependencies[*dependencyID] = true
	}
}

// CheckRecipeDependencies refuses a recipe whose dependencies, once it has them, would lead back
// to it.
//
// It is the cross-recipe half of the rule that a recipe's products flow one way; the in-recipe
// half is recipeanalysis' ValidateRecipeCreationRequestInputIsDAG. dependencies are the ones
// recipeID is about to have; every other recipe's are read through dependenciesOf. The refusal
// wraps ErrInvalidRecipeInput, because a recipe that would draw on itself is a malformed request.
func CheckRecipeDependencies(ctx context.Context, recipeID string, dependencies map[string]bool, dependenciesOf RecipeDependencyReader) error {
	if dependencies[recipeID] {
		return fmt.Errorf("%w: recipe %s cannot draw a product from itself", ErrInvalidRecipeInput, recipeID)
	}

	visited := map[string]bool{}
	onPath := map[string]bool{}

	var visit func(id string) error
	visit = func(id string) error {
		if onPath[id] {
			return fmt.Errorf("%w: recipe %s is part of a dependency cycle", ErrInvalidRecipeInput, id)
		}
		if visited[id] {
			return nil
		}

		visited[id] = true
		onPath[id] = true
		defer delete(onPath, id)

		next := dependencies
		if id != recipeID {
			var err error
			if next, err = dependenciesOf(ctx, id); err != nil {
				return fmt.Errorf("reading the dependencies of recipe %s: %w", id, err)
			}
		}

		// Sorted so that a recipe in more than one cycle is always reported by the same one.
		for _, dependency := range slices.Sorted(maps.Keys(next)) {
			if err := visit(dependency); err != nil {
				return err
			}
		}

		return nil
	}

	return visit(recipeID)
}
