package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	uploadedMediaTableName = "uploaded_media"
	storagePathColumn      = "storage_path"
	mimeTypeColumn         = "mime_type"
)

func init() {
	registerTableName(uploadedMediaTableName)
}

var uploadedMediaColumns = []string{
	idColumn,
	storagePathColumn,
	mimeTypeColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	createdByUserColumn,
}

func buildUploadedMediaQueries(database string) []*Query {
	switch database {
	case postgres:
		fullSelectColumns := applyToEach(uploadedMediaColumns, func(_ int, s string) string {
			return querygen.Qualify(uploadedMediaTableName, s)
		})

		return slices.Concat(
			pgGen.StandardCRUD(uploadedMediaTableName, uploadedMediaColumns,
				querygen.WithEntity("UploadedMedia", "UploadedMedias"),
				querygen.WithImmutable(createdByUserColumn),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveUploadedMedia",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						uploadedMediaTableName,
						lastUpdatedAtColumn, querygen.NowExpression,
						archivedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "CheckUploadedMediaExistence",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS(
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
						uploadedMediaTableName, idColumn,
						uploadedMediaTableName,
						uploadedMediaTableName, archivedAtColumn,
						uploadedMediaTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUploadedMediaWithIDs",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[]);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						uploadedMediaTableName,
						uploadedMediaTableName, archivedAtColumn,
						uploadedMediaTableName, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUploadedMediaForUser",
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
						pgGen.FilterCountSelect(uploadedMediaTableName, uploadedMediaColumns, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", uploadedMediaTableName, createdByUserColumn, createdByUserColumn)),
						pgGen.TotalCountSelect(uploadedMediaTableName, uploadedMediaColumns, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", uploadedMediaTableName, createdByUserColumn, createdByUserColumn)),
						uploadedMediaTableName,
						pgGen.FilterConditions(uploadedMediaTableName, uploadedMediaColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", uploadedMediaTableName, createdByUserColumn, createdByUserColumn),
						),
						pgGen.CursorLimitClause(uploadedMediaTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
