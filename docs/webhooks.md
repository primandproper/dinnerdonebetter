# Webhooks

Outbound webhooks let an account receive an HTTP request whenever something happens in it. They
are built on platform-go's `webhooks` package: signed, retried per subscriber, ordered, and
replayable.

## Before you read the rest

**Webhook delivery did not work before this.** `webhook_trigger_configs.trigger_event` was a
foreign key into a `webhook_trigger_events` table whose IDs were randomly generated, while the
fan-out looked webhooks up by the event type *string* — a random ID can never equal
`meal_plan_created`, so no webhook ever matched an event and nothing was ever delivered. Anything
you remember about how webhooks behaved is about a code path that never fired.

That is why this change breaks the API without a deprecation window: there is nobody to
deprecate.

## The model

| Thing | Where it lives | What it is |
| --- | --- | --- |
| Webhook | `webhooks` | The account-facing record: name, URL, owner, account |
| Trigger config | `webhook_trigger_configs` | One webhook's subscription to one event type |
| Endpoint | `webhooks_endpoints` | The delivery record: URL, content type, signing secret |
| Subscription | `webhooks_subscriptions` | What fan-out actually reads |
| Delivery / dispatch / attempt | `webhooks_deliveries`, `_dispatches`, `_attempts` | One event, one endpoint's copy of it, and every attempt at it |

A webhook and its endpoint share an ID, so "the endpoint for this webhook" is never a lookup that
can return the wrong answer.

### Tenancy

An endpoint belongs to an account, and that is a column: `webhooks_endpoints.scope` holds the
account ID, as a `tenancy.Scope`. Every delivery carries the same scope, and `EndpointsForEvent`
resolves subscribers within it — so one account's `meal_plan_created` never reaches another
account's endpoints.

```text
webhooks_endpoints                     webhooks_subscriptions
  id        | scope                      endpoint_id | event_type
  wh_abc123 | acct_9f2                   wh_abc123   | meal_plan_created
```

The event type stored against an endpoint is the plain catalog type — the same string the catalog
holds, and the same one a subscriber reads out of `X-Platform-Event`.

Registration and dispatch both go through platform-go's `webhooks.Dispatcher`, which is what
applies the catalog gate and the SSRF policy. The write side is
`internal/repositories/postgres/webhooks` (see `endpoints.go`); it reaches the `Store` directly
for the two operations `Dispatcher` does not offer — replacing an endpoint's subscription set,
and retiring it.

## Events

The catalog is **generated** from the `*EventType` constants the domains declare
(`internal/domain/webhooks/catalog`, regenerate with `make webhook_catalog`). An event that
exists is subscribable; one that does not cannot be subscribed to by typo.

Some events the application publishes are deliberately **not** deliverable — sign-ins, session
lifecycle, password and two-factor changes, account membership changes, OAuth2 client lifecycle.
An endpoint URL is attacker-supplied, and these would be a live feed of an account's security
activity. See `internal/domain/webhooks/catalog/excluded.go`. Subscribing to one is rejected;
emitting one dispatches nothing.

`GetWebhookEventTypes` returns what an account may subscribe to.

## Dispatch is transactional

Deliveries are rows written inside the transaction that caused the event, through
`internal/repositories/postgres/events`. A delivery and the state change it describes commit
together or not at all — there is no window in which a meal plan exists and its notification was
lost.

The delivery worker runs in the `scheduler` process beside the outbox relay. Its own tick also
reaps delivered dispatches past the retention window, so retention needs no separate job.

## Verifying a delivery

Every request carries:

| Header | Meaning |
| --- | --- |
| `X-Platform-Signature` | `v1,t=<unix>,s=<hex>` — see below |
| `X-Platform-Timestamp` | The signing timestamp, so you can reject a stale request before doing any HMAC work |
| `X-Platform-Event` | The event type, so you can route without parsing the body |
| `X-Platform-Delivery` | Stable across every retry and replay — **this is your deduplication key** |
| `X-Platform-Attempt` | Which attempt this is, 1-indexed |

The signature is `HMAC-SHA256(secret, "v1." + timestamp + "." + body)`, hex-encoded. Both the
scheme and the timestamp are inside the signed material. The timestamp is what makes a captured
request expire; verify it before computing any HMAC, so a replay flood costs you nothing.

During a secret rotation the header carries several `s=` components — accept the delivery if
**any** of them matches.

```go
import "github.com/primandproper/platform-go/v9/webhooks"

func handler(w http.ResponseWriter, r *http.Request) {
    // The exact bytes received. Decoding and re-encoding changes key order and whitespace,
    // and the signature covers bytes, not meaning.
    body, err := io.ReadAll(r.Body)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }

    if err = webhooks.Verify(
        webhooks.Secret{Current: mySecret},
        body,
        r.Header.Get(webhooks.SignatureHeader),
    ); err != nil {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    // ...
}
```

Do not compare signatures with `==`, and do not skip the timestamp check. `webhooks.Verify` does
both correctly; reimplementing it from this prose is how these schemes get got wrong.

## Secrets

Each **endpoint** has its own signing secret, not each account. A single account-wide key cannot
be rolled without breaking every subscriber for that account at the same instant, which in
practice means it never gets rolled.

The secret is returned **once**, by the call that mints it — `CreateWebhook`, or
`RotateWebhookSecret`. No read path can produce it. If you lose it, rotate.

Rotating keeps the outgoing key active: deliveries are signed under both until you rotate again,
so a subscriber can accept either while it switches over.

> The per-account `Account.WebhookEncryptionKey` is no longer used for signing. The column
> remains for now; removing it is a separate change.

## Delivery behavior

- **Retry** is per endpoint. A delivery that fanned out to five subscribers is not "failed" when
  four accepted it and the fifth is on its sixth attempt.
- **Backoff** is exponential with full jitter, persisted as a timestamp so it survives a worker
  restart. Ten attempts reaching an hour; past that the dispatch is marked dead and left for an
  operator to replay.
- **4xx other than 408 and 429** goes straight to dead — the subscriber understood and refused,
  and retrying twenty more times changes nothing.
- **Ordering** is per `(endpoint, ordering key)`. Deliveries sharing a key reach an endpoint in
  dispatch order, so `resource.updated` cannot overtake `resource.created`. The key defaults to
  the account; repositories that know their subject resource pass
  `events.WithOrderingKey(resourceID)`.
- **Circuit breaking** is per endpoint, so a subscriber that has been 500ing for a week stops
  competing for worker slots with healthy ones. A short-circuited delivery is not charged an
  attempt.
- **At-least-once.** A crash between a 200 and the row update redelivers. Deduplicate on
  `X-Platform-Delivery`.

### SSRF

Endpoint URLs are user-supplied and the worker makes authenticated requests to them. `https` only,
and no host resolving into loopback, link-local, private, or non-global space — checked at
registration, where the rejection can be reported, and again at delivery, because DNS is mutable.
Redirects are refused rather than followed.

## What changed for API clients

- `WebhookCreationRequestInput.events` is now `event_types`, a list of catalog event type strings.
- `WebhookTriggerConfig.trigger_event_id` is now `event_type`.
- `CreateWebhookTriggerEvent` / `Get` / `Update` / `ArchiveWebhookTriggerEvent` are gone. The
  catalog is generated and read-only; use `GetWebhookEventTypes`.
- `RotateWebhookSecret` is new. `CreateWebhookResponse` now carries `secret`.
- **Method is POST only** and **content type is `application/json` only**. XML is gone: a delivery
  carries one payload shared by every subscriber of that event, so per-endpoint XML would mean
  dispatching the same event twice, and nothing consumed it.

### Webhooks created before this change

They have no delivery endpoint. The migration deliberately backfilled none — minting a signing
secret in SQL would have meant either requiring `pgcrypto` everywhere or using a non-CSPRNG, and
a weak secret would start signing real deliveries the moment its owner subscribed it to
something.

Their trigger configs were deleted, because the catalog rows they pointed at carry
operator-chosen names that are not event types, and guessing a mapping would produce
subscriptions nobody asked for.

To adopt one: call `RotateWebhookSecret`, which registers the endpoint and returns a secret, then
subscribe it with `AddWebhookTriggerConfig`.

## Watching it

The two that matter most are `webhooks_backlog_depth` and `webhooks_backlog_age_seconds`. Every
other instrument is a rate or a latency, and none of them separates "delivering steadily" from
"delivering steadily while falling further behind".

Alert on any increase in `webhooks_deliveries_dead` — a dead dispatch is an event a subscriber
will never see.
