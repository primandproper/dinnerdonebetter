package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
)

// BuildFakeUserIngredientPreference builds a faked user ingredient preference.
func BuildFakeUserIngredientPreference() *types.UserIngredientPreference {
	preference := fake.BuildFakeRecord[types.UserIngredientPreference]()

	preference.Ingredient = *BuildFakeValidIngredient()

	// A rating inside the range the type validates, which is narrow enough that a
	// number chosen from anywhere else is outside it.
	preference.Rating = 1

	return preference
}

// BuildFakeUserIngredientPreferencesList builds a faked UserIngredientPreferenceList.
func BuildFakeUserIngredientPreferencesList() *filtering.QueryFilteredResult[types.UserIngredientPreference] {
	return fake.BuildFakePage(BuildFakeUserIngredientPreference)
}

// BuildFakeUserIngredientPreferenceUpdateRequestInput builds a faked UserIngredientPreferenceUpdateRequestInput from a preference.
func BuildFakeUserIngredientPreferenceUpdateRequestInput() *types.UserIngredientPreferenceUpdateRequestInput {
	preference := BuildFakeUserIngredientPreference()

	return converters.ConvertUserIngredientPreferenceToUserIngredientPreferenceUpdateRequestInput(preference)
}

// BuildFakeUserIngredientPreferenceCreationRequestInput builds a faked UserIngredientPreferenceCreationRequestInput.
func BuildFakeUserIngredientPreferenceCreationRequestInput() *types.UserIngredientPreferenceCreationRequestInput {
	preference := BuildFakeUserIngredientPreference()

	return converters.ConvertUserIngredientPreferenceToUserIngredientPreferenceCreationRequestInput(preference)
}
