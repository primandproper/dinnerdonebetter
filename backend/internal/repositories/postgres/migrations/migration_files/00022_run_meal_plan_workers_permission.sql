-- =============================================================================
-- Seed the run.meal_plan_workers permission and grant it to service_admin.
--
-- The RunMealPlanWorkersPermission ("run.meal_plan_workers") was introduced in code as a dedicated,
-- service-admin-only scope for triggering the global meal-plan background workers
-- (RunFinalizeMealPlanWorker, RunMealPlanGroceryListInitializerWorker, RunMealPlanTaskCreatorWorker),
-- replacing the previous, over-broad UpdateMealPlansPermission requirement. Session permissions are
-- loaded from the database (GetServicePermissionsForUser), so the permission must be seeded here for
-- the service_admin role, otherwise even service admins get PermissionDenied on those endpoints.
-- =============================================================================

INSERT INTO permissions (id, name, description) VALUES
    ('d9run0mealplanwrkr00', 'run.meal_plan_workers', 'Trigger global meal-plan background workers')
ON CONFLICT DO NOTHING;

INSERT INTO user_role_permissions (id, role_id, permission_id) VALUES
    ('d9run0mealplanwrkr01', 'role_service_admin', 'd9run0mealplanwrkr00')
ON CONFLICT DO NOTHING;
