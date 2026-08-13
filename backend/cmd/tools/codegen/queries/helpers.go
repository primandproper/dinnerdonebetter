package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	idColumn               = "id"
	nameColumn             = "name"
	pluralNameColumn       = "plural_name"
	notesColumn            = "notes"
	descriptionColumn      = "description"
	iconPathColumn         = "icon_path"
	slugColumn             = "slug"
	createdAtColumn        = "created_at"
	expiresAtColumn        = "expires_at"
	lastUpdatedAtColumn    = "last_updated_at"
	archivedAtColumn       = "archived_at"
	lastIndexedAtColumn    = "last_indexed_at"
	belongsToAccountColumn = "belongs_to_account"
	belongsToUserColumn    = "belongs_to_user"

	includeArchivedArg = "include_archived"
	cursorArg          = "cursor"
	limitArg           = "result_limit"

	currentTimeExpression = "NOW()"
)

func applyToEach[T comparable](x []T, f func(int, T) T) []T {
	output := []T{}

	for i, v := range x {
		output = append(output, f(i, v))
	}

	return output
}

func buildRawQuery(builder *builq.Builder) string {
	query, _, err := builder.Build()
	if err != nil {
		panic(err)
	}

	return query
}

func filterForInsert(columns []string, exceptions ...string) []string {
	return filterFromSlice(columns, append([]string{archivedAtColumn, createdAtColumn, lastUpdatedAtColumn, lastIndexedAtColumn}, exceptions...)...)
}

func filterForUpdate(columns []string, exceptions ...string) []string {
	return filterForInsert(columns, append(exceptions, idColumn)...)
}

func fullColumnName(tableName, columnName string) string {
	return fmt.Sprintf("%s.%s", tableName, columnName)
}

func filterFromSlice(slice []string, filtered ...string) []string {
	output := []string{}

	for _, s := range slice {
		if !slices.Contains(filtered, s) {
			output = append(output, s)
		}
	}

	return output
}

func mergeColumns(columns1, columns2 []string, indexToInsertSecondSet int) []string {
	output := []string{}

	for i, col1 := range columns1 {
		if i == indexToInsertSecondSet {
			output = append(output, columns2...)
		}
		output = append(output, col1)
	}

	return output
}

func buildFilterConditions(tableName string, withUpdateColumn, withArchivedAtColumn bool, conditions ...string) string {
	updateAddendum := ""
	if withUpdateColumn {
		updateAddendum = fmt.Sprintf("\n\t%s", strings.TrimSpace(buildRawQuery((&builq.Builder{}).Addf(`
	AND (
		%s.%s IS NULL
		OR %s.%s > COALESCE(sqlc.narg(updated_after), (SELECT %s - '999 years'::INTERVAL))
	)
	AND (
		%s.%s IS NULL
		OR %s.%s < COALESCE(sqlc.narg(updated_before), (SELECT %s + '999 years'::INTERVAL))
	)
		`,
			tableName,
			lastUpdatedAtColumn,
			tableName,
			lastUpdatedAtColumn,
			currentTimeExpression,
			tableName,
			lastUpdatedAtColumn,
			tableName,
			lastUpdatedAtColumn,
			currentTimeExpression,
		))))
	}

	archivedAddendum := ""
	if withArchivedAtColumn {
		archivedAddendum = fmt.Sprintf("\n\t\t\tAND (NOT COALESCE(sqlc.narg(%s), false)::boolean OR %s.%s IS NULL)", includeArchivedArg, tableName, archivedAtColumn)
	}

	var allConditions strings.Builder
	for _, condition := range conditions {
		if _, err := fmt.Fprintf(&allConditions, "\n\tAND %s", condition); err != nil {
			panic(err)
		}
	}

	// Add cursor-based pagination condition
	cursorCondition := fmt.Sprintf("\n\t%s", buildCursorCondition(tableName))

	rv := strings.TrimSpace(buildRawQuery((&builq.Builder{}).Addf(`AND %s.%s > COALESCE(sqlc.narg(created_after), (SELECT %s - '999 years'::INTERVAL))
	AND %s.%s < COALESCE(sqlc.narg(created_before), (SELECT %s + '999 years'::INTERVAL))%s%s%s%s`,
		tableName,
		createdAtColumn,
		currentTimeExpression,
		tableName,
		createdAtColumn,
		currentTimeExpression,
		updateAddendum,
		archivedAddendum,
		allConditions.String(),
		cursorCondition,
	)))

	return rv
}

func buildFilterCountSelect(tableName string, withUpdateColumn, withArchivedAtColumn bool, joins []string, conditions ...string) string {
	updateAddendum := ""
	if withUpdateColumn {
		updateAddendum = fmt.Sprintf("\n\t\t\t%s", strings.TrimSpace(buildRawQuery((&builq.Builder{}).Addf(`
			AND (
				%s.%s IS NULL
				OR %s.%s > COALESCE(sqlc.narg(updated_after), (SELECT %s - '999 years'::INTERVAL))
			)
			AND (
				%s.%s IS NULL
				OR %s.%s < COALESCE(sqlc.narg(updated_before), (SELECT %s + '999 years'::INTERVAL))
			)
		`,
			tableName, lastUpdatedAtColumn,
			tableName, lastUpdatedAtColumn, currentTimeExpression,
			tableName, lastUpdatedAtColumn,
			tableName, lastUpdatedAtColumn, currentTimeExpression,
		))))
	}

	archivedAddendum := ""
	if withArchivedAtColumn {
		archivedAddendum = fmt.Sprintf("\n\t\t\tAND (NOT COALESCE(sqlc.narg(%s), false)::boolean OR %s.%s IS NULL)", includeArchivedArg, tableName, archivedAtColumn)
	}

	var allConditions strings.Builder
	for _, condition := range conditions {
		if _, err := fmt.Fprintf(&allConditions, "\n\t\t\tAND %s", condition); err != nil {
			panic(err)
		}
	}

	archivedAtAddendum := "\n\t\tWHERE"
	if withArchivedAtColumn {
		archivedAtAddendum = fmt.Sprintf("\n\t\tWHERE %s.%s IS NULL\n\t\t\tAND", tableName, archivedAtColumn)
	}

	joinStmnt := ""
	if len(joins) > 0 {
		joinStmnt = fmt.Sprintf("\n\t\tJOIN %s", strings.Join(joins, "\n\tJOIN "))
	}

	return strings.TrimSpace(buildRawQuery((&builq.Builder{}).Addf(`(
		SELECT COUNT(%s.%s)
		FROM %s%s%s 
			%s.%s > COALESCE(sqlc.narg(created_after), (SELECT %s - '999 years'::INTERVAL))
			AND %s.%s < COALESCE(sqlc.narg(created_before), (SELECT %s + '999 years'::INTERVAL))%s%s%s
	) AS filtered_count`,
		tableName, idColumn,
		tableName, joinStmnt,
		archivedAtAddendum, tableName, createdAtColumn, currentTimeExpression,
		tableName, createdAtColumn, currentTimeExpression,
		updateAddendum,
		archivedAddendum,
		allConditions.String(),
	)))
}

func buildTotalCountSelect(tableName string, withArchivedAtColumn bool, joins []string, conditions ...string) string {
	var allConditons strings.Builder
	for i, condition := range conditions {
		prefix := "AND "
		if !withArchivedAtColumn && i == 0 {
			prefix = ""
		}
		if _, err := fmt.Fprintf(&allConditons, "\n\t\t\t%s%s", prefix, strings.TrimSpace(condition)); err != nil {
			panic(err)
		}
	}

	archivedAtAddendum := "WHERE"
	if withArchivedAtColumn {
		archivedAtAddendum = fmt.Sprintf("WHERE %s.%s IS NULL", tableName, archivedAtColumn)
	}

	joinStmnt := ""
	if len(joins) > 0 {
		joinStmnt = fmt.Sprintf("\n\t\tJOIN %s", strings.Join(joins, "\n\tJOIN "))
	}

	return strings.TrimSpace(buildRawQuery((&builq.Builder{}).Addf(`(
		SELECT COUNT(%s.%s)
		FROM %s%s
		%s%s
	) AS total_count`,
		tableName, idColumn,
		tableName,
		joinStmnt,
		archivedAtAddendum,
		allConditons.String(),
	)))
}

func buildILIKEForArgument(argumentName string) string {
	return fmt.Sprintf(`ILIKE '%%' || sqlc.arg(%s)::text || '%%'`, argumentName)
}

type joinStatement struct {
	joinTarget   string
	targetColumn string
	onTable      string
	onColumn     string
}

func buildJoinStatement(js joinStatement) string {
	return fmt.Sprintf("JOIN %s ON %s.%s=%s.%s", js.joinTarget, js.onTable, js.onColumn, js.joinTarget, js.targetColumn)
}

// buildCursorCondition creates a WHERE clause for cursor-based pagination.
// Since xid is sortable by time, we can use simple string comparison.
func buildCursorCondition(tableName string) string {
	return fmt.Sprintf("AND %s.%s > COALESCE(sqlc.narg(%s), '')", tableName, idColumn, cursorArg)
}

// buildCursorLimitClause creates the ORDER BY and LIMIT clause for cursor-based pagination.
// This provides a consistent ordering by ID (which is sortable with xid) and applies the limit.
func buildCursorLimitClause(tableName string) string {
	return fmt.Sprintf("ORDER BY %s.%s ASC\nLIMIT COALESCE(sqlc.narg(%s), 50)", tableName, idColumn, limitArg)
}

// buildReindexScanQuery builds the keyset walk a search reindex reads its source through.
//
// It returns IDs rather than rows on purpose. searchsync.Scanner and searchsync.Fetcher both
// have to produce the same document for the same row, and the cheapest way to guarantee that
// is to have one of them call the other: the scan names the next page of IDs and the fetch —
// the same one the change feed uses — turns them into documents. Selecting rows here would be
// a second row-to-document transform, and two transforms that are supposed to agree are two
// transforms that can drift.
//
// The ordering is a byte comparison, COLLATE "C", not the database's default collation.
// searchsync requires ascending byte order because the pruning half of a reindex merges this
// stream against the index's own stream of IDs, and Postgres's en_US.UTF-8 sorts
// case-insensitively and ignores punctuation — a different order. Two ordered streams merged
// under disagreeing orders do not fail; they conclude that live documents are absent from the
// source and delete them. The Reindexer verifies the order it is given for the same reason.
func buildReindexScanQuery(tableName string) string {
	return buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s COLLATE "C" > sqlc.arg(%s)
ORDER BY %s.%s COLLATE "C"
LIMIT COALESCE(sqlc.narg(%s), 50);`,
		tableName, idColumn,
		tableName,
		tableName, archivedAtColumn,
		tableName, idColumn, cursorArg,
		tableName, idColumn,
		limitArg,
	))
}

// buildCursorPaginationFragment creates a complete cursor-based pagination fragment
// for use in queries that don't already have buildFilterConditions.
func buildCursorPaginationFragment(tableName string) string {
	return fmt.Sprintf("%s\n%s", buildCursorCondition(tableName), buildCursorLimitClause(tableName))
}
