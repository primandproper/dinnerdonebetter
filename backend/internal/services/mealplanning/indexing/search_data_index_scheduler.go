package indexing

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v9/search/text/indexing"
)

const (
	// IndexTypeRecipes represents the recipes index.
	IndexTypeRecipes = "recipes"
	// IndexTypeMeals represents the meals index.
	IndexTypeMeals = "meals"
	// IndexTypeValidIngredients represents the valid_ingredients index.
	IndexTypeValidIngredients = "valid_ingredients"
	// IndexTypeValidInstruments represents the valid_instruments index.
	IndexTypeValidInstruments = "valid_instruments"
	// IndexTypeValidMeasurementUnits represents the valid_measurement_units index.
	IndexTypeValidMeasurementUnits = "valid_measurement_units"
	// IndexTypeValidPreparations represents the  valid_preparations index.
	IndexTypeValidPreparations = "valid_preparations"
	// IndexTypeValidIngredientStates represents the valid_ingredient_states index.
	IndexTypeValidIngredientStates = "valid_ingredient_states"
	// IndexTypeValidVessels represents the vessels index.
	IndexTypeValidVessels = "vessels"
)

func BuildEatingDataIndexingFunctions(dataManager mealplanning.Repository) map[string]indexing.Function {
	return map[string]indexing.Function{
		IndexTypeValidPreparations:     dataManager.GetValidPreparationIDsThatNeedSearchIndexing,
		IndexTypeRecipes:               dataManager.GetRecipeIDsThatNeedSearchIndexing,
		IndexTypeMeals:                 dataManager.GetMealIDsThatNeedSearchIndexing,
		IndexTypeValidIngredients:      dataManager.GetValidIngredientIDsThatNeedSearchIndexing,
		IndexTypeValidInstruments:      dataManager.GetValidInstrumentIDsThatNeedSearchIndexing,
		IndexTypeValidMeasurementUnits: dataManager.GetValidMeasurementUnitIDsThatNeedSearchIndexing,
		IndexTypeValidIngredientStates: dataManager.GetValidIngredientStateIDsThatNeedSearchIndexing,
		IndexTypeValidVessels:          dataManager.GetValidVesselIDsThatNeedSearchIndexing,
	}
}
