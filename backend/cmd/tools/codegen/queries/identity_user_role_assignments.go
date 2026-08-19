package main

import (
	"fmt"
	"slices"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	userRoleAssignmentsTableName = "user_role_assignments"

	userIDColumn    = "user_id"
	accountIDColumn = "account_id"
)

func init() {
	registerTableName(userRoleAssignmentsTableName)
}

var userRoleAssignmentsColumns = []string{
	idColumn,
	userIDColumn,
	roleIDColumn,
	accountIDColumn,
	createdAtColumn,
	archivedAtColumn,
}

func buildUserRoleAssignmentsQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			querygen.StandardCRUD(userRoleAssignmentsTableName, userRoleAssignmentsColumns,
				querygen.WithEntity("UserRoleAssignments", "UserRoleAssignmentss"),
				querygen.WithQueryName(querygen.CreateQuery, "AssignRoleToUser"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveRoleAssignment",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						userRoleAssignmentsTableName,
						archivedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						idColumn, idColumn,
						userIDColumn, userIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveRoleAssignmentsForUserAndAccount",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						userRoleAssignmentsTableName,
						archivedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						userIDColumn, userIDColumn,
						accountIDColumn, accountIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateAccountRoleAssignment",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = sqlc.arg(new_role_id)
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						userRoleAssignmentsTableName,
						roleIDColumn,
						archivedAtColumn,
						userIDColumn, userIDColumn,
						accountIDColumn, accountIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetServicePermissionsForUser",
						Type: ManyType,
					},
					Content: fmt.Sprintf(`WITH RECURSIVE role_tree AS (
	SELECT %s.%s AS %s, %s.%s AS role_name
	FROM %s
	JOIN %s ON %s.%s = %s.%s
	WHERE %s.%s = sqlc.arg(%s)
		AND %s.%s IS NULL
		AND %s.%s IS NULL
		AND %s.%s IS NULL
	UNION
	SELECT %s.%s, %s.%s
	FROM role_tree rt
	JOIN %s ON %s.child_role_id = rt.%s
	JOIN %s ON %s.%s = %s.parent_role_id
	WHERE %s.%s IS NULL
		AND %s.%s IS NULL
)
SELECT DISTINCT %s.%s AS permission_name
FROM role_tree rt
JOIN %s ON %s.%s = rt.%s
JOIN %s ON %s.%s = %s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL`,
						// Base case SELECT
						userRolesTableName, idColumn, roleIDColumn,
						userRolesTableName, nameColumn,
						// FROM / JOIN
						userRoleAssignmentsTableName,
						userRolesTableName, userRolesTableName, idColumn, userRoleAssignmentsTableName, roleIDColumn,
						// WHERE
						userRoleAssignmentsTableName, userIDColumn, userIDColumn,
						userRoleAssignmentsTableName, accountIDColumn,
						userRoleAssignmentsTableName, archivedAtColumn,
						userRolesTableName, archivedAtColumn,
						// Recursive SELECT
						userRolesTableName, idColumn,
						userRolesTableName, nameColumn,
						// Recursive JOIN
						userRoleHierarchyTableName, userRoleHierarchyTableName, roleIDColumn,
						userRolesTableName, userRolesTableName, idColumn, userRoleHierarchyTableName,
						// Recursive WHERE
						userRoleHierarchyTableName, archivedAtColumn,
						userRolesTableName, archivedAtColumn,
						// Final SELECT
						permissionsTableName, nameColumn,
						// Final JOINs
						userRolePermissionsTableName, userRolePermissionsTableName, roleIDColumn, roleIDColumn,
						permissionsTableName, permissionsTableName, idColumn, userRolePermissionsTableName, permissionIDColumn,
						// Final WHERE
						userRolePermissionsTableName, archivedAtColumn,
						permissionsTableName, archivedAtColumn,
					),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetAccountPermissionsForUser",
						Type: ManyType,
					},
					Content: fmt.Sprintf(`WITH RECURSIVE role_tree AS (
	SELECT %s.%s AS %s, %s.%s AS role_name, %s.%s AS %s
	FROM %s
	JOIN %s ON %s.%s = %s.%s
	WHERE %s.%s = sqlc.arg(%s)
		AND %s.%s IS NOT NULL
		AND %s.%s IS NULL
		AND %s.%s IS NULL
	UNION
	SELECT %s.%s, %s.%s, rt.%s
	FROM role_tree rt
	JOIN %s ON %s.child_role_id = rt.%s
	JOIN %s ON %s.%s = %s.parent_role_id
	WHERE %s.%s IS NULL
		AND %s.%s IS NULL
)
SELECT DISTINCT rt.%s, %s.%s AS permission_name
FROM role_tree rt
JOIN %s ON %s.%s = rt.%s
JOIN %s ON %s.%s = %s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL`,
						// Base case SELECT
						userRolesTableName, idColumn, roleIDColumn,
						userRolesTableName, nameColumn,
						userRoleAssignmentsTableName, accountIDColumn, accountIDColumn,
						// FROM / JOIN
						userRoleAssignmentsTableName,
						userRolesTableName, userRolesTableName, idColumn, userRoleAssignmentsTableName, roleIDColumn,
						// WHERE
						userRoleAssignmentsTableName, userIDColumn, userIDColumn,
						userRoleAssignmentsTableName, accountIDColumn,
						userRoleAssignmentsTableName, archivedAtColumn,
						userRolesTableName, archivedAtColumn,
						// Recursive SELECT
						userRolesTableName, idColumn,
						userRolesTableName, nameColumn,
						accountIDColumn,
						// Recursive JOIN
						userRoleHierarchyTableName, userRoleHierarchyTableName, roleIDColumn,
						userRolesTableName, userRolesTableName, idColumn, userRoleHierarchyTableName,
						// Recursive WHERE
						userRoleHierarchyTableName, archivedAtColumn,
						userRolesTableName, archivedAtColumn,
						// Final SELECT
						accountIDColumn,
						permissionsTableName, nameColumn,
						// Final JOINs
						userRolePermissionsTableName, userRolePermissionsTableName, roleIDColumn, roleIDColumn,
						permissionsTableName, permissionsTableName, idColumn, userRolePermissionsTableName, permissionIDColumn,
						// Final WHERE
						userRolePermissionsTableName, archivedAtColumn,
						permissionsTableName, archivedAtColumn,
					),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetServiceRoleNamesForUser",
						Type: ManyType,
					},
					Content: fmt.Sprintf(`SELECT %s.%s AS role_name
FROM %s
JOIN %s ON %s.%s = %s.%s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL`,
						userRolesTableName, nameColumn,
						userRoleAssignmentsTableName,
						userRolesTableName, userRolesTableName, idColumn, userRoleAssignmentsTableName, roleIDColumn,
						userRoleAssignmentsTableName, userIDColumn, userIDColumn,
						userRoleAssignmentsTableName, accountIDColumn,
						userRoleAssignmentsTableName, archivedAtColumn,
						userRolesTableName, archivedAtColumn,
					),
				},
			},
		)
	default:
		return nil
	}
}
