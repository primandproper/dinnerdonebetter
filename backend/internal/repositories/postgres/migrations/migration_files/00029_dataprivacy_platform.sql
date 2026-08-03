-- Data privacy: retire user_data_disclosures for the platform request table.
--
-- ddb_dataprivacy_requests is created by the generated migration at version 28, rendered from
-- platform-go's own DDL. This file removes what it replaces.
--
-- One table, not two. The old model split a request across a disclosure row and a report ID
-- naming an object in a bucket, and gained nothing from the pair except the possibility of the
-- two disagreeing about whether an artifact still existed — which they did, because nothing ever
-- deleted the object.

-- =============================================================================
-- 1. In-flight disclosures are not migrated
-- =============================================================================

-- There is nothing to migrate them to. A completed disclosure's value is its artifact, and the
-- artifacts these rows name are unreadable to the new code: they were written as bare base64
-- ciphertext at <reportID>.json.enc, while the platform writes a compressed-then-encrypted
-- object under dataprivacy/exports/ and reverses both on the way out. A migrated row would
-- promise a subject a file that fails to decode.
--
-- Pending rows are not carried across either. The work behind them was a message on a queue that
-- no longer has a consumer, so they would sit in 'pending' forever — which is worse than absent,
-- because 'pending' is a status a subject is told to wait on.
--
-- OPERATOR NOTE: the objects these rows named are now orphaned, and each one contains everything
-- this system knows about one person. Nothing reaps them, because the rows that named them are
-- gone. Empty the disclosure artifact bucket of *.json and *.json.enc at the object-storage
-- layer as part of deploying this. From here on the platform sweeper handles expiry, which is
-- the whole reason this table is going away.

DROP TABLE IF EXISTS user_data_disclosures;

DROP TYPE IF EXISTS user_data_disclosure_status;
