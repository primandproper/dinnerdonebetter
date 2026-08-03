-- Drop the hand-rolled audit log.
--
-- Its replacement is the platform's, created by the generated migration that
-- follows this one under the ddb_ prefix. The two do not collide, so this drop
-- is not required for the new tables to exist — it is here because leaving the
-- old table behind would leave two things called the audit log, one of which
-- nothing writes to and nothing reads from.
--
-- The rows are not carried across. A chain computed at migration time over rows
-- that were never chained proves nothing about the period before the migration:
-- it attests that the backfill read what the table said at that moment, which is
-- exactly the claim in question. Rather than ship a chain that looks like
-- evidence and is not, the log starts here, and what it attests to is everything
-- from this migration forward. See docs/audit.md.

DROP TABLE IF EXISTS audit_log_entries;

DROP TYPE IF EXISTS audit_log_event_type;
