-- name: CreateWaitlistSignup :exec
INSERT INTO waitlist_signups (
	id,
	notes,
	belongs_to_waitlist,
	belongs_to_user,
	belongs_to_account
) VALUES (
	sqlc.arg(id),
	sqlc.arg(notes),
	sqlc.arg(belongs_to_waitlist),
	sqlc.arg(belongs_to_user),
	sqlc.arg(belongs_to_account)
);

-- name: GetWaitlistSignupByID :one
SELECT
	waitlist_signups.id,
	waitlist_signups.notes,
	waitlist_signups.belongs_to_waitlist,
	waitlist_signups.created_at,
	waitlist_signups.last_updated_at,
	waitlist_signups.archived_at,
	waitlist_signups.belongs_to_user,
	waitlist_signups.belongs_to_account
FROM waitlist_signups
WHERE waitlist_signups.archived_at IS NULL
	AND waitlist_signups.id = sqlc.arg(id);

-- name: UpdateWaitlistSignup :execrows
UPDATE waitlist_signups SET
	notes = sqlc.arg(notes),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveWaitlistSignup :execrows
UPDATE waitlist_signups SET
	last_updated_at = NOW(),
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: GetWaitlistSignup :one
SELECT
	waitlist_signups.id,
	waitlist_signups.notes,
	waitlist_signups.belongs_to_waitlist,
	waitlist_signups.created_at,
	waitlist_signups.last_updated_at,
	waitlist_signups.archived_at,
	waitlist_signups.belongs_to_user,
	waitlist_signups.belongs_to_account
FROM waitlist_signups
WHERE waitlist_signups.archived_at IS NULL
	AND waitlist_signups.id = sqlc.arg(id)
	AND waitlist_signups.belongs_to_waitlist = sqlc.arg(belongs_to_waitlist);

-- name: CheckWaitlistSignupExistence :one
SELECT EXISTS(
	SELECT waitlist_signups.id
	FROM waitlist_signups
	WHERE waitlist_signups.archived_at IS NULL
		AND waitlist_signups.id = sqlc.arg(id)
		AND waitlist_signups.belongs_to_waitlist = sqlc.arg(belongs_to_waitlist)
);

-- name: GetWaitlistSignupsForWaitlist :many
SELECT
	waitlist_signups.id,
	waitlist_signups.notes,
	waitlist_signups.belongs_to_waitlist,
	waitlist_signups.created_at,
	waitlist_signups.last_updated_at,
	waitlist_signups.archived_at,
	waitlist_signups.belongs_to_user,
	waitlist_signups.belongs_to_account,
	(
		SELECT COUNT(waitlist_signups.id)
		FROM waitlist_signups
		WHERE waitlist_signups.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND waitlist_signups.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				waitlist_signups.last_updated_at IS NULL
				OR waitlist_signups.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				waitlist_signups.last_updated_at IS NULL
				OR waitlist_signups.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR waitlist_signups.archived_at IS NULL)
			AND waitlist_signups.belongs_to_waitlist = sqlc.arg(belongs_to_waitlist)
	) AS filtered_count,
	(
		SELECT COUNT(waitlist_signups.id)
		FROM waitlist_signups
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR waitlist_signups.archived_at IS NULL)
			AND waitlist_signups.belongs_to_waitlist = sqlc.arg(belongs_to_waitlist)
	) AS total_count
FROM waitlist_signups
WHERE waitlist_signups.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND waitlist_signups.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		waitlist_signups.last_updated_at IS NULL
		OR waitlist_signups.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		waitlist_signups.last_updated_at IS NULL
		OR waitlist_signups.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR waitlist_signups.archived_at IS NULL)
	AND waitlist_signups.belongs_to_waitlist = sqlc.arg(belongs_to_waitlist)
	AND waitlist_signups.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY waitlist_signups.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetWaitlistSignupsForUser :many
SELECT
	waitlist_signups.id,
	waitlist_signups.notes,
	waitlist_signups.belongs_to_waitlist,
	waitlist_signups.created_at,
	waitlist_signups.last_updated_at,
	waitlist_signups.archived_at,
	waitlist_signups.belongs_to_user,
	waitlist_signups.belongs_to_account,
	(
		SELECT COUNT(waitlist_signups.id)
		FROM waitlist_signups
		WHERE waitlist_signups.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND waitlist_signups.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				waitlist_signups.last_updated_at IS NULL
				OR waitlist_signups.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				waitlist_signups.last_updated_at IS NULL
				OR waitlist_signups.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR waitlist_signups.archived_at IS NULL)
			AND waitlist_signups.belongs_to_user = sqlc.arg(belongs_to_user)
	) AS filtered_count,
	(
		SELECT COUNT(waitlist_signups.id)
		FROM waitlist_signups
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR waitlist_signups.archived_at IS NULL)
			AND waitlist_signups.belongs_to_user = sqlc.arg(belongs_to_user)
	) AS total_count
FROM waitlist_signups
WHERE waitlist_signups.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND waitlist_signups.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		waitlist_signups.last_updated_at IS NULL
		OR waitlist_signups.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		waitlist_signups.last_updated_at IS NULL
		OR waitlist_signups.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR waitlist_signups.archived_at IS NULL)
	AND waitlist_signups.belongs_to_user = sqlc.arg(belongs_to_user)
	AND waitlist_signups.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY waitlist_signups.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
