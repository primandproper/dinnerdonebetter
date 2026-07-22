-- name: CheckMealInMealList :one
SELECT EXISTS (
	SELECT meal_list_items.id
	FROM meal_list_items
	WHERE meal_list_items.archived_at IS NULL
		AND meal_list_items.belongs_to_meal_list = sqlc.arg(belongs_to_meal_list)
		AND meal_list_items.meal_id = sqlc.arg(meal_id)
);

-- name: ArchiveMealListItem :execrows
UPDATE meal_list_items SET archived_at = NOW() WHERE archived_at IS NULL AND belongs_to_meal_list = sqlc.arg(belongs_to_meal_list) AND id = sqlc.arg(id);

-- name: CreateMealListItem :exec
INSERT INTO meal_list_items (
	id,
	meal_id,
	notes,
	belongs_to_meal_list
) VALUES (
	sqlc.arg(id),
	sqlc.arg(meal_id),
	sqlc.arg(notes),
	sqlc.arg(belongs_to_meal_list)
);

-- name: GetMealListItems :many
SELECT
	meal_list_items.id,
	meal_list_items.meal_id,
	meal_list_items.notes,
	meal_list_items.created_at,
	meal_list_items.last_updated_at,
	meal_list_items.archived_at,
	meal_list_items.belongs_to_meal_list,
	(
		SELECT COUNT(meal_list_items.id)
		FROM meal_list_items
		WHERE meal_list_items.archived_at IS NULL
			AND
			meal_list_items.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND meal_list_items.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				meal_list_items.last_updated_at IS NULL
				OR meal_list_items.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				meal_list_items.last_updated_at IS NULL
				OR meal_list_items.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR meal_list_items.archived_at IS NULL)
			AND meal_list_items.belongs_to_meal_list = sqlc.arg(meal_list_id)
			AND EXISTS (SELECT 1 FROM meal_lists WHERE meal_lists.id = meal_list_items.belongs_to_meal_list AND meal_lists.belongs_to_user = sqlc.arg(belongs_to_user))
	) AS filtered_count,
	(
		SELECT COUNT(meal_list_items.id)
		FROM meal_list_items
		WHERE meal_list_items.archived_at IS NULL
	) AS total_count
FROM meal_list_items
WHERE meal_list_items.archived_at IS NULL
	AND meal_list_items.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND meal_list_items.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		meal_list_items.last_updated_at IS NULL
		OR meal_list_items.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		meal_list_items.last_updated_at IS NULL
		OR meal_list_items.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND meal_list_items.belongs_to_meal_list = sqlc.arg(meal_list_id)
	AND EXISTS (SELECT 1 FROM meal_lists WHERE meal_lists.id = meal_list_items.belongs_to_meal_list AND meal_lists.belongs_to_user = sqlc.arg(belongs_to_user))
	AND meal_list_items.id > COALESCE(sqlc.narg(cursor), '')
	AND meal_list_items.belongs_to_meal_list = sqlc.arg(meal_list_id)
ORDER BY meal_list_items.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: UpdateMealListItem :execrows
UPDATE meal_list_items SET
	meal_id = sqlc.arg(meal_id),
	notes = sqlc.arg(notes),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND belongs_to_meal_list = sqlc.arg(belongs_to_meal_list)
	AND id = sqlc.arg(id);
