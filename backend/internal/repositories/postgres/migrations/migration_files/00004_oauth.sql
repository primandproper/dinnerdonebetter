-- OAuth Domain Migration
-- OAuth2 client registrations.
--
-- Only the client registry lives here. Authorization codes, access tokens and refresh
-- tokens belong to the platform's authorization server and live in the ddb_oauth2_*
-- tables migration 33 renders — see internal/domain/oauth.TablePrefix for why they are
-- prefixed. This table stays ours because it is an administered registry with a listing
-- API, permissions and audit behind it, none of which the platform's Store models.

CREATE TABLE IF NOT EXISTS oauth2_clients (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT DEFAULT ''::TEXT NOT NULL,
    client_id TEXT NOT NULL,
    -- Despite the name, this stores the SHA-256 hex digest of the secret, never the
    -- plaintext: the plaintext is returned to the creator exactly once at creation
    -- time, and verification compares digests. The UNIQUE constraint below is over
    -- digests.
    client_secret TEXT NOT NULL,
    -- The exact URIs a code may be sent to, matched byte for byte. An empty array is a
    -- client that can hold a secret but can never complete an authorization request,
    -- which is the correct answer for one registered before this column existed: the
    -- alternative is defaulting to some URI nobody nominated.
    redirect_uris TEXT[] DEFAULT '{}'::TEXT[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(client_id),
    UNIQUE(client_secret)
);

CREATE INDEX idx_oauth2_clients_archived_at ON oauth2_clients (archived_at) WHERE archived_at IS NULL;
