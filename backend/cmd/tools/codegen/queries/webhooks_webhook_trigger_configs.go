package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	webhookTriggerConfigsTableName = "webhook_trigger_configs"
	belongsToWebhookColumn         = "belongs_to_webhook"
	triggerEventColumn             = "trigger_event"
)

func init() {
	registerTableName(webhookTriggerConfigsTableName)
}

var (
	webhookTriggerConfigsColumns = []string{
		idColumn,
		triggerEventColumn,
		belongsToWebhookColumn,
		createdAtColumn,
		archivedAtColumn,
	}
)

func buildWebhookTriggerConfigsQueries(database string) []*Query {
	switch database {
	case postgres:
		insertColumns := filterForInsert(webhookTriggerConfigsColumns)

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "CreateWebhookTriggerConfig",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					webhookTriggerConfigsTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(i int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveWebhookTriggerConfig",
					Type: ExecRowsType,
				},
				// Scoped to the account through the owning webhook, not only to the
				// webhook. The account is checked at the service layer too, but a query
				// that will happily archive any account's trigger config given its two
				// IDs is one call site away from being an IDOR, and the join costs an
				// index lookup on a primary key.
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s IN (
		SELECT %s FROM %s
		WHERE %s = sqlc.arg(%s)
			AND %s = sqlc.arg(%s)
			AND %s IS NULL
	);`,
					webhookTriggerConfigsTableName,
					archivedAtColumn, currentTimeExpression,
					fullColumnName(webhookTriggerConfigsTableName, archivedAtColumn),
					fullColumnName(webhookTriggerConfigsTableName, idColumn), idColumn,
					fullColumnName(webhookTriggerConfigsTableName, belongsToWebhookColumn),
					// Fully qualified inside the subquery: the outer UPDATE and the inner
					// SELECT both have an id and an archived_at, and an unqualified
					// reference to either is ambiguous.
					fullColumnName(webhooksTableName, idColumn), webhooksTableName,
					fullColumnName(webhooksTableName, idColumn), belongsToWebhookColumn,
					fullColumnName(webhooksTableName, belongsToAccountColumn), belongsToAccountColumn,
					fullColumnName(webhooksTableName, archivedAtColumn),
				)),
			},
		}
	default:
		return nil
	}
}
