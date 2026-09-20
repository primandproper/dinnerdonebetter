/*
Package webhooksstore is platform-go's webhook store with this application's
recording around it.

It replaces internal/repositories/postgres/webhooks, which owned three tables of
its own — webhooks, webhook_trigger_configs and webhook_trigger_events — and ran
them in parallel with platform's endpoints, joined by a shared identifier. The
endpoints were the delivery machinery and the local rows were the model; now
there is one model, and it is platform's.

One field did not survive, and it was already inert. The local webhooks row
carried a method column, validated on every write and echoed back on every read,
while delivery went through platform's dispatcher — which posts. A client that
asked for PUT was told PUT and sent POST. Dropping the column removes the lie
rather than a capability.

What is not platform's is the audit entry and the data change event every write
owes, which is what lives here.
*/
package webhooksstore

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbwebhooks "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	webhookkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformwebhooks "github.com/primandproper/platform-go/v14/webhooks"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

const (
	o11yName = "webhooks_db_client"

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

// store is platform's webhook store with this application's recording around its
// writes.
//
// The reads are embedded. Twelve of the nineteen methods add nothing, and the
// worker path — Claim, MarkDelivered, RecordFailure, Reap — is the dispatcher's
// own and owes no entry: nobody investigates a delivery attempt through the
// audit log, and the attempts table is the record of those.
type store struct {
	platformwebhooks.Store

	tracer   tracing.Tracer
	logger   logging.Logger
	recorder *recording.Recorder
}

var _ platformwebhooks.Store = (*store)(nil)

// ProvideStore wraps a platform webhook store in this application's recording.
func ProvideStore(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	auditLogEntryRepo audit.Repository,
	eventEmitter *events.Emitter,
	inner platformwebhooks.Store,
) (platformwebhooks.Store, error) {
	if inner == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil webhook store")
	}

	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	return &store{
		Store:    inner,
		tracer:   tracer,
		logger:   logging.NewNamedLogger(logger, o11yName),
		recorder: recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
	}, nil
}

// SaveEndpoint writes the endpoint, then records it.
//
// One entry for both halves of a save, because the store's one method is both:
// an endpoint with an id it has seen is an update and one without is a
// creation, and the entry says which by the event type it carries.
func (s *store) SaveEndpoint(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	endpoint *platformwebhooks.Endpoint,
) (*platformwebhooks.Endpoint, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	existing := endpoint != nil && endpoint.ID != ""

	saved, err := s.Store.SaveEndpoint(ctx, tx, scope, endpoint)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, saved.ID)

	eventType := ddbwebhooks.WebhookCreatedServiceEventType
	auditEventType := audit.AuditLogEventTypeCreated
	if existing {
		eventType = ddbwebhooks.WebhookUpdatedServiceEventType
		auditEventType = audit.AuditLogEventTypeUpdated
	}

	if err = s.record(ctx, tx, scope, saved.ID, resourceTypeWebhooks, auditEventType, eventType,
		webhookkeys.WebhookIDKey); err != nil {
		return nil, err
	}

	return saved, nil
}

// ArchiveEndpoint retires the endpoint, then records it.
func (s *store) ArchiveEndpoint(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	endpointID string,
) (*platformwebhooks.Endpoint, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	archived, err := s.Store.ArchiveEndpoint(ctx, tx, scope, endpointID)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, endpointID)

	if err = s.record(ctx, tx, scope, endpointID, resourceTypeWebhooks,
		audit.AuditLogEventTypeArchived, ddbwebhooks.WebhookArchivedServiceEventType,
		webhookkeys.WebhookIDKey); err != nil {
		return nil, err
	}

	return archived, nil
}

// AddSubscription subscribes the endpoint to an event type, then records it.
func (s *store) AddSubscription(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	endpointID string,
	eventType platformwebhooks.EventType,
) (*platformwebhooks.Subscription, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	added, err := s.Store.AddSubscription(ctx, tx, scope, endpointID, eventType)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, webhookkeys.WebhookTriggerConfigIDKey, added.ID)

	if err = s.record(ctx, tx, scope, added.ID, resourceTypeWebhookTriggerConfigs,
		audit.AuditLogEventTypeCreated, ddbwebhooks.WebhookTriggerConfigCreatedServiceEventType,
		webhookkeys.WebhookTriggerConfigIDKey); err != nil {
		return nil, err
	}

	return added, nil
}

// ArchiveSubscription unsubscribes the endpoint, then records it.
func (s *store) ArchiveSubscription(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	subscriptionID string,
) (*platformwebhooks.Subscription, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	archived, err := s.Store.ArchiveSubscription(ctx, tx, scope, subscriptionID)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, webhookkeys.WebhookTriggerConfigIDKey, subscriptionID)

	if err = s.record(ctx, tx, scope, subscriptionID, resourceTypeWebhookTriggerConfigs,
		audit.AuditLogEventTypeArchived, ddbwebhooks.WebhookTriggerConfigArchivedServiceEventType,
		webhookkeys.WebhookTriggerConfigIDKey); err != nil {
		return nil, err
	}

	return archived, nil
}

// RotateSecret mints a new signing secret, then records that it happened.
//
// The entry names no secret and neither does the event. What is worth recording
// is that the key changed and who changed it; the value is the one thing in this
// schema that is written to sign with and never read back out.
func (s *store) RotateSecret(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	endpointID string,
	next []byte,
) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if err := s.Store.RotateSecret(ctx, tx, scope, endpointID, next); err != nil {
		return err
	}

	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, endpointID)

	return s.record(ctx, tx, scope, endpointID, resourceTypeWebhooks,
		audit.AuditLogEventTypeUpdated, ddbwebhooks.WebhookSecretRotatedServiceEventType,
		webhookkeys.WebhookIDKey)
}

// record writes the audit entry and enqueues the data change event, inside the
// caller's transaction.
//
// The account comes off the scope, which is where an endpoint's tenant lives:
// this application files every endpoint under the account that owns it, so the
// scope's owner is the account an event reaches subscribers under.
func (s *store) record(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	relevantID, resourceType, auditEventType, changeEventType, logKey string,
) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	accountID := scope.Owner()
	logger := s.logger.WithSpan(span).WithValue(logKey, relevantID)

	entry := &audit.AuditLogEntry{
		ID:           identifiers.New(),
		ResourceType: resourceType,
		RelevantID:   relevantID,
		EventType:    auditEventType,
	}
	if accountID != "" {
		entry.BelongsToAccount = &accountID
	}

	return s.recorder.RecordAndEmit(ctx, tx, logger, entry, changeEventType, accountID, map[string]any{
		logKey: relevantID,
	})
}
