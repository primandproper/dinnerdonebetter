package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
)

// BuildFakeValidIngredientMeasurementUnit builds a faked valid ingredient measurement unit.
func BuildFakeValidIngredientMeasurementUnit() *mealplanning.ValidIngredientMeasurementUnit {
	validIngredientMeasurementUnit := fake.BuildFakeRecord[mealplanning.ValidIngredientMeasurementUnit]()

	validIngredientMeasurementUnit.MeasurementUnit = *BuildFakeValidMeasurementUnit()
	validIngredientMeasurementUnit.Ingredient = *BuildFakeValidIngredient()
	validIngredientMeasurementUnit.MinAllowableQuantity, validIngredientMeasurementUnit.MaxAllowableQuantity = BuildFakeFloat32WithOptionalMax()

	return validIngredientMeasurementUnit
}

// BuildFakeValidIngredientMeasurementUnitsList builds a faked ValidIngredientMeasurementUnitList.
func BuildFakeValidIngredientMeasurementUnitsList() *filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit] {
	return fake.BuildFakePage(BuildFakeValidIngredientMeasurementUnit)
}

// BuildFakeValidIngredientMeasurementUnitUpdateRequestInput builds a faked ValidIngredientMeasurementUnitUpdateRequestInput from a valid ingredient measurement unit.
func BuildFakeValidIngredientMeasurementUnitUpdateRequestInput() *mealplanning.ValidIngredientMeasurementUnitUpdateRequestInput {
	validIngredientMeasurementUnit := BuildFakeValidIngredientMeasurementUnit()

	return &mealplanning.ValidIngredientMeasurementUnitUpdateRequestInput{
		Notes:                  &validIngredientMeasurementUnit.Notes,
		ValidMeasurementUnitID: &validIngredientMeasurementUnit.MeasurementUnit.ID,
		ValidIngredientID:      &validIngredientMeasurementUnit.Ingredient.ID,
		MinAllowableQuantity:   &validIngredientMeasurementUnit.MinAllowableQuantity,
		MaxAllowableQuantity:   validIngredientMeasurementUnit.MaxAllowableQuantity,
	}
}

// BuildFakeValidIngredientMeasurementUnitCreationRequestInput builds a faked ValidIngredientMeasurementUnitCreationRequestInput.
func BuildFakeValidIngredientMeasurementUnitCreationRequestInput() *mealplanning.ValidIngredientMeasurementUnitCreationRequestInput {
	validIngredientMeasurementUnit := BuildFakeValidIngredientMeasurementUnit()

	return converters.ConvertValidIngredientMeasurementUnitToValidIngredientMeasurementUnitCreationRequestInput(validIngredientMeasurementUnit)
}
