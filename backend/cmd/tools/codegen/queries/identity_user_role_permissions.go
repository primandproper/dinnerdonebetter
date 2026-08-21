package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	userRolePermissionsTableName = "user_role_permissions"

	roleIDColumn       = "role_id"
	permissionIDColumn = "permission_id"
)

var userRolePermissionsColumns = []string{
	idColumn,
	roleIDColumn,
	permissionIDColumn,
	createdAtColumn,
	archivedAtColumn,
}

func buildUserRolePermissionsQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			pgGen.StandardCRUD(userRolePermissionsTableName, userRolePermissionsColumns,
				querygen.WithEntity("UserRolePermission", "UserRolePermissions"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetUserRolePermissionsForRole",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(applyToEach(userRolePermissionsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", userRolePermissionsTableName, s)
						}), ",\n\t"),
						userRolePermissionsTableName,
						userRolePermissionsTableName, archivedAtColumn,
						userRolePermissionsTableName, roleIDColumn, roleIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveUserRolePermission",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						userRolePermissionsTableName,
						archivedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						roleIDColumn, roleIDColumn,
						permissionIDColumn, permissionIDColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
