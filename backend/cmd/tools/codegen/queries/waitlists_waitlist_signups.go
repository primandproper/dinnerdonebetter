package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	waitlistSignupsTableName  = "waitlist_signups"
	belongsToWaitlistColumn   = "belongs_to_waitlist"
	waitlistSignupNotesColumn = "notes"
)

func init() {
	registerTableName(waitlistSignupsTableName)
}

var waitlistSignupColumns = []string{
	idColumn,
	waitlistSignupNotesColumn,
	belongsToWaitlistColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	belongsToUserColumn,
	belongsToAccountColumn,
}

func buildWaitlistSignupsQueries(database string) []*Query {
	switch database {
	case postgres:
		insertColumns := filterForInsert(waitlistSignupColumns)
		fullSelectColumns := applyToEach(waitlistSignupColumns, func(_ int, s string) string {
			return fullColumnName(waitlistSignupsTableName, s)
		})

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "CreateWaitlistSignup",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					waitlistSignupsTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(_ int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateWaitlistSignup",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
					waitlistSignupsTableName,
					strings.Join(applyToEach(filterForUpdate(waitlistSignupColumns, belongsToWaitlistColumn, belongsToUserColumn, belongsToAccountColumn), func(_ int, s string) string {
						return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
					}), ",\n\t"),
					lastUpdatedAtColumn, currentTimeExpression,
					archivedAtColumn,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveWaitlistSignup",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
					waitlistSignupsTableName,
					lastUpdatedAtColumn, currentTimeExpression,
					archivedAtColumn, currentTimeExpression,
					archivedAtColumn,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetWaitlistSignup",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s);`,
					strings.Join(fullSelectColumns, ",\n\t"),
					waitlistSignupsTableName,
					waitlistSignupsTableName, archivedAtColumn,
					waitlistSignupsTableName, idColumn, idColumn,
					waitlistSignupsTableName, belongsToWaitlistColumn, belongsToWaitlistColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetWaitlistSignupByID",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
					strings.Join(fullSelectColumns, ",\n\t"),
					waitlistSignupsTableName,
					waitlistSignupsTableName, archivedAtColumn,
					waitlistSignupsTableName, idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CheckWaitlistSignupExistence",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS(
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s = sqlc.arg(%s)
);`,
					waitlistSignupsTableName, idColumn,
					waitlistSignupsTableName,
					waitlistSignupsTableName, archivedAtColumn,
					waitlistSignupsTableName, idColumn, idColumn,
					waitlistSignupsTableName, belongsToWaitlistColumn, belongsToWaitlistColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetWaitlistSignupsForWaitlist",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(waitlistSignupsTableName, true, true, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", waitlistSignupsTableName, belongsToWaitlistColumn, belongsToWaitlistColumn)),
					buildTotalCountSelect(waitlistSignupsTableName, true, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", waitlistSignupsTableName, belongsToWaitlistColumn, belongsToWaitlistColumn)),
					waitlistSignupsTableName,
					waitlistSignupsTableName, archivedAtColumn,
					waitlistSignupsTableName, belongsToWaitlistColumn, belongsToWaitlistColumn,
					buildFilterConditions(waitlistSignupsTableName, true, false, fmt.Sprintf("%s.%s = sqlc.arg(%s)", waitlistSignupsTableName, belongsToWaitlistColumn, belongsToWaitlistColumn)),
					buildCursorLimitClause(waitlistSignupsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetWaitlistSignupsForUser",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(waitlistSignupsTableName, true, true, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", waitlistSignupsTableName, belongsToUserColumn, belongsToUserColumn)),
					buildTotalCountSelect(waitlistSignupsTableName, true, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", waitlistSignupsTableName, belongsToUserColumn, belongsToUserColumn)),
					waitlistSignupsTableName,
					waitlistSignupsTableName, archivedAtColumn,
					waitlistSignupsTableName, belongsToUserColumn, belongsToUserColumn,
					buildFilterConditions(waitlistSignupsTableName, true, false, fmt.Sprintf("%s.%s = sqlc.arg(%s)", waitlistSignupsTableName, belongsToUserColumn, belongsToUserColumn)),
					buildCursorLimitClause(waitlistSignupsTableName),
				)),
			},
		}
	default:
		return nil
	}
}
