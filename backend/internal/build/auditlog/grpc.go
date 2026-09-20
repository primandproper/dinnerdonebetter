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

spanningReader reads both and merges, so "every entry about this user" is an API
read again. It is a decorator rather than a fix because the fix is upstream and
one predicate wide — see its doc — and the privacy export remains the answer for
a subject asking for everything held about them, because that reaches the
repository rather than the wire.
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
			// Wrapped so a caller reads their own chain alongside their account's.
			// See spanningReader, and the one-predicate change upstream that deletes it.
			spanningReader{Reader: do.MustInvoke[platformaudit.Reader](i)},
			do.MustInvoke[database.Client](i),
			auditgrpc.WithScopeResolver(activeAccountScope),
			auditgrpc.WithLogger(do.MustInvoke[logging.Logger](i)),
			auditgrpc.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			auditgrpc.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
