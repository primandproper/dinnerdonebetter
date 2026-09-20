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

	"github.com/primandproper/platform-go/v14/callers"
	platformissuereports "github.com/primandproper/platform-go/v14/issuereports"
	issuereportsgrpc "github.com/primandproper/platform-go/v14/issuereports/grpc"
	"github.com/primandproper/platform-go/v14/issuereports/issuereportspb"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// ownReportOrAdmin is this deployment's rule: a report is its reporter's, and a
// service administrator's.
//
// The scope has already done most of the work by the time this is asked — a
// report in another account is not found rather than refused — so what is left
// is one account's members not reading each other's reports.
type ownReportOrAdmin struct{}

// AuthorizeReport is asked once a keyed read has resolved whose report it is.
func (ownReportOrAdmin) AuthorizeReport(ctx context.Context, caller callers.Principal, report *platformissuereports.Report) error {
	if report != nil && caller != nil && report.Reporter == caller.UserID() {
		return nil
	}

	if sessions.FromContext(ctx).GetServicePermissions().IsServiceAdmin() {
		return nil
	}

	return callers.ErrTargetNotPermitted
}

// AuthorizeReporter is asked before paging the reports one person filed.
func (ownReportOrAdmin) AuthorizeReporter(ctx context.Context, caller callers.Principal, reporter string) error {
	if caller != nil && reporter == caller.UserID() {
		return nil
	}

	if sessions.FromContext(ctx).GetServicePermissions().IsServiceAdmin() {
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
			ownReportOrAdmin{},
			issuereportsgrpc.WithGrantsExtractor(sessions.GrantsFromContext),
			issuereportsgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			issuereportsgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			issuereportsgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
