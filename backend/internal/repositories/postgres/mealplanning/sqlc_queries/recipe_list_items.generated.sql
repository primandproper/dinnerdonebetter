-- name: ArchiveRecipeListItem :execrows
UPDATE recipe_list_items SET archived_at = NOW() WHERE archived_at IS NULL AND belongs_to_recipe_list = sqlc.arg(belongs_to_recipe_list) AND id = sqlc.arg(id);

-- name: CreateRecipeListItem :exec
INSERT INTO recipe_list_items (
	id,
	recipe_id,
	notes,
	belongs_to_recipe_list
) VALUES (
	sqlc.arg(id),
	sqlc.arg(recipe_id),
	sqlc.arg(notes),
	sqlc.arg(belongs_to_recipe_list)
);

-- name: GetRecipeListItems :many
SELECT
	recipe_list_items.id,
	recipe_list_items.recipe_id,
	recipe_list_items.notes,
	recipe_list_items.created_at,
	recipe_list_items.last_updated_at,
	recipe_list_items.archived_at,
	recipe_list_items.belongs_to_recipe_list,
	(
		SELECT COUNT(recipe_list_items.id)
		FROM recipe_list_items
		WHERE recipe_list_items.archived_at IS NULL
			AND
			recipe_list_items.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND recipe_list_items.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				recipe_list_items.last_updated_at IS NULL
				OR recipe_list_items.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				recipe_list_items.last_updated_at IS NULL
				OR recipe_list_items.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_list_items.archived_at IS NULL)
			AND recipe_list_items.belongs_to_recipe_list = sqlc.arg(recipe_list_id)
	) AS filtered_count,
	(
		SELECT COUNT(recipe_list_items.id)
		FROM recipe_list_items
		WHERE recipe_list_items.archived_at IS NULL
	) AS total_count
FROM recipe_list_items
WHERE recipe_list_items.archived_at IS NULL
	AND recipe_list_items.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND recipe_list_items.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		recipe_list_items.last_updated_at IS NULL
		OR recipe_list_items.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		recipe_list_items.last_updated_at IS NULL
		OR recipe_list_items.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND recipe_list_items.belongs_to_recipe_list = sqlc.arg(recipe_list_id)
	AND recipe_list_items.id > COALESCE(sqlc.narg(cursor), '')
	AND recipe_list_items.belongs_to_recipe_list = sqlc.arg(recipe_list_id)
ORDER BY recipe_list_items.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: UpdateRecipeListItem :execrows
UPDATE recipe_list_items SET
	recipe_id = sqlc.arg(recipe_id),
	notes = sqlc.arg(notes),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND belongs_to_recipe_list = sqlc.arg(belongs_to_recipe_list)
	AND id = sqlc.arg(id);
