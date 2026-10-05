package webhooksstore

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbwebhooks "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	webhookkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformwebhooks "github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

const (
	// resourceTypeWebhooks is what an audit entry about an endpoint names.
	//
	// It keeps the old table's name rather than taking the new one's, because an
	// audit log is read across the change: an investigation asking what happened
	// to a webhook should find the entries written before the store moved as well
	// as the ones after.
	resourceTypeWebhooks = "webhooks"
	// resourceTypeWebhookTriggerConfigs is what an entry about a subscription names.
	resourceTypeWebhookTriggerConfigs = "webhook_trigger_configs"
)

// hooks records every endpoint and subscription write.
//
// It implements Hooks outright rather than embedding NoopHooks, so a write
// platform adds later breaks this build until somebody decides what it records.
// Embedded, the new write would compile and record nothing, which is the one
// failure an audit log cannot notice.
//
// The delivery machinery — Claim, MarkDelivered, RecordFailure, Reap — has no
// hooks and owes no entry: nobody investigates a delivery attempt through the
// audit log, and the attempts table is the record of those.
type hooks struct {
	tracer   tracing.Tracer
	logger   logging.Logger
	recorder *recording.Recorder
}

var _ platformwebhooks.Hooks = (*hooks)(nil)

// AfterSaveEndpoint records an endpoint being registered or re-registered, and on
// a re-registration which of its fields moved.
//
// One entry for both halves of a save, because the store's one method is both,
// and the entry says which by the event type it carries. Which it was is the
// store's answer — the row it had to look for before the upsert — and not the
// argument's: the caller is webhooks.StoreDispatcher.Register, which mints the
// id before the store sees it, so "arrived with no id" would make every
// registration an update.
func (h *hooks) AfterSaveEndpoint(ctx context.Context, tx database.Tx, scope tenancy.Scope, before, after *platformwebhooks.Endpoint) error {
	if after.Created {
		return h.record(ctx, tx, scope, after.ID, resourceTypeWebhooks,
			platformaudit.EventCreated, ddbwebhooks.WebhookCreatedServiceEventType,
			webhookkeys.WebhookIDKey, nil)
	}

	// The endpoint's Secret is tagged json:"-", so the diff never carries a key.
	changes, err := platformaudit.Diff(before, after)
	if err != nil {
		return platformerrors.Wrap(err, "diffing the re-registered webhook")
	}

	return h.record(ctx, tx, scope, after.ID, resourceTypeWebhooks,
		platformaudit.EventUpdated, ddbwebhooks.WebhookUpdatedServiceEventType,
		webhookkeys.WebhookIDKey, changes)
}

// AfterArchiveEndpoint records an endpoint being retired.
func (h *hooks) AfterArchiveEndpoint(ctx context.Context, tx database.Tx, scope tenancy.Scope, endpoint *platformwebhooks.Endpoint) error {
	return h.record(ctx, tx, scope, endpoint.ID, resourceTypeWebhooks,
		platformaudit.EventArchived, ddbwebhooks.WebhookArchivedServiceEventType,
		webhookkeys.WebhookIDKey, nil)
}

// AfterRotateSecret records that an endpoint's signing key changed.
//
// The entry names no secret and neither does the event. What is worth recording
// is that the key changed and who changed it; the value is the one thing in this
// schema that is written to sign with and never read back out.
func (h *hooks) AfterRotateSecret(ctx context.Context, tx database.Tx, scope tenancy.Scope, endpointID string) error {
	return h.record(ctx, tx, scope, endpointID, resourceTypeWebhooks,
		platformaudit.EventUpdated, ddbwebhooks.WebhookSecretRotatedServiceEventType,
		webhookkeys.WebhookIDKey, nil)
}

// AfterAddSubscription records an endpoint being subscribed to an event type.
//
// It is recorded as a creation whether the store created the row, revived an
// archived one, or found it already live, which is what this application has
// always recorded: from where a subscriber stands, each of the three is "I
// subscribed".
func (h *hooks) AfterAddSubscription(ctx context.Context, tx database.Tx, scope tenancy.Scope, _, after *platformwebhooks.Subscription) error {
	return h.record(ctx, tx, scope, after.ID, resourceTypeWebhookTriggerConfigs,
		platformaudit.EventCreated, ddbwebhooks.WebhookTriggerConfigCreatedServiceEventType,
		webhookkeys.WebhookTriggerConfigIDKey, nil)
}

// AfterArchiveSubscription records an endpoint being unsubscribed.
func (h *hooks) AfterArchiveSubscription(ctx context.Context, tx database.Tx, scope tenancy.Scope, subscription *platformwebhooks.Subscription) error {
	return h.record(ctx, tx, scope, subscription.ID, resourceTypeWebhookTriggerConfigs,
		platformaudit.EventArchived, ddbwebhooks.WebhookTriggerConfigArchivedServiceEventType,
		webhookkeys.WebhookTriggerConfigIDKey, nil)
}

// record writes the audit entry and enqueues the data change event, on the
// transaction the write ran in. changes is the field-level diff a
// re-registration carries, and nil for every other write.
//
// The account comes off the scope, which is where an endpoint's tenant lives:
// this application files every endpoint under the account that owns it, so the
// scope's owner is the account an event reaches subscribers under.
func (h *hooks) record(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	relevantID, resourceType string, auditEventType platformaudit.EventType, changeEventType, logKey string,
	changes map[string]platformaudit.Change,
) error {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, logKey, relevantID)

	accountID := scope.Owner()
	logger := h.logger.WithSpan(span).WithValue(logKey, relevantID)

	entry := audit.NewEntry("", accountID, resourceType, relevantID, auditEventType)
	entry.Changes = changes

	return h.recorder.RecordAndEmit(ctx, tx, logger, entry, changeEventType, accountID, map[string]any{
		logKey: relevantID,
	})
}
