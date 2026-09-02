-- RBAC Migration
-- Database-driven roles, permissions, role-permission mappings, role hierarchy, and user-role assignments

-- Drop old hardcoded role columns (no deployed DB, no data to migrate)
ALTER TABLE users DROP COLUMN IF EXISTS service_role;
ALTER TABLE account_user_memberships DROP COLUMN IF EXISTS account_role;

-- =============================================================================
-- TABLE: user_roles
-- =============================================================================

CREATE TABLE IF NOT EXISTS user_roles (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT DEFAULT ''::TEXT NOT NULL,
    scope TEXT NOT NULL CHECK (scope IN ('service', 'account')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(name)
);

CREATE INDEX idx_user_roles_archived ON user_roles (archived_at) WHERE archived_at IS NULL;

-- =============================================================================
-- TABLE: permissions
-- =============================================================================

CREATE TABLE IF NOT EXISTS permissions (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT DEFAULT ''::TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(name)
);

CREATE INDEX idx_permissions_archived ON permissions (archived_at) WHERE archived_at IS NULL;
CREATE INDEX idx_permissions_name ON permissions (name) WHERE archived_at IS NULL;

-- =============================================================================
-- TABLE: user_role_permissions (many-to-many role<->permission)
-- =============================================================================

CREATE TABLE IF NOT EXISTS user_role_permissions (
    id TEXT NOT NULL PRIMARY KEY,
    role_id TEXT NOT NULL REFERENCES user_roles("id") ON DELETE CASCADE,
    permission_id TEXT NOT NULL REFERENCES permissions("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(role_id, permission_id)
);

CREATE INDEX idx_user_role_permissions_role ON user_role_permissions (role_id) WHERE archived_at IS NULL;
CREATE INDEX idx_user_role_permissions_permission ON user_role_permissions (permission_id) WHERE archived_at IS NULL;

-- =============================================================================
-- TABLE: user_role_assignments (maps users to roles, optional account scope)
-- =============================================================================

CREATE TABLE IF NOT EXISTS user_role_assignments (
    id TEXT NOT NULL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users("id") ON DELETE CASCADE,
    role_id TEXT NOT NULL REFERENCES user_roles("id") ON DELETE CASCADE,
    account_id TEXT REFERENCES accounts("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX idx_user_role_assignments_unique_with_account
    ON user_role_assignments (user_id, role_id, account_id)
    WHERE archived_at IS NULL AND account_id IS NOT NULL;

CREATE UNIQUE INDEX idx_user_role_assignments_unique_without_account
    ON user_role_assignments (user_id, role_id)
    WHERE archived_at IS NULL AND account_id IS NULL;

CREATE INDEX idx_user_role_assignments_user ON user_role_assignments (user_id) WHERE archived_at IS NULL;
CREATE INDEX idx_user_role_assignments_user_account ON user_role_assignments (user_id, account_id) WHERE archived_at IS NULL;
CREATE INDEX idx_user_role_assignments_account ON user_role_assignments (account_id) WHERE archived_at IS NULL AND account_id IS NOT NULL;

-- =============================================================================
-- TABLE: user_role_hierarchy (child inherits parent's permissions)
-- =============================================================================

CREATE TABLE IF NOT EXISTS user_role_hierarchy (
    id TEXT NOT NULL PRIMARY KEY,
    parent_role_id TEXT NOT NULL REFERENCES user_roles("id") ON DELETE CASCADE,
    child_role_id TEXT NOT NULL REFERENCES user_roles("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(parent_role_id, child_role_id),
    CHECK(parent_role_id != child_role_id)
);

CREATE INDEX idx_user_role_hierarchy_child ON user_role_hierarchy (child_role_id) WHERE archived_at IS NULL;
CREATE INDEX idx_user_role_hierarchy_parent ON user_role_hierarchy (parent_role_id) WHERE archived_at IS NULL;

-- =============================================================================
-- SEED DATA: roles
-- =============================================================================

INSERT INTO user_roles (id, name, description, scope) VALUES
    ('role_service_user', 'service_user', 'Default non-admin user', 'service'),
    ('role_service_admin', 'service_admin', 'Full service administrator', 'service'),
    ('role_service_data_admin', 'service_data_admin', 'Service data management', 'service'),
    ('role_account_member', 'account_member', 'Basic account member', 'account'),
    ('role_account_admin', 'account_admin', 'Account administrator', 'account');

-- =============================================================================
-- SEED DATA: role hierarchy
-- =============================================================================

INSERT INTO user_role_hierarchy (id, parent_role_id, child_role_id) VALUES
    ('d74elbkn9qd4nv86eo20', 'role_account_member', 'role_account_admin'),
    ('d74elbkn9qd4nv86eo2g', 'role_account_admin', 'role_service_admin');

-- =============================================================================
-- SEED DATA: core permissions (non-mealplanning)
-- =============================================================================

INSERT INTO permissions (id, name, description) VALUES
    -- admin
    ('d74epksn9qd63pdj432g', 'admin.read_user_data', 'Read user data (admin)'),
    ('d74epksn9qd63pdj4330', 'queues.publish.message', 'Publish arbitrary queue messages'),
    -- analytics
    ('d74epksn9qd63pdj433g', 'report.analytics_events', 'Report analytics events'),
    -- auth
    ('d74epksn9qd63pdj4340', 'update.user_status', 'Update user account status'),
    ('d74epksn9qd63pdj434g', 'imitate.user', 'Impersonate users'),
    ('d74epksn9qd63pdj4350', 'read.user', 'Read user data'),
    ('d74epksn9qd63pdj435g', 'search.user', 'Search users'),
    ('d74epksn9qd63pdj4360', 'archive.user', 'Archive users'),
    ('d74epksn9qd63pdj436g', 'manage.user_sessions', 'Manage user sessions'),
    -- identity
    ('d74epksn9qd63pdj4370', 'update.account', 'Update account'),
    ('d74epksn9qd63pdj437g', 'archive.account', 'Archive account'),
    ('d74epksn9qd63pdj4380', 'transfer.account', 'Transfer account ownership'),
    ('d74epksn9qd63pdj438g', 'account.add.member', 'Invite user to account'),
    ('d74epksn9qd63pdj4390', 'account.membership.modify', 'Modify member permissions'),
    ('d74epksn9qd63pdj439g', 'remove_member.account', 'Remove member from account'),
    -- oauth
    ('d74epksn9qd63pdj43a0', 'create.oauth2_clients', 'Create OAuth2 clients'),
    ('d74epksn9qd63pdj43ag', 'read.oauth2_clients', 'Read OAuth2 clients'),
    ('d74epksn9qd63pdj43b0', 'archive.oauth2_clients', 'Archive OAuth2 clients'),
    -- settings
    --
    -- The ids are the ones the service_settings names carried, because the
    -- capabilities are the same capabilities under the nouns platform's store
    -- uses: a definition is what a service setting was, and a value is what a
    -- configuration was. Keeping them is what leaves every grant below pointing
    -- at the permission it was written for.
    --
    -- Two do not survive. search.service_settings went with the search endpoint,
    -- which platform's store has no read behind; update.service_setting_configurations
    -- went because setting a value is one write — the store converges on the row,
    -- so a first answer and a changed one are the same statement.
    ('d74epksn9qd63pdj43bg', 'create.setting_definitions', 'Create setting definitions'),
    ('d74epksn9qd63pdj43c0', 'read.setting_definitions', 'Read setting definitions'),
    ('d74epksn9qd63pdj45kg', 'update.setting_definitions', 'Update setting definitions'),
    ('d74epksn9qd63pdj43d0', 'archive.setting_definitions', 'Archive setting definitions'),
    ('d74epksn9qd63pdj43dg', 'create.setting_values', 'Create setting values'),
    ('d74epksn9qd63pdj43e0', 'read.setting_values', 'Read setting values'),
    ('d74epksn9qd63pdj43f0', 'archive.setting_values', 'Archive setting values'),
    -- notifications
    ('d74epksn9qd63pdj43fg', 'create.user_notifications', 'Create user notifications'),
    ('d74epksn9qd63pdj43g0', 'read.user_notifications', 'Read user notifications'),
    ('d74epksn9qd63pdj43gg', 'update.user_notifications', 'Update user notifications'),
    ('d74epksn9qd63pdj43h0', 'create.user_device_tokens', 'Create user device tokens'),
    ('d74epksn9qd63pdj43hg', 'read.user_device_tokens', 'Read user device tokens'),
    ('d74epksn9qd63pdj43i0', 'archive.user_device_tokens', 'Archive user device tokens'),
    -- webhooks
    ('d74epksn9qd63pdj43ig', 'create.webhooks', 'Create webhooks'),
    ('d74epksn9qd63pdj43j0', 'read.webhooks', 'Read webhooks'),
    ('d74epksn9qd63pdj43jg', 'update.webhooks', 'Update webhooks'),
    ('d74epksn9qd63pdj43k0', 'archive.webhooks', 'Archive webhooks'),
    ('d74epksn9qd63pdj43kg', 'create.webhook_trigger_configs', 'Create webhook trigger configs'),
    ('d74epksn9qd63pdj43l0', 'archive.webhook_trigger_configs', 'Archive webhook trigger configs'),
    ('d74epksn9qd63pdj43lg', 'create.webhook_trigger_events', 'Create webhook trigger events'),
    ('d74epksn9qd63pdj43m0', 'read.webhook_trigger_events', 'Read webhook trigger events'),
    ('d74epksn9qd63pdj43mg', 'update.webhook_trigger_events', 'Update webhook trigger events'),
    ('d74epksn9qd63pdj43n0', 'archive.webhook_trigger_events', 'Archive webhook trigger events'),
    -- uploaded media
    ('d74epksn9qd63pdj43ng', 'create.uploaded_media', 'Create uploaded media'),
    ('d74epksn9qd63pdj43o0', 'read.uploaded_media', 'Read uploaded media'),
    ('d74epksn9qd63pdj43og', 'update.uploaded_media', 'Update uploaded media'),
    ('d74epksn9qd63pdj43p0', 'archive.uploaded_media', 'Archive uploaded media'),
    -- issue reports
    ('d74epksn9qd63pdj43pg', 'create.issue_reports', 'Create issue reports'),
    ('d74epksn9qd63pdj43q0', 'read.issue_reports', 'Read issue reports'),
    ('d74epksn9qd63pdj43qg', 'update.issue_reports', 'Update issue reports'),
    ('d74epksn9qd63pdj43r0', 'archive.issue_reports', 'Archive issue reports'),
    -- audit
    ('d74epksn9qd63pdj43rg', 'read.audit_log_entries', 'Read audit log entries'),
    -- comments
    ('d74epksn9qd63pdj43s0', 'create.comments', 'Create comments'),
    ('d74epksn9qd63pdj43sg', 'read.comments', 'Read comments'),
    ('d74epksn9qd63pdj43t0', 'update.comments', 'Update comments'),
    ('d74epksn9qd63pdj43tg', 'archive.comments', 'Archive comments'),
    -- waitlists
    ('d74epksn9qd63pdj43u0', 'create.waitlists', 'Create waitlists'),
    ('d74epksn9qd63pdj43ug', 'read.waitlists', 'Read waitlists'),
    ('d74epksn9qd63pdj43v0', 'update.waitlists', 'Update waitlists'),
    ('d74epksn9qd63pdj43vg', 'archive.waitlists', 'Archive waitlists'),
    ('d74epksn9qd63pdj4400', 'create.waitlist_signups', 'Create waitlist signups'),
    ('d74epksn9qd63pdj440g', 'read.waitlist_signups', 'Read waitlist signups'),
    ('d74epksn9qd63pdj4410', 'update.waitlist_signups', 'Update waitlist signups'),
    ('d74epksn9qd63pdj441g', 'archive.waitlist_signups', 'Archive waitlist signups'),
    -- payments
    ('d74epksn9qd63pdj4420', 'create.checkout_sessions', 'Create checkout sessions'),
    ('d74epksn9qd63pdj442g', 'cancel.subscriptions', 'Cancel subscriptions'),
    ('d74epksn9qd63pdj4430', 'read.purchases', 'Read purchases'),
    ('d74epksn9qd63pdj443g', 'read.payment_history', 'Read payment history'),
    ('d74epksn9qd63pdj4440', 'create.products', 'Create products'),
    ('d74epksn9qd63pdj444g', 'read.products', 'Read products'),
    ('d74epksn9qd63pdj4450', 'update.products', 'Update products'),
    ('d74epksn9qd63pdj445g', 'archive.products', 'Archive products'),
    ('d74epksn9qd63pdj4460', 'create.subscriptions', 'Create subscriptions'),
    ('d74epksn9qd63pdj446g', 'read.subscriptions', 'Read subscriptions'),
    ('d74epksn9qd63pdj4470', 'update.subscriptions', 'Update subscriptions'),
    ('d74epksn9qd63pdj447g', 'archive.subscriptions', 'Archive subscriptions'),
    -- data privacy
    ('d74epksn9qd63pdj45h0', 'create.user_data_reports', 'Request an export of one''s own data'),
    ('d74epksn9qd63pdj45hg', 'read.user_data_reports', 'Read one''s own data privacy requests and artifacts'),
    ('d74epksn9qd63pdj45i0', 'destroy.user_data', 'Request erasure of one''s own data');

-- =============================================================================
-- SEED DATA: core role-permission mappings
-- =============================================================================

-- service_admin: core permissions
INSERT INTO user_role_permissions (id, role_id, permission_id) VALUES
    ('d74epksn9qd63pdj4480', 'role_service_admin', 'd74epksn9qd63pdj432g'),
    ('d74epksn9qd63pdj448g', 'role_service_admin', 'd74epksn9qd63pdj4340'),
    ('d74epksn9qd63pdj4490', 'role_service_admin', 'd74epksn9qd63pdj4350'),
    ('d74epksn9qd63pdj449g', 'role_service_admin', 'd74epksn9qd63pdj435g'),
    ('d74epksn9qd63pdj44a0', 'role_service_admin', 'd74epksn9qd63pdj4360'),
    ('d74epksn9qd63pdj44ag', 'role_service_admin', 'd74epksn9qd63pdj43a0'),
    ('d74epksn9qd63pdj44b0', 'role_service_admin', 'd74epksn9qd63pdj43b0'),
    ('d74epksn9qd63pdj44bg', 'role_service_admin', 'd74epksn9qd63pdj43d0'),
    ('d74epksn9qd63pdj44c0', 'role_service_admin', 'd74epksn9qd63pdj43fg'),
    ('d74epksn9qd63pdj44cg', 'role_service_admin', 'd74epksn9qd63pdj434g'),
    ('d74epksn9qd63pdj44d0', 'role_service_admin', 'd74epksn9qd63pdj436g'),
    ('d74epksn9qd63pdj44dg', 'role_service_admin', 'd74epksn9qd63pdj4330'),
    ('d74epksn9qd63pdj44e0', 'role_service_admin', 'd74epksn9qd63pdj43bg'),
    ('d74epksn9qd63pdj45l0', 'role_service_admin', 'd74epksn9qd63pdj45kg'),
    ('d74epksn9qd63pdj44eg', 'role_service_admin', 'd74epksn9qd63pdj43u0'),
    ('d74epksn9qd63pdj44f0', 'role_service_admin', 'd74epksn9qd63pdj43v0'),
    ('d74epksn9qd63pdj44fg', 'role_service_admin', 'd74epksn9qd63pdj43vg'),
    ('d74epksn9qd63pdj44g0', 'role_service_admin', 'd74epksn9qd63pdj4440'),
    ('d74epksn9qd63pdj44gg', 'role_service_admin', 'd74epksn9qd63pdj444g'),
    ('d74epksn9qd63pdj44h0', 'role_service_admin', 'd74epksn9qd63pdj4450'),
    ('d74epksn9qd63pdj44hg', 'role_service_admin', 'd74epksn9qd63pdj445g'),
    ('d74epksn9qd63pdj44i0', 'role_service_admin', 'd74epksn9qd63pdj4460'),
    ('d74epksn9qd63pdj44ig', 'role_service_admin', 'd74epksn9qd63pdj446g'),
    ('d74epksn9qd63pdj44j0', 'role_service_admin', 'd74epksn9qd63pdj4470'),
    ('d74epksn9qd63pdj44jg', 'role_service_admin', 'd74epksn9qd63pdj447g');

-- account_admin: core permissions
INSERT INTO user_role_permissions (id, role_id, permission_id) VALUES
    ('d74epksn9qd63pdj44k0', 'role_account_admin', 'd74epksn9qd63pdj4370'),
    ('d74epksn9qd63pdj44kg', 'role_account_admin', 'd74epksn9qd63pdj437g'),
    ('d74epksn9qd63pdj44l0', 'role_account_admin', 'd74epksn9qd63pdj4380'),
    ('d74epksn9qd63pdj44lg', 'role_account_admin', 'd74epksn9qd63pdj438g'),
    ('d74epksn9qd63pdj44m0', 'role_account_admin', 'd74epksn9qd63pdj4390'),
    ('d74epksn9qd63pdj44mg', 'role_account_admin', 'd74epksn9qd63pdj439g'),
    ('d74epksn9qd63pdj44n0', 'role_account_admin', 'd74epksn9qd63pdj43ig'),
    ('d74epksn9qd63pdj44ng', 'role_account_admin', 'd74epksn9qd63pdj43jg'),
    ('d74epksn9qd63pdj44o0', 'role_account_admin', 'd74epksn9qd63pdj43k0'),
    ('d74epksn9qd63pdj44og', 'role_account_admin', 'd74epksn9qd63pdj43pg'),
    ('d74epksn9qd63pdj44p0', 'role_account_admin', 'd74epksn9qd63pdj43qg'),
    ('d74epksn9qd63pdj44pg', 'role_account_admin', 'd74epksn9qd63pdj43r0'),
    ('d74epksn9qd63pdj44q0', 'role_account_admin', 'd74epksn9qd63pdj43kg'),
    ('d74epksn9qd63pdj44qg', 'role_account_admin', 'd74epksn9qd63pdj43l0'),
    ('d74epksn9qd63pdj44r0', 'role_account_admin', 'd74epksn9qd63pdj43lg'),
    ('d74epksn9qd63pdj44rg', 'role_account_admin', 'd74epksn9qd63pdj43m0'),
    ('d74epksn9qd63pdj44s0', 'role_account_admin', 'd74epksn9qd63pdj43mg'),
    ('d74epksn9qd63pdj44sg', 'role_account_admin', 'd74epksn9qd63pdj43n0'),
    ('d74epksn9qd63pdj44t0', 'role_account_admin', 'd74epksn9qd63pdj4420'),
    ('d74epksn9qd63pdj44tg', 'role_account_admin', 'd74epksn9qd63pdj442g'),
    ('d74epksn9qd63pdj44u0', 'role_account_admin', 'd74epksn9qd63pdj4430'),
    ('d74epksn9qd63pdj44ug', 'role_account_admin', 'd74epksn9qd63pdj443g'),
    ('d74epksn9qd63pdj44v0', 'role_account_admin', 'd74epksn9qd63pdj446g');

-- account_member: core permissions
INSERT INTO user_role_permissions (id, role_id, permission_id) VALUES
    ('d74epksn9qd63pdj44vg', 'role_account_member', 'd74epksn9qd63pdj433g'),
    ('d74epksn9qd63pdj4500', 'role_account_member', 'd74epksn9qd63pdj43j0'),
    ('d74epksn9qd63pdj450g', 'role_account_member', 'd74epksn9qd63pdj43q0'),
    ('d74epksn9qd63pdj4510', 'role_account_member', 'd74epksn9qd63pdj43rg'),
    ('d74epksn9qd63pdj451g', 'role_account_member', 'd74epksn9qd63pdj43ag'),
    ('d74epksn9qd63pdj4520', 'role_account_member', 'd74epksn9qd63pdj43c0'),
    ('d74epksn9qd63pdj4530', 'role_account_member', 'd74epksn9qd63pdj43ng'),
    ('d74epksn9qd63pdj453g', 'role_account_member', 'd74epksn9qd63pdj43o0'),
    ('d74epksn9qd63pdj4540', 'role_account_member', 'd74epksn9qd63pdj43og'),
    ('d74epksn9qd63pdj454g', 'role_account_member', 'd74epksn9qd63pdj43p0'),
    ('d74epksn9qd63pdj4550', 'role_account_member', 'd74epksn9qd63pdj43dg'),
    ('d74epksn9qd63pdj455g', 'role_account_member', 'd74epksn9qd63pdj43e0'),
    ('d74epksn9qd63pdj456g', 'role_account_member', 'd74epksn9qd63pdj43f0'),
    ('d74epksn9qd63pdj4570', 'role_account_member', 'd74epksn9qd63pdj43s0'),
    ('d74epksn9qd63pdj457g', 'role_account_member', 'd74epksn9qd63pdj43sg'),
    ('d74epksn9qd63pdj4580', 'role_account_member', 'd74epksn9qd63pdj43t0'),
    ('d74epksn9qd63pdj458g', 'role_account_member', 'd74epksn9qd63pdj43tg'),
    ('d74epksn9qd63pdj4590', 'role_account_member', 'd74epksn9qd63pdj43g0'),
    ('d74epksn9qd63pdj459g', 'role_account_member', 'd74epksn9qd63pdj43gg'),
    ('d74epksn9qd63pdj45a0', 'role_account_member', 'd74epksn9qd63pdj43h0'),
    ('d74epksn9qd63pdj45ag', 'role_account_member', 'd74epksn9qd63pdj43hg'),
    ('d74epksn9qd63pdj45b0', 'role_account_member', 'd74epksn9qd63pdj43i0'),
    ('d74epksn9qd63pdj45bg', 'role_account_member', 'd74epksn9qd63pdj4400'),
    ('d74epksn9qd63pdj45c0', 'role_account_member', 'd74epksn9qd63pdj4410'),
    ('d74epksn9qd63pdj45cg', 'role_account_member', 'd74epksn9qd63pdj441g'),
    ('d74epksn9qd63pdj45d0', 'role_account_member', 'd74epksn9qd63pdj43ug'),
    ('d74epksn9qd63pdj45dg', 'role_account_member', 'd74epksn9qd63pdj440g'),
    ('d74epksn9qd63pdj45e0', 'role_account_member', 'd74epksn9qd63pdj4420'),
    ('d74epksn9qd63pdj45eg', 'role_account_member', 'd74epksn9qd63pdj442g'),
    ('d74epksn9qd63pdj45f0', 'role_account_member', 'd74epksn9qd63pdj4430'),
    ('d74epksn9qd63pdj45fg', 'role_account_member', 'd74epksn9qd63pdj443g'),
    ('d74epksn9qd63pdj45g0', 'role_account_member', 'd74epksn9qd63pdj446g'),
    -- Data privacy, granted to the base role every principal inherits. A subject access
    -- request is a right the subject holds, not an account privilege: it is asked for by a
    -- person about themselves, and there is no role a real user could hold that should be
    -- unable to ask.
    ('d74epksn9qd63pdj45j0', 'role_account_member', 'd74epksn9qd63pdj45h0'),
    ('d74epksn9qd63pdj45jg', 'role_account_member', 'd74epksn9qd63pdj45hg'),
    ('d74epksn9qd63pdj45k0', 'role_account_member', 'd74epksn9qd63pdj45i0');
