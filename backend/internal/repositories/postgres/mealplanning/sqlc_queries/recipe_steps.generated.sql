-- name: CreateRecipeStep :exec
INSERT INTO recipe_steps (
	id,
	index,
	preparation_id,
	minimum_estimated_time_in_seconds,
	maximum_estimated_time_in_seconds,
	minimum_temperature_in_celsius,
	maximum_temperature_in_celsius,
	notes,
	explicit_instructions,
	condition_expression,
	optional,
	start_timer_automatically,
	belongs_to_recipe
) VALUES (
	sqlc.arg(id),
	sqlc.arg(index),
	sqlc.arg(preparation_id),
	sqlc.arg(minimum_estimated_time_in_seconds),
	sqlc.arg(maximum_estimated_time_in_seconds),
	sqlc.arg(minimum_temperature_in_celsius),
	sqlc.arg(maximum_temperature_in_celsius),
	sqlc.arg(notes),
	sqlc.arg(explicit_instructions),
	sqlc.arg(condition_expression),
	sqlc.arg(optional),
	sqlc.arg(start_timer_automatically),
	sqlc.arg(belongs_to_recipe)
);

-- name: UpdateRecipeStep :execrows
UPDATE recipe_steps SET
	index = sqlc.arg(index),
	preparation_id = sqlc.arg(preparation_id),
	minimum_estimated_time_in_seconds = sqlc.arg(minimum_estimated_time_in_seconds),
	maximum_estimated_time_in_seconds = sqlc.arg(maximum_estimated_time_in_seconds),
	minimum_temperature_in_celsius = sqlc.arg(minimum_temperature_in_celsius),
	maximum_temperature_in_celsius = sqlc.arg(maximum_temperature_in_celsius),
	notes = sqlc.arg(notes),
	explicit_instructions = sqlc.arg(explicit_instructions),
	condition_expression = sqlc.arg(condition_expression),
	optional = sqlc.arg(optional),
	start_timer_automatically = sqlc.arg(start_timer_automatically),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_recipe = sqlc.arg(belongs_to_recipe);

-- name: ArchiveRecipeStep :execrows
UPDATE recipe_steps SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_recipe = sqlc.arg(belongs_to_recipe);

-- name: CheckRecipeStepExistence :one
SELECT EXISTS (
	SELECT recipe_steps.id
	FROM recipe_steps
		JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
	WHERE recipe_steps.archived_at IS NULL
		AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
		AND recipe_steps.id = sqlc.arg(recipe_step_id)
		AND recipes.archived_at IS NULL
		AND recipes.id = sqlc.arg(recipe_id)
);

-- name: GetRecipeStep :one
SELECT
	recipe_steps.id,
	recipe_steps.index,
	valid_preparations.id as valid_preparation_id,
	valid_preparations.name as valid_preparation_name,
	valid_preparations.description as valid_preparation_description,
	valid_preparations.icon_path as valid_preparation_icon_path,
	valid_preparations.yields_nothing as valid_preparation_yields_nothing,
	valid_preparations.restrict_to_ingredients as valid_preparation_restrict_to_ingredients,
	valid_preparations.past_tense as valid_preparation_past_tense,
	valid_preparations.slug as valid_preparation_slug,
	valid_preparations.minimum_ingredient_count as valid_preparation_minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count as valid_preparation_maximum_ingredient_count,
	valid_preparations.minimum_instrument_count as valid_preparation_minimum_instrument_count,
	valid_preparations.maximum_instrument_count as valid_preparation_maximum_instrument_count,
	valid_preparations.temperature_required as valid_preparation_temperature_required,
	valid_preparations.time_estimate_required as valid_preparation_time_estimate_required,
	valid_preparations.condition_expression_required as valid_preparation_condition_expression_required,
	valid_preparations.consumes_vessel as valid_preparation_consumes_vessel,
	valid_preparations.only_for_vessels as valid_preparation_only_for_vessels,
	valid_preparations.minimum_vessel_count as valid_preparation_minimum_vessel_count,
	valid_preparations.maximum_vessel_count as valid_preparation_maximum_vessel_count,
	valid_preparations.last_indexed_at as valid_preparation_last_indexed_at,
	valid_preparations.created_at as valid_preparation_created_at,
	valid_preparations.last_updated_at as valid_preparation_last_updated_at,
	valid_preparations.archived_at as valid_preparation_archived_at,
	recipe_steps.minimum_estimated_time_in_seconds,
	recipe_steps.maximum_estimated_time_in_seconds,
	recipe_steps.minimum_temperature_in_celsius,
	recipe_steps.maximum_temperature_in_celsius,
	recipe_steps.notes,
	recipe_steps.explicit_instructions,
	recipe_steps.condition_expression,
	recipe_steps.optional,
	recipe_steps.start_timer_automatically,
	recipe_steps.created_at,
	recipe_steps.last_updated_at,
	recipe_steps.archived_at,
	recipe_steps.belongs_to_recipe
FROM recipe_steps
	JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
	JOIN valid_preparations ON recipe_steps.preparation_id=valid_preparations.id
WHERE recipe_steps.archived_at IS NULL
	AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipe_steps.id = sqlc.arg(recipe_step_id)
	AND recipes.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id);

-- name: GetRecipeSteps :many
SELECT
	recipe_steps.id,
	recipe_steps.index,
	valid_preparations.id as valid_preparation_id,
	valid_preparations.name as valid_preparation_name,
	valid_preparations.description as valid_preparation_description,
	valid_preparations.icon_path as valid_preparation_icon_path,
	valid_preparations.yields_nothing as valid_preparation_yields_nothing,
	valid_preparations.restrict_to_ingredients as valid_preparation_restrict_to_ingredients,
	valid_preparations.past_tense as valid_preparation_past_tense,
	valid_preparations.slug as valid_preparation_slug,
	valid_preparations.minimum_ingredient_count as valid_preparation_minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count as valid_preparation_maximum_ingredient_count,
	valid_preparations.minimum_instrument_count as valid_preparation_minimum_instrument_count,
	valid_preparations.maximum_instrument_count as valid_preparation_maximum_instrument_count,
	valid_preparations.temperature_required as valid_preparation_temperature_required,
	valid_preparations.time_estimate_required as valid_preparation_time_estimate_required,
	valid_preparations.condition_expression_required as valid_preparation_condition_expression_required,
	valid_preparations.consumes_vessel as valid_preparation_consumes_vessel,
	valid_preparations.only_for_vessels as valid_preparation_only_for_vessels,
	valid_preparations.minimum_vessel_count as valid_preparation_minimum_vessel_count,
	valid_preparations.maximum_vessel_count as valid_preparation_maximum_vessel_count,
	valid_preparations.last_indexed_at as valid_preparation_last_indexed_at,
	valid_preparations.created_at as valid_preparation_created_at,
	valid_preparations.last_updated_at as valid_preparation_last_updated_at,
	valid_preparations.archived_at as valid_preparation_archived_at,
	recipe_steps.minimum_estimated_time_in_seconds,
	recipe_steps.maximum_estimated_time_in_seconds,
	recipe_steps.minimum_temperature_in_celsius,
	recipe_steps.maximum_temperature_in_celsius,
	recipe_steps.notes,
	recipe_steps.explicit_instructions,
	recipe_steps.condition_expression,
	recipe_steps.optional,
	recipe_steps.start_timer_automatically,
	recipe_steps.created_at,
	recipe_steps.last_updated_at,
	recipe_steps.archived_at,
	recipe_steps.belongs_to_recipe,
	(
		SELECT COUNT(recipe_steps.id)
		FROM recipe_steps
		WHERE recipe_steps.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND recipe_steps.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				recipe_steps.last_updated_at IS NULL
				OR recipe_steps.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				recipe_steps.last_updated_at IS NULL
				OR recipe_steps.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_steps.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(recipe_steps.id)
		FROM recipe_steps
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_steps.archived_at IS NULL)
	) AS total_count
FROM recipe_steps
	JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
	JOIN valid_preparations ON recipe_steps.preparation_id=valid_preparations.id
WHERE recipe_steps.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND recipe_steps.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		recipe_steps.last_updated_at IS NULL
		OR recipe_steps.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		recipe_steps.last_updated_at IS NULL
		OR recipe_steps.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_steps.archived_at IS NULL)
	AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipes.archived_at IS NULL
	AND recipe_steps.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY recipe_steps.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetRecipeStepByRecipeID :one
SELECT
	recipe_steps.id,
	recipe_steps.index,
	valid_preparations.id as valid_preparation_id,
	valid_preparations.name as valid_preparation_name,
	valid_preparations.description as valid_preparation_description,
	valid_preparations.icon_path as valid_preparation_icon_path,
	valid_preparations.yields_nothing as valid_preparation_yields_nothing,
	valid_preparations.restrict_to_ingredients as valid_preparation_restrict_to_ingredients,
	valid_preparations.past_tense as valid_preparation_past_tense,
	valid_preparations.slug as valid_preparation_slug,
	valid_preparations.minimum_ingredient_count as valid_preparation_minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count as valid_preparation_maximum_ingredient_count,
	valid_preparations.minimum_instrument_count as valid_preparation_minimum_instrument_count,
	valid_preparations.maximum_instrument_count as valid_preparation_maximum_instrument_count,
	valid_preparations.temperature_required as valid_preparation_temperature_required,
	valid_preparations.time_estimate_required as valid_preparation_time_estimate_required,
	valid_preparations.condition_expression_required as valid_preparation_condition_expression_required,
	valid_preparations.consumes_vessel as valid_preparation_consumes_vessel,
	valid_preparations.only_for_vessels as valid_preparation_only_for_vessels,
	valid_preparations.minimum_vessel_count as valid_preparation_minimum_vessel_count,
	valid_preparations.maximum_vessel_count as valid_preparation_maximum_vessel_count,
	valid_preparations.last_indexed_at as valid_preparation_last_indexed_at,
	valid_preparations.created_at as valid_preparation_created_at,
	valid_preparations.last_updated_at as valid_preparation_last_updated_at,
	valid_preparations.archived_at as valid_preparation_archived_at,
	recipe_steps.minimum_estimated_time_in_seconds,
	recipe_steps.maximum_estimated_time_in_seconds,
	recipe_steps.minimum_temperature_in_celsius,
	recipe_steps.maximum_temperature_in_celsius,
	recipe_steps.notes,
	recipe_steps.explicit_instructions,
	recipe_steps.condition_expression,
	recipe_steps.optional,
	recipe_steps.start_timer_automatically,
	recipe_steps.created_at,
	recipe_steps.last_updated_at,
	recipe_steps.archived_at,
	recipe_steps.belongs_to_recipe
FROM recipe_steps
	JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
	JOIN valid_preparations ON recipe_steps.preparation_id=valid_preparations.id
WHERE recipe_steps.archived_at IS NULL
	AND recipe_steps.id = sqlc.arg(id);
