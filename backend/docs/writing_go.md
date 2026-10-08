# Writing Go

## Testing Conventions

### Subtest Structure

**Always use subtests, even for single test cases.** This provides a consistent structure that makes it easy to add additional test scenarios later.

#### Parameter Naming Convention

- **Main test function**: Use capital `T` for the `*testing.T` parameter
- **Subtest functions**: Use lowercase `t` for the `*testing.T` parameter

```go
func TestMyFunction(T *testing.T) {
    T.Parallel()

    T.Run("standard", func(t *testing.T) {
        t.Parallel()
        // test implementation
    })

    T.Run("with error condition", func(t *testing.T) {
        t.Parallel()
        // error test implementation
    })
}
```

#### Happy Path First

Structure your subtests with the happy path (successful operation) as the first subtest. This pattern makes it easy to add derivative tests that simulate failure conditions:

- Database connection failures
- Queue timeouts
- Invalid input validation
- Permission errors

### Table Tests - Use Sparingly

**Avoid using table tests.** Only use them when there's meaningful code savings or when testing many similar scenarios with different inputs.

### Parallel Execution

**Always use parallel test execution** where possible:

- Add `T.Parallel()` to main test functions
- Add `t.Parallel()` to subtests
- This significantly improves test suite performance

### Test Organization

- **Unit tests**: Place alongside the code they test (`*_test.go` files)
- **Integration tests**: Located in `testing/integration/` directory
- **Test utilities**: Shared helpers in `github.com/primandproper/platform-go/v2/testutils`

Note: The `tests/` directory contains only integration tests, not unit tests.

### Context Usage

**Always use context in tests** that interact with external services:

```go
func TestMyService(T *testing.T) {
    T.Run("standard", func(t *testing.T) {
        ctx := t.Context()
        // use ctx in service calls
    })
}
```

### Test Reliability

- **No skipped tests**: The linter explicitly forbids `t.SkipNow()` calls. Container-backed tests do not need one: `pgtesting.BuildDatabaseContainerForTest` goes through platform's `pgtest.Run`, which skips the test itself unless `RUN_CONTAINER_TESTS=true`. Don't add your own gate on top of it — call the helper and write the test as if the database is always there. `make test` sets the variable, so these run by default; `RUN_CONTAINER_TESTS=false make test` opts out on a machine without a Docker daemon.
- **No global state**: Each test should be independent
- **Clean up resources**: Properly tear down any created test data. Containers are the helper's job, not yours — `BuildDatabaseContainerForTest` registers teardown of both the container and its pool with `t.Cleanup`, so tests never terminate a container by hand.

## General Go Conventions

### Error Handling

This codebase follows strict error handling practices:

- Always check errors (enforced by linter configuration)
- Wrap external package errors for context
- Use meaningful error messages

### Code Organization

#### Directory Structure

```text
internal/
├── authentication/  # Auth manager, tokens, WebAuthn
├── authorization/   # RBAC, permissions
├── build/           # DI (samber/do) and router construction
├── config/          # Configuration structs, env var loading
├── domain/          # Business logic and types
├── repositories/    # Data access layer
└── services/        # Application services
```

The platform framework (database, cache, observability, messaging, etc.) is an external dependency at `github.com/primandproper/platform-go/v2`.

#### Package Dependencies

- Keep dependencies flowing downward in the architecture
- Use dependency injection (samber/do service locator)

### Linting and Code Quality

The project uses `golangci-lint` with custom configuration:

- **Strict error checking**: No ignored errors
- **Import organization**: Use `gci` for import sorting
- **Test-specific rules**: Relaxed linting for test files where appropriate
- **Complexity limits**: Enforced cyclomatic complexity bounds

### Performance Considerations

- **Parallel processing**: Leverage goroutines where appropriate
- **Connection pooling**: Reuse database and HTTP connections
- **Circuit breakers**: Implement fault tolerance for external services
- **Caching**: Strategic use of in-memory and distributed caching

## Naming Conventions

### Interfaces

- Use descriptive suffixes that indicate purpose:
  - `Logger` - for logging abstractions
  - `Manager` or `<Domain>Manager` - for business logic coordinators (e.g., `IdentityManager`, `WehbhookManager`)
  - `Repository` - for data access abstractions (consistently just `Repository` within each domain)
  - `DataManager` - for data layer abstractions (e.g., `IdentityDataManager`)
  - `Handler` - for request handlers

### Structs

- **Configuration structs**: End with `Config` (e.g., `APIServiceConfig`, `DatabaseConfig`)
- **Domain entities**: Use clear business terms (e.g., `User`, `Account`, `Webhook`)

### Constants and Variables

- Use fully descriptive names: `ConfigurationFilePathEnvVarKey`
- Group related constants in `const` blocks
- Use `var` blocks for package-level variables with initialization

### Functions and Methods

- **Constructors**: Use `New` prefix (e.g., `NewService`, `NewGenerator`)
- **Providers**: Many use `Provide` prefix but not universally enforced
- **Converters**: Use `Convert` prefix describing the transformation

## Struct Design Patterns

### Privacy and Safety

Always include an unexported struct field to prevent accidental struct construction and comparison:

```go
type MyStruct struct {
    _ struct{} `json:"-"`
    
    ID   string `json:"id"`
    Name string `json:"name"`
}
```

### Type Definitions

Group related types in type blocks with clear documentation:

```go
type (
    // Level is a simple string alias for dependency injection's sake.
    Level *level
    
    // RequestIDFunc fetches a string ID from a request.
    RequestIDFunc func(*http.Request) string
)
```

## Dependency Injection with samber/do

The project uses [`samber/do`](https://github.com/samber/do) as a service locator for dependency injection. The DI container is built in `internal/build/services/api/grpc/build.go` via `BuildInjector`, which creates a `do.RootScope` and registers all services, repositories, and managers.

## Common Patterns

### Service Initialization

Services typically follow this pattern:

```go
type Service struct {
    _ struct{} `json:"-"`
    
    logger logger.Logger
    tracer tracing.Tracer
    // ... other dependencies
}

func NewService(
    logger logger.Logger,
    tracer tracing.Tracer,
) *Service {
    return &Service{
        logger: logger,
        tracer: tracer,
    }
}
```

### Interface Design

Keep interfaces focused and purposeful:

```go
// Logger represents a simple logging interface we can build wrappers around.
type Logger interface {
    Info(string)
    Debug(string)
    Error(whatWasHappeningWhenErrorOccurred string, err error)
    
    WithName(string) Logger
    WithValues(map[string]any) Logger
    WithSpan(span trace.Span) Logger
}
```

### Request/Response Handling

- Use structured logging with context
- Implement proper HTTP status codes
- Validate input at service boundaries
- Transform domain objects to/from API representations
- Use consistent response wrapper types (`APIResponse[T]`)

## Project Structure Patterns

### Directory Organization

The codebase follows a clean architecture approach. The platform framework (`github.com/primandproper/platform-go/v2`) is an external dependency providing infrastructure (database, cache, observability, messaging, etc.).

```text
├── artifacts/           # Gitignored folder for temporary files and coverage output
├── cmd/                 # All compiled binaries
│   ├── functions/       # Cloud function implementations
│   │   └── async_message_handler/  # Async message processing function
│   ├── playground/      # Gitignored development sandbox for testing library interactions
│   ├── services/        # Main application services
│   │   ├── api/         # Primary API server (HTTP + gRPC)
│   │   └── admin/       # Admin web app
│   ├── tools/           # Repository-specific development tools
│   │   ├── codegen/     # Code generation utilities
│   │   │   ├── configs/     # Configuration struct generation
│   │   │   ├── queries/     # Database query generation
│   │   │   └── valid_env_vars/ # Environment variable validation
│   │   ├── search_index_initializer/ # Search index setup (disabled)
│   │   └── sqlc_struct_checker/      # Database struct validation
│   └── workers/         # Background job processors
│       ├── db_cleaner/  # Database cleanup jobs
│       └── scheduler/   # Every periodic job, plus the outbox relay and the saga worker
├── deploy/              # Deployment configurations
│   ├── dockerfiles/     # Container build definitions
│   ├── environments/    # Environment-specific configs
│   │   ├── dev/         # Development environment
│   │   ├── localdev/    # Local development setup
│   │   └── testing/     # Testing environment
│   └── kustomize/       # Kubernetes customization configs
├── internal/            # Private application code
│   ├── authentication/  # Auth manager, tokens (JWT/PASETO), WebAuthn
│   ├── authorization/   # Authorization and RBAC logic
│   ├── build/           # Dependency injection (samber/do) and router construction
│   ├── config/          # Configuration management
│   ├── domain/          # Business logic, entities, and domain services
│   ├── functions/       # Cloud function implementations
│   ├── grpc/            # gRPC converters and generated code
│   ├── repositories/    # Data access implementations (Postgres via sqlc)
│   └── services/        # Application services and handlers (gRPC + HTTP)
├── pkg/                 # Public API packages
│   └── client/          # Public API client library for external consumers
└── testing/             # Integration and load tests
    ├── integration/
    │   └── apiserver/   # API server integration tests
    └── load/            # k6 load tests
```

Dependencies are not vendored — `go.mod` and `go.sum` are the committed source of truth,
and Go resolves packages from the module cache.

### Platform as External Framework

The platform framework lives in a separate repo at `github.com/primandproper/platform-go/v15` (with `primitives-go/v2` underneath it). It provides domain-agnostic infrastructure:

- **Infrastructure abstractions**: Database clients, message queues, caching
- **Utilities**: ID generation, encoding, compression, cryptography
- **Observability**: Logging, tracing, metrics collection
- **Search capabilities**: Text indexing, vector search (future-ready)
- **File handling**: Upload management and processing

Nothing meal-planning-specific lives in the platform package.

### Package Import Rules

- **Platform**: External dependency with no business-logic or domain-specific code
- **Domain packages**: Can freely import platform utilities
- **Services**: Can import both domain and platform packages
- **Repositories**: Implement domain interfaces using platform infrastructure
- **cmd/ binaries**: Import whatever they need to compile and run

### Generated Code

Several directories contain generated code:

- `internal/grpc/generated/` - gRPC service definitions
- Database query files (via sqlc)
- Configuration structs (via custom codegen)
- Code generation tools live in `cmd/tools/codegen/`

Never edit generated files directly; modify the generators in `cmd/tools/` instead.

## Should This Be Generic?

Most of this repository's data layer is the same shape over and over — `Exists / Get / GetMany / Create / Update / Archive` over an owner scope, plus a scoped read or two — so the reflex is to ask whether a domain could *declare* that shape instead of writing it out. **The question has been asked, built, vetted and answered: no.** Do not re-file it.

[#1304] spiked a generic scoped-CRUD resource kit, piloted on `comments`, and got a working one: zero domain-specific escape hatches, the whole comments data layer as a `Definition` plus a column list. It was promoted upstream as [platform-go#292], vetted against this repo as its first intended consumer, and **closed on the vetting**. `resources`, `Definition`, `Lookup`, `Match` and `Gate` were dropped and are not coming back.

Why, in the order the reasons matter:

- **The queries were never the duplication.** `cmd/tools/codegen/queries/` already builds them from `querygen`, one file per table. The handful of statements a domain needs beyond standard CRUD are assembled from the same `querygen` fragments a runtime kit would have called at execution time — `Generator.FilterConditions`, `Generator.CursorLimitClause`, the two count selects — so the kit does not remove a second rendering of the filter semantics, it adds a *third* declaration of the column list with a weaker query language than the one already in use.
- **The generated layer is nobody's to maintain.** For the pilot domain a human maintained 409 lines of repository and 141 of codegen; the ~670 lines of sqlc output and `.sql` underneath cost nothing to keep. Counting total lines overstates the win by more than a factor of two.
- **A declared query language re-implements what sqlc gives free.** The kit needed an `ErrUndeclaredLookup` guardrail to stop a generic list from answering predicate combinations nobody chose to index. sqlc prevents that by construction: the query is in a file or it is not.
- **The bugs clustered in the derived machinery, on the two boundaries this codebase can least afford them.** The vetting found a lookup on the scope column binding the caller's value where the tenancy gate's went — one tenant reading another's rows by naming them; a cascade that archived without a limit but reported a single page, so rows past the boundary got no audit entry and no event and nothing said so; and a list vouching for counts an empty page never scanned. None of the three can exist in a hand-written sqlc store, because there is no generic layer there to get them wrong.

There is a coverage ceiling underneath all of that. Of the 66 generated query files here, only 22 are free of `JOIN` / CTE / `GROUP BY` / `UNION`, and `mealplanning` — the bulk of the application — is 32 of 41 with joins. The kit's matcher was equality-only with no partial escape, so the first query shape it could not express dropped that whole resource back to a hand-written store anyway.

### What the vetting did find

The repetition worth removing is not the SQL. It is the ceremony around each write — `WithTransaction`, then record the audit entry, then emit the event — repeated per method per domain, and 11 of the pilot domain's 409 repository lines. That became [#1392] and then the recording spine: a local helper, not a framework.

### Read three instances instead of building a kit

Platform ships `comments`, `issuereports` and `waitlists` — the same scoped-CRUD shape written out three times, tested against three dialects, with nobody owning an abstraction over them. Adopting them ([#1375], [#1377], [#1378]) got the benefit the kit was reaching for, and every non-mealplanning store here is now ~100 lines of wiring over a platform store.

**So, before writing a new store: check whether platform already ships the domain.** If it does, adopt it. If it does not, hand-roll it against sqlc the way `internal/repositories/postgres/mealplanning` does — that is the intended cost, and it is cheaper than the alternative that was tried. No new domains are planned for this service; the record is kept so the question is not asked again.

[#1304]: https://github.com/primandproper/dinnerdonebetter/issues/1304
[#1375]: https://github.com/primandproper/dinnerdonebetter/issues/1375
[#1377]: https://github.com/primandproper/dinnerdonebetter/issues/1377
[#1378]: https://github.com/primandproper/dinnerdonebetter/issues/1378
[#1392]: https://github.com/primandproper/dinnerdonebetter/issues/1392
[platform-go#292]: https://github.com/primandproper/platform-go/pull/292

## Development Workflow

### Before Submitting Code

1. **Run the full test suite**: `make test`
2. **Check linting**: `make lint`
3. **Run integration tests**: `make integration_tests`

### Code Generation

This project uses extensive code generation via tools in `cmd/tools/`:

- **Database queries**: `make querier` (using sqlc)
- **Configuration structs**: `make configs`

Always regenerate code after schema changes.

### Binary Compilation

Everything in `cmd/` compiles to a binary:

- `cmd/ddb/` - The one deployed binary. The API server, MCP server, workers, jobs, and migrations
  are all subcommands of it, so a release ships one image rather than one per workload. See
  `cmd/ddb/README.md`.
- `cmd/tools/` - Development-time tools (codegen, bootstrap, importers), never deployed.

## Getting Started

### For New Contributors

1. **Read the existing tests** in your area of focus to understand patterns
2. **Follow the subtest structure** even for simple tests  
3. **Use the linter** to catch common issues early
4. **Study `internal/build/`** to understand dependency injection setup
5. **Look at similar implementations** before writing new code
6. **Run the full test suite** before submitting changes

### When to Ask Questions

- If you see patterns that seem inconsistent
- When choosing between multiple valid approaches
- Before making architectural changes
- When adding new dependencies or technologies

## Conclusion

These conventions represent years of learned experience maintaining a Go codebase. They prioritize:

- **Consistency** over individual preference
- **Testability** over brevity
- **Maintainability** over cleverness
- **Explicitness** over implicit behavior

When in doubt, follow the existing patterns you see in similar code. The goal is to write code that the next developer (including future you) can easily understand, test, and modify.

Remember: Code is read far more often than it's written. These conventions optimize for the reading experience.
