-- name: CreateServiceSetting :exec
INSERT INTO service_settings (
	id,
	name,
	type,
	description,
	default_value,
	enumeration,
	admins_only
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(type),
	sqlc.arg(description),
	sqlc.arg(default_value),
	sqlc.arg(enumeration),
	sqlc.arg(admins_only)
);

-- name: GetServiceSetting :one
SELECT
	service_settings.id,
	service_settings.name,
	service_settings.type,
	service_settings.description,
	service_settings.default_value,
	service_settings.enumeration,
	service_settings.admins_only,
	service_settings.created_at,
	service_settings.last_updated_at,
	service_settings.archived_at
FROM service_settings
WHERE service_settings.archived_at IS NULL
	AND service_settings.id = sqlc.arg(id);

-- name: CheckServiceSettingExistence :one
SELECT EXISTS (
	SELECT service_settings.id
	FROM service_settings
	WHERE service_settings.archived_at IS NULL
		AND service_settings.id = sqlc.arg(id)
);

-- name: ArchiveServiceSetting :execrows
UPDATE service_settings SET archived_at = NOW() WHERE id = sqlc.arg(id);

-- name: GetServiceSettings :many
SELECT
	service_settings.id,
	service_settings.name,
	service_settings.type,
	service_settings.description,
	service_settings.default_value,
	service_settings.enumeration,
	service_settings.admins_only,
	service_settings.created_at,
	service_settings.last_updated_at,
	service_settings.archived_at,
	(
		SELECT COUNT(service_settings.id)
		FROM service_settings
		WHERE service_settings.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND service_settings.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				service_settings.last_updated_at IS NULL
				OR service_settings.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				service_settings.last_updated_at IS NULL
				OR service_settings.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR service_settings.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(service_settings.id)
		FROM service_settings
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR service_settings.archived_at IS NULL)
	) AS total_count
FROM service_settings
WHERE service_settings.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND service_settings.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		service_settings.last_updated_at IS NULL
		OR service_settings.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		service_settings.last_updated_at IS NULL
		OR service_settings.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR service_settings.archived_at IS NULL)
	AND service_settings.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY service_settings.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: SearchForServiceSettings :many
SELECT
	service_settings.id,
	service_settings.name,
	service_settings.type,
	service_settings.description,
	service_settings.default_value,
	service_settings.enumeration,
	service_settings.admins_only,
	service_settings.created_at,
	service_settings.last_updated_at,
	service_settings.archived_at,
	(
		SELECT COUNT(service_settings.id)
		FROM service_settings
		WHERE service_settings.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND service_settings.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				service_settings.last_updated_at IS NULL
				OR service_settings.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				service_settings.last_updated_at IS NULL
				OR service_settings.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR service_settings.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(service_settings.id)
		FROM service_settings
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR service_settings.archived_at IS NULL)
	) AS total_count
FROM service_settings
WHERE service_settings.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND service_settings.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		service_settings.last_updated_at IS NULL
		OR service_settings.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		service_settings.last_updated_at IS NULL
		OR service_settings.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR service_settings.archived_at IS NULL)
	AND service_settings.name ILIKE '%' || sqlc.arg(name_query)::text || '%'
	AND service_settings.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY service_settings.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
