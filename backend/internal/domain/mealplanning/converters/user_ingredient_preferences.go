package converters

import types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

// ConvertUserIngredientPreferenceToUserIngredientPreferenceUpdateRequestInput creates a UserIngredientPreferenceUpdateRequestInput from a UserIngredientPreferences.
func ConvertUserIngredientPreferenceToUserIngredientPreferenceUpdateRequestInput(input *types.UserIngredientPreference) *types.UserIngredientPreferenceUpdateRequestInput {
	x := &types.UserIngredientPreferenceUpdateRequestInput{
		Notes:        &input.Notes,
		IngredientID: &input.Ingredient.ID,
		Rating:       &input.Rating,
		Allergy:      &input.Allergy,
	}

	return x
}

// ConvertUserIngredientPreferenceCreationRequestInputToUserIngredientPreferenceDatabaseCreationInput creates a UserIngredientPreferenceDatabaseCreationInput from a UserIngredientPreferenceCreationRequestInput.
func ConvertUserIngredientPreferenceCreationRequestInputToUserIngredientPreferenceDatabaseCreationInput(input *types.UserIngredientPreferenceCreationRequestInput) *types.UserIngredientPreferenceDatabaseCreationInput {
	x := &types.UserIngredientPreferenceDatabaseCreationInput{
		ValidIngredientGroupID: input.ValidIngredientGroupID,
		ValidIngredientID:      input.ValidIngredientID,
		Rating:                 input.Rating,
		Notes:                  input.Notes,
		Allergy:                input.Allergy,
	}

	return x
}

// ConvertUserIngredientPreferenceToUserIngredientPreferenceCreationRequestInput builds a UserIngredientPreferenceCreationRequestInput from a UserIngredientPreferences.
func ConvertUserIngredientPreferenceToUserIngredientPreferenceCreationRequestInput(x *types.UserIngredientPreference) *types.UserIngredientPreferenceCreationRequestInput {
	return &types.UserIngredientPreferenceCreationRequestInput{
		ValidIngredientID: x.Ingredient.ID,
		Rating:            x.Rating,
		Notes:             x.Notes,
		Allergy:           x.Allergy,
	}
}

// ConvertUserIngredientPreferenceToUserIngredientPreferenceDatabaseCreationInput builds a UserIngredientPreferenceDatabaseCreationInput from a UserIngredientPreferences.
func ConvertUserIngredientPreferenceToUserIngredientPreferenceDatabaseCreationInput(x *types.UserIngredientPreference) *types.UserIngredientPreferenceDatabaseCreationInput {
	return &types.UserIngredientPreferenceDatabaseCreationInput{
		ValidIngredientID: x.Ingredient.ID,
		Rating:            x.Rating,
		Notes:             x.Notes,
		Allergy:           x.Allergy,
		CreatedByUser:     x.CreatedByUser,
	}
}
