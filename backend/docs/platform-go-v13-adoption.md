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
| ~~`settings`~~ (adopted, #1379) | `postgres/settings` + `domain/settings` + `settings_*.go` codegen | #388 |
| ~~`comments`~~ (adopted, #1375) | `postgres/comments` + `domain/comments` + `comments_*.go` codegen | #450 |
| ~~`issuereports`~~ (adopted, #1377) | `postgres/issuereports` + `domain/issuereports` + `issuereports_*.go` codegen | #449 |
| ~~`waitlists`~~ (adopted, #1378) | `postgres/waitlists` + `domain/waitlists` + `waitlists_*.go` codegen | #452 |
| ~~`billing`~~ (adopted, #1380) | `postgres/payments` + `domain/payments`'s stored types + `payments_*.go` codegen; `capitalism` stays | #454 |
| `notifications` (store half) | `postgres/notifications` | #390, #439 |
| ~~`authentication/passwordreset`~~ (adopted, #1372) | `postgres/auth/password_reset_tokens.go` | #387 |
| ~~`sessions`~~ (adopted, #1373) | `postgres/auth/user_sessions.go` | #399, #430 |
| ~~`uploads/registry`~~ (adopted, #1376) | `postgres/uploadedmedia` + `domain/uploadedmedia` + `uploadedmedia_*.go` codegen | #389 |
| ~~`filtering/filteringpb` + `filtering/grpc`~~ (adopted, #1370) | `proto/filtering.proto` + `internal/grpc/converters/query_filter.go` | #311 |
| `oauth2server` resource server | the MCP server's resource-server half | #451 |
| ~~`authorization/database`~~ (adopted, #1386) | the four RBAC tables in `00019_rbac.sql` plus their seed | #400 |

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
guarded move between them — arrived with the store; #1377 did it. `waitlists` was the second adoption that *added* behavior, and the addition is the point rather
than a side effect: the local tables could hide a signup and could not suppress one, so somebody
who unsubscribed was somebody the next form submission re-subscribed. #1378 did it — see below.
`identity` is the largest by far and should be its own epic.

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

## What adopting a store with an *obligation* costs

`waitlists` (#1378) is the first store this repository adopted whose point is a promise rather
than a shape. The local tables were CRUD: a list with a `valid_until`, a signup with a note and
two ownership columns. What they could not express is the one thing a waitlist has to get right —
somebody who asks to come off a list stays off it. Archiving a signup frees nothing and suppresses
nothing, so the next submission from the same person simply succeeded.

Three things followed:

- **The signup now holds an address, and that changed who may read one.** The contact is the
  session's email — a signup that could name its own address is one anybody could make on
  anybody's behalf, and a suppression anybody could evade — so the list-wide signup read is a read
  of every signatory's email. It was already service-admin-only over gRPC. It was **not** gated on
  the MCP server, which has no role on its token, so the two signup tools were removed there
  rather than shipped as a way to ask a model for a list of addresses. The three catalog tools
  stay.
- **No foreign key could be re-created, and that is the feature.** `uploads/registry` and
  `issuereports` both re-pointed their cascade at `users` when they were adopted. This one cannot:
  a withdrawal blanks `subject_id` to the empty string, so a key there would refuse the withdrawal.
  What replaces the cascade is an eraser whose erasure *is* a withdrawal — see
  `docs/data-privacy.md`.
- **The wire surface grew rather than shrank.** `Invite`, `Convert` and `Withdraw` are new RPCs
  because they are new capabilities; `GetActiveWaitlists` became `GetOpenWaitlists` and
  `WaitlistIsNotExpired` became `WaitlistIsOpen`, because the platform's boundary counts an
  archived list as closed and the old names did not.

Both gaps this turned up are filed upstream as [platform-go #458] rather than papered over
locally: the store's writes own their transactions, so the audit entry and the data change event
land in a second one (the same shape as `comments`' [#457]), and — the one specific to this
package — it ships no `waitlists/privacy`, so the eraser is the consumer's to write against a
`Store` that cannot run it inside the request's transaction or reach an archived signup. Neither
has a local fix that is not a hand-written statement against another package's schema.

## What adopting a store that *narrows* a schema costs

`settings` (#1379) is the third adoption that changed behavior rather than only re-backing it, and
the change is a narrowing: the local `service_setting_configurations` filed a row against a user
**and** an account, and the adopted table files a value against one subject. That was not a
concession to the platform's shape. Two things were wrong with the pair of columns, and one
subject type is what fixes both:

- **The account read returned other people's answers.** `GetServiceSettingConfigurationsForAccount`
  filtered on `belongs_to_account` alone, so any account member holding
  `read.service_setting_configurations` — every member did — was handed every other member's
  personal preferences. The replacement read is the requester's own answers; the administrative
  "who has overridden this setting" is `GetSettingValuesForDefinition`, and it is service-admin
  only.
- **Nothing was ever account-owned anyway.** Every write set both columns from the session, so a
  preference chosen in one account was invisible to the same person in another — a per-person fact
  silently filed per membership.

Three more things followed:

- **The foreign key came back, and it enforces the decision.** platform files a value against a
  subject *type* and an id and leaves the type set open, so `subject_id` is a mixed column in
  general and no key is possible. Using one type makes every `subject_id` a row in `users`, so
  `renderSettingsDDL` re-creates `belongs_to_user`'s `ON DELETE CASCADE` and the single identity
  eraser goes on covering settings — unlike `waitlists`, which had to write an eraser instead. An
  account-owned setting starts by dropping that key.
- **`type` did not survive, and `admin_only` took its job.** The old `setting_type` enum said
  which sort of principal a setting was for; the platform's `kind` says what a value parses as,
  which is a different fact. Nothing enforced `type` — every configuration row carried both
  columns whatever it said — so all it decided was which settings the iOS preferences page listed,
  and `admin_only` decides that now. It is also enforced server-side for the first time: platform
  records `AdminOnly` and deliberately does not act on it, and the handlers do.
- **The wire surface moved to the platform's model.** `ServiceSetting` became `SettingDefinition`,
  `ServiceSettingConfiguration` became `SettingValue`, and `SettingResolution` is new — it is the
  tri-state (the subject chose, the default answered, nobody has) that a bare value cannot express,
  and `ResolveSettings` is what a preferences page now calls instead of joining a catalog against a
  list of values client-side. `SearchForServiceSettings` went: platform's store has no search
  behind it, and no client called it. `notes` went too — platform points a consumer at its own
  table for an annotation about a value, and the only writer set it to `""`.

The transaction gap is filed upstream as [platform-go #460] rather than papered over, along with
the second half of it that does not bite here: the store has no delete for a subject's values, so
a consumer whose `subject_id` is mixed has nothing to build an eraser on.

[platform-go #458]: https://github.com/primandproper/platform-go/issues/458
[#457]: https://github.com/primandproper/platform-go/issues/457
[platform-go #460]: https://github.com/primandproper/platform-go/issues/460

## What adopting a store whose *record* is not its *wire* costs

`billing` (#1380) is the first adoption that replaced half a domain. `capitalism` — the wire to
Stripe and RevenueCat — was already platform's; what the hand-rolled `payments` layer owned was
the four tables the provider's events landed in, and platform's own documentation now draws the
line the same way: *capitalism is the wire and billing is the record.* So the store, the fakes,
the mock, the codegen and every stored type went; the webhook adapters, the manager that reads a
provider's event, the gRPC surface and the account-standing mapping stayed. Scoping the seam was
the whole question the ticket posed, and the answer was that platform had already drawn it.

Four things followed:

- **The vocabulary is capitalism's, and one word changed.** `Subscription.Status` is
  `capitalism.SubscriptionStatus` rather than a five-value enum of this application's, which is
  the same judgment — which of Stripe's words is "cancelled" — made once instead of twice. The
  RevenueCat adapter's mapping table went with it: it had been folding eight platform statuses
  onto five local ones, losing `paused` and `unpaid` on the way. The casualty is a spelling: the
  store writes `canceled`, and a client sending `cancelled` is refused with `InvalidArgument`.
- **The manager shrank to what a store cannot hold.** `PaymentsDataManager` used to proxy every
  read and write with a validation step in front; the store owns those rules now, and the gRPC
  service reads and writes it directly, as `settings` does. What is left is `ProcessWebhookEvent`
  — which provider event changes what about a subscription, and what that does to the account's
  billing standing — which is exactly the half platform's package doc says a consumer still
  writes.
- **Redelivery is in the statement, not the handler.** The three provider-side id columns are
  unique within the scope, so a replayed webhook collides instead of recording a second charge,
  and the status writes are guarded so a replayed status is `ErrStatusUnchanged`. The old tables
  had neither: `external_transaction_id` defaulted to `""` and was unique nowhere. The manager
  reads `ErrStatusUnchanged` as an acknowledgement rather than a failure, and the repository
  records nothing for a write that changed nothing.
- **Amounts widened.** `amount_cents` is `int64` on the wire and in the schema, where it was
  `int32` — a signed 32-bit count of cents runs out at about twenty-one million dollars, which a
  zero-decimal currency reaches sooner. `billing_interval_months` stopped being optional on the
  wire: zero means one-time, which is what the store stores.

Two things did **not** change, and each is a decision recorded rather than a default taken. The
plan source is platform's `billing/plans` with this application's chooser — `free` stays a plan
rather than an absence, exactly as #1383 decided, because the chooser never declines; what did
move is that the read is of *current* subscriptions, so a lapsed period stops entitling on its
own. And the `belongs_to_account` cascade is re-created on all three account-owned tables,
preserving the erasure behavior the schema had — platform ships no eraser for billing on
retention grounds, and `docs/data-privacy.md` records that switching to one is a policy call this
adoption did not make.

The transaction gap is filed upstream as [platform-go #466], the fifth application of the shape
[#457] landed; `docs/audit.md` names all five and #1419 tracks what deletes here when each ships.

[platform-go #466]: https://github.com/primandproper/platform-go/issues/466

## What adopting a store that *duplicates a declaration* costs

`authorization/database` (#1386) is the first store this repository adopted whose value is
not a table it did not have, and not behaviour it was missing. Both existed. What it deletes
is a *second copy* of something the code already declared.

Permissions were declared twice. `internal/authorization` holds the slices the method table
and the platform policy are written from; `00019_rbac.sql` and `00021_mealplanning.sql` held
~620 lines of `INSERT` statements, and those rows were what authorization actually read at
runtime. Between the two stood one test that string-matched permission *names* against the
concatenated migration text — so it could catch a name declared in Go and never seeded, and
could not see a role→permission mapping that was wrong.

They had drifted, on three of five roles, and the drift was not small:

| Role | Declared in Go | Actually granted |
| --- | --- | --- |
| `account_member` | 133 | 133 |
| `service_data_admin` | 42 | 42 |
| `account_admin` | 43 | **171** |
| `service_admin` | 28 | **240** |
| `service_user` | **133** | **0** |

Two causes, and neither was a typo. The database has had a role hierarchy since #1215;
`PlatformPolicy()` modelled roles as flat, with a comment saying that expressing inheritance
"would be a behavioral change disguised as a refactor" — which had it backwards. And the meal
planning migration granted `service_admin` the data-admin set directly, row for row, which
nothing in Go recorded. In the other direction, `PlatformPolicy()` gave `service_user` the
account-member permissions; the database gives that role nothing, which is correct, because
every user holds `service_user` service-wide and account authority is held per account.

None of it was reachable. `ProvideAuthorizationEnforcer` validated the policy through
`static.NewResolver` and discarded the resolver, so the wrong table was never asked anything.
Both tests built on it — including the equivalence proof between the enforcer and the
hand-rolled interceptor — were driving principals that cannot exist. Correcting the policy is
the first commit of the adoption for that reason: seeding the flat one would have been a
two-way authorization break.

Four things followed:

- **The seed became a call.** `Seed` takes the caller's executor, so it runs in the migrator's
  own transaction — none of the transaction-ownership friction `comments` and `waitlists` hit
  ([#457], [platform-go #458]). It also *revokes*: a role's grants are rewritten rather than
  added to, so a permission deleted in Go disappears on the next migration. That had already
  been needed once by hand — #1376 removed `update.uploaded_media` and had to add a
  compensating `DELETE FROM permissions` to an unrelated migration, which this deletes.
- **Seeding lives behind `Migrate`, not at a wiring site.** An unseeded policy grants nothing,
  so a process that forgot would come up refusing every request. Putting it there also means
  the fifteen container harnesses that migrate a template database get a seeded policy without
  each remembering to ask for one.
- **The foreign key survived, unlike the last two adoptions.** `uploads/registry` and
  `issuereports` re-pointed their cascades at `users`; `waitlists` could re-create nothing.
  Here the assignment references the role *by name*, and platform indexes `authz_roles.name`
  uniquely — deliberately, since reusing an archived role's name would re-grant its authority
  — and PostgreSQL accepts a unique index as a key target. Three joins went with the change,
  including the admin login query's, because sqlc's schema is `migration_files` and a
  generated table is not in it.
- **A column nobody read was hiding a privilege escalation.** `user_roles.scope` said whether a
  role was a service role or an account role, and no query ever filtered on it — including
  `ModifyUserPermissions`, which resolved a caller-supplied role name and wrote it into an
  account-scoped assignment. An account admin could name `service_admin` there and grant a
  member the whole 240-permission closure inside the account. The column has no platform
  counterpart and is not re-created; an allow-list on the input replaces it, and is enforced
  where the column was not.

`authorization/cached` is **not** adopted, and the reason is not the one that closed #1385.
A memory cache would be *correct* here — policy is derived, identical in every replica, and
written only by a migration — but resolution is already one statement against five roles, and
a per-process cache makes a policy change visible one replica at a time. That is a fine trade
to make deliberately and a bad one to make silently, so it waits for somewhere shared to put
it. `authorization/http` is not adopted either: `ProvideAPIRouter` is the only router in the
repository and every route on it is unauthenticated, so there is nothing to guard.

One gap went upstream rather than being papered over: `Seed` is not safe to run concurrently
with itself — it clears a role's grants and re-inserts them with no `ON CONFLICT`, and
migrations run at startup on every replica. Filed as [platform-go #463]; the local workaround is
an advisory lock around the seeding transaction, and there is a test that three concurrent
migrations converge.

[platform-go #463]: https://github.com/primandproper/platform-go/issues/463

## Verification

Everything below was run against the released `v13.0.0` tag, not a `replace` directive:

- `make build` — 253 packages
- `make test` — 85 packages, including the Postgres container suites
- `make lint` — 0 issues (containers, queries, Go, shellcheck)
- `make integration_tests` — apiserver and mcpserver suites
- Every generator re-run, with no drift, so CI's `git diff --exit-code` passes
