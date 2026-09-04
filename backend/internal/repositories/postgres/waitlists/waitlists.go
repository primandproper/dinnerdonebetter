/*
Package waitlists records what a waitlist write means to the rest of this
application. The lists and the signups themselves are platform-go's: the schema,
the paging, the tenancy column, the lifecycle and the withdrawal all live there,
and this package neither reimplements nor wraps them.

What it adds is the half platform cannot know about — an audit log entry naming
who did what, and a data change event on the outbox that the webhook dispatcher
fans out. Every event this emits is in the webhook event catalog
(internal/domain/webhooks/catalog), so a subscriber can already ask for them; a
write that skipped the pair would be a row with no provenance and a subscriber
that never heard.

# The transaction the events are not in

Every hand-written repository here emits inside the transaction that wrote the
row, so the event lives or dies with what it describes (see
internal/repositories/postgres/events). This one cannot: platform's writes own
their transactions and take no executor, so the audit entry and the event are a
second transaction after the first has committed.

The gap that opens is the ordinary one — the row lands, the process dies, and
nothing is recorded about it. It is narrow and it is one-directional: a signup
can exist with no event, but no event can name a signup that was not written.
Closing it needs platform's write methods to accept a database.Tx. That is the
same gap comments has (platform-go #457), and it is filed for this package as
platform-go #458 rather than worked around here — a gap papered over locally
stops being a gap anyone remembers.

# Why a withdrawal is recorded apart from a transition

Invite and Convert move somebody through a queue; Withdraw is a standing
instruction to stop writing to an address. They emit different events for that
reason, and the withdrawal's audit entry is written from a read taken before the
store runs, because a withdrawal erases the subject reference the entry has to
name. Read afterwards, every withdrawal in the audit log would belong to nobody.
*/
package waitlists

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbwaitlists "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	waitlistkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/keys"

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/tenancy"
	platformwaitlists "github.com/primandproper/platform-go/v13/waitlists"
)

const (
	// resourceTypeWaitlists is what an audit entry about a list names.
	resourceTypeWaitlists = "waitlists"
	// resourceTypeWaitlistSignups is what an audit entry about a signup names.
	resourceTypeWaitlistSignups = "waitlist_signups"
)

var _ platformwaitlists.Store = (*repository)(nil)

// CreateList opens the waitlist, then records it.
func (r *repository) CreateList(ctx context.Context, scope tenancy.Scope, list *platformwaitlists.List) (*platformwaitlists.List, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	created, err := r.Store.CreateList(ctx, scope, list)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, created.ID)

	if err = r.recordList(ctx, created.ID, audit.AuditLogEventTypeCreated, ddbwaitlists.WaitlistCreatedServiceEventType); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateList rewrites the list, then records it.
func (r *repository) UpdateList(ctx context.Context, scope tenancy.Scope, list *platformwaitlists.List) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if err := r.Store.UpdateList(ctx, scope, list); err != nil {
		return err
	}

	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, list.ID)

	return r.recordList(ctx, list.ID, audit.AuditLogEventTypeUpdated, ddbwaitlists.WaitlistUpdatedServiceEventType)
}

// ArchiveList retires the list, then records it.
func (r *repository) ArchiveList(ctx context.Context, scope tenancy.Scope, listID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, listID)

	if err := r.Store.ArchiveList(ctx, scope, listID); err != nil {
		return err
	}

	return r.recordList(ctx, listID, audit.AuditLogEventTypeArchived, ddbwaitlists.WaitlistArchivedServiceEventType)
}

// Join adds somebody to the list, then records it.
func (r *repository) Join(ctx context.Context, scope tenancy.Scope, listID string, signup *platformwaitlists.Signup) (*platformwaitlists.Signup, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, listID)

	joined, err := r.Store.Join(ctx, scope, listID, signup)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupIDKey, joined.ID)

	if err = r.recordSignup(ctx, joined, audit.AuditLogEventTypeCreated, ddbwaitlists.WaitlistSignupCreatedServiceEventType); err != nil {
		return nil, err
	}

	return joined, nil
}

// UpdateSignupNotes rewrites the operator's note, then records it.
//
// It records an update rather than a transition, which is the distinction the
// method exists to make: a note is the one write that touches a signup without
// moving anybody.
func (r *repository) UpdateSignupNotes(ctx context.Context, scope tenancy.Scope, listID, signupID, notes string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	signup, err := r.readSignupToRecord(ctx, span, scope, listID, signupID)
	if err != nil {
		return err
	}

	if err = r.Store.UpdateSignupNotes(ctx, scope, listID, signupID, notes); err != nil {
		return err
	}

	return r.recordSignup(ctx, signup, audit.AuditLogEventTypeUpdated, ddbwaitlists.WaitlistSignupUpdatedServiceEventType)
}

// Invite lets somebody in, then records the move.
func (r *repository) Invite(ctx context.Context, scope tenancy.Scope, listID, signupID string) error {
	return r.recordedTransition(ctx, scope, listID, signupID, platformwaitlists.StatusInvited, r.Store.Invite)
}

// Convert marks an invitation taken up, then records the move.
func (r *repository) Convert(ctx context.Context, scope tenancy.Scope, listID, signupID string) error {
	return r.recordedTransition(ctx, scope, listID, signupID, platformwaitlists.StatusConverted, r.Store.Convert)
}

// Withdraw takes somebody off the list at their own request, then records it.
//
// The signup is read before the store runs, because a withdrawal blanks the
// subject reference — and an audit entry whose actor is empty is an entry nobody
// can find when they ask who came off which list.
func (r *repository) Withdraw(ctx context.Context, scope tenancy.Scope, listID, signupID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	signup, err := r.readSignupToRecord(ctx, span, scope, listID, signupID)
	if err != nil {
		return err
	}

	if err = r.Store.Withdraw(ctx, scope, listID, signupID); err != nil {
		return err
	}

	signup.Status = platformwaitlists.StatusWithdrawn

	return r.recordSignup(ctx, signup, audit.AuditLogEventTypeUpdated, ddbwaitlists.WaitlistSignupWithdrawnServiceEventType)
}

// ArchiveSignup retires the signup administratively, then records it.
func (r *repository) ArchiveSignup(ctx context.Context, scope tenancy.Scope, listID, signupID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	signup, err := r.readSignupToRecord(ctx, span, scope, listID, signupID)
	if err != nil {
		return err
	}

	if err = r.Store.ArchiveSignup(ctx, scope, listID, signupID); err != nil {
		return err
	}

	return r.recordSignup(ctx, signup, audit.AuditLogEventTypeArchived, ddbwaitlists.WaitlistSignupArchivedServiceEventType)
}

// recordedTransition is Invite and Convert: read whose signup it is, move it,
// and record the move under the status it landed in.
//
// The status is the one the guard required rather than one read back, because
// the guard is what decided it: a transition that reported no error moved the
// row to exactly this status, and a second read could only disagree by
// describing somebody else's later write.
func (r *repository) recordedTransition(
	ctx context.Context,
	scope tenancy.Scope,
	listID, signupID string,
	to platformwaitlists.Status,
	move func(ctx context.Context, scope tenancy.Scope, listID, signupID string) error,
) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupStatusKey, to.String())

	signup, err := r.readSignupToRecord(ctx, span, scope, listID, signupID)
	if err != nil {
		return err
	}

	if err = move(ctx, scope, listID, signupID); err != nil {
		return err
	}

	signup.Status = to

	return r.recordSignup(ctx, signup, audit.AuditLogEventTypeUpdated, ddbwaitlists.WaitlistSignupTransitionedServiceEventType)
}

// readSignupToRecord fetches the signup a write is about to change, so the entry
// that describes it can name whose it was.
//
// A read that fails is the write's failure too: platform answers an absent,
// archived, or other-scope signup as ErrSignupNotFound either way, so returning
// it from here is the same answer one call earlier.
func (r *repository) readSignupToRecord(
	ctx context.Context,
	span tracing.Span,
	scope tenancy.Scope,
	listID, signupID string,
) (*platformwaitlists.Signup, error) {
	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, listID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupIDKey, signupID)

	signup, err := r.GetSignup(ctx, scope, listID, signupID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching waitlist signup to record")
	}

	return signup, nil
}

// recordList writes the audit entry and the data change event for a write to the
// catalog.
//
// The entry names no user, because a list is an administrative row that belongs
// to nobody: who opened it is the actor on the context, which is what the audit
// recorder resolves. That is the same shape the table this replaced recorded
// under.
func (r *repository) recordList(ctx context.Context, listID, auditEventType, changeEventType string) error {
	return r.record(ctx, "", resourceTypeWaitlists, listID, auditEventType, changeEventType, map[string]any{
		waitlistkeys.WaitlistIDKey: listID,
	})
}

// recordSignup writes the audit entry and the data change event for a write to
// one signup.
//
// The entry belongs to the signup's subject rather than to whoever made the
// request, so that "which lists was this person on, and what happened to them"
// is answerable from the audit log after the row itself has been withdrawn and
// no longer says.
func (r *repository) recordSignup(ctx context.Context, signup *platformwaitlists.Signup, auditEventType, changeEventType string) error {
	return r.record(ctx, signup.Subject.ID, resourceTypeWaitlistSignups, signup.ID, auditEventType, changeEventType, map[string]any{
		waitlistkeys.WaitlistSignupIDKey:     signup.ID,
		waitlistkeys.WaitlistIDKey:           signup.ListID,
		waitlistkeys.WaitlistSignupStatusKey: signup.Status.String(),
	})
}

// record writes the audit entry and enqueues the data change event, in one
// transaction of their own.
//
// The two travel together because they answer the same question from opposite
// sides — the audit log for whoever asks later who did this, the outbox for
// whoever needs to know now — and a write that carried one without the other
// would be a write nobody could tell was incomplete.
//
// The event names no account, which is what makes it a service-wide event: these
// rows live in the global scope, so the account it reaches a webhook subscriber
// under is whichever one the requester had active, resolved from the context by
// the emitter.
func (r *repository) record(
	ctx context.Context,
	userID, resourceType, relevantID, auditEventType, changeEventType string,
	metadata map[string]any,
) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, relevantID)

	return r.client.WithTransaction(ctx, func(tx database.Tx) error {
		return r.recorder.RecordAndEmit(ctx, tx, logger, &audit.AuditLogEntry{
			ID:            identifiers.New(),
			ResourceType:  resourceType,
			RelevantID:    relevantID,
			EventType:     auditEventType,
			BelongsToUser: userID,
		}, changeEventType, "", metadata)
	})
}
