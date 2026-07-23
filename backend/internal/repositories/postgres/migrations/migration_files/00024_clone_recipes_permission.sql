-- =============================================================================
-- Seed the clone.recipes permission and grant it to account_member.
--
-- The CloneRecipesPermission ("clone.recipes") gates the CloneRecipe RPC, which previously
-- reused ReadRecipesPermission. Recipe cloning is a normal member action (it creates a new
-- recipe owned by the caller), so it is granted to role_account_member, mirroring
-- create.recipes. Session permissions are loaded from the database, so the permission must
-- be seeded here or every caller gets PermissionDenied on CloneRecipe.
-- =============================================================================

INSERT INTO permissions (id, name, description) VALUES
    ('d9clone0recipes00000', 'clone.recipes', 'Clone recipes')
ON CONFLICT DO NOTHING;

INSERT INTO user_role_permissions (id, role_id, permission_id) VALUES
    ('d9clone0recipes00001', 'role_account_member', 'd9clone0recipes00000')
ON CONFLICT DO NOTHING;
