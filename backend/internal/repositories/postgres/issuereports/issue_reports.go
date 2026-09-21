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

# The transaction the events are in

Every hand-written repository here emits inside the transaction that wrote
the row, so the event lives or dies with what it describes (see
internal/repositories/postgres/events). This one now does too. It could not
before: platform's writes owned their transactions and took no executor, so
the audit entry and the event were a second transaction after the first had
committed, and a report could exist that nothing had recorded.

As of platform-go v14 a store write takes the caller's database.Tx, so the
write, the entry and the event are one transaction and share one fate. The
gap filed upstream as platform-go #465 is closed by that convention rather
than by anything here, which is why this package has no workaround to
delete.
*/
package issuereports

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/keys"

	platformissuereports "github.com/primandproper/platform-go/v14/issuereports"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// resourceTypeIssueReports is what an audit entry about an issue report names.
const resourceTypeIssueReports = "issue_reports"

var _ platformissuereports.Store = (*repository)(nil)

// CreateReport files the report, then records it.
func (r *repository) CreateReport(ctx context.Context, tx database.Tx, scope tenancy.Scope, report *platformissuereports.Report) (*platformissuereports.Report, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	result, err := r.Store.CreateReport(ctx, tx, scope, report)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, report.ID)

	// The stored row, not the input. As of platform-go v14 the store answers with what it
	// wrote and leaves the argument alone, so the input's id is still empty here — an entry
	// recorded from it names no report, and so does the event a subscriber receives.
	if err = r.record(ctx, tx, result, audit.AuditLogEventTypeCreated, ddbissuereports.IssueReportCreatedServiceEventType); err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateReport revises what the reporter said, then records it.
//
// It does not move the status and cannot: the lifecycle's one door is
// TransitionReport. So this always records an update, never a transition.
func (r *repository) UpdateReport(ctx context.Context, tx database.Tx, scope tenancy.Scope, report *platformissuereports.Report) (*platformissuereports.Report, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	result, err := r.Store.UpdateReport(ctx, tx, scope, report)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, report.ID)

	// The stored row for the same reason, though this input does carry an id: what is
	// recorded should be what was written, not what was asked for.
	if err = r.record(ctx, tx, result, audit.AuditLogEventTypeUpdated, ddbissuereports.IssueReportUpdatedServiceEventType); err != nil {
		return nil, err
	}

	return result, nil
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
	tx database.Tx,
	scope tenancy.Scope,
	reportID string,
	from, to platformissuereports.Status,
	resolution string,
) (*platformissuereports.Report, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, reportID)

	report, err := r.Store.TransitionReport(ctx, tx, scope, reportID, from, to, resolution)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, issuereportkeys.IssueReportStatusKey, report.Status.String())

	if err = r.record(ctx, tx, report, audit.AuditLogEventTypeUpdated, ddbissuereports.IssueReportTransitionedServiceEventType); err != nil {
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
func (r *repository) ArchiveReport(ctx context.Context, tx database.Tx, scope tenancy.Scope, reportID string) (*platformissuereports.Report, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, issuereportkeys.IssueReportIDKey, reportID)

	report, err := r.GetReport(ctx, tx, scope, reportID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching issue report for archive")
	}

	result, err := r.Store.ArchiveReport(ctx, tx, scope, reportID)
	if err != nil {
		return nil, err
	}

	if err = r.record(ctx, tx, report, audit.AuditLogEventTypeArchived, ddbissuereports.IssueReportArchivedServiceEventType); err != nil {
		return nil, err
	}

	return result, nil
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
// The account comes off the report's scope rather than off the context, because
// a report's tenant is the account it was filed under and that is the account a
// webhook subscriber is resolved within. A background job reaching here has no
// session, and an event with no account reaches no subscriber at all.
func (r *repository) record(ctx context.Context, tx database.Tx, report *platformissuereports.Report, auditEventType, changeEventType string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithSpan(span).WithValue(issuereportkeys.IssueReportIDKey, report.ID)

	accountID := report.Scope.Owner()

	return r.recorder.RecordAndEmit(ctx, tx, logger, &audit.AuditLogEntry{
		ID:               identifiers.New(),
		ResourceType:     resourceTypeIssueReports,
		RelevantID:       report.ID,
		EventType:        auditEventType,
		BelongsToUser:    report.Reporter,
		BelongsToAccount: &accountID,
	}, changeEventType, accountID, map[string]any{
		issuereportkeys.IssueReportIDKey:     report.ID,
		issuereportkeys.IssueReportStatusKey: report.Status.String(),
	})
}
