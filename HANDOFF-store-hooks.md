# Handoff: platform store hooks

Delete this file before merging.

## What this branch is

DDB's repository packages used to wrap platform-go stores to write an audit
entry and a data change event beside every write. platform-go now gives each
store a `Hooks` interface (one `After…` method per write, called on the
caller's transaction, updates handed the row before and after), and this
branch replaces every wrapper with a hooks implementation.

- **DDB branch:** `store-hooks` (off `cleanup` at `818eab2b7`)
  - `624b78bd9` — the port: waitlists, settings, comments, issuereports,
    uploadedmedia, notificationsstore, payments, webhooksstore, auth
    (password reset), oauth2clientsstore (now uses the Service's hooks),
    identitystore (adapted to the new signatures, no longer embeds NoopHooks).
  - `3c3932690` — **TEMPORARY** `go.mod` replace pointing at `../../platform-go`.
    Drop it and bump platform-go once the platform PR is released. On another
    machine it only builds if platform-go is checked out on `waitlists-hooks`
    as a sibling of this repo.
- **platform-go branch:** `waitlists-hooks`, PR primandproper/platform-go#1079
  - `107df8f6` waitlists hooks (the worked example)
  - `0cd5e28a` hooks for settings, comments, issuereports, mediaregistry,
    notifications, billing, webhooks, passwordreset; before/after on updates
  - `45046fd5` identity + oauth2clients update hooks take before/after
    (**breaking** — see decisions)
  - The PR description only covers the first commit; update it.

## Conventions the port follows

- DDB hooks types implement the platform interface **outright** (no
  `NoopHooks` embed) so a new platform write breaks the build until someone
  decides what it records. Deliberately unrecorded writes are explicit no-ops
  with a reason.
- Provider constructors kept their signatures, so DI is unchanged
  (exceptions: webhooksstore `ProvideStore`, notificationsstore lost the unused
  `ProvideInbox`/`ProvideRegistry`, oauth2clientsstore `ProvideStore` →
  `ProvideHooks`).
- Recorded entries/events are the same as before, except updates now carry
  `entry.Changes` from `platformaudit.Diff(before, after)`.
- Platform: the before row is read only when hooks are installed; erasures and
  bulk deletes get a count, never rows; writes that take no caller Tx get no hook.

## Impact

DDB repository packages, non-test Go: 4,229 → 3,494 lines (−17%). Biggest:
payments 548→366, waitlists 447→280, notificationsstore 410→291. auth and
identitystore grew slightly. notificationsstore, oauth2clientsstore and
webhooksstore gained integration tests (they had none).

## Test status (as of handoff)

- platform-go: `make lint` clean; every touched package passes on SQLite and on
  Postgres/MySQL containers (`RUN_CONTAINER_TESTS=true`).
- Pre-existing, not from this work: billing and identity SQLite tests flake
  with "database schema has changed (17)" at the same rate on `origin/main`;
  running the whole suite at once can exhaust MySQL containers (the affected
  packages pass alone).
- DDB: builds and vets; all 11 ported packages pass with
  `RUN_CONTAINER_TESTS=true go test ./internal/repositories/postgres/...`.
  `cmd/tools/codegen/converters` fails, identically on clean `cleanup`.
- DDB `make lint` was **not** run.

## Open decisions

1. **Breaking change in platform-go.** `identity.Hooks.AfterUpdateAccount`,
   `identity.Hooks.AfterUpdateUserAccountStatus` and
   `oauth2clients.Hooks.AfterUpdateClient` changed signature; both interfaces
   shipped in v14.2.0. Accept as a deliberate break, or pull commit `45046fd5`
   out of the release (DDB's identitystore/oauth2clientsstore would then need
   reverting to the old signatures).
2. **Personal data in audit diffs** (the audit log is tamper-evident, so hard
   to erase):
   - webhook endpoint `Headers` may hold credentials → suggest `audit:"-"` on
     the platform field;
   - comment `body` and waitlist signup `notes` → suggest hash redactions in
     `backend/internal/domain/audit/redaction.go`
     (e.g. `"comments": {Hash: []string{"body"}}`).
3. Smaller calls (recommendation in brackets):
   - issuereports' transition hook gets before/after because it clears the
     resolution note [keep];
   - settings' value hooks also receive the `*Definition` [keep];
   - auth's Consume diff uses a synthesized before row [drop it];
   - identity's account event `changed` metadata now includes `lastUpdatedAt`
     [filter it out];
   - `identity.AfterUpdateProfile` stays without a before row, for erasure
     reasons [leave];
   - platform `NoopHooks` docs say "embed"; DDB deliberately doesn't [add one
     sentence to the docs naming both choices];
   - webhooksstore no longer records an "archived" entry for an id that matched
     nothing; AddSubscription reviving an archived row is still recorded as a
     creation.

## Next steps

1. Decide 1–3, apply, re-run the tests above plus DDB `make lint`.
2. Update the platform PR description, get it merged and released.
3. Drop `3c3932690`, bump platform-go in `backend/go.mod`, delete this file.

Earlier context from the same session, not yet acted on: the analytics gRPC
proxy (`internal/services/analytics/grpc`, iOS `BackendEventReporter.swift`)
is the ad-blocker-style proxy the "no speculative protective services" rule
says to delete; and platform's `webhooks.Emitter` may make DDB's local
`internal/repositories/postgres/events.Emitter` redundant.
