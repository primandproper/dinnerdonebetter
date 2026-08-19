-- name: CreateRecipeStepVessel :exec
INSERT INTO recipe_step_vessels (
	id,
	name,
	notes,
	belongs_to_recipe_step,
	recipe_step_product_id,
	valid_vessel_id,
	vessel_predicate,
	minimum_quantity,
	maximum_quantity,
	unavailable_after_step,
	index,
	option_index,
	scale_factor
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(notes),
	sqlc.arg(belongs_to_recipe_step),
	sqlc.arg(recipe_step_product_id),
	sqlc.arg(valid_vessel_id),
	sqlc.arg(vessel_predicate),
	sqlc.arg(minimum_quantity),
	sqlc.arg(maximum_quantity),
	sqlc.arg(unavailable_after_step),
	sqlc.arg(index),
	sqlc.arg(option_index),
	sqlc.arg(scale_factor)
);

-- name: ArchiveRecipeStepVessel :execrows
UPDATE recipe_step_vessels SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_recipe_step = sqlc.arg(belongs_to_recipe_step);

-- name: CheckRecipeStepVesselExistence :one
SELECT EXISTS (
	SELECT recipe_step_vessels.id
	FROM recipe_step_vessels
		JOIN recipe_steps ON recipe_step_vessels.belongs_to_recipe_step=recipe_steps.id
		JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
	WHERE recipe_step_vessels.archived_at IS NULL
		AND recipe_step_vessels.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
		AND recipe_step_vessels.id = sqlc.arg(recipe_step_vessel_id)
		AND recipe_steps.archived_at IS NULL
		AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
		AND recipe_steps.id = sqlc.arg(recipe_step_id)
		AND recipes.archived_at IS NULL
		AND recipes.id = sqlc.arg(recipe_id)
);

-- name: GetRecipeStepVesselsForRecipe :many
SELECT
	recipe_step_vessels.id,
	valid_vessels.id as valid_vessel_id,
	valid_vessels.name as valid_vessel_name,
	valid_vessels.plural_name as valid_vessel_plural_name,
	valid_vessels.description as valid_vessel_description,
	valid_vessels.icon_path as valid_vessel_icon_path,
	valid_vessels.usable_for_storage as valid_vessel_usable_for_storage,
	valid_vessels.slug as valid_vessel_slug,
	valid_vessels.display_in_summary_lists as valid_vessel_display_in_summary_lists,
	valid_vessels.include_in_generated_instructions as valid_vessel_include_in_generated_instructions,
	valid_vessels.capacity as valid_vessel_capacity,
	valid_measurement_units.id as valid_measurement_unit_id,
	valid_measurement_units.name as valid_measurement_unit_name,
	valid_measurement_units.description as valid_measurement_unit_description,
	valid_measurement_units.volumetric as valid_measurement_unit_volumetric,
	valid_measurement_units.icon_path as valid_measurement_unit_icon_path,
	valid_measurement_units.universal as valid_measurement_unit_universal,
	valid_measurement_units.metric as valid_measurement_unit_metric,
	valid_measurement_units.imperial as valid_measurement_unit_imperial,
	valid_measurement_units.slug as valid_measurement_unit_slug,
	valid_measurement_units.plural_name as valid_measurement_unit_plural_name,
	valid_measurement_units.last_indexed_at as valid_measurement_unit_last_indexed_at,
	valid_measurement_units.created_at as valid_measurement_unit_created_at,
	valid_measurement_units.last_updated_at as valid_measurement_unit_last_updated_at,
	valid_measurement_units.archived_at as valid_measurement_unit_archived_at,
	valid_vessels.width_in_millimeters as valid_vessel_width_in_millimeters,
	valid_vessels.length_in_millimeters as valid_vessel_length_in_millimeters,
	valid_vessels.height_in_millimeters as valid_vessel_height_in_millimeters,
	valid_vessels.shape as valid_vessel_shape,
	valid_vessels.last_indexed_at as valid_vessel_last_indexed_at,
	valid_vessels.created_at as valid_vessel_created_at,
	valid_vessels.last_updated_at as valid_vessel_last_updated_at,
	valid_vessels.archived_at as valid_vessel_archived_at,
	recipe_step_vessels.name,
	recipe_step_vessels.notes,
	recipe_step_vessels.belongs_to_recipe_step,
	recipe_step_vessels.recipe_step_product_id,
	recipe_step_vessels.vessel_predicate,
	recipe_step_vessels.minimum_quantity,
	recipe_step_vessels.maximum_quantity,
	recipe_step_vessels.unavailable_after_step,
	recipe_step_vessels.index,
	recipe_step_vessels.option_index,
	recipe_step_vessels.scale_factor,
	recipe_step_vessels.created_at,
	recipe_step_vessels.last_updated_at,
	recipe_step_vessels.archived_at
FROM recipe_step_vessels
	LEFT JOIN valid_vessels ON recipe_step_vessels.valid_vessel_id=valid_vessels.id
	LEFT JOIN valid_measurement_units ON valid_vessels.capacity_unit=valid_measurement_units.id
	JOIN recipe_steps ON recipe_step_vessels.belongs_to_recipe_step=recipe_steps.id
	JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
WHERE recipe_step_vessels.archived_at IS NULL
	AND recipe_steps.archived_at IS NULL
	AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipes.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id);

-- name: GetRecipeStepVessel :one
SELECT
	recipe_step_vessels.id,
	valid_vessels.id as valid_vessel_id,
	valid_vessels.name as valid_vessel_name,
	valid_vessels.plural_name as valid_vessel_plural_name,
	valid_vessels.description as valid_vessel_description,
	valid_vessels.icon_path as valid_vessel_icon_path,
	valid_vessels.usable_for_storage as valid_vessel_usable_for_storage,
	valid_vessels.slug as valid_vessel_slug,
	valid_vessels.display_in_summary_lists as valid_vessel_display_in_summary_lists,
	valid_vessels.include_in_generated_instructions as valid_vessel_include_in_generated_instructions,
	valid_vessels.capacity as valid_vessel_capacity,
	valid_measurement_units.id as valid_measurement_unit_id,
	valid_measurement_units.name as valid_measurement_unit_name,
	valid_measurement_units.description as valid_measurement_unit_description,
	valid_measurement_units.volumetric as valid_measurement_unit_volumetric,
	valid_measurement_units.icon_path as valid_measurement_unit_icon_path,
	valid_measurement_units.universal as valid_measurement_unit_universal,
	valid_measurement_units.metric as valid_measurement_unit_metric,
	valid_measurement_units.imperial as valid_measurement_unit_imperial,
	valid_measurement_units.slug as valid_measurement_unit_slug,
	valid_measurement_units.plural_name as valid_measurement_unit_plural_name,
	valid_measurement_units.last_indexed_at as valid_measurement_unit_last_indexed_at,
	valid_measurement_units.created_at as valid_measurement_unit_created_at,
	valid_measurement_units.last_updated_at as valid_measurement_unit_last_updated_at,
	valid_measurement_units.archived_at as valid_measurement_unit_archived_at,
	valid_vessels.width_in_millimeters as valid_vessel_width_in_millimeters,
	valid_vessels.length_in_millimeters as valid_vessel_length_in_millimeters,
	valid_vessels.height_in_millimeters as valid_vessel_height_in_millimeters,
	valid_vessels.shape as valid_vessel_shape,
	valid_vessels.last_indexed_at as valid_vessel_last_indexed_at,
	valid_vessels.created_at as valid_vessel_created_at,
	valid_vessels.last_updated_at as valid_vessel_last_updated_at,
	valid_vessels.archived_at as valid_vessel_archived_at,
	recipe_step_vessels.name,
	recipe_step_vessels.notes,
	recipe_step_vessels.belongs_to_recipe_step,
	recipe_step_vessels.recipe_step_product_id,
	recipe_step_vessels.vessel_predicate,
	recipe_step_vessels.minimum_quantity,
	recipe_step_vessels.maximum_quantity,
	recipe_step_vessels.unavailable_after_step,
	recipe_step_vessels.index,
	recipe_step_vessels.option_index,
	recipe_step_vessels.scale_factor,
	recipe_step_vessels.created_at,
	recipe_step_vessels.last_updated_at,
	recipe_step_vessels.archived_at
FROM recipe_step_vessels
	LEFT JOIN valid_vessels ON recipe_step_vessels.valid_vessel_id=valid_vessels.id
	LEFT JOIN valid_measurement_units ON valid_vessels.capacity_unit=valid_measurement_units.id
	JOIN recipe_steps ON recipe_step_vessels.belongs_to_recipe_step=recipe_steps.id
	JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
WHERE recipe_step_vessels.archived_at IS NULL
	AND recipe_step_vessels.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
	AND recipe_step_vessels.id = sqlc.arg(recipe_step_vessel_id)
	AND recipe_steps.archived_at IS NULL
	AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipe_steps.id = sqlc.arg(recipe_step_id)
	AND recipes.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id);

-- name: GetRecipeStepVessels :many
SELECT
	recipe_step_vessels.id,
	valid_vessels.id as valid_vessel_id,
	valid_vessels.name as valid_vessel_name,
	valid_vessels.plural_name as valid_vessel_plural_name,
	valid_vessels.description as valid_vessel_description,
	valid_vessels.icon_path as valid_vessel_icon_path,
	valid_vessels.usable_for_storage as valid_vessel_usable_for_storage,
	valid_vessels.slug as valid_vessel_slug,
	valid_vessels.display_in_summary_lists as valid_vessel_display_in_summary_lists,
	valid_vessels.include_in_generated_instructions as valid_vessel_include_in_generated_instructions,
	valid_vessels.capacity as valid_vessel_capacity,
	valid_measurement_units.id as valid_measurement_unit_id,
	valid_measurement_units.name as valid_measurement_unit_name,
	valid_measurement_units.description as valid_measurement_unit_description,
	valid_measurement_units.volumetric as valid_measurement_unit_volumetric,
	valid_measurement_units.icon_path as valid_measurement_unit_icon_path,
	valid_measurement_units.universal as valid_measurement_unit_universal,
	valid_measurement_units.metric as valid_measurement_unit_metric,
	valid_measurement_units.imperial as valid_measurement_unit_imperial,
	valid_measurement_units.slug as valid_measurement_unit_slug,
	valid_measurement_units.plural_name as valid_measurement_unit_plural_name,
	valid_measurement_units.last_indexed_at as valid_measurement_unit_last_indexed_at,
	valid_measurement_units.created_at as valid_measurement_unit_created_at,
	valid_measurement_units.last_updated_at as valid_measurement_unit_last_updated_at,
	valid_measurement_units.archived_at as valid_measurement_unit_archived_at,
	valid_vessels.width_in_millimeters as valid_vessel_width_in_millimeters,
	valid_vessels.length_in_millimeters as valid_vessel_length_in_millimeters,
	valid_vessels.height_in_millimeters as valid_vessel_height_in_millimeters,
	valid_vessels.shape as valid_vessel_shape,
	valid_vessels.last_indexed_at as valid_vessel_last_indexed_at,
	valid_vessels.created_at as valid_vessel_created_at,
	valid_vessels.last_updated_at as valid_vessel_last_updated_at,
	valid_vessels.archived_at as valid_vessel_archived_at,
	recipe_step_vessels.name,
	recipe_step_vessels.notes,
	recipe_step_vessels.belongs_to_recipe_step,
	recipe_step_vessels.recipe_step_product_id,
	recipe_step_vessels.vessel_predicate,
	recipe_step_vessels.minimum_quantity,
	recipe_step_vessels.maximum_quantity,
	recipe_step_vessels.unavailable_after_step,
	recipe_step_vessels.index,
	recipe_step_vessels.option_index,
	recipe_step_vessels.scale_factor,
	recipe_step_vessels.created_at,
	recipe_step_vessels.last_updated_at,
	recipe_step_vessels.archived_at,
	(
		SELECT COUNT(recipe_step_vessels.id)
		FROM recipe_step_vessels
		WHERE recipe_step_vessels.archived_at IS NULL
			AND
			recipe_step_vessels.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND recipe_step_vessels.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				recipe_step_vessels.last_updated_at IS NULL
				OR recipe_step_vessels.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				recipe_step_vessels.last_updated_at IS NULL
				OR recipe_step_vessels.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_step_vessels.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(recipe_step_vessels.id)
		FROM recipe_step_vessels
		WHERE recipe_step_vessels.archived_at IS NULL
	) AS total_count
FROM recipe_step_vessels
	 LEFT JOIN valid_vessels ON recipe_step_vessels.valid_vessel_id=valid_vessels.id
	 LEFT JOIN valid_measurement_units ON valid_vessels.capacity_unit=valid_measurement_units.id
	 JOIN recipe_steps ON recipe_step_vessels.belongs_to_recipe_step=recipe_steps.id
	 JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
WHERE recipe_step_vessels.archived_at IS NULL
	AND recipe_step_vessels.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
	AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipe_steps.archived_at IS NULL
	AND recipe_steps.id = sqlc.arg(recipe_step_id)
	AND recipes.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id)
	AND recipe_step_vessels.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND recipe_step_vessels.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		recipe_step_vessels.last_updated_at IS NULL
		OR recipe_step_vessels.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		recipe_step_vessels.last_updated_at IS NULL
		OR recipe_step_vessels.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND recipe_step_vessels.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY recipe_step_vessels.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: UpdateRecipeStepVessel :execrows
UPDATE recipe_step_vessels SET
	name = sqlc.arg(name),
	notes = sqlc.arg(notes),
	belongs_to_recipe_step = sqlc.arg(belongs_to_recipe_step),
	recipe_step_product_id = sqlc.arg(recipe_step_product_id),
	valid_vessel_id = sqlc.arg(valid_vessel_id),
	vessel_predicate = sqlc.arg(vessel_predicate),
	minimum_quantity = sqlc.arg(minimum_quantity),
	maximum_quantity = sqlc.arg(maximum_quantity),
	unavailable_after_step = sqlc.arg(unavailable_after_step),
	index = sqlc.arg(index),
	option_index = sqlc.arg(option_index),
	scale_factor = sqlc.arg(scale_factor),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND belongs_to_recipe_step = sqlc.arg(belongs_to_recipe_step)
	AND id = sqlc.arg(id);
