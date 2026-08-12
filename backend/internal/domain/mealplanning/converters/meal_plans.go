package converters

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v10/identifiers"
)

// ConvertMealPlanToMealPlanUpdateRequestInput creates a MealPlanUpdateRequestInput from a MealPlan.
func ConvertMealPlanToMealPlanUpdateRequestInput(input *types.MealPlan) *types.MealPlanUpdateRequestInput {
	x := &types.MealPlanUpdateRequestInput{
		BelongsToAccount: &input.BelongsToAccount,
		Notes:            &input.Notes,
		VotingDeadline:   &input.VotingDeadline,
	}

	return x
}

// ConvertMealPlanCreationRequestInputToMealPlanDatabaseCreationInput creates a MealPlanDatabaseCreationInput from a MealPlanCreationRequestInput.
func ConvertMealPlanCreationRequestInputToMealPlanDatabaseCreationInput(input *types.MealPlanCreationRequestInput) *types.MealPlanDatabaseCreationInput {
	mealPlanID := identifiers.New()
	events := []*types.MealPlanEventDatabaseCreationInput{}
	for _, e := range input.Events {
		eventInput := ConvertMealPlanEventCreationRequestInputToMealPlanEventDatabaseCreationInput(e)
		eventInput.BelongsToMealPlan = mealPlanID
		events = append(events, eventInput)
	}

	// Convert selections from creation request input to database creation input
	selections := []*types.MealPlanRecipeOptionSelectionDatabaseCreationInput{}
	for _, s := range input.Selections {
		selections = append(selections, &types.MealPlanRecipeOptionSelectionDatabaseCreationInput{
			ID:                  identifiers.New(),
			RecipeID:            s.RecipeID,
			RecipeStepID:        s.RecipeStepID,
			IngredientIndex:     s.IngredientIndex,
			SelectedOptionIndex: s.SelectedOptionIndex,
			SelectionType:       s.SelectionType,
			// BelongsToMealPlanOption will be set later when matching with options
		})
	}

	x := &types.MealPlanDatabaseCreationInput{
		ID:    mealPlanID,
		Notes: input.Notes,
		// CreatedByUser is intentionally left empty here; the caller (the manager) sets it from the
		// authenticated creator ID. Generating a random ID would persist garbage if a caller forgot to.
		VotingDeadline: input.VotingDeadline,
		Events:         events,
		ElectionMethod: input.ElectionMethod,
		Selections:     selections,
	}

	return x
}

// ConvertMealPlanToMealPlanCreationRequestInput builds a MealPlanCreationRequestInput from a MealPlan.
func ConvertMealPlanToMealPlanCreationRequestInput(mealPlan *types.MealPlan) *types.MealPlanCreationRequestInput {
	events := []*types.MealPlanEventCreationRequestInput{}
	for _, evt := range mealPlan.Events {
		events = append(events, ConvertMealPlanEventToMealPlanEventCreationRequestInput(evt))
	}

	return &types.MealPlanCreationRequestInput{
		Notes:          mealPlan.Notes,
		VotingDeadline: mealPlan.VotingDeadline,
		Events:         events,
		ElectionMethod: mealPlan.ElectionMethod,
	}
}

// ConvertMealPlanToMealPlanDatabaseCreationInput builds a MealPlanDatabaseCreationInput from a MealPlan.
func ConvertMealPlanToMealPlanDatabaseCreationInput(mealPlan *types.MealPlan) *types.MealPlanDatabaseCreationInput {
	events := []*types.MealPlanEventDatabaseCreationInput{}
	for _, event := range mealPlan.Events {
		events = append(events, ConvertMealPlanEventToMealPlanEventDatabaseCreationInput(event))
	}

	selections := []*types.MealPlanRecipeOptionSelectionDatabaseCreationInput{}
	for _, selection := range mealPlan.Selections {
		selections = append(selections, ConvertMealPlanRecipeOptionSelectionToMealPlanRecipeOptionSelectionDatabaseCreationInput(selection))
	}

	return &types.MealPlanDatabaseCreationInput{
		ID:               mealPlan.ID,
		Notes:            mealPlan.Notes,
		VotingDeadline:   mealPlan.VotingDeadline,
		Events:           events,
		ElectionMethod:   mealPlan.ElectionMethod,
		BelongsToAccount: mealPlan.BelongsToAccount,
		CreatedByUser:    mealPlan.CreatedByUser,
		Selections:       selections,
	}
}
