package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v9/identifiers"
)

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

// ConvertRecipeStepIngredientToRecipeStepIngredientUpdateRequestInput creates a RecipeStepIngredientUpdateRequestInput from a RecipeStepIngredient.
func ConvertRecipeStepIngredientToRecipeStepIngredientUpdateRequestInput(input *mealplanning.RecipeStepIngredient) *mealplanning.RecipeStepIngredientUpdateRequestInput {
	x := &mealplanning.RecipeStepIngredientUpdateRequestInput{
		IngredientID:           &input.Ingredient.ID,
		RecipeStepProductID:    input.RecipeStepProductID,
		Name:                   &input.Name,
		MeasurementUnitID:      &input.MeasurementUnit.ID,
		QuantityNotes:          &input.QuantityNotes,
		IngredientNotes:        &input.IngredientNotes,
		BelongsToRecipeStep:    &input.BelongsToRecipeStep,
		MinQuantity:            &input.MinQuantity,
		MaxQuantity:            input.MaxQuantity,
		Optional:               &input.Optional,
		Index:                  &input.Index,
		OptionIndex:            &input.OptionIndex,
		VesselIndex:            input.VesselIndex,
		ToTaste:                &input.ToTaste,
		ProductPercentageToUse: input.ProductPercentageToUse,
		ScaleFactor:            &input.ScaleFactor,
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

// ConvertRecipeStepIngredientToRecipeStepIngredientDatabaseCreationInput builds a RecipeStepIngredientDatabaseCreationInput from a RecipeStepIngredient.
func ConvertRecipeStepIngredientToRecipeStepIngredientDatabaseCreationInput(input *mealplanning.RecipeStepIngredient) *mealplanning.RecipeStepIngredientDatabaseCreationInput {
	var ingredientID *string
	if input.Ingredient != nil {
		ingredientID = &input.Ingredient.ID
	}

	return &mealplanning.RecipeStepIngredientDatabaseCreationInput{
		ID:                        input.ID,
		Name:                      input.Name,
		Optional:                  input.Optional,
		IngredientID:              ingredientID,
		MeasurementUnitID:         input.MeasurementUnit.ID,
		MinQuantity:               input.MinQuantity,
		MaxQuantity:               input.MaxQuantity,
		QuantityNotes:             input.QuantityNotes,
		IngredientNotes:           input.IngredientNotes,
		BelongsToRecipeStep:       input.BelongsToRecipeStep,
		RecipeStepProductRecipeID: input.RecipeStepProductRecipeID,
		RecipeStepProductID:       input.RecipeStepProductID,
		Index:                     input.Index,
		OptionIndex:               input.OptionIndex,
		VesselIndex:               input.VesselIndex,
		ToTaste:                   input.ToTaste,
		ProductPercentageToUse:    input.ProductPercentageToUse,
		ScaleFactor:               scaleFactorOrDefault(input.ScaleFactor),
	}
}
