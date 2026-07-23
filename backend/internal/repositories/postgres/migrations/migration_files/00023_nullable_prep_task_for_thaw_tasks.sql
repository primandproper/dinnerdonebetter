-- Ad-hoc thaw tasks (generated for frozen ingredients) have no backing recipe prep task,
-- so the reference must be optional. NULL means "standalone task not derived from a prep task".
ALTER TABLE meal_plan_tasks ALTER COLUMN belongs_to_recipe_prep_task DROP NOT NULL;
