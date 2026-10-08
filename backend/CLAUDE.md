# Backend

Go backend for Dinner Done Better. See `docs/` for detailed domain documentation.

## Build & Run Commands

```bash
make dev                # Local dev environment (docker-compose: Postgres, Redis, Jaeger, etc.)
make build              # Build all binaries
make test               # Run unit tests
make lint               # Run all linters (containers, queries, Go, shellcheck)
make format             # Format all code (Go + Terraform)
make integration_tests  # Run integration tests against Postgres
```

## Code Generation

```bash
make querier    # Regenerate SQL query code (codegen + sqlc via Docker — NOT `sqlc generate`)
make configs    # Regenerate config structs per environment
make env_vars   # Regenerate .env.example from the config structs
make proto      # Generate proto (run from repo root, not backend/)
```

Never edit files in `*/generated/` directories. Modify the generators in `cmd/tools/codegen/` instead.

## Architecture

The platform framework (database, cache, observability, messaging, uploads, search, encoding, etc.) is an external dependency at `github.com/primandproper/platform-go/v15`. Application code lives under `internal/`:

| Layer              | Path                       | Role                                                      |
|--------------------|----------------------------|-----------------------------------------------------------|
| **Domain**         | `internal/domain/`         | Business logic, types, managers, fakes, converters, mocks |
| **Repositories**   | `internal/repositories/`   | Data access (Postgres via sqlc)                           |
| **Services**       | `internal/services/`       | gRPC and HTTP handlers                                    |
| **Authentication** | `internal/authentication/` | Auth manager, tokens (JWT/PASETO), WebAuthn               |
| **Authorization**  | `internal/authorization/`  | RBAC, permissions                                         |
| **Build**          | `internal/build/`          | Dependency injection (samber/do) and router construction  |
| **Config**         | `internal/config/`         | Configuration structs, env var loading                    |

**Import rules**: Platform is an external library with no business logic. Domain imports platform freely. Services import domain + platform. Repositories implement domain interfaces using platform infra.

### Key Directories

```bash
cmd/ddb/                 # The one deployed binary; every workload is a subcommand (see cmd/ddb/README.md)
cmd/tools/codegen/       # Code generation: queries, configs, env vars
cmd/tools/               # Development-time tools, never deployed
internal/mcpserver/      # MCP server, run by `ddb serve mcp`
pkg/client/              # Public Go API client library
testing/integration/     # Integration tests
deploy/                  # Dockerfile, Kustomize, environment configs
```

Every deployed workload ships in one image and is selected by subcommand: `ddb serve`,
`ddb serve mcp`, `ddb worker scheduler`, `ddb worker async-messages`, `ddb job db-cleaner`,
`ddb job email-deliverability`, `ddb migrate`. Adding a workload means adding a subcommand and a
manifest that names it — not a new `cmd/` directory, Dockerfile, and skaffold artifact.

## Code Conventions

See `docs/writing_go.md` for full details.

- **Testing**: Always use subtests. Main test func uses `T *testing.T`, subtests use `t *testing.T`. Happy path first. Always `T.Parallel()` / `t.Parallel()`. Avoid table tests. No `t.SkipNow()`.
- **Structs**: Include `_ struct{} \`json:"-"\`` as first field to prevent accidental construction/comparison.
- **Naming**: Constructors use `New` prefix. Interfaces: `Repository`, `Manager`/`XxxDataManager`, `Handler`. Config structs end with `Config`.
- **DI**: `samber/do` service locator. Container built in `internal/build/services/api/grpc/build.go`.
- **Observability**: Every repo/service/manager defines `o11yName` constant used for `tracing.NewNamedTracer` and `logging.NewNamedLogger`.
- **Errors**: Always check errors (enforced by linter). Wrap external errors for context.
- **Context**: Always use `ctx := t.Context()` in tests that interact with services.

## One Domain, Platform Underneath

Mealplanning is the only domain this service has, and no new ones are planned: the work is to
simplify it, not extend it. Every other store is platform-go's with ~100 lines of wiring (`comments`,
`issuereports`, `waitlists`, `settings`, `billing`, `webhooks`, and the rest). Before writing a new
store, check whether platform already ships the domain; if it does not, hand-roll it against sqlc
the way `internal/repositories/postgres/mealplanning` does. There is no CRUD kit, and the reasons
are recorded in `docs/writing_go.md` ("Should This Be Generic?") so the question is not re-filed.

Every site outside the three mealplanning roots that names the domain carries a
`// Domain: mealplanning` marker, and `TestDomainMarkerCensus` holds the composition root to it;
the domain registers through `internal/domain/mealplanning/registration`.

## Configuration

Two-stage config: JSON file (path from `CONFIGURATION_FILEPATH` env var) + environment variable overrides. Env vars use `DINNER_DONE_BETTER_` prefix. See `docs/configuration.md`.

## Documentation

- `docs/audit.md` — the tamper-evident audit log: how to record an entry, what the chain guarantees, retention, redaction
- `docs/search-pagination.md` — how clients page search results, and what the cursor means on each path
- `docs/writing_go.md` — Go conventions, testing, naming, patterns, and why there is no CRUD kit
- `docs/migrations.md` — Database migration workflow
- `docs/history/` — Adoption logs for the v13 and v14 ports: records of what each changed, not living docs
- `docs/configuration.md` — Config loading, env vars, deployment
- `docs/payments.md` — Payments domain architecture
- `docs/metering.md` — Usage metering: what is counted, why nothing is limited or billed yet
- `docs/entitlements.md` — The plan catalog: what each tier includes, and how metering is told
- `docs/mcp-usage-guide.md` — MCP server setup and auth flow
