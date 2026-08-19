package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	userNotificationsTableName      = "user_notifications"
	contentColumn                   = "content"
	userNotificationStatusDismissed = "dismissed"
)

func init() {
	registerTableName(userNotificationsTableName)
}

var (
	userNotificationsColumns = []string{
		idColumn,
		"content",
		"status",
		belongsToUserColumn,
		createdAtColumn,
		lastUpdatedAtColumn,
	}
)

func buildUserNotificationQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := applyToEach(userNotificationsColumns, func(_ int, s string) string {
			return fullColumnName(userNotificationsTableName, s)
		})

		return slices.Concat(
			querygen.StandardCRUD(userNotificationsTableName, userNotificationsColumns,
				querygen.WithEntity("UserNotification", "UserNotifications"),
				querygen.WithDatabaseOwned("status"),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetUserNotification",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s = sqlc.arg(%s)
AND %s.%s = sqlc.arg(%s);`,
						strings.Join(applyToEach(userNotificationsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", userNotificationsTableName, s)
						}), ",\n\t"),
						userNotificationsTableName,
						belongsToUserColumn, belongsToUserColumn,
						userNotificationsTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "CheckUserNotificationExistence",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS(
	SELECT %s.%s
	FROM %s
	WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
);`,
						userNotificationsTableName, idColumn,
						userNotificationsTableName,
						userNotificationsTableName, idColumn, idColumn,
						userNotificationsTableName, belongsToUserColumn, belongsToUserColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUserNotificationsForUser",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						querygen.FilterCountSelect(
							userNotificationsTableName,
							userNotificationsColumns,
							nil,
							fmt.Sprintf("user_notifications.status != '%s'", userNotificationStatusDismissed),
							"user_notifications.belongs_to_user = sqlc.arg(user_id)",
						),
						querygen.TotalCountSelect(
							userNotificationsTableName,
							userNotificationsColumns,
							nil,
							fmt.Sprintf("user_notifications.status != '%s'", userNotificationStatusDismissed),
							"user_notifications.belongs_to_user = sqlc.arg(user_id)",
						),
						userNotificationsTableName,
						querygen.FilterConditions(userNotificationsTableName, userNotificationsColumns,
							fmt.Sprintf("user_notifications.status != '%s'", userNotificationStatusDismissed),
							"user_notifications.belongs_to_user = sqlc.arg(user_id)",
						),
						querygen.CursorLimitClause(userNotificationsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateUserNotification",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s = sqlc.arg(%s);`,
						userNotificationsTableName,
						strings.Join(applyToEach(filterForUpdate(userNotificationsColumns, contentColumn, belongsToUserColumn), func(i int, s string) string {
							return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
						}), ",\n\t"),
						lastUpdatedAtColumn, currentTimeExpression,
						idColumn, idColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
