package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

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
		fullSelectColumns := applyToEach(waitlistSignupColumns, func(_ int, s string) string {
			return querygen.Qualify(waitlistSignupsTableName, s)
		})

		return slices.Concat(
			querygen.StandardCRUD(waitlistSignupsTableName, waitlistSignupColumns,
				querygen.WithEntity("WaitlistSignup", "WaitlistSignups"),
				querygen.WithImmutable(belongsToWaitlistColumn, belongsToUserColumn, belongsToAccountColumn),
				querygen.WithQueryName(querygen.GetQuery, "GetWaitlistSignupByID"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.ListQuery),
			),
			[]*Query{
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
						lastUpdatedAtColumn, querygen.NowExpression,
						archivedAtColumn, querygen.NowExpression,
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
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						querygen.FilterCountSelect(waitlistSignupsTableName, waitlistSignupColumns, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", waitlistSignupsTableName, belongsToWaitlistColumn, belongsToWaitlistColumn)),
						querygen.TotalCountSelect(waitlistSignupsTableName, waitlistSignupColumns, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", waitlistSignupsTableName, belongsToWaitlistColumn, belongsToWaitlistColumn)),
						waitlistSignupsTableName,
						querygen.FilterConditions(waitlistSignupsTableName, waitlistSignupColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", waitlistSignupsTableName, belongsToWaitlistColumn, belongsToWaitlistColumn),
						),
						querygen.CursorLimitClause(waitlistSignupsTableName),
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
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						querygen.FilterCountSelect(waitlistSignupsTableName, waitlistSignupColumns, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", waitlistSignupsTableName, belongsToUserColumn, belongsToUserColumn)),
						querygen.TotalCountSelect(waitlistSignupsTableName, waitlistSignupColumns, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", waitlistSignupsTableName, belongsToUserColumn, belongsToUserColumn)),
						waitlistSignupsTableName,
						querygen.FilterConditions(waitlistSignupsTableName, waitlistSignupColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", waitlistSignupsTableName, belongsToUserColumn, belongsToUserColumn),
						),
						querygen.CursorLimitClause(waitlistSignupsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
