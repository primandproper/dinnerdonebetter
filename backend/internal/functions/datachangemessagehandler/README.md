# Async Message Handler

The Async Message Handler is the **central consumer** for all (Pub/Sub)/Redis/SQS) topics in the system. It runs as a long-lived process, subscribing to every async message queue and dispatching events to the appropriate handlers.

This is one of the most critical services in the backend: if it stops running, emails won't be sent, search indexes won't update, webhooks won't fire, and data change events won't propagate.

## Topics Consumed

| Topic                          | Purpose                                                                    |
|--------------------------------|----------------------------------------------------------------------------|
| **Data Changes**               | Audit events emitted when data changes (user signup, recipe created, etc.) |
| **Outbound Emails**            | Email send requests (verification, password reset, notifications, etc.)    |
| **Search Index Requests**      | Requests to index or remove records from the text search index             |
| **Webhook Execution Requests** | Outbound webhook deliveries to customer-configured URLs                    |
| **User Data Aggregation**      | GDPR/CCPA data export requests                                             |

## Event Handlers

### Data Changes (`DataChangesEventHandler`)

Consumes audit events from across the system. For each event, it:

1. **Analytics** — Reports the event to the customer data platform (CDP)
2. **Webhooks** — If the event type has registered webhooks, publishes `WebhookExecutionRequest` messages
3. **Outbound notifications** — Publishes to the Outbound Emails topic when users need emails (e.g., verification, password reset)
4. **Search index** — Publishes `IndexRequest` messages for entities that changed (users, recipes, meals, etc.)

Event types include user lifecycle (signup, archive, email/username changes), meal planning (recipes, meals, grocery lists), OAuth clients, and more.

### Outbound Emails (`OutboundEmailsEventHandler`)

Sends emails via the configured email provider. Messages contain recipient, subject, body, and metadata. Also reports send events to analytics.

### Search Index Requests (`SearchIndexRequestsEventHandler`)

Performs the actual search indexing. Dispatches by `IndexType`:

- **Meal planning** — Recipes, meals, valid ingredients, instruments, measurement units, preparations, ingredient states, vessels
- **Identity** — Users

Index requests can originate from the Data Changes handler (real-time) or from the search data index scheduler (batch backfill), registered in `internal/build/jobs/scheduler/jobs.go`.

### Webhook Execution Requests (`WebhookExecutionRequestsEventHandler`)

Executes outbound webhooks: fetches the webhook config, signs the payload, and HTTP POSTs to the customer's configured URL. Handles retries and error reporting.

### User Data Aggregation (`UserDataAggregationEventHandler`)

GDPR/CCPA compliance: fetches a user's complete data collection from the data privacy repo, marshals it to JSON, and saves it to object storage under the report ID.

## Message Queue Provider

The handler supports multiple backends (configured via `CONSUMER_PROVIDER`):

- **Pub/Sub** — GCP Pub/Sub (production)
- **Redis** — Redis Streams (local dev)
- **SQS** — AWS SQS
- **Noop** — For testing

## Deployment

Runs as `ddb worker async-messages`: a Kubernetes Deployment that stays up and consumes messages continuously. Configured via `AsyncMessageHandlerConfig` and the standard queue config (`QueuesConfig`). See `cmd/ddb/README.md`.

## Related

- **Publishers** — The API service and various cron jobs publish to these topics
- Search data index scheduler — scheduled job that backfills search index requests for records that need indexing; see `internal/build/jobs/search_data_index_scheduler` and `cmd/ddb/README.md`
