package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v9/identifiers"
)

// ConvertRecipeStepVesselToRecipeStepVesselUpdateRequestInput creates a RecipeStepVesselUpdateRequestInput from a RecipeStepVessel.
func ConvertRecipeStepVesselToRecipeStepVesselUpdateRequestInput(input *mealplanning.RecipeStepVessel) *mealplanning.RecipeStepVesselUpdateRequestInput {
	x := &mealplanning.RecipeStepVesselUpdateRequestInput{
		VesselID:             &input.Vessel.ID,
		Notes:                &input.Notes,
		RecipeStepProductID:  input.RecipeStepProductID,
		Name:                 &input.Name,
		BelongsToRecipeStep:  &input.BelongsToRecipeStep,
		MinQuantity:          &input.MinQuantity,
		MaxQuantity:          input.MaxQuantity,
		Index:                &input.Index,
		OptionIndex:          &input.OptionIndex,
		VesselPreposition:    &input.VesselPreposition,
		UnavailableAfterStep: &input.UnavailableAfterStep,
		ScaleFactor:          &input.ScaleFactor,
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

// ConvertRecipeStepVesselToRecipeStepVesselDatabaseCreationInput builds a RecipeStepVesselDatabaseCreationInput from a RecipeStepVessel.
func ConvertRecipeStepVesselToRecipeStepVesselDatabaseCreationInput(input *mealplanning.RecipeStepVessel) *mealplanning.RecipeStepVesselDatabaseCreationInput {
	var vesselID *string
	if input.Vessel != nil {
		vesselID = &input.Vessel.ID
	}

	return &mealplanning.RecipeStepVesselDatabaseCreationInput{
		ID:                   input.ID,
		VesselID:             vesselID,
		Name:                 input.Name,
		RecipeStepProductID:  input.RecipeStepProductID,
		Notes:                input.Notes,
		BelongsToRecipeStep:  input.BelongsToRecipeStep,
		VesselPreposition:    input.VesselPreposition,
		UnavailableAfterStep: input.UnavailableAfterStep,
		Index:                input.Index,
		OptionIndex:          input.OptionIndex,
		MinQuantity:          input.MinQuantity,
		MaxQuantity:          input.MaxQuantity,
		ScaleFactor:          scaleFactorOrDefault(input.ScaleFactor),
	}
}
