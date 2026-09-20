/*
Package auditlog mounts platform-go's audit read surface.

The service this replaces had three RPCs — an entry by id, a page for an
account, a page for a user — and platform's three are the same reads with the
account and user selectors moved into a Query.

One of them narrows, and this package widens it back. platform binds the query's
scope to the one the connection resolved, and this application's chains are
per-account or per-user: ScopeFor files an entry under its account when it has
one and under its actor otherwise, so that logins do not all serialize on one
chain. A signed-in caller therefore belongs to two chains at once, and a read of
one of them is a read that never shows somebody their own logins, signups or
password resets.

callerChains names both, and platform reads them as one. The page is merged
under a single cursor, and an entry by id is looked for in each in turn, off the
same set — so the list and the get cannot disagree about which chains somebody
belongs to, which is a disagreement that lands on the caller who did everything
right. That lives upstream rather than here because the merge is easy to get
quietly wrong and impossible to do in the store: a set of scopes cannot be bound
alongside a cursor and a limit on two of the three dialects audit serves.

One read is still this package's, and it is the one the resolver deliberately
cannot express: the operator's, every chain at once, which is a nil query scope
rather than a slice anybody could enumerate. operatorReader makes that decision,
and which chains a caller may read is not something a grant on the method can
say — the grant is an account member's, for the reason waitlists' subject read
is decided inside its handler rather than by PermissionReadSignups.

The privacy export remains the answer for a subject asking for everything held
about them, because that reaches the repository rather than the wire.
*/
package auditlog

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	platformaudit "github.com/primandproper/platform-go/v14/audit"
	"github.com/primandproper/platform-go/v14/audit/auditpb"
	auditgrpc "github.com/primandproper/platform-go/v14/audit/grpc"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/samber/do/v2"
)

// activeAccountScope is the chain a request reads.
//
// The account, because that is the chain an account's own activity is filed
// under and the one an operator or a member is asking about. A caller with no
// active account yields the zero Scope, which the reader refuses rather than
// widening — see tenancy.Of.
func activeAccountScope(ctx context.Context) (tenancy.Scope, error) {
	return tenancy.Of(sessions.FromContext(ctx).GetActiveAccountID()), nil
}

// RegisterAuditService registers platform's audit read surface.
func RegisterAuditService(i do.Injector) {
	do.Provide[auditpb.AuditServiceServer](i, func(i do.Injector) (auditpb.AuditServiceServer, error) {
		return auditgrpc.NewServer(
			// Wrapped for the two reads the chains resolver does not reach: an entry by
			// id, and an operator's. See operatorReader.
			operatorReader{Reader: do.MustInvoke[platformaudit.Reader](i)},
			do.MustInvoke[database.Client](i),
			auditgrpc.WithScopeResolver(activeAccountScope),
			auditgrpc.WithChainsResolver(callerChains),
			auditgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			auditgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			auditgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
