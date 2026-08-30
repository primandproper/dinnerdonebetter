package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	serviceSettingsTableName = "service_settings"
)

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
			pgGen.StandardCRUD(serviceSettingsTableName, serviceSettingsColumns,
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
						querygen.NowExpression,
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
WHERE %s
%s;`,
						strings.Join(applyToEach(serviceSettingsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", serviceSettingsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(serviceSettingsTableName, serviceSettingsColumns, []string{}),
						pgGen.TotalCountSelect(serviceSettingsTableName, serviceSettingsColumns, []string{}),
						serviceSettingsTableName,
						pgGen.FilterConditions(serviceSettingsTableName, serviceSettingsColumns, querygen.Ascending),
						pgGen.CursorLimitClause(serviceSettingsTableName, querygen.Ascending),
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
WHERE %s
%s;`,
						strings.Join(applyToEach(serviceSettingsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", serviceSettingsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(serviceSettingsTableName, serviceSettingsColumns, []string{}),
						pgGen.TotalCountSelect(serviceSettingsTableName, serviceSettingsColumns, []string{}),
						serviceSettingsTableName,
						pgGen.FilterConditions(serviceSettingsTableName, serviceSettingsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s %s", serviceSettingsTableName, nameColumn, buildILIKEForArgument("name_query")),
						),
						pgGen.CursorLimitClause(serviceSettingsTableName, querygen.Ascending),
					)),
				},
			},
		)
	default:
		return nil
	}
}
