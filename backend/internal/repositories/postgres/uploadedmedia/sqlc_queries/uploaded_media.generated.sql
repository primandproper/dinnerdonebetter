-- name: CreateUploadedMedia :exec
INSERT INTO uploaded_media (
	id,
	storage_path,
	mime_type,
	created_by_user
) VALUES (
	sqlc.arg(id),
	sqlc.arg(storage_path),
	sqlc.arg(mime_type),
	sqlc.arg(created_by_user)
);

-- name: GetUploadedMedia :one
SELECT
	uploaded_media.id,
	uploaded_media.storage_path,
	uploaded_media.mime_type,
	uploaded_media.created_at,
	uploaded_media.last_updated_at,
	uploaded_media.archived_at,
	uploaded_media.created_by_user
FROM uploaded_media
WHERE uploaded_media.archived_at IS NULL
	AND uploaded_media.id = sqlc.arg(id);

-- name: UpdateUploadedMedia :execrows
UPDATE uploaded_media SET
	storage_path = sqlc.arg(storage_path),
	mime_type = sqlc.arg(mime_type),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveUploadedMedia :execrows
UPDATE uploaded_media SET
	last_updated_at = NOW(),
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: CheckUploadedMediaExistence :one
SELECT EXISTS(
	SELECT uploaded_media.id
	FROM uploaded_media
	WHERE uploaded_media.archived_at IS NULL
		AND uploaded_media.id = sqlc.arg(id)
);

-- name: GetUploadedMediaWithIDs :many
SELECT
	uploaded_media.id,
	uploaded_media.storage_path,
	uploaded_media.mime_type,
	uploaded_media.created_at,
	uploaded_media.last_updated_at,
	uploaded_media.archived_at,
	uploaded_media.created_by_user
FROM uploaded_media
WHERE uploaded_media.archived_at IS NULL
	AND uploaded_media.id = ANY(sqlc.arg(ids)::text[]);

-- name: GetUploadedMediaForUser :many
SELECT
	uploaded_media.id,
	uploaded_media.storage_path,
	uploaded_media.mime_type,
	uploaded_media.created_at,
	uploaded_media.last_updated_at,
	uploaded_media.archived_at,
	uploaded_media.created_by_user,
	(
		SELECT COUNT(uploaded_media.id)
		FROM uploaded_media
		WHERE uploaded_media.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND uploaded_media.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				uploaded_media.last_updated_at IS NULL
				OR uploaded_media.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				uploaded_media.last_updated_at IS NULL
				OR uploaded_media.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR uploaded_media.archived_at IS NULL)
			AND uploaded_media.created_by_user = sqlc.arg(created_by_user)
	) AS filtered_count,
	(
		SELECT COUNT(uploaded_media.id)
		FROM uploaded_media
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR uploaded_media.archived_at IS NULL)
			AND uploaded_media.created_by_user = sqlc.arg(created_by_user)
	) AS total_count
FROM uploaded_media
WHERE uploaded_media.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND uploaded_media.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		uploaded_media.last_updated_at IS NULL
		OR uploaded_media.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		uploaded_media.last_updated_at IS NULL
		OR uploaded_media.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR uploaded_media.archived_at IS NULL)
	AND uploaded_media.created_by_user = sqlc.arg(created_by_user)
	AND uploaded_media.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY uploaded_media.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
