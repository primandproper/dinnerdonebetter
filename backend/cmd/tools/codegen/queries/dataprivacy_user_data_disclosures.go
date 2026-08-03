package main

import (
	"fmt"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"

	"github.com/cristalhq/builq"
)

const (
	userDataDisclosuresTableName = "user_data_disclosures"
	reportIDColumn               = "report_id"
	expiresAtColumn              = "expires_at"
	completedAtColumn            = "completed_at"
)

func init() {
	registerTableName(userDataDisclosuresTableName)
}

var (
	userDataDisclosuresColumns = []string{
		idColumn,
		belongsToUserColumn,
		statusColumn,
		reportIDColumn,
		expiresAtColumn,
		createdAtColumn,
		lastUpdatedAtColumn,
		completedAtColumn,
		archivedAtColumn,
	}
)

func buildUserDataDisclosuresQueries(database string) []*Query {
	switch database {
	case postgres:
		insertColumns := filterForInsert(userDataDisclosuresColumns, statusColumn, reportIDColumn, completedAtColumn)
		fullSelectColumns := applyToEach(userDataDisclosuresColumns, func(_ int, s string) string {
			return fullColumnName(userDataDisclosuresTableName, s)
		})

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "CreateUserDataDisclosure",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					userDataDisclosuresTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(_ int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetUserDataDisclosure",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					userDataDisclosuresTableName,
					userDataDisclosuresTableName, idColumn, idColumn,
					userDataDisclosuresTableName, archivedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetUserDataDisclosuresForUser",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(user_id)
	%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(
						userDataDisclosuresTableName,
						false,
						true,
						nil,
						fmt.Sprintf("%s.%s = sqlc.arg(user_id)", userDataDisclosuresTableName, belongsToUserColumn),
					),
					buildTotalCountSelect(
						userDataDisclosuresTableName,
						true,
						nil,
						fmt.Sprintf("%s.%s = sqlc.arg(user_id)", userDataDisclosuresTableName, belongsToUserColumn),
					),
					userDataDisclosuresTableName,
					userDataDisclosuresTableName, archivedAtColumn,
					userDataDisclosuresTableName, belongsToUserColumn,
					buildFilterConditions(userDataDisclosuresTableName, false, false),
					buildCursorLimitClause(userDataDisclosuresTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetExpiredUserDataDisclosures",
					Type: ManyType,
				},
				// Every disclosure past its expiry that has not already been reaped, oldest
				// first. Not filtered to 'completed': a disclosure that never produced a
				// report still has an expiry that ought to mean something, and one that did
				// produce a report has a report_id for the reaper to destroy. Excluding
				// 'expired' is what makes the sweep converge — a reaped row stops coming back.
				//
				// The limit is fixed rather than a parameter because sqlc cannot infer a type
				// for a bare LIMIT argument and types it as interface{}. The reaper drains a
				// backlog by taking more than one batch per run, not by asking for a bigger one.
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s != 'expired'
	AND %s.%s <= %s
ORDER BY %s.%s
LIMIT %d;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					userDataDisclosuresTableName,
					userDataDisclosuresTableName, archivedAtColumn,
					userDataDisclosuresTableName, statusColumn,
					userDataDisclosuresTableName, expiresAtColumn, currentTimeExpression,
					userDataDisclosuresTableName, expiresAtColumn,
					dataprivacy.ExpiredUserDataDisclosureBatchSize,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "MarkUserDataDisclosureExpired",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = 'expired',
	%s = %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
					userDataDisclosuresTableName,
					statusColumn,
					lastUpdatedAtColumn, currentTimeExpression,
					userDataDisclosuresTableName, idColumn, idColumn,
					userDataDisclosuresTableName, archivedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "MarkUserDataDisclosureCompleted",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = 'completed',
	%s = sqlc.arg(%s),
	%s = %s,
	%s = %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
					userDataDisclosuresTableName,
					statusColumn,
					reportIDColumn, reportIDColumn,
					completedAtColumn, currentTimeExpression,
					lastUpdatedAtColumn, currentTimeExpression,
					userDataDisclosuresTableName, idColumn, idColumn,
					userDataDisclosuresTableName, archivedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "MarkUserDataDisclosureFailed",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = 'failed',
	%s = %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
					userDataDisclosuresTableName,
					statusColumn,
					lastUpdatedAtColumn, currentTimeExpression,
					userDataDisclosuresTableName, idColumn, idColumn,
					userDataDisclosuresTableName, archivedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "MarkUserDataDisclosureProcessing",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = 'processing',
	%s = %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
					userDataDisclosuresTableName,
					statusColumn,
					lastUpdatedAtColumn, currentTimeExpression,
					userDataDisclosuresTableName, idColumn, idColumn,
					userDataDisclosuresTableName, archivedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveUserDataDisclosure",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
					userDataDisclosuresTableName, archivedAtColumn, currentTimeExpression,
					userDataDisclosuresTableName, idColumn, idColumn,
					userDataDisclosuresTableName, archivedAtColumn,
				)),
			},
		}
	default:
		return nil
	}
}
