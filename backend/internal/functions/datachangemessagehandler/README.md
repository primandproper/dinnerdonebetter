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

Consumes the data change events every write publishes through the outbox. Each message is a
`webhooks.Envelope` naming its event; the handler routes on the event type and decodes the payload
into the type that event names — this application's own `datachanges.Message` for its meal planning
events, platform's payload types (`identity.UserEvent`, `passwordreset.TokenEvent`, …) for the
events platform's stores record. For each event, it:

1. **Analytics** — Reports the event to the customer data platform, for the events on the
   allowlist in `internal/domain/analytics` and no others
2. **Outbound notifications** — Publishes to the Outbound Emails topic when somebody is owed a mail
   that carries no secret (a password changed, a reset spent), and to the Mobile Notifications
   topic when a household is owed a push. A mail that carries a link — verification, password
   reset, username reminder, invitation, waitlist confirmation — never passes through here; see
   Queued Mail below. Which of a domain's events imply a mail is that domain's to say: each
   contributes an `OutboundNotificationHandler` through `RegisterAsyncDataChangeMessageHandler`
   (the meal planning one is `internal/services/mealplanning/outbound`), and the handler here
   ranges over the list. The handler over platform's identity events is this package's own

Webhook deliveries and search index events are no longer this handler's: both are rows the
transaction that caused the event writes, through platform's recording spine (built in
`internal/recordingspine`).

### Queued Mail (`mail.Drainer`)

platform's `notifications/mail` transport for the mail its identity, sign-in, password reset and
waitlist doors owe. Each door hands its mail to the `QueuedMailer` once its write commits, which
puts it on the outbox under the queued mail topic; this process runs the `Drainer` on that topic,
renders each mail in this application's words through `internal/services/identity/emails`'
`Renderer`, and sends it. The topic is the Drainer's alone: every message on it carries the secret
its mail delivers.

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
