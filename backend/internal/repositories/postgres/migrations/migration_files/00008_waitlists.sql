-- Waitlists are global, service-admin-managed records ("which of our users wants to
-- opt into feature X"); they intentionally carry no ownership column.
CREATE TABLE IF NOT EXISTS waitlists (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    valid_until TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE
);

-- Signups, by contrast, are user-owned opt-ins: belongs_to_user/belongs_to_account are
-- stamped from the caller's session at creation, and belongs_to_user is the
-- authorization anchor — reads and mutations require owner-or-service-admin, and the
-- waitlist-wide signup listing is service-admin-only.
CREATE TABLE IF NOT EXISTS waitlist_signups (
    id TEXT NOT NULL PRIMARY KEY,
    notes TEXT NOT NULL DEFAULT '',
    belongs_to_waitlist TEXT REFERENCES waitlists("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    belongs_to_user TEXT REFERENCES users("id") ON DELETE CASCADE,
    belongs_to_account TEXT REFERENCES accounts("id") ON DELETE CASCADE
);
