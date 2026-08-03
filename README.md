# Dinner Done Better

**Dinner Done Better** is a meal plan management platform that helps groups decide what to eat through collaborative voting, recipe discovery, and automated grocery lists.

- **Meal planning** — Create meal plans with events (e.g. “Dinner on Tuesday”), propose options, and vote (Schulze method).
- **Recipes & meals** — Rich recipes (ingredients, steps, timing) and meals (collections of recipes).
- **Grocery lists** — Generated from finalized meal plans.
- **Web and mobile** — Admin and consumer web apps, plus a native iOS app.

---

## Repository layout

This is a **monorepo**. Main areas:

| Path | Description |
| ------ | ------------- |
| **`backend/`** | Go API server, workers, and tooling. gRPC + HTTP, PostgreSQL, Pub/Sub, Wire, sqlc. See [backend/README.md](backend/README.md). |
| **`frontend/`** | Web apps: **admin** and **consumer** (SvelteKit + Vite). Shared packages: `api-client` (proto-generated TS), `ui`, `logger`. Built and deployed via own Skaffold/Docker/Kustomize. |
| **`ios/`** | Native iOS app (Swift). Xcode project, Fastlane, tests. |
| **`proto/`** | Shared API definitions (Protocol Buffers). Generated Go, Swift, and TypeScript code used by backend, frontend, and iOS. |
| **`infra/`** | Infrastructure and deploy: Terraform (GCP, GKE, networking, DNS), Skaffold, scripts. |
| **`docs/`** | Design and runbooks: auth, identity, meal planning, recipes, deployment, spin-up, secrets. |

### Backend at a glance

- **Services** (in `backend/cmd/services/`): `api` (main gRPC/HTTP), `mcp`. Admin and consumer **webapps** live in **`frontend/`** (SvelteKit).
- **Workers** (in `backend/cmd/workers/`): meal plan finalizer, grocery list initializer, task creator, search index scheduler, mobile notification scheduler, DB cleaner, etc.
- **Functions**: async handlers (e.g. data-change message handler) for Pub/Sub and queues.
- **Stack**: Go 1.26, Chi, Wire, sqlc, PostgreSQL, Redis, GCP (Cloud SQL, Pub/Sub, Secret Manager, etc.), Algolia, Stripe, Resend, Firebase/APNs.

---

## Quick start

**Prerequisites:** Go 1.26, Node.js (for frontend), Make, Docker & Docker Compose. See [docs/spin-up-from-scratch.md](docs/spin-up-from-scratch.md) for a full list (Terraform, sqlc, wire, etc.).

```bash
# One-time setup (root)
make setup

# Backend: configs, codegen
cd backend && make setup

# Frontend: install dependencies (from repo root)
cd frontend && npm install

# Run local dev server (API + workers, local Postgres via compose)
cd backend && make dev
```

In separate terminals, run the web apps (with the API already running):

```bash
# Admin webapp (from frontend/)
cd frontend && npm run dev -w admin

# Consumer webapp (from frontend/)
cd frontend && npm run dev -w consumer
```

- **API:** <http://localhost:8000> (HTTP) · `localhost:8001` (gRPC)
- **Admin webapp:** Vite dev server (default <http://localhost:5173> when run alone)
- **Consumer webapp:** Vite dev server (use a different port if both run, e.g. `--port 5174` for the second)

See [backend/README.md](backend/README.md) for backend-only quick start. Frontend apps use `.env` in `frontend/admin` and `frontend/consumer` for API URLs and auth (see `.env.example` there).

---

## Common commands (from repo root)

| Command | Description |
| -------- | ------------- |
| `make setup` | Install tools, format YAML; then `backend` setup. |
| `make format` | Format YAML, Go, Terraform, Swift. |
| `make lint` | Lint backend and iOS. |
| `make test` | Run backend and iOS tests. |
| `make proto` | Format protos, generate Go, Swift, and TypeScript from `proto/`. |
| `make deploy_localdev` | Deploy to Docker Desktop Kubernetes (infra + backend via Skaffold). |
| `make deploy_terraform_prod` | Apply prod Terraform (infra then backend). |
| `make deploy_prod` | Run prod application deploy (Skaffold) via `./scripts/deploy-prod-local.sh`. |
| `make deploy_prod_frontend` | Deploy only the frontend (consumer + admin webapps) to prod via `./scripts/deploy-prod-frontend.sh`. |
| `make verify_prod` | Post-deploy verification (Skaffold verify + scripts). |

**Frontend** (from `frontend/`): `npm install`, `npm run dev -w admin` / `npm run dev -w consumer`, `npm run build`, `npm run lint`, `npm run test`.

---

## Branch and deployment strategy

- **Default branch:** `main` — day-to-day work and CI target `main`.
- **Production:** Deploy by **publishing a GitHub Release** (tag = semver, e.g. `v1.2.3`). Publishing the release triggers the deploy pipeline.
- **Version** is embedded in backend binaries (version/health endpoints), in Grafana deployment/terraform annotations, and on Kubernetes resources via the `app.kubernetes.io/version` label.

Deploy pipeline (see [docs/deployment.md](docs/deployment.md)):

1. **baseline-infra** — Terraform for GCP, GKE, networking, DNS, email, IAM.
2. **backend-infra** — Terraform for Cloud SQL, storage, Pub/Sub, Algolia, K8s namespaces.
3. **applications** — Build and deploy app workloads with Skaffold (API, workers, cron jobs).

---

## Documentation

| Document | Purpose |
| ---------- | --------- |
| [docs/deployment.md](docs/deployment.md) | Release-based deployment, deploy pipeline, local prod deploy, verification. |
| [docs/spin-up-from-scratch.md](docs/spin-up-from-scratch.md) | Greenfield setup: GCP, Terraform Cloud, external services, secrets. |
| [docs/required-secrets-and-variables.md](docs/required-secrets-and-variables.md) | Required secrets and env vars. |
| [docs/meal_planning.md](docs/meal_planning.md) | Meal planning concepts, flows, and architecture. |
| [docs/recipes.md](docs/recipes.md) | Recipe model and behavior. |
| [docs/meals.md](docs/meals.md) | Meals (collections of recipes). |
| [docs/identity.md](docs/identity.md) | Identity and accounts. |
| [docs/auth-flow.md](docs/auth-flow.md) | Authentication flows. |
| [backend/README.md](backend/README.md) | Backend quick start, Make targets, architecture diagram. |
| [backend/docs/](backend/docs/) | Configuration, migrations, payments, writing Go, adding domains. |

---

## Protobuf and code generation

- **Source:** `proto/` — per-domain `.proto` files (e.g. `mealplanning/`, `auth/`, `identity/`).
- **Generate:** `make proto` (format with buf, then `proto_golang` and `proto_swift`). For the frontend API client: `make proto_typescript` (outputs to `frontend/packages/api-client/src`).
- **Output:** Go → `backend/internal/grpc`, Swift → `ios/ios/Generated`, TypeScript → `frontend/packages/api-client/src`.

Requires `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, and for Swift `protoc-gen-swift` and `protoc-gen-grpc-swift`. The root Makefile has `ensure_*` targets; see `make proto` and [Makefile](Makefile) around the `PROTO_*` variables.

---

## Local Kubernetes (Docker Desktop)

```bash
make deploy_localdev
```

Uses Skaffold to deploy infra and backend to a `localdev` namespace. The **admin** and **consumer** webapps are built and deployed from **`frontend/`** (see `frontend/skaffold.yaml`); run that separately if you want them in the cluster. Afterward:

- API: <http://localhost:8000>
- Admin and consumer webapps: as configured by frontend deploy (e.g. port-forward from the `dinner-done-better-admin-webapp-svc` / `dinner-done-better-consumer-webapp-svc` services).

Teardown: `make nuke_localdev`.

---

## License and contributing

See the repository’s license file and contribution guidelines (if present). For external services and accounts required to run or deploy, see [docs/spin-up-from-scratch.md](docs/spin-up-from-scratch.md) and [docs/required-secrets-and-variables.md](docs/required-secrets-and-variables.md).
