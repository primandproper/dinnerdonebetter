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
make env_vars   # Regenerate valid environment variable constants
make proto      # Generate proto (run from repo root, not backend/)
```

Never edit files in `*/generated/` directories. Modify the generators in `cmd/tools/codegen/` instead.

## Architecture

The platform framework (database, cache, observability, messaging, uploads, search, encoding, etc.) is an external dependency at `github.com/primandproper/platform-go/v2`. Application code lives under `internal/`:

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
cmd/services/api/        # Primary API server (HTTP + gRPC)
cmd/services/admin/      # Admin web app
cmd/workers/             # Background job processors (meal plan finalizer, grocery lists, etc.)
cmd/tools/codegen/       # Code generation: queries, configs, env vars
cmd/functions/           # Cloud functions (async message handler)
pkg/client/              # Public Go API client library
testing/integration/     # Integration tests
deploy/                  # Dockerfiles, Kustomize, environment configs
```

## Code Conventions

See `docs/writing_go.md` for full details.

- **Testing**: Always use subtests. Main test func uses `T *testing.T`, subtests use `t *testing.T`. Happy path first. Always `T.Parallel()` / `t.Parallel()`. Avoid table tests. No `t.SkipNow()`.
- **Structs**: Include `_ struct{} \`json:"-"\`` as first field to prevent accidental construction/comparison.
- **Naming**: Constructors use `New` prefix. Interfaces: `Repository`, `Manager`/`XxxDataManager`, `Handler`. Config structs end with `Config`.
- **DI**: `samber/do` service locator. Container built in `internal/build/services/api/grpc/build.go`.
- **Observability**: Every repo/service/manager defines `o11yName` constant used for `tracing.NewNamedTracer` and `logging.NewNamedLogger`.
- **Errors**: Always check errors (enforced by linter). Wrap external errors for context.
- **Context**: Always use `ctx := t.Context()` in tests that interact with services.

## Adding a New Domain

See `docs/adding_a_new_domain.md` for the full checklist (note: doc still references Wire but DI is now `samber/do`). Key steps:

1. Migration (`migrations/migration_files/NNNNN_name.sql`, register in `migrate.go`)
2. Query codegen (`cmd/tools/codegen/queries/`, register in `main.go`, run `make querier`)
3. Domain types (`internal/domain/<domain>/` — structs, fakes, converters, mocks)
4. Repository (`internal/repositories/postgres/<domain>/`)
5. Manager (`internal/domain/<domain>/manager/`)
6. gRPC proto (`proto/<domain>/`, run `make proto` from repo root)
7. gRPC service (`internal/services/<domain>/grpc/`)
8. Permissions (`internal/authorization/`, aggregate in `extras.go`)
9. DI registration (`internal/build/services/api/grpc/build.go` + `extras.go`)

## Configuration

Two-stage config: JSON file (path from `CONFIGURATION_FILEPATH` env var) + environment variable overrides. Env vars use `DINNER_DONE_BETTER_` prefix. See `docs/configuration.md`.

## Documentation

- `docs/writing_go.md` — Go conventions, testing, naming, patterns
- `docs/adding_a_new_domain.md` — Step-by-step domain creation checklist
- `docs/migrations.md` — Database migration workflow
- `docs/configuration.md` — Config loading, env vars, deployment
- `docs/payments.md` — Payments domain architecture
- `docs/mcp-usage-guide.md` — MCP server setup and auth flow
