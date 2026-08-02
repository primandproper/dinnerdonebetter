package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v9/identifiers"
)

// ConvertRecipeStepToRecipeStepUpdateRequestInput creates a RecipeStepUpdateRequestInput from a RecipeStep.
func ConvertRecipeStepToRecipeStepUpdateRequestInput(input *mealplanning.RecipeStep) *mealplanning.RecipeStepUpdateRequestInput {
	x := &mealplanning.RecipeStepUpdateRequestInput{
		Notes:                     &input.Notes,
		BelongsToRecipe:           &input.BelongsToRecipe,
		Preparation:               &input.Preparation,
		Index:                     &input.Index,
		MinEstimatedTimeInSeconds: input.MinEstimatedTimeInSeconds,
		MaxEstimatedTimeInSeconds: input.MaxEstimatedTimeInSeconds,
		MinTemperatureInCelsius:   input.MinTemperatureInCelsius,
		MaxTemperatureInCelsius:   input.MaxTemperatureInCelsius,
		Optional:                  &input.Optional,
		ExplicitInstructions:      &input.ExplicitInstructions,
		ConditionExpression:       &input.ConditionExpression,
		StartTimerAutomatically:   &input.StartTimerAutomatically,
	}

	return x
}

// ConvertRecipeStepCreationInputToRecipeStepDatabaseCreationInput creates a RecipeStepDatabaseCreationInput from a RecipeStepCreationRequestInput.
func ConvertRecipeStepCreationInputToRecipeStepDatabaseCreationInput(input *mealplanning.RecipeStepCreationRequestInput) *mealplanning.RecipeStepDatabaseCreationInput {
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
		convertedProduct := ConvertRecipeStepProductCreationInputToRecipeStepProductDatabaseCreationInput(product)
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

// ConvertRecipeStepToRecipeStepDatabaseCreationInput builds a RecipeStepDatabaseCreationInput from a RecipeStep.
func ConvertRecipeStepToRecipeStepDatabaseCreationInput(recipeStep *mealplanning.RecipeStep) *mealplanning.RecipeStepDatabaseCreationInput {
	ingredients := []*mealplanning.RecipeStepIngredientDatabaseCreationInput{}
	for _, i := range recipeStep.Ingredients {
		ingredients = append(ingredients, ConvertRecipeStepIngredientToRecipeStepIngredientDatabaseCreationInput(i))
	}

	instruments := []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{}
	for _, i := range recipeStep.Instruments {
		instruments = append(instruments, ConvertRecipeStepInstrumentToRecipeStepInstrumentDatabaseCreationInput(i))
	}

	vessels := []*mealplanning.RecipeStepVesselDatabaseCreationInput{}
	for _, v := range recipeStep.Vessels {
		vessels = append(vessels, ConvertRecipeStepVesselToRecipeStepVesselDatabaseCreationInput(v))
	}

	products := []*mealplanning.RecipeStepProductDatabaseCreationInput{}
	for _, p := range recipeStep.Products {
		products = append(products, ConvertRecipeStepProductToRecipeStepProductDatabaseCreationInput(p))
	}

	completionConditions := []*mealplanning.RecipeStepCompletionConditionDatabaseCreationInput{}
	for _, completionCondition := range recipeStep.CompletionConditions {
		completionConditions = append(completionConditions, ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionDatabaseCreationInput(completionCondition))
	}

	return &mealplanning.RecipeStepDatabaseCreationInput{
		ID:                        recipeStep.ID,
		Index:                     recipeStep.Index,
		PreparationID:             recipeStep.Preparation.ID,
		Optional:                  recipeStep.Optional,
		MinEstimatedTimeInSeconds: recipeStep.MinEstimatedTimeInSeconds,
		MaxEstimatedTimeInSeconds: recipeStep.MaxEstimatedTimeInSeconds,
		MinTemperatureInCelsius:   recipeStep.MinTemperatureInCelsius,
		MaxTemperatureInCelsius:   recipeStep.MaxTemperatureInCelsius,
		StartTimerAutomatically:   recipeStep.StartTimerAutomatically,
		Notes:                     recipeStep.Notes,
		ExplicitInstructions:      recipeStep.ExplicitInstructions,
		ConditionExpression:       recipeStep.ConditionExpression,
		Ingredients:               ingredients,
		Instruments:               instruments,
		Products:                  products,
		Vessels:                   vessels,
		BelongsToRecipe:           recipeStep.BelongsToRecipe,
		CompletionConditions:      completionConditions,
	}
}
