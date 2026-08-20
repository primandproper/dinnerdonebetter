package main

import (
	"github.com/primandproper/platform-go/v12/database/querygen"
)

const (
	mealComponentsTableName = "meal_components"
	belongsToMealColumn     = "belongs_to_meal"
)

func init() {
	registerTableName(mealComponentsTableName)
}

var mealComponentsColumns = []string{
	idColumn,
	belongsToMealColumn,
	recipeIDColumn,
	"meal_component_type",
	"recipe_scale",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildMealComponentsQueries(database string) []*Query {
	switch database {
	case postgres:

		return pgGen.StandardCRUD(mealComponentsTableName, mealComponentsColumns,
			querygen.WithEntity("MealComponent", "MealComponents"),
			querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
		)
	default:
		return nil
	}
}
