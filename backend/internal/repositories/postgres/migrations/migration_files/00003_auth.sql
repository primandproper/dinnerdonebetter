-- Auth Domain Migration
-- Password reset functionality
--
-- SUPERSEDED. The table below is dropped at version 34, which replaces it with
-- ddb_password_reset_tokens, rendered from platform-go's own DDL. Nothing here is read by
-- any running code; the file stays because renumbering or removing an applied migration is
-- how goose loses track of what a database has already run.

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id TEXT NOT NULL PRIMARY KEY,
    token TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    redeemed_at TIMESTAMP WITH TIME ZONE,
    belongs_to_user TEXT NOT NULL REFERENCES users("id") ON DELETE CASCADE
);

-- =============================================================================
-- INDEXES FOR AUTH TABLES
-- =============================================================================

-- Password reset tokens indexes
CREATE INDEX idx_password_reset_user ON password_reset_tokens (belongs_to_user);
CREATE INDEX idx_password_reset_token ON password_reset_tokens (token);
CREATE INDEX idx_password_reset_expires ON password_reset_tokens (expires_at);
CREATE INDEX idx_password_reset_unredeemed ON password_reset_tokens (belongs_to_user, expires_at) WHERE redeemed_at IS NULL;
