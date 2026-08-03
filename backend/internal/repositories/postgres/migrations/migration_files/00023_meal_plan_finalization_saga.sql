-- Meal plan finalization saga
--
-- Finalization used to be three independent interval-polled jobs coordinated by two boolean
-- columns, each rediscovering its own work with a "finalized but not yet X" query. It is now
-- one saga (see internal/services/mealplanning/workers/meal_plan_finalization), and this
-- column is what makes starting one idempotent: the job that starts sagas selects plans with
-- no saga attached, and the attach happens in the same transaction as the instance row.
--
-- grocery_list_initialized and tasks_created stay. They stopped being the coordinator and
-- became per-step idempotency guards: each is written in the same transaction as the work it
-- describes, which is a stronger guarantee than the saga's own idempotency keys can offer for
-- a step that writes to this database. They are also still on the API's MealPlan message, so
-- dropping them is a separate change to that contract.

ALTER TABLE meal_plans ADD COLUMN IF NOT EXISTS finalization_saga_id TEXT;

-- One saga per plan, enforced rather than assumed. The attach is a conditional UPDATE that
-- rolls the instance row back when it matches nothing, so this index is the backstop for the
-- case that predicate cannot see: two transactions attaching different sagas at once.
CREATE UNIQUE INDEX IF NOT EXISTS idx_meal_plans_finalization_saga_id
    ON meal_plans (finalization_saga_id)
    WHERE finalization_saga_id IS NOT NULL;

-- Serves the starter's predicate. Partial on "no saga yet", which is the whole of the working
-- set: a plan gets a saga once and keeps it, so the index tracks the plans still waiting
-- rather than every plan the system has ever finalized.
CREATE INDEX IF NOT EXISTS idx_meal_plans_awaiting_finalization_saga
    ON meal_plans (status, voting_deadline)
    WHERE archived_at IS NULL AND finalization_saga_id IS NULL;
