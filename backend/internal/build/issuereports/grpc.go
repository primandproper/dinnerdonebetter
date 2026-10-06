/*
Package issuereports mounts platform-go's issue reports surface.

It is the first surface this repo mounts whose rows belong to an account rather
than to the deployment, so it takes sessions.AccountScopedPrincipalFromContext
rather than the global one every other surface takes. See that type for why
getting it wrong is silent.
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
// the operator's who holds ReadAnyIssueReportsPermission.
//
// The scope has already done most of the work by the time this is asked — a
// report in another account is not found rather than refused — so what is left
// is one account's members not reading each other's reports.
//
// The operator half reads a permission off the caller's grants rather than the
// name of their role. platform's ReportAuthorizer documentation asks for exactly
// that, and names the same grant: the one that reads every account's queue is
// the one that reads any report in it. The reporter half is platform's own
// ReporterAuthorizer, which refuses a caller with no identifier rather than
// matching them against a report filed by nobody.
type ownReportOrAdmin struct {
	issuereportsgrpc.ReporterAuthorizer

	grants platformauthz.GrantsExtractor
}

// operator reports whether the caller holds the grant that reads any report.
func (a ownReportOrAdmin) operator(ctx context.Context) bool {
	grants, ok := a.grants(ctx)

	return ok && grants.Has(authorization.ReadAnyIssueReportsPermission)
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
			// Account-scoped, not global. See the package comment.
			sessions.AccountScopedPrincipalFromContext,
			ownReportOrAdmin{grants: sessions.GrantsFromContext},
			issuereportsgrpc.WithGrantsExtractor(sessions.GrantsFromContext),
			issuereportsgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			issuereportsgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			issuereportsgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
