package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	userIngredientPreferencesTableName = "user_ingredient_preferences"

	userIngredientPreferencesIngredientColumn = "ingredient"
)

func init() {
	registerTableName(userIngredientPreferencesTableName)
}

var userIngredientPreferencesColumns = []string{
	idColumn,
	userIngredientPreferencesIngredientColumn,
	"rating",
	notesColumn,
	"allergy",
	belongsToUserColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildUserIngredientPreferencesQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			applyToEach(filterFromSlice(userIngredientPreferencesColumns, userIngredientPreferencesIngredientColumn), func(i int, s string) string {
				return fmt.Sprintf("%s.%s", userIngredientPreferencesTableName, s)
			}),
			applyToEach(validIngredientsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as valid_ingredient_%s", validIngredientsTableName, s, s)
			}),
			1,
		)

		return slices.Concat(
			pgGen.StandardCRUD(userIngredientPreferencesTableName, userIngredientPreferencesColumns,
				querygen.WithEntity("UserIngredientPreference", "UserIngredientPreferences"),
				querygen.WithOwnership(belongsToUserColumn),
				querygen.WithOmitted(querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetUserIngredientPreferencesForUser",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(userIngredientPreferencesTableName, userIngredientPreferencesColumns, []string{}),
						pgGen.TotalCountSelect(userIngredientPreferencesTableName, userIngredientPreferencesColumns, []string{}),
						userIngredientPreferencesTableName,
						validIngredientsTableName,
						validIngredientsTableName,
						idColumn,
						userIngredientPreferencesTableName,
						userIngredientPreferencesIngredientColumn,
						pgGen.FilterConditions(userIngredientPreferencesTableName, userIngredientPreferencesColumns,
							"user_ingredient_preferences.belongs_to_user = sqlc.arg(belongs_to_user)",
							"valid_ingredients.archived_at IS NULL",
						),
						pgGen.CursorLimitClause(userIngredientPreferencesTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUserIngredientPreference",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						userIngredientPreferencesTableName,
						validIngredientsTableName, validIngredientsTableName, idColumn, userIngredientPreferencesTableName, userIngredientPreferencesIngredientColumn,
						userIngredientPreferencesTableName, archivedAtColumn,
						validIngredientsTableName, archivedAtColumn,
						userIngredientPreferencesTableName, idColumn, idColumn,
						userIngredientPreferencesTableName, belongsToUserColumn, belongsToUserColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
