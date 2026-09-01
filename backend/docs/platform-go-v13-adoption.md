# platform-go v13 adoption

Record of the v12 → v13 port, and of what it deliberately left for later.

## What v13 is

88 commits since `v12.0.0`, thirteen of them breaking. The module path bump to `/v13` and the
`go 1.27` directive both happened upstream before the tag ([#302], [#305]), so this port carries a
Go toolchain bump with it.

[#302]: https://github.com/primandproper/platform-go/pull/302
[#305]: https://github.com/primandproper/platform-go/pull/305

The five changes that reached this repo:

| Change | Upstream | What it cost here |
| --- | --- | --- |
| `database.Tx` — a transaction gets its own executor | [#310] | The bulk of the port: 82 `WithTransaction` sites plus propagation |
| querygen fragments take a `Direction` | [#316], [#377], [#381] | 196 call sites, and a `cursor` → `page_cursor` rename |
| `filtering` ceiling moved into the package | [#299], [#311] | `MaxQueryFilterLimit` is now a `uint16` var |
| webhooks: a subscription is a row | [#376], [#395] | `Endpoint.Events` → `Endpoint.Subscriptions` |
| capitalism: the inbound half got a vocabulary | [#315] | `capstripe.Event` → `capitalism.Event` |

[#310]: https://github.com/primandproper/platform-go/pull/310
[#316]: https://github.com/primandproper/platform-go/pull/316
[#377]: https://github.com/primandproper/platform-go/pull/377
[#381]: https://github.com/primandproper/platform-go/pull/381
[#299]: https://github.com/primandproper/platform-go/pull/299
[#311]: https://github.com/primandproper/platform-go/pull/311
[#376]: https://github.com/primandproper/platform-go/pull/376
[#395]: https://github.com/primandproper/platform-go/pull/395
[#315]: https://github.com/primandproper/platform-go/pull/315

## `database.Tx` is the one that needs judgment

`Tx` is `SQLQueryExecutor` plus an unexported marker, produced only by `RunInTransaction`. A
parameter typed `Tx` is a compile-time claim that the caller is inside a transaction.

Most of the propagation is mechanical, but a blanket rewrite gets it wrong in both directions, and
the rule is [#310]'s: a parameter becomes a `Tx` when the write has to commit with something else,
and stays a `SQLQueryExecutor` when it is a read. Both cases turned up here:

- `getRecipeStepByID` is a read called with `readDB`. Widening it was wrong.
- `createRecipeStep` writes a step plus its products, instruments and vessels. Widening it was
  right — those have to commit together.
- `MealPlanFinalizationSagaStarter`'s doc already read "writes a saga instance using the caller's
  transaction". [#310] is the change that turns that sentence into a type.

`database.NewTxForTesting` is the escape hatch for test doubles; `forbidigo` bans it outside
`_test.go`.

## querygen cost far less than it looks

Upstream, [#381] makes a paged list two statements — an ascending/descending pair a store picks
between via `filtering.QueryFilter.SortsDescending`. That pair is emitted only by querygen's own
list *forms*: `StandardCRUD`'s list, `ListQueries`, `JunctionListQueries` and the prefix search.

This repo's paged reads are hand-authored statements that call the fragments directly, so nothing
here inherits the pair. **Zero descending statements appear in the generated corpus.** The cost was
passing `querygen.Ascending` — the zero value, behaviour-identical to what the fragments did before
— at 196 call sites, which changed the generated SQL by exactly one thing: `cursor` →
`page_cursor`, 107 lines across 54 files, surfacing in Go as `Cursor` → `PageCursor` on the params
structs.

sqlc 1.26.0 regenerates the v13 corpus correctly and idempotently, so the bump to platform's 1.31.1
was **not** required and is not in this port.

## The audit narrowing

`audit.Query.ResourceTypes []string` became `ResourceType string` upstream ([#404]) — deliberately:
a bound set cannot sit in a paged read on two of the three dialects.

[#404]: https://github.com/primandproper/platform-go/pull/404

It cost nothing here. The `...AndResourceTypes` methods had no gRPC caller — the wire only ever
carried a singular `resource_type` — so they were dead plumbing through the domain interface,
manager and mock. All four narrowed to a single `resourceType string`.

## The Go 1.27 bump brought its own work

Two things rode in on the toolchain, neither related to platform's API:

- **`go fix` modernized `errors.As` into `errors.AsType`** in `recipe_analyzer.go`. CI runs
  `format_golang`, which runs `go fix`, so this is not optional.
- **golangci-lint v2.10.1 refuses to run against a 1.27 module**, so the pin moved to platform's
  v2.13.1 (`Makefile` and `backend_lint.yaml`). That version reported 371 findings that already
  existed on `main` — 357 `goconst`, 10 `noctx`, 4 `nolintlint` — none caused by v13. They are
  fixed in this port rather than deferred: repeated literals became constants (mostly MCP tool
  property names), `httptest.NewRequest` became `NewRequestWithContext`, and four stale `gosec`
  directives naming rule IDs that no longer exist were removed. `gomodguard` also deprecated in
  favour of `gomodguard_v2`.

The constant extraction was checked against the generators: the rendered SQL and the rendered
config JSON are byte-identical, since every constant holds the value the literal did.

## What this port deliberately does not do

- **No descending pagination.** `sortBy=desc` is still silently ignored, exactly as before. The
  statements to switch to are not emitted for hand-authored lists, so implementing it is real work.
- **No new platform packages adopted.** v13 is largely platform absorbing what this repo
  hand-rolls; none of it is imported yet. See the table below.
- **No `filtering` proto/converter adoption**, so `proto/`, `frontend/` and `ios/` are untouched.
  (Adopted afterwards in #1370 — the row below is kept struck through rather than deleted, because
  the sequencing is the point: v13 landed without it.)
- **No sqlc bump.**

## Adoption opportunities

v13 ships stores this repo still hand-rolls. Each is a chance to delete local code, sized and
sequenced separately — none is required to compile.

| Platform package | What it would replace | Upstream |
| --- | --- | --- |
| `identity` | `postgres/identity` + `identity_*.go` codegen | #306, #317, #341 |
| `settings` | `postgres/settings` | #388 |
| ~~`comments`~~ (adopted, #1375) | `postgres/comments` + `domain/comments` + `comments_*.go` codegen | #450 |
| ~~`issuereports`~~ (adopted, #1377) | `postgres/issuereports` + `domain/issuereports` + `issuereports_*.go` codegen | #449 |
| `waitlists` | `postgres/waitlists` | #452 |
| `billing` | `postgres/payments`, partly | #454 |
| `notifications` (store half) | `postgres/notifications` | #390, #439 |
| ~~`authentication/passwordreset`~~ (adopted, #1372) | `postgres/auth/password_reset_tokens.go` | #387 |
| ~~`sessions`~~ (adopted, #1373) | `postgres/auth/user_sessions.go` | #399, #430 |
| ~~`uploads/registry`~~ (adopted, #1376) | `postgres/uploadedmedia` + `domain/uploadedmedia` + `uploadedmedia_*.go` codegen | #389 |
| ~~`filtering/filteringpb` + `filtering/grpc`~~ (adopted, #1370) | `proto/filtering.proto` + `internal/grpc/converters/query_filter.go` | #311 |
| `oauth2server` resource server | the MCP server's resource-server half | #451 |

`filtering` had the best value-to-risk ratio — it deleted a page-size clamp this repo restated by
hand — and was the only one that reached `frontend/` and `ios/`; #1370 did it.
`authentication/passwordreset` had the best *security* return, which is a different axis: it
moved single use from the call site into the store and stopped the reset token being stored as
itself; #1372 did it. `sessions` was the same axis again — it collapsed a session table kept
beside the sign-in into the one row a revocation actually removes, so a sign-out cannot be read
past; #1373 did it. `comments` was the smallest complete domain, and was taken first as the proof
that a whole domain — table, types, manager, mocks and generated SQL — can be deleted rather than
merely re-backed; #1375 did it. `uploads/registry` was the first adoption whose table other
domains join against, and #1376 is the record of what that costs — see below. `issuereports` was
the first adoption that *added* behavior rather than only re-backing it: the local table had no
status column at all, so the triage lifecycle — open, acknowledged, resolved, declined, with a
guarded move between them — arrived with the store; #1377 did it. `identity` is the largest by far
and should be its own epic.

## What adopting a *referenced* table costs

`uploads/registry` (#1376) is the first store this repository adopted whose table other domains
were reading. Three things followed, none of them visible from the package's own surface:

- **The joins had to go.** sqlc's schema is `migration_files`, and a platform table is rendered by
  a generated migration rather than by a file there — so sqlc cannot see `ddb_uploads_objects` and
  a query joining it does not compile. Identity's user reads used to hydrate an avatar in the same
  statement; they now carry `user_avatars.uploaded_media_id` and read the object separately, which
  is one read per user including inside the paged reads. Meal planning's set-shaped
  `GetUploadedMediaWithIDs` is a read per id for the same reason. Both are honest costs of the
  table moving, and both are paid down by a bulk read upstream rather than by hand-written SQL
  against another package's schema.
- **The foreign keys had to be re-pointed.** Six of this repository's tables reference the media
  row. `DROP TABLE ... CASCADE` drops those constraints, and the migration re-adds them against
  the new table — plus one the platform cannot ship, `owner_id REFERENCES users ON DELETE CASCADE`,
  which is what keeps the single identity eraser covering uploads. See `docs/data-privacy.md`.
- **Two operations had no equivalent, for a reason.** The registry has no update — every column is
  a fact about bytes already in a bucket — so `UpdateUploadedMedia` went, along with its
  permission. It has no bulk read by id either; that one is a genuine gap and is a loop here.

## Verification

Everything below was run against the released `v13.0.0` tag, not a `replace` directive:

- `make build` — 253 packages
- `make test` — 85 packages, including the Postgres container suites
- `make lint` — 0 issues (containers, queries, Go, shellcheck)
- `make integration_tests` — apiserver and mcpserver suites
- Every generator re-run, with no drift, so CI's `git diff --exit-code` passes
