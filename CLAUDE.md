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

This is one domain on top of an imported platform. The framework (database, cache, observability, messaging, identity, audit, billing, webhooks, and the rest) lives in a separate repo at `github.com/primandproper/platform-go/v15`, with `primitives-go/v2` underneath it, and is imported as a dependency; every non-mealplanning store here is a thin wiring of a platform package. `internal/domain/mealplanning` is the only domain, and no new ones are planned. The work is simplification over extension: when something mealplanning does is something platform could own, the answer is to adopt platform's and delete the local copy, not to generalize the local copy. There is no add-a-domain guide and no CRUD kit, on purpose — `backend/docs/writing_go.md` records why.

## Deployment Status

**Nothing is deployed.** No instance of this service runs anywhere, and there are zero downstream
consumers. Backwards-incompatible changes — schema, proto and gRPC surface, wire formats, stored
data — cost nothing in compatibility terms, so prefer the change that leaves the better end state
over the one that preserves the current one.

Two things this does *not* license:

- **platform-go is versioned separately** and is a reusable framework in its own right. A change
  that has to happen there still needs a release and a version bump here, and that sequencing is a
  real constraint on work in this repo regardless of what is deployed.
- **Editing an already-applied migration is not free**, just cheap — see `backend/docs/migrations.md`.

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
- `docs/data-privacy.md` — Disclosure lifecycle, artifact encryption, the expiry reaper
- `docs/auth-flow.md` — Authentication flow (password, passkey, OAuth2, gRPC interceptor)
- `docs/recipes.md` — Recipe object model, bridge tables, option groups, scaling
- `docs/meals.md` — Meals, components, scaling
- `docs/meal_planning.md` — Meal plans, voting (Schulze), grocery lists, background workers
- `docs/webhooks.md` — Outbound webhooks: event catalog, signing, retry, secrets
- `docs/deployment.md` — Release-based deployment, GitHub Actions, Terraform Cloud
- `docs/spin-up-from-scratch.md` — Greenfield provisioning guide
- `docs/required-secrets-and-variables.md` — All Terraform and GitHub Actions secrets
