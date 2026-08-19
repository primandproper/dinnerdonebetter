package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	validIngredientPreparationsTableName = "valid_ingredient_preparations"
)

func init() {
	registerTableName(validIngredientPreparationsTableName)
}

var validIngredientPreparationsColumns = []string{
	idColumn,
	notesColumn,
	validPreparationIDColumn,
	validIngredientIDColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidIngredientPreparationsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			mergeColumns(
				applyToEach(filterFromSlice(validIngredientPreparationsColumns, "valid_preparation_id", "valid_ingredient_id"), func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_ingredient_preparation_%s", validIngredientPreparationsTableName, s, s)
				}),
				applyToEach(validIngredientsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_ingredient_%s", validIngredientsTableName, s, s)
				}),
				2,
			),
			applyToEach(validPreparationsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as valid_preparation_%s", validPreparationsTableName, s, s)
			}),
			2,
		)

		return slices.Concat(
			querygen.StandardCRUD(validIngredientPreparationsTableName, validIngredientPreparationsColumns,
				querygen.WithEntity("ValidIngredientPreparation", "ValidIngredientPreparations"),
				querygen.WithOmitted(querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientPreparationsForIngredient",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(validIngredientPreparationsTableName, true, true, []string{}),
						buildTotalCountSelect(validIngredientPreparationsTableName, true, []string{}),
						validIngredientPreparationsTableName,
						validIngredientsTableName,
						validIngredientPreparationsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validIngredientPreparationsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validIngredientPreparationsTableName,
						archivedAtColumn,
						validIngredientPreparationsTableName,
						validIngredientIDColumn,
						idColumn,
						buildFilterConditions(validIngredientPreparationsTableName, true, false),
						buildCursorLimitClause(validIngredientPreparationsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientPreparationsForPreparation",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(validIngredientPreparationsTableName, true, true, []string{}),
						buildTotalCountSelect(validIngredientPreparationsTableName, true, []string{}),
						validIngredientPreparationsTableName,
						validIngredientsTableName,
						validIngredientPreparationsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validIngredientPreparationsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validIngredientPreparationsTableName,
						archivedAtColumn,
						validIngredientPreparationsTableName,
						validPreparationIDColumn,
						idColumn,
						buildFilterConditions(validIngredientPreparationsTableName, true, false),
						buildCursorLimitClause(validIngredientPreparationsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientPreparations",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(validIngredientPreparationsTableName, true, true, []string{}),
						buildTotalCountSelect(validIngredientPreparationsTableName, true, []string{}),
						validIngredientPreparationsTableName,
						validIngredientsTableName,
						validIngredientPreparationsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validIngredientPreparationsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validIngredientPreparationsTableName,
						archivedAtColumn,
						buildFilterConditions(validIngredientPreparationsTableName, true, false),
						buildCursorLimitClause(validIngredientPreparationsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientPreparation",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						validIngredientPreparationsTableName,
						validIngredientsTableName,
						validIngredientPreparationsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validIngredientPreparationsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validIngredientPreparationsTableName,
						archivedAtColumn,
						validIngredientsTableName,
						archivedAtColumn,
						validPreparationsTableName,
						archivedAtColumn,
						validIngredientPreparationsTableName,
						idColumn,
						idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientPreparationsByIDs",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[]);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						validIngredientPreparationsTableName,
						validIngredientsTableName,
						validIngredientPreparationsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validIngredientPreparationsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validIngredientPreparationsTableName,
						archivedAtColumn,
						validIngredientsTableName,
						archivedAtColumn,
						validPreparationsTableName,
						archivedAtColumn,
						validIngredientPreparationsTableName,
						idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "ValidIngredientPreparationPairIsValid",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS(
	SELECT %s
	FROM %s
	WHERE %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s)
	AND %s IS NULL
);`,
						idColumn,
						validIngredientPreparationsTableName,
						validIngredientIDColumn,
						validIngredientIDColumn,
						validPreparationIDColumn,
						validPreparationIDColumn,
						archivedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SearchValidIngredientPreparationsByPreparationAndIngredientName",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s %s
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(validIngredientPreparationsTableName, true, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPreparationsTableName, idColumn, idColumn)),
						buildTotalCountSelect(validIngredientPreparationsTableName, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPreparationsTableName, idColumn, idColumn)),
						validIngredientPreparationsTableName,
						validIngredientsTableName,
						validIngredientPreparationsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validIngredientPreparationsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validIngredientPreparationsTableName,
						archivedAtColumn,
						validIngredientsTableName,
						archivedAtColumn,
						validPreparationsTableName,
						archivedAtColumn,
						validPreparationsTableName,
						idColumn,
						idColumn,
						validIngredientsTableName,
						nameColumn,
						buildILIKEForArgument("name_query"),
						buildFilterConditions(
							validIngredientPreparationsTableName,
							true,
							true,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPreparationsTableName, idColumn, idColumn),
						),
						buildCursorLimitClause(validIngredientPreparationsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
