package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
)

// BuildFakeValidIngredient builds a faked valid ingredient.
func BuildFakeValidIngredient() *types.ValidIngredient {
	validIngredient := fake.BuildFakeRecord[types.ValidIngredient]()

	validIngredient.MinStorageTemperatureInCelsius, validIngredient.MaxStorageTemperatureInCelsius = BuildFakeOptionalFloat32MinMax()

	return validIngredient
}

// BuildFakeValidIngredientsList builds a faked ValidIngredientList.
func BuildFakeValidIngredientsList() *filtering.QueryFilteredResult[types.ValidIngredient] {
	return fake.BuildFakePage(BuildFakeValidIngredient)
}

// BuildFakeValidIngredientUpdateRequestInput builds a faked ValidIngredientUpdateRequestInput from a valid ingredient.
func BuildFakeValidIngredientUpdateRequestInput() *types.ValidIngredientUpdateRequestInput {
	validIngredient := BuildFakeValidIngredient()

	return converters.ConvertValidIngredientToValidIngredientUpdateRequestInput(validIngredient)
}

// BuildFakeValidIngredientCreationRequestInput builds a faked ValidIngredientCreationRequestInput.
func BuildFakeValidIngredientCreationRequestInput() *types.ValidIngredientCreationRequestInput {
	validIngredient := BuildFakeValidIngredient()

	return converters.ConvertValidIngredientToValidIngredientCreationRequestInput(validIngredient)
}
