-- name: CreateRecipeStepProduct :exec
INSERT INTO recipe_step_products (
	id,
	name,
	type,
	measurement_unit,
	minimum_measurement_quantity_value,
	maximum_measurement_quantity_value,
	minimum_item_quantity_value,
	maximum_item_quantity_value,
	quantity_notes,
	compostable,
	maximum_storage_duration_in_seconds,
	minimum_storage_temperature_in_celsius,
	maximum_storage_temperature_in_celsius,
	storage_instructions,
	is_liquid,
	is_waste,
	index,
	contained_in_vessel_index,
	belongs_to_recipe_step
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(type),
	sqlc.arg(measurement_unit),
	sqlc.arg(minimum_measurement_quantity_value),
	sqlc.arg(maximum_measurement_quantity_value),
	sqlc.arg(minimum_item_quantity_value),
	sqlc.arg(maximum_item_quantity_value),
	sqlc.arg(quantity_notes),
	sqlc.arg(compostable),
	sqlc.arg(maximum_storage_duration_in_seconds),
	sqlc.arg(minimum_storage_temperature_in_celsius),
	sqlc.arg(maximum_storage_temperature_in_celsius),
	sqlc.arg(storage_instructions),
	sqlc.arg(is_liquid),
	sqlc.arg(is_waste),
	sqlc.arg(index),
	sqlc.arg(contained_in_vessel_index),
	sqlc.arg(belongs_to_recipe_step)
);

-- name: UpdateRecipeStepProduct :execrows
UPDATE recipe_step_products SET
	name = sqlc.arg(name),
	type = sqlc.arg(type),
	measurement_unit = sqlc.arg(measurement_unit),
	minimum_measurement_quantity_value = sqlc.arg(minimum_measurement_quantity_value),
	maximum_measurement_quantity_value = sqlc.arg(maximum_measurement_quantity_value),
	minimum_item_quantity_value = sqlc.arg(minimum_item_quantity_value),
	maximum_item_quantity_value = sqlc.arg(maximum_item_quantity_value),
	quantity_notes = sqlc.arg(quantity_notes),
	compostable = sqlc.arg(compostable),
	maximum_storage_duration_in_seconds = sqlc.arg(maximum_storage_duration_in_seconds),
	minimum_storage_temperature_in_celsius = sqlc.arg(minimum_storage_temperature_in_celsius),
	maximum_storage_temperature_in_celsius = sqlc.arg(maximum_storage_temperature_in_celsius),
	storage_instructions = sqlc.arg(storage_instructions),
	is_liquid = sqlc.arg(is_liquid),
	is_waste = sqlc.arg(is_waste),
	index = sqlc.arg(index),
	contained_in_vessel_index = sqlc.arg(contained_in_vessel_index),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_recipe_step = sqlc.arg(belongs_to_recipe_step);

-- name: ArchiveRecipeStepProduct :execrows
UPDATE recipe_step_products SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_recipe_step = sqlc.arg(belongs_to_recipe_step);

-- name: CheckRecipeStepProductExistence :one
SELECT EXISTS (
	SELECT recipe_step_products.id
	FROM recipe_step_products
		JOIN recipe_steps ON recipe_step_products.belongs_to_recipe_step=recipe_steps.id
		JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
	WHERE recipe_step_products.archived_at IS NULL
		AND recipe_step_products.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
		AND recipe_step_products.id = sqlc.arg(recipe_step_product_id)
		AND recipe_steps.archived_at IS NULL
		AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
		AND recipe_steps.id = sqlc.arg(recipe_step_id)
		AND recipes.archived_at IS NULL
		AND recipes.id = sqlc.arg(recipe_id)
);

-- name: GetRecipeStepProductsForRecipe :many
SELECT
	recipe_step_products.id,
	recipe_step_products.name,
	recipe_step_products.type,
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
	recipe_step_products.minimum_measurement_quantity_value,
	recipe_step_products.maximum_measurement_quantity_value,
	recipe_step_products.minimum_item_quantity_value,
	recipe_step_products.maximum_item_quantity_value,
	recipe_step_products.quantity_notes,
	recipe_step_products.compostable,
	recipe_step_products.maximum_storage_duration_in_seconds,
	recipe_step_products.minimum_storage_temperature_in_celsius,
	recipe_step_products.maximum_storage_temperature_in_celsius,
	recipe_step_products.storage_instructions,
	recipe_step_products.is_liquid,
	recipe_step_products.is_waste,
	recipe_step_products.index,
	recipe_step_products.contained_in_vessel_index,
	recipe_step_products.created_at,
	recipe_step_products.last_updated_at,
	recipe_step_products.archived_at,
	recipe_step_products.belongs_to_recipe_step
FROM recipe_step_products
	JOIN recipe_steps ON recipe_step_products.belongs_to_recipe_step=recipe_steps.id
	JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
	LEFT JOIN valid_measurement_units ON recipe_step_products.measurement_unit=valid_measurement_units.id
WHERE recipe_step_products.archived_at IS NULL
	AND recipe_steps.archived_at IS NULL
	AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipes.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id);

-- name: GetRecipeStepProducts :many
SELECT
	recipe_step_products.id,
	recipe_step_products.name,
	recipe_step_products.type,
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
	recipe_step_products.minimum_measurement_quantity_value,
	recipe_step_products.maximum_measurement_quantity_value,
	recipe_step_products.minimum_item_quantity_value,
	recipe_step_products.maximum_item_quantity_value,
	recipe_step_products.quantity_notes,
	recipe_step_products.compostable,
	recipe_step_products.maximum_storage_duration_in_seconds,
	recipe_step_products.minimum_storage_temperature_in_celsius,
	recipe_step_products.maximum_storage_temperature_in_celsius,
	recipe_step_products.storage_instructions,
	recipe_step_products.is_liquid,
	recipe_step_products.is_waste,
	recipe_step_products.index,
	recipe_step_products.contained_in_vessel_index,
	recipe_step_products.created_at,
	recipe_step_products.last_updated_at,
	recipe_step_products.archived_at,
	recipe_step_products.belongs_to_recipe_step,
	(
		SELECT COUNT(recipe_step_products.id)
		FROM recipe_step_products
		WHERE recipe_step_products.archived_at IS NULL
			AND
			recipe_step_products.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND recipe_step_products.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				recipe_step_products.last_updated_at IS NULL
				OR recipe_step_products.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				recipe_step_products.last_updated_at IS NULL
				OR recipe_step_products.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_step_products.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(recipe_step_products.id)
		FROM recipe_step_products
		WHERE recipe_step_products.archived_at IS NULL
				AND recipe_step_products.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
	) AS total_count
FROM recipe_step_products
	JOIN recipe_steps ON recipe_step_products.belongs_to_recipe_step=recipe_steps.id
	JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
	LEFT JOIN valid_measurement_units ON recipe_step_products.measurement_unit=valid_measurement_units.id
WHERE recipe_step_products.archived_at IS NULL
	AND recipe_step_products.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
	AND recipe_steps.archived_at IS NULL
	AND recipe_steps.id = sqlc.arg(recipe_step_id)
	AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipes.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id)
	AND recipe_step_products.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND recipe_step_products.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		recipe_step_products.last_updated_at IS NULL
		OR recipe_step_products.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		recipe_step_products.last_updated_at IS NULL
		OR recipe_step_products.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_step_products.archived_at IS NULL)
	AND recipe_step_products.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY recipe_step_products.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetRecipeStepProduct :one
SELECT
	recipe_step_products.id,
	recipe_step_products.name,
	recipe_step_products.type,
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
	recipe_step_products.minimum_measurement_quantity_value,
	recipe_step_products.maximum_measurement_quantity_value,
	recipe_step_products.minimum_item_quantity_value,
	recipe_step_products.maximum_item_quantity_value,
	recipe_step_products.quantity_notes,
	recipe_step_products.compostable,
	recipe_step_products.maximum_storage_duration_in_seconds,
	recipe_step_products.minimum_storage_temperature_in_celsius,
	recipe_step_products.maximum_storage_temperature_in_celsius,
	recipe_step_products.storage_instructions,
	recipe_step_products.is_liquid,
	recipe_step_products.is_waste,
	recipe_step_products.index,
	recipe_step_products.contained_in_vessel_index,
	recipe_step_products.created_at,
	recipe_step_products.last_updated_at,
	recipe_step_products.archived_at,
	recipe_step_products.belongs_to_recipe_step
FROM recipe_step_products
	JOIN recipe_steps ON recipe_step_products.belongs_to_recipe_step=recipe_steps.id
	JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
	LEFT JOIN valid_measurement_units ON recipe_step_products.measurement_unit=valid_measurement_units.id
WHERE recipe_step_products.archived_at IS NULL
	AND recipe_step_products.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
	AND recipe_step_products.id = sqlc.arg(recipe_step_product_id)
	AND recipe_steps.archived_at IS NULL
	AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipe_steps.id = sqlc.arg(recipe_step_id)
	AND recipes.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id);
