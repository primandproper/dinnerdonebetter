-- name: CreateUserDataDisclosure :exec
INSERT INTO user_data_disclosures (
	id,
	belongs_to_user,
	expires_at
) VALUES (
	sqlc.arg(id),
	sqlc.arg(belongs_to_user),
	sqlc.arg(expires_at)
);

-- name: GetUserDataDisclosure :one
SELECT
	user_data_disclosures.id,
	user_data_disclosures.belongs_to_user,
	user_data_disclosures.status,
	user_data_disclosures.report_id,
	user_data_disclosures.expires_at,
	user_data_disclosures.created_at,
	user_data_disclosures.last_updated_at,
	user_data_disclosures.completed_at,
	user_data_disclosures.archived_at
FROM user_data_disclosures
WHERE user_data_disclosures.id = sqlc.arg(id)
	AND user_data_disclosures.archived_at IS NULL;

-- name: GetUserDataDisclosuresForUser :many
SELECT
	user_data_disclosures.id,
	user_data_disclosures.belongs_to_user,
	user_data_disclosures.status,
	user_data_disclosures.report_id,
	user_data_disclosures.expires_at,
	user_data_disclosures.created_at,
	user_data_disclosures.last_updated_at,
	user_data_disclosures.completed_at,
	user_data_disclosures.archived_at,
	(
		SELECT COUNT(user_data_disclosures.id)
		FROM user_data_disclosures
		WHERE user_data_disclosures.archived_at IS NULL
			AND
			user_data_disclosures.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND user_data_disclosures.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR user_data_disclosures.archived_at IS NULL)
			AND user_data_disclosures.belongs_to_user = sqlc.arg(user_id)
	) AS filtered_count,
	(
		SELECT COUNT(user_data_disclosures.id)
		FROM user_data_disclosures
		WHERE user_data_disclosures.archived_at IS NULL
			AND user_data_disclosures.belongs_to_user = sqlc.arg(user_id)
	) AS total_count
FROM user_data_disclosures
WHERE user_data_disclosures.archived_at IS NULL
	AND user_data_disclosures.belongs_to_user = sqlc.arg(user_id)
	AND user_data_disclosures.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND user_data_disclosures.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND user_data_disclosures.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY user_data_disclosures.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: MarkUserDataDisclosureCompleted :exec
UPDATE user_data_disclosures SET
	status = 'completed',
	report_id = sqlc.arg(report_id),
	completed_at = NOW(),
	last_updated_at = NOW()
WHERE user_data_disclosures.id = sqlc.arg(id)
	AND user_data_disclosures.archived_at IS NULL;

-- name: MarkUserDataDisclosureFailed :exec
UPDATE user_data_disclosures SET
	status = 'failed',
	last_updated_at = NOW()
WHERE user_data_disclosures.id = sqlc.arg(id)
	AND user_data_disclosures.archived_at IS NULL;

-- name: MarkUserDataDisclosureProcessing :exec
UPDATE user_data_disclosures SET
	status = 'processing',
	last_updated_at = NOW()
WHERE user_data_disclosures.id = sqlc.arg(id)
	AND user_data_disclosures.archived_at IS NULL;

-- name: ArchiveUserDataDisclosure :exec
UPDATE user_data_disclosures SET archived_at = NOW()
WHERE user_data_disclosures.id = sqlc.arg(id)
	AND user_data_disclosures.archived_at IS NULL;
