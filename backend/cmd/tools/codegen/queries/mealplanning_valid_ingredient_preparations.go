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
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						querygen.FilterCountSelect(validIngredientPreparationsTableName, validIngredientPreparationsColumns, []string{}),
						querygen.TotalCountSelect(validIngredientPreparationsTableName, validIngredientPreparationsColumns, []string{}),
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
						querygen.FilterConditions(validIngredientPreparationsTableName, validIngredientPreparationsColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validIngredientPreparationsTableName, validIngredientIDColumn, idColumn),
						),
						querygen.CursorLimitClause(validIngredientPreparationsTableName),
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
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						querygen.FilterCountSelect(validIngredientPreparationsTableName, validIngredientPreparationsColumns, []string{}),
						querygen.TotalCountSelect(validIngredientPreparationsTableName, validIngredientPreparationsColumns, []string{}),
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
						querygen.FilterConditions(validIngredientPreparationsTableName, validIngredientPreparationsColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validIngredientPreparationsTableName, validPreparationIDColumn, idColumn),
						),
						querygen.CursorLimitClause(validIngredientPreparationsTableName),
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
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						querygen.FilterCountSelect(validIngredientPreparationsTableName, validIngredientPreparationsColumns, []string{}),
						querygen.TotalCountSelect(validIngredientPreparationsTableName, validIngredientPreparationsColumns, []string{}),
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
						querygen.FilterConditions(validIngredientPreparationsTableName, validIngredientPreparationsColumns),
						querygen.CursorLimitClause(validIngredientPreparationsTableName),
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
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						querygen.FilterCountSelect(validIngredientPreparationsTableName, validIngredientPreparationsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPreparationsTableName, idColumn, idColumn)),
						querygen.TotalCountSelect(validIngredientPreparationsTableName, validIngredientPreparationsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPreparationsTableName, idColumn, idColumn)),
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
						querygen.FilterConditions(validIngredientPreparationsTableName, validIngredientPreparationsColumns,
							"valid_ingredients.archived_at IS NULL",
							"valid_preparations.archived_at IS NULL",
							fmt.Sprintf("%s.%s %s", validIngredientsTableName, nameColumn, buildILIKEForArgument("name_query")),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPreparationsTableName, idColumn, idColumn),
						),
						querygen.CursorLimitClause(validIngredientPreparationsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
