package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	recipesTableName = "recipes"

	belongsToRecipeColumn  = "belongs_to_recipe"
	recipeIDColumn         = "recipe_id"
	lastValidatedAtColumn  = "last_validated_at"
	eligibleForMealsColumn = "eligible_for_meals"
	statusColumn           = "status"
)

func init() {
	registerTableName(recipesTableName)
}

var recipesColumns = []string{
	idColumn,
	nameColumn,
	slugColumn,
	"source",
	"source_isbn",
	descriptionColumn,
	statusColumn,
	"inspired_by_recipe_id",
	"min_estimated_portions",
	"max_estimated_portions",
	"portion_name",
	"plural_portion_name",
	eligibleForMealsColumn,
	"yields_component_type",
	lastIndexedAtColumn,
	lastValidatedAtColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	createdByUserColumn,
}

func buildRecipesQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := append(
			applyToEach(recipesColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s", recipesTableName, s)
			}),
			mergeColumns(
				applyToEach(filterFromSlice(recipeStepsColumns, preparationIDColumn), func(i int, s string) string {
					return fmt.Sprintf("%s.%s as recipe_step_%s", recipeStepsTableName, s, s)
				}),
				applyToEach(validPreparationsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as recipe_step_preparation_%s", validPreparationsTableName, s, s)
				}),
				2,
			)...,
		)

		return slices.Concat(
			querygen.StandardCRUD(recipesTableName, recipesColumns,
				querygen.WithEntity("Recipe", "Recipes"),
				querygen.WithDatabaseOwned(lastValidatedAtColumn),
				querygen.WithImmutable(nameColumn, slugColumn, "source", "source_isbn", descriptionColumn, "inspired_by_recipe_id", "min_estimated_portions", "max_estimated_portions", "portion_name", "plural_portion_name", eligibleForMealsColumn, "yields_component_type", createdByUserColumn),
				querygen.WithQueryName(querygen.UpdateQuery, "UpdateRecipeStatus"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.GetQuery, querygen.ListQuery, querygen.MarkAsIndexedQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveRecipe",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s IS NULL AND %s = sqlc.arg(%s) AND %s = sqlc.arg(%s);`,
						recipesTableName,
						archivedAtColumn,
						currentTimeExpression,
						archivedAtColumn,
						createdByUserColumn,
						createdByUserColumn,
						idColumn,
						idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeByID",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	LEFT JOIN %s ON %s.%s=%s.%s AND %s.%s IS NULL
	LEFT JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
ORDER BY %s.%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						recipesTableName,
						recipeStepsTableName, recipesTableName, idColumn, recipeStepsTableName, belongsToRecipeColumn, recipeStepsTableName, archivedAtColumn,
						validPreparationsTableName, recipeStepsTableName, preparationIDColumn, validPreparationsTableName, idColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
						recipeStepsTableName, indexColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeByIDAndAuthorID",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	LEFT JOIN %s ON %s.%s=%s.%s AND %s.%s IS NULL
	LEFT JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
ORDER BY %s.%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						recipesTableName,
						recipeStepsTableName, recipesTableName, idColumn, recipeStepsTableName, belongsToRecipeColumn, recipeStepsTableName, archivedAtColumn,
						validPreparationsTableName, recipeStepsTableName, preparationIDColumn, validPreparationsTableName, idColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
						recipesTableName, createdByUserColumn, createdByUserColumn,
						recipeStepsTableName, indexColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipes",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(applyToEach(recipesColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", recipesTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(recipesTableName, recipesColumns, []string{}),
						querygen.TotalCountSelect(recipesTableName, recipesColumns, []string{}),
						recipesTableName,
						querygen.FilterConditions(recipesTableName, recipesColumns,
							fmt.Sprintf("%s.%s = COALESCE(sqlc.narg(%s), 'approved')::recipe_status", recipesTableName, statusColumn, statusColumn),
						),
						querygen.CursorLimitClause(recipesTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipesCreatedByUser",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(applyToEach(recipesColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", recipesTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(recipesTableName, recipesColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipesTableName, createdByUserColumn, createdByUserColumn)),
						querygen.TotalCountSelect(recipesTableName, recipesColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipesTableName, createdByUserColumn, createdByUserColumn)),
						recipesTableName,
						querygen.FilterConditions(recipesTableName, recipesColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipesTableName, createdByUserColumn, createdByUserColumn),
						),
						querygen.CursorLimitClause(recipesTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipesWithIDs",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	LEFT JOIN %s ON %s.%s=%s.%s AND %s.%s IS NULL
	LEFT JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[])
ORDER BY %s.%s ASC;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						recipesTableName,
						recipeStepsTableName, recipesTableName, idColumn, recipeStepsTableName, belongsToRecipeColumn, recipeStepsTableName, archivedAtColumn,
						validPreparationsTableName, recipeStepsTableName, preparationIDColumn, validPreparationsTableName, idColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn,
						recipesTableName, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "RecipeSearch",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(applyToEach(recipesColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", recipesTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(recipesTableName, recipesColumns, []string{}),
						querygen.TotalCountSelect(recipesTableName, recipesColumns, []string{}),
						recipesTableName,
						querygen.FilterConditions(recipesTableName, recipesColumns,
							fmt.Sprintf("%s.%s %s", recipesTableName, nameColumn, buildILIKEForArgument("query")),
						),
						querygen.CursorLimitClause(recipesTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SearchForMealEligibleRecipes",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(applyToEach(recipesColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", recipesTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(recipesTableName, recipesColumns, []string{}),
						querygen.TotalCountSelect(recipesTableName, recipesColumns, []string{}),
						recipesTableName,
						querygen.FilterConditions(recipesTableName, recipesColumns,
							fmt.Sprintf("%s.%s = true", recipesTableName, eligibleForMealsColumn),
							fmt.Sprintf("%s.%s = 'approved'", recipesTableName, statusColumn),
							fmt.Sprintf("%s.%s %s", recipesTableName, nameColumn, buildILIKEForArgument("query")),
						),
						querygen.CursorLimitClause(recipesTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SearchForRecipesWithInstrumentOwnership",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(applyToEach(recipesColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", recipesTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(recipesTableName, recipesColumns, []string{}),
						querygen.TotalCountSelect(recipesTableName, recipesColumns, []string{}),
						recipesTableName,
						querygen.FilterConditions(recipesTableName, recipesColumns,
							fmt.Sprintf("%s.%s %s", recipesTableName, nameColumn, buildILIKEForArgument("query")),
							"NOT EXISTS (\n\t\tSELECT 1 FROM recipe_step_instruments rsi\n\t\tJOIN recipe_steps rs ON rsi.belongs_to_recipe_step = rs.id\n\t\tWHERE rs.belongs_to_recipe = recipes.id\n\t\t\t\tAND rsi.archived_at IS NULL\n\t\t\t\tAND rs.archived_at IS NULL\n\t\t\t\tAND rsi.optional = false\n\t\t\t\tAND rsi.instrument_id IS NOT NULL\n\t\t\t\tAND rsi.instrument_id NOT IN (\n\t\t\t\t\tSELECT valid_instrument_id FROM account_instrument_ownerships\n\t\t\t\t\tWHERE belongs_to_account = sqlc.arg(account_id) AND archived_at IS NULL\n\t\t\t\t)\n\t)",
						),
						querygen.CursorLimitClause(recipesTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipesNeedingIndexing",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
FROM %s
WHERE %s.%s IS NULL
	AND (
		%s.%s IS NULL
		OR %s.%s < %s - '24 hours'::INTERVAL
	);`,
						recipesTableName, idColumn,
						recipesTableName,
						recipesTableName, archivedAtColumn,
						recipesTableName, lastIndexedAtColumn,
						recipesTableName, lastIndexedAtColumn, currentTimeExpression,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeIDsForMeal",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
GROUP BY %s.%s
ORDER BY %s.%s;`,
						recipesTableName, idColumn,
						recipesTableName,
						mealComponentsTableName, mealComponentsTableName, recipeIDColumn, recipesTableName, idColumn,
						mealsTableName, mealComponentsTableName, belongsToMealColumn, mealsTableName, idColumn,
						recipesTableName, archivedAtColumn,
						mealsTableName, idColumn, mealIDColumn,
						recipesTableName, idColumn,
						recipesTableName, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateRecipe",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						recipesTableName,
						strings.Join(applyToEach(filterForUpdate(recipesColumns, statusColumn, lastValidatedAtColumn, createdByUserColumn), func(i int, s string) string {
							return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
						}), ",\n\t"),
						lastUpdatedAtColumn, currentTimeExpression,
						archivedAtColumn,
						createdByUserColumn, createdByUserColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "MarkRecipesAsIndexed",
						Type: ExecRowsType,
					},
					// Bulk, and named for what querygen emits, so migrating this table to
					// StandardCRUD deletes this declaration without renaming anything.
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s
WHERE %s = ANY(sqlc.arg(%s)::text[]);`,
						recipesTableName,
						lastIndexedAtColumn,
						currentTimeExpression,
						idColumn,
						idsArg,
					)),
				},
			},
		)
	default:
		return nil
	}
}
