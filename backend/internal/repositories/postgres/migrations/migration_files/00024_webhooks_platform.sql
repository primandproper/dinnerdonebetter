-- Webhooks: adopt the platform delivery tables.
--
-- The five webhooks_* tables are created by the generated migration at version 23, rendered from
-- platform-go's own DDL. This file reconciles the application's side of the model with them.

-- =============================================================================
-- 1. Trigger configs hold event types, not catalog row IDs
-- =============================================================================

-- Every existing trigger config is unreachable and always has been. trigger_event referenced
-- webhook_trigger_events.id, whose values are randomly generated identifiers, while the fan-out
-- looked webhooks up by the event type *string* ("meal_plan_created"). A random ID can never
-- equal one of those, so no webhook has ever matched an event and none of these rows has ever
-- caused a delivery.
--
-- They are deleted rather than migrated because there is nothing to migrate them to: the catalog
-- row a config points at carries an operator-chosen name that need not be, and generally is not,
-- one of the event types this application publishes. Guessing a mapping would produce
-- subscriptions nobody asked for, on webhooks that are about to start working for the first
-- time.
DELETE FROM webhook_trigger_configs;

ALTER TABLE webhook_trigger_configs
    DROP CONSTRAINT IF EXISTS webhook_trigger_configs_trigger_event_fkey;

COMMENT ON COLUMN webhook_trigger_configs.trigger_event IS
    'A catalog event type, e.g. meal_plan_created. Validated against the generated catalog in internal/domain/webhooks/catalog.';

DROP TABLE IF EXISTS webhook_trigger_events;

-- =============================================================================
-- 2. Only POST and application/json are delivered
-- =============================================================================

-- The delivery worker POSTs JSON. The other enum values described requests nobody was making —
-- a GET carries no body, so there is nothing to sign or to receive — and a webhook row still
-- claiming one would misreport what it will actually be sent.
--
-- The enum types keep their other values: dropping one from a Postgres enum means recreating the
-- type and rewriting every column that uses it, which is a table rewrite to remove something
-- nothing can set any more. The API rejects them on the way in.
UPDATE webhooks SET method = 'POST' WHERE method <> 'POST';
UPDATE webhooks SET content_type = 'application/json' WHERE content_type <> 'application/json';

-- =============================================================================
-- 3. No endpoints are backfilled
-- =============================================================================

-- Existing webhooks deliberately get no row in webhooks_endpoints here.
--
-- An endpoint needs a signing secret, and generating one in SQL would mean either requiring the
-- pgcrypto extension in every environment or falling back to random(), which is not a CSPRNG. A
-- weak secret would be harmless only for as long as the endpoint stayed inert, and it would not:
-- the moment the owner subscribed it to an event through the API it would start signing real
-- deliveries with it.
--
-- So the secret is minted in Go, on demand. RotateWebhookSecret registers the endpoint if it is
-- missing, which makes "adopt a webhook that predates delivery working" and "replace a secret I
-- lost" the same operation. Nothing is lost by waiting: these webhooks have never delivered, and
-- their owners hold no working secret today either.
