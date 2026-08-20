package main

import (
	"fmt"
	"slices"

	"github.com/primandproper/platform-go/v12/database/querygen"

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

	// idsArg is querygen's name for the id-list argument of a bulk stamp, taken from
	// there rather than spelled again so the hand-written stamps below match the
	// generated ones exactly.
	idsArg = querygen.IDsArg
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

func buildILIKEForArgument(argumentName string) string {
	return fmt.Sprintf(`ILIKE '%%' || sqlc.arg(%s)::text || '%%'`, argumentName)
}
