package managers

import (
	"context"
	"database/sql"
)

// The checks below are how a write to something a recipe or a meal plan is made of learns who
// may make it. Each asks the database an owner-scoped question and answers sql.ErrNoRows when
// the answer is no, so a caller who names something they do not own is told it is not there,
// exactly as ArchiveRecipe and ArchiveMealPlan tell them when their own owner-scoped statements
// match no row.

// requireRecipeOwnership refuses anyone but a recipe's author.
func (m *mealPlanningManager) requireRecipeOwnership(ctx context.Context, recipeID, ownerID string) error {
	owned, err := m.db.RecipeIsOwnedBy(ctx, recipeID, ownerID)
	if err != nil {
		return err
	}

	if !owned {
		return sql.ErrNoRows
	}

	return nil
}

// requireRecipeStepOwnership refuses anyone but a recipe's author, and a step that is not the
// recipe's: the writes beneath a step are keyed by the step alone, so a step borrowed from
// somebody else's recipe would otherwise pass for one of the caller's.
func (m *mealPlanningManager) requireRecipeStepOwnership(ctx context.Context, recipeID, recipeStepID, ownerID string) error {
	if err := m.requireRecipeOwnership(ctx, recipeID, ownerID); err != nil {
		return err
	}

	exists, err := m.db.RecipeStepExists(ctx, recipeID, recipeStepID)
	if err != nil {
		return err
	}

	if !exists {
		return sql.ErrNoRows
	}

	return nil
}

// requireMealPlanAccess refuses anyone outside the account a meal plan belongs to.
func (m *mealPlanningManager) requireMealPlanAccess(ctx context.Context, mealPlanID, accountID string) error {
	exists, err := m.db.MealPlanExists(ctx, mealPlanID, accountID)
	if err != nil {
		return err
	}

	if !exists {
		return sql.ErrNoRows
	}

	return nil
}

// requireMealPlanOptionAccess refuses anyone outside the account a meal plan option, by way of
// its event and plan, belongs to. It is for the writes that name an option and nothing above it.
func (m *mealPlanningManager) requireMealPlanOptionAccess(ctx context.Context, mealPlanOptionID, accountID string) error {
	belongs, err := m.db.MealPlanOptionBelongsToAccount(ctx, mealPlanOptionID, accountID)
	if err != nil {
		return err
	}

	if !belongs {
		return sql.ErrNoRows
	}

	return nil
}

// requireMealPlanEventAccess refuses anyone outside a meal plan's account, and an event that is
// not the plan's.
func (m *mealPlanningManager) requireMealPlanEventAccess(ctx context.Context, mealPlanID, mealPlanEventID, accountID string) error {
	if err := m.requireMealPlanAccess(ctx, mealPlanID, accountID); err != nil {
		return err
	}

	exists, err := m.db.MealPlanEventExists(ctx, mealPlanID, mealPlanEventID)
	if err != nil {
		return err
	}

	if !exists {
		return sql.ErrNoRows
	}

	return nil
}

// requireMealPlanTaskAccess refuses anyone outside a meal plan's account, and a task that is not
// the plan's.
func (m *mealPlanningManager) requireMealPlanTaskAccess(ctx context.Context, mealPlanID, mealPlanTaskID, accountID string) error {
	if err := m.requireMealPlanAccess(ctx, mealPlanID, accountID); err != nil {
		return err
	}

	exists, err := m.db.MealPlanTaskExists(ctx, mealPlanID, mealPlanTaskID)
	if err != nil {
		return err
	}

	if !exists {
		return sql.ErrNoRows
	}

	return nil
}

// requireMealPlanGroceryListItemAccess refuses anyone outside a meal plan's account, and a grocery
// list item that is not the plan's.
func (m *mealPlanningManager) requireMealPlanGroceryListItemAccess(ctx context.Context, mealPlanID, mealPlanGroceryListItemID, accountID string) error {
	if err := m.requireMealPlanAccess(ctx, mealPlanID, accountID); err != nil {
		return err
	}

	exists, err := m.db.MealPlanGroceryListItemExists(ctx, mealPlanID, mealPlanGroceryListItemID)
	if err != nil {
		return err
	}

	if !exists {
		return sql.ErrNoRows
	}

	return nil
}

// requireMealPlanOptionInPlanAccess refuses anyone outside a meal plan's account, and an option
// that is not one of the named event's, or an event that is not the plan's.
func (m *mealPlanningManager) requireMealPlanOptionInPlanAccess(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, accountID string) error {
	if err := m.requireMealPlanAccess(ctx, mealPlanID, accountID); err != nil {
		return err
	}

	exists, err := m.db.MealPlanOptionExists(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID)
	if err != nil {
		return err
	}

	if !exists {
		return sql.ErrNoRows
	}

	return nil
}
