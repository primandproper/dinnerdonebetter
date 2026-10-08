# Configuration

Every backend binary is configured through a two-stage process: a JSON config file provides the base values, and environment variables can override any config field. Environment variables always take precedence over the JSON config.

## Overview

1. **JSON config file**: The binary reads the path to its config file from the `CONFIGURATION_FILEPATH` environment variable, then loads and parses that JSON into a config struct.
2. **Environment variable overrides**: [github.com/caarlos0/env/v11](https://github.com/caarlos0/env) is used to overlay config values from environment variables. Any field defined in the config struct can be overridden by setting the corresponding env var.

**Example**: If the JSON config sets `config.Service.GRPCServer.Port` to `1234` and you run the application with `DINNER_DONE_BETTER_GRPC_SERVER_PORT=4321`, the service will use `4321`.

## Loading the config

Binaries load config via `config.LoadConfigFromEnvironment[T]()` or `config.LoadConfigFromPath[T](ctx, path)`:

```go
cfg, err := config.LoadConfigFromEnvironment[config.APIServiceConfig]()
```

Both functions:

1. Read and decode the JSON config file
2. Call `ApplyEnvironmentVariables(cfg)` to apply env var overrides

The `CONFIGURATION_FILEPATH` environment variable must point to a valid JSON file. In Kubernetes, this is typically set to `/etc/service-config.json` with the actual config mounted from a ConfigMap.

## Environment variable naming

Env vars follow this convention: `DINNER_DONE_BETTER_` + `PREFIX_` + `FIELD`.

- **Prefix**: `DINNER_DONE_BETTER_` (see `config.EnvVarPrefix`)
- **Section prefix**: From the `envPrefix` struct tag on nested config fields (e.g. `GRPC_`, `META_`, `COOKIES_`)
- **Field name**: From the `env` struct tag on each field (e.g. `PORT`, `DEBUG`, `DOMAIN`)

For example, `config.Service.GRPCServer.Port` maps to `DINNER_DONE_BETTER_GRPC_SERVER_PORT` because:

- `Service` has no `envPrefix`, so its blocks keep their own
- `GRPCServer` has `envPrefix:"GRPC_SERVER_"`
- `Port` has `env:"PORT"`

## Valid environment variables

A canonical list of the variables that override the API server's config is **programmatically generated** in:

```bash
.env.example
```

That file is produced by `cmd/tools/codegen/valid_env_vars`, which walks the config structs (`APIServiceConfig`, `MCPServiceConfig`, etc.) and extracts env var names from `env` and `envPrefix` tags. Regenerate it with:

```bash
make env_vars
```

## Config struct tagging

Config structs use struct tags to control JSON and env behavior:

```go
// Root-level section: envPrefix defines the env var prefix for all nested fields
GRPCServer grpc.Config `envPrefix:"GRPC_" json:"grpc"`

// Nested field: env defines the env var suffix
Port uint16 `env:"PORT" json:"port"`
```

- `json:"..."` – JSON field name for the config file
- `env:"..."` – Suffix for the environment variable (combined with prefix)
- `envPrefix:"..."` – Prefix for nested struct fields (only on struct-typed fields)
- `json:"-"` – Field is excluded from both JSON and env parsing

## platform-go's `service.Config`, and the blocks that stay out of it

Every long-lived process and the database cleaner is composed by platform's `service` package.
Each one's config struct carries a `service.Config` as its `Service` field, with no `envPrefix`, so
its blocks keep the names platform gives them — `DATABASE_`, `OBSERVABILITY_`, `OPERATIONS_`,
`GRPC_SERVER_`. The composition root validates the config, calls `service.Register` with that
field, registers what platform cannot know, and builds the process with `service.New`.

**A platform subsystem is configured by putting its block in `Service`, and registered by
`service.Register` from it.** This replaces the earlier rule, which embedded platform's bare
`Config` structs (`outbox.RelayConfig`, `saga.WorkerConfig`, ...) in this application's own tree
and refused platform's `*/config` packages as a second configuration path. `service.Config` is
the tree those packages belong to, so there is no second path: a block present is a subsystem
built, a block absent is one that is not, and the reapers a store owns — operations' recovery and
reap, saga retention — are scheduled because their block is present rather than because somebody
remembered.

Validation is not optional. `env:",init"` allocates every block whether or not anything was put in
it, and `service.Config.ValidateWithContext` is what releases the empty ones; each process config's
`ValidateWithContext` runs it first, and every `BuildInjector` validates before it registers. Each
process also names the blocks it cannot run without — a scheduler with no `Operations` block boots
and never fulfills a privacy request — so a missing one is a failed render rather than a quiet
deployment.

A block stays out of `Service`, and its subsystem stays registered by hand, when `service.Register`
would build the wrong thing. Each of these is named where its field lives:

| block | process | why it is not in `Service` |
| --- | --- | --- |
| `Outbox` | scheduler, API | Every outbox row this application writes goes through a writer carrying the search index side effect, and `outbox/config` builds its writer from configuration alone. The relay's settings are `SchedulerConfig.OutboxRelay` (`OUTBOX_RELAY_`). Without `Outbox` the recording spine is assembled by hand too, since `service.Register` builds it only beside `Audit`, `Webhooks` and `Outbox`. Upstream: [platform-go#1147](https://github.com/primandproper/platform-go/issues/1147). |
| `Audit` | scheduler, API | This application's recorder is platform's with the impersonating administrator attached to each entry. Retention reads `SchedulerConfig.AuditLog` (`AUDIT_LOG_`). |
| `MobileNotifications` | scheduler, async messages | `service.Config` validates every block it holds, and a rendered file has no APNs credentials — they arrive from the environment at startup. `PushNotifications` stays a field of its own. |
| `Routing` | API | The API server builds its router itself, with its routes on it; a `Routing` block would register a second, empty one. |
| `Operations`, `DataPrivacy`, `Webhooks`, `Saga` | API | Each block registers its worker too, and the workers run in the scheduler. The API server registers the enqueue-and-read halves by hand — through platform's own `Register*` bridges where they exist (`dataprivacycfg.RegisterRequestService`). |
| `OAuth2Server` | db cleaner | It registers the authorization server with the store, and a server needs a subject authenticator a sweep has no use for. The cleaner reads `DBCleanerConfig.OAuth2` (`OAUTH2_`). |

The domain stores this application registers itself — identity, comments, settings, waitlists,
issue reports, billing, notifications — are not in `Service` either. They are wired with this
application's table prefix and recording, and moving each onto its block is its own change.

A field that sits beside a block it could be confused with uses a prefix that cannot collide with
it — `OUTBOX_RELAY_` rather than `OUTBOX_`, `AUDIT_LOG_` rather than `AUDIT_` — because an
environment variable that reached the `Service` block would switch that subsystem on beside the
hand-registered one, and samber/do refuses a second provider by panicking.

### The table prefix is deliberately not configuration

Every prefix-only config package is refused for the same reason, and it is worth stating once. A
prefix has to match the prefix the migration was rendered with. Both come from one Go constant —
`branding.TablePrefix` — read by `internal/repositories/postgres/migrations`
when it renders the DDL and by the store when it builds its statements. An environment variable
that could set one of those without the other is a way to point a store at tables that do not
exist, and the failure is at the first query rather than at boot.

### And the four `*/http` surfaces, for completeness

`dataprivacy/http`, `mediaregistry/http`, `operations/http` and `sessions/http` are **no**, on one
reason: this application's API is gRPC, and its HTTP server exists for the handful of things that
cannot be — the OAuth2 authorization endpoints, the payment processors' webhooks, and health. A
domain reachable over gRPC does not want a second transport with a second authorization surface to
keep in step.

`sessions/cache` is **no** for the reason `links` was parked: it stores records in a
`cache.Cache`, and nothing but localdev provisions a Redis. This deployment uses
`sessions/database`, which is what the session table in the migrations is.

### The `*/config` packages registered by hand

Where a subsystem stays out of `Service`, its `*/config` package's constructors and `Register*`
bridges are still how it is built. One of them is worth knowing about: the notifications store is built
through `notificationscfg.NewStore` with an **empty** `Config{}` and then has the only field that
`Config` carries overridden by `WithStoreOptions(WithTablePrefix(...))`. It reads as though the
prefix comes from the environment and it does not. Harmless, and left alone rather than churned —
but if that file is being edited for another reason, calling `notifications.NewSQLStore` directly
is what it actually does.

`dataprivacy/config` is the one earning its keep most clearly: `RegisterAuditEraser` is a policy
flag — whether this deployment erases audit records at all — and that is genuinely a deployment's
answer rather than a constant.

## Deployment

- **Kubernetes**: Each deployment/cronjob sets `CONFIGURATION_FILEPATH` and mounts a ConfigMap containing the service-specific JSON config. Environment-specific overrides (e.g. database URLs, secrets) are often applied via patches that inject additional env vars.
- **Docker Compose**: `CONFIGURATION_FILEPATH` is set to `/etc/config`, with config files mounted from `deploy/environments/*/config_files/`.
- **Local development**: Config files are generated by `cmd/tools/codegen/configs` into `deploy/environments/localdev/config_files/`; `CONFIGURATION_FILEPATH` is set accordingly. The localdev kustomization reads those same files rather than a second rendered copy, which is why `skaffold.yaml` passes `--load-restrictor LoadRestrictionsNone` on that profile.

## Production deployment (Terraform Cloud)

Backend Terraform (`backend/deploy/environments/prod/terraform`) requires these variables. Add them in Terraform Cloud: Workspace → Variables.

| Variable                               | Description                                                                                                                                        |
|----------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------|
| `POSTHOG_API_KEY`                      | PostHog Project API Key (for event ingestion: main analytics and proxy sources)                                                                    |
| `POSTHOG_PERSONAL_API_KEY`             | PostHog Personal API Key (for feature flags API; create in [PostHog Settings → Personal API Keys](https://app.posthog.com/settings/user-api-keys)) |
| `SENDGRID_API_KEY`                     | SendGrid API token                                                                                                                                 |
| `ALGOLIA_APPLICATION_ID`               | Algolia app ID                                                                                                                                     |
| `ALGOLIA_API_KEY`                      | Algolia write API key                                                                                                                              |
| `ADMIN_WEBAPP_OAUTH2_CLIENT_ID`        | Admin OAuth2 client ID                                                                                                                             |
| `ADMIN_WEBAPP_OAUTH2_CLIENT_SECRET`    | Admin OAuth2 client secret                                                                                                                         |
| `CONSUMER_WEBAPP_OAUTH2_CLIENT_ID`     | Consumer webapp (root site) OAuth2 client ID                                                                                                       |
| `CONSUMER_WEBAPP_OAUTH2_CLIENT_SECRET` | Consumer webapp (root site) OAuth2 client secret                                                                                                   |

## Related

- `internal/config/configs.go` – Config struct definitions and `LoadConfigFromEnvironment` / `LoadConfigFromPath`
- `internal/config/env_vars.go` – The env var options every loader shares (prefix, debug logging)
- `.env.example` – Generated list of the API server's env var names and defaults
- `cmd/tools/codegen/valid_env_vars` – Code generator for `.env.example`
- `cmd/tools/codegen/configs` – Generates JSON config files per environment
