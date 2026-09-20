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

callerChains names both, and platform pages them as one — one cursor across
every chain, merged into the page a single-chain read would have returned. That
lives upstream rather than here because the merge is easy to get quietly wrong
and impossible to do in the store: a set of scopes cannot be bound alongside a
cursor and a limit on two of the three dialects audit serves.

Two reads are still this package's. An entry by id takes the connection's single
scope, so a caller who finds one of their own in a list and asks for it by id
would be told it does not exist; and the operator's read — every chain at once,
which is a nil query scope and not a slice anybody can enumerate — is a
deployment policy platform's own surface declines to have. operatorReader
answers both. Which chains a caller may read is not something a grant on the
method can say, and it is decided there for the reason waitlists decides a
subject read inside its handler: the grant is held by an account member, and the
question it cannot answer is whose log this is.

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
