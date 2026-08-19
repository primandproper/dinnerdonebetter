package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	recipeStepCompletionConditionIngredientsTableName = "recipe_step_completion_condition_ingredients"

	belongsToRecipeStepCompletionConditionColumn = "belongs_to_recipe_step_completion_condition"
)

func init() {
	registerTableName(recipeStepCompletionConditionIngredientsTableName)
}

var recipeStepCompletionConditionIngredientsColumns = []string{
	idColumn,
	belongsToRecipeStepCompletionConditionColumn,
	"recipe_step_ingredient",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildRecipeStepCompletionConditionIngredientsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := append(
			applyToEach(recipeStepCompletionConditionIngredientsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as %s_%s", recipeStepCompletionConditionIngredientsTableName, s, strings.TrimSuffix(recipeStepCompletionConditionIngredientsTableName, "s"), s)
			}),
			applyToEach(validIngredientStatesColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as %s_%s", validIngredientStatesTableName, s, strings.TrimSuffix(validIngredientStatesTableName, "s"), s)
			})...,
		)

		return slices.Concat(
			querygen.StandardCRUD(recipeStepCompletionConditionIngredientsTableName, recipeStepCompletionConditionIngredientsColumns,
				querygen.WithEntity("RecipeStepCompletionConditionIngredient", "RecipeStepCompletionConditionIngredients"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetAllRecipeStepCompletionConditionIngredientsForRecipeCompletionIDs",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[])
	AND %s.%s IS NULL;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						recipeStepCompletionConditionIngredientsTableName,
						recipeStepCompletionConditionsTableName, recipeStepCompletionConditionIngredientsTableName, belongsToRecipeStepCompletionConditionColumn, recipeStepCompletionConditionsTableName, idColumn,
						validIngredientStatesTableName, recipeStepCompletionConditionsTableName, ingredientStateColumn, validIngredientStatesTableName, idColumn,
						recipeStepCompletionConditionsTableName, archivedAtColumn,
						recipeStepCompletionConditionIngredientsTableName, archivedAtColumn,
						recipeStepCompletionConditionIngredientsTableName, belongsToRecipeStepCompletionConditionColumn,
						validIngredientStatesTableName, archivedAtColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
