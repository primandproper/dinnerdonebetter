package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v9/filtering"
)

// BuildFakeValidIngredientStateIngredient builds a faked valid ingredient preparation.
func BuildFakeValidIngredientStateIngredient() *types.ValidIngredientStateIngredient {
	return &types.ValidIngredientStateIngredient{
		ID:              BuildFakeID(),
		Notes:           buildUniqueString(),
		IngredientState: *BuildFakeValidIngredientState(),
		Ingredient:      *BuildFakeValidIngredient(),
		CreatedAt:       BuildFakeTime(),
	}
}

// BuildFakeValidIngredientStateIngredientsList builds a faked ValidIngredientStateIngredientList.
func BuildFakeValidIngredientStateIngredientsList() *filtering.QueryFilteredResult[types.ValidIngredientStateIngredient] {
	var examples []*types.ValidIngredientStateIngredient
	for range exampleQuantity {
		examples = append(examples, BuildFakeValidIngredientStateIngredient())
	}

	return &filtering.QueryFilteredResult[types.ValidIngredientStateIngredient]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeValidIngredientStateIngredientUpdateRequestInput builds a faked ValidIngredientStateIngredientUpdateRequestInput from a valid ingredient preparation.
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
