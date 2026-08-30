package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	validPrepTaskConfigsTableName = "valid_prep_task_configs"
)

var validPrepTaskConfigsColumns = []string{
	idColumn,
	validIngredientIDColumn,
	validPreparationIDColumn,
	"minimum_storage_duration_in_seconds",
	"maximum_storage_duration_in_seconds",
	"storage_container_type",
	minimumStorageTemperatureInCelsiusColumn,
	maximumStorageTemperatureInCelsiusColumn,
	storageInstructionsColumn,
	notesColumn,
	"source",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidPrepTaskConfigsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			mergeColumns(
				applyToEach(filterFromSlice(validPrepTaskConfigsColumns, "valid_preparation_id", "valid_ingredient_id"), func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_prep_task_config_%s", validPrepTaskConfigsTableName, s, s)
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
			pgGen.StandardCRUD(validPrepTaskConfigsTableName, validPrepTaskConfigsColumns,
				querygen.WithEntity("ValidPrepTaskConfig", "ValidPrepTaskConfigs"),
				querygen.WithNullable("maximum_storage_duration_in_seconds", maximumStorageTemperatureInCelsiusColumn, minimumStorageTemperatureInCelsiusColumn),
				querygen.WithOmitted(querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPrepTaskConfigsForIngredient",
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
						pgGen.FilterCountSelect(validPrepTaskConfigsTableName, validPrepTaskConfigsColumns, []string{}),
						pgGen.TotalCountSelect(validPrepTaskConfigsTableName, validPrepTaskConfigsColumns, []string{}),
						validPrepTaskConfigsTableName,
						validIngredientsTableName,
						validPrepTaskConfigsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validPrepTaskConfigsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						pgGen.FilterConditions(validPrepTaskConfigsTableName, validPrepTaskConfigsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPrepTaskConfigsTableName, validIngredientIDColumn, idColumn),
						),
						pgGen.CursorLimitClause(validPrepTaskConfigsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPrepTaskConfigsForPreparation",
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
						pgGen.FilterCountSelect(validPrepTaskConfigsTableName, validPrepTaskConfigsColumns, []string{}),
						pgGen.TotalCountSelect(validPrepTaskConfigsTableName, validPrepTaskConfigsColumns, []string{}),
						validPrepTaskConfigsTableName,
						validIngredientsTableName,
						validPrepTaskConfigsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validPrepTaskConfigsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						pgGen.FilterConditions(validPrepTaskConfigsTableName, validPrepTaskConfigsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPrepTaskConfigsTableName, validPreparationIDColumn, idColumn),
						),
						pgGen.CursorLimitClause(validPrepTaskConfigsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPrepTaskConfigsForIngredientAndPreparation",
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
						pgGen.FilterCountSelect(validPrepTaskConfigsTableName, validPrepTaskConfigsColumns, []string{}),
						pgGen.TotalCountSelect(validPrepTaskConfigsTableName, validPrepTaskConfigsColumns, []string{}),
						validPrepTaskConfigsTableName,
						validIngredientsTableName,
						validPrepTaskConfigsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validPrepTaskConfigsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						pgGen.FilterConditions(validPrepTaskConfigsTableName, validPrepTaskConfigsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPrepTaskConfigsTableName, validIngredientIDColumn, validIngredientIDColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPrepTaskConfigsTableName, validPreparationIDColumn, validPreparationIDColumn),
						),
						pgGen.CursorLimitClause(validPrepTaskConfigsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPrepTaskConfigs",
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
						pgGen.FilterCountSelect(validPrepTaskConfigsTableName, validPrepTaskConfigsColumns, []string{}),
						pgGen.TotalCountSelect(validPrepTaskConfigsTableName, validPrepTaskConfigsColumns, []string{}),
						validPrepTaskConfigsTableName,
						validIngredientsTableName,
						validPrepTaskConfigsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validPrepTaskConfigsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						pgGen.FilterConditions(validPrepTaskConfigsTableName, validPrepTaskConfigsColumns, querygen.Ascending),
						pgGen.CursorLimitClause(validPrepTaskConfigsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPrepTaskConfig",
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
						validPrepTaskConfigsTableName,
						validIngredientsTableName,
						validPrepTaskConfigsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validPrepTaskConfigsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validPrepTaskConfigsTableName,
						archivedAtColumn,
						validIngredientsTableName,
						archivedAtColumn,
						validPreparationsTableName,
						archivedAtColumn,
						validPrepTaskConfigsTableName,
						idColumn,
						idColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
