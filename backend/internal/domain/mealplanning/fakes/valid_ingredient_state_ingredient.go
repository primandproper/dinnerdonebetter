package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"
)

// BuildFakeValidIngredientStateIngredient builds a faked valid ingredient state ingredient.
func BuildFakeValidIngredientStateIngredient() *types.ValidIngredientStateIngredient {
	validIngredientStateIngredient := fake.BuildFakeRecord[types.ValidIngredientStateIngredient]()

	validIngredientStateIngredient.IngredientState = *BuildFakeValidIngredientState()
	validIngredientStateIngredient.Ingredient = *BuildFakeValidIngredient()

	return validIngredientStateIngredient
}

// BuildFakeValidIngredientStateIngredientsList builds a faked ValidIngredientStateIngredientList.
func BuildFakeValidIngredientStateIngredientsList() *filtering.QueryFilteredResult[types.ValidIngredientStateIngredient] {
	return fake.BuildFakePage(BuildFakeValidIngredientStateIngredient)
}

// BuildFakeValidIngredientStateIngredientUpdateRequestInput builds a faked ValidIngredientStateIngredientUpdateRequestInput from a valid ingredient state ingredient.
func BuildFakeValidIngredientStateIngredientUpdateRequestInput() *types.ValidIngredientStateIngredientUpdateRequestInput {
	validIngredientStateIngredient := BuildFakeValidIngredientStateIngredient()

	return &types.ValidIngredientStateIngredientUpdateRequestInput{
		Notes:                  &validIngredientStateIngredient.Notes,
		ValidIngredientStateID: &validIngredientStateIngredient.IngredientState.ID,
		ValidIngredientID:      &validIngredientStateIngredient.Ingredient.ID,
	}
}

// BuildFakeValidIngredientStateIngredientCreationRequestInput builds a faked ValidIngredientStateIngredientCreationRequestInput.
func BuildFakeValidIngredientStateIngredientCreationRequestInput() *types.ValidIngredientStateIngredientCreationRequestInput {
	validIngredientStateIngredient := BuildFakeValidIngredientStateIngredient()

	return converters.ConvertValidIngredientStateIngredientToValidIngredientStateIngredientCreationRequestInput(validIngredientStateIngredient)
}
