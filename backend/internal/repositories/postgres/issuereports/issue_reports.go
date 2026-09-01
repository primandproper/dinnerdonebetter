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

# The transaction the events are not in

Every hand-written repository here emits inside the transaction that wrote the
row, so the event lives or dies with what it describes (see
internal/repositories/postgres/events). This one cannot: platform's
CreateReport, UpdateReport, TransitionReport and ArchiveReport own their
transactions and take no executor, so the audit entry and the event are a second
transaction after the first has committed.

The gap that opens is the ordinary one — the report lands, the process dies, and
nothing is recorded about it. It is narrow and it is one-directional: a report
can exist with no event, but no event can name a report that was not written.
Closing it needs platform's write methods to accept a database.Tx the way
DeleteReportsByReporter already does. That is filed upstream as platform-go
#457 rather than worked around here — a gap papered over locally stops being a
gap anyone remembers.
*/
package issuereports

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/keys"

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/identifiers"
	platformissuereports "github.com/primandproper/platform-go/v13/issuereports"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/tenancy"
)

// resourceTypeIssueReports is what an audit entry about an issue report names.
const resourceTypeIssueReports = "issue_reports"

var _ platformissuereports.Store = (*repository)(nil)

// CreateReport files the report, then records it.
func (r *repository) CreateReport(ctx context.Context, report *platformissuereports.Report) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if err := r.Store.CreateReport(ctx, report); err != nil {
		return err
	}

	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, report.ID)

	return r.record(ctx, report, audit.AuditLogEventTypeCreated, ddbissuereports.IssueReportCreatedServiceEventType)
}

// UpdateReport revises what the reporter said, then records it.
//
// It does not move the status and cannot: the lifecycle's one door is
// TransitionReport. So this always records an update, never a transition.
func (r *repository) UpdateReport(ctx context.Context, report *platformissuereports.Report) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if err := r.Store.UpdateReport(ctx, report); err != nil {
		return err
	}

	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, report.ID)

	return r.record(ctx, report, audit.AuditLogEventTypeUpdated, ddbissuereports.IssueReportUpdatedServiceEventType)
}

// TransitionReport moves the report through the triage lifecycle, then records
// it.
//
// The report it records is the one platform returns rather than the one the
// caller described, so the event names the status the row actually holds. A
// transition whose guard did not match writes nothing and records nothing: the
// caller's view of the row was one write out of date, which is not a fact about
// the report worth putting in its audit trail.
func (r *repository) TransitionReport(
	ctx context.Context,
	scope tenancy.Scope,
	reportID string,
	from, to platformissuereports.Status,
	resolution string,
) (*platformissuereports.Report, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, reportID)

	report, err := r.Store.TransitionReport(ctx, scope, reportID, from, to, resolution)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, issuereportkeys.IssueReportStatusKey, report.Status.String())

	if err = r.record(ctx, report, audit.AuditLogEventTypeUpdated, ddbissuereports.IssueReportTransitionedServiceEventType); err != nil {
		return nil, err
	}

	return report, nil
}

// ArchiveReport removes the report from the queue, then records it.
//
// The report is read before the archive rather than after, because an audit
// entry names whose row it was and the archived row is the one this method is
// about. A read that fails is the archive's failure too: platform answers an
// absent, archived, or other-scope report as ErrReportNotFound either way, so
// returning it from here is the same answer one call earlier.
func (r *repository) ArchiveReport(ctx context.Context, scope tenancy.Scope, reportID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, reportID)

	report, err := r.GetReport(ctx, scope, reportID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching issue report for archive")
	}

	if err = r.Store.ArchiveReport(ctx, scope, reportID); err != nil {
		return err
	}

	return r.record(ctx, report, audit.AuditLogEventTypeArchived, ddbissuereports.IssueReportArchivedServiceEventType)
}

// record writes the audit entry and enqueues the data change event, in one
// transaction of their own.
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
func (r *repository) record(ctx context.Context, report *platformissuereports.Report, auditEventType, changeEventType string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithSpan(span).WithValue(issuereportkeys.IssueReportIDKey, report.ID)

	accountID := report.Scope.Owner()

	return r.client.WithTransaction(ctx, func(tx database.Tx) error {
		if err := r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ID:               identifiers.New(),
			ResourceType:     resourceTypeIssueReports,
			RelevantID:       report.ID,
			EventType:        auditEventType,
			BelongsToUser:    report.Reporter,
			BelongsToAccount: &accountID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "creating audit log entry")
		}

		if err := r.events.Emit(ctx, tx, logger, changeEventType, accountID, map[string]any{
			issuereportkeys.IssueReportIDKey:     report.ID,
			issuereportkeys.IssueReportStatusKey: report.Status.String(),
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "enqueuing data change event")
		}

		return nil
	})
}
