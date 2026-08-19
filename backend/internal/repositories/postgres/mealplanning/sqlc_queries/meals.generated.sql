-- name: CreateMeal :exec
INSERT INTO meals (
	id,
	name,
	description,
	min_estimated_portions,
	max_estimated_portions,
	eligible_for_meal_plans,
	created_by_user
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(description),
	sqlc.arg(min_estimated_portions),
	sqlc.arg(max_estimated_portions),
	sqlc.arg(eligible_for_meal_plans),
	sqlc.arg(created_by_user)
);

-- name: CheckMealExistence :one
SELECT EXISTS (
	SELECT meals.id
	FROM meals
	WHERE meals.archived_at IS NULL
		AND meals.id = sqlc.arg(id)
);

-- name: ScanMealIDsForReindex :many
SELECT meals.id
FROM meals
WHERE meals.archived_at IS NULL
	AND meals.id COLLATE "C" > sqlc.arg(cursor)
ORDER BY meals.id COLLATE "C"
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: ArchiveMeal :execrows
UPDATE meals SET archived_at = NOW() WHERE archived_at IS NULL AND created_by_user = sqlc.arg(created_by_user) AND id = sqlc.arg(id);

-- name: GetMealsByCreatorAndName :many
SELECT
	meals.id,
	meals.name,
	meals.description,
	meals.min_estimated_portions,
	meals.max_estimated_portions,
	meals.eligible_for_meal_plans,
	meals.last_indexed_at,
	meals.created_at,
	meals.last_updated_at,
	meals.archived_at,
	meals.created_by_user,
	meal_components.id as component_id,
	meal_components.belongs_to_meal as component_belongs_to_meal,
	meal_components.recipe_id as component_recipe_id,
	meal_components.meal_component_type as component_meal_component_type,
	meal_components.recipe_scale as component_recipe_scale,
	meal_components.created_at as component_created_at,
	meal_components.last_updated_at as component_last_updated_at,
	meal_components.archived_at as component_archived_at
FROM meals
	JOIN meal_components ON meal_components.belongs_to_meal=meals.id
		AND meal_components.archived_at IS NULL
		AND EXISTS (SELECT 1 FROM recipes WHERE recipes.id = meal_components.recipe_id AND recipes.archived_at IS NULL)
WHERE meals.archived_at IS NULL
	AND meals.created_by_user = sqlc.arg(created_by_user)
	AND meals.name = sqlc.arg(name)
ORDER BY meals.id ASC, meal_components.id ASC;

-- name: GetMealsNeedingIndexing :many
SELECT meals.id
	FROM meals
	WHERE meals.archived_at IS NULL
	AND (
		meals.last_indexed_at IS NULL
		OR meals.last_indexed_at < NOW() - '24 hours'::INTERVAL
	);

-- name: GetMeal :many
SELECT
	meals.id,
	meals.name,
	meals.description,
	meals.min_estimated_portions,
	meals.max_estimated_portions,
	meals.eligible_for_meal_plans,
	meals.last_indexed_at,
	meals.created_at,
	meals.last_updated_at,
	meals.archived_at,
	meals.created_by_user,
	meal_components.id as component_id,
	meal_components.belongs_to_meal as component_belongs_to_meal,
	meal_components.recipe_id as component_recipe_id,
	meal_components.meal_component_type as component_meal_component_type,
	meal_components.recipe_scale as component_recipe_scale,
	meal_components.created_at as component_created_at,
	meal_components.last_updated_at as component_last_updated_at,
	meal_components.archived_at as component_archived_at
FROM meals
	JOIN meal_components ON meal_components.belongs_to_meal=meals.id
		AND meal_components.archived_at IS NULL
		AND EXISTS (SELECT 1 FROM recipes WHERE recipes.id = meal_components.recipe_id AND recipes.archived_at IS NULL)
WHERE meals.archived_at IS NULL
  AND meals.id = sqlc.arg(id);

-- name: GetMeals :many
SELECT
	meals.id,
	meals.name,
	meals.description,
	meals.min_estimated_portions,
	meals.max_estimated_portions,
	meals.eligible_for_meal_plans,
	meals.last_indexed_at,
	meals.created_at,
	meals.last_updated_at,
	meals.archived_at,
	meals.created_by_user,
	meal_components.id as component_id,
	meal_components.belongs_to_meal as component_belongs_to_meal,
	meal_components.recipe_id as component_recipe_id,
	meal_components.meal_component_type as component_meal_component_type,
	meal_components.recipe_scale as component_recipe_scale,
	meal_components.created_at as component_created_at,
	meal_components.last_updated_at as component_last_updated_at,
	meal_components.archived_at as component_archived_at,
	(
		SELECT COUNT(meals.id)
		FROM meals
		WHERE meals.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND meals.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				meals.last_updated_at IS NULL
				OR meals.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				meals.last_updated_at IS NULL
				OR meals.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR meals.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(meals.id)
		FROM meals
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR meals.archived_at IS NULL)
	) AS total_count
FROM meals
	LEFT JOIN meal_components ON meal_components.belongs_to_meal=meals.id AND meal_components.archived_at IS NULL
		AND EXISTS (SELECT 1 FROM recipes WHERE recipes.id = meal_components.recipe_id AND recipes.archived_at IS NULL)
WHERE meals.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND meals.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		meals.last_updated_at IS NULL
		OR meals.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		meals.last_updated_at IS NULL
		OR meals.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR meals.archived_at IS NULL)
	AND meals.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY meals.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetMealsWithIDs :many
SELECT
	meals.id,
	meals.name,
	meals.description,
	meals.min_estimated_portions,
	meals.max_estimated_portions,
	meals.eligible_for_meal_plans,
	meals.last_indexed_at,
	meals.created_at,
	meals.last_updated_at,
	meals.archived_at,
	meals.created_by_user,
	meal_components.id as component_id,
	meal_components.belongs_to_meal as component_belongs_to_meal,
	meal_components.recipe_id as component_recipe_id,
	meal_components.meal_component_type as component_meal_component_type,
	meal_components.recipe_scale as component_recipe_scale,
	meal_components.created_at as component_created_at,
	meal_components.last_updated_at as component_last_updated_at,
	meal_components.archived_at as component_archived_at
FROM meals
	JOIN meal_components ON meal_components.belongs_to_meal=meals.id
		AND meal_components.archived_at IS NULL
		AND EXISTS (SELECT 1 FROM recipes WHERE recipes.id = meal_components.recipe_id AND recipes.archived_at IS NULL)
WHERE meals.archived_at IS NULL
  AND meals.id = ANY(sqlc.arg(ids)::text[])
ORDER BY meals.id ASC;

-- name: GetMealsCreatedByUser :many
SELECT
	meals.id,
	meals.name,
	meals.description,
	meals.min_estimated_portions,
	meals.max_estimated_portions,
	meals.eligible_for_meal_plans,
	meals.last_indexed_at,
	meals.created_at,
	meals.last_updated_at,
	meals.archived_at,
	meals.created_by_user,
	meal_components.id as component_id,
	meal_components.belongs_to_meal as component_belongs_to_meal,
	meal_components.recipe_id as component_recipe_id,
	meal_components.meal_component_type as component_meal_component_type,
	meal_components.recipe_scale as component_recipe_scale,
	meal_components.created_at as component_created_at,
	meal_components.last_updated_at as component_last_updated_at,
	meal_components.archived_at as component_archived_at,
	(
		SELECT COUNT(meals.id)
		FROM meals
		WHERE meals.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND meals.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				meals.last_updated_at IS NULL
				OR meals.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				meals.last_updated_at IS NULL
				OR meals.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR meals.archived_at IS NULL)
			AND meals.created_by_user = sqlc.arg(created_by_user)
	) AS filtered_count,
	(
		SELECT COUNT(meals.id)
		FROM meals
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR meals.archived_at IS NULL)
			AND meals.created_by_user = sqlc.arg(created_by_user)
	) AS total_count
FROM meals
	LEFT JOIN meal_components ON meal_components.belongs_to_meal=meals.id AND meal_components.archived_at IS NULL
		AND EXISTS (SELECT 1 FROM recipes WHERE recipes.id = meal_components.recipe_id AND recipes.archived_at IS NULL)
WHERE meals.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND meals.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		meals.last_updated_at IS NULL
		OR meals.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		meals.last_updated_at IS NULL
		OR meals.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR meals.archived_at IS NULL)
	AND meals.created_by_user = sqlc.arg(created_by_user)
	AND meals.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY meals.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: SearchForMeals :many
SELECT
	meals.id,
	meals.name,
	meals.description,
	meals.min_estimated_portions,
	meals.max_estimated_portions,
	meals.eligible_for_meal_plans,
	meals.last_indexed_at,
	meals.created_at,
	meals.last_updated_at,
	meals.archived_at,
	meals.created_by_user,
	meal_components.id as component_id,
	meal_components.belongs_to_meal as component_belongs_to_meal,
	meal_components.recipe_id as component_recipe_id,
	meal_components.meal_component_type as component_meal_component_type,
	meal_components.recipe_scale as component_recipe_scale,
	meal_components.created_at as component_created_at,
	meal_components.last_updated_at as component_last_updated_at,
	meal_components.archived_at as component_archived_at,
	(
		SELECT COUNT(meals.id)
		FROM meals
		WHERE meals.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND meals.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				meals.last_updated_at IS NULL
				OR meals.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				meals.last_updated_at IS NULL
				OR meals.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR meals.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(meals.id)
		FROM meals
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR meals.archived_at IS NULL)
	) AS total_count
FROM meals
	JOIN meal_components ON meal_components.belongs_to_meal=meals.id
		AND meal_components.archived_at IS NULL
		AND EXISTS (SELECT 1 FROM recipes WHERE recipes.id = meal_components.recipe_id AND recipes.archived_at IS NULL)
WHERE meals.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND meals.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		meals.last_updated_at IS NULL
		OR meals.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		meals.last_updated_at IS NULL
		OR meals.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR meals.archived_at IS NULL)
	AND meals.name ILIKE '%' || sqlc.arg(query)::text || '%'
	AND meals.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY meals.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: MarkMealsAsIndexed :execrows
UPDATE meals SET
	last_indexed_at = NOW()
WHERE id = ANY(sqlc.arg(ids)::text[]);
