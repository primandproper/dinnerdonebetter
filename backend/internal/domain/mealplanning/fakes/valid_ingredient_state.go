package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v5/filtering"
)

// BuildFakeValidIngredientState builds a faked valid preparation.
func BuildFakeValidIngredientState() *types.ValidIngredientState {
	return &types.ValidIngredientState{
		ID:            BuildFakeID(),
		Name:          buildUniqueString(),
		Description:   buildUniqueString(),
		IconPath:      buildUniqueString(),
		Slug:          buildUniqueString(),
		PastTense:     buildUniqueString(),
		AttributeType: types.ValidIngredientStateAttributeTypeOther,
		CreatedAt:     BuildFakeTime(),
	}
}

// BuildFakeValidIngredientStatesList builds a faked ValidIngredientStateList.
func BuildFakeValidIngredientStatesList() *filtering.QueryFilteredResult[types.ValidIngredientState] {
	var examples []*types.ValidIngredientState
	for range exampleQuantity {
		examples = append(examples, BuildFakeValidIngredientState())
	}

	return &filtering.QueryFilteredResult[types.ValidIngredientState]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeValidIngredientStateUpdateRequestInput builds a faked ValidIngredientStateUpdateRequestInput from a valid preparation.
func BuildFakeValidIngredientStateUpdateRequestInput() *types.ValidIngredientStateUpdateRequestInput {
	validIngredientState := BuildFakeValidIngredientState()
	return converters.ConvertValidIngredientStateToValidIngredientStateUpdateRequestInput(validIngredientState)
}

// BuildFakeValidIngredientStateCreationRequestInput builds a faked ValidIngredientStateCreationRequestInput.
func BuildFakeValidIngredientStateCreationRequestInput() *types.ValidIngredientStateCreationRequestInput {
	validIngredientState := BuildFakeValidIngredientState()
	return converters.ConvertValidIngredientStateToValidIngredientStateCreationRequestInput(validIngredientState)
}
