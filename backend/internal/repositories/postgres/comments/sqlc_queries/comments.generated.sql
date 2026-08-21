-- name: CreateComment :exec
INSERT INTO comments (
	id,
	content,
	target_type,
	referenced_id,
	parent_comment_id,
	belongs_to_user
) VALUES (
	sqlc.arg(id),
	sqlc.arg(content),
	sqlc.arg(target_type),
	sqlc.arg(referenced_id),
	sqlc.narg(parent_comment_id),
	sqlc.arg(belongs_to_user)
);

-- name: GetComment :one
SELECT
	comments.id,
	comments.content,
	comments.target_type,
	comments.referenced_id,
	comments.parent_comment_id,
	comments.belongs_to_user,
	comments.created_at,
	comments.last_updated_at,
	comments.archived_at
FROM comments
WHERE comments.archived_at IS NULL
	AND comments.id = sqlc.arg(id);

-- name: ArchiveComment :execrows
UPDATE comments SET
	archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveCommentsForReference :many
UPDATE comments SET archived_at = CURRENT_TIMESTAMP WHERE archived_at IS NULL
	AND target_type = sqlc.arg(target_type)
	AND referenced_id = sqlc.arg(referenced_id)
RETURNING id, belongs_to_user;

-- name: GetCommentsForReference :many
SELECT
	comments.id,
	comments.content,
	comments.target_type,
	comments.referenced_id,
	comments.parent_comment_id,
	comments.belongs_to_user,
	comments.created_at,
	comments.last_updated_at,
	comments.archived_at,
	(
		SELECT COUNT(comments.id)
		FROM comments
		WHERE comments.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND comments.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				comments.last_updated_at IS NULL
				OR comments.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				comments.last_updated_at IS NULL
				OR comments.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR comments.archived_at IS NULL)
			AND comments.target_type = sqlc.arg(target_type)
			AND comments.referenced_id = sqlc.arg(referenced_id)
	) AS filtered_count,
	(
		SELECT COUNT(comments.id)
		FROM comments
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR comments.archived_at IS NULL)
			AND comments.target_type = sqlc.arg(target_type)
			AND comments.referenced_id = sqlc.arg(referenced_id)
	) AS total_count
FROM comments
WHERE comments.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND comments.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		comments.last_updated_at IS NULL
		OR comments.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		comments.last_updated_at IS NULL
		OR comments.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR comments.archived_at IS NULL)
	AND comments.target_type = sqlc.arg(target_type)
	AND comments.referenced_id = sqlc.arg(referenced_id)
	AND comments.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY comments.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetCommentsForUser :many
SELECT
	comments.id,
	comments.content,
	comments.target_type,
	comments.referenced_id,
	comments.parent_comment_id,
	comments.belongs_to_user,
	comments.created_at,
	comments.last_updated_at,
	comments.archived_at,
	(
		SELECT COUNT(comments.id)
		FROM comments
		WHERE comments.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND comments.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				comments.last_updated_at IS NULL
				OR comments.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				comments.last_updated_at IS NULL
				OR comments.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR comments.archived_at IS NULL)
			AND comments.belongs_to_user = sqlc.arg(belongs_to_user)
	) AS filtered_count,
	(
		SELECT COUNT(comments.id)
		FROM comments
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR comments.archived_at IS NULL)
			AND comments.belongs_to_user = sqlc.arg(belongs_to_user)
	) AS total_count
FROM comments
WHERE comments.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND comments.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		comments.last_updated_at IS NULL
		OR comments.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		comments.last_updated_at IS NULL
		OR comments.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR comments.archived_at IS NULL)
	AND comments.belongs_to_user = sqlc.arg(belongs_to_user)
	AND comments.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY comments.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: UpdateComment :execrows
UPDATE comments SET
	content = sqlc.arg(content),
	last_updated_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_user = sqlc.arg(belongs_to_user);
