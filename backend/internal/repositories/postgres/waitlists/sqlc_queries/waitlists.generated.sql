-- name: CreateWaitlist :exec
INSERT INTO waitlists (
	id,
	name,
	description,
	valid_until
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(description),
	sqlc.arg(valid_until)
);

-- name: GetWaitlist :one
SELECT
	waitlists.id,
	waitlists.name,
	waitlists.description,
	waitlists.valid_until,
	waitlists.created_at,
	waitlists.last_updated_at,
	waitlists.archived_at
FROM waitlists
WHERE waitlists.archived_at IS NULL
	AND waitlists.id = sqlc.arg(id);

-- name: UpdateWaitlist :execrows
UPDATE waitlists SET
	name = sqlc.arg(name),
	description = sqlc.arg(description),
	valid_until = sqlc.arg(valid_until),
	last_updated_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveWaitlist :execrows
UPDATE waitlists SET
	last_updated_at = CURRENT_TIMESTAMP,
	archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: CheckWaitlistExistence :one
SELECT EXISTS(
	SELECT waitlists.id
	FROM waitlists
	WHERE waitlists.archived_at IS NULL
		AND waitlists.id = sqlc.arg(id)
);

-- name: WaitlistIsNotExpired :one
SELECT EXISTS(
	SELECT waitlists.id
	FROM waitlists
	WHERE waitlists.archived_at IS NULL
		AND waitlists.id = sqlc.arg(id)
		AND waitlists.valid_until >= CURRENT_TIMESTAMP
);

-- name: GetWaitlists :many
SELECT
	waitlists.id,
	waitlists.name,
	waitlists.description,
	waitlists.valid_until,
	waitlists.created_at,
	waitlists.last_updated_at,
	waitlists.archived_at,
	(
		SELECT COUNT(waitlists.id)
		FROM waitlists
		WHERE waitlists.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND waitlists.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				waitlists.last_updated_at IS NULL
				OR waitlists.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				waitlists.last_updated_at IS NULL
				OR waitlists.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR waitlists.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(waitlists.id)
		FROM waitlists
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR waitlists.archived_at IS NULL)
	) AS total_count
FROM waitlists
WHERE waitlists.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND waitlists.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		waitlists.last_updated_at IS NULL
		OR waitlists.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		waitlists.last_updated_at IS NULL
		OR waitlists.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR waitlists.archived_at IS NULL)
	AND waitlists.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY waitlists.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetActiveWaitlists :many
SELECT
	waitlists.id,
	waitlists.name,
	waitlists.description,
	waitlists.valid_until,
	waitlists.created_at,
	waitlists.last_updated_at,
	waitlists.archived_at,
	(
		SELECT COUNT(waitlists.id)
		FROM waitlists
		WHERE waitlists.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND waitlists.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				waitlists.last_updated_at IS NULL
				OR waitlists.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				waitlists.last_updated_at IS NULL
				OR waitlists.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR waitlists.archived_at IS NULL)
			AND waitlists.valid_until >= CURRENT_TIMESTAMP
	) AS filtered_count,
	(
		SELECT COUNT(waitlists.id)
		FROM waitlists
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR waitlists.archived_at IS NULL)
			AND waitlists.valid_until >= CURRENT_TIMESTAMP
	) AS total_count
FROM waitlists
WHERE waitlists.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND waitlists.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		waitlists.last_updated_at IS NULL
		OR waitlists.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		waitlists.last_updated_at IS NULL
		OR waitlists.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR waitlists.archived_at IS NULL)
	AND waitlists.valid_until >= CURRENT_TIMESTAMP
	AND waitlists.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY waitlists.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
