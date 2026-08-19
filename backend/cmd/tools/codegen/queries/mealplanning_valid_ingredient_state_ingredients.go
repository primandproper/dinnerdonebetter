package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	validIngredientStateIngredientsTableName = "valid_ingredient_state_ingredients"

	validIngredientStateColumn = "valid_ingredient_state"
	validIngredientColumn      = "valid_ingredient"
)

func init() {
	registerTableName(validIngredientStateIngredientsTableName)
}

var validIngredientStateIngredientsColumns = []string{
	idColumn,
	notesColumn,
	validIngredientStateColumn,
	validIngredientColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidIngredientStateIngredientsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			applyToEach(filterFromSlice(validIngredientStateIngredientsColumns, "valid_ingredient_id", "valid_measurement_unit_id"), func(i int, s string) string {
				return fmt.Sprintf("%s.%s as valid_ingredient_state_ingredient_%s", validIngredientStateIngredientsTableName, s, s)
			}),
			append(
				applyToEach(validIngredientStatesColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_ingredient_state_%s", validIngredientStatesTableName, s, s)
				}),
				applyToEach(validIngredientsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_ingredient_%s", validIngredientsTableName, s, s)
				})...),
			2,
		)

		return slices.Concat(
			querygen.StandardCRUD(validIngredientStateIngredientsTableName, validIngredientStateIngredientsColumns,
				querygen.WithEntity("ValidIngredientStateIngredient", "ValidIngredientStateIngredients"),
				querygen.WithOmitted(querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientStateIngredientsForIngredient",
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
						querygen.FilterCountSelect(validIngredientStateIngredientsTableName, validIngredientStateIngredientsColumns, []string{}),
						querygen.TotalCountSelect(validIngredientStateIngredientsTableName, validIngredientStateIngredientsColumns, []string{}),
						validIngredientStateIngredientsTableName,
						validIngredientsTableName,
						validIngredientStateIngredientsTableName,
						validIngredientColumn,
						validIngredientsTableName,
						idColumn,
						validIngredientStatesTableName,
						validIngredientStateIngredientsTableName,
						validIngredientStateColumn,
						validIngredientStatesTableName,
						idColumn,
						querygen.FilterConditions(validIngredientStateIngredientsTableName, validIngredientStateIngredientsColumns,
							"valid_ingredients.archived_at IS NULL",
							"valid_ingredient_states.archived_at IS NULL",
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validIngredientStateIngredientsTableName, validIngredientColumn, validIngredientColumn),
						),
						querygen.CursorLimitClause(validIngredientStateIngredientsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientStateIngredientsForIngredientState",
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
						querygen.FilterCountSelect(validIngredientStateIngredientsTableName, validIngredientStateIngredientsColumns, []string{}),
						querygen.TotalCountSelect(validIngredientStateIngredientsTableName, validIngredientStateIngredientsColumns, []string{}),
						validIngredientStateIngredientsTableName,
						validIngredientsTableName,
						validIngredientStateIngredientsTableName,
						validIngredientColumn,
						validIngredientsTableName,
						idColumn,
						validIngredientStatesTableName,
						validIngredientStateIngredientsTableName,
						validIngredientStateColumn,
						validIngredientStatesTableName,
						idColumn,
						querygen.FilterConditions(validIngredientStateIngredientsTableName, validIngredientStateIngredientsColumns,
							"valid_ingredients.archived_at IS NULL",
							"valid_ingredient_states.archived_at IS NULL",
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validIngredientStateIngredientsTableName, validIngredientStateColumn, validIngredientStateColumn),
						),
						querygen.CursorLimitClause(validIngredientStateIngredientsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientStateIngredients",
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
						querygen.FilterCountSelect(validIngredientStateIngredientsTableName, validIngredientStateIngredientsColumns, []string{}),
						querygen.TotalCountSelect(validIngredientStateIngredientsTableName, validIngredientStateIngredientsColumns, []string{}),
						validIngredientStateIngredientsTableName,
						validIngredientsTableName,
						validIngredientStateIngredientsTableName,
						validIngredientColumn,
						validIngredientsTableName,
						idColumn,
						validIngredientStatesTableName,
						validIngredientStateIngredientsTableName,
						validIngredientStateColumn,
						validIngredientStatesTableName,
						idColumn,
						querygen.FilterConditions(validIngredientStateIngredientsTableName, validIngredientStateIngredientsColumns,
							"valid_ingredients.archived_at IS NULL",
							"valid_ingredient_states.archived_at IS NULL",
						),
						querygen.CursorLimitClause(validIngredientStateIngredientsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientStateIngredient",
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
						validIngredientStateIngredientsTableName,
						validIngredientsTableName, validIngredientStateIngredientsTableName, validIngredientColumn, validIngredientsTableName, idColumn,
						validIngredientStatesTableName, validIngredientStateIngredientsTableName, validIngredientStateColumn, validIngredientStatesTableName, idColumn,
						validIngredientStateIngredientsTableName, archivedAtColumn,
						validIngredientsTableName, archivedAtColumn,
						validIngredientStatesTableName, archivedAtColumn,
						validIngredientStateIngredientsTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientStateIngredientsWithIDs",
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
						validIngredientStateIngredientsTableName,
						validIngredientsTableName, validIngredientStateIngredientsTableName, validIngredientColumn, validIngredientsTableName, idColumn,
						validIngredientStatesTableName, validIngredientStateIngredientsTableName, validIngredientStateColumn, validIngredientStatesTableName, idColumn,
						validIngredientStateIngredientsTableName, archivedAtColumn,
						validIngredientsTableName, archivedAtColumn,
						validIngredientStatesTableName, archivedAtColumn,
						validIngredientStateIngredientsTableName, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "CheckValidityOfValidIngredientStateIngredientPair",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS(
	SELECT %s.%s
	FROM %s
	WHERE %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s)
	AND %s IS NULL
);`,
						validIngredientStateIngredientsTableName, idColumn,
						validIngredientStateIngredientsTableName,
						validIngredientColumn, validIngredientColumn,
						validIngredientStateColumn, validIngredientStateColumn,
						archivedAtColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
