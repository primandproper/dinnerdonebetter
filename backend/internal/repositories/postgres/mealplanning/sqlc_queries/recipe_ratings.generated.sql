-- name: CreateRecipeRating :exec
INSERT INTO recipe_ratings (
	id,
	belongs_to_recipe,
	taste,
	difficulty,
	cleanup,
	instructions,
	overall,
	notes,
	created_by_user
) VALUES (
	sqlc.arg(id),
	sqlc.arg(belongs_to_recipe),
	sqlc.arg(taste),
	sqlc.arg(difficulty),
	sqlc.arg(cleanup),
	sqlc.arg(instructions),
	sqlc.arg(overall),
	sqlc.arg(notes),
	sqlc.arg(created_by_user)
);

-- name: GetRecipeRating :one
SELECT
	recipe_ratings.id,
	recipe_ratings.belongs_to_recipe,
	recipe_ratings.taste,
	recipe_ratings.difficulty,
	recipe_ratings.cleanup,
	recipe_ratings.instructions,
	recipe_ratings.overall,
	recipe_ratings.notes,
	recipe_ratings.created_by_user,
	recipe_ratings.created_at,
	recipe_ratings.last_updated_at,
	recipe_ratings.archived_at
FROM recipe_ratings
WHERE recipe_ratings.archived_at IS NULL
	AND recipe_ratings.id = sqlc.arg(id);

-- name: CheckRecipeRatingExistence :one
SELECT EXISTS (
	SELECT recipe_ratings.id
	FROM recipe_ratings
	WHERE recipe_ratings.archived_at IS NULL
		AND recipe_ratings.id = sqlc.arg(id)
);

-- name: UpdateRecipeRating :execrows
UPDATE recipe_ratings SET
	belongs_to_recipe = sqlc.arg(belongs_to_recipe),
	taste = sqlc.arg(taste),
	difficulty = sqlc.arg(difficulty),
	cleanup = sqlc.arg(cleanup),
	instructions = sqlc.arg(instructions),
	overall = sqlc.arg(overall),
	notes = sqlc.arg(notes),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveRecipeRating :execrows
UPDATE recipe_ratings SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: GetRecipeRatingsForRecipe :many
SELECT
	recipe_ratings.id,
	recipe_ratings.belongs_to_recipe,
	recipe_ratings.taste,
	recipe_ratings.difficulty,
	recipe_ratings.cleanup,
	recipe_ratings.instructions,
	recipe_ratings.overall,
	recipe_ratings.notes,
	recipe_ratings.created_by_user,
	recipe_ratings.created_at,
	recipe_ratings.last_updated_at,
	recipe_ratings.archived_at,
	(
		SELECT COUNT(recipe_ratings.id)
		FROM recipe_ratings
		WHERE recipe_ratings.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND recipe_ratings.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				recipe_ratings.last_updated_at IS NULL
				OR recipe_ratings.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				recipe_ratings.last_updated_at IS NULL
				OR recipe_ratings.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_ratings.archived_at IS NULL)
			AND recipe_ratings.belongs_to_recipe = sqlc.arg(belongs_to_recipe)
	) AS filtered_count,
	(
		SELECT COUNT(recipe_ratings.id)
		FROM recipe_ratings
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_ratings.archived_at IS NULL)
			AND recipe_ratings.belongs_to_recipe = sqlc.arg(belongs_to_recipe)
	) AS total_count
FROM recipe_ratings
WHERE recipe_ratings.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND recipe_ratings.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		recipe_ratings.last_updated_at IS NULL
		OR recipe_ratings.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		recipe_ratings.last_updated_at IS NULL
		OR recipe_ratings.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_ratings.archived_at IS NULL)
	AND recipe_ratings.archived_at IS NULL AND
		recipe_ratings.belongs_to_recipe = sqlc.arg(belongs_to_recipe)
	AND recipe_ratings.belongs_to_recipe = sqlc.arg(belongs_to_recipe)
	AND recipe_ratings.id > COALESCE(sqlc.narg(cursor), '')
GROUP BY recipe_ratings.id
ORDER BY recipe_ratings.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetRecipeRatingsForUser :many
SELECT
	recipe_ratings.id,
	recipe_ratings.belongs_to_recipe,
	recipe_ratings.taste,
	recipe_ratings.difficulty,
	recipe_ratings.cleanup,
	recipe_ratings.instructions,
	recipe_ratings.overall,
	recipe_ratings.notes,
	recipe_ratings.created_by_user,
	recipe_ratings.created_at,
	recipe_ratings.last_updated_at,
	recipe_ratings.archived_at,
	(
		SELECT COUNT(recipe_ratings.id)
		FROM recipe_ratings
		WHERE recipe_ratings.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND recipe_ratings.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				recipe_ratings.last_updated_at IS NULL
				OR recipe_ratings.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				recipe_ratings.last_updated_at IS NULL
				OR recipe_ratings.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_ratings.archived_at IS NULL)
			AND recipe_ratings.created_by_user = sqlc.arg(created_by_user)
	) AS filtered_count,
	(
		SELECT COUNT(recipe_ratings.id)
		FROM recipe_ratings
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_ratings.archived_at IS NULL)
			AND recipe_ratings.created_by_user = sqlc.arg(created_by_user)
	) AS total_count
FROM recipe_ratings
WHERE recipe_ratings.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND recipe_ratings.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		recipe_ratings.last_updated_at IS NULL
		OR recipe_ratings.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		recipe_ratings.last_updated_at IS NULL
		OR recipe_ratings.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_ratings.archived_at IS NULL)
	AND recipe_ratings.archived_at IS NULL AND
		recipe_ratings.created_by_user = sqlc.arg(created_by_user)
	AND recipe_ratings.created_by_user = sqlc.arg(created_by_user)
	AND recipe_ratings.id > COALESCE(sqlc.narg(cursor), '')
GROUP BY recipe_ratings.id
ORDER BY recipe_ratings.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
