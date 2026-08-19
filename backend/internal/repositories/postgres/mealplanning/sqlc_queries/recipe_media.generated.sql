-- name: CreateRecipeMedia :exec
INSERT INTO recipe_media (
	id,
	belongs_to_recipe,
	belongs_to_recipe_step,
	mime_type,
	internal_path,
	external_path,
	index
) VALUES (
	sqlc.arg(id),
	sqlc.arg(belongs_to_recipe),
	sqlc.arg(belongs_to_recipe_step),
	sqlc.arg(mime_type),
	sqlc.arg(internal_path),
	sqlc.arg(external_path),
	sqlc.arg(index)
);

-- name: GetRecipeMedia :one
SELECT
	recipe_media.id,
	recipe_media.belongs_to_recipe,
	recipe_media.belongs_to_recipe_step,
	recipe_media.mime_type,
	recipe_media.internal_path,
	recipe_media.external_path,
	recipe_media.index,
	recipe_media.created_at,
	recipe_media.last_updated_at,
	recipe_media.archived_at
FROM recipe_media
WHERE recipe_media.archived_at IS NULL
	AND recipe_media.id = sqlc.arg(id);

-- name: CheckRecipeMediaExistence :one
SELECT EXISTS (
	SELECT recipe_media.id
	FROM recipe_media
	WHERE recipe_media.archived_at IS NULL
		AND recipe_media.id = sqlc.arg(id)
);

-- name: UpdateRecipeMedia :execrows
UPDATE recipe_media SET
	belongs_to_recipe = sqlc.arg(belongs_to_recipe),
	belongs_to_recipe_step = sqlc.arg(belongs_to_recipe_step),
	mime_type = sqlc.arg(mime_type),
	internal_path = sqlc.arg(internal_path),
	external_path = sqlc.arg(external_path),
	index = sqlc.arg(index),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveRecipeMedia :execrows
UPDATE recipe_media SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: GetRecipeMediaForRecipe :many
SELECT
	recipe_media.id,
	recipe_media.belongs_to_recipe,
	recipe_media.belongs_to_recipe_step,
	recipe_media.mime_type,
	recipe_media.internal_path,
	recipe_media.external_path,
	recipe_media.index,
	recipe_media.created_at,
	recipe_media.last_updated_at,
	recipe_media.archived_at
FROM recipe_media
WHERE recipe_media.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipe_media.belongs_to_recipe_step IS NULL
	AND recipe_media.archived_at IS NULL
GROUP BY recipe_media.id
ORDER BY recipe_media.id;

-- name: GetRecipeMediaForRecipeStep :many
SELECT
	recipe_media.id,
	recipe_media.belongs_to_recipe,
	recipe_media.belongs_to_recipe_step,
	recipe_media.mime_type,
	recipe_media.internal_path,
	recipe_media.external_path,
	recipe_media.index,
	recipe_media.created_at,
	recipe_media.last_updated_at,
	recipe_media.archived_at
FROM recipe_media
WHERE recipe_media.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipe_media.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
	AND recipe_media.archived_at IS NULL
GROUP BY recipe_media.id
ORDER BY recipe_media.id;
