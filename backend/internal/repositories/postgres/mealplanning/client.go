package mealplanning

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/platform-go/v15/mediaregistry"
	"github.com/primandproper/platform-go/v15/outbox"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

const (
	o11yName = "meal_planning_db_client"
)

// repository is the meal planning repository implementation.
//
// Every write announces itself through platform's recording spine, on the write's own
// transaction: emitter for an event alone, recorder for an audit entry and the event describing
// the same write, and writer's EnqueueDerived for an index event alone. The one thing this
// application adds to them is the payload, which datachanges.Event builds.
type repository struct {
	database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	generatedQuerier  generated.Querier
	auditLogEntryRepo audit.Repository
	emitter           *webhooks.Emitter
	recorder          *platformrecording.Recorder
	writer            *outbox.Writer

	// uploads answers what a bridge row's uploaded_media_id names. The media
	// itself lives in platform-go's upload registry, whose table this repository's
	// statements cannot join, so it is read through the registry's batched read —
	// see GetUploadedMediaWithIDs.
	uploads mediaregistry.Store

	readDB  database.SQLQueryExecutor
	writeDB database.SQLQueryExecutor
}

// ProvideMealPlanningRepository provides a new repository.
//
// emitter, recorder and writer are platform's recording spine — see internal/recordingspine for
// how a process builds them. They are the same three in every process, and a write made through a
// repository built without them fails rather than writing a row nothing announced.
func ProvideMealPlanningRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	emitter *webhooks.Emitter,
	recorder *platformrecording.Recorder,
	writer *outbox.Writer,
	uploads mediaregistry.Store,
) mealplanning.Repository {
	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	c := &repository{
		Client:            client,
		readDB:            client.Reader(),
		writeDB:           client.Writer(),
		tracer:            tracer,
		generatedQuerier:  generated.New(),
		auditLogEntryRepo: auditLogEntryRepo,
		emitter:           emitter,
		recorder:          recorder,
		writer:            writer,
		uploads:           uploads,
		logger:            logging.NewNamedLogger(logger, o11yName),
	}

	return c
}

// withEvent runs a write and the data change event describing it in one transaction, so the
// event cannot survive a write that rolled back — nor be lost after one that committed.
//
// accountID is passed explicitly wherever the repository knows it; see datachanges.Event.
// The enumeration tables (valid ingredients, vessels, preparations, and friends) are global
// catalog data owned by no account, so they pass "".
//
// The search index event a write to an indexed table owes is not passed here: that obligation
// is registered on the outbox writer, and derived from this event. See internal/indexevents.
func (q *repository) withEvent(
	ctx context.Context,
	logger logging.Logger,
	eventType, accountID string,
	metadata map[string]any,
	write func(tx database.Tx) error,
) error {
	return q.WithTransaction(ctx, func(tx database.Tx) error {
		if err := write(tx); err != nil {
			return err
		}

		return q.emit(ctx, tx, logger, eventType, accountID, metadata)
	})
}

// emit publishes one data change event on tx and fans it out to the account's webhook
// subscribers, so it commits with whatever else tx did.
func (q *repository) emit(ctx context.Context, tx database.Tx, logger logging.Logger, eventType, accountID string, metadata map[string]any) error {
	event, msg := datachanges.Event(ctx, logger, eventType, accountID, metadata)

	return q.emitter.Emit(ctx, tx, datachanges.Scope(msg.AccountID), event)
}

// record writes entry to the audit log and publishes the event describing the same write, both
// on tx, through platform's Recorder. Who did it is the principal on the context.
//
// The entry is one audit.NewEntry built: its Scope is the chain this application's attribution
// rule chose for it, and the event fans out within that same scope.
func (q *repository) record(
	ctx context.Context,
	tx database.Tx,
	logger logging.Logger,
	entry *platformaudit.Entry,
	eventType, accountID string,
	metadata map[string]any,
) error {
	event, msg := datachanges.Event(ctx, logger, eventType, accountID, metadata)

	scope := entry.Scope
	if scope == (tenancy.Scope{}) {
		scope = datachanges.Scope(msg.AccountID)
	}

	return q.recorder.Record(ctx, tx, scope, event, &platformrecording.Entry{
		ResourceType: entry.ResourceType,
		ResourceID:   entry.ResourceID,
		EventType:    entry.EventType,
		Changes:      entry.Changes,
		Metadata:     entry.Metadata,
	})
}

// emitIndex enqueues the index events a trigger implies, without announcing anything.
//
// emit is the usual path, because a write worth indexing is nearly always a write worth
// announcing, and the side effect derives the index event from the announcement. This exists for
// the writes where that is not true — where putting the event on the wire would be a decision
// about the public event stream rather than about the index.
//
// It is the writer's EnqueueDerived over the same shape of message every announced write sends,
// so the rules it is matched against are the ones registered on the writer, not a copy kept
// here. The message itself is never enqueued; only what the side effects derive from it is.
func (q *repository) emitIndex(ctx context.Context, tx database.Tx, trigger string, metadata map[string]any) error {
	return q.writer.EnqueueDerived(ctx, tx, outbox.Message{
		Payload: &datachanges.Message{EventType: trigger, Context: metadata},
	})
}
