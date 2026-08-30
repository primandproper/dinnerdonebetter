package main

import (
	"slices"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	webhookTriggerConfigsTableName = "webhook_trigger_configs"
	belongsToWebhookColumn         = "belongs_to_webhook"
	triggerEventColumn             = "trigger_event"
)

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

		return slices.Concat(
			pgGen.StandardCRUD(webhookTriggerConfigsTableName, webhookTriggerConfigsColumns,
				querygen.WithEntity("WebhookTriggerConfig", "WebhookTriggerConfigs"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
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
						archivedAtColumn, querygen.NowExpression,
						querygen.Qualify(webhookTriggerConfigsTableName, archivedAtColumn),
						querygen.Qualify(webhookTriggerConfigsTableName, idColumn), idColumn,
						querygen.Qualify(webhookTriggerConfigsTableName, belongsToWebhookColumn),
						// Fully qualified inside the subquery: the outer UPDATE and the inner
						// SELECT both have an id and an archived_at, and an unqualified
						// reference to either is ambiguous.
						querygen.Qualify(webhooksTableName, idColumn), webhooksTableName,
						querygen.Qualify(webhooksTableName, idColumn), belongsToWebhookColumn,
						querygen.Qualify(webhooksTableName, belongsToAccountColumn), belongsToAccountColumn,
						querygen.Qualify(webhooksTableName, archivedAtColumn),
					)),
				},
			},
		)
	default:
		return nil
	}
}
