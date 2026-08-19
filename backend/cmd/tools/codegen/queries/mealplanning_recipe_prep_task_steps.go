package main

import (
	"github.com/primandproper/platform-go/v11/database/querygen"
)

const (
	recipePrepTaskStepsTableName = "recipe_prep_task_steps"

	satisfiesRecipeStepColumn = "satisfies_recipe_step"
)

func init() {
	registerTableName(recipePrepTaskStepsTableName)
}

var recipePrepTaskStepsColumns = []string{
	idColumn,
	belongsToRecipeStepColumn,
	"belongs_to_recipe_prep_task",
	satisfiesRecipeStepColumn,
}

func buildRecipePrepTaskStepsQueries(database string) []*Query {
	switch database {
	case postgres:

		return querygen.StandardCRUD(recipePrepTaskStepsTableName, recipePrepTaskStepsColumns,
			querygen.WithEntity("RecipePrepTaskStep", "RecipePrepTaskSteps"),
			querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
		)
	default:
		return nil
	}
}
