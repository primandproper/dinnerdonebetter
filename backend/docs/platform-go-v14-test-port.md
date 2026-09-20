# platform-go v14: test port findings

A `replace`-directive port of this repo onto `platform-go` at `main` (untagged
v14) plus `primitives-go/v2` v2.2.0, done on the branch `v14-test-port` to
answer one question: **is there anything in v14 that would force a v15?**

Non-test code builds. Test code does not yet — 58 packages, all the same two
mechanical shapes — and mocks want regenerating. Nothing here is a bump: both
modules are wired with `replace`.

## What v14 is

Two modules where there was one.

- **`primitives-go/v2`** — the substrate: `database`, `filtering`, `errors`,
  `observability`, `encoding`, `uploads`, `search`, `routing`, `server`,
  `tenancy`, `pointer`, `identifiers`, `messagequeue`, `capitalism`. 131 of the
  200 platform packages this repo imported.
- **`platform-go/v14`** — the domains and their transports: `identity`,
  `comments`, `settings`, `waitlists`, `issuereports`, `billing`, `webhooks`,
  `audit`, `notifications`, `dataprivacy`, `metering`, `entitlements`,
  `mediaregistry`, `rbac`, `sessions`, `saga`, `outbox`, `workqueue`. 69.

Every one of the 200 paths this repo used maps to exactly one of them, with 11
renames where a package moved tier: `uploads/registry` → `mediaregistry`,
`search/sync` → `searchsync`, `authorization/database` → `rbac`,
`authentication/oauth2server/database` → `authentication/oauth2serverstore`,
`authentication/webauthn/database` → `authentication/webauthnsessions`.

## The two conventions that changed every call site

**A store holds no database handle.** Reads take a
`database.SQLQueryExecutor`, writes take a `database.Tx`, and the caller
supplies both. Measured across all 34 exported `Store`/`Reader`/`Dispatcher`
interfaces: 18 are fully converted, 6 are split, 3 take none. The split ones
(`dataprivacy`, `metering`, `webhooks`, `operations`, `saga`,
`notifications.Registry`) take no executor on exactly their worker-path methods
— `Claim`, `Reap`, `MarkFlushed`, `Stranded`, `Release`, `Requeue` — and the
rule is stated per method (`webhooks/doc.go:40`, `metering/doc.go:118`). The
three that take none at all are `links.Store` (cache-backed), `shredding.Store`
(may not be on the caller's database at all) and `dataprivacy.Service` (not a
store). **This is coherent and documented. It is not v15 pressure.**

**`tenancy.Scope` is a value type.** It was a string; it is now a struct that
cannot be built from nothing, with `Global()` as the explicit way to say "no
tenant". A read that lost its scope fails to compile rather than quietly
widening. This is the single largest source of mechanical churn in the port and
is worth every line of it.

## What the port bought

These are not speculative: each is a workaround this repo was carrying that came
out with the port.

- **platform-go #458** (waitlists erasure). The local eraser documented two
  departures from the `Eraser` contract — `Withdraw` owned its own transaction,
  and archived signups were unreachable. v14's `Withdraw` takes the caller's
  `Tx` and `ListSignupsForSubject` pages archived rows. Both caveats deleted.
- **platform-go #456** (oauth2 sweep interval). `SweepInterval` is a
  `*time.Duration`, so an explicit zero now means "no sweeper" instead of being
  rewritten to the ten-minute default. The comment explaining that the config
  did not take effect is gone.
- **Audit entries now commit with the writes they describe.** Seven repository
  decorators (`settings`, `comments`, `waitlists`, `payments`, `issuereports`,
  `uploadedmedia`, password-reset tokens) each carried a `record` helper that
  opened a transaction of its own, because the platform store's write had
  already committed. Every one of those comments said the same thing: the pair
  is not atomic, and the gap loses the entry rather than the write. v14 closes
  it — the write, the audit entry and the outbox event are one transaction.
- **Read-backs deleted.** Writes return the row they wrote, so
  `waitlists.recordedTransition`, `comments.ArchiveComment` and
  `mediaregistry.ArchiveObject` no longer need a read before or after the write
  to name what changed, along with the window each read opened.
- **`EnsurePackaging` and the fulfiller's `encrypted bool`** are gone, replaced
  by `WithCompressor`/`WithEncryptor` on the shared option set. Whether
  artifacts are encrypted is now read off the encryptor rather than off a flag
  beside it that could disagree with it.
- **Fenced work-queue completion.** `workqueue.Complete` takes the claimed
  `Item` rather than its key, so a completion cannot land against a lease
  another replica now holds. Handing back the key no longer compiles.

## Findings

### 1. Three constructors kept their path and name and lost their SQL backend

The one thing worth fixing before the tag.

| path (unchanged) | v13 | primitives v2 | SQL version now at |
|---|---|---|---|
| `authentication/oauth2server/config` | `NewStore(ctx, cfg, db, opts...)` | `NewStore(ctx, cfg, opts...)` → **memory** | `platform/authentication/oauth2serverstore/config` |
| `authentication/webauthn/config` | `NewSessionStore(ctx, cfg, db, opts...)` | `NewSessionStore(ctx, cfg, opts...)` → **cache** | `platform/authentication/webauthnsessions/config` |
| `authorization/config` | `NewPolicyResolver(ctx, cfg, db, cache, opts...)` | `NewPolicyResolver(ctx, cfg, cache, opts...)` → **static** | `platform/rbac/config` |

A full-signature diff of every package present at the same path in v13 and in
primitives v2 finds exactly three exported constructors that lost a parameter,
and all three lost the database. The same three `Config` types lost their
`Provider string \`env:"PROVIDER"\`` field, so the environment variable that
selects the backend is read by nobody at those paths — and caarlos0/env does not
error on a variable no field claims.

In all three the *call* fails to compile, because the dropped parameter's type
does not match the one that took its place. The hazard is the fix: deleting the
argument compiles, and silently switches the backend. For `oauth2server` that
means OAuth2 codes and tokens in process memory — fine on one replica in dev,
and every login failing across two. For `authorization/config` it means
authorization decisions come from `cfg.Roles` rather than from the policy table,
and PR #1418 already found those two drifting apart on three of five roles, so
the direction of the change is not predictable.

This port hit the oauth2 one and was saved by the argument count. That is luck.

**Fix:** rename the primitives-tier constructors to say what they build
(`NewMemoryStore`, `NewStaticPolicyResolver`), or keep `Provider` on the
primitives `Config` and refuse `ProviderDatabase` with an error naming the
platform package. Either makes the substitution unrepresentable. Cheap now,
invisible later.

### 2. `audit.NewReader` takes a `database.Client`; `audit.NewRecorder` takes a `dialect.Dialect`

`audit/reader.go:354` and `audit/recorder.go:114`. Both are stateless now —
every method takes the caller's executor — and the Reader uses its `Client`
argument for nothing but `client.Dialect()`. Same package, same release, two
answers to the same question, and the `Client` argument implies a handle the
Reader does not keep. Changing it after the tag is a rewrite at every
construction site. `NewReader(d dialect.Dialect, opts ...ReaderOption)`.

### 3. `WithServerOptions` exists at both config tiers with different element types

`oauth2serverstore/config.WithServerOptions` takes
`oauth2servercfg.Option`, which is the *other* `WithServerOptions`'s own option
type, so a wiring site that wants a server option writes
`WithServerOptions(WithServerOptions(x))` against two identically-named
functions. It reads as a mistake and is not one. Cosmetic, fixable in a patch
release — noted, not blocking.

### Checked and cleared

- The executor split (see above) — principled and documented per method.
- `oauth2serverstore.Store` keeping its client — it backs an external library's
  interface it does not control.
- `billing/privacy.AccountResolver` being a defined type where six siblings
  alias `dataprivacy.ScopeResolver` — a billing row carries both a scope and an
  account id and neither is inferable from the other.
- `dataprivacy.Collector.Collect` taking no executor, so every collector must
  capture a reader at construction — stated as a deliberate ruling in
  `settings/privacy/privacy.go:75` and applied consistently, including by
  platform's own collectors.

## What this means for the "mostly platform imports" goal

v14 ships 11 `.proto` files and complete gRPC servers — 134 RPCs across
`identity`, `billing`, `comments`, `settings`, `waitlists`, `issuereports`,
`webhooks`, `audit`, `notifications`, `signin`, `oauth2clients`.

Comparing RPC sets against this repo's own protos, **they are the same
operations under different names**: `GetSettingDefinitions`/`ListDefinitions`,
`GetWaitlists`/`ListLists`, `CreateWaitlist`/`CreateList`,
`RegisterDeviceToken`/`RegisterDevice`. This repo says `VerbDomainNoun`;
platform says `VerbNoun` with the domain implied by the service. Nothing local
is missing upstream except a short list of genuine extensions — avatars and
service-role admin on identity, trigger configs and the event catalog on
webhooks, `WaitlistIsOpen`, and audit's by-account/by-user read variants that
platform expresses as one `ListEntries` with a `Query`.

**The audit and outbox decorators are not in the way.** Every platform server
takes the `Store` *interface*, not the concrete `SQLStore`, and opens the
transaction itself before calling into it. This repo's decorator repositories
already implement those interfaces, so handing one to a platform server puts the
audit entry and the data-change event inside that server's transaction — which
is exactly what the decorators want and could not have before v14.

So the pass-2 shape is: **keep the decorator repositories, delete the local
proto, converters and gRPC services, mount platform's servers.** The domains
platform also ships account for ~85,700 lines here against mealplanning's
~158,300. Not all of it goes — identity and webhooks carry real local extension
— but the thin ones (comments, settings, waitlists, issuereports, uploadedmedia,
audit, notifications) are mostly forwarding and translation.

The cost is client-facing: adopting platform's protos renames RPCs and messages
across the generated TypeScript and Swift clients. Nothing is deployed, so that
is a regenerate and a frontend sweep rather than a compatibility problem.

## Verdict

Nothing found here forces a v15. The one thing worth fixing before the tag is
finding 1, and it is a rename.

---

# Pass 2 spike: comments

Done on the same branch. The standalone comments service, its proto and its
generated code are gone; platform's `comments/grpc.Server` is mounted over this
repo's existing repository. The backend builds and the comments suite passes.

## What the spike had to prove

The whole pass-2 plan rests on one claim: that this repo's audit entry and
outbox event survive adoption, because platform's server takes the
`comments.Store` interface and opens the transaction before calling into it. If
that were wrong — if the server committed the row and left the decorator's two
statements outside — then adopting any of the thirteen surfaces would mean
trading the audit log for the deletion, which is not a trade this application
can make.

`internal/repositories/postgres/comments/server_transaction_test.go` settles it
against a real Postgres, driving platform's server with this repo's repository
behind it and a real outbox writer:

- **It commits together.** One `CreateComment` through platform's surface leaves
  the comment, its audit entry and its outbox message behind.
- **It rolls back together.** With the audit repository made to fail, the call
  errors and *neither the comment nor the outbox row exists*.
- **A refused write records nothing.** An unknown target type is refused by the
  store, and no entry and no event are left behind.

The middle test is the load-bearing one, so it was checked against a mutant
rather than trusted for being green. Restoring the pre-v14 shape — the store
opening a transaction of its own, which is exactly what v13 did — makes it fail
on precisely the right assertion:

    Error:    Should be zero, but was 1
    Messages: the comment should have rolled back with the audit entry

So the test is not passing vacuously: it distinguishes v14's shape from v13's,
which is the only thing it was written to do.

## What came out, and what went in

| | lines |
|---|---|
| deleted: local service, converters, permissions, generated stubs, `proto/comments` | −2,611 |
| deleted: the four `AddCommentTo…` RPCs and their messages from the mealplanning and issue_reports protos | −722 |
| added: `sessions.Principal` + extractor (one-time, serves all thirteen) | +57 |
| added: `internal/build/comments` server wiring | +55 |
| added: the transaction tests | +296 |

Net −2,925 for the domain, and the only new production code that is not
comments-specific is the 57-line principal adapter.

## Four things the spike found that reading could not

**The convenience RPCs go too.** `AddCommentToRecipe`, `AddCommentToMeal`,
`AddCommentToMealPlan` and `AddCommentToIssueReport` lived on two *other*
services and looked like separate surface. Each is sugar: it names the target
from the RPC rather than the body and forwards to `CreateComment`. The existence
check they appear to add is the store's, through the target catalog. Platform's
`CreateComment` takes the target as a field, so all four are redundant — but
they mean comments cannot be adopted without regenerating two neighbouring
protos, which is the first thing that makes this a cross-domain change.

**`Permission` had to become an alias.** Platform's surfaces ship their own
method→permission maps, and `commentsgrpc.Permissions()` returns
`map[string][]primitives.Permission`. This repo's `Permission` was a defined
type, so the maps would not compose — one conversion per domain, thirteen times.
Platform documents the fix in `authorization/authorization.go:18`: declare the
local type as an alias. One line, and every existing constant, map key and
switch kept compiling.

**The permission strings change.** Local `"create.comments"` becomes platform's
`"comments.create"` — domain first, so that composed domains cannot collide on a
bare verb. `internal/authorization/comments_permissions.go` now re-exports
platform's constants under the names the policy already spells, so only the
seeded strings move. Free here because nothing is deployed; a deployed service
would need a policy migration per adopted domain.

**Two of the seams need nothing.** Platform's default `AuthorAuthorizer`,
`OwnCommentsOnly`, is exactly what the deleted service's `ownedComment` did, and
the absent `GrantsExtractor` clears `include_archived`, which is the behaviour
the deleted service had (it exposed no archived read at all). Both are left at
their defaults, with the reasons written down where the server is built.

## Two findings fixed upstream mid-spike

`audit.NewReader` now takes a `dialect.Dialect` rather than a `database.Client`,
and `oauth2serverstore/config.WithServerOptions` is now
`WithServerConfigOptions`. Both call sites here have been moved over. That
clears findings 2 and 3; finding 1 — the three constructors that silently lose
their SQL backend — is the one still open, and it is still the only thing worth
blocking the tag on.

## What this says about the remaining twelve

The shape holds and the cost is now measured rather than estimated. Per domain,
expect: delete the service, converters, proto and generated stubs; add a
build-layer registration of about fifty lines; re-export platform's permission
constants; and regenerate any neighbouring proto that imported the domain's
messages. The repository decorator — the thing that makes the write auditable —
is untouched, which is the whole reason this is worth doing.

What will not be uniform is the cross-domain proto coupling. Comments was
imported by two neighbours; identity and webhooks will be worse, and the
generated TypeScript and Swift clients have to be regenerated and swept either
way.
