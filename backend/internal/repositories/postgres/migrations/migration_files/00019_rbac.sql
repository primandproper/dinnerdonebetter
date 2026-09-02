-- RBAC Migration
-- Which roles a principal holds, service-wide or within one account.
--
-- What a role *grants* is not here. That is authorization/database's four tables,
-- rendered and seeded by the migrator from authorization.PlatformPolicy() — see
-- renderAuthorizationDDL. This file used to hold both, and the grant half was ~600
-- lines of INSERT statements restating the permission slices in Go.
--
-- The split is the platform's, and it is the reason a role's permissions can be
-- cached: policy is keyed by role name, so five roles are five entries every
-- principal shares. An assignment names this application's users and accounts,
-- which no platform package can model without owning them, so it stays here.

-- Drop old hardcoded role columns (no deployed DB, no data to migrate)
ALTER TABLE users DROP COLUMN IF EXISTS service_role;
ALTER TABLE account_user_memberships DROP COLUMN IF EXISTS account_role;

-- =============================================================================
-- TABLE: user_role_assignments (maps users to roles, optional account scope)
-- =============================================================================
--
-- role_name rather than a role id. The roles table is rendered by a generated
-- migration, so sqlc — whose schema is this directory — cannot see it, and a
-- statement joining it would not compile. Names are what the platform's own model
-- uses for the same reason: "role names are the identifiers a principal's
-- assignments refer to", which is why that table's name column carries a unique
-- index and why archiving a role keeps its name reserved.
--
-- The foreign key onto that name is added in the generated migration that creates
-- the table, since it cannot be declared before it exists.
--
-- account_id IS NULL means the assignment is service-wide.
CREATE TABLE IF NOT EXISTS user_role_assignments (
    id TEXT NOT NULL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users("id") ON DELETE CASCADE,
    role_name TEXT NOT NULL,
    account_id TEXT REFERENCES accounts("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX idx_user_role_assignments_unique_with_account
    ON user_role_assignments (user_id, role_name, account_id)
    WHERE archived_at IS NULL AND account_id IS NOT NULL;

CREATE UNIQUE INDEX idx_user_role_assignments_unique_without_account
    ON user_role_assignments (user_id, role_name)
    WHERE archived_at IS NULL AND account_id IS NULL;

CREATE INDEX idx_user_role_assignments_user ON user_role_assignments (user_id) WHERE archived_at IS NULL;
CREATE INDEX idx_user_role_assignments_user_account ON user_role_assignments (user_id, account_id) WHERE archived_at IS NULL;
CREATE INDEX idx_user_role_assignments_account ON user_role_assignments (account_id) WHERE archived_at IS NULL AND account_id IS NOT NULL;
