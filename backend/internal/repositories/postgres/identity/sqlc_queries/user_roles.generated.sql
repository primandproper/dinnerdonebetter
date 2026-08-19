-- name: CreateUserRole :exec
INSERT INTO user_roles (
	id,
	name,
	description,
	scope
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(description),
	sqlc.arg(scope)
);

-- name: GetUserRoleByID :one
SELECT
	user_roles.id,
	user_roles.name,
	user_roles.description,
	user_roles.scope,
	user_roles.created_at,
	user_roles.last_updated_at,
	user_roles.archived_at
FROM user_roles
WHERE user_roles.archived_at IS NULL
	AND user_roles.id = sqlc.arg(id);

-- name: ArchiveUserRole :execrows
UPDATE user_roles SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: GetUserRoleByName :one
SELECT
	user_roles.id,
	user_roles.name,
	user_roles.description,
	user_roles.scope,
	user_roles.created_at,
	user_roles.last_updated_at,
	user_roles.archived_at
FROM user_roles
WHERE user_roles.archived_at IS NULL
	AND user_roles.name = sqlc.arg(name);

-- name: GetUserRoles :many
SELECT
	user_roles.id,
	user_roles.name,
	user_roles.description,
	user_roles.scope,
	user_roles.created_at,
	user_roles.last_updated_at,
	user_roles.archived_at,
	(
		SELECT COUNT(user_roles.id)
		FROM user_roles
		WHERE user_roles.archived_at IS NULL
			AND
			user_roles.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND user_roles.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				user_roles.last_updated_at IS NULL
				OR user_roles.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				user_roles.last_updated_at IS NULL
				OR user_roles.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR user_roles.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(user_roles.id)
		FROM user_roles
		WHERE user_roles.archived_at IS NULL
	) AS total_count
FROM user_roles
WHERE user_roles.archived_at IS NULL
	AND user_roles.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND user_roles.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		user_roles.last_updated_at IS NULL
		OR user_roles.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		user_roles.last_updated_at IS NULL
		OR user_roles.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR user_roles.archived_at IS NULL)
	AND user_roles.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY user_roles.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
