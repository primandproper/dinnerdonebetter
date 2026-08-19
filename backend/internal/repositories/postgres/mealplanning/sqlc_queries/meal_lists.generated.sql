-- name: CreateMealList :exec
INSERT INTO meal_lists (
	id,
	name,
	description,
	belongs_to_user
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(description),
	sqlc.arg(belongs_to_user)
);

-- name: UpdateMealList :execrows
UPDATE meal_lists SET
	name = sqlc.arg(name),
	description = sqlc.arg(description),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_user = sqlc.arg(belongs_to_user);

-- name: ArchiveMealList :execrows
UPDATE meal_lists SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_user = sqlc.arg(belongs_to_user);

-- name: GetMealLists :many
SELECT
	meal_lists.id,
	meal_lists.name,
	meal_list_items.id as meal_list_item_id,
	meal_list_items.meal_id as meal_list_item_meal_id,
	meal_list_items.notes as meal_list_item_notes,
	meal_list_items.created_at as meal_list_item_created_at,
	meal_list_items.last_updated_at as meal_list_item_last_updated_at,
	meal_list_items.archived_at as meal_list_item_archived_at,
	meal_list_items.belongs_to_meal_list as meal_list_item_belongs_to_meal_list,
	meal_lists.description,
	meal_lists.created_at,
	meal_lists.last_updated_at,
	meal_lists.archived_at,
	meal_lists.belongs_to_user,
	(
		SELECT COUNT(meal_lists.id)
		FROM meal_lists
		WHERE meal_lists.archived_at IS NULL
			AND
			meal_lists.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND meal_lists.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				meal_lists.last_updated_at IS NULL
				OR meal_lists.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				meal_lists.last_updated_at IS NULL
				OR meal_lists.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR meal_lists.archived_at IS NULL)
			AND meal_lists.belongs_to_user = sqlc.arg(belongs_to_user)
	) AS filtered_count,
	(
		SELECT COUNT(meal_lists.id)
		FROM meal_lists
		WHERE meal_lists.archived_at IS NULL
			AND meal_lists.belongs_to_user = sqlc.arg(belongs_to_user)
	) AS total_count
FROM meal_lists
	LEFT JOIN meal_list_items ON meal_list_items.belongs_to_meal_list = meal_lists.id AND meal_list_items.archived_at IS NULL
	WHERE meal_lists.archived_at IS NULL
	AND meal_lists.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND meal_lists.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		meal_lists.last_updated_at IS NULL
		OR meal_lists.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		meal_lists.last_updated_at IS NULL
		OR meal_lists.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND meal_lists.belongs_to_user = sqlc.arg(belongs_to_user)
	AND meal_lists.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY meal_lists.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
