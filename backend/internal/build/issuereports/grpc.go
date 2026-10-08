/*
Package issuereports mounts platform-go's issue reports surface.

A report is between the person who filed it and the service's administrators: a bug, a
complaint about another user, a creation somebody thinks is low quality. It is nobody else's,
and in particular not the business of whoever administers the reporter's household, so every
report is filed under the global scope and the surface takes sessions.PrincipalFromContext as
every other global surface here does. What keeps that from being one queue everybody can read
is the policy rather than the scope: filing and reading one's own reports are every user's,
and paging, revising, moving and archiving the queue are a service administrator's alone.
*/
package issuereports

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	"github.com/primandproper/platform-go/v15/callers"
	platformissuereports "github.com/primandproper/platform-go/v15/issuereports"
	issuereportsgrpc "github.com/primandproper/platform-go/v15/issuereports/grpc"
	"github.com/primandproper/platform-go/v15/issuereports/issuereportspb"
	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// ownReportOrAdmin is this deployment's rule: a report is its reporter's, and
// the service administrator's who works the queue.
//
// Every report is in the one global scope, so the scope decides nothing here and
// this is the whole of what keeps one user from reading another's reports — the
// reporter of a harassment complaint and the member it is about included.
//
// The administrator half reads a permission off the caller's grants rather than
// the name of their role. platform's ReportAuthorizer documentation asks for
// exactly that, and its example admits a caller who triages: the grant that pages
// the queue is the one that opens a report in it. The reporter half is platform's
// own ReporterAuthorizer, which refuses a caller with no identifier rather than
// matching them against a report filed by nobody.
type ownReportOrAdmin struct {
	issuereportsgrpc.ReporterAuthorizer

	grants platformauthz.GrantsExtractor
}

// OwnReportOrAdmin is the rule above over grants, for a surface other than the gRPC one —
// the MCP tools — so that who may read a report is decided once.
func OwnReportOrAdmin(grants platformauthz.GrantsExtractor) issuereportsgrpc.ReportAuthorizer {
	return ownReportOrAdmin{grants: grants}
}

// operator reports whether the caller holds the grant that works the queue.
func (a ownReportOrAdmin) operator(ctx context.Context) bool {
	grants, ok := a.grants(ctx)

	return ok && grants.Has(authorization.TriageIssueReportsPermission)
}

// AuthorizeReport is asked once a keyed read has resolved whose report it is.
func (a ownReportOrAdmin) AuthorizeReport(ctx context.Context, caller callers.Principal, report *platformissuereports.Report) error {
	if a.ReporterAuthorizer.AuthorizeReport(ctx, caller, report) == nil || a.operator(ctx) {
		return nil
	}

	return callers.ErrTargetNotPermitted
}

// AuthorizeReporter is asked before paging the reports one person filed.
func (a ownReportOrAdmin) AuthorizeReporter(ctx context.Context, caller callers.Principal, reporter string) error {
	if a.ReporterAuthorizer.AuthorizeReporter(ctx, caller, reporter) == nil || a.operator(ctx) {
		return nil
	}

	return callers.ErrTargetNotPermitted
}

// RegisterIssueReportsService registers platform's issue reports surface.
func RegisterIssueReportsService(i do.Injector) {
	do.Provide[issuereportspb.IssueReportsServiceServer](i, func(i do.Injector) (issuereportspb.IssueReportsServiceServer, error) {
		return issuereportsgrpc.NewServer(
			do.MustInvoke[platformissuereports.Store](i),
			do.MustInvoke[database.Client](i),
			// Global, not account-scoped. See the package comment.
			sessions.PrincipalFromContext,
			ownReportOrAdmin{grants: sessions.GrantsFromContext},
			issuereportsgrpc.WithGrantsExtractor(sessions.GrantsFromContext),
			issuereportsgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			issuereportsgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			issuereportsgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

// PermissionOverrides re-declares the two cross-scope reads under the grant that
// pages the queue.
//
// platform puts them behind PermissionReadAnyReports, for a deployment whose
// triagers each work one tenant's queue and whose operator reads all of them. This
// one files every report under the global scope, so there is one queue, the
// ordinary ListReports and ListReportsByStatus already page all of it, and the
// administrator who triages it is the only caller with any business reading it.
// Nobody holds the read-any grant, and without these the two methods would be
// declared behind a permission no role grants: callable by nobody, and refused
// exactly as they would be for a caller who genuinely lacked it.
func PermissionOverrides() map[string][]authorization.Permission {
	return map[string][]authorization.Permission{
		issuereportspb.IssueReportsService_ListReportsAcrossScopes_FullMethodName:         {authorization.TriageIssueReportsPermission},
		issuereportspb.IssueReportsService_ListReportsByStatusAcrossScopes_FullMethodName: {authorization.TriageIssueReportsPermission},
	}
}
