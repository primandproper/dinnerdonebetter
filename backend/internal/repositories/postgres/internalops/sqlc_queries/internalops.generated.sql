-- name: DestroyAllData :exec
DO $$
DECLARE
	truncation TEXT;
BEGIN
	SELECT 'TRUNCATE ' || string_agg(format('%I.%I', schemaname, tablename), ', ') || ' CASCADE'
	INTO truncation
	FROM pg_tables
	WHERE schemaname = 'public'
	  AND tablename <> 'goose_db_version';

	IF truncation IS NOT NULL THEN
		EXECUTE truncation;
	END IF;
END
$$;

-- name: CreateQueueTestMessage :exec
INSERT INTO queue_test_messages (id, queue_name) VALUES (sqlc.arg(id), sqlc.arg(queue_name));

-- name: AcknowledgeQueueTestMessage :exec
UPDATE queue_test_messages SET acknowledged_at = NOW() WHERE id = sqlc.arg(id);

-- name: GetQueueTestMessage :one
SELECT id, queue_name, created_at, acknowledged_at FROM queue_test_messages WHERE id = sqlc.arg(id);

-- name: PruneQueueTestMessages :exec
DELETE FROM queue_test_messages AS qtm
WHERE qtm.queue_name = sqlc.arg(queue_name)
  AND qtm.id NOT IN (
      SELECT keep.id FROM queue_test_messages AS keep
      WHERE keep.queue_name = sqlc.arg(queue_name)
      ORDER BY keep.created_at DESC
      LIMIT 100
  );
