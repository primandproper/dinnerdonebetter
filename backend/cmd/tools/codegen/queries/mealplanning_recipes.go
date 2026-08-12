package main

import (
	"fmt"
	"strings"

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

		insertColumns := filterForInsert(recipesColumns, lastValidatedAtColumn)

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

		return []*Query{
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
					Name: "CreateRecipe",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					recipesTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(i int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CheckRecipeExistence",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
					recipesTableName, idColumn,
					recipesTableName,
					recipesTableName, archivedAtColumn,
					recipesTableName, idColumn, idColumn,
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
	WHERE %s.%s IS NULL
	AND %s.%s = COALESCE(sqlc.narg(%s), 'approved')::recipe_status
	%s
%s;`,
					strings.Join(applyToEach(recipesColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", recipesTableName, s)
					}), ",\n\t"),
					buildFilterCountSelect(recipesTableName, true, true, []string{}),
					buildTotalCountSelect(recipesTableName, true, []string{}),
					recipesTableName,
					recipesTableName, archivedAtColumn,
					recipesTableName, statusColumn, statusColumn,
					buildFilterConditions(recipesTableName, true, false),
					buildCursorLimitClause(recipesTableName),
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
	WHERE %s.%s IS NULL AND
	%s.%s = sqlc.arg(%s)
	%s
%s;`,
					strings.Join(applyToEach(recipesColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", recipesTableName, s)
					}), ",\n\t"),
					buildFilterCountSelect(recipesTableName, true, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipesTableName, createdByUserColumn, createdByUserColumn)),
					buildTotalCountSelect(recipesTableName, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipesTableName, createdByUserColumn, createdByUserColumn)),
					recipesTableName,
					recipesTableName, archivedAtColumn,
					recipesTableName, createdByUserColumn, createdByUserColumn,
					buildFilterConditions(recipesTableName, true, false, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipesTableName, createdByUserColumn, createdByUserColumn)),
					buildCursorLimitClause(recipesTableName),
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
WHERE %s.%s IS NULL
	AND %s.%s %s
	%s
%s;`,
					strings.Join(applyToEach(recipesColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", recipesTableName, s)
					}), ",\n\t"),
					buildFilterCountSelect(recipesTableName, true, true, []string{}),
					buildTotalCountSelect(recipesTableName, true, []string{}),
					recipesTableName,
					recipesTableName, archivedAtColumn,
					recipesTableName, nameColumn, buildILIKEForArgument("query"),
					buildFilterConditions(recipesTableName, true, false),
					buildCursorLimitClause(recipesTableName),
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
WHERE %s.%s IS NULL
	AND %s.%s = true 
	AND %s.%s = 'approved'
	AND %s.%s %s
	%s
%s;`,
					strings.Join(applyToEach(recipesColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", recipesTableName, s)
					}), ",\n\t"),
					buildFilterCountSelect(recipesTableName, true, true, []string{}),
					buildTotalCountSelect(recipesTableName, true, []string{}),
					recipesTableName,
					recipesTableName, archivedAtColumn,
					recipesTableName, eligibleForMealsColumn,
					recipesTableName, statusColumn,
					recipesTableName, nameColumn, buildILIKEForArgument("query"),
					buildFilterConditions(recipesTableName, true, false),
					buildCursorLimitClause(recipesTableName),
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
WHERE %s.%s IS NULL
	AND %s.%s %s
	%s
	AND NOT EXISTS (
		SELECT 1 FROM recipe_step_instruments rsi
		JOIN recipe_steps rs ON rsi.belongs_to_recipe_step = rs.id
		WHERE rs.belongs_to_recipe = %s.%s
			AND rsi.archived_at IS NULL
			AND rs.archived_at IS NULL
			AND rsi.optional = false
			AND rsi.instrument_id IS NOT NULL
			AND rsi.instrument_id NOT IN (
				SELECT valid_instrument_id FROM account_instrument_ownerships
				WHERE belongs_to_account = sqlc.arg(account_id) AND archived_at IS NULL
			)
	)
%s;`,
					strings.Join(applyToEach(recipesColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", recipesTableName, s)
					}), ",\n\t"),
					buildFilterCountSelect(recipesTableName, true, true, []string{}),
					buildTotalCountSelect(recipesTableName, true, []string{}),
					recipesTableName,
					recipesTableName, archivedAtColumn,
					recipesTableName, nameColumn, buildILIKEForArgument("query"),
					buildFilterConditions(recipesTableName, true, false),
					recipesTableName, idColumn,
					buildCursorLimitClause(recipesTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ScanRecipeIDsForReindex",
					Type: ManyType,
				},
				Content: buildReindexScanQuery(recipesTableName),
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
					Name: "UpdateRecipeStatus",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s),
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
					recipesTableName,
					statusColumn, statusColumn,
					lastUpdatedAtColumn, currentTimeExpression,
					archivedAtColumn,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateRecipeLastIndexedAt",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s = sqlc.arg(%s) AND %s IS NULL;`,
					recipesTableName,
					lastIndexedAtColumn,
					currentTimeExpression,
					idColumn,
					idColumn,
					archivedAtColumn,
				)),
			},
		}
	default:
		return nil
	}
}
