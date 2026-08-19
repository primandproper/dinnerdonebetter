package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	userSessionsTableName = "user_sessions"

	sessionTokenIDColumn = "session_token_id"
	refreshTokenIDColumn = "refresh_token_id"
	clientIPColumn       = "client_ip"
	userAgentColumn      = "user_agent"
	deviceNameColumn     = "device_name"
	loginMethodColumn    = "login_method"
	lastActiveAtColumn   = "last_active_at"
	revokedAtColumn      = "revoked_at"
)

func init() {
	registerTableName(userSessionsTableName)
}

var userSessionsColumns = []string{
	idColumn,
	belongsToUserColumn,
	sessionTokenIDColumn,
	refreshTokenIDColumn,
	clientIPColumn,
	userAgentColumn,
	deviceNameColumn,
	loginMethodColumn,
	createdAtColumn,
	lastActiveAtColumn,
	expiresAtColumn,
	revokedAtColumn,
}

func buildUserSessionsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := applyToEach(userSessionsColumns, func(i int, s string) string {
			return fmt.Sprintf("%s.%s", userSessionsTableName, s)
		})

		return slices.Concat(
			querygen.StandardCRUD(userSessionsTableName, userSessionsColumns,
				querygen.WithEntity("UserSession", "UserSessions"),
				querygen.WithDatabaseOwned(lastActiveAtColumn, revokedAtColumn),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetUserSessionBySessionTokenID",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s > %s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						userSessionsTableName,
						userSessionsTableName, sessionTokenIDColumn, sessionTokenIDColumn,
						userSessionsTableName, revokedAtColumn,
						userSessionsTableName, expiresAtColumn, currentTimeExpression,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUserSessionByRefreshTokenID",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						userSessionsTableName,
						userSessionsTableName, refreshTokenIDColumn, refreshTokenIDColumn,
						userSessionsTableName, revokedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetActiveSessionsForUser",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	(
		SELECT COUNT(%s.%s)
		FROM %s
		WHERE %s.%s = sqlc.arg(%s)
				AND %s.%s IS NULL
				AND %s.%s > %s
				AND %s.%s > COALESCE(sqlc.narg(created_after), (SELECT %s - '999 years'::INTERVAL))
				AND %s.%s < COALESCE(sqlc.narg(created_before), (SELECT %s + '999 years'::INTERVAL))
	) AS filtered_count,
	(
		SELECT COUNT(%s.%s)
		FROM %s
		WHERE %s.%s = sqlc.arg(%s)
				AND %s.%s IS NULL
				AND %s.%s > %s
	) AS total_count
FROM %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s > %s
	AND %s.%s > COALESCE(sqlc.narg(created_after), (SELECT %s - '999 years'::INTERVAL))
	AND %s.%s < COALESCE(sqlc.narg(created_before), (SELECT %s + '999 years'::INTERVAL))
	AND %s.%s > COALESCE(sqlc.narg(cursor), '')
ORDER BY %s.%s DESC
LIMIT COALESCE(sqlc.narg(result_limit), 50);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						// filtered_count subquery
						userSessionsTableName, idColumn,
						userSessionsTableName,
						userSessionsTableName, belongsToUserColumn, belongsToUserColumn,
						userSessionsTableName, revokedAtColumn,
						userSessionsTableName, expiresAtColumn, currentTimeExpression,
						userSessionsTableName, createdAtColumn, currentTimeExpression,
						userSessionsTableName, createdAtColumn, currentTimeExpression,
						// total_count subquery
						userSessionsTableName, idColumn,
						userSessionsTableName,
						userSessionsTableName, belongsToUserColumn, belongsToUserColumn,
						userSessionsTableName, revokedAtColumn,
						userSessionsTableName, expiresAtColumn, currentTimeExpression,
						// main query
						userSessionsTableName,
						userSessionsTableName, belongsToUserColumn, belongsToUserColumn,
						userSessionsTableName, revokedAtColumn,
						userSessionsTableName, expiresAtColumn, currentTimeExpression,
						userSessionsTableName, createdAtColumn, currentTimeExpression,
						userSessionsTableName, createdAtColumn, currentTimeExpression,
						userSessionsTableName, idColumn,
						userSessionsTableName, lastActiveAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "RevokeUserSession",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
						userSessionsTableName,
						revokedAtColumn, currentTimeExpression,
						userSessionsTableName, idColumn, idColumn,
						userSessionsTableName, belongsToUserColumn, belongsToUserColumn,
						userSessionsTableName, revokedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "RevokeAllSessionsForUser",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
						userSessionsTableName,
						revokedAtColumn, currentTimeExpression,
						userSessionsTableName, belongsToUserColumn, belongsToUserColumn,
						userSessionsTableName, revokedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "RevokeAllSessionsForUserExcept",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s != sqlc.arg(session_id)
	AND %s.%s IS NULL;`,
						userSessionsTableName,
						revokedAtColumn, currentTimeExpression,
						userSessionsTableName, belongsToUserColumn, belongsToUserColumn,
						userSessionsTableName, idColumn,
						userSessionsTableName, revokedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateSessionTokenIDs",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s),
	%s = sqlc.arg(%s),
	%s = sqlc.arg(%s),
	%s = %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
						userSessionsTableName,
						sessionTokenIDColumn, sessionTokenIDColumn,
						refreshTokenIDColumn, refreshTokenIDColumn,
						expiresAtColumn, expiresAtColumn,
						lastActiveAtColumn, currentTimeExpression,
						userSessionsTableName, idColumn, idColumn,
						userSessionsTableName, revokedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "TouchSessionLastActive",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
						userSessionsTableName,
						lastActiveAtColumn, currentTimeExpression,
						userSessionsTableName, sessionTokenIDColumn, sessionTokenIDColumn,
						userSessionsTableName, revokedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "CleanupExpiredSessions",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s
WHERE %s.%s IS NULL
	AND %s.%s < %s;`,
						userSessionsTableName,
						revokedAtColumn, currentTimeExpression,
						userSessionsTableName, revokedAtColumn,
						userSessionsTableName, expiresAtColumn, currentTimeExpression,
					)),
				},
			},
		)
	default:
		return nil
	}
}
