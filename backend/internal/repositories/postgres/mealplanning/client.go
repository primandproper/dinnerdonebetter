package mealplanning

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/mediaregistry"
	"github.com/primandproper/platform-go/v15/outbox"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "meal_planning_db_client"
)

// repository is the meal planning repository implementation.
//
// Every write announces itself through platform's recording spine, on the write's own
// transaction: emitter for an event alone (withEvent, emit), recorder for audit entries and the
// event describing the same write (withRecord, record), and writer's EnqueueDerived for an index
// event alone (emitIndex). The one thing this application adds to them is the payload, which
// datachanges.Event builds.
//
// TestEveryWriteIsRecorded holds every exported write to that, and names the few that are
// bookkeeping rather than change — and so deliberately announce nothing — with the reason.
type repository struct {
	database.Client
	tracer           tracing.Tracer
	logger           logging.Logger
	generatedQuerier generated.Querier
	roster           platformidentity.DirectoryReader
	emitter          *webhooks.Emitter
	recorder         *platformrecording.Recorder
	writer           *outbox.Writer

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
// repository built without them fails rather than writing a row nothing announced. The writes
// exempt from announcing are named, with their reasons, in TestEveryWriteIsRecorded.
func ProvideMealPlanningRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	roster platformidentity.DirectoryReader,
	client database.Client,
	emitter *webhooks.Emitter,
	recorder *platformrecording.Recorder,
	writer *outbox.Writer,
	uploads mediaregistry.Store,
) mealplanning.Repository {
	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	c := &repository{
		Client:           client,
		readDB:           client.Reader(),
		writeDB:          client.Writer(),
		tracer:           tracer,
		generatedQuerier: generated.New(),
		roster:           roster,
		emitter:          emitter,
		recorder:         recorder,
		writer:           writer,
		uploads:          uploads,
		logger:           logging.NewNamedLogger(logger, o11yName),
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

// withRecord is withEvent for the writes that are audited as well as announced: the audit entry,
// the event and the write share one transaction, so none of the three survives without the others.
func (q *repository) withRecord(
	ctx context.Context,
	logger logging.Logger,
	entry *platformrecording.Entry,
	eventType, accountID string,
	metadata map[string]any,
	write func(tx database.Tx) error,
) error {
	return q.WithTransaction(ctx, func(tx database.Tx) error {
		if err := write(tx); err != nil {
			return err
		}

		return q.record(ctx, tx, logger, eventType, accountID, metadata, entry)
	})
}

// record writes entries to the audit log and publishes the event describing the same write, all
// on tx, through platform's Recorder. Who did it is the principal on the context.
//
// The event and the entries share one scope, because platform's Recorder takes one: the chain
// audit.ScopeFor files the event's account and actor under, so an entry lands in the account
// whose subscribers hear about it, and in the actor's own chain when the write happened in no
// account. accountID is resolved exactly as emit resolves it; see datachanges.Event.
func (q *repository) record(
	ctx context.Context,
	tx database.Tx,
	logger logging.Logger,
	eventType, accountID string,
	metadata map[string]any,
	entries ...*platformrecording.Entry,
) error {
	event, msg := datachanges.Event(ctx, logger, eventType, accountID, metadata)

	return q.recorder.Record(ctx, tx, audit.ScopeFor(msg.AccountID, msg.UserID), event, entries...)
}

// auditEntry is the caller-supplied half of an audit entry, which is all platform's Recorder
// takes: the actor comes off the context, and the scope off the event; see record.
func auditEntry(resourceType, resourceID string, eventType platformaudit.EventType) *platformrecording.Entry {
	return &platformrecording.Entry{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		EventType:    eventType,
	}
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
