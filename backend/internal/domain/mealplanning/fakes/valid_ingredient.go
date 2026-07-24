package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v6/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeValidIngredient builds a faked valid ingredient.
func BuildFakeValidIngredient() *types.ValidIngredient {
	minST, maxST := BuildFakeOptionalFloat32MinMax()
	return &types.ValidIngredient{
		ID:                             BuildFakeID(),
		Name:                           buildUniqueString(),
		Description:                    buildUniqueString(),
		Warning:                        buildUniqueString(),
		ContainsEgg:                    fake.Bool(),
		ContainsDairy:                  fake.Bool(),
		ContainsPeanut:                 fake.Bool(),
		ContainsTreeNut:                fake.Bool(),
		ContainsSoy:                    fake.Bool(),
		ContainsWheat:                  fake.Bool(),
		ContainsShellfish:              fake.Bool(),
		ContainsSesame:                 fake.Bool(),
		ContainsFish:                   fake.Bool(),
		ContainsGluten:                 fake.Bool(),
		AnimalFlesh:                    fake.Bool(),
		IsLiquid:                       fake.Bool(),
		IconPath:                       buildUniqueString(),
		CreatedAt:                      BuildFakeTime(),
		PluralName:                     buildUniqueString(),
		AnimalDerived:                  fake.Bool(),
		RestrictToPreparations:         fake.Bool(),
		MinStorageTemperatureInCelsius: minST,
		MaxStorageTemperatureInCelsius: maxST,
		StorageInstructions:            buildUniqueString(),
		Slug:                           buildUniqueString(),
		ContainsAlcohol:                fake.Bool(),
		ShoppingSuggestions:            buildUniqueString(),
		IsStarch:                       fake.Bool(),
		IsProtein:                      fake.Bool(),
		IsGrain:                        fake.Bool(),
		IsFruit:                        fake.Bool(),
		IsSalt:                         fake.Bool(),
		IsFat:                          fake.Bool(),
		IsAcid:                         fake.Bool(),
		IsHeat:                         fake.Bool(),
	}
}

// BuildFakeValidIngredientsList builds a faked ValidIngredientList.
func BuildFakeValidIngredientsList() *filtering.QueryFilteredResult[types.ValidIngredient] {
	var examples []*types.ValidIngredient
	for range exampleQuantity {
		examples = append(examples, BuildFakeValidIngredient())
	}

	return &filtering.QueryFilteredResult[types.ValidIngredient]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
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
