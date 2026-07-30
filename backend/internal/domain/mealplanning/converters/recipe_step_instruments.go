package converters

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v8/identifiers"
)

// ConvertRecipeStepInstrumentToRecipeStepInstrumentUpdateRequestInput creates a RecipeStepInstrumentUpdateRequestInput from a RecipeStepInstrument.
func ConvertRecipeStepInstrumentToRecipeStepInstrumentUpdateRequestInput(input *mealplanning.RecipeStepInstrument) *mealplanning.RecipeStepInstrumentUpdateRequestInput {
	x := &mealplanning.RecipeStepInstrumentUpdateRequestInput{
		InstrumentID:        &input.Instrument.ID,
		Notes:               &input.Notes,
		RecipeStepProductID: input.RecipeStepProductID,
		Name:                &input.Name,
		PreferenceRank:      &input.PreferenceRank,
		BelongsToRecipeStep: &input.BelongsToRecipeStep,
		Optional:            &input.Optional,
		Index:               &input.Index,
		OptionIndex:         &input.OptionIndex,
		MinQuantity:         &input.MinQuantity,
		MaxQuantity:         input.MaxQuantity,
		ScaleFactor:         &input.ScaleFactor,
	}

	return x
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

// ConvertRecipeStepInstrumentToRecipeStepInstrumentDatabaseCreationInput builds a RecipeStepInstrumentDatabaseCreationInput from a RecipeStepInstrument.
func ConvertRecipeStepInstrumentToRecipeStepInstrumentDatabaseCreationInput(input *mealplanning.RecipeStepInstrument) *mealplanning.RecipeStepInstrumentDatabaseCreationInput {
	var instrumentID *string
	if input.Instrument != nil {
		instrumentID = &input.Instrument.ID
	}

	return &mealplanning.RecipeStepInstrumentDatabaseCreationInput{
		ID:                  input.ID,
		InstrumentID:        instrumentID,
		Name:                input.Name,
		RecipeStepProductID: input.RecipeStepProductID,
		Notes:               input.Notes,
		PreferenceRank:      input.PreferenceRank,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Optional:            input.Optional,
		Index:               input.Index,
		OptionIndex:         input.OptionIndex,
		MinQuantity:         input.MinQuantity,
		MaxQuantity:         input.MaxQuantity,
		ScaleFactor:         scaleFactorOrDefault(input.ScaleFactor),
	}
}
