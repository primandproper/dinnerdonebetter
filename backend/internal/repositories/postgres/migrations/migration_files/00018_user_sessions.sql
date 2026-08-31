-- User Sessions Migration
-- Active login session tracking for users
--
-- SUPERSEDED. The table below is dropped at version 35, which replaces it with
-- ddb_sessions, rendered from platform-go's own DDL. Nothing here is read by any running
-- code; the file stays because renumbering or removing an applied migration is how goose
-- loses track of what a database has already run.

CREATE TABLE IF NOT EXISTS user_sessions (
    id TEXT NOT NULL PRIMARY KEY,
    belongs_to_user TEXT NOT NULL REFERENCES users("id") ON DELETE CASCADE,
    session_token_id TEXT NOT NULL,
    refresh_token_id TEXT NOT NULL,
    client_ip TEXT DEFAULT ''::TEXT NOT NULL,
    user_agent TEXT DEFAULT ''::TEXT NOT NULL,
    device_name TEXT DEFAULT ''::TEXT NOT NULL,
    login_method TEXT DEFAULT ''::TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user ON user_sessions (belongs_to_user) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_sessions_token_id ON user_sessions (session_token_id) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_sessions_refresh_token_id ON user_sessions (refresh_token_id) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires ON user_sessions (expires_at);
