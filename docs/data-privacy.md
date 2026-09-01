# Data Privacy

How a GDPR/CCPA subject access request becomes an export, how that export is protected while it
exists, how it stops existing, and what a right-to-be-forgotten request actually erases.

The machinery is platform-go's `dataprivacy` package. This application supplies three things and
nothing else: **collectors** (what each domain holds about a person), **erasers** (what each
domain removes), and configuration. The state machine, the request table, the artifact packaging,
the fulfillment loop, and the expiry sweep all live upstream.

## The request lifecycle

A **request** is one row in `ddb_dataprivacy_requests`, of type `export` or `erasure`. It replaces
the old `user_data_disclosures` table and the report ID that named an object beside it — one row
that owns its artifact, rather than a pair that could disagree about whether the artifact still
existed.

```text
[*] --> pending      Submit
pending --> processing        claimed by the fulfillment worker
processing --> pending        retryable failure, attempts remaining
processing --> completed      fulfilled
processing --> failed         attempts exhausted
completed --> expired         artifact deleted (exports only)
```

`awaiting_confirmation` and `cancelled` exist upstream but are unreachable here: our
`ConfirmationWindow` is zero, so an erasure is queued on submission and `Confirm` is never needed.
That is the deliberate choice — a confirmation step is worth turning on where an accidental
erasure would be unrecoverable, and ours is initiated by the subject themselves through an
authenticated call.

| Step | Who | What happens |
| ------ | ----- | -------------- |
| Submit | API server (`AggregateUserDataReport` / `DestroyAllUserData`) | Writes a `pending` row, stamps `due_at` from the response window, returns the request. |
| Fulfill | Scheduler (`dataprivacy.Worker`) | Fans out over the registry; writes the artifact, or runs every eraser in one transaction. |
| Delivery | API server (`FetchUserDataReport`) | `Open`s the artifact — decrypt, decompress — and returns the JSON. |
| Expiry | Scheduler (`data_privacy_sweep`) | Deletes the artifact, then clears the reference and marks the request `expired`. |

Submission no longer publishes to a message queue. A request is a row a worker claims, so the
durability that a topic was providing now comes from the table — and a request can no longer be
accepted, acknowledged, and then lost because the broker dropped it. The
`user_data_aggregation_requests` topic and its subscription are gone.

## Collectors: what goes in an export

Each domain implements `dataprivacy.Collector` in `internal/domain/<domain>/privacy`, returning
already-encoded JSON. The library composes the fragments into a document by key and never looks
inside one.

| Key | Package | Covers |
| ----- | --------- | -------- |
| `identity` | `identity/privacy` | The user, their accounts, invitations sent and received |
| `meal_planning` | `mealplanning/privacy` | Recipes, meals, meal plans, ingredient preferences, ratings |
| `webhooks` | `webhooks/privacy` | Webhooks, keyed by account |
| `settings` | `settings/privacy` | User and account setting configurations |
| `notifications` | `notifications/privacy` | In-app notifications |
| `payments` | `payments/privacy` | Subscriptions, purchases, payment transactions |
| `audit_log` | `audit/privacy` | Audit entries recorded about the subject |
| `issue_reports` | `issuereports/privacy` | Issue reports from the subject's accounts |
| `uploaded_media` | `uploadedmedia/privacy` | Media records (not the bytes) |
| `waitlists` | `waitlists/privacy` | Waitlist signups |
| `comments` | platform-go's `comments/privacy` | Comments the subject authored |

Registration happens in one place, `internal/build/dataprivacy/registry.go`. **Adding a domain to
an export is a line there and a collector beside the domain.** It replaces a `UserDataCollection`
struct that imported all eleven domains and that every collector wrote into — a central type
edited on every schema change, in the file most likely to conflict.

The other thing that changed with it: **a partial export is delivered.** A collector that errors
or times out costs its own section. The artifact is still written, its manifest names the missing
sections and why, and `Request.Failures` carries the same information out to the client. Only an
export where *every* collector failed is a hard failure — a document asserting that nothing is
held about a person is the one wrong answer available. Previously any single collector's error
failed the whole export.

A collector that holds nothing returns `nil`, and the section is omitted rather than written as
`null`. An export's sections are the domains that actually held something.

Neither that rule nor the paging beneath it is restated per domain. `dataprivacy.CollectAll`
walks a cursor-paginated read to its end, `dataprivacy.Fragment` turns "did this domain hold
anything" into a fragment or a `nil`, and `dataprivacy.CollectorFor` is both of those wrapped
around a single list read — which is the whole body of the `audit_log`, `uploaded_media`, and
`waitlists` collectors, so those are one call each with no type of their own. `comments` is the
same shape and is not written here at all: platform-go's `comments/privacy` ships it.

A collector that stopped after one page would produce a truncated export that is
well-formed, present in the manifest, and wrong only in the rows nobody can see are missing;
that is the failure these live in platform-go to prevent.

The one hop that stays here is `CollectAcrossAccounts` in `internal/domain/dataprivacy`, because
"a subject's data hangs off the accounts they belong to" is a fact about this schema rather than
about subject access requests. It pages each account through `CollectAll` and concatenates.

## Erasers: what a deletion removes

Three erasers are registered, and they run **serially inside one transaction** along with the
bookkeeping that records the erasure happened. A subject is never left deleted from eight domains
and present in three.

**`audit`** (platform-go's `auditerasure`, registered via config) deletes whole audit scopes
belonging to the subject: the accounts they own, and their own user scope. It cannot delete
entries from the middle of a chain — that would make `audit.Reader.Verify` report tampering for
the rest of that scope's history — so their actions inside *other people's* accounts are reported
as retained, with a stated basis. See `backend/docs/audit.md`.

**`comments`** (platform-go's `comments/privacy`) hard-deletes every comment the subject wrote. It
is registered because comments are the one ordinary store the cascade below cannot reach: the
table this repository used to own carried `belongs_to_user REFERENCES users ON DELETE CASCADE`,
and platform-go's carries a plain `author` column, because that package does not own the directory
people live in and has no table to point a foreign key at. Without this eraser a deletion would
leave every comment the subject wrote in place, live and listable.

It is a hard delete rather than an anonymization, and that is the right way round here. A
comment's body is free text somebody typed, so what has to go is the words; stripping the author
and keeping the sentence would remove the part that is not identifying and keep the part that
might be.

**`identity`** deletes the user row. Every `belongs_to_user` and `belongs_to_account` foreign key
in this schema carries `ON DELETE CASCADE`, so that single `DELETE` is the erasure for every other
domain.

There is deliberately no eraser per domain beyond those. Eleven statements that can only agree
with the one that ran first are eleven places for that agreement to rot. What makes a domain's own
eraser worth writing is retention, anonymization, or — as with comments — no foreign key to
cascade from. Retention is the case that has not arrived yet; the likeliest first instance is
payment records, which tax law generally requires be retained for years. When it does, that domain
registers its own `Eraser` under its own key and reports what it kept and why. Nothing here has to
change for that.

**Ordering is load-bearing.** Erasers run in sorted key order: `audit`, then `comments`, then
`identity`. The audit eraser resolves its scopes by asking which accounts the subject owns, and
that question has no answer once the user row is gone. `comments` sorting before `identity` costs
nothing — there is no foreign key between them any more — and reads correctly anyway.

## The artifact

One artifact is everything the system knows about one person, in one object. It is canonical JSON,
compressed with zstd, then encrypted, written under `dataprivacy/exports/`.

**`Open` is the only delivery path.** There is no signed URL, because the stored object is
ciphertext the recipient has no key for — a subject following that link gets a file they cannot
open, and finds out some days into a statutory window. platform-go enforces this rather than
documenting it: configuring an encryptor makes `Download` refuse with `ErrArtifactEncrypted`. The
server reads, decrypts, decompresses, and returns the bytes over the authenticated gRPC call.

The response carries raw JSON rather than a typed message. The typed one named every domain in the
application, so a schema change anywhere was a proto change, a regeneration in three languages,
and a client release.

**The bucket is private and unversioned.** Both are load-bearing:

- It grants `objectAdmin` to the workload identity service account and nothing else.
- Versioning is off. With versioning on, deleting an artifact only writes a tombstone and the
  object survives as a noncurrent version, so the expiry the sweep exists to enforce would not
  actually happen.

A 14-day lifecycle rule deletes anything the sweep never got to. It is a backstop; the seven-day
expiry is enforced by the sweep.

## The sweep

`data_privacy_sweep` runs hourly on the scheduler, under the same distributed lock as every other
scheduled job, so one replica sweeps per tick. Each pass:

1. Deletes the artifact of every completed export past `expires_at`, **then** clears the reference
   and marks the request `expired`. In that order, so a row can never claim an artifact is gone
   while it is not.
2. Cancels erasures whose confirmation window lapsed (a no-op here — our window is zero).
3. Samples `dataprivacy_requests_overdue`, the count of requests past `due_at` that are still
   unfulfilled.
4. Reaps terminal request records past the retention window. A record of a privacy request is
   itself personal data — it says a named person asked, and when.

**A deployment that runs the worker and not the sweep accumulates artifacts forever**, and nothing
about the request rows suggests otherwise. That is why it is a named registration rather than a
flag.

## Deadlines

`due_at` is stamped at submission from the configured response window — 30 days by default, GDPR's
figure rather than CCPA's 45. `dataprivacy_requests_overdue` is the gauge; alerting on it is an
operator decision, because what counts as an incident is a policy question and depends on which
jurisdiction applies.

For an erasure, a missed deadline is a legal problem rather than an ops one. There was no such
tracking before this.

## Configuration

`internal/services/dataprivacy/config.Config` configures two processes: the API server reads
artifacts, and the scheduler writes and expires them. They must agree on the bucket, the cipher,
the compressor, and the table prefix; a mismatch surfaces as an unreadable export or a sweep that
deletes nothing and reports success, neither of which fails at startup. `EnsurePackaging` and a
pinned table prefix are what keep them from being configurable apart.

| Process    | Env var for the key                                               |
|------------|-------------------------------------------------------------------|
| API server | `DINNER_DONE_BETTER_SERVICE_DATA_PRIVACY_ARTIFACT_ENCRYPTION_KEY` |
| Scheduler  | `DINNER_DONE_BETTER_DATA_PRIVACY_ARTIFACT_ENCRYPTION_KEY`         |

The API server's spelling differs because it reaches this config through `Services`. Both take the
same value, from the `DISCLOSURE_ARTIFACT_ENCRYPTION_KEY` entry of the `api-service-config`
secret, generated by Terraform.

**Rotating the key** makes every artifact written under the old key permanently unreadable. That
is survivable — artifacts expire in seven days and a subject can request a fresh one — but it is
not a no-op, and the sweep still destroys the unreadable objects on schedule since deletion does
not involve the cipher.

Audit erasure is a config flag (`...DATA_PRIVACY_REQUESTS_AUDIT_ERASURE_DISABLED`) rather than a
code change, because "do we erase our own audit records about this person" has a different answer
per jurisdiction. It is on by default: an erasure that silently skipped a store of personal data
would be the more surprising default.

## Upgrade notes

The migration at version 29 drops `user_data_disclosures` without migrating its rows.

Completed disclosures cannot be carried across: their artifacts were written as bare base64
ciphertext at `<reportID>.json.enc`, while the platform writes a compressed-then-encrypted object
under `dataprivacy/exports/` and reverses both on the way out. A migrated row would promise a
subject a file that fails to decode. Pending rows cannot be carried across either — the work
behind them was a message on a queue that no longer has a consumer, so they would sit in `pending`
forever, which is worse than absent.

**Operator action:** the objects those rows named are orphaned by this, and each contains
everything the system knows about one person. Empty the disclosure artifact bucket of `*.json` and
`*.json.enc` at the object-storage layer as part of deploying it. From here on the sweep handles
expiry.

## Related

- `backend/docs/audit.md` — the audit log, and what erasure can and cannot remove from it
- `docs/identity.md` — users, accounts, memberships
- `internal/build/dataprivacy/registry.go` — the one file that knows about every domain
