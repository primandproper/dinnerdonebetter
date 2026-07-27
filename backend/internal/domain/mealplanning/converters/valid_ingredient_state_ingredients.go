package converters

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v7/identifiers"
)

// ConvertValidIngredientStateIngredientCreationRequestInputToValidIngredientStateIngredientDatabaseCreationInput creates a ValidIngredientStateIngredientDatabaseCreationInput from a ValidIngredientStateIngredientCreationRequestInput.
func ConvertValidIngredientStateIngredientCreationRequestInputToValidIngredientStateIngredientDatabaseCreationInput(input *types.ValidIngredientStateIngredientCreationRequestInput) *types.ValidIngredientStateIngredientDatabaseCreationInput {
	x := &types.ValidIngredientStateIngredientDatabaseCreationInput{
		ID:                     identifiers.New(),
		Notes:                  input.Notes,
		ValidIngredientStateID: input.ValidIngredientStateID,
		ValidIngredientID:      input.ValidIngredientID,
	}

	return x
}

// ConvertValidIngredientStateIngredientToValidIngredientStateIngredientUpdateRequestInput builds a ValidIngredientStateIngredientUpdateRequestInput from a ValidIngredientStateIngredient.
func ConvertValidIngredientStateIngredientToValidIngredientStateIngredientUpdateRequestInput(validIngredientStateIngredient *types.ValidIngredientStateIngredient) *types.ValidIngredientStateIngredientUpdateRequestInput {
	return &types.ValidIngredientStateIngredientUpdateRequestInput{
		Notes:                  &validIngredientStateIngredient.Notes,
		ValidIngredientStateID: &validIngredientStateIngredient.IngredientState.ID,
		ValidIngredientID:      &validIngredientStateIngredient.Ingredient.ID,
	}
}

// ConvertValidIngredientStateIngredientToValidIngredientStateIngredientCreationRequestInput builds a ValidIngredientStateIngredientCreationRequestInput from a ValidIngredientStateIngredient.
func ConvertValidIngredientStateIngredientToValidIngredientStateIngredientCreationRequestInput(validIngredientStateIngredient *types.ValidIngredientStateIngredient) *types.ValidIngredientStateIngredientCreationRequestInput {
	return &types.ValidIngredientStateIngredientCreationRequestInput{
		Notes:                  validIngredientStateIngredient.Notes,
		ValidIngredientStateID: validIngredientStateIngredient.IngredientState.ID,
		ValidIngredientID:      validIngredientStateIngredient.Ingredient.ID,
	}
}

// ConvertValidIngredientStateIngredientToValidIngredientStateIngredientDatabaseCreationInput builds a ValidIngredientStateIngredientDatabaseCreationInput from a ValidIngredientStateIngredient.
func ConvertValidIngredientStateIngredientToValidIngredientStateIngredientDatabaseCreationInput(validIngredientStateIngredient *types.ValidIngredientStateIngredient) *types.ValidIngredientStateIngredientDatabaseCreationInput {
	return &types.ValidIngredientStateIngredientDatabaseCreationInput{
		ID:                     validIngredientStateIngredient.ID,
		Notes:                  validIngredientStateIngredient.Notes,
		ValidIngredientStateID: validIngredientStateIngredient.IngredientState.ID,
		ValidIngredientID:      validIngredientStateIngredient.Ingredient.ID,
	}
}
