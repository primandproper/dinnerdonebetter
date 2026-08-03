-- name: DeleteMealPlanGroceryListItems :exec
DELETE FROM meal_plan_grocery_list_items WHERE id = ANY(sqlc.arg(ids)::text[]);

-- name: ArchiveMealPlanGroceryListItem :execrows
UPDATE meal_plan_grocery_list_items SET archived_at = NOW() WHERE archived_at IS NULL AND id = sqlc.arg(id);

-- name: CreateMealPlanGroceryListItem :exec
INSERT INTO meal_plan_grocery_list_items (
	id,
	belongs_to_meal_plan,
	valid_ingredient,
	valid_measurement_unit,
	minimum_quantity_needed,
	maximum_quantity_needed,
	quantity_purchased,
	purchased_measurement_unit,
	purchased_upc,
	purchase_price,
	status_explanation,
	status,
	belongs_to_meal_plan_option,
	recipe_id,
	recipe_step_id,
	ingredient_index,
	option_index
) VALUES (
	sqlc.arg(id),
	sqlc.arg(belongs_to_meal_plan),
	sqlc.arg(valid_ingredient),
	sqlc.arg(valid_measurement_unit),
	sqlc.arg(minimum_quantity_needed),
	sqlc.arg(maximum_quantity_needed),
	sqlc.arg(quantity_purchased),
	sqlc.arg(purchased_measurement_unit),
	sqlc.arg(purchased_upc),
	sqlc.arg(purchase_price),
	sqlc.arg(status_explanation),
	sqlc.arg(status),
	sqlc.arg(belongs_to_meal_plan_option),
	sqlc.arg(recipe_id),
	sqlc.arg(recipe_step_id),
	sqlc.arg(ingredient_index),
	sqlc.arg(option_index)
);

-- name: CheckMealPlanGroceryListItemExistence :one
SELECT EXISTS (
	SELECT meal_plan_grocery_list_items.id
	FROM meal_plan_grocery_list_items
	WHERE meal_plan_grocery_list_items.archived_at IS NULL
		AND meal_plan_grocery_list_items.id = sqlc.arg(meal_plan_grocery_list_item_id)
		AND meal_plan_grocery_list_items.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
);

-- name: GetMealPlanGroceryListItemsForMealPlan :many
SELECT
	meal_plan_grocery_list_items.id,
	meal_plan_grocery_list_items.belongs_to_meal_plan,
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
	meal_plan_grocery_list_items.minimum_quantity_needed,
	meal_plan_grocery_list_items.maximum_quantity_needed,
	meal_plan_grocery_list_items.quantity_purchased,
	meal_plan_grocery_list_items.purchased_measurement_unit,
	meal_plan_grocery_list_items.purchased_upc,
	meal_plan_grocery_list_items.purchase_price,
	meal_plan_grocery_list_items.status_explanation,
	meal_plan_grocery_list_items.status,
	meal_plan_grocery_list_items.created_at,
	meal_plan_grocery_list_items.last_updated_at,
	meal_plan_grocery_list_items.archived_at,
	meal_plan_grocery_list_items.belongs_to_meal_plan_option,
	meal_plan_grocery_list_items.recipe_id,
	meal_plan_grocery_list_items.recipe_step_id,
	meal_plan_grocery_list_items.ingredient_index,
	meal_plan_grocery_list_items.option_index,
	(
		SELECT COUNT(meal_plan_grocery_list_items.id)
		FROM meal_plan_grocery_list_items
		JOIN meal_plans ON meal_plan_grocery_list_items.belongs_to_meal_plan = meal_plans.id
	JOIN valid_ingredients ON meal_plan_grocery_list_items.valid_ingredient = valid_ingredients.id
	JOIN valid_measurement_units ON meal_plan_grocery_list_items.valid_measurement_unit = valid_measurement_units.id
		WHERE meal_plan_grocery_list_items.archived_at IS NULL
			AND
			meal_plan_grocery_list_items.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND meal_plan_grocery_list_items.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				meal_plan_grocery_list_items.last_updated_at IS NULL
				OR meal_plan_grocery_list_items.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				meal_plan_grocery_list_items.last_updated_at IS NULL
				OR meal_plan_grocery_list_items.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR meal_plan_grocery_list_items.archived_at IS NULL)
			AND valid_measurement_units.archived_at IS NULL
			AND valid_ingredients.archived_at IS NULL
			AND meal_plans.archived_at IS NULL
			AND meal_plan_grocery_list_items.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
			AND (meal_plan_grocery_list_items.belongs_to_meal_plan_option IS NULL OR NOT EXISTS (SELECT 1 FROM meal_plan_options o JOIN meal_plan_events e ON o.belongs_to_meal_plan_event = e.id WHERE o.id = meal_plan_grocery_list_items.belongs_to_meal_plan_option AND e.archived_at IS NOT NULL))
	) AS filtered_count,
	(
		SELECT COUNT(meal_plan_grocery_list_items.id)
		FROM meal_plan_grocery_list_items
		JOIN meal_plans ON meal_plan_grocery_list_items.belongs_to_meal_plan = meal_plans.id
	JOIN valid_ingredients ON meal_plan_grocery_list_items.valid_ingredient = valid_ingredients.id
	JOIN valid_measurement_units ON meal_plan_grocery_list_items.valid_measurement_unit = valid_measurement_units.id
		WHERE meal_plan_grocery_list_items.archived_at IS NULL
			AND valid_measurement_units.archived_at IS NULL
			AND valid_ingredients.archived_at IS NULL
			AND meal_plans.archived_at IS NULL
			AND meal_plan_grocery_list_items.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
			AND (meal_plan_grocery_list_items.belongs_to_meal_plan_option IS NULL OR NOT EXISTS (SELECT 1 FROM meal_plan_options o JOIN meal_plan_events e ON o.belongs_to_meal_plan_event = e.id WHERE o.id = meal_plan_grocery_list_items.belongs_to_meal_plan_option AND e.archived_at IS NOT NULL))
	) AS total_count
FROM meal_plan_grocery_list_items
	JOIN meal_plans ON meal_plan_grocery_list_items.belongs_to_meal_plan=meal_plans.id
	JOIN valid_ingredients ON meal_plan_grocery_list_items.valid_ingredient=valid_ingredients.id
	JOIN valid_measurement_units ON meal_plan_grocery_list_items.valid_measurement_unit=valid_measurement_units.id
WHERE meal_plan_grocery_list_items.archived_at IS NULL
	AND meal_plan_grocery_list_items.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND meal_plan_grocery_list_items.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		meal_plan_grocery_list_items.last_updated_at IS NULL
		OR meal_plan_grocery_list_items.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		meal_plan_grocery_list_items.last_updated_at IS NULL
		OR meal_plan_grocery_list_items.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR meal_plan_grocery_list_items.archived_at IS NULL)
	AND meal_plan_grocery_list_items.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
	AND valid_measurement_units.archived_at IS NULL
	AND valid_ingredients.archived_at IS NULL
	AND meal_plans.archived_at IS NULL
	AND (meal_plan_grocery_list_items.belongs_to_meal_plan_option IS NULL OR NOT EXISTS (SELECT 1 FROM meal_plan_options o JOIN meal_plan_events e ON o.belongs_to_meal_plan_event = e.id WHERE o.id = meal_plan_grocery_list_items.belongs_to_meal_plan_option AND e.archived_at IS NOT NULL))
	AND meal_plan_grocery_list_items.id > COALESCE(sqlc.narg(cursor), '')
GROUP BY meal_plan_grocery_list_items.id,
	valid_ingredients.id,
	valid_measurement_units.id,
	meal_plans.id
ORDER BY meal_plan_grocery_list_items.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetMealPlanGroceryListItem :one
SELECT
	meal_plan_grocery_list_items.id,
	meal_plan_grocery_list_items.belongs_to_meal_plan,
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
	meal_plan_grocery_list_items.minimum_quantity_needed,
	meal_plan_grocery_list_items.maximum_quantity_needed,
	meal_plan_grocery_list_items.quantity_purchased,
	meal_plan_grocery_list_items.purchased_measurement_unit,
	meal_plan_grocery_list_items.purchased_upc,
	meal_plan_grocery_list_items.purchase_price,
	meal_plan_grocery_list_items.status_explanation,
	meal_plan_grocery_list_items.status,
	meal_plan_grocery_list_items.created_at,
	meal_plan_grocery_list_items.last_updated_at,
	meal_plan_grocery_list_items.archived_at,
	meal_plan_grocery_list_items.belongs_to_meal_plan_option,
	meal_plan_grocery_list_items.recipe_id,
	meal_plan_grocery_list_items.recipe_step_id,
	meal_plan_grocery_list_items.ingredient_index,
	meal_plan_grocery_list_items.option_index
FROM meal_plan_grocery_list_items
	JOIN meal_plans ON meal_plan_grocery_list_items.belongs_to_meal_plan=meal_plans.id
	JOIN valid_ingredients ON meal_plan_grocery_list_items.valid_ingredient=valid_ingredients.id
	JOIN valid_measurement_units ON meal_plan_grocery_list_items.valid_measurement_unit=valid_measurement_units.id
WHERE meal_plan_grocery_list_items.archived_at IS NULL
	AND valid_measurement_units.archived_at IS NULL
	AND valid_ingredients.archived_at IS NULL
	AND meal_plan_grocery_list_items.id = sqlc.arg(meal_plan_grocery_list_item_id)
	AND meal_plan_grocery_list_items.belongs_to_meal_plan = sqlc.arg(meal_plan_id);

-- name: UpdateMealPlanGroceryListItem :execrows
UPDATE meal_plan_grocery_list_items SET
	belongs_to_meal_plan = sqlc.arg(belongs_to_meal_plan),
	valid_ingredient = sqlc.arg(valid_ingredient),
	valid_measurement_unit = sqlc.arg(valid_measurement_unit),
	minimum_quantity_needed = sqlc.arg(minimum_quantity_needed),
	maximum_quantity_needed = sqlc.arg(maximum_quantity_needed),
	quantity_purchased = sqlc.arg(quantity_purchased),
	purchased_measurement_unit = sqlc.arg(purchased_measurement_unit),
	purchased_upc = sqlc.arg(purchased_upc),
	purchase_price = sqlc.arg(purchase_price),
	status_explanation = sqlc.arg(status_explanation),
	status = sqlc.arg(status),
	belongs_to_meal_plan_option = sqlc.arg(belongs_to_meal_plan_option),
	recipe_id = sqlc.arg(recipe_id),
	recipe_step_id = sqlc.arg(recipe_step_id),
	ingredient_index = sqlc.arg(ingredient_index),
	option_index = sqlc.arg(option_index),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);
