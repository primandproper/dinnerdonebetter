/*
Package recording writes the audit log entry and the data change event that every repository
write owes, as further statements in the transaction that performed it.

The pair is one helper rather than two blocks at each call site because the failure mode of two
blocks is omission, and omission is silent: a row nothing recorded has no provenance and the
tamper-evident chain cannot notice, because a chain records what it was given; a change nothing
emitted leaves the search index stale and no webhook fired. Neither leaves anything behind to
find later. One call cannot half-happen.

It is one type here rather than a method per repository because the body is the same body in
every one of them — nine copies of it, differing only in the name of the receiver — and a rule
stated nine times is a rule that can be restated wrongly once. What varies between repositories
is which tracer names the span, and that is a constructor argument.

# Why this is not in platform-go

Both halves are this application's vocabulary sitting on platform's engines. audit.AuditLogEntry
names a BelongsToUser and a BelongsToAccount because tenancy depth is an application's decision
and ours is two-level; the local audit.Recorder translates that to platform's Actor and Scope.
events.Emitter builds this application's DataChangeMessage from the context and hands it to
platform's outbox.Writer. A platform-side version would have to be generic over both types, and
the body inside those two interfaces is a call, a wrap, a call and a wrap — more interface for
the consumer to declare than implementation for platform to provide. See platform-go#285, which
closed the larger version of this question for related reasons.

What platform does own is the part that makes the transaction guarantee real rather than
conventional: outbox.Enqueue takes the database.Tx, so an event cannot be enqueued outside the
transaction that wrote its row even by mistake.
*/
package recording

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
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
// The emitter may be nil, which is inert and emits nothing — see events.NewEmitter. The audit
// repository may not: a process built without an outbox still owes the log an entry, which is
// why the two are separate fields here rather than one emitter that does both.
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
// The options are forwarded to Emit. Nothing passes one today, but the alternative to accepting
// them is that a write needing an ordering key has to reach past this method and spell both
// blocks itself — which is the shape this exists to remove.
//
// Reach past it for a bare Record when a write owes an entry and no event, or owes two entries,
// which Record's variadic form takes and this deliberately does not. See docs/audit.md.
func (r *Recorder) RecordAndEmit(
	ctx context.Context,
	tx database.Tx,
	logger logging.Logger,
	entry *audit.AuditLogEntry,
	eventType, accountID string,
	metadata map[string]any,
	opts ...events.EmitOption,
) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if err := r.auditLogEntryRepo.Record(ctx, tx, entry); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "creating audit log entry")
	}

	if err := r.events.Emit(ctx, tx, logger, eventType, accountID, metadata, opts...); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "enqueuing data change event")
	}

	return nil
}
