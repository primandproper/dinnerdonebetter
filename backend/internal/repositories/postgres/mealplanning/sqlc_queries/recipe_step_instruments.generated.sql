-- name: ArchiveRecipeStepInstrument :execrows
UPDATE recipe_step_instruments SET archived_at = NOW() WHERE archived_at IS NULL AND belongs_to_recipe_step = sqlc.arg(belongs_to_recipe_step) AND id = sqlc.arg(id);

-- name: CreateRecipeStepInstrument :exec
INSERT INTO recipe_step_instruments (
	id,
	instrument_id,
	recipe_step_product_id,
	name,
	notes,
	preference_rank,
	optional,
	minimum_quantity,
	maximum_quantity,
	index,
	option_index,
	scale_factor,
	belongs_to_recipe_step
) VALUES (
	sqlc.arg(id),
	sqlc.arg(instrument_id),
	sqlc.arg(recipe_step_product_id),
	sqlc.arg(name),
	sqlc.arg(notes),
	sqlc.arg(preference_rank),
	sqlc.arg(optional),
	sqlc.arg(minimum_quantity),
	sqlc.arg(maximum_quantity),
	sqlc.arg(index),
	sqlc.arg(option_index),
	sqlc.arg(scale_factor),
	sqlc.arg(belongs_to_recipe_step)
);

-- name: CheckRecipeStepInstrumentExistence :one
SELECT EXISTS (
	SELECT recipe_step_instruments.id
	FROM recipe_step_instruments
		JOIN recipe_steps ON recipe_step_instruments.belongs_to_recipe_step=recipe_steps.id
		JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
	WHERE recipe_step_instruments.archived_at IS NULL
		AND recipe_step_instruments.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
		AND recipe_step_instruments.id = sqlc.arg(recipe_step_instrument_id)
		AND recipe_steps.archived_at IS NULL
		AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
		AND recipe_steps.id = sqlc.arg(recipe_step_id)
		AND recipes.archived_at IS NULL
		AND recipes.id = sqlc.arg(recipe_id)
);

-- name: GetRecipeStepInstrumentsForRecipe :many
SELECT
	recipe_step_instruments.id,
	valid_instruments.id as valid_instrument_id,
	valid_instruments.name as valid_instrument_name,
	valid_instruments.description as valid_instrument_description,
	valid_instruments.icon_path as valid_instrument_icon_path,
	valid_instruments.plural_name as valid_instrument_plural_name,
	valid_instruments.usable_for_storage as valid_instrument_usable_for_storage,
	valid_instruments.slug as valid_instrument_slug,
	valid_instruments.display_in_summary_lists as valid_instrument_display_in_summary_lists,
	valid_instruments.include_in_generated_instructions as valid_instrument_include_in_generated_instructions,
	valid_instruments.last_indexed_at as valid_instrument_last_indexed_at,
	valid_instruments.created_at as valid_instrument_created_at,
	valid_instruments.last_updated_at as valid_instrument_last_updated_at,
	valid_instruments.archived_at as valid_instrument_archived_at,
	recipe_step_instruments.recipe_step_product_id,
	recipe_step_instruments.name,
	recipe_step_instruments.notes,
	recipe_step_instruments.preference_rank,
	recipe_step_instruments.optional,
	recipe_step_instruments.minimum_quantity,
	recipe_step_instruments.maximum_quantity,
	recipe_step_instruments.index,
	recipe_step_instruments.option_index,
	recipe_step_instruments.scale_factor,
	recipe_step_instruments.created_at,
	recipe_step_instruments.last_updated_at,
	recipe_step_instruments.archived_at,
	recipe_step_instruments.belongs_to_recipe_step
FROM recipe_step_instruments
	LEFT JOIN valid_instruments ON recipe_step_instruments.instrument_id=valid_instruments.id
	JOIN recipe_steps ON recipe_step_instruments.belongs_to_recipe_step=recipe_steps.id
	JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
WHERE recipe_step_instruments.archived_at IS NULL
	AND recipe_steps.archived_at IS NULL
	AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipes.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id);

-- name: GetRecipeStepInstrument :one
SELECT
	recipe_step_instruments.id,
	valid_instruments.id as valid_instrument_id,
	valid_instruments.name as valid_instrument_name,
	valid_instruments.description as valid_instrument_description,
	valid_instruments.icon_path as valid_instrument_icon_path,
	valid_instruments.plural_name as valid_instrument_plural_name,
	valid_instruments.usable_for_storage as valid_instrument_usable_for_storage,
	valid_instruments.slug as valid_instrument_slug,
	valid_instruments.display_in_summary_lists as valid_instrument_display_in_summary_lists,
	valid_instruments.include_in_generated_instructions as valid_instrument_include_in_generated_instructions,
	valid_instruments.last_indexed_at as valid_instrument_last_indexed_at,
	valid_instruments.created_at as valid_instrument_created_at,
	valid_instruments.last_updated_at as valid_instrument_last_updated_at,
	valid_instruments.archived_at as valid_instrument_archived_at,
	recipe_step_instruments.recipe_step_product_id,
	recipe_step_instruments.name,
	recipe_step_instruments.notes,
	recipe_step_instruments.preference_rank,
	recipe_step_instruments.optional,
	recipe_step_instruments.minimum_quantity,
	recipe_step_instruments.maximum_quantity,
	recipe_step_instruments.index,
	recipe_step_instruments.option_index,
	recipe_step_instruments.scale_factor,
	recipe_step_instruments.created_at,
	recipe_step_instruments.last_updated_at,
	recipe_step_instruments.archived_at,
	recipe_step_instruments.belongs_to_recipe_step
FROM recipe_step_instruments
	LEFT JOIN valid_instruments ON recipe_step_instruments.instrument_id=valid_instruments.id
	JOIN recipe_steps ON recipe_step_instruments.belongs_to_recipe_step=recipe_steps.id
	JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
WHERE recipe_step_instruments.archived_at IS NULL
	AND recipe_step_instruments.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
	AND recipe_step_instruments.id = sqlc.arg(recipe_step_instrument_id)
	AND recipe_steps.archived_at IS NULL
	AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipe_steps.id = sqlc.arg(recipe_step_id)
	AND recipes.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id);

-- name: GetRecipeStepInstruments :many
SELECT
	recipe_step_instruments.id,
	valid_instruments.id as valid_instrument_id,
	valid_instruments.name as valid_instrument_name,
	valid_instruments.description as valid_instrument_description,
	valid_instruments.icon_path as valid_instrument_icon_path,
	valid_instruments.plural_name as valid_instrument_plural_name,
	valid_instruments.usable_for_storage as valid_instrument_usable_for_storage,
	valid_instruments.slug as valid_instrument_slug,
	valid_instruments.display_in_summary_lists as valid_instrument_display_in_summary_lists,
	valid_instruments.include_in_generated_instructions as valid_instrument_include_in_generated_instructions,
	valid_instruments.last_indexed_at as valid_instrument_last_indexed_at,
	valid_instruments.created_at as valid_instrument_created_at,
	valid_instruments.last_updated_at as valid_instrument_last_updated_at,
	valid_instruments.archived_at as valid_instrument_archived_at,
	recipe_step_instruments.recipe_step_product_id,
	recipe_step_instruments.name,
	recipe_step_instruments.notes,
	recipe_step_instruments.preference_rank,
	recipe_step_instruments.optional,
	recipe_step_instruments.minimum_quantity,
	recipe_step_instruments.maximum_quantity,
	recipe_step_instruments.index,
	recipe_step_instruments.option_index,
	recipe_step_instruments.scale_factor,
	recipe_step_instruments.created_at,
	recipe_step_instruments.last_updated_at,
	recipe_step_instruments.archived_at,
	recipe_step_instruments.belongs_to_recipe_step,
	(
		SELECT COUNT(recipe_step_instruments.id)
		FROM recipe_step_instruments
		WHERE recipe_step_instruments.archived_at IS NULL
			AND
			recipe_step_instruments.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND recipe_step_instruments.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				recipe_step_instruments.last_updated_at IS NULL
				OR recipe_step_instruments.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				recipe_step_instruments.last_updated_at IS NULL
				OR recipe_step_instruments.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_step_instruments.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(recipe_step_instruments.id)
		FROM recipe_step_instruments
		WHERE recipe_step_instruments.archived_at IS NULL
	) AS total_count
FROM recipe_step_instruments
	LEFT JOIN valid_instruments ON recipe_step_instruments.instrument_id=valid_instruments.id
	JOIN recipe_steps ON recipe_step_instruments.belongs_to_recipe_step=recipe_steps.id
	JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
WHERE
	recipe_step_instruments.archived_at IS NULL
	AND recipe_step_instruments.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
	AND recipe_steps.archived_at IS NULL
	AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipe_steps.id = sqlc.arg(recipe_step_id)
	AND recipes.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id)
	AND recipe_step_instruments.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND recipe_step_instruments.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		recipe_step_instruments.last_updated_at IS NULL
		OR recipe_step_instruments.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		recipe_step_instruments.last_updated_at IS NULL
		OR recipe_step_instruments.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND recipe_step_instruments.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY recipe_step_instruments.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: UpdateRecipeStepInstrument :execrows
UPDATE recipe_step_instruments SET
	instrument_id = sqlc.arg(instrument_id),
	recipe_step_product_id = sqlc.arg(recipe_step_product_id),
	name = sqlc.arg(name),
	notes = sqlc.arg(notes),
	preference_rank = sqlc.arg(preference_rank),
	optional = sqlc.arg(optional),
	minimum_quantity = sqlc.arg(minimum_quantity),
	maximum_quantity = sqlc.arg(maximum_quantity),
	index = sqlc.arg(index),
	option_index = sqlc.arg(option_index),
	scale_factor = sqlc.arg(scale_factor),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND belongs_to_recipe_step = sqlc.arg(belongs_to_recipe_step)
	AND id = sqlc.arg(id);
