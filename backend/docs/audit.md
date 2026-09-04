# Audit log

The audit log is the durable, queryable, tamper-evident record of who did what to
which resource. It is not logging: a log is a best-effort account written beside
the work, this is a record written *inside* it, in a table the database refuses to
edit, chained so that removing or altering an entry is detectable afterwards.

It is built on `platform-go/v13`'s `audit` package. The platform owns the schema,
the hash chain, and retention; this repository owns the vocabulary — an entry
belongs to a *user* and usually to an *account*, where the platform speaks of an
*actor* and a *scope*.

The tables are `ddb_audit_log_entries` and `ddb_audit_log_chains`. The prefix is
`audit.TablePrefix`, and it is deliberate: the platform's default renders
`audit_log_entries`, which is the name the hand-rolled log this replaced already
held. Its DDL says `CREATE TABLE IF NOT EXISTS`, so against any database that ever
applied the old migration the new schema would be a silent no-op and the audit code
would then run against the wrong columns — and goose records the old migration as
applied, so deleting its file does not undo it. A prefix makes the two unable to
collide. The constant is referenced by the migration that creates the tables, the
Recorder that writes them, the Reader that queries them, and the Sweeper that
prunes them; all four have to agree, and a prefix that differed between the writer
and the reader is the one misconfiguration that stays invisible until somebody asks
the log a question and gets an empty answer.

## Writing an entry

`Record` takes the caller's query executor, which is the whole design. An entry
commits with the change it describes or not at all:

```go
return q.WithTransaction(ctx, func(tx database.Tx) error {
    if err := q.generatedQuerier.UpdateRecipe(ctx, tx, params); err != nil {
        return err
    }

    changes, err := audit.Diff(before, after)
    if err != nil {
        return err
    }

    return q.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
        ResourceType:     resourceTypeRecipes,
        RelevantID:       after.ID,
        EventType:        audit.AuditLogEventTypeUpdated,
        BelongsToUser:    userID,
        BelongsToAccount: &accountID,
        Changes:          changes,
    })
})
```

There is no way to record outside a transaction by accident: holding a
`database.Tx` from `WithTransaction` means you are already in one.

### Almost always, use `RecordAndEmit`

A write that records an entry nearly always owes a data change event too, and every
repository holds a `recording.Recorder` for exactly that pair:

```go
return q.WithTransaction(ctx, func(tx database.Tx) error {
    if err := q.generatedQuerier.UpdateRecipe(ctx, tx, params); err != nil {
        return err
    }

    return q.recorder.RecordAndEmit(ctx, tx, logger, &audit.AuditLogEntry{
        ResourceType: resourceTypeRecipes,
        RelevantID:   after.ID,
        EventType:    audit.AuditLogEventTypeUpdated,
    }, mealplanning.RecipeUpdatedServiceEventType, accountID, map[string]any{
        mealplanningkeys.RecipeIDKey: after.ID,
    })
})
```

The two used to be written out as separate blocks at every write, which made the
easiest mistake to make the one nothing catches: skip the entry and the row has no
provenance, and the chain does not notice, because a chain records what it was
given; skip the event and the search index goes stale and no webhook fires. Neither
leaves anything behind to find later. One call cannot half-happen.

Reach past it for `Record` alone only where the pair genuinely does not apply — an
entry with no event of its own, or a transaction recording several entries at once,
which `Record`'s variadic form handles and `RecordAndEmit` deliberately does not.

Four packages are exceptions to "in the transaction that performed the write", and
they are exactly the four whose writes are an adopted platform store: `comments`,
`issuereports`, `settings` and `waitlists`. Those stores own their transactions and do
not lend them out, so each of these repositories calls the store, lets it commit, and
then opens a second transaction to record — an entry the database refuses fails the
call but no longer takes the row down with it. The gap is narrow and one-directional:
a row can exist with no entry, but no entry can name a row that was not written.

That is a property of the stores, not of `RecordAndEmit`, and each closes when platform
accepts a caller's transaction — filed upstream as platform-go #457 (comments, fixed
but unreleased), #458 (waitlists), #460 (settings) and #465 (issuereports). Tracked
locally on #1419, which lists what deletes here when each lands.

The recorder is one type in `internal/repositories/postgres/recording` rather than a
method on each repository, because the body was the same body in all nine of them and
a rule stated nine times is a rule that can be restated wrongly once. It is built from
the repository's own named tracer, so a span it raises is still attributed to the
package whose write raised it. Its doc comment records why the pair is not in
platform-go: both halves are this application's vocabulary over platform's engines, so
a platform-side version would be generic over two types whose bodies are one call
each.

`Record` is variadic. A transaction touching three resources should pass three
entries to one call rather than making three calls — one chain-head lookup and one
INSERT instead of three of each, and half the lock hold time on the scope's chain
row. `SwapMealPlanEvents` and `MarkUserTwoFactorSecretAsUnverified` do this.

Prefer `audit.Diff(before, after)` to a hand-assembled change map. Hand assembly is
tedious where it is right and silently incomplete where it is wrong, and the field
somebody forgot to add to the map when they added it to the struct is exactly the
field an investigation will want. `Diff` needs both sides to be the same struct
type; where a repository only has an update input, a hand-built map is still the
honest option.

## Scope: which chain an entry lands in

Entries chain per scope, and `Record` holds that scope's chain row for the length
of the caller's transaction. Everything sharing a scope therefore serializes.

`audit.ScopeFor` decides, and it is the only place that decides:

| Entry has | Scope | Why |
| --- | --- | --- |
| an account | the account ID | the natural tenancy boundary; covers most events |
| no account, a user | the user ID | signup, login, password reset — filing these under the empty scope would put every login in the application behind one row lock |
| neither | `""` | platform-level events, rare enough to serialize |

Reads are unaffected by the fallback: the account read path filters on scope, the
user read path filters on the actor. The cost is that a user's events live in two
chains, so verifying a user's full history means walking both.

On the way back out, a scope that is not the actor's own ID is an account ID. The
row carries one scope column rather than a discriminated pair, so that is the only
signal available — and it is exact, because an account never shares an ID with a
user.

## What is guaranteed

A clean `VerifyChain` proves nobody edited, removed, or reordered an entry without
also rewriting every entry after it. It does **not** prove the table was not
replaced wholesale by a consistent forgery — nothing self-contained can. The answer
to that is to publish head hashes somewhere this database's owner does not control;
`Record` writes each entry's `Hash` back into the value you passed, which is what
you would publish. Nothing does that today.

Three mechanisms, in decreasing order of how much they prove:

- **Append-only triggers.** `ddb_audit_log_entries` rejects `UPDATE` outright, at the
  database. Editing a recorded entry is not something the chain reveals after the
  fact, it is something that cannot happen. Installed by the migration; asserted by
  `TestQuerier_Migrate/audit_entries_reject_updates` against real Postgres.
- **The hash chain.** Each entry carries its predecessor's hash and its own hash
  over that plus its own contents. This is what covers *deletion*, which the
  triggers deliberately do not — see retention below.
- **`UNIQUE (scope, seq)`.** A forked chain is not something a verifier has to
  detect, it is something the table cannot hold. Two writers racing on one scope
  both compute the same next position and the index refuses the second.

If the deployment can revoke `UPDATE` and `DELETE` from the application role, do
that as well. It is strictly stronger than the triggers, and it stops the deletions
they cannot.

## Retention

The `Sweeper` runs in the scheduler process as the `audit_retention_sweeper`
scheduled job, and prunes entries past **two years**
(`SchedulerConfig.Audit.Retention`, `DINNER_DONE_BETTER_AUDIT_RETENTION`).

A registered job rather than the Sweeper's own `Run` loop, so that exactly one
replica prunes per tick because the distributed lock says so rather than because
the deployment happens to run one replica. It is scheduled overnight and has no
`RunOnStart`: the other jobs' first run is catch-up work, and this one's would be a
deletion.

Two years rather than the platform's seven-year default: seven is the window the
regulations that ask for an audit log in the first place tend to name, and this
application is under none of them. Two still covers a dispute, an incident review,
or a question about an account somebody closed last year. **Widening the window
later only affects what has not already been swept** — retention deletes, and a
window that was too short is not recoverable by lengthening it afterwards.

`DELETE` stays permitted on the entries table for exactly this reason: retention
has to remove aged entries, and no trigger can tell that sweep apart from an
attacker. The sweeper takes two precautions an ordinary reaper would not. It
removes only a *prefix* of a scope's chain, never a row from the middle, so the
survivors stay contiguous and verifiable against each other. And it records the
hash of the last entry it removed as that scope's watermark, in the same
transaction as the delete, so the oldest surviving entry still links to something
and `Verify` can tell retention's gap from a deletion.

## Redaction

`audit.Redactions` (in `internal/domain/audit/redaction.go`) declares what never
becomes durable, keyed by resource type, with the empty key applying to every
resource type.

It lives in code rather than in a config file on purpose: a bearer token is a bearer
token in every environment, and a policy that could be relaxed by an environment
variable is one bad rollout away from writing secrets into the one table designed
to be immutable and kept for years. Filtering at query time is not the same thing —
by then the value is written.

The catch-all is the important half. It is a rule about the *word*, not about one
table: a field named `password` is a password wherever it shows up, including in a
`Diff` of a struct nobody thought about. A resource type's own rules can only add to
the catch-all; there is no way to opt back out of it.

`Drop` where even the value's shape is not worth keeping; `Hash` where the question
is "did this change, and is it the same value as that one" — rotating a credential
is a real event and the new credential is not a thing to write down.

The static counterpart is the `audit:"-"` struct tag, which keeps a field out of
every `Diff` wherever it appears. `identity.Account.Members` carries it: auditing an
account update would otherwise copy the whole membership roster, including user
email addresses and birthdays, into the audit table on every update. A field tagged
`json:"-"` — `User.HashedPassword`, `User.TwoFactorSecret` — is skipped by `Diff`
already.

## Reading it

`GetAuditLogEntriesForAccount` / `ForUser` (and the `AndResourceTypes` variants) and
`GetAuditLogEntry` are unchanged in shape and still page with `filtering.QueryFilter`.
`VerifyChain(ctx, scope, from, to)` walks a chain and reports the first break.

The gRPC `AuditLogEntry` now carries `scope`, `seq`, `prev_hash`, `hash`,
`actor_type` and `actor_ip` alongside the fields it always had, so a caller can
check the chain rather than trust this service to have checked it. `ChangeLog`
still carries two strings: values are stored typed, and the rendering to text
happens at the edge.

## Watching it

Alert on **`audit_chain_breaks`**. Everything else the package emits describes
throughput; that one means the log has stopped being evidence. The rest are
`audit_entries_recorded`, `audit_record_latency_ms` (recorded inside somebody's
transaction, so it is lock hold time on their rows), `audit_verifications`,
`audit_entries_pruned`, `audit_sweep_errors`, `audit_sweep_latency_ms`.

No span or log line carries a value from `Changes` — those hold exactly what
redaction exists to keep out of durable storage, and a span exporter is durable
storage.

## Erasure

Deleting a user used to be a single cascade: every `belongs_to_user` foreign key
carried `ON DELETE CASCADE`, the old audit table carried one too, and the entries
went with everything else. These tables have no such key and could not — removing a
row from the middle of a chain is indistinguishable from an attacker removing it,
which is the property the chain exists to provide.

So the audit eraser does the one deletion the structure permits: **whole scopes**,
entries and chain rows together. A scope that disappears entirely leaves no gap in
any surviving chain, because there is nothing left to verify against. It is
platform-go's `dataprivacy/auditerasure`, registered under the `audit` key with the
scope resolver in `internal/domain/audit/privacy`.

`ScopeFor` is what makes that cover most of a departing user's trail:

| Their entries | Scope | Erased? |
| --- | --- | --- |
| In accounts they own | the account | ✅ deleted whole |
| Outside any account (signup, login, password reset) | their user ID | ✅ deleted whole |
| In accounts they merely belong to | somebody else's account | ❌ retained, and reported |

The scopes are resolved **before** the user row is deleted, because the cascade
takes the accounts that name them and a deleted user owns nothing. That ordering is
guaranteed rather than arranged: every eraser shares one transaction and they run in
sorted key order, and `audit` sorts before `identity`. The shared transaction is
also what makes the ordering safe in the other direction — a failure at any point
rolls the whole erasure back, so there is no state in which the audit trail of a
surviving account has been destroyed.

What is retained is counted into the `ErasureOutcome` and logged rather than
silently kept — "we kept some audit entries, on this basis" is something a subject
can be told and a regulator can read, instead of something discovered later.

## Known gaps

These are real and worth fixing; none of them is created by the adoption.

**Unattributed entries.** The platform requires an actor on every entry, and it is
right to: an event with nobody responsible for it is half a record. Most repository
methods have no requester to give — `ArchiveSettingDefinition` takes an ID and nothing
else — and the old schema let `belongs_to_user` be NULL, so the gap predates this.
The three that already carried the acting user in their input (webhook creation,
issue report creation, waitlist signup) now record it; the rest are recorded under
`audit.UnattributedActorID` (`"unattributed"`) with actor type `system`, which makes
the gap countable:

```sql
SELECT resource_type, event_type, count(*)
FROM ddb_audit_log_entries WHERE actor_id = 'unattributed'
GROUP BY 1, 2 ORDER BY 3 DESC;
```

Closing it means threading the requester through those repository signatures, which
is a larger change than adopting the log was.

**Updates that record no changes.** About fifteen `Update*` methods take a fully
populated record and record only *that* something changed, because they never read
the prior state and so have nothing to diff against. Closing this means a `SELECT`
before each `UPDATE`, inside the transaction — a real behaviour change per method,
worth doing deliberately rather than in bulk. `UpdateAccount` is the shape to copy:
it already fetches the prior record and now uses `audit.Diff`.

**No published head hashes.** Tamper *evidence* becomes tamper *proof* only when
the head hash lives somewhere the database's owner does not control. `Record`
returns the hash; nothing publishes it.

**Retained entries survive erasure.** A departing user's actions inside accounts
they merely belonged to cannot be removed without breaking those accounts' chains,
so they are kept and reported. That is the intended trade rather than an oversight
— see Erasure above — but it is a thing to be able to explain to a subject who
asks.
