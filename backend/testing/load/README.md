# k6 Read-Only Load Tests

Load tests for the Dinner Done Better API that exercise read-only gRPC endpoints. No data is written; only shared enumeration data (valid ingredients, instruments, vessels, etc.) is read.

## Auth Flow

Uses the same flow as the mobile and web apps:

1. **LoginForToken** (gRPC, unauthenticated) – submit username/password, get JWT
2. **OAuth2 authorize** – GET `/oauth2/authorize` with Bearer JWT + client creds, receive auth code via redirect
3. **OAuth2 token exchange** – POST `/oauth2/token` with code + client creds, get access token
4. **gRPC calls** – use access token as Bearer for read-only RPCs

## Prerequisites

- [k6](https://k6.io/docs/getting-started/installation/) installed
- Load test user (username + password)
- Load test OAuth2 client (client_id + client_secret)
- gRPC server reflection enabled (for proto discovery). The `proto/` symlink points to the repo proto dir for manual proto loading if reflection is unavailable; those files import `primandproper/platform/filtering/v1/filtering.proto` from platform-go, so a manual load needs that module's `filtering/proto` directory on the import path too (see `proto/README.md`).

## Environment Variables

| Variable                  | Description                                                                                      |
|---------------------------|--------------------------------------------------------------------------------------------------|
| `K6_GRPC_TARGET`          | gRPC API address, e.g. `api.dinnerdonebetter.com:443` (default: production)                      |
| `K6_HTTP_BASE`            | HTTP API base URL for OAuth2, e.g. `https://http-api.dinnerdonebetter.com` (default: production) |
| `K6_GRPC_PLAINTEXT`       | Set to `true` for local dev (no TLS)                                                             |
| `K6_OAUTH2_CLIENT_ID`     | Load test OAuth2 client ID                                                                       |
| `K6_OAUTH2_CLIENT_SECRET` | Load test OAuth2 client secret                                                                   |
| `K6_LOADTEST_USERNAME`    | Load test user username                                                                          |
| `K6_LOADTEST_PASSWORD`    | Load test user password                                                                          |

## Running

From the **repo root**:

```bash
k6 run \
  --vus 10 \
  --duration 5m \
  -e K6_GRPC_TARGET=api.dinnerdonebetter.com:443 \
  -e K6_HTTP_BASE=https://http-api.dinnerdonebetter.com \
  -e K6_OAUTH2_CLIENT_ID=your_client_id \
  -e K6_OAUTH2_CLIENT_SECRET=your_client_secret \
  -e K6_LOADTEST_USERNAME=loadtest_user \
  -e K6_LOADTEST_PASSWORD=your_password \
  backend/testing/load/script.js
```

From the **backend** directory, use `testing/load/script.js` instead.

Or use a `.env` file (do not commit secrets):

```bash
export $(grep -v '^#' .env.loadtest | xargs)
k6 run --vus 10 --duration 5m backend/testing/load/script.js
```

## Local Development

For local dev (plaintext gRPC, different ports), set:

```bash
K6_GRPC_TARGET=localhost:8001
K6_HTTP_BASE=http://localhost:8000
K6_GRPC_PLAINTEXT=true
```

Then run with `--env` or export before `k6 run`.

## Endpoints Exercised

Read-only only. Setup fetches `account_id` and `user_id` via GetActiveAccount/GetSelf for context-dependent calls.

**Valid enums (shared):** GetValidInstruments, GetValidIngredients, GetValidVessels, GetValidPreparations, GetValidIngredientGroups, GetValidMeasurementUnits, SearchForValidIngredients, SearchForValidInstruments, SearchForValidVessels, SearchForValidPreparations, SearchForValidMeasurementUnits

**Mealplanning (account/user-scoped):** GetRecipes, GetMealPlansForAccount, GetMeals, GetRecipeLists, GetMealLists, SearchForRecipes, SearchForMealEligibleRecipes, SearchForMeals, GetAccountInstrumentOwnerships, GetUserIngredientPreferences, SearchForRecipesWithInstrumentOwnership, SearchForValidIngredientGroups, SearchForValidIngredientStates, GetValidPrepTaskConfigs, GetValidIngredientStates, GetValidPreparationInstruments, GetValidPreparationVessels

**Webhooks:** GetWebhooks, GetWebhookTriggerEvents

**Identity:** GetAccounts, GetReceivedAccountInvitations, GetSentAccountInvitations (admin-only GetUsersForAccount, GetAccount, GetAccountsForUser, SearchForUsers excluded)

**Settings:** GetSettingDefinitions, GetSettingValues, ResolveSettings (there is no per-account read: a setting value belongs to a person, not a membership)

**Audit:** GetAuditLogEntriesForAccount, GetAuditLogEntriesForUser

**Notifications:** GetUserNotifications, GetUserDeviceTokens

**Issue reports:** GetIssueReports, GetIssueReportsByStatus, GetIssueReportsBySubjectType

**Auth:** GetAuthStatus, GetActiveAccount, GetSelf

## Token Lifetime

The access token is fetched once in `setup()` and shared by all VUs. For runs longer than the token lifetime (~24h), consider adding token refresh or re-running the test in segments.
