-- name: CreateRecipeStepIngredient :exec
INSERT INTO recipe_step_ingredients (
	id,
	name,
	optional,
	ingredient_id,
	measurement_unit,
	minimum_quantity_value,
	maximum_quantity_value,
	quantity_notes,
	recipe_step_product_id,
	ingredient_notes,
	index,
	option_index,
	to_taste,
	product_percentage_to_use,
	vessel_index,
	scale_factor,
	recipe_step_product_recipe_id,
	belongs_to_recipe_step
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(optional),
	sqlc.arg(ingredient_id),
	sqlc.arg(measurement_unit),
	sqlc.arg(minimum_quantity_value),
	sqlc.arg(maximum_quantity_value),
	sqlc.arg(quantity_notes),
	sqlc.arg(recipe_step_product_id),
	sqlc.arg(ingredient_notes),
	sqlc.arg(index),
	sqlc.arg(option_index),
	sqlc.arg(to_taste),
	sqlc.arg(product_percentage_to_use),
	sqlc.arg(vessel_index),
	sqlc.arg(scale_factor),
	sqlc.arg(recipe_step_product_recipe_id),
	sqlc.arg(belongs_to_recipe_step)
);

-- name: UpdateRecipeStepIngredient :execrows
UPDATE recipe_step_ingredients SET
	name = sqlc.arg(name),
	optional = sqlc.arg(optional),
	ingredient_id = sqlc.arg(ingredient_id),
	measurement_unit = sqlc.arg(measurement_unit),
	minimum_quantity_value = sqlc.arg(minimum_quantity_value),
	maximum_quantity_value = sqlc.arg(maximum_quantity_value),
	quantity_notes = sqlc.arg(quantity_notes),
	recipe_step_product_id = sqlc.arg(recipe_step_product_id),
	ingredient_notes = sqlc.arg(ingredient_notes),
	index = sqlc.arg(index),
	option_index = sqlc.arg(option_index),
	to_taste = sqlc.arg(to_taste),
	product_percentage_to_use = sqlc.arg(product_percentage_to_use),
	vessel_index = sqlc.arg(vessel_index),
	scale_factor = sqlc.arg(scale_factor),
	recipe_step_product_recipe_id = sqlc.arg(recipe_step_product_recipe_id),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_recipe_step = sqlc.arg(belongs_to_recipe_step);

-- name: ArchiveRecipeStepIngredient :execrows
UPDATE recipe_step_ingredients SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_recipe_step = sqlc.arg(belongs_to_recipe_step);

-- name: CheckRecipeStepIngredientExistence :one
SELECT EXISTS (
	SELECT recipe_step_ingredients.id
	FROM recipe_step_ingredients
		JOIN recipe_steps ON recipe_step_ingredients.belongs_to_recipe_step=recipe_steps.id
		JOIN recipes ON recipe_steps.belongs_to_recipe=recipes.id
	WHERE recipe_step_ingredients.archived_at IS NULL
		AND recipe_step_ingredients.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
		AND recipe_step_ingredients.id = sqlc.arg(recipe_step_ingredient_id)
		AND recipe_steps.archived_at IS NULL
		AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
		AND recipe_steps.id = sqlc.arg(recipe_step_id)
		AND recipes.archived_at IS NULL
		AND recipes.id = sqlc.arg(recipe_id)
);

-- name: GetAllRecipeStepIngredientsForRecipe :many
SELECT
	recipe_step_ingredients.id,
	recipe_step_ingredients.name,
	recipe_step_ingredients.optional,
	valid_ingredients.id as valid_ingredient_id,
	valid_ingredients.name as valid_ingredient_name,
	valid_ingredients.description as valid_ingredient_description,
	valid_ingredients.warning as valid_ingredient_warning,
	valid_ingredients.contains_egg as valid_ingredient_contains_egg,
	valid_ingredients.contains_dairy as valid_ingredient_contains_dairy,
	valid_ingredients.contains_peanut as valid_ingredient_contains_peanut,
	valid_ingredients.contains_tree_nut as valid_ingredient_contains_tree_nut,
	valid_ingredients.contains_soy as valid_ingredient_contains_soy,
	valid_ingredients.contains_wheat as valid_ingredient_contains_wheat,
	valid_ingredients.contains_shellfish as valid_ingredient_contains_shellfish,
	valid_ingredients.contains_sesame as valid_ingredient_contains_sesame,
	valid_ingredients.contains_fish as valid_ingredient_contains_fish,
	valid_ingredients.contains_gluten as valid_ingredient_contains_gluten,
	valid_ingredients.animal_flesh as valid_ingredient_animal_flesh,
	valid_ingredients.is_liquid as valid_ingredient_is_liquid,
	valid_ingredients.icon_path as valid_ingredient_icon_path,
	valid_ingredients.animal_derived as valid_ingredient_animal_derived,
	valid_ingredients.plural_name as valid_ingredient_plural_name,
	valid_ingredients.restrict_to_preparations as valid_ingredient_restrict_to_preparations,
	valid_ingredients.contaminates_equipment as valid_ingredient_contaminates_equipment,
	valid_ingredients.minimum_ideal_storage_temperature_in_celsius as valid_ingredient_minimum_ideal_storage_temperature_in_celsius,
	valid_ingredients.maximum_ideal_storage_temperature_in_celsius as valid_ingredient_maximum_ideal_storage_temperature_in_celsius,
	valid_ingredients.storage_instructions as valid_ingredient_storage_instructions,
	valid_ingredients.slug as valid_ingredient_slug,
	valid_ingredients.contains_alcohol as valid_ingredient_contains_alcohol,
	valid_ingredients.shopping_suggestions as valid_ingredient_shopping_suggestions,
	valid_ingredients.is_starch as valid_ingredient_is_starch,
	valid_ingredients.is_protein as valid_ingredient_is_protein,
	valid_ingredients.is_grain as valid_ingredient_is_grain,
	valid_ingredients.is_fruit as valid_ingredient_is_fruit,
	valid_ingredients.is_salt as valid_ingredient_is_salt,
	valid_ingredients.is_fat as valid_ingredient_is_fat,
	valid_ingredients.is_acid as valid_ingredient_is_acid,
	valid_ingredients.is_heat as valid_ingredient_is_heat,
	valid_ingredients.last_indexed_at as valid_ingredient_last_indexed_at,
	valid_ingredients.created_at as valid_ingredient_created_at,
	valid_ingredients.last_updated_at as valid_ingredient_last_updated_at,
	valid_ingredients.archived_at as valid_ingredient_archived_at,
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
	recipe_step_ingredients.minimum_quantity_value,
	recipe_step_ingredients.maximum_quantity_value,
	recipe_step_ingredients.quantity_notes,
	recipe_step_ingredients.recipe_step_product_id,
	recipe_step_ingredients.ingredient_notes,
	recipe_step_ingredients.index,
	recipe_step_ingredients.option_index,
	recipe_step_ingredients.to_taste,
	recipe_step_ingredients.product_percentage_to_use,
	recipe_step_ingredients.vessel_index,
	recipe_step_ingredients.scale_factor,
	recipe_step_ingredients.created_at,
	recipe_step_ingredients.last_updated_at,
	recipe_step_ingredients.archived_at,
	recipe_step_ingredients.recipe_step_product_recipe_id,
	recipe_step_ingredients.belongs_to_recipe_step
FROM recipe_step_ingredients
	JOIN recipe_steps ON recipe_step_ingredients.belongs_to_recipe_step = recipe_steps.id
	JOIN recipes ON recipe_steps.belongs_to_recipe = recipes.id
	LEFT JOIN valid_ingredients ON recipe_step_ingredients.ingredient_id = valid_ingredients.id
	JOIN valid_measurement_units ON recipe_step_ingredients.measurement_unit = valid_measurement_units.id
WHERE
	recipe_step_ingredients.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id)
	AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id);

-- name: GetRecipeStepIngredients :many
SELECT
	recipe_step_ingredients.id,
	recipe_step_ingredients.name,
	recipe_step_ingredients.optional,
	valid_ingredients.id as valid_ingredient_id,
	valid_ingredients.name as valid_ingredient_name,
	valid_ingredients.description as valid_ingredient_description,
	valid_ingredients.warning as valid_ingredient_warning,
	valid_ingredients.contains_egg as valid_ingredient_contains_egg,
	valid_ingredients.contains_dairy as valid_ingredient_contains_dairy,
	valid_ingredients.contains_peanut as valid_ingredient_contains_peanut,
	valid_ingredients.contains_tree_nut as valid_ingredient_contains_tree_nut,
	valid_ingredients.contains_soy as valid_ingredient_contains_soy,
	valid_ingredients.contains_wheat as valid_ingredient_contains_wheat,
	valid_ingredients.contains_shellfish as valid_ingredient_contains_shellfish,
	valid_ingredients.contains_sesame as valid_ingredient_contains_sesame,
	valid_ingredients.contains_fish as valid_ingredient_contains_fish,
	valid_ingredients.contains_gluten as valid_ingredient_contains_gluten,
	valid_ingredients.animal_flesh as valid_ingredient_animal_flesh,
	valid_ingredients.is_liquid as valid_ingredient_is_liquid,
	valid_ingredients.icon_path as valid_ingredient_icon_path,
	valid_ingredients.animal_derived as valid_ingredient_animal_derived,
	valid_ingredients.plural_name as valid_ingredient_plural_name,
	valid_ingredients.restrict_to_preparations as valid_ingredient_restrict_to_preparations,
	valid_ingredients.contaminates_equipment as valid_ingredient_contaminates_equipment,
	valid_ingredients.minimum_ideal_storage_temperature_in_celsius as valid_ingredient_minimum_ideal_storage_temperature_in_celsius,
	valid_ingredients.maximum_ideal_storage_temperature_in_celsius as valid_ingredient_maximum_ideal_storage_temperature_in_celsius,
	valid_ingredients.storage_instructions as valid_ingredient_storage_instructions,
	valid_ingredients.slug as valid_ingredient_slug,
	valid_ingredients.contains_alcohol as valid_ingredient_contains_alcohol,
	valid_ingredients.shopping_suggestions as valid_ingredient_shopping_suggestions,
	valid_ingredients.is_starch as valid_ingredient_is_starch,
	valid_ingredients.is_protein as valid_ingredient_is_protein,
	valid_ingredients.is_grain as valid_ingredient_is_grain,
	valid_ingredients.is_fruit as valid_ingredient_is_fruit,
	valid_ingredients.is_salt as valid_ingredient_is_salt,
	valid_ingredients.is_fat as valid_ingredient_is_fat,
	valid_ingredients.is_acid as valid_ingredient_is_acid,
	valid_ingredients.is_heat as valid_ingredient_is_heat,
	valid_ingredients.last_indexed_at as valid_ingredient_last_indexed_at,
	valid_ingredients.created_at as valid_ingredient_created_at,
	valid_ingredients.last_updated_at as valid_ingredient_last_updated_at,
	valid_ingredients.archived_at as valid_ingredient_archived_at,
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
	recipe_step_ingredients.minimum_quantity_value,
	recipe_step_ingredients.maximum_quantity_value,
	recipe_step_ingredients.quantity_notes,
	recipe_step_ingredients.recipe_step_product_id,
	recipe_step_ingredients.ingredient_notes,
	recipe_step_ingredients.index,
	recipe_step_ingredients.option_index,
	recipe_step_ingredients.to_taste,
	recipe_step_ingredients.product_percentage_to_use,
	recipe_step_ingredients.vessel_index,
	recipe_step_ingredients.scale_factor,
	recipe_step_ingredients.created_at,
	recipe_step_ingredients.last_updated_at,
	recipe_step_ingredients.archived_at,
	recipe_step_ingredients.recipe_step_product_recipe_id,
	recipe_step_ingredients.belongs_to_recipe_step,
	(
		SELECT COUNT(recipe_step_ingredients.id)
		FROM recipe_step_ingredients
		JOIN recipe_steps ON recipe_step_ingredients.belongs_to_recipe_step = recipe_steps.id
	JOIN recipes ON recipe_steps.belongs_to_recipe = recipes.id
		WHERE recipe_step_ingredients.archived_at IS NULL
			AND
			recipe_step_ingredients.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND recipe_step_ingredients.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				recipe_step_ingredients.last_updated_at IS NULL
				OR recipe_step_ingredients.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				recipe_step_ingredients.last_updated_at IS NULL
				OR recipe_step_ingredients.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR recipe_step_ingredients.archived_at IS NULL)
			AND recipes.id = sqlc.arg(recipe_id)
			AND recipe_steps.id = sqlc.arg(recipe_step_id)
			AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
			AND recipe_step_ingredients.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
	) AS filtered_count,
	(
		SELECT COUNT(recipe_step_ingredients.id)
		FROM recipe_step_ingredients
		JOIN recipe_steps ON recipe_step_ingredients.belongs_to_recipe_step = recipe_steps.id
	JOIN recipes ON recipe_steps.belongs_to_recipe = recipes.id
		WHERE recipe_step_ingredients.archived_at IS NULL
			AND recipes.id = sqlc.arg(recipe_id)
			AND recipe_steps.id = sqlc.arg(recipe_step_id)
			AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
			AND recipe_step_ingredients.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
	) AS total_count
FROM recipe_step_ingredients
	JOIN recipe_steps ON recipe_step_ingredients.belongs_to_recipe_step = recipe_steps.id
	JOIN recipes ON recipe_steps.belongs_to_recipe = recipes.id
	LEFT JOIN valid_ingredients ON recipe_step_ingredients.ingredient_id = valid_ingredients.id
	JOIN valid_measurement_units ON recipe_step_ingredients.measurement_unit = valid_measurement_units.id
WHERE
	recipe_step_ingredients.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id)
	AND recipe_steps.id = sqlc.arg(recipe_step_id)
	AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipe_step_ingredients.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
	AND recipe_step_ingredients.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND recipe_step_ingredients.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		recipe_step_ingredients.last_updated_at IS NULL
		OR recipe_step_ingredients.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		recipe_step_ingredients.last_updated_at IS NULL
		OR recipe_step_ingredients.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND recipe_step_ingredients.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY recipe_step_ingredients.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetRecipeStepIngredient :one
SELECT
	recipe_step_ingredients.id,
	recipe_step_ingredients.name,
	recipe_step_ingredients.optional,
	valid_ingredients.id as valid_ingredient_id,
	valid_ingredients.name as valid_ingredient_name,
	valid_ingredients.description as valid_ingredient_description,
	valid_ingredients.warning as valid_ingredient_warning,
	valid_ingredients.contains_egg as valid_ingredient_contains_egg,
	valid_ingredients.contains_dairy as valid_ingredient_contains_dairy,
	valid_ingredients.contains_peanut as valid_ingredient_contains_peanut,
	valid_ingredients.contains_tree_nut as valid_ingredient_contains_tree_nut,
	valid_ingredients.contains_soy as valid_ingredient_contains_soy,
	valid_ingredients.contains_wheat as valid_ingredient_contains_wheat,
	valid_ingredients.contains_shellfish as valid_ingredient_contains_shellfish,
	valid_ingredients.contains_sesame as valid_ingredient_contains_sesame,
	valid_ingredients.contains_fish as valid_ingredient_contains_fish,
	valid_ingredients.contains_gluten as valid_ingredient_contains_gluten,
	valid_ingredients.animal_flesh as valid_ingredient_animal_flesh,
	valid_ingredients.is_liquid as valid_ingredient_is_liquid,
	valid_ingredients.icon_path as valid_ingredient_icon_path,
	valid_ingredients.animal_derived as valid_ingredient_animal_derived,
	valid_ingredients.plural_name as valid_ingredient_plural_name,
	valid_ingredients.restrict_to_preparations as valid_ingredient_restrict_to_preparations,
	valid_ingredients.contaminates_equipment as valid_ingredient_contaminates_equipment,
	valid_ingredients.minimum_ideal_storage_temperature_in_celsius as valid_ingredient_minimum_ideal_storage_temperature_in_celsius,
	valid_ingredients.maximum_ideal_storage_temperature_in_celsius as valid_ingredient_maximum_ideal_storage_temperature_in_celsius,
	valid_ingredients.storage_instructions as valid_ingredient_storage_instructions,
	valid_ingredients.slug as valid_ingredient_slug,
	valid_ingredients.contains_alcohol as valid_ingredient_contains_alcohol,
	valid_ingredients.shopping_suggestions as valid_ingredient_shopping_suggestions,
	valid_ingredients.is_starch as valid_ingredient_is_starch,
	valid_ingredients.is_protein as valid_ingredient_is_protein,
	valid_ingredients.is_grain as valid_ingredient_is_grain,
	valid_ingredients.is_fruit as valid_ingredient_is_fruit,
	valid_ingredients.is_salt as valid_ingredient_is_salt,
	valid_ingredients.is_fat as valid_ingredient_is_fat,
	valid_ingredients.is_acid as valid_ingredient_is_acid,
	valid_ingredients.is_heat as valid_ingredient_is_heat,
	valid_ingredients.last_indexed_at as valid_ingredient_last_indexed_at,
	valid_ingredients.created_at as valid_ingredient_created_at,
	valid_ingredients.last_updated_at as valid_ingredient_last_updated_at,
	valid_ingredients.archived_at as valid_ingredient_archived_at,
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
	recipe_step_ingredients.minimum_quantity_value,
	recipe_step_ingredients.maximum_quantity_value,
	recipe_step_ingredients.quantity_notes,
	recipe_step_ingredients.recipe_step_product_id,
	recipe_step_ingredients.ingredient_notes,
	recipe_step_ingredients.index,
	recipe_step_ingredients.option_index,
	recipe_step_ingredients.to_taste,
	recipe_step_ingredients.product_percentage_to_use,
	recipe_step_ingredients.vessel_index,
	recipe_step_ingredients.scale_factor,
	recipe_step_ingredients.created_at,
	recipe_step_ingredients.last_updated_at,
	recipe_step_ingredients.archived_at,
	recipe_step_ingredients.recipe_step_product_recipe_id,
	recipe_step_ingredients.belongs_to_recipe_step
FROM recipe_step_ingredients
	JOIN recipe_steps ON recipe_step_ingredients.belongs_to_recipe_step = recipe_steps.id
	JOIN recipes ON recipe_steps.belongs_to_recipe = recipes.id
	LEFT JOIN valid_ingredients ON recipe_step_ingredients.ingredient_id = valid_ingredients.id
	JOIN valid_measurement_units ON recipe_step_ingredients.measurement_unit = valid_measurement_units.id
WHERE recipe_step_ingredients.archived_at IS NULL
	AND recipe_step_ingredients.belongs_to_recipe_step = sqlc.arg(recipe_step_id)
	AND recipe_step_ingredients.id = sqlc.arg(recipe_step_ingredient_id)
	AND recipe_steps.archived_at IS NULL
	AND recipe_steps.belongs_to_recipe = sqlc.arg(recipe_id)
	AND recipe_steps.id = sqlc.arg(recipe_step_id)
	AND recipes.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id);
