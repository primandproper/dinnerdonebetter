-- OAuth Domain Migration
-- OAuth2 client and token management

CREATE TABLE IF NOT EXISTS oauth2_clients (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT DEFAULT ''::TEXT NOT NULL,
    client_id TEXT NOT NULL,
    -- Despite the name, this stores the SHA-256 hex digest of the secret, never the
    -- plaintext: the plaintext is returned to the creator exactly once at creation
    -- time, and verification compares digests (constant-time, via the oauth library's
    -- ClientPasswordVerifier). The UNIQUE constraint below is over digests.
    client_secret TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(client_id),
    UNIQUE(client_secret)
);

CREATE TABLE IF NOT EXISTS oauth2_client_tokens (
    id TEXT NOT NULL PRIMARY KEY,
    client_id TEXT NOT NULL REFERENCES oauth2_clients("client_id") ON DELETE CASCADE,
    belongs_to_user TEXT NOT NULL REFERENCES users("id") ON DELETE CASCADE,
    redirect_uri TEXT DEFAULT ''::TEXT NOT NULL,
    code TEXT DEFAULT ''::TEXT NOT NULL,
    code_challenge TEXT DEFAULT ''::TEXT NOT NULL,
    code_challenge_method TEXT DEFAULT ''::TEXT NOT NULL,
    code_created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    code_expires_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + '01:00:00'::INTERVAL) NOT NULL,
    access TEXT DEFAULT ''::TEXT NOT NULL,
    access_created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    access_expires_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + '01:00:00'::INTERVAL) NOT NULL,
    refresh TEXT DEFAULT ''::TEXT NOT NULL,
    refresh_created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    refresh_expires_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + '01:00:00'::INTERVAL) NOT NULL,
    -- Deterministic blind indexes over the plaintext code/access/refresh values.
    -- The code/access/refresh columns themselves are stored encrypted (randomized nonce),
    -- so lookups happen against these HMAC hashes instead.
    code_hash TEXT DEFAULT ''::TEXT NOT NULL,
    access_hash TEXT DEFAULT ''::TEXT NOT NULL,
    refresh_hash TEXT DEFAULT ''::TEXT NOT NULL,
    UNIQUE(belongs_to_user, client_id, code_expires_at, access_expires_at, refresh_expires_at)
);

-- =============================================================================
-- INDEXES FOR OAUTH TABLES
-- =============================================================================

-- OAuth2 clients indexes
CREATE INDEX idx_oauth2_clients_archived_at ON oauth2_clients (archived_at) WHERE archived_at IS NULL;

-- OAuth2 client tokens indexes
CREATE INDEX idx_oauth2_tokens_client_id ON oauth2_client_tokens (client_id);
CREATE INDEX idx_oauth2_tokens_user ON oauth2_client_tokens (belongs_to_user);
CREATE INDEX idx_oauth2_tokens_user_client ON oauth2_client_tokens (belongs_to_user, client_id);
CREATE INDEX idx_oauth2_tokens_code_expires ON oauth2_client_tokens (code_expires_at);
CREATE INDEX idx_oauth2_tokens_access_expires ON oauth2_client_tokens (access_expires_at);
CREATE INDEX idx_oauth2_tokens_refresh_expires ON oauth2_client_tokens (refresh_expires_at);
CREATE INDEX idx_oauth2_tokens_code_hash ON oauth2_client_tokens (code_hash);
CREATE INDEX idx_oauth2_tokens_access_hash ON oauth2_client_tokens (access_hash);
CREATE INDEX idx_oauth2_tokens_refresh_hash ON oauth2_client_tokens (refresh_hash);
