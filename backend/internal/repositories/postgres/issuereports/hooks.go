/*
Package issuereports records what an issue report write means to the rest of
this application. The reports themselves are platform-go's: the schema, the
paging, the tenancy column, the triage lifecycle and the erasure all live there,
and this package neither reimplements nor wraps them.

What it adds is the half platform cannot know about — an audit log entry naming
who did what, and a data change event on the outbox that the webhook dispatcher
fans out. issue_report_created, issue_report_updated, issue_report_transitioned
and issue_report_archived are all in the webhook event catalog, so a subscriber
can already ask for them; a write that skipped the pair would be a row with no
provenance and a subscriber that never heard.

Both are written from platform's issuereports.Hooks, which the store calls on
the caller's transaction once each write has landed. A hook's error fails the
write, so the report, the entry and the event commit together or not at all.
*/
package issuereports

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	platformissuereports "github.com/primandproper/platform-go/v14/issuereports"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// resourceTypeIssueReports is what an audit entry about an issue report names.
const resourceTypeIssueReports = "issue_reports"

// hooks records every issue report write.
//
// It implements Hooks outright rather than embedding NoopHooks, so a write
// platform adds later breaks this build until somebody decides what it records.
// Embedded, the new write would compile and record nothing, which is the one
// failure an audit log cannot notice.
type hooks struct {
	logger   logging.Logger
	recorder *recording.Recorder
}

var _ platformissuereports.Hooks = (*hooks)(nil)

// AfterCreateReport records a report being filed.
func (h *hooks) AfterCreateReport(ctx context.Context, tx database.Tx, _ tenancy.Scope, report *platformissuereports.Report) error {
	return h.record(ctx, tx, report, platformaudit.EventCreated, ddbissuereports.IssueReportCreatedServiceEventType, nil)
}

// AfterUpdateReport records what the reporter said being revised, and which of
// its fields moved. It never records a transition: a revision cannot move the
// status, because the lifecycle's one door is TransitionReport.
func (h *hooks) AfterUpdateReport(ctx context.Context, tx database.Tx, _ tenancy.Scope, before, after *platformissuereports.Report) error {
	changes, err := platformaudit.Diff(before, after)
	if err != nil {
		return platformerrors.Wrap(err, "diffing the updated issue report")
	}

	return h.record(ctx, tx, after, platformaudit.EventUpdated, ddbissuereports.IssueReportUpdatedServiceEventType, changes)
}

// AfterTransitionReport records a report moving through the triage lifecycle,
// with the status, note and closing stamp it moved away from in the diff, so
// "who resolved this, and what did it say before they reopened it" is
// answerable from the audit log rather than from the one row the last write
// left behind.
//
// A transition whose guard did not match never reaches here: the caller's view
// of the row was one write out of date, which is not a fact about the report
// worth putting in its audit trail.
func (h *hooks) AfterTransitionReport(ctx context.Context, tx database.Tx, _ tenancy.Scope, before, after *platformissuereports.Report) error {
	changes, err := platformaudit.Diff(before, after)
	if err != nil {
		return platformerrors.Wrap(err, "diffing the transitioned issue report")
	}

	return h.record(ctx, tx, after, platformaudit.EventUpdated, ddbissuereports.IssueReportTransitionedServiceEventType, changes)
}

// AfterArchiveReport records a report being taken out of the queue. The row is
// the one the archive left, which names the same reporter, scope and status the
// report held before it.
func (h *hooks) AfterArchiveReport(ctx context.Context, tx database.Tx, _ tenancy.Scope, report *platformissuereports.Report) error {
	return h.record(ctx, tx, report, platformaudit.EventArchived, ddbissuereports.IssueReportArchivedServiceEventType, nil)
}

// AfterDeleteReportsByReporter records nothing, deliberately: an erasure was
// never recorded here, and an entry naming the reporter would put back into the
// audit log the reference the erasure exists to remove.
func (*hooks) AfterDeleteReportsByReporter(context.Context, database.Tx, tenancy.Scope, string, int64) error {
	return nil
}

// record writes the audit entry and enqueues the data change event on the
// transaction the write ran in. changes is the field-level diff an update or a
// transition carries, and nil for every other write.
//
// The two travel together because they answer the same question from opposite
// sides — the audit log for whoever asks later who did this, the outbox for
// whoever needs to know now — and a write that carried one without the other
// would be a write nobody could tell was incomplete.
//
// The account comes off the report's scope rather than off the context, because
// a report's tenant is the account it was filed under and that is the account a
// webhook subscriber is resolved within. A background job reaching here has no
// session, and an event with no account reaches no subscriber at all.
func (h *hooks) record(
	ctx context.Context,
	tx database.Tx,
	report *platformissuereports.Report,
	auditEventType platformaudit.EventType,
	changeEventType string,
	changes map[string]platformaudit.Change,
) error {
	logger := h.logger.WithValue(issuereportkeys.IssueReportIDKey, report.ID)

	accountID := report.Scope.Owner()

	entry := audit.NewEntry(report.Reporter, accountID, resourceTypeIssueReports, report.ID, auditEventType)
	entry.Changes = changes

	return h.recorder.RecordAndEmit(ctx, tx, logger, entry, changeEventType, accountID, map[string]any{
		issuereportkeys.IssueReportIDKey:     report.ID,
		issuereportkeys.IssueReportStatusKey: report.Status.String(),
	})
}
