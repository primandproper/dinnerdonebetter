-- name: DeleteExpiredOAuth2ClientTokens :execrows
DELETE FROM oauth2_client_tokens WHERE code_expires_at < (NOW() - interval '1 day') AND access_expires_at < (NOW() - interval '1 day') AND refresh_expires_at < (NOW() - interval '1 day');

-- name: DestroyAllData :exec
TRUNCATE account_instrument_ownerships, account_invitations, account_user_memberships, accounts, audit_log_entries, comments, issue_reports, meal_components, meal_list_items, meal_lists, meal_plan_events, meal_plan_grocery_list_items, meal_plan_option_votes, meal_plan_options, meal_plan_recipe_option_selections, meal_plan_tasks, meal_plans, meals, oauth2_client_tokens, oauth2_clients, password_reset_tokens, payment_transactions, permissions, products, purchases, queue_test_messages, recipe_list_items, recipe_lists, recipe_media, recipe_prep_task_steps, recipe_prep_tasks, recipe_ratings, recipe_step_completion_condition_ingredients, recipe_step_completion_conditions, recipe_step_ingredients, recipe_step_instruments, recipe_step_products, recipe_step_vessels, recipe_steps, recipes, service_setting_configurations, service_settings, subscriptions, uploaded_media, user_avatars, user_data_disclosures, user_ingredient_preferences, user_notifications, user_role_assignments, user_role_hierarchy, user_role_permissions, user_roles, user_sessions, users, valid_ingredient_group_members, valid_ingredient_groups, valid_ingredient_measurement_units, valid_ingredient_preparations, valid_ingredient_state_ingredients, valid_ingredient_states, valid_ingredients, valid_instruments, valid_measurement_unit_conversions, valid_measurement_units, valid_prep_task_configs, valid_preparation_instruments, valid_preparation_vessels, valid_preparations, valid_vessels, waitlist_signups, waitlists, webhook_trigger_configs, webhooks CASCADE;

-- name: CreateQueueTestMessage :exec
INSERT INTO queue_test_messages (id, queue_name) VALUES (sqlc.arg(id), sqlc.arg(queue_name));

-- name: AcknowledgeQueueTestMessage :exec
UPDATE queue_test_messages SET acknowledged_at = NOW() WHERE id = sqlc.arg(id);

-- name: GetQueueTestMessage :one
SELECT id, queue_name, created_at, acknowledged_at FROM queue_test_messages WHERE id = sqlc.arg(id);

-- name: PruneQueueTestMessages :exec
DELETE FROM queue_test_messages AS qtm
WHERE qtm.queue_name = sqlc.arg(queue_name)
  AND qtm.id NOT IN (
      SELECT keep.id FROM queue_test_messages AS keep
      WHERE keep.queue_name = sqlc.arg(queue_name)
      ORDER BY keep.created_at DESC
      LIMIT 100
  );
