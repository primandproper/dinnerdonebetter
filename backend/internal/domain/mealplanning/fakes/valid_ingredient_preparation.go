package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"
)

// BuildFakeValidIngredientPreparation builds a faked valid ingredient preparation.
func BuildFakeValidIngredientPreparation() *types.ValidIngredientPreparation {
	validIngredientPreparation := fake.BuildFakeRecord[types.ValidIngredientPreparation]()

	// The two records the row joins, each built by its own builder: this is a bridge
	// table, and a bridge to two records nothing else in the test has is a row every
	// join drops.
	validIngredientPreparation.Preparation = *BuildFakeValidPreparation()
	validIngredientPreparation.Ingredient = *BuildFakeValidIngredient()

	return validIngredientPreparation
}

// BuildFakeValidIngredientPreparationsList builds a faked ValidIngredientPreparationList.
func BuildFakeValidIngredientPreparationsList() *filtering.QueryFilteredResult[types.ValidIngredientPreparation] {
	return fake.BuildFakePage(BuildFakeValidIngredientPreparation)
}

// BuildFakeValidIngredientPreparationUpdateRequestInput builds a faked ValidIngredientPreparationUpdateRequestInput from a valid ingredient preparation.
func BuildFakeValidIngredientPreparationUpdateRequestInput() *types.ValidIngredientPreparationUpdateRequestInput {
	validIngredientPreparation := BuildFakeValidIngredientPreparation()

	return &types.ValidIngredientPreparationUpdateRequestInput{
		Notes:              &validIngredientPreparation.Notes,
		ValidPreparationID: &validIngredientPreparation.Preparation.ID,
		ValidIngredientID:  &validIngredientPreparation.Ingredient.ID,
	}
}

// BuildFakeValidIngredientPreparationCreationRequestInput builds a faked ValidIngredientPreparationCreationRequestInput.
func BuildFakeValidIngredientPreparationCreationRequestInput() *types.ValidIngredientPreparationCreationRequestInput {
	validIngredientPreparation := BuildFakeValidIngredientPreparation()

	return converters.ConvertValidIngredientPreparationToValidIngredientPreparationCreationRequestInput(validIngredientPreparation)
}
