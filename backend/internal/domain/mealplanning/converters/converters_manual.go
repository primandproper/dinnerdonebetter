package converters

// The conversions in this file are hand-written: each does something the generator in
// cmd/tools/codegen/converters does not produce — it fails, it fans one value out into many, it
// defaults something, it needs a second entity to make sense of the first. exceptions.go names
// each one and says why.
//
// Everything else in this package is generated. A conversion that is a field copy with a handful
// of exceptions belongs there, where no destination field can be silently forgotten.

import (
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v12/identifiers"

	"github.com/ccoveille/go-safecast"
)

func mustnt(err error) {
	if err != nil {
		panic(err)
	}
}

// ConvertMealPlanEventToMealPlanEventCreationRequestInput builds a MealPlanEventCreationRequestInput from a meal plan.
func ConvertMealPlanEventToMealPlanEventCreationRequestInput(mealPlanEvent *mealplanning.MealPlanEvent) *mealplanning.MealPlanEventCreationRequestInput {
	options := []*mealplanning.MealPlanOptionCreationRequestInput{}
	for _, opt := range mealPlanEvent.Options {
		opt.BelongsToMealPlanEvent = mealPlanEvent.ID
		options = append(options, ConvertMealPlanOptionToMealPlanOptionCreationRequestInput(opt))
	}

	return &mealplanning.MealPlanEventCreationRequestInput{
		Notes:    mealPlanEvent.Notes,
		StartsAt: mealPlanEvent.StartsAt,
		EndsAt:   mealPlanEvent.EndsAt,
		MealName: mealPlanEvent.MealName,
		Options:  options,
	}
}

// ConvertMealPlanOptionVoteCreationRequestInputToMealPlanOptionVotesDatabaseCreationInput creates a MealPlanOptionVotesDatabaseCreationInput from a MealPlanOptionVoteCreationRequestInput.
func ConvertMealPlanOptionVoteCreationRequestInputToMealPlanOptionVotesDatabaseCreationInput(input *mealplanning.MealPlanOptionVoteCreationRequestInput) *mealplanning.MealPlanOptionVotesDatabaseCreationInput {
	var votes []*mealplanning.MealPlanOptionVoteDatabaseCreationInput
	for _, vote := range input.Votes {
		votes = append(votes, ConvertMealPlanOptionVoteCreationRequestInputToMealPlanOptionVoteDatabaseCreationInput(vote))
	}

	x := &mealplanning.MealPlanOptionVotesDatabaseCreationInput{
		Votes: votes,
	}

	return x
}

func ConvertMealPlanOptionVoteCreationRequestInputToMealPlanOptionVoteDatabaseCreationInput(input *mealplanning.MealPlanOptionVoteCreationInput) *mealplanning.MealPlanOptionVoteDatabaseCreationInput {
	return &mealplanning.MealPlanOptionVoteDatabaseCreationInput{
		ID:                      identifiers.New(),
		Notes:                   input.Notes,
		ByUser:                  input.ByUser,
		BelongsToMealPlanOption: input.BelongsToMealPlanOption,
		Rank:                    input.Rank,
		Abstain:                 input.Abstain,
	}
}

// ConvertMealPlanOptionVoteToMealPlanOptionVoteCreationRequestInput builds a MealPlanOptionVoteCreationRequestInput from a meal plan option vote.
func ConvertMealPlanOptionVoteToMealPlanOptionVoteCreationRequestInput(mealPlanOptionVote *mealplanning.MealPlanOptionVote) *mealplanning.MealPlanOptionVoteCreationRequestInput {
	return &mealplanning.MealPlanOptionVoteCreationRequestInput{
		Votes: []*mealplanning.MealPlanOptionVoteCreationInput{
			{
				ID:                      mealPlanOptionVote.ID,
				Rank:                    mealPlanOptionVote.Rank,
				Abstain:                 mealPlanOptionVote.Abstain,
				Notes:                   mealPlanOptionVote.Notes,
				BelongsToMealPlanOption: mealPlanOptionVote.BelongsToMealPlanOption,
				ByUser:                  mealPlanOptionVote.ByUser,
			},
		},
	}
}

// ConvertMealPlanOptionVoteToMealPlanOptionVotesDatabaseCreationInput builds a MealPlanOptionVotesDatabaseCreationInput from a meal plan option vote.
func ConvertMealPlanOptionVoteToMealPlanOptionVotesDatabaseCreationInput(mealPlanOptionVote *mealplanning.MealPlanOptionVote) *mealplanning.MealPlanOptionVotesDatabaseCreationInput {
	return &mealplanning.MealPlanOptionVotesDatabaseCreationInput{
		Votes: []*mealplanning.MealPlanOptionVoteDatabaseCreationInput{
			{
				ID:                      mealPlanOptionVote.ID,
				Rank:                    mealPlanOptionVote.Rank,
				Abstain:                 mealPlanOptionVote.Abstain,
				Notes:                   mealPlanOptionVote.Notes,
				BelongsToMealPlanOption: mealPlanOptionVote.BelongsToMealPlanOption,
				ByUser:                  mealPlanOptionVote.ByUser,
			},
		},
	}
}

// ConvertMealPlanRecipeOptionSelectionDatabaseCreationInputToMealPlanRecipeOptionSelectionDatabaseCreationInput creates a new DatabaseCreationInput with a new ID.
func ConvertMealPlanRecipeOptionSelectionDatabaseCreationInputToMealPlanRecipeOptionSelectionDatabaseCreationInput(input *mealplanning.MealPlanRecipeOptionSelectionCreationRequestInput, mealPlanOptionID string) *mealplanning.MealPlanRecipeOptionSelectionDatabaseCreationInput {
	return &mealplanning.MealPlanRecipeOptionSelectionDatabaseCreationInput{
		ID:                      identifiers.New(),
		BelongsToMealPlanOption: mealPlanOptionID,
		RecipeID:                input.RecipeID,
		RecipeStepID:            input.RecipeStepID,
		IngredientIndex:         input.IngredientIndex,
		SelectedOptionIndex:     input.SelectedOptionIndex,
		SelectionType:           input.SelectionType,
	}
}

// ConvertMealPlanCreationRequestInputToMealPlanDatabaseCreationInput creates a MealPlanDatabaseCreationInput from a MealPlanCreationRequestInput.
func ConvertMealPlanCreationRequestInputToMealPlanDatabaseCreationInput(input *mealplanning.MealPlanCreationRequestInput) *mealplanning.MealPlanDatabaseCreationInput {
	mealPlanID := identifiers.New()
	events := []*mealplanning.MealPlanEventDatabaseCreationInput{}
	for _, e := range input.Events {
		eventInput := ConvertMealPlanEventCreationRequestInputToMealPlanEventDatabaseCreationInput(e)
		eventInput.BelongsToMealPlan = mealPlanID
		events = append(events, eventInput)
	}

	// Convert selections from creation request input to database creation input
	selections := []*mealplanning.MealPlanRecipeOptionSelectionDatabaseCreationInput{}
	for _, s := range input.Selections {
		selections = append(selections, &mealplanning.MealPlanRecipeOptionSelectionDatabaseCreationInput{
			ID:                  identifiers.New(),
			RecipeID:            s.RecipeID,
			RecipeStepID:        s.RecipeStepID,
			IngredientIndex:     s.IngredientIndex,
			SelectedOptionIndex: s.SelectedOptionIndex,
			SelectionType:       s.SelectionType,
			// BelongsToMealPlanOption will be set later when matching with options
		})
	}

	x := &mealplanning.MealPlanDatabaseCreationInput{
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

// ConvertRecipePrepTaskToRecipePrepTaskUpdateRequestInput creates a RecipePrepTaskUpdateRequestInput from a RecipePrepTask.
func ConvertRecipePrepTaskToRecipePrepTaskUpdateRequestInput(input *mealplanning.RecipePrepTask) *mealplanning.RecipePrepTaskUpdateRequestInput {
	taskSteps := []*mealplanning.RecipePrepTaskStepUpdateRequestInput{}
	for _, x := range input.TaskSteps {
		y := x
		taskSteps = append(taskSteps, &mealplanning.RecipePrepTaskStepUpdateRequestInput{
			BelongsToRecipeStep:     &y.BelongsToRecipeStep,
			BelongsToRecipePrepTask: &y.BelongsToRecipePrepTask,
			SatisfiesRecipeStep:     &y.SatisfiesRecipeStep,
		})
	}
	x := &mealplanning.RecipePrepTaskUpdateRequestInput{
		Name:                               &input.Name,
		Description:                        &input.Description,
		Notes:                              &input.Notes,
		ExplicitStorageInstructions:        &input.ExplicitStorageInstructions,
		Optional:                           &input.Optional,
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: &input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
		StorageType:                        &input.StorageType,
		BelongsToRecipe:                    &input.BelongsToRecipe,
		TaskSteps:                          taskSteps,
	}

	return x
}

// ConvertRecipePrepTaskCreationRequestInputToRecipePrepTaskDatabaseCreationInput creates a DatabaseCreationInput from a CreationInput.
func ConvertRecipePrepTaskCreationRequestInputToRecipePrepTaskDatabaseCreationInput(input *mealplanning.RecipePrepTaskCreationRequestInput) *mealplanning.RecipePrepTaskDatabaseCreationInput {
	taskSteps := []*mealplanning.RecipePrepTaskStepDatabaseCreationInput{}
	for _, x := range input.RecipeSteps {
		taskSteps = append(taskSteps, &mealplanning.RecipePrepTaskStepDatabaseCreationInput{
			ID:                  identifiers.New(),
			BelongsToRecipeStep: x.BelongsToRecipeStep,
			SatisfiesRecipeStep: x.SatisfiesRecipeStep,
		})
	}

	x := &mealplanning.RecipePrepTaskDatabaseCreationInput{
		ID:                                 identifiers.New(),
		Name:                               input.Name,
		Description:                        input.Description,
		Notes:                              input.Notes,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		Optional:                           input.Optional,
		StorageType:                        input.StorageType,
		BelongsToRecipe:                    input.BelongsToRecipe,
		TaskSteps:                          taskSteps,
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
	}

	return x
}

// ConvertRecipePrepTaskWithinRecipeCreationRequestInputToRecipePrepTaskDatabaseCreationInput creates a DatabaseCreationInput from a CreationInput.
func ConvertRecipePrepTaskWithinRecipeCreationRequestInputToRecipePrepTaskDatabaseCreationInput(recipe *mealplanning.RecipeDatabaseCreationInput, input *mealplanning.RecipePrepTaskWithinRecipeCreationRequestInput) (*mealplanning.RecipePrepTaskDatabaseCreationInput, error) {
	x := &mealplanning.RecipePrepTaskDatabaseCreationInput{
		ID:                                 identifiers.New(),
		Name:                               input.Name,
		Description:                        input.Description,
		Notes:                              input.Notes,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		Optional:                           input.Optional,
		StorageType:                        input.StorageType,
		BelongsToRecipe:                    input.BelongsToRecipe,
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
	}

	x.TaskSteps = []*mealplanning.RecipePrepTaskStepDatabaseCreationInput{}
	for i, ts := range input.RecipeSteps {
		if rs := recipe.FindStepByIndex(ts.BelongsToRecipeStepIndex); rs != nil {
			x.TaskSteps = append(x.TaskSteps, &mealplanning.RecipePrepTaskStepDatabaseCreationInput{
				ID:                      identifiers.New(),
				BelongsToRecipeStep:     rs.ID,
				BelongsToRecipePrepTask: x.ID,
				SatisfiesRecipeStep:     ts.SatisfiesRecipeStep,
			})
		} else {
			return nil, fmt.Errorf("task step #%d has an invalid recipe step index", i+1)
		}
	}

	return x, nil
}

func ConvertRecipePrepTaskToRecipePrepTaskWithinRecipeCreationRequestInput(recipe *mealplanning.Recipe, input *mealplanning.RecipePrepTask) *mealplanning.RecipePrepTaskWithinRecipeCreationRequestInput {
	taskSteps := []*mealplanning.RecipePrepTaskStepWithinRecipeCreationRequestInput{}
	for _, x := range input.TaskSteps {
		taskSteps = append(taskSteps, ConvertRecipePrepTaskStepToRecipePrepTaskStepWithinRecipeCreationRequestInput(recipe, x))
	}

	return &mealplanning.RecipePrepTaskWithinRecipeCreationRequestInput{
		Name:                               input.Name,
		Description:                        input.Description,
		Notes:                              input.Notes,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		Optional:                           input.Optional,
		StorageType:                        input.StorageType,
		BelongsToRecipe:                    input.BelongsToRecipe,
		RecipeSteps:                        taskSteps,
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
	}
}

func ConvertRecipePrepTaskStepToRecipePrepTaskStepWithinRecipeCreationRequestInput(recipe *mealplanning.Recipe, input *mealplanning.RecipePrepTaskStep) *mealplanning.RecipePrepTaskStepWithinRecipeCreationRequestInput {
	var belongsToIndex uint32
	if x := recipe.FindStepByID(input.BelongsToRecipeStep); x != nil {
		belongsToIndex = x.Index
	}

	return &mealplanning.RecipePrepTaskStepWithinRecipeCreationRequestInput{
		BelongsToRecipeStepIndex: belongsToIndex,
		SatisfiesRecipeStep:      input.SatisfiesRecipeStep,
	}
}

// ConvertRecipeStepCompletionConditionCreationRequestInputToRecipeStepCompletionConditionDatabaseCreationInput creates a RecipeStepCompletionConditionDatabaseCreationInput from a RecipeStepCompletionConditionCreationRequestInput.
func ConvertRecipeStepCompletionConditionCreationRequestInputToRecipeStepCompletionConditionDatabaseCreationInput(recipeStep *mealplanning.RecipeStepDatabaseCreationInput, input *mealplanning.RecipeStepCompletionConditionCreationRequestInput) *mealplanning.RecipeStepCompletionConditionDatabaseCreationInput {
	recipeStepCompletionConditionID := identifiers.New()

	var ingredients []*mealplanning.RecipeStepCompletionConditionIngredientDatabaseCreationInput
	for _, i := range input.Ingredients {
		x := &mealplanning.RecipeStepCompletionConditionIngredientDatabaseCreationInput{
			ID:                                     identifiers.New(),
			RecipeStepIngredient:                   recipeStep.Ingredients[i].ID,
			BelongsToRecipeStepCompletionCondition: recipeStepCompletionConditionID,
		}

		ingredients = append(ingredients, x)
	}

	x := &mealplanning.RecipeStepCompletionConditionDatabaseCreationInput{
		ID:                  recipeStepCompletionConditionID,
		IngredientStateID:   input.IngredientStateID,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Notes:               input.Notes,
		Ingredients:         ingredients,
		Optional:            input.Optional,
	}

	return x
}

// ConvertRecipeStepCompletionConditionForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionDatabaseCreationInput creates a RecipeStepCompletionConditionDatabaseCreationInput from a RecipeStepCompletionConditionForExitingRecipeCreationRequestInput.
func ConvertRecipeStepCompletionConditionForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionDatabaseCreationInput(input *mealplanning.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput) *mealplanning.RecipeStepCompletionConditionDatabaseCreationInput {
	id := identifiers.New()

	var ingredients []*mealplanning.RecipeStepCompletionConditionIngredientDatabaseCreationInput
	for _, i := range input.Ingredients {
		x := ConvertRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionIngredientDatabaseCreationInput(i)
		x.BelongsToRecipeStepCompletionCondition = id
		ingredients = append(ingredients, x)
	}

	x := &mealplanning.RecipeStepCompletionConditionDatabaseCreationInput{
		ID:                  id,
		IngredientStateID:   input.IngredientStateID,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Notes:               input.Notes,
		Ingredients:         ingredients,
		Optional:            input.Optional,
	}

	return x
}

// ConvertRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionIngredientDatabaseCreationInput creates a RecipeStepCompletionConditionIngredientDatabaseCreationInput from a RecipeStepCompletionConditionCreationRequestInput.
func ConvertRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionIngredientDatabaseCreationInput(input *mealplanning.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput) *mealplanning.RecipeStepCompletionConditionIngredientDatabaseCreationInput {
	x := &mealplanning.RecipeStepCompletionConditionIngredientDatabaseCreationInput{
		ID:                   identifiers.New(),
		RecipeStepIngredient: input.RecipeStepIngredient,
	}

	return x
}

// ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionCreationRequestInput builds a RecipeStepCompletionConditionCreationRequestInput from a RecipeStepCompletionCondition.
func ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionCreationRequestInput(recipeStep *mealplanning.RecipeStep, recipeStepCompletionCondition *mealplanning.RecipeStepCompletionCondition) *mealplanning.RecipeStepCompletionConditionCreationRequestInput {
	var ingredients []uint64
	for _, ingredientIndex := range recipeStepCompletionCondition.Ingredients {
		for i, ingredient := range recipeStep.Ingredients {
			if ingredient.ID == ingredientIndex.RecipeStepIngredient {
				x, err := safecast.Convert[uint64](i)
				mustnt(err)
				ingredients = append(ingredients, x)
			}
		}
	}

	return &mealplanning.RecipeStepCompletionConditionCreationRequestInput{
		IngredientStateID:   recipeStepCompletionCondition.IngredientState.ID,
		BelongsToRecipeStep: recipeStepCompletionCondition.BelongsToRecipeStep,
		Notes:               recipeStepCompletionCondition.Notes,
		Ingredients:         ingredients,
		Optional:            recipeStepCompletionCondition.Optional,
	}
}

// ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionForExistingRecipeCreationRequestInput builds a RecipeStepCompletionConditionCreationRequestInput from a RecipeStepCompletionCondition.
func ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionForExistingRecipeCreationRequestInput(recipeStepCompletionCondition *mealplanning.RecipeStepCompletionCondition) *mealplanning.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput {
	var ingredients []*mealplanning.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput
	for _, i := range recipeStepCompletionCondition.Ingredients {
		x := ConvertRecipeStepCompletionConditionIngredientToRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput(i)
		ingredients = append(ingredients, x)
	}

	return &mealplanning.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput{
		IngredientStateID:   recipeStepCompletionCondition.IngredientState.ID,
		BelongsToRecipeStep: recipeStepCompletionCondition.BelongsToRecipeStep,
		Notes:               recipeStepCompletionCondition.Notes,
		Ingredients:         ingredients,
		Optional:            recipeStepCompletionCondition.Optional,
	}
}

// ConvertRecipeStepCompletionConditionIngredientToRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput builds a RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput from a RecipeStepCompletionCondition.
func ConvertRecipeStepCompletionConditionIngredientToRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput(recipeStepCompletionConditionIngredient *mealplanning.RecipeStepCompletionConditionIngredient) *mealplanning.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput {
	return &mealplanning.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput{
		RecipeStepIngredient: recipeStepCompletionConditionIngredient.RecipeStepIngredient,
	}
}

// ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionDatabaseCreationInput builds a RecipeStepCompletionConditionDatabaseCreationInput from a RecipeStepCompletionCondition.
func ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionDatabaseCreationInput(recipeStepCompletionCondition *mealplanning.RecipeStepCompletionCondition) *mealplanning.RecipeStepCompletionConditionDatabaseCreationInput {
	ingredients := []*mealplanning.RecipeStepCompletionConditionIngredientDatabaseCreationInput{}
	for _, ingredient := range recipeStepCompletionCondition.Ingredients {
		ingredients = append(ingredients, &mealplanning.RecipeStepCompletionConditionIngredientDatabaseCreationInput{
			ID:                                     ingredient.ID,
			BelongsToRecipeStepCompletionCondition: ingredient.BelongsToRecipeStepCompletionCondition,
			RecipeStepIngredient:                   ingredient.RecipeStepIngredient,
		})
	}

	return &mealplanning.RecipeStepCompletionConditionDatabaseCreationInput{
		ID:                  recipeStepCompletionCondition.ID,
		Optional:            recipeStepCompletionCondition.Optional,
		Notes:               recipeStepCompletionCondition.Notes,
		IngredientStateID:   recipeStepCompletionCondition.IngredientState.ID,
		BelongsToRecipeStep: recipeStepCompletionCondition.BelongsToRecipeStep,
		Ingredients:         ingredients,
	}
}

// ConvertRecipeStepIngredientCreationRequestInputToRecipeStepIngredientDatabaseCreationInput creates a RecipeStepIngredientDatabaseCreationInput from a RecipeStepIngredientCreationRequestInput.
// If input.Index is nil, it will be set to the provided arrayIndex.
func ConvertRecipeStepIngredientCreationRequestInputToRecipeStepIngredientDatabaseCreationInput(input *mealplanning.RecipeStepIngredientCreationRequestInput, arrayIndex uint16) *mealplanning.RecipeStepIngredientDatabaseCreationInput {
	index := arrayIndex
	if input.Index != nil {
		index = *input.Index
	}

	x := &mealplanning.RecipeStepIngredientDatabaseCreationInput{
		ID:                               identifiers.New(),
		ValidIngredientPreparationID:     input.ValidIngredientPreparationID,
		ValidIngredientMeasurementUnitID: input.ValidIngredientMeasurementUnitID,
		Name:                             input.Name,
		MinQuantity:                      input.MinQuantity,
		MaxQuantity:                      input.MaxQuantity,
		QuantityNotes:                    input.QuantityNotes,
		IngredientNotes:                  input.IngredientNotes,
		Optional:                         input.Optional,
		Index:                            index,
		OptionIndex:                      input.OptionIndex,
		ProductOfRecipeStepIndex:         input.ProductOfRecipeStepIndex,
		ProductOfRecipeStepProductIndex:  input.ProductOfRecipeStepProductIndex,
		RecipeStepProductRecipeID:        input.RecipeStepProductRecipeID,
		VesselIndex:                      input.VesselIndex,
		ToTaste:                          input.ToTaste,
		ProductPercentageToUse:           input.ProductPercentageToUse,
		ScaleFactor:                      scaleFactorOrDefault(input.ScaleFactor),
	}

	return x
}

// ConvertRecipeStepIngredientToRecipeStepIngredientCreationRequestInput builds a RecipeStepIngredientCreationRequestInput from a RecipeStepIngredient.
// Note: This conversion loses bridge table ID information since RecipeStepIngredient doesn't store them.
// If Index is 0, it will be set to nil so that the converter can use the array index during recipe creation.
func ConvertRecipeStepIngredientToRecipeStepIngredientCreationRequestInput(input *mealplanning.RecipeStepIngredient) *mealplanning.RecipeStepIngredientCreationRequestInput {
	var indexPtr *uint16
	if input.Index != 0 {
		indexPtr = new(input.Index)
	}
	return &mealplanning.RecipeStepIngredientCreationRequestInput{
		Name:                   input.Name,
		Optional:               input.Optional,
		MinQuantity:            input.MinQuantity,
		MaxQuantity:            input.MaxQuantity,
		QuantityNotes:          input.QuantityNotes,
		IngredientNotes:        input.IngredientNotes,
		Index:                  indexPtr,
		OptionIndex:            input.OptionIndex,
		VesselIndex:            input.VesselIndex,
		ToTaste:                input.ToTaste,
		ProductPercentageToUse: input.ProductPercentageToUse,
		ScaleFactor:            input.ScaleFactor,
	}
}

// ConvertRecipeStepInstrumentCreationRequestInputToRecipeStepInstrumentDatabaseCreationInput creates a RecipeStepInstrumentDatabaseCreationInput from a RecipeStepInstrumentCreationRequestInput.
// If input.Index is nil, it will be set to the provided arrayIndex.
func ConvertRecipeStepInstrumentCreationRequestInputToRecipeStepInstrumentDatabaseCreationInput(input *mealplanning.RecipeStepInstrumentCreationRequestInput, arrayIndex uint16) *mealplanning.RecipeStepInstrumentDatabaseCreationInput {
	index := arrayIndex
	if input.Index != nil {
		index = *input.Index
	}

	x := &mealplanning.RecipeStepInstrumentDatabaseCreationInput{
		ID:                              identifiers.New(),
		ValidPreparationInstrumentID:    input.ValidPreparationInstrumentID,
		RecipeStepProductID:             input.RecipeStepProductID,
		Name:                            input.Name,
		Notes:                           input.Notes,
		PreferenceRank:                  input.PreferenceRank,
		Optional:                        input.Optional,
		Index:                           index,
		OptionIndex:                     input.OptionIndex,
		MinQuantity:                     input.MinQuantity,
		MaxQuantity:                     input.MaxQuantity,
		ProductOfRecipeStepIndex:        input.ProductOfRecipeStepIndex,
		ProductOfRecipeStepProductIndex: input.ProductOfRecipeStepProductIndex,
		ScaleFactor:                     scaleFactorOrDefault(input.ScaleFactor),
	}

	return x
}

func scaleFactorOrDefault(v float32) float32 {
	if v <= 0 {
		return 1.0
	}
	return v
}

// ConvertRecipeStepInstrumentToRecipeStepInstrumentCreationRequestInput builds a RecipeStepInstrumentCreationRequestInput from a RecipeStepInstrument.
// Note: This conversion loses bridge table ID information since RecipeStepInstrument doesn't store them.
// If Index is 0, it will be set to nil so that the converter can use the array index during recipe creation.
func ConvertRecipeStepInstrumentToRecipeStepInstrumentCreationRequestInput(input *mealplanning.RecipeStepInstrument) *mealplanning.RecipeStepInstrumentCreationRequestInput {
	var indexPtr *uint16
	if input.Index != 0 {
		indexPtr = new(input.Index)
	}
	return &mealplanning.RecipeStepInstrumentCreationRequestInput{
		Name:                input.Name,
		RecipeStepProductID: input.RecipeStepProductID,
		Notes:               input.Notes,
		PreferenceRank:      input.PreferenceRank,
		Optional:            input.Optional,
		Index:               indexPtr,
		OptionIndex:         input.OptionIndex,
		MinQuantity:         input.MinQuantity,
		MaxQuantity:         input.MaxQuantity,
		ScaleFactor:         input.ScaleFactor,
	}
}

// ConvertRecipeStepProductToRecipeStepProductUpdateRequestInput creates a RecipeStepProductUpdateRequestInput from a RecipeStepProduct.
func ConvertRecipeStepProductToRecipeStepProductUpdateRequestInput(input *mealplanning.RecipeStepProduct) *mealplanning.RecipeStepProductUpdateRequestInput {
	if input == nil {
		return nil
	}

	x := &mealplanning.RecipeStepProductUpdateRequestInput{
		Name:                           &input.Name,
		Type:                           &input.Type,
		MeasurementUnitID:              &input.MeasurementUnit.ID,
		QuantityNotes:                  &input.QuantityNotes,
		BelongsToRecipeStep:            &input.BelongsToRecipeStep,
		Compostable:                    &input.Compostable,
		MinMeasurementQuantity:         input.MinMeasurementQuantity,
		MaxMeasurementQuantity:         input.MaxMeasurementQuantity,
		MinItemQuantity:                input.MinItemQuantity,
		MaxItemQuantity:                input.MaxItemQuantity,
		MinStorageDurationInSeconds:    input.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    input.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: input.MaxStorageTemperatureInCelsius,
		StorageInstructions:            &input.StorageInstructions,
		IsWaste:                        &input.IsWaste,
		IsLiquid:                       &input.IsLiquid,
		Index:                          &input.Index,
		ContainedInVesselIndex:         input.ContainedInVesselIndex,
	}

	return x
}

// ConvertRecipeStepProductCreationRequestInputToRecipeStepProductDatabaseCreationInput creates a RecipeStepProductDatabaseCreationInput from a RecipeStepProductCreationRequestInput.
func ConvertRecipeStepProductCreationRequestInputToRecipeStepProductDatabaseCreationInput(input *mealplanning.RecipeStepProductCreationRequestInput) *mealplanning.RecipeStepProductDatabaseCreationInput {
	if input == nil {
		return nil
	}

	x := &mealplanning.RecipeStepProductDatabaseCreationInput{
		ID:                             identifiers.New(),
		Name:                           input.Name,
		Type:                           input.Type,
		MeasurementUnitID:              input.MeasurementUnitID,
		QuantityNotes:                  input.QuantityNotes,
		Compostable:                    input.Compostable,
		MinMeasurementQuantity:         input.MinMeasurementQuantity,
		MaxMeasurementQuantity:         input.MaxMeasurementQuantity,
		MinItemQuantity:                input.MinItemQuantity,
		MaxItemQuantity:                input.MaxItemQuantity,
		MinStorageDurationInSeconds:    input.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    input.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: input.MaxStorageTemperatureInCelsius,
		StorageInstructions:            input.StorageInstructions,
		IsWaste:                        input.IsWaste,
		IsLiquid:                       input.IsLiquid,
		Index:                          input.Index,
		ContainedInVesselIndex:         input.ContainedInVesselIndex,
	}

	return x
}

// ConvertRecipeStepVesselCreationRequestInputToRecipeStepVesselDatabaseCreationInput creates a RecipeStepVesselDatabaseCreationInput from a RecipeStepVesselCreationRequestInput.
// If input.Index is nil, it will be set to the provided arrayIndex.
func ConvertRecipeStepVesselCreationRequestInputToRecipeStepVesselDatabaseCreationInput(input *mealplanning.RecipeStepVesselCreationRequestInput, arrayIndex uint16) *mealplanning.RecipeStepVesselDatabaseCreationInput {
	index := arrayIndex
	if input.Index != nil {
		index = *input.Index
	}

	x := &mealplanning.RecipeStepVesselDatabaseCreationInput{
		ID:                              identifiers.New(),
		ValidPreparationVesselID:        input.ValidPreparationVesselID,
		RecipeStepProductID:             input.RecipeStepProductID,
		Name:                            input.Name,
		Notes:                           input.Notes,
		MinQuantity:                     input.MinQuantity,
		MaxQuantity:                     input.MaxQuantity,
		Index:                           index,
		OptionIndex:                     input.OptionIndex,
		ProductOfRecipeStepIndex:        input.ProductOfRecipeStepIndex,
		ProductOfRecipeStepProductIndex: input.ProductOfRecipeStepProductIndex,
		VesselPreposition:               input.VesselPreposition,
		UnavailableAfterStep:            input.UnavailableAfterStep,
		ScaleFactor:                     scaleFactorOrDefault(input.ScaleFactor),
	}

	return x
}

// ConvertRecipeStepVesselToRecipeStepVesselCreationRequestInput builds a RecipeStepVesselCreationRequestInput from a RecipeStepVessel.
// Note: This conversion loses bridge table ID information since RecipeStepVessel doesn't store them.
// If Index is 0, it will be set to nil so that the converter can use the array index during recipe creation.
func ConvertRecipeStepVesselToRecipeStepVesselCreationRequestInput(input *mealplanning.RecipeStepVessel) *mealplanning.RecipeStepVesselCreationRequestInput {
	var indexPtr *uint16
	if input.Index != 0 {
		indexPtr = new(input.Index)
	}
	return &mealplanning.RecipeStepVesselCreationRequestInput{
		Name:                 input.Name,
		RecipeStepProductID:  input.RecipeStepProductID,
		Notes:                input.Notes,
		VesselPreposition:    input.VesselPreposition,
		UnavailableAfterStep: input.UnavailableAfterStep,
		Index:                indexPtr,
		OptionIndex:          input.OptionIndex,
		MinQuantity:          input.MinQuantity,
		MaxQuantity:          input.MaxQuantity,
		ScaleFactor:          input.ScaleFactor,
	}
}

// ConvertRecipeStepCreationRequestInputToRecipeStepDatabaseCreationInput creates a RecipeStepDatabaseCreationInput from a RecipeStepCreationRequestInput.
func ConvertRecipeStepCreationRequestInputToRecipeStepDatabaseCreationInput(input *mealplanning.RecipeStepCreationRequestInput) *mealplanning.RecipeStepDatabaseCreationInput {
	stepID := identifiers.New()

	ingredients := []*mealplanning.RecipeStepIngredientDatabaseCreationInput{}
	for i, ingredient := range input.Ingredients {
		convertedIngredient := ConvertRecipeStepIngredientCreationRequestInputToRecipeStepIngredientDatabaseCreationInput(ingredient, uint16(i))
		convertedIngredient.ID = identifiers.New()
		convertedIngredient.BelongsToRecipeStep = stepID
		ingredients = append(ingredients, convertedIngredient)
	}

	instruments := []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{}
	for i, instrument := range input.Instruments {
		convertedInstrument := ConvertRecipeStepInstrumentCreationRequestInputToRecipeStepInstrumentDatabaseCreationInput(instrument, uint16(i))
		convertedInstrument.ID = identifiers.New()
		convertedInstrument.BelongsToRecipeStep = stepID
		instruments = append(instruments, convertedInstrument)
	}

	vessels := []*mealplanning.RecipeStepVesselDatabaseCreationInput{}
	for i, vessel := range input.Vessels {
		convertedVessel := ConvertRecipeStepVesselCreationRequestInputToRecipeStepVesselDatabaseCreationInput(vessel, uint16(i))
		convertedVessel.ID = identifiers.New()
		convertedVessel.BelongsToRecipeStep = stepID
		vessels = append(vessels, convertedVessel)
	}

	products := []*mealplanning.RecipeStepProductDatabaseCreationInput{}
	for _, product := range input.Products {
		convertedProduct := ConvertRecipeStepProductCreationRequestInputToRecipeStepProductDatabaseCreationInput(product)
		convertedProduct.ID = identifiers.New()
		convertedProduct.BelongsToRecipeStep = stepID
		products = append(products, convertedProduct)
	}

	completionConditions := []*mealplanning.RecipeStepCompletionConditionDatabaseCreationInput{}
	// Create a temporary struct with ingredients populated for the completion condition converter
	tempStepForCompletionConditions := &mealplanning.RecipeStepDatabaseCreationInput{
		ID:          stepID,
		Ingredients: ingredients,
	}
	for _, completionCondition := range input.CompletionConditions {
		convertedCompletionCondition := ConvertRecipeStepCompletionConditionCreationRequestInputToRecipeStepCompletionConditionDatabaseCreationInput(
			tempStepForCompletionConditions,
			completionCondition,
		)
		convertedCompletionCondition.ID = identifiers.New()
		convertedCompletionCondition.BelongsToRecipeStep = stepID
		completionConditions = append(completionConditions, convertedCompletionCondition)
	}

	return &mealplanning.RecipeStepDatabaseCreationInput{
		ID:                        stepID,
		Index:                     input.Index,
		PreparationID:             input.PreparationID,
		MinEstimatedTimeInSeconds: input.MinEstimatedTimeInSeconds,
		MaxEstimatedTimeInSeconds: input.MaxEstimatedTimeInSeconds,
		MinTemperatureInCelsius:   input.MinTemperatureInCelsius,
		MaxTemperatureInCelsius:   input.MaxTemperatureInCelsius,
		Notes:                     input.Notes,
		Optional:                  input.Optional,
		ExplicitInstructions:      input.ExplicitInstructions,
		ConditionExpression:       input.ConditionExpression,
		StartTimerAutomatically:   input.StartTimerAutomatically,
		Ingredients:               ingredients,
		Instruments:               instruments,
		Vessels:                   vessels,
		Products:                  products,
		CompletionConditions:      completionConditions,
	}
}

// ConvertRecipeStepToRecipeStepCreationRequestInput builds a RecipeStepCreationRequestInput from a RecipeStep.
func ConvertRecipeStepToRecipeStepCreationRequestInput(recipeStep *mealplanning.RecipeStep) *mealplanning.RecipeStepCreationRequestInput {
	ingredients := []*mealplanning.RecipeStepIngredientCreationRequestInput{}
	for _, ingredient := range recipeStep.Ingredients {
		ingredients = append(ingredients, ConvertRecipeStepIngredientToRecipeStepIngredientCreationRequestInput(ingredient))
	}

	instruments := []*mealplanning.RecipeStepInstrumentCreationRequestInput{}
	for _, instrument := range recipeStep.Instruments {
		instruments = append(instruments, ConvertRecipeStepInstrumentToRecipeStepInstrumentCreationRequestInput(instrument))
	}

	vessels := []*mealplanning.RecipeStepVesselCreationRequestInput{}
	for _, vessel := range recipeStep.Vessels {
		vessels = append(vessels, ConvertRecipeStepVesselToRecipeStepVesselCreationRequestInput(vessel))
	}

	products := []*mealplanning.RecipeStepProductCreationRequestInput{}
	for _, product := range recipeStep.Products {
		products = append(products, ConvertRecipeStepProductToRecipeStepProductCreationRequestInput(product))
	}

	completionConditions := []*mealplanning.RecipeStepCompletionConditionCreationRequestInput{}
	for _, completionCondition := range recipeStep.CompletionConditions {
		completionConditions = append(completionConditions, ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionCreationRequestInput(recipeStep, completionCondition))
	}

	return &mealplanning.RecipeStepCreationRequestInput{
		Optional:                  recipeStep.Optional,
		Index:                     recipeStep.Index,
		PreparationID:             recipeStep.Preparation.ID,
		MinEstimatedTimeInSeconds: recipeStep.MinEstimatedTimeInSeconds,
		MaxEstimatedTimeInSeconds: recipeStep.MaxEstimatedTimeInSeconds,
		MinTemperatureInCelsius:   recipeStep.MinTemperatureInCelsius,
		MaxTemperatureInCelsius:   recipeStep.MaxTemperatureInCelsius,
		Notes:                     recipeStep.Notes,
		ExplicitInstructions:      recipeStep.ExplicitInstructions,
		ConditionExpression:       recipeStep.ConditionExpression,
		StartTimerAutomatically:   recipeStep.StartTimerAutomatically,
		Products:                  products,
		Ingredients:               ingredients,
		Instruments:               instruments,
		Vessels:                   vessels,
		CompletionConditions:      completionConditions,
	}
}

// ConvertRecipeCreationRequestInputToRecipeDatabaseCreationInput creates a DatabaseCreationInput from a CreationInput.
func ConvertRecipeCreationRequestInputToRecipeDatabaseCreationInput(input *mealplanning.RecipeCreationRequestInput) (*mealplanning.RecipeDatabaseCreationInput, error) {
	x := &mealplanning.RecipeDatabaseCreationInput{
		ID:                   identifiers.New(),
		AlsoCreateMeal:       input.AlsoCreateMeal,
		Name:                 input.Name,
		Slug:                 input.Slug,
		Source:               input.Source,
		SourceISBN:           input.SourceISBN,
		Description:          input.Description,
		InspiredByRecipeID:   input.InspiredByRecipeID,
		MinEstimatedPortions: input.MinEstimatedPortions,
		MaxEstimatedPortions: input.MaxEstimatedPortions,
		PortionName:          input.PortionName,
		PluralPortionName:    input.PluralPortionName,
		EligibleForMeals:     input.EligibleForMeals,
		YieldsComponentType:  input.YieldsComponentType,
	}

	for _, step := range input.Steps {
		s := ConvertRecipeStepCreationRequestInputToRecipeStepDatabaseCreationInput(step)
		s.BelongsToRecipe = x.ID
		x.Steps = append(x.Steps, s)
	}

	for _, task := range input.PrepTasks {
		prepTaskDatabaseCreationInput, err := ConvertRecipePrepTaskWithinRecipeCreationRequestInputToRecipePrepTaskDatabaseCreationInput(x, task)
		if err != nil {
			return nil, err
		}
		prepTaskDatabaseCreationInput.BelongsToRecipe = x.ID
		x.PrepTasks = append(x.PrepTasks, prepTaskDatabaseCreationInput)
	}

	return x, nil
}

// ConvertRecipeToRecipeCreationRequestInput builds a RecipeCreationRequestInput from a recipe.
func ConvertRecipeToRecipeCreationRequestInput(input *mealplanning.Recipe) *mealplanning.RecipeCreationRequestInput {
	steps := []*mealplanning.RecipeStepCreationRequestInput{}
	for _, step := range input.Steps {
		steps = append(steps, ConvertRecipeStepToRecipeStepCreationRequestInput(step))
	}

	prepTasks := []*mealplanning.RecipePrepTaskWithinRecipeCreationRequestInput{}
	for _, prepTask := range input.PrepTasks {
		prepTasks = append(prepTasks, ConvertRecipePrepTaskToRecipePrepTaskWithinRecipeCreationRequestInput(input, prepTask))
	}

	return &mealplanning.RecipeCreationRequestInput{
		Name:                 input.Name,
		Slug:                 input.Slug,
		Source:               input.Source,
		SourceISBN:           input.SourceISBN,
		Description:          input.Description,
		InspiredByRecipeID:   input.InspiredByRecipeID,
		MinEstimatedPortions: input.MinEstimatedPortions,
		MaxEstimatedPortions: input.MaxEstimatedPortions,
		PortionName:          input.PortionName,
		PluralPortionName:    input.PluralPortionName,
		Steps:                steps,
		PrepTasks:            prepTasks,
		EligibleForMeals:     input.EligibleForMeals,
		YieldsComponentType:  input.YieldsComponentType,
	}
}

// ConvertValidIngredientGroupCreationRequestInputToValidIngredientGroupDatabaseCreationInput creates a DatabaseCreationInput from a CreationInput.
func ConvertValidIngredientGroupCreationRequestInputToValidIngredientGroupDatabaseCreationInput(input *mealplanning.ValidIngredientGroupCreationRequestInput) *mealplanning.ValidIngredientGroupDatabaseCreationInput {
	var members []*mealplanning.ValidIngredientGroupMemberDatabaseCreationInput
	for _, member := range input.Members {
		members = append(members, &mealplanning.ValidIngredientGroupMemberDatabaseCreationInput{
			ID:                identifiers.New(),
			ValidIngredientID: member.ValidIngredientID,
		})
	}

	x := &mealplanning.ValidIngredientGroupDatabaseCreationInput{
		ID:          identifiers.New(),
		Name:        input.Name,
		Description: input.Description,
		Slug:        input.Slug,
		Members:     members,
	}

	return x
}

// ConvertValidIngredientGroupToValidIngredientGroupDatabaseCreationInput builds a ValidIngredientGroupDatabaseCreationInput from a ValidIngredientGroup.
func ConvertValidIngredientGroupToValidIngredientGroupDatabaseCreationInput(validIngredient *mealplanning.ValidIngredientGroup) *mealplanning.ValidIngredientGroupDatabaseCreationInput {
	members := make([]*mealplanning.ValidIngredientGroupMemberDatabaseCreationInput, len(validIngredient.Members))
	for i, member := range validIngredient.Members {
		members[i] = &mealplanning.ValidIngredientGroupMemberDatabaseCreationInput{
			ID:                member.ID,
			ValidIngredientID: member.ValidIngredient.ID,
		}
	}

	return &mealplanning.ValidIngredientGroupDatabaseCreationInput{
		ID:          validIngredient.ID,
		Name:        validIngredient.Name,
		Description: validIngredient.Description,
		Slug:        validIngredient.Slug,
		Members:     members,
	}
}

// ConvertNullableValidInstrumentToValidInstrument produces a ValidInstrument from a NullableValidInstrument.
func ConvertNullableValidInstrumentToValidInstrument(x *mealplanning.NullableValidInstrument) *mealplanning.ValidInstrument {
	return &mealplanning.ValidInstrument{
		LastUpdatedAt:                  x.LastUpdatedAt,
		ArchivedAt:                     x.ArchivedAt,
		Description:                    *x.Description,
		IconPath:                       *x.IconPath,
		ID:                             *x.ID,
		Name:                           *x.Name,
		PluralName:                     *x.PluralName,
		CreatedAt:                      *x.CreatedAt,
		UsableForStorage:               *x.UsableForStorage,
		Slug:                           *x.Slug,
		DisplayInSummaryLists:          *x.DisplayInSummaryLists,
		IncludeInGeneratedInstructions: *x.IncludeInGeneratedInstructions,
	}
}

// ConvertNullableValidMeasurementUnitToValidMeasurementUnit produces a ValidMeasurementUnit from a NullableValidMeasurementUnit.
func ConvertNullableValidMeasurementUnitToValidMeasurementUnit(x *mealplanning.NullableValidMeasurementUnit) *mealplanning.ValidMeasurementUnit {
	if x != nil && x.ID != nil {
		return &mealplanning.ValidMeasurementUnit{
			CreatedAt:     *x.CreatedAt,
			LastUpdatedAt: x.LastUpdatedAt,
			ArchivedAt:    x.ArchivedAt,
			Name:          *x.Name,
			IconPath:      *x.IconPath,
			ID:            *x.ID,
			Description:   *x.Description,
			PluralName:    *x.PluralName,
			Slug:          *x.Slug,
			Volumetric:    *x.Volumetric,
			Universal:     *x.Universal,
			Metric:        *x.Metric,
			Imperial:      *x.Imperial,
		}
	}
	return nil
}

// ConvertValidMeasurementUnitToNullableValidMeasurementUnit converts a NullableValidMeasurementUnit to a ValidMeasurementUnit.
func ConvertValidMeasurementUnitToNullableValidMeasurementUnit(input *mealplanning.ValidMeasurementUnit) *mealplanning.NullableValidMeasurementUnit {
	return &mealplanning.NullableValidMeasurementUnit{
		CreatedAt:     &input.CreatedAt,
		LastUpdatedAt: input.LastUpdatedAt,
		ArchivedAt:    input.ArchivedAt,
		Name:          &input.Name,
		IconPath:      &input.IconPath,
		ID:            &input.ID,
		Description:   &input.Description,
		PluralName:    &input.PluralName,
		Slug:          &input.Slug,
		Volumetric:    &input.Volumetric,
		Universal:     &input.Universal,
		Metric:        &input.Metric,
		Imperial:      &input.Imperial,
	}
}

// ConvertNullableValidVesselToValidVessel produces a ValidVessel from a NullableValidVessel.
func ConvertNullableValidVesselToValidVessel(x *mealplanning.NullableValidVessel) *mealplanning.ValidVessel {
	v := &mealplanning.ValidVessel{
		CapacityUnit:  ConvertNullableValidMeasurementUnitToValidMeasurementUnit(x.CapacityUnit),
		LastUpdatedAt: x.LastUpdatedAt,
		ArchivedAt:    x.ArchivedAt,
	}

	if x.ID != nil {
		v.ID = *x.ID
	}
	if x.Name != nil {
		v.Name = *x.Name
	}
	if x.PluralName != nil {
		v.PluralName = *x.PluralName
	}
	if x.Description != nil {
		v.Description = *x.Description
	}
	if x.IconPath != nil {
		v.IconPath = *x.IconPath
	}
	if x.UsableForStorage != nil {
		v.UsableForStorage = *x.UsableForStorage
	}
	if x.Slug != nil {
		v.Slug = *x.Slug
	}
	if x.DisplayInSummaryLists != nil {
		v.DisplayInSummaryLists = *x.DisplayInSummaryLists
	}
	if x.IncludeInGeneratedInstructions != nil {
		v.IncludeInGeneratedInstructions = *x.IncludeInGeneratedInstructions
	}
	if x.Capacity != nil {
		v.Capacity = *x.Capacity
	}
	if x.WidthInMillimeters != nil {
		v.WidthInMillimeters = *x.WidthInMillimeters
	}
	if x.LengthInMillimeters != nil {
		v.LengthInMillimeters = *x.LengthInMillimeters
	}
	if x.HeightInMillimeters != nil {
		v.HeightInMillimeters = *x.HeightInMillimeters
	}
	if x.Shape != nil {
		v.Shape = *x.Shape
	}
	if x.CreatedAt != nil {
		v.CreatedAt = *x.CreatedAt
	}

	return v
}
