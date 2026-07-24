-- name: CheckMealInMealPlanEvent :one
SELECT EXISTS (
	SELECT meal_plan_options.id
	FROM meal_plan_options
	WHERE meal_plan_options.archived_at IS NULL
		AND meal_plan_options.belongs_to_meal_plan_event = sqlc.arg(belongs_to_meal_plan_event)
		AND meal_plan_options.meal_id = sqlc.arg(meal_id)
);

-- name: ArchiveMealPlanOption :execrows
UPDATE meal_plan_options SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND belongs_to_meal_plan_event = sqlc.arg(belongs_to_meal_plan_event)
	AND id = sqlc.arg(id);

-- name: CreateMealPlanOption :exec
INSERT INTO meal_plan_options (
	id,
	assigned_cook,
	assigned_dishwasher,
	chosen,
	meal_scale,
	meal_id,
	notes,
	belongs_to_meal_plan_event
) VALUES (
	sqlc.arg(id),
	sqlc.arg(assigned_cook),
	sqlc.arg(assigned_dishwasher),
	sqlc.arg(chosen),
	sqlc.arg(meal_scale),
	sqlc.arg(meal_id),
	sqlc.arg(notes),
	sqlc.arg(belongs_to_meal_plan_event)
);

-- name: CheckMealPlanOptionExistence :one
SELECT EXISTS (
	SELECT meal_plan_options.id
	FROM meal_plan_options
		JOIN meal_plan_events ON meal_plan_options.belongs_to_meal_plan_event = meal_plan_events.id
		JOIN meal_plans ON meal_plan_events.belongs_to_meal_plan = meal_plans.id
	WHERE meal_plan_options.archived_at IS NULL
		AND meal_plan_options.belongs_to_meal_plan_event = sqlc.arg(meal_plan_event_id)
		AND meal_plan_options.id = sqlc.arg(meal_plan_option_id)
		AND meal_plan_events.archived_at IS NULL
		AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
		AND meal_plan_events.id = sqlc.arg(meal_plan_event_id)
		AND meal_plans.archived_at IS NULL
		AND meal_plans.id = sqlc.arg(meal_plan_id)
);

-- name: CheckMealPlanOptionBelongsToAccount :one
SELECT EXISTS (
	SELECT meal_plan_options.id
	FROM meal_plan_options
		JOIN meal_plan_events ON meal_plan_options.belongs_to_meal_plan_event = meal_plan_events.id
		JOIN meal_plans ON meal_plan_events.belongs_to_meal_plan = meal_plans.id
	WHERE meal_plan_options.archived_at IS NULL
		AND meal_plan_options.id = sqlc.arg(meal_plan_option_id)
		AND meal_plans.archived_at IS NULL
		AND meal_plans.belongs_to_account = sqlc.arg(belongs_to_account)
);

-- name: FinalizeMealPlanOption :exec
UPDATE meal_plan_options SET
	chosen = (belongs_to_meal_plan_event = sqlc.arg(meal_plan_event_id) AND id = sqlc.arg(id)),
	tiebroken = sqlc.arg(tiebroken)
WHERE archived_at IS NULL
	AND belongs_to_meal_plan_event = sqlc.arg(meal_plan_event_id)
	AND id = sqlc.arg(id);

-- name: GetAllMealPlanOptionsForMealPlanEvent :many
SELECT
	meal_plan_options.id,
	meal_plan_options.assigned_cook,
	meal_plan_options.assigned_dishwasher,
	meal_plan_options.chosen,
	meal_plan_options.tiebroken,
	meal_plan_options.meal_scale,
	meal_plan_options.meal_id,
	meal_plan_options.notes,
	meal_plan_options.created_at,
	meal_plan_options.last_updated_at,
	meal_plan_options.archived_at,
	meal_plan_options.belongs_to_meal_plan_event,
	meals.id as meal_id,
	meals.name as meal_name,
	meals.description as meal_description,
	meals.min_estimated_portions as meal_min_estimated_portions,
	meals.max_estimated_portions as meal_max_estimated_portions,
	meals.eligible_for_meal_plans as meal_eligible_for_meal_plans,
	meals.last_indexed_at as meal_last_indexed_at,
	meals.created_at as meal_created_at,
	meals.last_updated_at as meal_last_updated_at,
	meals.archived_at as meal_archived_at,
	meals.created_by_user as meal_created_by_user
FROM meal_plan_options
	JOIN meal_plan_events ON meal_plan_options.belongs_to_meal_plan_event = meal_plan_events.id
	JOIN meal_plans ON meal_plan_events.belongs_to_meal_plan = meal_plans.id
	JOIN meals ON meal_plan_options.meal_id = meals.id
WHERE
	meal_plan_options.archived_at IS NULL
	AND meal_plan_options.belongs_to_meal_plan_event = sqlc.arg(meal_plan_event_id)
	AND meal_plan_events.id = sqlc.arg(meal_plan_event_id)
	AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
	AND meal_plans.archived_at IS NULL
	AND meal_plans.id = sqlc.arg(meal_plan_id);

-- name: GetMealPlanOptions :many
SELECT
	meal_plan_options.id,
	meal_plan_options.assigned_cook,
	meal_plan_options.assigned_dishwasher,
	meal_plan_options.chosen,
	meal_plan_options.tiebroken,
	meal_plan_options.meal_scale,
	meal_plan_options.meal_id,
	meal_plan_options.notes,
	meal_plan_options.created_at,
	meal_plan_options.last_updated_at,
	meal_plan_options.archived_at,
	meal_plan_options.belongs_to_meal_plan_event,
	meals.id as meal_id,
	meals.name as meal_name,
	meals.description as meal_description,
	meals.min_estimated_portions as meal_min_estimated_portions,
	meals.max_estimated_portions as meal_max_estimated_portions,
	meals.eligible_for_meal_plans as meal_eligible_for_meal_plans,
	meals.last_indexed_at as meal_last_indexed_at,
	meals.created_at as meal_created_at,
	meals.last_updated_at as meal_last_updated_at,
	meals.archived_at as meal_archived_at,
	meals.created_by_user as meal_created_by_user,
	(
		SELECT COUNT(meal_plan_options.id)
		FROM meal_plan_options
		WHERE meal_plan_options.archived_at IS NULL
			AND
			meal_plan_options.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND meal_plan_options.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				meal_plan_options.last_updated_at IS NULL
				OR meal_plan_options.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				meal_plan_options.last_updated_at IS NULL
				OR meal_plan_options.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR meal_plan_options.archived_at IS NULL)
			AND meal_plan_options.belongs_to_meal_plan_event = sqlc.arg(meal_plan_event_id)
	) AS filtered_count,
	(
		SELECT COUNT(meal_plan_options.id)
		FROM meal_plan_options
		WHERE meal_plan_options.archived_at IS NULL
	) AS total_count
FROM meal_plan_options
	JOIN meal_plan_events ON meal_plan_options.belongs_to_meal_plan_event = meal_plan_events.id
	JOIN meal_plans ON meal_plan_events.belongs_to_meal_plan = meal_plans.id
	JOIN meals ON meal_plan_options.meal_id = meals.id
WHERE
	meal_plan_options.archived_at IS NULL
	AND meal_plan_options.belongs_to_meal_plan_event = sqlc.arg(meal_plan_event_id)
	AND meal_plan_events.id = sqlc.arg(meal_plan_event_id)
	AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
	AND meal_plans.archived_at IS NULL
	AND meal_plans.id = sqlc.arg(meal_plan_id)
	AND meal_plan_options.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND meal_plan_options.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		meal_plan_options.last_updated_at IS NULL
		OR meal_plan_options.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		meal_plan_options.last_updated_at IS NULL
		OR meal_plan_options.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND meal_plan_options.belongs_to_meal_plan_event = sqlc.arg(meal_plan_event_id)
	AND meal_plan_options.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY meal_plan_options.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetMealPlanOption :one
SELECT
	meal_plan_options.id,
	meal_plan_options.assigned_cook,
	meal_plan_options.assigned_dishwasher,
	meal_plan_options.chosen,
	meal_plan_options.tiebroken,
	meal_plan_options.meal_scale,
	meal_plan_options.meal_id,
	meal_plan_options.notes,
	meal_plan_options.created_at,
	meal_plan_options.last_updated_at,
	meal_plan_options.archived_at,
	meal_plan_options.belongs_to_meal_plan_event,
	meals.id as meal_id,
	meals.name as meal_name,
	meals.description as meal_description,
	meals.min_estimated_portions as meal_min_estimated_portions,
	meals.max_estimated_portions as meal_max_estimated_portions,
	meals.eligible_for_meal_plans as meal_eligible_for_meal_plans,
	meals.last_indexed_at as meal_last_indexed_at,
	meals.created_at as meal_created_at,
	meals.last_updated_at as meal_last_updated_at,
	meals.archived_at as meal_archived_at,
	meals.created_by_user as meal_created_by_user
FROM meal_plan_options
	JOIN meal_plan_events ON meal_plan_options.belongs_to_meal_plan_event = meal_plan_events.id
	JOIN meal_plans ON meal_plan_events.belongs_to_meal_plan = meal_plans.id
	JOIN meals ON meal_plan_options.meal_id = meals.id
WHERE meal_plan_options.archived_at IS NULL
	AND meal_plan_options.belongs_to_meal_plan_event = sqlc.arg(meal_plan_event_id)
	AND meal_plan_options.id = sqlc.arg(meal_plan_option_id)
	AND meal_plan_events.id = sqlc.arg(meal_plan_event_id)
	AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
	AND meal_plans.archived_at IS NULL
	AND meal_plans.id = sqlc.arg(meal_plan_id);

-- name: GetMealPlanOptionByID :one
SELECT
	meal_plan_options.id,
	meal_plan_options.assigned_cook,
	meal_plan_options.assigned_dishwasher,
	meal_plan_options.chosen,
	meal_plan_options.tiebroken,
	meal_plan_options.meal_scale,
	meal_plan_options.meal_id,
	meal_plan_options.notes,
	meal_plan_options.created_at,
	meal_plan_options.last_updated_at,
	meal_plan_options.archived_at,
	meal_plan_options.belongs_to_meal_plan_event,
	meals.id as meal_id,
	meals.name as meal_name,
	meals.description as meal_description,
	meals.min_estimated_portions as meal_min_estimated_portions,
	meals.max_estimated_portions as meal_max_estimated_portions,
	meals.eligible_for_meal_plans as meal_eligible_for_meal_plans,
	meals.last_indexed_at as meal_last_indexed_at,
	meals.created_at as meal_created_at,
	meals.last_updated_at as meal_last_updated_at,
	meals.archived_at as meal_archived_at,
	meals.created_by_user as meal_created_by_user
FROM meal_plan_options
	JOIN meal_plan_events ON meal_plan_options.belongs_to_meal_plan_event = meal_plan_events.id
	JOIN meal_plans ON meal_plan_events.belongs_to_meal_plan = meal_plans.id
	JOIN meals ON meal_plan_options.meal_id = meals.id
WHERE meal_plan_options.archived_at IS NULL
	AND meal_plan_options.id = sqlc.arg(meal_plan_option_id);

-- name: UpdateMealPlanOption :execrows
UPDATE meal_plan_options SET
	assigned_cook = sqlc.arg(assigned_cook),
	assigned_dishwasher = sqlc.arg(assigned_dishwasher),
	meal_scale = sqlc.arg(meal_scale),
	meal_id = sqlc.arg(meal_id),
	notes = sqlc.arg(notes),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND belongs_to_meal_plan_event = sqlc.arg(meal_plan_event_id)
	AND id = sqlc.arg(meal_plan_option_id);
