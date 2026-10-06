# Handoff: platform-go v15 adoption

Delete this file before merging.

## What this branch is

`store-hooks-adoption` began as the port of DDB's repository wrappers onto platform's `Hooks`
interfaces (platform-go #1080). platform then shipped `recording` and a `RecordingHooks` per
store, and the branch became the first consumer of platform-go **v15**: every hand-written hook
this application had over a platform-owned store is deleted in favour of platform's, and the
recording spine underneath them is platform's too.

**None of the libraries are tagged.** That is deliberate: the point is to consume them here
first and find their bugs before a release. `backend/go.mod` therefore carries two local
replaces — `platform-go/v15 => ../../platform-go` and `primitives-go/v2 => ../../primitives-go`
— and builds only with those siblings checked out at or after platform-go `dfa6ad3e` and
primitives-go `e6509109`. `make lint` runs in a container that cannot see them; lint natively
(`golangci-lint run ./...`) and ignore the two `gomoddirectives` hits on the replaces. The
frontend and iOS still reference published `0.0.1` client libraries; moving them to local paths
is step 6 below and has not started.

## Commits, in order

- `3c7e49b32` — import path bump to `/v15`, both replaces. Compiled first try; the whole
  non-container suite passed.
- `f9960dd0e` — the recording spine. `internal/repositories/postgres/events.Emitter` is now a
  thin adapter over platform's `webhooks.Emitter` and `recording.Recorder`, registered through
  platform's own `webhookscfg.NewEmitter` and `recordingcfg.Register` (`FileBySubject`,
  `sessions.PrincipalFromContext` as the `callers.PrincipalExtractor`). `indexevents` is a
  `[]searchsync.Rule` table and platform's side effect derives the index events;
  `datachanges.Message` implements `searchsync.Change`. The nil-inert emitter is gone.
- `45430c8c0` — the webhook catalog is `webhooks.Merge` of this application's generated
  fragment with every adopted package's `EventCatalog()`; excluded events are marked `Internal`
  rather than omitted. `internal/domain/audit/redaction.go` deleted (platform's defaults cover
  it). `events.New` builds the spine for processes assembled outside the injector.
- (this commit) — the ten hook swaps: waitlists, settings, comments, issuereports,
  uploadedmedia→mediaregistry, notificationsstore, payments→billing, webhooksstore,
  oauth2clientsstore, passkeys. Each `client.go` takes a `*recording.Recorder` and installs
  `<pkg>.NewRecordingHooks`; each `hooks.go` is deleted; the DDB event-type constants for those
  nouns are deleted and the catalog regenerated; `localdev.Spine` replaces the seven
  hand-assembled processes' nil emitters.

## What stays local, and why

- **identity, signin, passwordreset hooks** (`identitystore/hooks.go`, `authentication/
  signin_hooks.go`, `auth/password_reset_tokens.go`). Their events drive the async handler's
  mail flows, which route on `datachanges.Message.EventType`. platform's `Emit` puts a bare
  payload on the broker with no event type (**platform-go #1130**), so adopting
  `identity.RecordingHooks` would silently stop verification emails. They move the day #1130
  lands; they already record through the platform Recorder via the adapter.
- `events.EmitIndex` — index an entity without announcing it; **platform-go #1112** asks the
  outbox writer for `EnqueueDerived`.
- `catalog.omittedByPlatform()` — passkeys' fragment leaves its credential events out
  (**platform-go #1131**); the local fragment describes them as `Internal`. `Merge` refuses the
  duplicate when platform carries them, which is the signal to delete it.
- `recording.Recorder` (the adapter) — keeps `RecordAndEmit`'s signature for the mealplanning
  repositories and the three local hooks. Goes with its last caller.

## Behaviour that changed on purpose

- **The actor is the requester.** DDB's hooks filed the row's subject as the audit actor (an
  admin archiving your comment was recorded as you doing it). platform reads the principal off
  the context and files the subject separately; `FileBySubject` puts a subject's entries on
  `tenancy.Of(subjectID)`. Tests that pinned the old behaviour were flipped, not deleted.
- **Erasures record nothing per store** (platform #1123 / #1115): dataprivacy records the
  erasure once.
- **Event vocabulary.** `comment_created` → `comments.comment.created`, etc. The analytics
  allowlist's one platform event (`billing.subscription.created`) cannot flow until #1130.
- Notifications: device re-registration is `Updated` (was `Created`), revocation `Deleted`
  (was `Archived`). Settings: a first `SetValue` is `Created`. Waitlists: confirm/invite/convert
  are all `audit.EventUpdated`; the distinction lives in the webhook event type only. Billing
  purchases and transactions now emit events (the catalog governs subscribability).
- A process with no broker writes its events to the outbox under
  `queuescfg.DefaultDataChangesTopicName` for the worker that has one. Seeders and one-shot
  tools used to announce nothing.

## Found upstream on the way

Filed: platform-go #1130 (broker payload carries no event type), #1131 (credential events
omitted rather than `Internal`). Already implemented from the earlier review: #1105–#1110,
#1113–#1117, #1119–#1128; still open: #1111, #1112, template-go #7.

Not a defect, but a consequence of building against primitives `main`: 65 `staticcheck` hits for
`filteringgrpc.FromProto`, deprecated in favour of `QueryFilterFromProto(in, archiveDecision)`.
That is the `include_archived` seam and a follow-up of its own.

## Test status (at the last commit)

- `go build ./...`, `go vet ./...`: clean.
- `RUN_CONTAINER_TESTS=false go test ./...`: every package passes, including the
  `testing/integration/{apiserver,mcpserver}` suites (which boot against containers regardless
  of the flag).
- `RUN_CONTAINER_TESTS=true go test ./internal/repositories/postgres/... ./internal/build/...
  ./internal/localdev/...`: all pass.
- `golangci-lint run ./...` natively: clean apart from the two `gomoddirectives` hits on the
  replaces and the 65 `FromProto` deprecations named above.

## Next steps

1. Step 3: `indexstamp` + the three per-process syncer lists → `searchsync.Registry`
   (`RegisterIndex`, `PoolSpecs`, `ReindexAll`), including `cmd/tools/search_index_initializer`.
2. Step 4: the six `BuildInjector`s and `cmd/ddb/worker.go` → `service.Register/New/Run`
   (the decorator-store collision that blocked it is gone). Revisit `docs/configuration.md`'s
   ruling against the `*/config` packages first — `service.Config` is that tree.
3. Step 5: `signingrpc.WithAccessTokens`, `rbac.Tiers`, `notifications/mail`, `*/mcp` +
   `oauth2server/mcp.Protect`, `metering.Unbilled`, `filtering.Observe`, `RandomQuery`,
   `billing/http`, `mediaregistry/grpc`, `QueryFilterFromProto`.
4. Step 6: frontend and iOS onto local paths for `platform-client-ts`/`-swift`.
5. When #1130 lands: delete the identity/signin/passwordreset hooks and the adapter.
6. Tag the libraries; drop the replaces; delete this file.
