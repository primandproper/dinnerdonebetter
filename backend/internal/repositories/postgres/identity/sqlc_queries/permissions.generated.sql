-- name: CreatePermission :exec
INSERT INTO permissions (
	id,
	name,
	description
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(description)
);

-- name: GetPermissionByID :one
SELECT
	permissions.id,
	permissions.name,
	permissions.description,
	permissions.created_at,
	permissions.last_updated_at,
	permissions.archived_at
FROM permissions
WHERE permissions.archived_at IS NULL
	AND permissions.id = sqlc.arg(id);

-- name: ArchivePermission :execrows
UPDATE permissions SET
	archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: GetPermissionByName :one
SELECT
	permissions.id,
	permissions.name,
	permissions.description,
	permissions.created_at,
	permissions.last_updated_at,
	permissions.archived_at
FROM permissions
WHERE permissions.archived_at IS NULL
	AND permissions.name = sqlc.arg(name);

-- name: GetPermissions :many
SELECT
	permissions.id,
	permissions.name,
	permissions.description,
	permissions.created_at,
	permissions.last_updated_at,
	permissions.archived_at,
	(
		SELECT COUNT(permissions.id)
		FROM permissions
		WHERE permissions.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND permissions.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				permissions.last_updated_at IS NULL
				OR permissions.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				permissions.last_updated_at IS NULL
				OR permissions.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR permissions.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(permissions.id)
		FROM permissions
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR permissions.archived_at IS NULL)
	) AS total_count
FROM permissions
WHERE permissions.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND permissions.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		permissions.last_updated_at IS NULL
		OR permissions.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		permissions.last_updated_at IS NULL
		OR permissions.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR permissions.archived_at IS NULL)
	AND permissions.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY permissions.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
