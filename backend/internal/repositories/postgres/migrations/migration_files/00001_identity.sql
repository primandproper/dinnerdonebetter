-- Identity Domain Migration
-- Core user management, accounts, and authentication sessions

CREATE TYPE time_zone AS ENUM (
    'UTC',
    'US/Pacific',
    'US/Mountain',
    'US/Central',
    'US/Eastern'
);

CREATE TYPE invitation_state AS ENUM (
    'pending',
    'cancelled',
    'accepted',
    'rejected'
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT NOT NULL PRIMARY KEY,
    username TEXT NOT NULL,
    email_address TEXT NOT NULL,
    hashed_password TEXT NOT NULL,
    password_last_changed_at TIMESTAMP WITH TIME ZONE,
    requires_password_change BOOLEAN DEFAULT FALSE NOT NULL,
    two_factor_secret TEXT NOT NULL,
    two_factor_secret_verified_at TIMESTAMP WITH TIME ZONE,
    service_role TEXT DEFAULT 'service_user'::TEXT NOT NULL,
    user_account_status TEXT DEFAULT 'unverified'::TEXT NOT NULL,
    user_account_status_explanation TEXT DEFAULT ''::TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    birthday TIMESTAMP WITH TIME ZONE,
    email_address_verification_token TEXT DEFAULT ''::TEXT,
    email_address_verified_at TIMESTAMP WITH TIME ZONE,
    first_name TEXT DEFAULT ''::TEXT NOT NULL,
    last_name TEXT DEFAULT ''::TEXT NOT NULL,
    last_accepted_terms_of_service TIMESTAMP WITH TIME ZONE,
    last_accepted_privacy_policy TIMESTAMP WITH TIME ZONE,
    last_indexed_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(username)
);

CREATE TABLE IF NOT EXISTS accounts (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    billing_status TEXT DEFAULT 'unpaid'::TEXT NOT NULL,
    contact_phone TEXT DEFAULT ''::TEXT NOT NULL,
    payment_processor_customer_id TEXT DEFAULT ''::TEXT NOT NULL,
    subscription_plan_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    belongs_to_user TEXT NOT NULL REFERENCES users("id") ON DELETE CASCADE,
    time_zone time_zone DEFAULT 'US/Central'::time_zone NOT NULL,
    address_line_1 TEXT DEFAULT ''::TEXT NOT NULL,
    address_line_2 TEXT DEFAULT ''::TEXT NOT NULL,
    city TEXT DEFAULT ''::TEXT NOT NULL,
    state TEXT DEFAULT ''::TEXT NOT NULL,
    zip_code TEXT DEFAULT ''::TEXT NOT NULL,
    country TEXT DEFAULT ''::TEXT NOT NULL,
    latitude NUMERIC(14,11),
    longitude NUMERIC(14,11),
    last_payment_provider_sync_occurred_at TIMESTAMP WITH TIME ZONE,
    webhook_hmac_secret TEXT DEFAULT ''::TEXT NOT NULL,
    UNIQUE(belongs_to_user, name)
);

CREATE TABLE IF NOT EXISTS account_user_memberships (
    id TEXT NOT NULL PRIMARY KEY,
    belongs_to_account TEXT NOT NULL REFERENCES accounts("id") ON DELETE CASCADE,
    belongs_to_user TEXT NOT NULL REFERENCES users("id") ON DELETE CASCADE,
    default_account BOOLEAN DEFAULT FALSE NOT NULL,
    account_role TEXT DEFAULT 'account_user'::TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(belongs_to_account, belongs_to_user)
);

CREATE TABLE IF NOT EXISTS account_invitations (
    id TEXT NOT NULL PRIMARY KEY,
    destination_account TEXT NOT NULL REFERENCES accounts("id") ON DELETE CASCADE,
    to_email TEXT NOT NULL,
    to_user TEXT  REFERENCES users("id") ON DELETE CASCADE,
    from_user TEXT NOT NULL  REFERENCES users("id") ON DELETE CASCADE,
    status invitation_state DEFAULT 'pending'::invitation_state NOT NULL,
    note TEXT DEFAULT ''::TEXT NOT NULL,
    status_note TEXT DEFAULT ''::TEXT NOT NULL,
    token TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + '7 days'::INTERVAL) NOT NULL,
    to_name TEXT DEFAULT ''::TEXT NOT NULL,
    UNIQUE(to_user, to_email, from_user, destination_account)
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(token)
);

-- =============================================================================
-- INDEXES FOR IDENTITY TABLES
-- =============================================================================

-- Users table indexes
CREATE INDEX idx_users_archived_at ON users (archived_at) WHERE archived_at IS NULL;
CREATE INDEX idx_users_email_address_active ON users (email_address) WHERE archived_at IS NULL;
CREATE INDEX idx_users_username_active ON users (username) WHERE archived_at IS NULL;
CREATE INDEX idx_users_email_verification_token ON users (email_address_verification_token) WHERE archived_at IS NULL AND email_address_verification_token != '';
CREATE INDEX idx_users_service_role_username ON users (service_role, username) WHERE archived_at IS NULL;
CREATE INDEX idx_users_two_factor_verified ON users (two_factor_secret_verified_at) WHERE archived_at IS NULL;
CREATE INDEX idx_users_indexing_status ON users (last_indexed_at) WHERE archived_at IS NULL;
CREATE INDEX idx_users_active_created_at ON users (created_at) WHERE archived_at IS NULL;
CREATE INDEX idx_users_active_updated_at ON users (last_updated_at) WHERE archived_at IS NULL;
CREATE INDEX idx_users_indexing_needed ON users (last_indexed_at) WHERE archived_at IS NULL;

-- Accounts table indexes
CREATE INDEX idx_accounts_belongs_to_user ON accounts (belongs_to_user) WHERE archived_at IS NULL;
CREATE INDEX idx_accounts_archived_at ON accounts (archived_at) WHERE archived_at IS NULL;
CREATE INDEX idx_accounts_user_name ON accounts (belongs_to_user, name) WHERE archived_at IS NULL;
CREATE INDEX idx_accounts_payment_sync ON accounts (last_payment_provider_sync_occurred_at) WHERE archived_at IS NULL;
CREATE INDEX idx_accounts_billing_status ON accounts (billing_status) WHERE archived_at IS NULL;
CREATE INDEX idx_accounts_user_created_at ON accounts (belongs_to_user, created_at) WHERE archived_at IS NULL;
CREATE INDEX idx_accounts_user_updated_at ON accounts (belongs_to_user, last_updated_at) WHERE archived_at IS NULL;
CREATE INDEX idx_accounts_user_billing ON accounts (belongs_to_user, billing_status) WHERE archived_at IS NULL;

-- Account user memberships indexes
CREATE INDEX idx_memberships_user ON account_user_memberships (belongs_to_user) WHERE archived_at IS NULL;
CREATE INDEX idx_memberships_account ON account_user_memberships (belongs_to_account) WHERE archived_at IS NULL;
CREATE INDEX idx_memberships_default_account ON account_user_memberships (belongs_to_user, default_account) WHERE archived_at IS NULL AND default_account = TRUE;
CREATE INDEX idx_memberships_user_account ON account_user_memberships (belongs_to_user, belongs_to_account) WHERE archived_at IS NULL;
CREATE INDEX idx_memberships_account_role ON account_user_memberships (belongs_to_account, account_role) WHERE archived_at IS NULL;

-- Account invitations indexes
CREATE INDEX idx_invitations_destination_account ON account_invitations (destination_account) WHERE archived_at IS NULL;
CREATE INDEX idx_invitations_from_user ON account_invitations (from_user) WHERE archived_at IS NULL;
CREATE INDEX idx_invitations_to_user ON account_invitations (to_user) WHERE archived_at IS NULL;
CREATE INDEX idx_invitations_to_email ON account_invitations (to_email) WHERE archived_at IS NULL;
CREATE INDEX idx_invitations_token ON account_invitations (token) WHERE archived_at IS NULL;
CREATE INDEX idx_invitations_status ON account_invitations (status) WHERE archived_at IS NULL;
CREATE INDEX idx_invitations_expires_at ON account_invitations (expires_at) WHERE archived_at IS NULL;

-- Sessions indexes
CREATE INDEX idx_sessions_expiry ON sessions (expiry);
CREATE INDEX idx_sessions_created_at ON sessions (created_at);

-- Text search indexes (for efficient LIKE and ILIKE operations)
-- Uncomment if pg_trgm extension is available:
-- CREATE EXTENSION IF NOT EXISTS pg_trgm;
-- CREATE INDEX idx_users_username_trgm ON users USING gin (username gin_trgm_ops) WHERE archived_at IS NULL;

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
