# Dinner Done Better

Monorepo for a meal planning application built as a reusable service template.

## Repository Structure

- `backend/` — Go backend (API server, workers, admin app, MCP server)
- `frontend/` — Web frontend
- `ios/` — iOS mobile app
- `proto/` — Protocol Buffer definitions for gRPC services
- `infra/` — Infrastructure Terraform (GKE, networking, DNS, Caddy)
- `docs/` — Cross-cutting documentation (identity, auth, meals, recipes, deployment)

## Template Philosophy

This repo serves dual purposes: a working meal planning app and a reusable service template. The platform framework (database, cache, observability, messaging, etc.) lives in a separate repo at `github.com/primandproper/platform-go/v2` and is imported as a dependency. `internal/domain/mealplanning` is the example domain built on top. Someone should be able to fork this and swap the meal planning domain for their own without modifying core infrastructure.

## Cross-Cutting Commands

```bash
make proto    # Format + generate proto (Go + Swift + Typescript) from repo root
make build    # builds the backend frontend and iOS folders
make format   # formats the backend frontend and iOS folders
make lint     # lints the backend frontend and iOS folders
make test     # tests the backend frontend and iOS folders
```

## Documentation

- `docs/identity.md` — Users, accounts, memberships, roles
- `docs/audit.md` — Tamper-evident audit log: the hash chain, scopes, redaction, retention
- `docs/data-privacy.md` — Disclosure lifecycle, artifact encryption, the expiry reaper
- `docs/auth-flow.md` — Authentication flow (password, passkey, OAuth2, gRPC interceptor)
- `docs/recipes.md` — Recipe object model, bridge tables, option groups, scaling
- `docs/meals.md` — Meals, components, scaling
- `docs/meal_planning.md` — Meal plans, voting (Schulze), grocery lists, background workers
- `docs/deployment.md` — Release-based deployment, GitHub Actions, Terraform Cloud
- `docs/spin-up-from-scratch.md` — Greenfield provisioning guide
- `docs/required-secrets-and-variables.md` — All Terraform and GitHub Actions secrets
