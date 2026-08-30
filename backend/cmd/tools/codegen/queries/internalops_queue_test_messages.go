package main

import (
	"fmt"
)

const queueTestMessagesTableName = "queue_test_messages"

func buildQueueTestMessagesQueries(database string) []*Query {
	switch database {
	case postgres:
		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "CreateQueueTestMessage",
					Type: ExecType,
				},
				Content: fmt.Sprintf(`INSERT INTO %s (id, queue_name) VALUES (sqlc.arg(id), sqlc.arg(queue_name));`, queueTestMessagesTableName),
			},
			{
				Annotation: QueryAnnotation{
					Name: "AcknowledgeQueueTestMessage",
					Type: ExecType,
				},
				Content: fmt.Sprintf(`UPDATE %s SET acknowledged_at = NOW() WHERE id = sqlc.arg(id);`, queueTestMessagesTableName),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetQueueTestMessage",
					Type: OneType,
				},
				Content: fmt.Sprintf(`SELECT id, queue_name, created_at, acknowledged_at FROM %s WHERE id = sqlc.arg(id);`, queueTestMessagesTableName),
			},
			{
				Annotation: QueryAnnotation{
					Name: "PruneQueueTestMessages",
					Type: ExecType,
				},
				Content: fmt.Sprintf(`DELETE FROM %s AS qtm
WHERE qtm.queue_name = sqlc.arg(queue_name)
  AND qtm.id NOT IN (
      SELECT keep.id FROM %s AS keep
      WHERE keep.queue_name = sqlc.arg(queue_name)
      ORDER BY keep.created_at DESC
      LIMIT 100
  );`, queueTestMessagesTableName, queueTestMessagesTableName),
			},
		}
	default:
		return nil
	}
}
