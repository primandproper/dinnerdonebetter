-- name: ArchiveRecipe :execrows
UPDATE recipes SET archived_at = NOW() WHERE archived_at IS NULL AND created_by_user = sqlc.arg(created_by_user) AND id = sqlc.arg(id);

-- name: CreateRecipe :exec
INSERT INTO recipes (
	id,
	name,
	slug,
	source,
	source_isbn,
	description,
	status,
	inspired_by_recipe_id,
	min_estimated_portions,
	max_estimated_portions,
	portion_name,
	plural_portion_name,
	eligible_for_meals,
	yields_component_type,
	created_by_user
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(slug),
	sqlc.arg(source),
	sqlc.arg(source_isbn),
	sqlc.arg(description),
	sqlc.arg(status),
	sqlc.arg(inspired_by_recipe_id),
	sqlc.arg(min_estimated_portions),
	sqlc.arg(max_estimated_portions),
	sqlc.arg(portion_name),
	sqlc.arg(plural_portion_name),
	sqlc.arg(eligible_for_meals),
	sqlc.arg(yields_component_type),
	sqlc.arg(created_by_user)
);

-- name: CheckRecipeExistence :one
SELECT EXISTS (
	SELECT recipes.id
	FROM recipes
	WHERE recipes.archived_at IS NULL
		AND recipes.id = sqlc.arg(id)
);

-- name: GetRecipeByID :many
SELECT
	recipes.id,
	recipes.name,
	recipes.slug,
	recipes.source,
	recipes.source_isbn,
	recipes.description,
	recipes.status,
	recipes.inspired_by_recipe_id,
	recipes.min_estimated_portions,
	recipes.max_estimated_portions,
	recipes.portion_name,
	recipes.plural_portion_name,
	recipes.eligible_for_meals,
	recipes.yields_component_type,
	recipes.last_indexed_at,
	recipes.last_validated_at,
	recipes.created_at,
	recipes.last_updated_at,
	recipes.archived_at,
	recipes.created_by_user,
	recipe_steps.id as recipe_step_id,
	recipe_steps.index as recipe_step_index,
	valid_preparations.id as recipe_step_preparation_id,
	valid_preparations.name as recipe_step_preparation_name,
	valid_preparations.description as recipe_step_preparation_description,
	valid_preparations.icon_path as recipe_step_preparation_icon_path,
	valid_preparations.yields_nothing as recipe_step_preparation_yields_nothing,
	valid_preparations.restrict_to_ingredients as recipe_step_preparation_restrict_to_ingredients,
	valid_preparations.past_tense as recipe_step_preparation_past_tense,
	valid_preparations.slug as recipe_step_preparation_slug,
	valid_preparations.minimum_ingredient_count as recipe_step_preparation_minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count as recipe_step_preparation_maximum_ingredient_count,
	valid_preparations.minimum_instrument_count as recipe_step_preparation_minimum_instrument_count,
	valid_preparations.maximum_instrument_count as recipe_step_preparation_maximum_instrument_count,
	valid_preparations.temperature_required as recipe_step_preparation_temperature_required,
	valid_preparations.time_estimate_required as recipe_step_preparation_time_estimate_required,
	valid_preparations.condition_expression_required as recipe_step_preparation_condition_expression_required,
	valid_preparations.consumes_vessel as recipe_step_preparation_consumes_vessel,
	valid_preparations.only_for_vessels as recipe_step_preparation_only_for_vessels,
	valid_preparations.minimum_vessel_count as recipe_step_preparation_minimum_vessel_count,
	valid_preparations.maximum_vessel_count as recipe_step_preparation_maximum_vessel_count,
	valid_preparations.last_indexed_at as recipe_step_preparation_last_indexed_at,
	valid_preparations.created_at as recipe_step_preparation_created_at,
	valid_preparations.last_updated_at as recipe_step_preparation_last_updated_at,
	valid_preparations.archived_at as recipe_step_preparation_archived_at,
	recipe_steps.minimum_estimated_time_in_seconds as recipe_step_minimum_estimated_time_in_seconds,
	recipe_steps.maximum_estimated_time_in_seconds as recipe_step_maximum_estimated_time_in_seconds,
	recipe_steps.minimum_temperature_in_celsius as recipe_step_minimum_temperature_in_celsius,
	recipe_steps.maximum_temperature_in_celsius as recipe_step_maximum_temperature_in_celsius,
	recipe_steps.notes as recipe_step_notes,
	recipe_steps.explicit_instructions as recipe_step_explicit_instructions,
	recipe_steps.condition_expression as recipe_step_condition_expression,
	recipe_steps.optional as recipe_step_optional,
	recipe_steps.start_timer_automatically as recipe_step_start_timer_automatically,
	recipe_steps.created_at as recipe_step_created_at,
	recipe_steps.last_updated_at as recipe_step_last_updated_at,
	recipe_steps.archived_at as recipe_step_archived_at,
	recipe_steps.belongs_to_recipe as recipe_step_belongs_to_recipe
FROM recipes
	LEFT JOIN recipe_steps ON recipes.id=recipe_steps.belongs_to_recipe
	LEFT JOIN valid_preparations ON recipe_steps.preparation_id=valid_preparations.id
WHERE recipes.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id)
ORDER BY recipe_steps.index;

-- name: GetRecipeByIDAndAuthorID :many
SELECT
	recipes.id,
	recipes.name,
	recipes.slug,
	recipes.source,
	recipes.source_isbn,
	recipes.description,
	recipes.status,
	recipes.inspired_by_recipe_id,
	recipes.min_estimated_portions,
	recipes.max_estimated_portions,
	recipes.portion_name,
	recipes.plural_portion_name,
	recipes.eligible_for_meals,
	recipes.yields_component_type,
	recipes.last_indexed_at,
	recipes.last_validated_at,
	recipes.created_at,
	recipes.last_updated_at,
	recipes.archived_at,
	recipes.created_by_user,
	recipe_steps.id as recipe_step_id,
	recipe_steps.index as recipe_step_index,
	valid_preparations.id as recipe_step_preparation_id,
	valid_preparations.name as recipe_step_preparation_name,
	valid_preparations.description as recipe_step_preparation_description,
	valid_preparations.icon_path as recipe_step_preparation_icon_path,
	valid_preparations.yields_nothing as recipe_step_preparation_yields_nothing,
	valid_preparations.restrict_to_ingredients as recipe_step_preparation_restrict_to_ingredients,
	valid_preparations.past_tense as recipe_step_preparation_past_tense,
	valid_preparations.slug as recipe_step_preparation_slug,
	valid_preparations.minimum_ingredient_count as recipe_step_preparation_minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count as recipe_step_preparation_maximum_ingredient_count,
	valid_preparations.minimum_instrument_count as recipe_step_preparation_minimum_instrument_count,
	valid_preparations.maximum_instrument_count as recipe_step_preparation_maximum_instrument_count,
	valid_preparations.temperature_required as recipe_step_preparation_temperature_required,
	valid_preparations.time_estimate_required as recipe_step_preparation_time_estimate_required,
	valid_preparations.condition_expression_required as recipe_step_preparation_condition_expression_required,
	valid_preparations.consumes_vessel as recipe_step_preparation_consumes_vessel,
	valid_preparations.only_for_vessels as recipe_step_preparation_only_for_vessels,
	valid_preparations.minimum_vessel_count as recipe_step_preparation_minimum_vessel_count,
	valid_preparations.maximum_vessel_count as recipe_step_preparation_maximum_vessel_count,
	valid_preparations.last_indexed_at as recipe_step_preparation_last_indexed_at,
	valid_preparations.created_at as recipe_step_preparation_created_at,
	valid_preparations.last_updated_at as recipe_step_preparation_last_updated_at,
	valid_preparations.archived_at as recipe_step_preparation_archived_at,
	recipe_steps.minimum_estimated_time_in_seconds as recipe_step_minimum_estimated_time_in_seconds,
	recipe_steps.maximum_estimated_time_in_seconds as recipe_step_maximum_estimated_time_in_seconds,
	recipe_steps.minimum_temperature_in_celsius as recipe_step_minimum_temperature_in_celsius,
	recipe_steps.maximum_temperature_in_celsius as recipe_step_maximum_temperature_in_celsius,
	recipe_steps.notes as recipe_step_notes,
	recipe_steps.explicit_instructions as recipe_step_explicit_instructions,
	recipe_steps.condition_expression as recipe_step_condition_expression,
	recipe_steps.optional as recipe_step_optional,
	recipe_steps.start_timer_automatically as recipe_step_start_timer_automatically,
	recipe_steps.created_at as recipe_step_created_at,
	recipe_steps.last_updated_at as recipe_step_last_updated_at,
	recipe_steps.archived_at as recipe_step_archived_at,
	recipe_steps.belongs_to_recipe as recipe_step_belongs_to_recipe
FROM recipes
	LEFT JOIN recipe_steps ON recipes.id=recipe_steps.belongs_to_recipe
	LEFT JOIN valid_preparations ON recipe_steps.preparation_id=valid_preparations.id
WHERE recipes.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id)
	AND recipes.created_by_user = sqlc.arg(created_by_user)
ORDER BY recipe_steps.index;

-- name: GetRecipes :many
SELECT
	recipes.id,
	recipes.name,
	recipes.slug,
	recipes.source,
	recipes.source_isbn,
	recipes.description,
	recipes.status,
	recipes.inspired_by_recipe_id,
	recipes.min_estimated_portions,
	recipes.max_estimated_portions,
	recipes.portion_name,
	recipes.plural_portion_name,
	recipes.eligible_for_meals,
	recipes.yields_component_type,
	recipes.last_indexed_at,
	recipes.last_validated_at,
	recipes.created_at,
	recipes.last_updated_at,
	recipes.archived_at,
	recipes.created_by_user,
	(
		SELECT COUNT(recipes.id)
		FROM recipes
		WHERE recipes.archived_at IS NULL
			AND
			recipes.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND recipes.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				recipes.last_updated_at IS NULL
				OR recipes.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				recipes.last_updated_at IS NULL
				OR recipes.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR recipes.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(recipes.id)
		FROM recipes
		WHERE recipes.archived_at IS NULL
	) AS total_count
FROM recipes
	WHERE recipes.archived_at IS NULL
	AND recipes.status = COALESCE(sqlc.narg(status), 'approved')::recipe_status
	AND recipes.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND recipes.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		recipes.last_updated_at IS NULL
		OR recipes.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		recipes.last_updated_at IS NULL
		OR recipes.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND recipes.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY recipes.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetRecipesCreatedByUser :many
SELECT
	recipes.id,
	recipes.name,
	recipes.slug,
	recipes.source,
	recipes.source_isbn,
	recipes.description,
	recipes.status,
	recipes.inspired_by_recipe_id,
	recipes.min_estimated_portions,
	recipes.max_estimated_portions,
	recipes.portion_name,
	recipes.plural_portion_name,
	recipes.eligible_for_meals,
	recipes.yields_component_type,
	recipes.last_indexed_at,
	recipes.last_validated_at,
	recipes.created_at,
	recipes.last_updated_at,
	recipes.archived_at,
	recipes.created_by_user,
	(
		SELECT COUNT(recipes.id)
		FROM recipes
		WHERE recipes.archived_at IS NULL
			AND
			recipes.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND recipes.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				recipes.last_updated_at IS NULL
				OR recipes.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				recipes.last_updated_at IS NULL
				OR recipes.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR recipes.archived_at IS NULL)
			AND recipes.created_by_user = sqlc.arg(created_by_user)
	) AS filtered_count,
	(
		SELECT COUNT(recipes.id)
		FROM recipes
		WHERE recipes.archived_at IS NULL
			AND recipes.created_by_user = sqlc.arg(created_by_user)
	) AS total_count
FROM recipes
	WHERE recipes.archived_at IS NULL AND
	recipes.created_by_user = sqlc.arg(created_by_user)
	AND recipes.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND recipes.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		recipes.last_updated_at IS NULL
		OR recipes.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		recipes.last_updated_at IS NULL
		OR recipes.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND recipes.created_by_user = sqlc.arg(created_by_user)
	AND recipes.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY recipes.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetRecipesWithIDs :many
SELECT
	recipes.id,
	recipes.name,
	recipes.slug,
	recipes.source,
	recipes.source_isbn,
	recipes.description,
	recipes.status,
	recipes.inspired_by_recipe_id,
	recipes.min_estimated_portions,
	recipes.max_estimated_portions,
	recipes.portion_name,
	recipes.plural_portion_name,
	recipes.eligible_for_meals,
	recipes.yields_component_type,
	recipes.last_indexed_at,
	recipes.last_validated_at,
	recipes.created_at,
	recipes.last_updated_at,
	recipes.archived_at,
	recipes.created_by_user,
	recipe_steps.id as recipe_step_id,
	recipe_steps.index as recipe_step_index,
	valid_preparations.id as recipe_step_preparation_id,
	valid_preparations.name as recipe_step_preparation_name,
	valid_preparations.description as recipe_step_preparation_description,
	valid_preparations.icon_path as recipe_step_preparation_icon_path,
	valid_preparations.yields_nothing as recipe_step_preparation_yields_nothing,
	valid_preparations.restrict_to_ingredients as recipe_step_preparation_restrict_to_ingredients,
	valid_preparations.past_tense as recipe_step_preparation_past_tense,
	valid_preparations.slug as recipe_step_preparation_slug,
	valid_preparations.minimum_ingredient_count as recipe_step_preparation_minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count as recipe_step_preparation_maximum_ingredient_count,
	valid_preparations.minimum_instrument_count as recipe_step_preparation_minimum_instrument_count,
	valid_preparations.maximum_instrument_count as recipe_step_preparation_maximum_instrument_count,
	valid_preparations.temperature_required as recipe_step_preparation_temperature_required,
	valid_preparations.time_estimate_required as recipe_step_preparation_time_estimate_required,
	valid_preparations.condition_expression_required as recipe_step_preparation_condition_expression_required,
	valid_preparations.consumes_vessel as recipe_step_preparation_consumes_vessel,
	valid_preparations.only_for_vessels as recipe_step_preparation_only_for_vessels,
	valid_preparations.minimum_vessel_count as recipe_step_preparation_minimum_vessel_count,
	valid_preparations.maximum_vessel_count as recipe_step_preparation_maximum_vessel_count,
	valid_preparations.last_indexed_at as recipe_step_preparation_last_indexed_at,
	valid_preparations.created_at as recipe_step_preparation_created_at,
	valid_preparations.last_updated_at as recipe_step_preparation_last_updated_at,
	valid_preparations.archived_at as recipe_step_preparation_archived_at,
	recipe_steps.minimum_estimated_time_in_seconds as recipe_step_minimum_estimated_time_in_seconds,
	recipe_steps.maximum_estimated_time_in_seconds as recipe_step_maximum_estimated_time_in_seconds,
	recipe_steps.minimum_temperature_in_celsius as recipe_step_minimum_temperature_in_celsius,
	recipe_steps.maximum_temperature_in_celsius as recipe_step_maximum_temperature_in_celsius,
	recipe_steps.notes as recipe_step_notes,
	recipe_steps.explicit_instructions as recipe_step_explicit_instructions,
	recipe_steps.condition_expression as recipe_step_condition_expression,
	recipe_steps.optional as recipe_step_optional,
	recipe_steps.start_timer_automatically as recipe_step_start_timer_automatically,
	recipe_steps.created_at as recipe_step_created_at,
	recipe_steps.last_updated_at as recipe_step_last_updated_at,
	recipe_steps.archived_at as recipe_step_archived_at,
	recipe_steps.belongs_to_recipe as recipe_step_belongs_to_recipe
FROM recipes
	LEFT JOIN recipe_steps ON recipes.id=recipe_steps.belongs_to_recipe
	LEFT JOIN valid_preparations ON recipe_steps.preparation_id=valid_preparations.id
WHERE recipes.archived_at IS NULL
	AND recipes.id = ANY(sqlc.arg(ids)::text[])
ORDER BY recipes.id ASC;

-- name: RecipeSearch :many
SELECT
	recipes.id,
	recipes.name,
	recipes.slug,
	recipes.source,
	recipes.source_isbn,
	recipes.description,
	recipes.status,
	recipes.inspired_by_recipe_id,
	recipes.min_estimated_portions,
	recipes.max_estimated_portions,
	recipes.portion_name,
	recipes.plural_portion_name,
	recipes.eligible_for_meals,
	recipes.yields_component_type,
	recipes.last_indexed_at,
	recipes.last_validated_at,
	recipes.created_at,
	recipes.last_updated_at,
	recipes.archived_at,
	recipes.created_by_user,
	(
		SELECT COUNT(recipes.id)
		FROM recipes
		WHERE recipes.archived_at IS NULL
			AND
			recipes.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND recipes.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				recipes.last_updated_at IS NULL
				OR recipes.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				recipes.last_updated_at IS NULL
				OR recipes.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR recipes.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(recipes.id)
		FROM recipes
		WHERE recipes.archived_at IS NULL
	) AS total_count
FROM recipes
WHERE recipes.archived_at IS NULL
	AND recipes.name ILIKE '%' || sqlc.arg(query)::text || '%'
	AND recipes.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND recipes.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		recipes.last_updated_at IS NULL
		OR recipes.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		recipes.last_updated_at IS NULL
		OR recipes.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND recipes.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY recipes.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: SearchForMealEligibleRecipes :many
SELECT
	recipes.id,
	recipes.name,
	recipes.slug,
	recipes.source,
	recipes.source_isbn,
	recipes.description,
	recipes.status,
	recipes.inspired_by_recipe_id,
	recipes.min_estimated_portions,
	recipes.max_estimated_portions,
	recipes.portion_name,
	recipes.plural_portion_name,
	recipes.eligible_for_meals,
	recipes.yields_component_type,
	recipes.last_indexed_at,
	recipes.last_validated_at,
	recipes.created_at,
	recipes.last_updated_at,
	recipes.archived_at,
	recipes.created_by_user,
	(
		SELECT COUNT(recipes.id)
		FROM recipes
		WHERE recipes.archived_at IS NULL
			AND
			recipes.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND recipes.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				recipes.last_updated_at IS NULL
				OR recipes.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				recipes.last_updated_at IS NULL
				OR recipes.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR recipes.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(recipes.id)
		FROM recipes
		WHERE recipes.archived_at IS NULL
	) AS total_count
FROM recipes
WHERE recipes.archived_at IS NULL
	AND recipes.eligible_for_meals = true
	AND recipes.status = 'approved'
	AND recipes.name ILIKE '%' || sqlc.arg(query)::text || '%'
	AND recipes.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND recipes.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		recipes.last_updated_at IS NULL
		OR recipes.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		recipes.last_updated_at IS NULL
		OR recipes.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND recipes.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY recipes.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: SearchForRecipesWithInstrumentOwnership :many
SELECT
	recipes.id,
	recipes.name,
	recipes.slug,
	recipes.source,
	recipes.source_isbn,
	recipes.description,
	recipes.status,
	recipes.inspired_by_recipe_id,
	recipes.min_estimated_portions,
	recipes.max_estimated_portions,
	recipes.portion_name,
	recipes.plural_portion_name,
	recipes.eligible_for_meals,
	recipes.yields_component_type,
	recipes.last_indexed_at,
	recipes.last_validated_at,
	recipes.created_at,
	recipes.last_updated_at,
	recipes.archived_at,
	recipes.created_by_user,
	(
		SELECT COUNT(recipes.id)
		FROM recipes
		WHERE recipes.archived_at IS NULL
			AND
			recipes.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND recipes.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				recipes.last_updated_at IS NULL
				OR recipes.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				recipes.last_updated_at IS NULL
				OR recipes.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR recipes.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(recipes.id)
		FROM recipes
		WHERE recipes.archived_at IS NULL
	) AS total_count
FROM recipes
WHERE recipes.archived_at IS NULL
	AND recipes.name ILIKE '%' || sqlc.arg(query)::text || '%'
	AND recipes.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND recipes.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		recipes.last_updated_at IS NULL
		OR recipes.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		recipes.last_updated_at IS NULL
		OR recipes.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND recipes.id > COALESCE(sqlc.narg(cursor), '')
	AND NOT EXISTS (
		SELECT 1 FROM recipe_step_instruments rsi
		JOIN recipe_steps rs ON rsi.belongs_to_recipe_step = rs.id
		WHERE rs.belongs_to_recipe = recipes.id
			AND rsi.archived_at IS NULL
			AND rs.archived_at IS NULL
			AND rsi.optional = false
			AND rsi.instrument_id IS NOT NULL
			AND rsi.instrument_id NOT IN (
				SELECT valid_instrument_id FROM account_instrument_ownerships
				WHERE belongs_to_account = sqlc.arg(account_id) AND archived_at IS NULL
			)
	)
ORDER BY recipes.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetRecipesNeedingIndexing :many
SELECT recipes.id
FROM recipes
WHERE recipes.archived_at IS NULL
	AND (
		recipes.last_indexed_at IS NULL
		OR recipes.last_indexed_at < NOW() - '24 hours'::INTERVAL
	);

-- name: GetRecipeIDsForMeal :many
SELECT recipes.id
FROM recipes
	JOIN meal_components ON meal_components.recipe_id = recipes.id
	JOIN meals ON meal_components.belongs_to_meal = meals.id
WHERE
	recipes.archived_at IS NULL
	AND meals.id = sqlc.arg(meal_id)
GROUP BY recipes.id
ORDER BY recipes.id;

-- name: UpdateRecipe :execrows
UPDATE recipes SET
	name = sqlc.arg(name),
	slug = sqlc.arg(slug),
	source = sqlc.arg(source),
	source_isbn = sqlc.arg(source_isbn),
	description = sqlc.arg(description),
	inspired_by_recipe_id = sqlc.arg(inspired_by_recipe_id),
	min_estimated_portions = sqlc.arg(min_estimated_portions),
	max_estimated_portions = sqlc.arg(max_estimated_portions),
	portion_name = sqlc.arg(portion_name),
	plural_portion_name = sqlc.arg(plural_portion_name),
	eligible_for_meals = sqlc.arg(eligible_for_meals),
	yields_component_type = sqlc.arg(yields_component_type),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND created_by_user = sqlc.arg(created_by_user)
	AND id = sqlc.arg(id);

-- name: UpdateRecipeStatus :execrows
UPDATE recipes SET
	status = sqlc.arg(status),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: UpdateRecipeLastIndexedAt :execrows
UPDATE recipes SET last_indexed_at = NOW() WHERE id = sqlc.arg(id) AND archived_at IS NULL;
