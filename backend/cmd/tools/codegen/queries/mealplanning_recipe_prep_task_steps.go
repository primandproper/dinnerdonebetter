package main

import (
	"github.com/primandproper/platform-go/v12/database/querygen"
)

const (
	recipePrepTaskStepsTableName = "recipe_prep_task_steps"

	satisfiesRecipeStepColumn = "satisfies_recipe_step"
)

var recipePrepTaskStepsColumns = []string{
	idColumn,
	belongsToRecipeStepColumn,
	"belongs_to_recipe_prep_task",
	satisfiesRecipeStepColumn,
}

func buildRecipePrepTaskStepsQueries(database string) []*Query {
	switch database {
	case postgres:

		return pgGen.StandardCRUD(recipePrepTaskStepsTableName, recipePrepTaskStepsColumns,
			querygen.WithEntity("RecipePrepTaskStep", "RecipePrepTaskSteps"),
			querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
		)
	default:
		return nil
	}
}
