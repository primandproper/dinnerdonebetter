/*
Package recording writes the audit log entry and the data change event that every repository
write owes, as further statements in the transaction that performed it.

The pair is one helper rather than two blocks at each call site because the failure mode of two
blocks is omission, and omission is silent: a row nothing recorded has no provenance and the
tamper-evident chain cannot notice, because a chain records what it was given; a change nothing
emitted leaves the search index stale and no webhook fired. Neither leaves anything behind to
find later. One call cannot half-happen.

# What is left of it

The pair itself is platform's now: recording.Recorder.Record writes the entries and the event on
one transaction, in one order, with the actor read off the context. What remains here is the
translation from this application's spelling of a write — an audit.NewEntry naming a user and an
account, an event type string, a context map — onto platform's Record, which events.Emitter
performs. This type keeps the call sites' signature while they move, and goes with the last of
them.

The one thing it still decides is what a process with no outbox does: it writes the entry. A
repository can be built without an emitter — the one-shot tools, a unit test — and the log is
owed its entry regardless, which is why the audit repository is a field here rather than
something reached through the emitter.
*/
package recording

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// Recorder writes the audit entry and the data change event for one write.
type Recorder struct {
	_                 struct{} `json:"-"`
	tracer            tracing.Tracer
	auditLogEntryRepo audit.Repository
	events            *events.Emitter
}

// NewRecorder builds a Recorder for one repository.
//
// The tracer is the repository's own named tracer rather than one this package builds, so a
// span raised here is attributed to the package whose write raised it: recording is where the
// code lives, not where the work belongs.
//
// The emitter may be nil, in which case a write records its entry and announces nothing. The
// audit repository may not: a process built without an outbox still owes the log an entry.
func NewRecorder(tracer tracing.Tracer, auditLogEntryRepo audit.Repository, emitter *events.Emitter) *Recorder {
	return &Recorder{
		tracer:            tracer,
		auditLogEntryRepo: auditLogEntryRepo,
		events:            emitter,
	}
}

// RecordAndEmit writes entry to the audit log and enqueues one data change event, both using
// the caller's transaction, so they commit with the rows they describe or not at all.
//
// accountID overrides the account the event is attributed to and should be passed wherever the
// repository knows it, because a background job has no session to read one from; pass "" when
// the event genuinely has no account. See events.Emitter.Emit.
//
// The options are forwarded to the emitter. Nothing passes one today, but the alternative to
// accepting them is that a write needing an ordering key has to reach past this method and spell
// both halves itself — which is the shape this exists to remove.
//
// Reach past it for a bare Record when a write owes an entry and no event, or owes two entries,
// which Record's variadic form takes and this deliberately does not. See docs/audit.md.
func (r *Recorder) RecordAndEmit(
	ctx context.Context,
	tx database.Tx,
	logger logging.Logger,
	entry *platformaudit.Entry,
	eventType, accountID string,
	metadata map[string]any,
	opts ...events.EmitOption,
) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if r.events == nil {
		if err := r.auditLogEntryRepo.Record(ctx, tx, entry); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "creating audit log entry")
		}

		return nil
	}

	if err := r.events.Record(ctx, tx, logger, entry, eventType, accountID, metadata, opts...); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "recording the write")
	}

	return nil
}
