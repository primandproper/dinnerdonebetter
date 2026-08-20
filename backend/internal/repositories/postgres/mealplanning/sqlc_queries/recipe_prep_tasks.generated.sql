-- name: CreateRecipePrepTask :exec
INSERT INTO recipe_prep_tasks (
	id,
	name,
	description,
	notes,
	optional,
	explicit_storage_instructions,
	minimum_time_buffer_before_recipe_in_seconds,
	maximum_time_buffer_before_recipe_in_seconds,
	storage_type,
	minimum_storage_temperature_in_celsius,
	maximum_storage_temperature_in_celsius,
	belongs_to_recipe
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(description),
	sqlc.arg(notes),
	sqlc.arg(optional),
	sqlc.arg(explicit_storage_instructions),
	sqlc.arg(minimum_time_buffer_before_recipe_in_seconds),
	sqlc.arg(maximum_time_buffer_before_recipe_in_seconds),
	sqlc.arg(storage_type),
	sqlc.arg(minimum_storage_temperature_in_celsius),
	sqlc.arg(maximum_storage_temperature_in_celsius),
	sqlc.arg(belongs_to_recipe)
);

-- name: UpdateRecipePrepTask :execrows
UPDATE recipe_prep_tasks SET
	name = sqlc.arg(name),
	description = sqlc.arg(description),
	notes = sqlc.arg(notes),
	optional = sqlc.arg(optional),
	explicit_storage_instructions = sqlc.arg(explicit_storage_instructions),
	minimum_time_buffer_before_recipe_in_seconds = sqlc.arg(minimum_time_buffer_before_recipe_in_seconds),
	maximum_time_buffer_before_recipe_in_seconds = sqlc.arg(maximum_time_buffer_before_recipe_in_seconds),
	storage_type = sqlc.arg(storage_type),
	minimum_storage_temperature_in_celsius = sqlc.arg(minimum_storage_temperature_in_celsius),
	maximum_storage_temperature_in_celsius = sqlc.arg(maximum_storage_temperature_in_celsius),
	belongs_to_recipe = sqlc.arg(belongs_to_recipe),
	last_updated_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveRecipePrepTask :execrows
UPDATE recipe_prep_tasks SET
	archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: CheckRecipePrepTaskExistence :one
SELECT EXISTS (
	SELECT recipe_prep_tasks.id
	FROM recipe_prep_tasks
		JOIN recipes ON recipe_prep_tasks.belongs_to_recipe=recipes.id
	WHERE recipe_prep_tasks.archived_at IS NULL
		AND recipe_prep_tasks.belongs_to_recipe = sqlc.arg(recipe_id)
		AND recipe_prep_tasks.id = sqlc.arg(recipe_prep_task_id)
		AND recipes.archived_at IS NULL
		AND recipes.id = sqlc.arg(recipe_id)
);

-- name: GetRecipePrepTask :many
SELECT
	recipe_prep_tasks.id,
	recipe_prep_tasks.name,
	recipe_prep_tasks.description,
	recipe_prep_tasks.notes,
	recipe_prep_tasks.optional,
	recipe_prep_tasks.explicit_storage_instructions,
	recipe_prep_tasks.minimum_time_buffer_before_recipe_in_seconds,
	recipe_prep_tasks.maximum_time_buffer_before_recipe_in_seconds,
	recipe_prep_tasks.storage_type,
	recipe_prep_tasks.minimum_storage_temperature_in_celsius,
	recipe_prep_tasks.maximum_storage_temperature_in_celsius,
	recipe_prep_tasks.created_at,
	recipe_prep_tasks.last_updated_at,
	recipe_prep_tasks.archived_at,
	recipe_prep_tasks.belongs_to_recipe,
	recipe_prep_task_steps.id as task_step_id,
	recipe_prep_task_steps.belongs_to_recipe_step as task_step_belongs_to_recipe_step,
	recipe_prep_task_steps.belongs_to_recipe_prep_task as task_step_belongs_to_recipe_prep_task,
	recipe_prep_task_steps.satisfies_recipe_step as task_step_satisfies_recipe_step
FROM recipe_prep_tasks
	JOIN recipe_prep_task_steps ON recipe_prep_tasks.id=recipe_prep_task_steps.belongs_to_recipe_prep_task
WHERE recipe_prep_tasks.archived_at IS NULL
	AND recipe_prep_tasks.id = sqlc.arg(recipe_prep_task_id);

-- name: ListAllRecipePrepTasksByRecipe :many
SELECT
	recipe_prep_tasks.id,
	recipe_prep_tasks.name,
	recipe_prep_tasks.description,
	recipe_prep_tasks.notes,
	recipe_prep_tasks.optional,
	recipe_prep_tasks.explicit_storage_instructions,
	recipe_prep_tasks.minimum_time_buffer_before_recipe_in_seconds,
	recipe_prep_tasks.maximum_time_buffer_before_recipe_in_seconds,
	recipe_prep_tasks.storage_type,
	recipe_prep_tasks.minimum_storage_temperature_in_celsius,
	recipe_prep_tasks.maximum_storage_temperature_in_celsius,
	recipe_prep_tasks.created_at,
	recipe_prep_tasks.last_updated_at,
	recipe_prep_tasks.archived_at,
	recipe_prep_tasks.belongs_to_recipe,
	recipe_prep_task_steps.id as task_step_id,
	recipe_prep_task_steps.belongs_to_recipe_step as task_step_belongs_to_recipe_step,
	recipe_prep_task_steps.belongs_to_recipe_prep_task as task_step_belongs_to_recipe_prep_task,
	recipe_prep_task_steps.satisfies_recipe_step as task_step_satisfies_recipe_step
FROM recipe_prep_tasks
	JOIN recipe_prep_task_steps ON recipe_prep_task_steps.belongs_to_recipe_prep_task=recipe_prep_tasks.id
	JOIN recipe_steps ON recipe_prep_task_steps.belongs_to_recipe_step=recipe_steps.id
	JOIN recipes ON recipe_prep_tasks.belongs_to_recipe=recipes.id
WHERE recipe_prep_tasks.archived_at IS NULL
	AND recipe_steps.archived_at IS NULL
	AND recipes.archived_at IS NULL
	AND recipes.id = sqlc.arg(recipe_id);
