# Adding a New Domain

This document enumerates every step required to add a new domain (or new entity within an existing domain) to the system. It serves as the authoritative checklist for humans and AI agents, with file paths, patterns, and decision points made explicit.

**Reference checklist**: [`.github/pull_request_template.md`](.github/pull_request_template.md)

## When to Use This Doc

- **New business domain**: Adding a completely new area (e.g., invoicing, inventory) — follow the full path.
- **New entity within domain**: Adding a new table/entity to an existing domain (e.g., new `valid_*` in mealplanning) — follow the shorter path in [Section 11](#11-new-entity-within-existing-domain).

## Should This Be Generic?

Most of what follows is the same shape every time — `Exists / Get / GetMany / Create / Update / Archive` over an owner scope, plus a scoped read or two — so the reflex is to ask whether a domain could *declare* that shape instead of walking this checklist. **The question has been asked, built, vetted and answered: no.** Do not re-file it.

[#1304] spiked a generic scoped-CRUD resource kit, piloted on `comments`, and got a working one: zero domain-specific escape hatches, the whole comments data layer as a `Definition` plus a column list. It was promoted upstream as [platform-go#292], vetted against this repo as its first intended consumer, and **closed on the vetting**. `resources`, `Definition`, `Lookup`, `Match` and `Gate` were dropped and are not coming back.

Why, in the order the reasons matter:

- **The queries were never the duplication.** `cmd/tools/codegen/queries/` already builds them from `querygen`, one file per table. The handful of statements a domain needs beyond standard CRUD are assembled from the same `querygen` fragments a runtime kit would have called at execution time — `Generator.FilterConditions`, `Generator.CursorLimitClause`, the two count selects — so the kit does not remove a second rendering of the filter semantics, it adds a *third* declaration of the column list with a weaker query language than the one already in use.
- **The generated layer is nobody's to maintain.** For the pilot domain a human maintained 409 lines of repository and 141 of codegen; the ~670 lines of sqlc output and `.sql` underneath cost nothing to keep. Counting total lines overstates the win by more than a factor of two.
- **A declared query language re-implements what sqlc gives free.** The kit needed an `ErrUndeclaredLookup` guardrail to stop a generic list from answering predicate combinations nobody chose to index. sqlc prevents that by construction: the query is in a file or it is not.
- **The bugs clustered in the derived machinery, on the two boundaries this codebase can least afford them.** The vetting found a lookup on the scope column binding the caller's value where the tenancy gate's went — one tenant reading another's rows by naming them; a cascade that archived without a limit but reported a single page, so rows past the boundary got no audit entry and no event and nothing said so; and a list vouching for counts an empty page never scanned. None of the three can exist in a hand-written sqlc store, because there is no generic layer there to get them wrong.

There is a coverage ceiling underneath all of that. Of the 66 generated query files here, only 22 are free of `JOIN` / CTE / `GROUP BY` / `UNION`, and `mealplanning` — the bulk of the application — is 32 of 41 with joins. The kit's matcher was equality-only with no partial escape, so the first query shape it could not express dropped that whole resource back to a hand-written store anyway.

### What the vetting did find

The repetition worth removing is not the SQL. It is the ceremony around each write — `WithTransaction`, then record the audit entry, then emit the event — repeated per method per domain, and 11 of the pilot domain's 409 repository lines. That is [#1392]: a local helper, not a framework.

### Read three instances instead of building a kit

Platform ships `comments`, `issuereports` and `waitlists` — the same scoped-CRUD shape written out three times, tested against three dialects, with nobody owning an abstraction over them. Adopting them is section 2 of [#1368] and gets the benefit the kit was reaching for. `comments` ([#1375]), `issuereports` ([#1377]) and `waitlists` ([#1378]) are all done.

**So, before following this checklist: check whether platform already ships the domain.** If it does, adopt it. If it does not, hand-roll it through the steps below — that is the intended cost, and it is cheaper than the alternative that was tried.

[#1304]: https://github.com/primandproper/dinnerdonebetter/issues/1304
[#1368]: https://github.com/primandproper/dinnerdonebetter/issues/1368
[#1375]: https://github.com/primandproper/dinnerdonebetter/issues/1375
[#1377]: https://github.com/primandproper/dinnerdonebetter/issues/1377
[#1378]: https://github.com/primandproper/dinnerdonebetter/issues/1378
[#1392]: https://github.com/primandproper/dinnerdonebetter/issues/1392
[platform-go#292]: https://github.com/primandproper/platform-go/pull/292

## High-Level Flow

```mermaid
flowchart TD
    subgraph required [Required Steps]
        M[Migration]
        Q[Queries]
        T[Types]
        R[Repository]
        Mg[Manager]
        G[gRPC]
        P[Permissions]
    end
    
    subgraph optional [Optional Steps]
        C[Configs]
        A[Admin Routes]
        I[Integration Tests]
    end
    
    M --> Q
    Q --> T
    T --> R
    R --> Mg
    Mg --> G
    G --> P
    P --> C
    P --> A
    P --> I
```

---

## 1. Migration

### Location

- `backend/internal/repositories/postgres/migrations/migration_files/`
- Naming: `NNNNN_<domain_or_feature>.sql` (use next sequence number)

### Contents

- `CREATE TYPE` for enums
- `CREATE TABLE` for entities

### Patterns

- **Soft delete**: `archived_at` (nullable timestamp)
- **Timestamps**: `created_at`, `last_updated_at`
- **ID**: text ID stored as `TEXT`
- **Data ownership**: `belongs_to_user` vs. `belongs_to_account` — see [docs/identity.md](identity.md)

### Registration

Migrations are explicitly listed in `backend/internal/repositories/postgres/migrations/migrate.go`. Add a new entry:

```go
{Version: N, Description: "your domain tables", Script: fetchMigration("00019_your_domain")},
```

Migrations run on API server startup.

---

## 2. Queries (Full Workflow)

The query pipeline has two stages:

1. **Go codegen** (`cmd/tools/codegen/queries`) outputs `.sql` files
2. **sqlc** reads those `.sql` files and generates Go code

### 2a. Codegen That Produces SQL

**Location**: `backend/cmd/tools/codegen/queries/`

#### Add a query builder file

Create a file like `your_domain_entity_name.go` (e.g., `settings_service_settings.go`). Each file:

1. Defines `const tableName = "your_table_name"`
2. Defines `var columns = []string{...}` (include `idColumn`, `createdAtColumn`, `lastUpdatedAtColumn`, `archivedAtColumn` from shared constants)
3. Implements `buildXxxQueries(database string) []*Query` returning a slice of `Query` structs

There is nothing to register your table with. `DestroyAllData` — the `TRUNCATE` that
empties a database between tests — reads `pg_tables` when it runs, so a table is covered
by virtue of existing rather than by virtue of somebody having written a query builder
for it.

#### Query types

- `OneType`: Returns single row
- `ManyType`: Returns multiple rows (use with filtered_count, total_count for list endpoints)
- `ExecType`: Execute (e.g., INSERT)
- `ExecRowsType`: Execute and return affected row count (e.g., UPDATE)

#### Common queries

- `CreateXxx`: INSERT
- `GetXxx`: SELECT by ID
- `GetXxxs`: SELECT list with filtering (created_after, created_before, cursor, result_limit)
- `SearchForXxxs`: SELECT list with text search (e.g., `ILIKE` on name)
- `ArchiveXxx`: UPDATE `archived_at = NOW()`
- `CheckXxxExistence`: SELECT EXISTS

**Reference**: [backend/cmd/tools/codegen/queries/settings_service_settings.go](backend/cmd/tools/codegen/queries/settings_service_settings.go)

#### Register in main.go

Add an entry to the `queryOutput` map in `backend/cmd/tools/codegen/queries/main.go`:

```go
"internal/repositories/postgres/<domain>/sqlc_queries/<entity>.sql": build<Entity>Queries(databaseToUse),
```

### 2b. sqlc Configuration

**Location**: `backend/sqlc.yaml`

For a **new domain** (new package under `postgres/`), add a new engine block. Copy structure from an existing block (e.g., settings):

- `engine`: `postgresql`
- `schema`: `internal/repositories/postgres/migrations/migration_files`
- `queries`: path to `internal/repositories/postgres/<domain>/sqlc_queries`
- `gen.go.out`: `internal/repositories/postgres/<domain>/generated`
- Use same `rules`, `gen` options (emit_interface, omit_unused_structs, etc.)

For a **new entity within existing domain**: add a new `.sql` file under the domain's `sqlc_queries/` folder; no sqlc config change needed (same queries path).

### 2c. Commands

```bash
cd backend
make querier   # Runs: queries (codegen) + queries_lint + sqlc generate
```

---

## 3. Types (Domain Layer)

**Location**: `backend/internal/domain/<domain>/`

### Definitions

- Struct with `_ struct{}` for safety (per [writing_go.md](backend/docs/writing_go.md))
- JSON tags on exported fields
- Validation via `ozzo-validation` in `ValidateWithContext(ctx)`
- Input structs: `XxxCreationRequestInput`, `XxxUpdateRequestInput`, `XxxDatabaseCreationInput`

### Fakes

**Location**: `backend/internal/domain/<domain>/fakes/`

- `BuildFakeXxx()` functions returning realistic test data
- Used in integration tests and unit tests

Converters:

- **Domain <-> DB models**: Convert between domain structs and sqlc-generated models
- Place in `backend/internal/domain/<domain>/converters/` or inline in repository
- Pattern: `ConvertSqlcModelToDomain(m *generated.Model) *domain.Entity`

### Mocks

- Use `mockgen` for Repository and Manager interfaces
- `backend/internal/domain/<domain>/mock/repository.go`
- `backend/internal/domain/<domain>/manager/mock/manager.go`

**References**: [backend/internal/domain/settings/](backend/internal/domain/settings/), [backend/internal/domain/webhooks/](backend/internal/domain/webhooks/)

---

## 4. Data Layer: Repository and Manager

Preferred pattern: **Manager wraps Repository** (not Repository-only).

This layer is where the repetition between domains is most visible. It is deliberate — see [Should This Be Generic?](#should-this-be-generic) for the kit that was built to remove it and why it was rejected.

### 4a. Repository

**Location**: `backend/internal/repositories/postgres/<domain>/`

#### client.go (or entity-specific files)

- Struct holds: `generated.Querier`, `logger`, `tracer`, `database.Client`, `readDB`, `writeDB`
- Implement domain `Repository` interface
- Methods: call generated querier, convert sqlc models to domain, return
- **o11yName**: `"<domain>_db_client"` (e.g., `"webhook_db_client"`)

wire.go:

- `ProvideXxxRepository(logger, tracerProvider, ..., client) Repository`

#### Unit tests

- `client_test.go` or `<entity>_test.go` — test repo methods against test DB
- Use `internal/repositories/postgres/testing` helpers

### 4b. Manager

**Location**: `backend/internal/domain/<domain>/manager/`

#### Interface (interface.go)

```go
type XxxDataManager interface {
    CreateXxx(ctx context.Context, ...) (*domain.Xxx, error)
    GetXxx(ctx context.Context, id, accountID string) (*domain.Xxx, error)
    GetXxxs(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[domain.Xxx], error)
    ArchiveXxx(ctx context.Context, id, accountID string) error
    // ...
}
```

#### Implementation

- Wraps Repository
- Adds: validation, multi-step logic, event publishing, ID generation
- **o11yName**: `"<domain>_data_manager"` (e.g., `"webhook_data_manager"`)

wire.go:

- `ProvideXxxManager(...)` or `NewXxxDataManager(...)`

**Reference**: [backend/internal/domain/webhooks/manager/](backend/internal/domain/webhooks/manager/)

### 4c. Wire and Build

Add to `backend/internal/build/services/api/grpc/build.go`:

- `xxxrepo.XxxRepoProviders` (or equivalent)
- `xxxmanager.XxxManagerProviders` (or equivalent)

Each repo needs:

- Entry in `sqlc.yaml` (if new domain)
- `backend/internal/repositories/postgres/<domain>/wire.go` exporting providers

---

## 5. gRPC

### 5a. Protobuf

**Location**: `proto/<domain>/`

Typical split:

- `xxx_service.proto`: service definition and RPCs
- `xxx_messages.proto`: message definitions
- `xxx_service_types.proto`: shared types (if used)

Define:

- `service XxxService { rpc CreateXxx(...) returns (...); ... }`
- Request/response messages
- `option go_package = "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/<domain>";`

**Generate**:

```bash
make proto   # From repo root: format_proto, proto_golang, proto_swift
```

Output: `backend/internal/grpc/generated/services/<domain>/*.pb.go`

### 5b. gRPC Service Implementation

**Location**: `backend/internal/services/<domain>/grpc/`

#### service.go

- Struct implementing generated `XxxServiceServer`
- Depends on Manager (not Repository directly)
- **o11yName**: `"<domain>_service"` (e.g., `"configuration_service"`)
- Method handlers: extract session from context, call Manager, convert domain -> proto, return (or handle gRPC status errors)

Converters:

- Domain types <-> proto types
- in `converters/` subpackage

wire.go:

- `NewService(logger, tracerProvider, xxxManager) XxxServiceServer`

### 5c. Registration

1. Add server to `BuildRegistrationFuncs` in [backend/internal/build/services/api/grpc/extras.go](backend/internal/build/services/api/grpc/extras.go):
   - Add parameter
   - Add `xxxsvc.RegisterXxxServiceServer(server, xxxService)` in registration func
2. Add service to wire.Build in [backend/internal/build/services/api/grpc/build.go](backend/internal/build/services/api/grpc/build.go)

---

## 6. Auth Interceptor and Permissions

### 6a. Permissions

**Location**: `backend/internal/authorization/`

Create or extend `*_permissions.go`:

```go
const (
    CreateXxxPermission Permission = "id_create.xxx"
    ReadXxxPermission   Permission = "id_read.xxx"
    ArchiveXxxPermission Permission = "id_archive.xxx"
    // ...
)
```

Pattern: `id_<action>.<resource>`. Add to any role/action registries if using RBAC (e.g., `permissible_action_ids.go`, `permission_set_checkers.go`).

**Reference**: [backend/internal/authorization/settings_permissions.go](backend/internal/authorization/settings_permissions.go)

### 6b. Method Permissions

**Location**: `backend/internal/services/<domain>/grpc/permissions.go`

```go
func ProvideMethodPermissions() XxxMethodPermissions {
    return XxxMethodPermissions{
        xxxsvc.XxxService_CreateXxx_FullMethodName: {authorization.CreateXxxPermission},
        xxxsvc.XxxService_GetXxx_FullMethodName:   {authorization.ReadXxxPermission},
        // ...
    }
}
```

Use generated `FullMethodName` constants from the gRPC generated code.

### 6c. Aggregate

In [backend/internal/build/services/api/grpc/extras.go](backend/internal/build/services/api/grpc/extras.go):

1. Add `xxxPermissions xxxgrpc.XxxMethodPermissions` parameter to `AggregateMethodPermissions`
2. Add merge loop: `for method, perms := range xxxPermissions { result[method] = perms }`

---

## 7. Observability Keys

Explicit `o11yName` constant in:

- Repository client: `"<domain>_db_client"`
- gRPC service: `"<domain>_service"`
- Manager: `"<domain>_data_manager"`

Used for:

- `tracing.NewNamedTracer(tracerProvider, o11yName)`
- `logging.NewNamedLogger(logger, o11yName)`

---

## 8. Configs (When Needed)

**Rarely needed.** Only add if the domain has runtime configuration (external URLs, feature flags, env-specific behavior).

If adding:

1. Add field to `ServicesConfig` in [backend/internal/config/services_config.go](backend/internal/config/services_config.go)
2. Add to `wire.FieldsOf(new(*ServicesConfig), "Xxx")` in [backend/internal/config/wire.go](backend/internal/config/wire.go)
3. Add values in codegen configs: `internal/config/environments/localdev.go`, `integrationtests.go`. For prod, update `deploy/environments/prod/kustomize/configs/*.json` as needed.

---

## 9. Integration Tests

**Location**: `backend/testing/integration/apiserver/`

**File**: `<domain>_<entity>_test.go` or `<domain>_<service>_test.go`

**Helpers** (from `init.go`):

- `createUserAndClientForTest(t)` — returns user and authenticated gRPC client
- `buildUnauthenticatedGRPCClientForTest(t)` — unauthenticated client

**Test cases**:

- Happy path (create, list, get, update, archive)
- Requires auth (call without token → error)
- Invalid input (validation failure → error)
- Permission denied (non-admin or wrong account → error), if applicable

**Reference**: [backend/testing/integration/apiserver/identity_accounts_test.go](backend/testing/integration/apiserver/identity_accounts_test.go)

---

## 10. Admin Web App Routes (When Applicable)

**Case-by-case.** API-only domains (e.g., internal ops, data privacy) may skip. Add when admins need CRUD (settings, valid_*, waitlists, issuereports).

**Location**: [backend/cmd/services/admin/routes.go](backend/cmd/services/admin/routes.go)

**Pattern**:

- List: `r.Get("/entities", ghttp.Adapt(s.EntitiesList))`
- Edit: `r.Get("/entities/{id}", ghttp.Adapt(s.EntityPage))`
- Create: `r.Get("/entities/new", ...)`, `r.Post("/api/entities", ...)`
- Search API: `r.Get("/api/entities/search", ...)`

**Handlers**: Implement in `cmd/services/admin/` — HTTP handlers that call gRPC or repository.

**Reference**: Settings, valid_ingredients, waitlists in routes.go

---

## 11. New Entity Within Existing Domain

Shorter path — reuse:

- Existing domain package
- Existing Repository package and sqlc block
- Existing protobuf service

**Add**:

1. Migration (new table)
2. Codegen query builder file + entry in codegen `main.go`
3. New `sqlc_queries/<entity>.sql` (generated by codegen)
4. Repository methods
5. Manager methods (if using Manager)
6. New RPCs in existing proto service
7. gRPC handlers
8. Permissions for new RPCs
9. Optionally: admin routes, integration tests

**Do not add**:

- New domain package
- New service package
- New sqlc engine block
- New wire blocks for service/repo (extend existing)

---

## 12. Quick Reference: File Checklist

| PR Checklist Item        | Files / Paths                                                                                                         |
|--------------------------|-----------------------------------------------------------------------------------------------------------------------|
| Migration                | `backend/internal/repositories/postgres/migrations/migration_files/NNNNN_name.sql`; register in `migrate.go`          |
| Queries                  | `cmd/tools/codegen/queries/<domain>_<entity>.go`; `main.go`; `sqlc_queries/<entity>.sql`; `sqlc.yaml` (if new domain) |
| Observability keys       | `o11yName` in repo client, gRPC service, manager                                                                      |
| Types - Definitions      | `backend/internal/domain/<domain>/*.go`                                                                               |
| Types - Fakes            | `backend/internal/domain/<domain>/fakes/`                                                                             |
| Types - Converters       | `backend/internal/domain/<domain>/converters/` or inline                                                              |
| Types - Mocks            | `backend/internal/domain/<domain>/mock/`, `manager/mock/`                                                             |
| Data Manager - Storage   | `backend/internal/repositories/postgres/<domain>/*.go`                                                                |
| Data Manager - Interface | `backend/internal/domain/<domain>/manager/interface.go`                                                               |
| Data Manager - Impl      | `backend/internal/domain/<domain>/manager/*.go`                                                                       |
| Data Manager - Wire      | `manager/wire.go`; `build.go`                                                                                         |
| gRPC - Proto             | `proto/<domain>/*.proto`                                                                                              |
| gRPC - Service           | `backend/internal/services/<domain>/grpc/service.go`                                                                  |
| gRPC - Converters        | `services/<domain>/grpc/converters/` or inline                                                                        |
| gRPC - Registration      | `build/services/api/grpc/extras.go`, `build.go`                                                                       |
| Auth interceptor         | `authorization/*_permissions.go`; `services/<domain>/grpc/permissions.go`; `extras.go` AggregateMethodPermissions     |
| Configs                  | `config/services_config.go`, `wire.go`, `config/environments/*.go` (if needed)                                        |
| Integration tests        | `testing/integration/apiserver/<domain>_<entity>_test.go`                                                             |
| Admin - List view        | `cmd/services/admin/routes.go`; handler in `cmd/services/admin/`                                                      |
| Admin - Edit view        | Same                                                                                                                  |
