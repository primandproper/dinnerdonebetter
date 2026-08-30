-- name: CreateRecipeList :exec
INSERT INTO recipe_lists (
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

-- name: UpdateRecipeList :execrows
UPDATE recipe_lists SET
	name = sqlc.arg(name),
	description = sqlc.arg(description),
	last_updated_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_user = sqlc.arg(belongs_to_user);

-- name: ArchiveRecipeList :execrows
UPDATE recipe_lists SET
	archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_user = sqlc.arg(belongs_to_user);

-- name: GetRecipeLists :many
SELECT
	recipe_lists.id,
	recipe_lists.name,
	recipe_list_items.id as recipe_list_item_id,
	recipe_list_items.recipe_id as recipe_list_item_recipe_id,
	recipe_list_items.notes as recipe_list_item_notes,
	recipe_list_items.created_at as recipe_list_item_created_at,
	recipe_list_items.last_updated_at as recipe_list_item_last_updated_at,
	recipe_list_items.archived_at as recipe_list_item_archived_at,
	recipe_list_items.belongs_to_recipe_list as recipe_list_item_belongs_to_recipe_list,
	recipe_lists.description,
	recipe_lists.created_at,
	recipe_lists.last_updated_at,
	recipe_lists.archived_at,
	recipe_lists.belongs_to_user,
	(
		SELECT COUNT(recipe_lists.id)
		FROM recipe_lists
		WHERE recipe_lists.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND recipe_lists.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				recipe_lists.last_updated_at IS NULL
				OR recipe_lists.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				recipe_lists.last_updated_at IS NULL
				OR recipe_lists.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_lists.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(recipe_lists.id)
		FROM recipe_lists
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_lists.archived_at IS NULL)
	) AS total_count
FROM recipe_lists
	LEFT JOIN recipe_list_items ON recipe_list_items.belongs_to_recipe_list = recipe_lists.id AND recipe_list_items.archived_at IS NULL
WHERE recipe_lists.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND recipe_lists.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		recipe_lists.last_updated_at IS NULL
		OR recipe_lists.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		recipe_lists.last_updated_at IS NULL
		OR recipe_lists.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_lists.archived_at IS NULL)
	AND recipe_lists.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY recipe_lists.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
