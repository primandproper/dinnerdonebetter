package main

// destroyAllData truncates every table the schema has, discovered when the query runs
// rather than listed when it is generated.
//
// The list used to be assembled by registerTableName calls sprinkled through the query
// builders, which made "is this table truncated?" a question about whether somebody wrote
// a query builder for it rather than about whether the table exists. Eleven tables had
// fallen out by the time anyone checked: the media and image tables nobody ever wrote a
// builder for, and the ones whose queries come from platform rather than from this
// generator — sessions, webauthn_credentials, webauthn_sessions, user_device_tokens, and
// every webhook, audit, data privacy, metering, operations, outbox, saga and
// authorization server table. A domain moving onto platform's resources kit drops out the
// same way, silently, the moment it stops generating queries here — and the symptom is an
// unrelated test failing later, or passing for the wrong reason on auth state a previous
// test left behind.
//
// pg_tables is the schema itself, so it cannot drift from the migrations. goose's version
// table is the one exclusion: truncating it would leave a fully migrated database claiming
// it had never been migrated.
const destroyAllData = `DO $$
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
$$;`

func buildMaintenanceQueries(database string) []*Query {
	switch database {
	case postgres:
		queries := []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "DestroyAllData",
					Type: ExecType,
				},
				Content: destroyAllData,
			},
		}
		return append(queries, buildQueueTestMessagesQueries(database)...)
	default:
		return nil
	}
}
