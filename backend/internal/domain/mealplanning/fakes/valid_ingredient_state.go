package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"
)

// BuildFakeValidIngredientState builds a faked valid ingredient state.
func BuildFakeValidIngredientState() *types.ValidIngredientState {
	validIngredientState := fake.BuildFakeRecord[types.ValidIngredientState]()

	// One of the attribute types the domain enumerates and validates against.
	validIngredientState.AttributeType = types.ValidIngredientStateAttributeTypeOther

	return validIngredientState
}

// BuildFakeValidIngredientStatesList builds a faked ValidIngredientStateList.
func BuildFakeValidIngredientStatesList() *filtering.QueryFilteredResult[types.ValidIngredientState] {
	return fake.BuildFakePage(BuildFakeValidIngredientState)
}

// BuildFakeValidIngredientStateUpdateRequestInput builds a faked ValidIngredientStateUpdateRequestInput from a valid ingredient state.
func BuildFakeValidIngredientStateUpdateRequestInput() *types.ValidIngredientStateUpdateRequestInput {
	validIngredientState := BuildFakeValidIngredientState()

	return converters.ConvertValidIngredientStateToValidIngredientStateUpdateRequestInput(validIngredientState)
}

// BuildFakeValidIngredientStateCreationRequestInput builds a faked ValidIngredientStateCreationRequestInput.
func BuildFakeValidIngredientStateCreationRequestInput() *types.ValidIngredientStateCreationRequestInput {
	validIngredientState := BuildFakeValidIngredientState()

	return converters.ConvertValidIngredientStateToValidIngredientStateCreationRequestInput(validIngredientState)
}
