package main

import (
	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

func buildAdminQueries(database string) []*Query {
	switch database {
	case postgres:

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "SetUserAccountStatus",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s,
	%s = sqlc.arg(%s),
	%s = sqlc.arg(%s)
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
					usersTableName,
					lastUpdatedAtColumn, querygen.NowExpression,
					userAccountStatusColumn, userAccountStatusColumn,
					userAccountStatusExplanationColumn, userAccountStatusExplanationColumn,
					archivedAtColumn,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "SetUserRequiresPasswordChange",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s,
	%s = sqlc.arg(%s)
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
					usersTableName,
					lastUpdatedAtColumn, querygen.NowExpression,
					requiresPasswordChangeColumn, requiresPasswordChangeColumn,
					archivedAtColumn,
					idColumn, idColumn,
				)),
			},
		}
	default:
		return nil
	}
}
