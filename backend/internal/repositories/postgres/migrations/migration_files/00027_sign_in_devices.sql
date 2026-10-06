-- Where each sign-in came from.
--
-- platform's sign-in records how a login happened and stores nothing about the device behind
-- it, deliberately: what is recorded, under which keys and for how long is the consumer's. This
-- is that record, written by the AfterIssueToken hook on the token's own transaction and read
-- back by the annotator that fills each login's attributes on ListSignIns — so a person looking
-- at where they are signed in sees an address and a browser rather than a list of dates.
--
-- One row per login, keyed by its refresh token family and refreshed as the login is: it says
-- where the login was last renewed from, which is the device holding it now. It lives as long
-- as the login can, and the db-cleaner job deletes it after; a person's rows go with them.
CREATE TABLE IF NOT EXISTS sign_in_devices (
    family_id TEXT NOT NULL PRIMARY KEY,
    belongs_to_user TEXT NOT NULL REFERENCES ddb_identity_users("id") ON DELETE CASCADE,
    ip_address TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    device_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sign_in_devices_user ON sign_in_devices (belongs_to_user);

CREATE INDEX IF NOT EXISTS idx_sign_in_devices_expires_at ON sign_in_devices (expires_at);
