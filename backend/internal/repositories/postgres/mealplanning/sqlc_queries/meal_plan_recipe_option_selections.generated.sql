-- name: CreateMealPlanRecipeOptionSelection :exec
INSERT INTO meal_plan_recipe_option_selections (
	id,
	belongs_to_meal_plan_option,
	recipe_id,
	recipe_step_id,
	ingredient_index,
	selected_option_index,
	selection_type
) VALUES (
	sqlc.arg(id),
	sqlc.arg(belongs_to_meal_plan_option),
	sqlc.arg(recipe_id),
	sqlc.arg(recipe_step_id),
	sqlc.arg(ingredient_index),
	sqlc.arg(selected_option_index),
	sqlc.arg(selection_type)
) ON CONFLICT (belongs_to_meal_plan_option, recipe_step_id, ingredient_index, selection_type) DO UPDATE SET
	selected_option_index = EXCLUDED.selected_option_index,
	last_updated_at = CURRENT_TIMESTAMP;

-- name: GetMealPlanRecipeOptionSelection :one
SELECT
	meal_plan_recipe_option_selections.id,
	meal_plan_recipe_option_selections.belongs_to_meal_plan_option,
	meal_plan_recipe_option_selections.recipe_id,
	meal_plan_recipe_option_selections.recipe_step_id,
	meal_plan_recipe_option_selections.ingredient_index,
	meal_plan_recipe_option_selections.selected_option_index,
	meal_plan_recipe_option_selections.selection_type,
	meal_plan_recipe_option_selections.created_at,
	meal_plan_recipe_option_selections.last_updated_at,
	meal_plan_recipe_option_selections.archived_at
FROM meal_plan_recipe_option_selections
WHERE belongs_to_meal_plan_option = sqlc.arg(meal_plan_option_id)
	AND recipe_step_id = sqlc.arg(recipe_step_id)
	AND ingredient_index = sqlc.arg(ingredient_index)
	AND selection_type = sqlc.arg(selection_type);

-- name: GetMealPlanRecipeOptionSelectionsForMealPlanOption :many
SELECT
	meal_plan_recipe_option_selections.id,
	meal_plan_recipe_option_selections.belongs_to_meal_plan_option,
	meal_plan_recipe_option_selections.recipe_id,
	meal_plan_recipe_option_selections.recipe_step_id,
	meal_plan_recipe_option_selections.ingredient_index,
	meal_plan_recipe_option_selections.selected_option_index,
	meal_plan_recipe_option_selections.selection_type,
	meal_plan_recipe_option_selections.created_at,
	meal_plan_recipe_option_selections.last_updated_at,
	meal_plan_recipe_option_selections.archived_at,
	(
		SELECT COUNT(meal_plan_recipe_option_selections.id)
		FROM meal_plan_recipe_option_selections
		WHERE meal_plan_recipe_option_selections.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND meal_plan_recipe_option_selections.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				meal_plan_recipe_option_selections.last_updated_at IS NULL
				OR meal_plan_recipe_option_selections.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				meal_plan_recipe_option_selections.last_updated_at IS NULL
				OR meal_plan_recipe_option_selections.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR meal_plan_recipe_option_selections.archived_at IS NULL)
			AND meal_plan_recipe_option_selections.belongs_to_meal_plan_option = sqlc.arg(meal_plan_option_id)
	) AS filtered_count,
	(
		SELECT COUNT(meal_plan_recipe_option_selections.id)
		FROM meal_plan_recipe_option_selections
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR meal_plan_recipe_option_selections.archived_at IS NULL)
			AND meal_plan_recipe_option_selections.belongs_to_meal_plan_option = sqlc.arg(meal_plan_option_id)
	) AS total_count
FROM meal_plan_recipe_option_selections
WHERE meal_plan_recipe_option_selections.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND meal_plan_recipe_option_selections.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		meal_plan_recipe_option_selections.last_updated_at IS NULL
		OR meal_plan_recipe_option_selections.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		meal_plan_recipe_option_selections.last_updated_at IS NULL
		OR meal_plan_recipe_option_selections.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR meal_plan_recipe_option_selections.archived_at IS NULL)
	AND belongs_to_meal_plan_option = sqlc.arg(meal_plan_option_id)
	AND meal_plan_recipe_option_selections.belongs_to_meal_plan_option = sqlc.arg(meal_plan_option_id)
	AND meal_plan_recipe_option_selections.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY meal_plan_recipe_option_selections.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetMealPlanRecipeOptionSelectionsForMealPlan :many
SELECT
	meal_plan_recipe_option_selections.id,
	meal_plan_recipe_option_selections.belongs_to_meal_plan_option,
	meal_plan_recipe_option_selections.recipe_id,
	meal_plan_recipe_option_selections.recipe_step_id,
	meal_plan_recipe_option_selections.ingredient_index,
	meal_plan_recipe_option_selections.selected_option_index,
	meal_plan_recipe_option_selections.selection_type,
	meal_plan_recipe_option_selections.created_at,
	meal_plan_recipe_option_selections.last_updated_at,
	meal_plan_recipe_option_selections.archived_at,
	(
		SELECT COUNT(meal_plan_recipe_option_selections.id)
		FROM meal_plan_recipe_option_selections
		JOIN meal_plan_options ON meal_plan_recipe_option_selections.belongs_to_meal_plan_option = meal_plan_options.id
		JOIN meal_plan_events ON meal_plan_options.belongs_to_meal_plan_event = meal_plan_events.id
		WHERE meal_plan_recipe_option_selections.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND meal_plan_recipe_option_selections.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				meal_plan_recipe_option_selections.last_updated_at IS NULL
				OR meal_plan_recipe_option_selections.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				meal_plan_recipe_option_selections.last_updated_at IS NULL
				OR meal_plan_recipe_option_selections.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR meal_plan_recipe_option_selections.archived_at IS NULL)
			AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id) AND meal_plan_options.archived_at IS NULL AND meal_plan_events.archived_at IS NULL
	) AS filtered_count,
	(
		SELECT COUNT(meal_plan_recipe_option_selections.id)
		FROM meal_plan_recipe_option_selections
		JOIN meal_plan_options ON meal_plan_recipe_option_selections.belongs_to_meal_plan_option = meal_plan_options.id
		JOIN meal_plan_events ON meal_plan_options.belongs_to_meal_plan_event = meal_plan_events.id
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR meal_plan_recipe_option_selections.archived_at IS NULL)
			AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id) AND meal_plan_options.archived_at IS NULL AND meal_plan_events.archived_at IS NULL
	) AS total_count
FROM meal_plan_recipe_option_selections
	JOIN meal_plan_options ON meal_plan_recipe_option_selections.belongs_to_meal_plan_option = meal_plan_options.id
	JOIN meal_plan_events ON meal_plan_options.belongs_to_meal_plan_event = meal_plan_events.id
WHERE meal_plan_recipe_option_selections.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND meal_plan_recipe_option_selections.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		meal_plan_recipe_option_selections.last_updated_at IS NULL
		OR meal_plan_recipe_option_selections.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		meal_plan_recipe_option_selections.last_updated_at IS NULL
		OR meal_plan_recipe_option_selections.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR meal_plan_recipe_option_selections.archived_at IS NULL)
	AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
	AND meal_plan_options.archived_at IS NULL
	AND meal_plan_events.archived_at IS NULL
	AND meal_plan_recipe_option_selections.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY meal_plan_recipe_option_selections.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: UpdateMealPlanRecipeOptionSelection :execrows
UPDATE meal_plan_recipe_option_selections SET
	recipe_id = sqlc.arg(recipe_id),
	selected_option_index = sqlc.arg(selected_option_index),
	last_updated_at = CURRENT_TIMESTAMP
WHERE belongs_to_meal_plan_option = sqlc.arg(meal_plan_option_id)
	AND recipe_step_id = sqlc.arg(recipe_step_id)
	AND ingredient_index = sqlc.arg(ingredient_index)
	AND selection_type = sqlc.arg(selection_type);

-- name: ArchiveMealPlanRecipeOptionSelection :execrows
UPDATE meal_plan_recipe_option_selections SET
	archived_at = CURRENT_TIMESTAMP
WHERE belongs_to_meal_plan_option = sqlc.arg(meal_plan_option_id)
	AND recipe_step_id = sqlc.arg(recipe_step_id)
	AND ingredient_index = sqlc.arg(ingredient_index)
	AND selection_type = sqlc.arg(selection_type);
