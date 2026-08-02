package converters

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v9/identifiers"
)

// ConvertValidIngredientPreparationCreationRequestInputToValidIngredientPreparationDatabaseCreationInput creates a ValidIngredientPreparationDatabaseCreationInput from a ValidIngredientPreparationCreationRequestInput.
func ConvertValidIngredientPreparationCreationRequestInputToValidIngredientPreparationDatabaseCreationInput(input *types.ValidIngredientPreparationCreationRequestInput) *types.ValidIngredientPreparationDatabaseCreationInput {
	x := &types.ValidIngredientPreparationDatabaseCreationInput{
		ID:                 identifiers.New(),
		Notes:              input.Notes,
		ValidPreparationID: input.ValidPreparationID,
		ValidIngredientID:  input.ValidIngredientID,
	}

	return x
}

// ConvertValidIngredientPreparationToValidIngredientPreparationUpdateRequestInput builds a ValidIngredientPreparationUpdateRequestInput from a ValidIngredientPreparation.
func ConvertValidIngredientPreparationToValidIngredientPreparationUpdateRequestInput(validIngredientPreparation *types.ValidIngredientPreparation) *types.ValidIngredientPreparationUpdateRequestInput {
	return &types.ValidIngredientPreparationUpdateRequestInput{
		Notes:              &validIngredientPreparation.Notes,
		ValidPreparationID: &validIngredientPreparation.Preparation.ID,
		ValidIngredientID:  &validIngredientPreparation.Ingredient.ID,
	}
}

// ConvertValidIngredientPreparationToValidIngredientPreparationCreationRequestInput builds a ValidIngredientPreparationCreationRequestInput from a ValidIngredientPreparation.
func ConvertValidIngredientPreparationToValidIngredientPreparationCreationRequestInput(validIngredientPreparation *types.ValidIngredientPreparation) *types.ValidIngredientPreparationCreationRequestInput {
	return &types.ValidIngredientPreparationCreationRequestInput{
		Notes:              validIngredientPreparation.Notes,
		ValidPreparationID: validIngredientPreparation.Preparation.ID,
		ValidIngredientID:  validIngredientPreparation.Ingredient.ID,
	}
}

// ConvertValidIngredientPreparationToValidIngredientPreparationDatabaseCreationInput builds a ValidIngredientPreparationDatabaseCreationInput from a ValidIngredientPreparation.
func ConvertValidIngredientPreparationToValidIngredientPreparationDatabaseCreationInput(validIngredientPreparation *types.ValidIngredientPreparation) *types.ValidIngredientPreparationDatabaseCreationInput {
	return &types.ValidIngredientPreparationDatabaseCreationInput{
		ID:                 validIngredientPreparation.ID,
		Notes:              validIngredientPreparation.Notes,
		ValidPreparationID: validIngredientPreparation.Preparation.ID,
		ValidIngredientID:  validIngredientPreparation.Ingredient.ID,
	}
}
