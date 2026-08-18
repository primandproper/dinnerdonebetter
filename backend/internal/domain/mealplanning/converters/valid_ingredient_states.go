package converters

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v11/identifiers"
)

// ConvertValidIngredientStateToValidIngredientStateUpdateRequestInput creates a ValidIngredientStateUpdateRequestInput from a ValidIngredientState.
func ConvertValidIngredientStateToValidIngredientStateUpdateRequestInput(input *types.ValidIngredientState) *types.ValidIngredientStateUpdateRequestInput {
	x := &types.ValidIngredientStateUpdateRequestInput{
		Name:          &input.Name,
		Description:   &input.Description,
		IconPath:      &input.IconPath,
		Slug:          &input.Slug,
		PastTense:     &input.PastTense,
		AttributeType: &input.AttributeType,
	}

	return x
}

// ConvertValidIngredientStateCreationRequestInputToValidIngredientStateDatabaseCreationInput creates a ValidIngredientStateDatabaseCreationInput from a ValidIngredientStateCreationRequestInput.
func ConvertValidIngredientStateCreationRequestInputToValidIngredientStateDatabaseCreationInput(input *types.ValidIngredientStateCreationRequestInput) *types.ValidIngredientStateDatabaseCreationInput {
	x := &types.ValidIngredientStateDatabaseCreationInput{
		ID:            identifiers.New(),
		Name:          input.Name,
		Description:   input.Description,
		IconPath:      input.IconPath,
		Slug:          input.Slug,
		PastTense:     input.PastTense,
		AttributeType: input.AttributeType,
	}

	return x
}

// ConvertValidIngredientStateToValidIngredientStateCreationRequestInput builds a ValidIngredientStateCreationRequestInput from a ValidIngredientState.
func ConvertValidIngredientStateToValidIngredientStateCreationRequestInput(validIngredientState *types.ValidIngredientState) *types.ValidIngredientStateCreationRequestInput {
	return &types.ValidIngredientStateCreationRequestInput{
		Name:          validIngredientState.Name,
		Description:   validIngredientState.Description,
		IconPath:      validIngredientState.IconPath,
		Slug:          validIngredientState.Slug,
		PastTense:     validIngredientState.PastTense,
		AttributeType: validIngredientState.AttributeType,
	}
}

// ConvertValidIngredientStateToValidIngredientStateDatabaseCreationInput builds a ValidIngredientStateDatabaseCreationInput from a ValidIngredientState.
func ConvertValidIngredientStateToValidIngredientStateDatabaseCreationInput(validIngredientState *types.ValidIngredientState) *types.ValidIngredientStateDatabaseCreationInput {
	return &types.ValidIngredientStateDatabaseCreationInput{
		ID:            validIngredientState.ID,
		Name:          validIngredientState.Name,
		Description:   validIngredientState.Description,
		IconPath:      validIngredientState.IconPath,
		Slug:          validIngredientState.Slug,
		PastTense:     validIngredientState.PastTense,
		AttributeType: validIngredientState.AttributeType,
	}
}
