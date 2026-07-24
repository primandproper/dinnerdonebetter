package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v6/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeUserIngredientPreference builds a faked valid preparation.
func BuildFakeUserIngredientPreference() *types.UserIngredientPreference {
	return &types.UserIngredientPreference{
		ID:         BuildFakeID(),
		Ingredient: *BuildFakeValidIngredient(),
		Rating:     1,
		Notes:      buildUniqueString(),
		Allergy:    fake.Bool(),
		CreatedAt:  BuildFakeTime(),
	}
}

// BuildFakeUserIngredientPreferencesList builds a faked UserIngredientPreferenceList.
func BuildFakeUserIngredientPreferencesList() *filtering.QueryFilteredResult[types.UserIngredientPreference] {
	var examples []*types.UserIngredientPreference
	for range exampleQuantity {
		examples = append(examples, BuildFakeUserIngredientPreference())
	}

	return &filtering.QueryFilteredResult[types.UserIngredientPreference]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeUserIngredientPreferenceUpdateRequestInput builds a faked UserIngredientPreferenceUpdateRequestInput from a valid preparation.
func BuildFakeUserIngredientPreferenceUpdateRequestInput() *types.UserIngredientPreferenceUpdateRequestInput {
	validPreparation := BuildFakeUserIngredientPreference()
	return converters.ConvertUserIngredientPreferenceToUserIngredientPreferenceUpdateRequestInput(validPreparation)
}

// BuildFakeUserIngredientPreferenceCreationRequestInput builds a faked UserIngredientPreferenceCreationRequestInput.
func BuildFakeUserIngredientPreferenceCreationRequestInput() *types.UserIngredientPreferenceCreationRequestInput {
	validPreparation := BuildFakeUserIngredientPreference()
	return converters.ConvertUserIngredientPreferenceToUserIngredientPreferenceCreationRequestInput(validPreparation)
}
