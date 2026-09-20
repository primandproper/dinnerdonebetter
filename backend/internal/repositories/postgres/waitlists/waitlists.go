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

# The transaction the events are in

Every hand-written repository here emits inside the transaction that wrote
the row, so the event lives or dies with what it describes (see
internal/repositories/postgres/events). This one now does too. It could not
before: platform's writes owned their transactions and took no executor, so
the audit entry and the event were a second transaction after the first had
committed, and a signup could exist that nothing had recorded.

As of platform-go v14 a store write takes the caller's database.Tx, so the
write, the entry and the event are one transaction and share one fate. The
gap filed for this package as platform-go #458 is closed by that convention
rather than by anything here, which is why this package has no workaround to
delete.

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

	platformwaitlists "github.com/primandproper/platform-go/v14/waitlists"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

const (
	// resourceTypeWaitlists is what an audit entry about a list names.
	resourceTypeWaitlists = "waitlists"
	// resourceTypeWaitlistSignups is what an audit entry about a signup names.
	resourceTypeWaitlistSignups = "waitlist_signups"
)

var _ platformwaitlists.Store = (*repository)(nil)

// CreateList opens the waitlist, then records it.
func (r *repository) CreateList(ctx context.Context, tx database.Tx, scope tenancy.Scope, list *platformwaitlists.List) (*platformwaitlists.List, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	created, err := r.Store.CreateList(ctx, tx, scope, list)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, created.ID)

	if err = r.recordList(ctx, tx, created.ID, audit.AuditLogEventTypeCreated, ddbwaitlists.WaitlistCreatedServiceEventType); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateList rewrites the list, then records it.
func (r *repository) UpdateList(ctx context.Context, tx database.Tx, scope tenancy.Scope, list *platformwaitlists.List) (*platformwaitlists.List, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	result, err := r.Store.UpdateList(ctx, tx, scope, list)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, list.ID)

	if err = r.recordList(ctx, tx, list.ID, audit.AuditLogEventTypeUpdated, ddbwaitlists.WaitlistUpdatedServiceEventType); err != nil {
		return nil, err
	}

	return result, nil
}

// ArchiveList retires the list, then records it.
func (r *repository) ArchiveList(ctx context.Context, tx database.Tx, scope tenancy.Scope, listID string) (*platformwaitlists.List, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, listID)

	result, err := r.Store.ArchiveList(ctx, tx, scope, listID)
	if err != nil {
		return nil, err
	}

	if err = r.recordList(ctx, tx, listID, audit.AuditLogEventTypeArchived, ddbwaitlists.WaitlistArchivedServiceEventType); err != nil {
		return nil, err
	}

	return result, nil
}

// Join adds somebody to the list, then records it.
func (r *repository) Join(ctx context.Context, tx database.Tx, scope tenancy.Scope, listID string, signup *platformwaitlists.Signup) (*platformwaitlists.Signup, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, listID)

	joined, err := r.Store.Join(ctx, tx, scope, listID, signup)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupIDKey, joined.ID)

	if err = r.recordSignup(ctx, tx, joined, audit.AuditLogEventTypeCreated, ddbwaitlists.WaitlistSignupCreatedServiceEventType); err != nil {
		return nil, err
	}

	return joined, nil
}

// UpdateSignupNotes rewrites the operator's note, then records it.
//
// It records an update rather than a transition, which is the distinction the
// method exists to make: a note is the one write that touches a signup without
// moving anybody.
func (r *repository) UpdateSignupNotes(ctx context.Context, tx database.Tx, scope tenancy.Scope, listID, signupID, notes string) (*platformwaitlists.Signup, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	signup, err := r.readSignupToRecord(ctx, tx, span, scope, listID, signupID)
	if err != nil {
		return nil, err
	}

	result, err := r.Store.UpdateSignupNotes(ctx, tx, scope, listID, signupID, notes)
	if err != nil {
		return nil, err
	}

	if err = r.recordSignup(ctx, tx, signup, audit.AuditLogEventTypeUpdated, ddbwaitlists.WaitlistSignupUpdatedServiceEventType); err != nil {
		return nil, err
	}

	return result, nil
}

// Invite lets somebody in, then records the move.
func (r *repository) Invite(ctx context.Context, tx database.Tx, scope tenancy.Scope, listID, signupID string) (*platformwaitlists.Signup, error) {
	return r.recordedTransition(ctx, tx, scope, listID, signupID, platformwaitlists.StatusInvited, r.Store.Invite)
}

// Convert marks an invitation taken up, then records the move.
func (r *repository) Convert(ctx context.Context, tx database.Tx, scope tenancy.Scope, listID, signupID string) (*platformwaitlists.Signup, error) {
	return r.recordedTransition(ctx, tx, scope, listID, signupID, platformwaitlists.StatusConverted, r.Store.Convert)
}

// Withdraw takes somebody off the list at their own request, then records it.
//
// The signup is read before the store runs, because a withdrawal blanks the
// subject reference — and an audit entry whose actor is empty is an entry nobody
// can find when they ask who came off which list.
func (r *repository) Withdraw(ctx context.Context, tx database.Tx, scope tenancy.Scope, listID, signupID string) (*platformwaitlists.Signup, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	signup, err := r.readSignupToRecord(ctx, tx, span, scope, listID, signupID)
	if err != nil {
		return nil, err
	}

	result, err := r.Store.Withdraw(ctx, tx, scope, listID, signupID)
	if err != nil {
		return nil, err
	}

	signup.Status = platformwaitlists.StatusWithdrawn

	if err = r.recordSignup(ctx, tx, signup, audit.AuditLogEventTypeUpdated, ddbwaitlists.WaitlistSignupWithdrawnServiceEventType); err != nil {
		return nil, err
	}

	return result, nil
}

// ArchiveSignup retires the signup administratively, then records it.
func (r *repository) ArchiveSignup(ctx context.Context, tx database.Tx, scope tenancy.Scope, listID, signupID string) (*platformwaitlists.Signup, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	signup, err := r.readSignupToRecord(ctx, tx, span, scope, listID, signupID)
	if err != nil {
		return nil, err
	}

	result, err := r.Store.ArchiveSignup(ctx, tx, scope, listID, signupID)
	if err != nil {
		return nil, err
	}

	if err = r.recordSignup(ctx, tx, signup, audit.AuditLogEventTypeArchived, ddbwaitlists.WaitlistSignupArchivedServiceEventType); err != nil {
		return nil, err
	}

	return result, nil
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
	tx database.Tx,
	scope tenancy.Scope,
	listID, signupID string,
	to platformwaitlists.Status,
	move func(ctx context.Context, tx database.Tx, scope tenancy.Scope, listID, signupID string) (*platformwaitlists.Signup, error),
) (*platformwaitlists.Signup, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupStatusKey, to.String())

	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, listID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupIDKey, signupID)

	// No read before the move. v14's transitions return the row they wrote, so
	// the entry names the signup as it now stands rather than as it was a
	// statement ago — and the read that used to establish that is gone with the
	// window it opened.
	moved, err := move(ctx, tx, scope, listID, signupID)
	if err != nil {
		return nil, err
	}

	if err = r.recordSignup(ctx, tx, moved, audit.AuditLogEventTypeUpdated, ddbwaitlists.WaitlistSignupTransitionedServiceEventType); err != nil {
		return nil, err
	}

	return moved, nil
}

// readSignupToRecord fetches the signup a write is about to change, so the entry
// that describes it can name whose it was.
//
// A read that fails is the write's failure too: platform answers an absent,
// archived, or other-scope signup as ErrSignupNotFound either way, so returning
// it from here is the same answer one call earlier.
func (r *repository) readSignupToRecord(
	ctx context.Context,
	tx database.Tx,
	span tracing.Span,
	scope tenancy.Scope,
	listID, signupID string,
) (*platformwaitlists.Signup, error) {
	tracing.AttachToSpan(span, waitlistkeys.WaitlistIDKey, listID)
	tracing.AttachToSpan(span, waitlistkeys.WaitlistSignupIDKey, signupID)

	signup, err := r.GetSignup(ctx, tx, scope, listID, signupID)
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
func (r *repository) recordList(ctx context.Context, tx database.Tx, listID, auditEventType, changeEventType string) error {
	return r.record(ctx, tx, "", resourceTypeWaitlists, listID, auditEventType, changeEventType, map[string]any{
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
func (r *repository) recordSignup(ctx context.Context, tx database.Tx, signup *platformwaitlists.Signup, auditEventType, changeEventType string) error {
	return r.record(ctx, tx, signup.Subject.ID, resourceTypeWaitlistSignups, signup.ID, auditEventType, changeEventType, map[string]any{
		waitlistkeys.WaitlistSignupIDKey:     signup.ID,
		waitlistkeys.WaitlistIDKey:           signup.ListID,
		waitlistkeys.WaitlistSignupStatusKey: signup.Status.String(),
	})
}

// record writes the audit entry and enqueues the data change event, inside the
// caller's transaction.
//
// It used to open one of its own, because the write it describes had already
// committed inside the platform store. As of platform-go v14 that store takes
// the caller's executor, so the row, the entry and the event commit together or
// not at all.
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
	ctx context.Context, tx database.Tx,
	userID, resourceType, relevantID, auditEventType, changeEventType string,
	metadata map[string]any,
) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithSpan(span).WithValue(waitlistkeys.WaitlistIDKey, relevantID)

	return r.recorder.RecordAndEmit(ctx, tx, logger, &audit.AuditLogEntry{
		ID:            identifiers.New(),
		ResourceType:  resourceType,
		RelevantID:    relevantID,
		EventType:     auditEventType,
		BelongsToUser: userID,
	}, changeEventType, "", metadata)
}
