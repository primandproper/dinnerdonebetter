package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/filtering"
)

// BuildFakeValidIngredientMeasurementUnit builds a faked valid ingredient measurement unit.
func BuildFakeValidIngredientMeasurementUnit() *mealplanning.ValidIngredientMeasurementUnit {
	minQty, maxQty := BuildFakeFloat32WithOptionalMax()
	return &mealplanning.ValidIngredientMeasurementUnit{
		ID:                   BuildFakeID(),
		Notes:                buildUniqueString(),
		MeasurementUnit:      *BuildFakeValidMeasurementUnit(),
		Ingredient:           *BuildFakeValidIngredient(),
		MinAllowableQuantity: minQty,
		MaxAllowableQuantity: maxQty,
		CreatedAt:            BuildFakeTime(),
	}
}

// BuildFakeValidIngredientMeasurementUnitsList builds a faked ValidIngredientMeasurementUnitList.
func BuildFakeValidIngredientMeasurementUnitsList() *filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit] {
	var examples []*mealplanning.ValidIngredientMeasurementUnit
	for range exampleQuantity {
		examples = append(examples, BuildFakeValidIngredientMeasurementUnit())
	}

	return &filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
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
