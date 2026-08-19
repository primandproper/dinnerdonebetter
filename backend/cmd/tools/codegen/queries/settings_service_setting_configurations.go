package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	serviceSettingConfigurationsTableName = "service_setting_configurations"

	serviceSettingIDColumn    = "service_setting_id"
	serviceSettingValueColumn = "value"
)

func init() {
	registerTableName(serviceSettingConfigurationsTableName)
}

var serviceSettingConfigurationsColumns = []string{
	idColumn,
	serviceSettingValueColumn,
	notesColumn,
	serviceSettingIDColumn,
	belongsToUserColumn,
	belongsToAccountColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildServiceSettingConfigurationQueries(database string) []*Query {
	switch database {
	case postgres:

		selectColumnsWithServiceSettingColumns := mergeColumns(
			applyToEach(filterFromSlice(serviceSettingConfigurationsColumns, "service_setting_id"), func(i int, s string) string {
				return fmt.Sprintf("%s.%s", serviceSettingConfigurationsTableName, s)
			}),
			applyToEach(serviceSettingsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as service_setting_%s", serviceSettingsTableName, s, s)
			}),
			3,
		)

		return slices.Concat(
			querygen.StandardCRUD(serviceSettingConfigurationsTableName, serviceSettingConfigurationsColumns,
				querygen.WithEntity("ServiceSettingConfiguration", "ServiceSettingConfigurations"),
				querygen.WithOmitted(querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetServiceSettingConfigurationByID",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(selectColumnsWithServiceSettingColumns, ",\n\t"),
						serviceSettingConfigurationsTableName,
						serviceSettingsTableName, serviceSettingConfigurationsTableName, serviceSettingIDColumn, serviceSettingsTableName, idColumn,
						serviceSettingsTableName, archivedAtColumn,
						serviceSettingConfigurationsTableName, archivedAtColumn,
						serviceSettingConfigurationsTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetServiceSettingConfigurationForAccountBySettingName",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(selectColumnsWithServiceSettingColumns, ",\n\t"),
						serviceSettingConfigurationsTableName,
						serviceSettingsTableName, serviceSettingConfigurationsTableName, serviceSettingIDColumn, serviceSettingsTableName, idColumn,
						serviceSettingsTableName, archivedAtColumn,
						serviceSettingConfigurationsTableName, archivedAtColumn,
						serviceSettingsTableName, nameColumn, nameColumn,
						serviceSettingConfigurationsTableName, belongsToAccountColumn, belongsToAccountColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetServiceSettingConfigurationForUserBySettingName",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(selectColumnsWithServiceSettingColumns, ",\n\t"),
						serviceSettingConfigurationsTableName,
						serviceSettingsTableName, serviceSettingConfigurationsTableName, serviceSettingIDColumn, serviceSettingsTableName, idColumn,
						serviceSettingsTableName, archivedAtColumn,
						serviceSettingConfigurationsTableName, archivedAtColumn,
						serviceSettingsTableName, nameColumn, nameColumn,
						serviceSettingConfigurationsTableName, belongsToUserColumn, belongsToUserColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetServiceSettingConfigurationsForAccount",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
						strings.Join(selectColumnsWithServiceSettingColumns, ",\n\t"),
						buildFilterCountSelect(serviceSettingConfigurationsTableName, true, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", serviceSettingConfigurationsTableName, belongsToAccountColumn, belongsToAccountColumn)),
						buildTotalCountSelect(serviceSettingConfigurationsTableName, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", serviceSettingConfigurationsTableName, belongsToAccountColumn, belongsToAccountColumn)),
						serviceSettingConfigurationsTableName,
						serviceSettingsTableName, serviceSettingConfigurationsTableName, serviceSettingIDColumn, serviceSettingsTableName, idColumn,
						serviceSettingsTableName, archivedAtColumn,
						serviceSettingConfigurationsTableName, archivedAtColumn,
						serviceSettingConfigurationsTableName, belongsToAccountColumn, belongsToAccountColumn,
						buildFilterConditions(
							serviceSettingConfigurationsTableName,
							true,
							true,
						),
						buildCursorLimitClause(serviceSettingConfigurationsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetServiceSettingConfigurationsForUser",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
						strings.Join(selectColumnsWithServiceSettingColumns, ",\n\t"),
						buildFilterCountSelect(serviceSettingConfigurationsTableName, true, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", serviceSettingConfigurationsTableName, belongsToUserColumn, belongsToUserColumn)),
						buildTotalCountSelect(serviceSettingConfigurationsTableName, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", serviceSettingConfigurationsTableName, belongsToUserColumn, belongsToUserColumn)),
						serviceSettingConfigurationsTableName,
						serviceSettingsTableName, serviceSettingConfigurationsTableName, serviceSettingIDColumn, serviceSettingsTableName, idColumn,
						serviceSettingsTableName, archivedAtColumn,
						serviceSettingConfigurationsTableName, archivedAtColumn,
						serviceSettingConfigurationsTableName, belongsToUserColumn, belongsToUserColumn,
						buildFilterConditions(
							serviceSettingConfigurationsTableName,
							true,
							true,
						),
						buildCursorLimitClause(serviceSettingConfigurationsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
