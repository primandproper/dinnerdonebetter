/*
Package indexing keeps the meal planning search indexes in step with the database.

Each of the eight indexed entities contributes the same three things — how to read one row, how
to page over IDs for a reindex, and how to shape a row into the subset that gets indexed — and
syncsource.Source turns that triple into both of the read seams platform-go's search sync needs.
What used to be one 200-line switch over an index-type string is now eight declarations, and the
compiler checks each one against the index it feeds rather than a map lookup failing at runtime.
*/
package indexing

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	syncsource "github.com/primandproper/platform-go/v12/search/sync/source"
)

type (
	// MealSource and its siblings read one entity as search documents. Each is both a
	// searchsync.Fetcher, for the change feed, and a searchsync.Scanner, for a reindex.
	MealSource = syncsource.Source[mealplanning.Meal, MealSearchSubset]
	// RecipeSource reads recipes as search documents.
	RecipeSource = syncsource.Source[mealplanning.Recipe, RecipeSearchSubset]
	// ValidIngredientSource reads valid ingredients as search documents.
	ValidIngredientSource = syncsource.Source[mealplanning.ValidIngredient, ValidIngredientSearchSubset]
	// ValidInstrumentSource reads valid instruments as search documents.
	ValidInstrumentSource = syncsource.Source[mealplanning.ValidInstrument, ValidInstrumentSearchSubset]
	// ValidMeasurementUnitSource reads valid measurement units as search documents.
	ValidMeasurementUnitSource = syncsource.Source[mealplanning.ValidMeasurementUnit, ValidMeasurementUnitSearchSubset]
	// ValidPreparationSource reads valid preparations as search documents.
	ValidPreparationSource = syncsource.Source[mealplanning.ValidPreparation, ValidPreparationSearchSubset]
	// ValidIngredientStateSource reads valid ingredient states as search documents.
	ValidIngredientStateSource = syncsource.Source[mealplanning.ValidIngredientState, ValidIngredientStateSearchSubset]
	// ValidVesselSource reads valid vessels as search documents.
	ValidVesselSource = syncsource.Source[mealplanning.ValidVessel, ValidVesselSearchSubset]
)

// NewMealSource builds the meals source.
func NewMealSource(repo mealplanning.Repository) (*MealSource, error) {
	return syncsource.New(IndexTypeMeals, repo.GetMeal, repo.ScanMealIDsForReindex, ConvertMealToMealSearchSubset)
}

// NewRecipeSource builds the recipes source.
func NewRecipeSource(repo mealplanning.Repository) (*RecipeSource, error) {
	return syncsource.New(IndexTypeRecipes, repo.GetRecipe, repo.ScanRecipeIDsForReindex, ConvertRecipeToRecipeSearchSubset)
}

// NewValidIngredientSource builds the valid ingredients source.
func NewValidIngredientSource(repo mealplanning.Repository) (*ValidIngredientSource, error) {
	return syncsource.New(
		IndexTypeValidIngredients,
		repo.GetValidIngredient,
		repo.ScanValidIngredientIDsForReindex,
		ConvertValidIngredientToValidIngredientSearchSubset,
	)
}

// NewValidInstrumentSource builds the valid instruments source.
func NewValidInstrumentSource(repo mealplanning.Repository) (*ValidInstrumentSource, error) {
	return syncsource.New(
		IndexTypeValidInstruments,
		repo.GetValidInstrument,
		repo.ScanValidInstrumentIDsForReindex,
		ConvertValidInstrumentToValidInstrumentSearchSubset,
	)
}

// NewValidMeasurementUnitSource builds the valid measurement units source.
func NewValidMeasurementUnitSource(repo mealplanning.Repository) (*ValidMeasurementUnitSource, error) {
	return syncsource.New(
		IndexTypeValidMeasurementUnits,
		repo.GetValidMeasurementUnit,
		repo.ScanValidMeasurementUnitIDsForReindex,
		ConvertValidMeasurementUnitToValidMeasurementUnitSearchSubset,
	)
}

// NewValidPreparationSource builds the valid preparations source.
func NewValidPreparationSource(repo mealplanning.Repository) (*ValidPreparationSource, error) {
	return syncsource.New(
		IndexTypeValidPreparations,
		repo.GetValidPreparation,
		repo.ScanValidPreparationIDsForReindex,
		ConvertValidPreparationToValidPreparationSearchSubset,
	)
}

// NewValidIngredientStateSource builds the valid ingredient states source.
func NewValidIngredientStateSource(repo mealplanning.Repository) (*ValidIngredientStateSource, error) {
	return syncsource.New(
		IndexTypeValidIngredientStates,
		repo.GetValidIngredientState,
		repo.ScanValidIngredientStateIDsForReindex,
		ConvertValidIngredientStateToValidIngredientStateSearchSubset,
	)
}

// NewValidVesselSource builds the valid vessels source.
func NewValidVesselSource(repo mealplanning.Repository) (*ValidVesselSource, error) {
	return syncsource.New(
		IndexTypeValidVessels,
		repo.GetValidVessel,
		repo.ScanValidVesselIDsForReindex,
		ConvertValidVesselToValidVesselSearchSubset,
	)
}
