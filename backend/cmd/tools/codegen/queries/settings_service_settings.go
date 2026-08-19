package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	serviceSettingsTableName = "service_settings"
)

func init() {
	registerTableName(serviceSettingsTableName)
}

var serviceSettingsColumns = []string{
	idColumn,
	nameColumn,
	"type",
	descriptionColumn,
	"default_value",
	"enumeration",
	"admins_only",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildServiceSettingQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			querygen.StandardCRUD(serviceSettingsTableName, serviceSettingsColumns,
				querygen.WithEntity("ServiceSetting", "ServiceSettings"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveServiceSetting",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s = sqlc.arg(%s);`,
						serviceSettingsTableName,
						archivedAtColumn,
						currentTimeExpression,
						idColumn,
						idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetServiceSettings",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	%s
%s;`,
						strings.Join(applyToEach(serviceSettingsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", serviceSettingsTableName, s)
						}), ",\n\t"),
						buildFilterCountSelect(serviceSettingsTableName, true, true, []string{}),
						buildTotalCountSelect(serviceSettingsTableName, true, []string{}),
						serviceSettingsTableName,
						serviceSettingsTableName, archivedAtColumn,
						buildFilterConditions(
							serviceSettingsTableName,
							true,
							true,
						),
						buildCursorLimitClause(serviceSettingsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SearchForServiceSettings",
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
						strings.Join(applyToEach(serviceSettingsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", serviceSettingsTableName, s)
						}), ",\n\t"),
						buildFilterCountSelect(serviceSettingsTableName, true, true, []string{}),
						buildTotalCountSelect(serviceSettingsTableName, true, []string{}),
						serviceSettingsTableName,
						serviceSettingsTableName,
						archivedAtColumn,
						serviceSettingsTableName,
						nameColumn,
						buildILIKEForArgument("name_query"),
						buildFilterConditions(
							serviceSettingsTableName,
							true,
							true,
						),
						buildCursorLimitClause(serviceSettingsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
