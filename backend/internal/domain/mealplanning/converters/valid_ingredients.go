package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v9/identifiers"
)

// ConvertValidIngredientToValidIngredientUpdateRequestInput creates a ValidIngredientUpdateRequestInput from a ValidIngredient.
func ConvertValidIngredientToValidIngredientUpdateRequestInput(x *mealplanning.ValidIngredient) *mealplanning.ValidIngredientUpdateRequestInput {
	out := &mealplanning.ValidIngredientUpdateRequestInput{
		Name:                           &x.Name,
		Description:                    &x.Description,
		Warning:                        &x.Warning,
		IconPath:                       &x.IconPath,
		ContainsDairy:                  &x.ContainsDairy,
		ContainsPeanut:                 &x.ContainsPeanut,
		ContainsTreeNut:                &x.ContainsTreeNut,
		ContainsEgg:                    &x.ContainsEgg,
		ContainsWheat:                  &x.ContainsWheat,
		ContainsShellfish:              &x.ContainsShellfish,
		ContainsSesame:                 &x.ContainsSesame,
		ContainsFish:                   &x.ContainsFish,
		ContainsGluten:                 &x.ContainsGluten,
		AnimalFlesh:                    &x.AnimalFlesh,
		IsLiquid:                       &x.IsLiquid,
		ContainsSoy:                    &x.ContainsSoy,
		PluralName:                     &x.PluralName,
		AnimalDerived:                  &x.AnimalDerived,
		RestrictToPreparations:         &x.RestrictToPreparations,
		ContaminatesEquipment:          &x.ContaminatesEquipment,
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		StorageInstructions:            &x.StorageInstructions,
		Slug:                           &x.Slug,
		ContainsAlcohol:                &x.ContainsAlcohol,
		ShoppingSuggestions:            &x.ShoppingSuggestions,
		IsStarch:                       &x.IsStarch,
		IsProtein:                      &x.IsProtein,
		IsGrain:                        &x.IsGrain,
		IsFruit:                        &x.IsFruit,
		IsSalt:                         &x.IsSalt,
		IsFat:                          &x.IsFat,
		IsAcid:                         &x.IsAcid,
		IsHeat:                         &x.IsHeat,
	}

	return out
}

// ConvertValidIngredientCreationRequestInputToValidIngredientDatabaseCreationInput creates a DatabaseCreationInput from a ValidIngredientCreationRequestInput.
func ConvertValidIngredientCreationRequestInputToValidIngredientDatabaseCreationInput(x *mealplanning.ValidIngredientCreationRequestInput) *mealplanning.ValidIngredientDatabaseCreationInput {
	out := &mealplanning.ValidIngredientDatabaseCreationInput{
		ID:                             identifiers.New(),
		Name:                           x.Name,
		Description:                    x.Description,
		Warning:                        x.Warning,
		ContainsEgg:                    x.ContainsEgg,
		ContainsDairy:                  x.ContainsDairy,
		ContainsPeanut:                 x.ContainsPeanut,
		ContainsTreeNut:                x.ContainsTreeNut,
		ContainsSoy:                    x.ContainsSoy,
		ContainsWheat:                  x.ContainsWheat,
		ContainsShellfish:              x.ContainsShellfish,
		ContainsSesame:                 x.ContainsSesame,
		ContainsFish:                   x.ContainsFish,
		ContainsGluten:                 x.ContainsGluten,
		AnimalFlesh:                    x.AnimalFlesh,
		IsLiquid:                       x.IsLiquid,
		IconPath:                       x.IconPath,
		PluralName:                     x.PluralName,
		AnimalDerived:                  x.AnimalDerived,
		RestrictToPreparations:         x.RestrictToPreparations,
		ContaminatesEquipment:          x.ContaminatesEquipment,
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		StorageInstructions:            x.StorageInstructions,
		Slug:                           x.Slug,
		ContainsAlcohol:                x.ContainsAlcohol,
		ShoppingSuggestions:            x.ShoppingSuggestions,
		IsStarch:                       x.IsStarch,
		IsProtein:                      x.IsProtein,
		IsGrain:                        x.IsGrain,
		IsFruit:                        x.IsFruit,
		IsSalt:                         x.IsSalt,
		IsFat:                          x.IsFat,
		IsAcid:                         x.IsAcid,
		IsHeat:                         x.IsHeat,
	}

	return out
}

// ConvertValidIngredientToValidIngredientCreationRequestInput builds a ValidIngredientCreationRequestInput from a Ingredient.
func ConvertValidIngredientToValidIngredientCreationRequestInput(x *mealplanning.ValidIngredient) *mealplanning.ValidIngredientCreationRequestInput {
	return &mealplanning.ValidIngredientCreationRequestInput{
		Name:                           x.Name,
		Description:                    x.Description,
		Warning:                        x.Warning,
		ContainsEgg:                    x.ContainsEgg,
		ContainsDairy:                  x.ContainsDairy,
		ContainsPeanut:                 x.ContainsPeanut,
		ContainsTreeNut:                x.ContainsTreeNut,
		ContainsSoy:                    x.ContainsSoy,
		ContainsWheat:                  x.ContainsWheat,
		ContainsShellfish:              x.ContainsShellfish,
		ContainsSesame:                 x.ContainsSesame,
		ContainsFish:                   x.ContainsFish,
		ContainsGluten:                 x.ContainsGluten,
		AnimalFlesh:                    x.AnimalFlesh,
		IsLiquid:                       x.IsLiquid,
		IconPath:                       x.IconPath,
		PluralName:                     x.PluralName,
		AnimalDerived:                  x.AnimalDerived,
		RestrictToPreparations:         x.RestrictToPreparations,
		ContaminatesEquipment:          x.ContaminatesEquipment,
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		StorageInstructions:            x.StorageInstructions,
		Slug:                           x.Slug,
		ContainsAlcohol:                x.ContainsAlcohol,
		ShoppingSuggestions:            x.ShoppingSuggestions,
		IsStarch:                       x.IsStarch,
		IsProtein:                      x.IsProtein,
		IsGrain:                        x.IsGrain,
		IsFruit:                        x.IsFruit,
		IsSalt:                         x.IsSalt,
		IsFat:                          x.IsFat,
		IsAcid:                         x.IsAcid,
		IsHeat:                         x.IsHeat,
	}
}

// ConvertValidIngredientToValidIngredientDatabaseCreationInput builds a ValidIngredientDatabaseCreationInput from a ValidIngredient.
func ConvertValidIngredientToValidIngredientDatabaseCreationInput(x *mealplanning.ValidIngredient) *mealplanning.ValidIngredientDatabaseCreationInput {
	return &mealplanning.ValidIngredientDatabaseCreationInput{
		ID:                             x.ID,
		Name:                           x.Name,
		Description:                    x.Description,
		Warning:                        x.Warning,
		ContainsEgg:                    x.ContainsEgg,
		ContainsDairy:                  x.ContainsDairy,
		ContainsPeanut:                 x.ContainsPeanut,
		ContainsTreeNut:                x.ContainsTreeNut,
		ContainsSoy:                    x.ContainsSoy,
		ContainsWheat:                  x.ContainsWheat,
		ContainsShellfish:              x.ContainsShellfish,
		ContainsSesame:                 x.ContainsSesame,
		ContainsFish:                   x.ContainsFish,
		ContainsGluten:                 x.ContainsGluten,
		AnimalFlesh:                    x.AnimalFlesh,
		IsLiquid:                       x.IsLiquid,
		IconPath:                       x.IconPath,
		PluralName:                     x.PluralName,
		AnimalDerived:                  x.AnimalDerived,
		RestrictToPreparations:         x.RestrictToPreparations,
		ContaminatesEquipment:          x.ContaminatesEquipment,
		MinStorageTemperatureInCelsius: x.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: x.MaxStorageTemperatureInCelsius,
		StorageInstructions:            x.StorageInstructions,
		Slug:                           x.Slug,
		ContainsAlcohol:                x.ContainsAlcohol,
		ShoppingSuggestions:            x.ShoppingSuggestions,
		IsStarch:                       x.IsStarch,
		IsProtein:                      x.IsProtein,
		IsGrain:                        x.IsGrain,
		IsFruit:                        x.IsFruit,
		IsSalt:                         x.IsSalt,
		IsFat:                          x.IsFat,
		IsAcid:                         x.IsAcid,
		IsHeat:                         x.IsHeat,
	}
}
