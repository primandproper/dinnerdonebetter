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

Both are written from platform's waitlists.Hooks, which the store calls on the
caller's transaction once each write has landed. A hook's error fails the
write, so the row, the entry and the event commit together or not at all.

# Why a withdrawal is recorded apart from a transition

Invite and Convert move somebody through a queue; Withdraw is a standing
instruction to stop writing to an address. They emit different events for that
reason. The withdrawal's entry names the subject from the row platform hands
AfterWithdraw, which is the row from before the blanking: read afterwards, every
withdrawal in the audit log would belong to nobody.
*/
package waitlists

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbwaitlists "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	waitlistkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformwaitlists "github.com/primandproper/platform-go/v15/waitlists"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

const (
	// resourceTypeWaitlists is what an audit entry about a list names.
	resourceTypeWaitlists = "waitlists"
	// resourceTypeWaitlistSignups is what an audit entry about a signup names.
	resourceTypeWaitlistSignups = "waitlist_signups"
)

// hooks records every waitlist write.
//
// It implements Hooks outright rather than embedding NoopHooks, so a write
// platform adds later breaks this build until somebody decides what it records.
// Embedded, the new write would compile and record nothing, which is the one
// failure an audit log cannot notice.
type hooks struct {
	logger   logging.Logger
	recorder *recording.Recorder
}

var _ platformwaitlists.Hooks = (*hooks)(nil)

// AfterCreateList records a list being opened.
func (h *hooks) AfterCreateList(ctx context.Context, tx database.Tx, _ tenancy.Scope, list *platformwaitlists.List) error {
	return h.recordList(ctx, tx, list.ID, platformaudit.EventCreated, ddbwaitlists.WaitlistCreatedServiceEventType)
}

// AfterUpdateList records a list being rewritten, and which of its fields moved.
func (h *hooks) AfterUpdateList(ctx context.Context, tx database.Tx, _ tenancy.Scope, before, after *platformwaitlists.List) error {
	changes, err := platformaudit.Diff(before, after)
	if err != nil {
		return platformerrors.Wrap(err, "diffing the updated waitlist")
	}

	return h.record(ctx, tx, "", resourceTypeWaitlists, after.ID, platformaudit.EventUpdated, ddbwaitlists.WaitlistUpdatedServiceEventType, changes, map[string]any{
		waitlistkeys.WaitlistIDKey: after.ID,
	})
}

// AfterArchiveList records a list being retired.
func (h *hooks) AfterArchiveList(ctx context.Context, tx database.Tx, _ tenancy.Scope, list *platformwaitlists.List) error {
	return h.recordList(ctx, tx, list.ID, platformaudit.EventArchived, ddbwaitlists.WaitlistArchivedServiceEventType)
}

// AfterJoin records somebody joining a list.
func (h *hooks) AfterJoin(ctx context.Context, tx database.Tx, _ tenancy.Scope, signup *platformwaitlists.Signup) error {
	return h.recordSignup(ctx, tx, signup, signup.Status, platformaudit.EventCreated, ddbwaitlists.WaitlistSignupCreatedServiceEventType)
}

// AfterUpdateSignupNotes records the operator's note changing. It is an update
// rather than a transition, which is the distinction the write exists to make:
// a note is the one write that touches a signup without moving anybody.
func (h *hooks) AfterUpdateSignupNotes(ctx context.Context, tx database.Tx, _ tenancy.Scope, before, after *platformwaitlists.Signup) error {
	changes, err := platformaudit.Diff(before, after)
	if err != nil {
		return platformerrors.Wrap(err, "diffing the updated waitlist signup")
	}

	return h.record(ctx, tx, after.Subject.ID, resourceTypeWaitlistSignups, after.ID, platformaudit.EventUpdated, ddbwaitlists.WaitlistSignupUpdatedServiceEventType, changes, signupMetadata(after, after.Status))
}

// AfterConfirm records a pending signup entering the queue.
func (h *hooks) AfterConfirm(ctx context.Context, tx database.Tx, _ tenancy.Scope, signup *platformwaitlists.Signup) error {
	return h.recordTransition(ctx, tx, signup)
}

// AfterInvite records somebody being let in.
func (h *hooks) AfterInvite(ctx context.Context, tx database.Tx, _ tenancy.Scope, signup *platformwaitlists.Signup) error {
	return h.recordTransition(ctx, tx, signup)
}

// AfterConvert records an invitation being taken up.
func (h *hooks) AfterConvert(ctx context.Context, tx database.Tx, _ tenancy.Scope, signup *platformwaitlists.Signup) error {
	return h.recordTransition(ctx, tx, signup)
}

// AfterWithdraw records somebody coming off a list at their own request.
//
// The row is the one from before the blanking, so it still names the subject;
// the status recorded is the one it is in now.
func (h *hooks) AfterWithdraw(ctx context.Context, tx database.Tx, _ tenancy.Scope, signup *platformwaitlists.Signup) error {
	return h.recordSignup(ctx, tx, signup, platformwaitlists.StatusWithdrawn, platformaudit.EventUpdated, ddbwaitlists.WaitlistSignupWithdrawnServiceEventType)
}

// AfterWithdrawSignupsForSubject records nothing, deliberately: an erasure was
// never recorded here, and an entry naming the subject would put back into the
// audit log the reference the erasure exists to remove.
func (*hooks) AfterWithdrawSignupsForSubject(context.Context, database.Tx, tenancy.Scope, platformwaitlists.Subject, int64) error {
	return nil
}

// AfterArchiveSignup records a signup being retired administratively.
func (h *hooks) AfterArchiveSignup(ctx context.Context, tx database.Tx, _ tenancy.Scope, signup *platformwaitlists.Signup) error {
	return h.recordSignup(ctx, tx, signup, signup.Status, platformaudit.EventArchived, ddbwaitlists.WaitlistSignupArchivedServiceEventType)
}

// recordTransition is Confirm, Invite and Convert: one event for a move through
// the queue, carrying the status the signup landed in.
func (h *hooks) recordTransition(ctx context.Context, tx database.Tx, signup *platformwaitlists.Signup) error {
	return h.recordSignup(ctx, tx, signup, signup.Status, platformaudit.EventUpdated, ddbwaitlists.WaitlistSignupTransitionedServiceEventType)
}

// recordList writes the audit entry and the data change event for a write to the
// catalog.
//
// The entry names no user, because a list is an administrative row that belongs
// to nobody: who opened it is the actor on the context, which is what the audit
// recorder resolves.
func (h *hooks) recordList(ctx context.Context, tx database.Tx, listID string, auditEventType platformaudit.EventType, changeEventType string) error {
	return h.record(ctx, tx, "", resourceTypeWaitlists, listID, auditEventType, changeEventType, nil, map[string]any{
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
func (h *hooks) recordSignup(
	ctx context.Context,
	tx database.Tx,
	signup *platformwaitlists.Signup,
	status platformwaitlists.Status,
	auditEventType platformaudit.EventType,
	changeEventType string,
) error {
	return h.record(ctx, tx, signup.Subject.ID, resourceTypeWaitlistSignups, signup.ID, auditEventType, changeEventType, nil, signupMetadata(signup, status))
}

// signupMetadata is what a signup's data change event carries.
func signupMetadata(signup *platformwaitlists.Signup, status platformwaitlists.Status) map[string]any {
	return map[string]any{
		waitlistkeys.WaitlistSignupIDKey:     signup.ID,
		waitlistkeys.WaitlistIDKey:           signup.ListID,
		waitlistkeys.WaitlistSignupStatusKey: status.String(),
	}
}

// record writes the audit entry and enqueues the data change event on the
// transaction the write ran in. changes is the field-level diff an update
// carries, and nil for every other write.
//
// The event names no account, which is what makes it a service-wide event: these
// rows live in the global scope, so the account it reaches a webhook subscriber
// under is whichever one the requester had active, resolved from the context by
// the emitter.
func (h *hooks) record(
	ctx context.Context, tx database.Tx,
	userID, resourceType, relevantID string, auditEventType platformaudit.EventType, changeEventType string,
	changes map[string]platformaudit.Change,
	metadata map[string]any,
) error {
	logger := h.logger.WithValue(waitlistkeys.WaitlistIDKey, relevantID)

	entry := audit.NewEntry(userID, "", resourceType, relevantID, auditEventType)
	entry.Changes = changes

	return h.recorder.RecordAndEmit(ctx, tx, logger, entry, changeEventType, "", metadata)
}
